/*
Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/*

Package rescheduling is using for HuaWei Ascend pin fault rescheduling.

*/
package rescheduling

import (
	"encoding/json"
	"errors"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog"
	"reflect"
	"strconv"
	"strings"
	"time"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"
)

func convertToReSchedulerNodesMapFromCM(buffer string) (map[string]FaultNodeState, error) {
	faultNode := map[string]FaultNodeState{}

	if unmarshalErr := json.Unmarshal([]byte(buffer), &faultNode); unmarshalErr != nil {
		return nil, unmarshalErr
	}

	return faultNode, nil
}

func convertToReSchedulerCardsMapFromCM(buffer string) (map[string]FaultNPUsOnNode, error) {
	faultNodeNPUs := map[string]FaultNPUsOnNode{}
	if unmarshalErr := json.Unmarshal([]byte(buffer), &faultNodeNPUs); unmarshalErr != nil {
		return nil, unmarshalErr
	}
	return faultNodeNPUs, nil
}

func convertToNodeHeartbeatMapFromCM(buffer string) (map[string]NormalNodeHeartbeat, error) {
	heartbeat := map[string]NormalNodeHeartbeat{}

	if unmarshalErr := json.Unmarshal([]byte(buffer), &heartbeat); unmarshalErr != nil {
		return nil, unmarshalErr
	}

	return heartbeat, nil
}

func getCMCardWriteData(nodeData interface{}) (string, error) {
	return marshalCacheDataToString(nodeData)
}

func getCMHeartbeatWriteData(nodeData interface{}) (string, error) {
	return marshalCacheDataToString(nodeData)
}

func marshalCacheDataToString(data interface{}) (string, error) {
	dataBuffer, err := json.Marshal(data)
	if err != nil {
		klog.V(logErrorLev).Infof("marshalCacheDataToString err: %v.", err)
		return "", err
	}
	return string(dataBuffer), nil
}

func getCMNodeWriteData(nodeData interface{}) (string, error) {
	return marshalCacheDataToString(nodeData)
}

func getNodeHeartbeatInfoFromCache(node *api.NodeInfo) (NormalNodeHeartbeat, error) {
	tmp, mapOK := ReSchedulerCache[CmNodeHeartbeatKind]
	if !mapOK {
		return NormalNodeHeartbeat{}, fmt.Errorf("ReSchedulerCache no kind of %v", CmNodeHeartbeatKind)
	}
	nodesHeartBeat, ok := tmp.(map[string]NormalNodeHeartbeat)
	if !ok {
		return NormalNodeHeartbeat{}, fmt.Errorf("getNodeHeartbeatFromCache assert %v failed", tmp)
	}
	nodeHeartBeat, getOk := nodesHeartBeat[node.Name]
	if !getOk {
		return NormalNodeHeartbeat{}, fmt.Errorf("getNodeHeartbeatFromCache %s nonexistent", node.Name)
	}

	return nodeHeartBeat, nil
}

func getNodeHeartbeatIntervalAndUpdateTimeFromCache(node *api.NodeInfo) (int64, int64, error) {
	heartbeatInfo, err := getNodeHeartbeatInfoFromCache(node)
	if err != nil {
		klog.V(logErrorLev).Infof("isNodeHealth %v.", err)
		return 0, 0, err
	}

	heartbeatInterval := heartbeatInfo.HeartbeatInterval
	maxInterval := int64(heartbeatInterval) * constIntNum3

	updateHeartbeatTime := heartbeatInfo.UpdateHeartbeatTime
	return maxInterval, updateHeartbeatTime, nil
}

func isNodeHealth(node *api.NodeInfo) bool {
	if !isEnableFaultNode(node) {
		klog.V(logDebugLev).Infof("isNodeHealth %s fault feature[%+v] not enable", node.Name, node.Node.Labels)
		return true
	}

	maxInterval, updateHeartbeatTime, err := getNodeHeartbeatIntervalAndUpdateTimeFromCache(node)
	if err != nil {
		return false
	}

	nowTime := time.Now().Unix()
	margin := nowTime - updateHeartbeatTime
	if margin < 0 {
		klog.V(logErrorLev).Infof(" isNodeHealth %s cache Time is newer[%d-%d], confused, skip.",
			node.Name, nowTime, updateHeartbeatTime)
	}
	if margin > maxInterval {
		klog.V(logErrorLev).Infof(" %s Time over %d [%d-%d],not health.",
			node.Name, maxInterval, nowTime, updateHeartbeatTime)
		return false
	}

	return true
}

// Delete expired node data.
func synReSchedulerNodeCache(ssn *framework.Session, tmpValue interface{}) error {
	nodeMap, assertOk := tmpValue.(map[string]FaultNodeState)
	if !assertOk {
		msg := fmt.Errorf("convert %v to map[string]FaultNodeState failed", tmpValue)
		klog.V(logDebugLev).Infof("synReSchedulerNodeCache %v.", msg)
		return msg
	}

	for nodeName, faultNode := range nodeMap {
		// 	No node
		nodeInf, ok := ssn.Nodes[nodeName]
		if !ok {
			klog.V(logErrorLev).Infof("delete %s from configMap due to not existence.", nodeName)
			delete(nodeMap, nodeName)
			continue
		}
		// For node changed health (Automatic recovery).
		if isNodeHealth(nodeInf) {
			klog.V(logErrorLev).Infof("delete %s from configMap due to node change health.", nodeName)
			delete(nodeMap, nodeName)
			continue
		}
		// For Node doesn't last too long
		preTime := faultNode.UpdateTime
		nowTime := time.Now().Unix()
		if nowTime-preTime > maxIntervalTime {
			klog.V(logErrorLev).Infof("delete %s from CM for overTime %v => %v.", nodeName, nowTime, preTime)
			delete(nodeMap, nodeName)
		}
	}
	ReSchedulerCache[CmNodeKind] = nodeMap

	return nil
}

func updateNodeIntoNodesHeartbeatTmp(ssn *framework.Session) error {
	nodesHeartbeat := make(map[string]NormalNodeHeartbeat, constIntNum3)
	temp, ok := ReSchedulerCache[CmNodeHeartbeatKind]
	if ok {
		nodesHeartbeat, ok = temp.(map[string]NormalNodeHeartbeat)
		if !ok {
			klog.V(logDebugLev).Infof("updateNodeIntoNodesHeartbeatTmp assert failed %v", temp)
			return fmt.Errorf("assert map[string]NormalNodeHeartbeat failed")
		}
	}

	for _, nodeInfo := range ssn.Nodes {
		if !isEnableFaultNode(nodeInfo) {
			klog.V(logDebugLev).Infof("%s fault feature not enable, not add in cache", nodeInfo.Name)
			continue
		}
		// 2.get node heartbeat
		updateTime := time.Now().Unix()
		oldHeartBeat := int64(-1)
		updateHeartbeatTime := updateTime
		nodeHeartBeat, ok := nodesHeartbeat[nodeInfo.Name]
		if ok {
			oldHeartBeat = nodeHeartBeat.NodeDHeartbeat
		}

		newHeartBeat, heartBeatErr := getNodeHeartbeat(nodeInfo)
		if heartBeatErr != nil {
			newHeartBeat = oldHeartBeat
			klog.V(logErrorLev).Infof("getNodeHeartbeat %v.", heartBeatErr)
		}
		if oldHeartBeat == newHeartBeat {
			updateHeartbeatTime = nodeHeartBeat.UpdateHeartbeatTime
		}
		// 3.get node HeartbeatInterval
		heartbeatInterval, intervalErr := getNodeHeartbeatInterval(nodeInfo)
		if intervalErr != nil {
			klog.V(logErrorLev).Infof("getNodeHeartbeatInterval %v, will use %d.", intervalErr, nodeUpdateTime)
		}
		// 4.add or update NodeHeartbeat
		newNodeHeartbeat := NormalNodeHeartbeat{
			NodeDHeartbeat:      newHeartBeat,
			UpdateHeartbeatTime: updateHeartbeatTime,
			HeartbeatInterval:   heartbeatInterval,
			UpdateTime:          updateTime,
		}
		nodesHeartbeat[nodeInfo.Name] = newNodeHeartbeat
	}
	ReSchedulerCache[CmNodeHeartbeatKind] = nodesHeartbeat
	return nil
}

// Delete expired heartbeat data.
func synNodeHeartbeatCache(ssn *framework.Session, tmpValue interface{}) error {
	nodesHeartbeat, assertOk := tmpValue.(map[string]NormalNodeHeartbeat)
	if !assertOk {
		msg := fmt.Errorf("assert %v to map[string]NormalNodeHeartbeat failed", tmpValue)
		klog.V(logErrorLev).Infof("synNodeHeartbeatCache %v.", msg)
		return msg
	}

	// delete nonexistent node
	for nodeName := range nodesHeartbeat {
		// 	No info
		_, ok := ssn.Nodes[nodeName]
		if !ok {
			klog.V(logErrorLev).Infof("delete %s from heartbeat due to not existence.", nodeName)
			delete(nodesHeartbeat, nodeName)
			continue
		}
	}

	ReSchedulerCache[CmNodeHeartbeatKind] = nodesHeartbeat

	return nil
}

// Delete expired card data.
func synReSchedulerCardCache(ssn *framework.Session, tmpValue interface{}) error {
	cardMap, assertOk := tmpValue.(map[string]FaultNPUsOnNode)
	if !assertOk {
		msg := fmt.Errorf("convert %v to map[string]FaultNPUsOnNode failed", tmpValue)
		klog.V(logErrorLev).Infof("synReSchedulerCardCache %v.", msg)
		return msg
	}

	for nodeName, cards := range cardMap {
		// 	No info
		nodeInf, ok := ssn.Nodes[nodeName]
		if !ok {
			klog.V(logErrorLev).Infof("delete %s from configMap due to not existence.", nodeName)
			delete(cardMap, nodeName)
			continue
		}
		// No fault NPUs.
		if !isNodeHasFaultNPUs(nodeInf) {
			klog.V(logErrorLev).Infof("delete %s from configMap due to card change health.", nodeName)
			delete(cardMap, nodeName)
			continue
		}
		// Timeout to delete
		preTime := cards.UpdateTime
		nowTime := time.Now().Unix()
		if nowTime-preTime > maxIntervalTime {
			klog.V(logErrorLev).Infof("delete %s from CM for overTime %v => %v.", nodeName, nowTime, preTime)
			delete(cardMap, nodeName)
		}
	}
	ReSchedulerCache[CmCardKind] = cardMap

	return nil
}

func getNodeFaultNPUs(node *api.NodeInfo, nodeNPUNumber int) ([]string, error) {
	npuStrings, ok := node.Node.Annotations[faultNPU]
	if !ok || len(npuStrings) == 0 {
		return nil, fmt.Errorf("%s get nil npus", node.Name)
	}

	faultNPUs := strings.Split(npuStrings, ",")
	if len(faultNPUs) > nodeNPUNumber {
		return nil, fmt.Errorf("%s get fault npus(%d)", node.Name, len(faultNPUs))
	}

	return faultNPUs, nil
}

func getNodeNetworkUnhealthyNPUs(node *api.NodeInfo, nodeNPUNumber int) ([]string, error) {
	npuStrings, ok := node.Node.Annotations[networkUnhealthyNPU]
	if !ok || len(npuStrings) == 0 {
		return nil, fmt.Errorf("%s get nil npus", node.Name)
	}

	faultNPUs := strings.Split(npuStrings, ",")
	if len(faultNPUs) > nodeNPUNumber {
		return nil, fmt.Errorf("%s get fault npus(%d)", node.Name, len(faultNPUs))
	}

	return faultNPUs, nil
}

// True:has fault NPUs/ network unhealthy card, otherwise return false.
func isNodeHasFaultNPUs(node *api.NodeInfo) bool {
	faultNPUStrings, npuOK := node.Node.Annotations[faultNPU]
	faultNetNPUStrings, netOK := node.Node.Annotations[networkUnhealthyNPU]
	if (!npuOK || len(faultNPUStrings) == 0) && (!netOK || len(faultNetNPUStrings) == 0) {
		return false
	}

	return true
}

func getNodeHeartbeatInterval(node *api.NodeInfo) (int, error) {
	var heartbeatInterval = nodeUpdateTime
	var err error
	value, ok := node.Node.Annotations[nodeHeartbeatInterval]
	if !ok || len(value) == 0 {
		klog.V(logErrorLev).Infof("isNodeHealth %s no [%s].", node.Name, nodeHeartbeat)
		return heartbeatInterval, fmt.Errorf("getFaultNodeState %s nil", node.Name)
	}

	// If the Time exceeds, the fault occurs.
	heartbeatInterval, err = strconv.Atoi(value)
	if err != nil {
		klog.V(logErrorLev).Infof("%s cover %s to int64 failed [%v].", node.Name, value, err)
		return heartbeatInterval, err
	}

	if heartbeatInterval > maxIntervalTime || heartbeatInterval < 1 {
		klog.V(logErrorLev).Infof("%s's heartbeatInterval %d over limit, will use %d.",
			node.Name, heartbeatInterval, nodeUpdateTime)
	}
	return heartbeatInterval, nil
}

func getNodeHeartbeat(node *api.NodeInfo) (int64, error) {
	const constNumber10 = 10
	const constNumber64 = 64
	value, ok := node.Node.Annotations[nodeHeartbeat]
	if !ok || len(value) == 0 {
		klog.V(logErrorLev).Infof("isNodeHealth %s no [%s].", node.Name, nodeHeartbeat)
		return 0, fmt.Errorf("getFaultNodeState %s nil", node.Name)
	}

	heartbeatTime, err := strconv.ParseInt(value, constNumber10, constNumber64)
	if err != nil {
		klog.V(logErrorLev).Infof("%s cover %s to int64 failed [%v].", node.Name, value, err)
		return 0, err
	}
	return heartbeatTime, nil
}

// In parameter 'node' is fault node.
func getFaultNodeState(node *api.NodeInfo) (FaultNodeState, error) {
	heartbeatTime, err := getNodeHeartbeat(node)
	if err != nil {
		return FaultNodeState{}, err
	}

	heartbeatInterval, intervalErr := getNodeHeartbeatInterval(node)
	if intervalErr != nil {
		klog.V(logErrorLev).Infof("getNodeHeartbeatInterval %v, will use %d.", err, nodeUpdateTime)
	}

	return FaultNodeState{
		NodeName:          node.Name,
		HealthCode:        0,
		UpdateTime:        time.Now().Unix(),
		Heartbeat:         heartbeatTime,
		HeartbeatInterval: heartbeatInterval,
	}, nil
}

func isEnableFaultNode(node *api.NodeInfo) bool {
	value, ok := node.Node.Labels[nodeDEnableKey]
	if !ok {
		return false
	}

	switch value {
	case nodeDEnableOnValue:
		return true
	case nodeDEnableOffValue:
		return false
	default:
		klog.V(logErrorLev).Infof("isEnableFaultNode not support %s.", value)
		return false
	}
}

func getInoperableNodes(nodes map[string]*api.NodeInfo) ([]FaultNodeState, error) {
	var faultNPUNodes []FaultNodeState

	for _, nodeInfo := range nodes {

		if isNodeHealth(nodeInfo) {
			continue
		}

		nodeState, stateErr := getFaultNodeState(nodeInfo)
		if stateErr != nil {
			klog.V(logDebugLev).Infof("getInoperableNodes %+v.", faultNPUNodes)
			continue
		}
		faultNPUNodes = append(faultNPUNodes, nodeState)
	}

	if len(faultNPUNodes) == 0 {
		return nil, errors.New("nil inoperable nodes")
	}
	klog.V(logDebugLev).Infof("getInoperableNodes %+v.", faultNPUNodes)

	return faultNPUNodes, nil
}

func getInoperableNPUCards(nodes map[string]*api.NodeInfo, npuNumber int) ([]FaultNPUsOnNode, error) {
	var faultNPUs []FaultNPUsOnNode

	for _, nodeInfo := range nodes {
		npus, err := getNodeFaultNPUs(nodeInfo, npuNumber)
		if err != nil {
			klog.V(logDebugLev).Infof("getNodeFaultNPUs err:%v.", err)
		}
		networkNPUs, netErr := getNodeNetworkUnhealthyNPUs(nodeInfo, npuNumber)
		if netErr != nil {
			klog.V(logDebugLev).Infof("getNodeNetworkUnhealthyNPUs err:%v.", netErr)
		}

		if err != nil && netErr != nil {
			continue
		}
		faultNPUs = append(faultNPUs, FaultNPUsOnNode{nodeInfo.Name, npus, networkNPUs, time.Now().Unix()})
	}

	if len(faultNPUs) == 0 {
		return nil, fmt.Errorf("%v nil inoperable NPU", reflect.ValueOf(nodes).MapKeys())
	}
	klog.V(logDebugLev).Infof("getInoperableNPUCards %+v.", faultNPUs)

	return faultNPUs, nil
}

func getFaultNodePODAndRankIndex(job *api.JobInfo, nodes map[string]*v1.Pod) (FaultNPUJob, error) {
	var faultJob = FaultNPUJob{
		faultNPUJobBase{
			job.Name,
			job.Namespace,
			make(map[string]string, constIntNum3),
			make(map[string]string, constIntNum3),
		},
		make(map[string]string, constIntNum3),
	}

	for _, task := range job.Tasks {
		if pod, ok := nodes[task.NodeName]; ok {
			rankIndex, err := getPodRankIndex(pod)
			if err != nil {
				klog.V(logErrorLev).Infof("getPodRankIndex %s %v.", pod.Name, err)
				return faultJob, err
			}
			faultJob.taskUseRankIndex[task.Name] = rankIndex
			faultJob.taskUseNode[task.Name] = task.NodeName
			faultJob.taskUseNPUs[task.Name] = pod.Annotations[npu800And9000CardName]
		}
	}

	if len(faultJob.taskUseRankIndex) == 0 {
		return faultJob, errors.New("get nil rankIndex")
	}

	return faultJob, nil
}

// For get ReSchedulerTasks from ReSchedulerData
func getReSchedulerTasksFromCache(task *api.TaskInfo) (ReSchedulerTasks, error) {
	jobValue, ok := ReSchedulerCache[CmJobKind]
	if !ok {
		klog.V(logErrorLev).Infof("getReSchedulerTasksFromCache no fault task in ReSchedulerCache.")
		return ReSchedulerTasks{}, nil
	}

	jobMap, ok := jobValue.(map[api.JobID]ReSchedulerTasks)
	if !ok {
		mgs := fmt.Errorf("not type(map[api.JobID]ReSchedulerTasks) %v ", jobMap)
		klog.V(logErrorLev).Infof("%v.", mgs)
		return ReSchedulerTasks{}, mgs
	}

	value, ok := jobMap[task.Job]
	if !ok {
		mgs := fmt.Errorf("no %v in jobMap", task.Job)
		klog.V(logErrorLev).Infof("%v.", mgs)
		return ReSchedulerTasks{}, mgs
	}

	return value, nil
}

// GetFaultTaskUseNodeInfo Get task used node.
func GetFaultTaskUseNodeInfo(task *api.TaskInfo, ssn *framework.Session) (*api.NodeInfo, error) {
	faultTasks, err := getReSchedulerTasksFromCache(task)
	if err != nil {
		return nil, err
	}

	nodeName, ok := faultTasks.NodeNames[task.Name]
	if !ok {
		return nil, fmt.Errorf("get taskName %s failed", task.Name)
	}

	node, ok := ssn.Nodes[nodeName]
	if !ok {
		return nil, fmt.Errorf("get node name %s failed", nodeName)
	}

	if IsNodeInFaultNodeList(node) {
		return nil, fmt.Errorf("GetFaultTaskUseNodeInfo %s in fault node list", nodeName)
	}
	return node, nil
}

// IsNodeInFaultNodeList Check whether the node is in the faulty node list, used for noded.
func IsNodeInFaultNodeList(node *api.NodeInfo) bool {
	nodeMap, ok := ReSchedulerCache[CmNodeKind]
	if !ok {
		return false
	}

	faultNodes, nodeErr := nodeMap.(map[string]FaultNodeState)
	if !nodeErr {
		return false
	}

	for _, value := range faultNodes {
		if value.NodeName == node.Name {
			return true
		}
	}
	return false
}

func isNodeInFaultJobUseList(node *api.NodeInfo) bool {
	faultJobMap, ok := ReSchedulerCache[CmJobKind]
	if !ok {
		return false
	}
	faultJob, jobErr := faultJobMap.(map[api.JobID]ReSchedulerTasks)
	if !jobErr {
		return false
	}

	for _, value := range faultJob {
		for _, nodeName := range value.NodeNames {
			if nodeName == node.Name {
				return true
			}
		}
	}

	return false
}

// GetNetworkUnhealthyCards Get the network Unhealthy npu Cards in a node.
func GetNetworkUnhealthyCards(nodeName string) []int {
	tmpData, ok := ReSchedulerCache[CmCardKind]
	if !ok {
		klog.V(logDebugLev).Infof("GetNetworkUnhealthyCards %s not in cache.", nodeName)
		return nil
	}
	faultNPUMap, cardErr := tmpData.(map[string]FaultNPUsOnNode)
	if !cardErr {
		klog.V(logErrorLev).Infof("GetNetworkUnhealthyCards %v convert to FaultNPUsOnNode map failed.", tmpData)
		return nil
	}
	faultNPUs, getErr := faultNPUMap[nodeName]
	if !getErr {
		klog.V(logDebugLev).Infof("GetNetworkUnhealthyCards FaultNPUsOnNode no %s.", nodeName)
		return nil
	}

	var topInt []int
	for _, cardStr := range faultNPUs.NetworkUnhealthyNPUs {
		v := strings.TrimPrefix(cardStr, npu800And9000CardPreName)
		klog.V(logDebugLev).Infof("GetNetworkUnhealthyCards after TrimPrefix %s.", v)
		cardInt, err := strconv.Atoi(v)
		if err != nil {
			klog.V(logErrorLev).Infof("GetNetworkUnhealthyCards conv failed %v.", err)
			return nil
		}

		topInt = append(topInt, cardInt)
	}

	return topInt
}

// GetDistributeUsableNPUTop Slice forehead difference set.
func GetDistributeUsableNPUTop(nodeNPUTopology, netUnhealthyCards []int) []int {
	var usableCards []int
	temp := map[int]struct{}{}

	for _, val := range netUnhealthyCards {
		temp[val] = struct{}{}
	}

	for _, nodeCard := range nodeNPUTopology {
		if _, ok := temp[nodeCard]; ok {
			continue
		}
		usableCards = append(usableCards, nodeCard)
	}

	return usableCards
}
