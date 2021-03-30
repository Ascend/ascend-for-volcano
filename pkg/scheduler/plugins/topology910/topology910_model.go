/*
Copyright(C) 2020. Huawei Technologies Co.,Ltd. All rights reserved.

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

Package topology910 is using for HuaWei Ascend910 pin affinity schedule.

*/
package topology910

import (
	"errors"
	"k8s.io/klog"
	schedulerApi "k8s.io/kube-scheduler/extender/v1"
	"strconv"
	"volcano.sh/volcano/pkg/scheduler/api"
)

const (
	leftHccsName        = "leftHccs"
	rightHccsName       = "rightHccs"
	allHccsName         = "allHccs"
	nodeNoFitNPUWarning = "node no fit npu number"
)

type hccsInf struct {
	nodeInf  *api.NodeInfo
	name     string
	availNum int
	top      []int
}

type npuPriNodeInf struct {
	// the priority for NPU HCCS top
	Name     string
	nodeName string
}

type selectNodeInf struct {
	nodeName    string
	allNpuNum   int
	leftNpuNum  int
	rightNpuNum int
}

type initPriNodeGroupFn func(priNodeGroup map[string]*npuPriNodeInf, groupName string)

func initScoreMp(nodes []*api.NodeInfo, interPodAffinityScore schedulerApi.HostPriorityList) map[string]float64 {
	for _, node := range nodes {
		interPodAffinityScore = append(interPodAffinityScore, schedulerApi.HostPriority{
			Host:  node.Name,
			Score: 0,
		})
	}
	scoreMp := make(map[string]float64, len(interPodAffinityScore))
	for _, host := range interPodAffinityScore {
		scoreMp[host.Host] = float64(host.Score)
	}
	return scoreMp
}

// the police for one need card
func initOneCardPriNodeGroups(
	sNodeInf selectNodeInf,
	priNodeGroups []map[string]*npuPriNodeInf,
	addPriNodeGroupFn initPriNodeGroupFn) error {
	// priority:1>3>2>4
	if sNodeInf.leftNpuNum == magicNumInt1 || sNodeInf.rightNpuNum == magicNumInt1 {
		// A group
		addPriNodeGroupFn(priNodeGroups[0], "A")
		return nil
	}

	if sNodeInf.leftNpuNum == magicNumInt3 || sNodeInf.rightNpuNum == magicNumInt3 {
		// B group
		addPriNodeGroupFn(priNodeGroups[magicNumInt1], "B")
		return nil
	}

	if sNodeInf.leftNpuNum == magicNumInt2 || sNodeInf.rightNpuNum == magicNumInt2 {
		// C group
		addPriNodeGroupFn(priNodeGroups[magicNumInt2], "C")
		return nil
	}

	if sNodeInf.leftNpuNum == npuNumPerHccs || sNodeInf.rightNpuNum == npuNumPerHccs {
		// D group
		addPriNodeGroupFn(priNodeGroups[magicNumInt3], "D")
		return nil
	}

	// no satisfy HCCS,8*N need whole nodes
	klog.V(logErrorLev).Infof("%s initOneCardPriNodeGroups node(%s) (left :%d,right :%d) cannot fit",
		PluginName, sNodeInf.nodeName, sNodeInf.leftNpuNum, sNodeInf.rightNpuNum)
	return errors.New(nodeNoFitNPUWarning)
}

// the police for Two need card
func initTwoCardPriNodeGroups(
	sNodeInf selectNodeInf,
	priNodeGroups []map[string]*npuPriNodeInf,
	addPriNodeGroupFn initPriNodeGroupFn,
	taskModule string) error {

	if taskModule == cardAcceleratorType {
		if sNodeInf.allNpuNum == magicNumInt2 {
			addPriNodeGroupFn(priNodeGroups[0], "A")
			return nil
		}
		klog.V(logDebugLev).Infof("%s %s not have 2 npu in card module", PluginName, sNodeInf.nodeName)
		return errors.New(nodeNoFitNPUWarning)
	}

	// priority：2>npuNumPerHccs>3
	if sNodeInf.leftNpuNum == magicNumInt2 || sNodeInf.rightNpuNum == magicNumInt2 {
		// A group
		addPriNodeGroupFn(priNodeGroups[0], "A")
		return nil
	}

	if sNodeInf.leftNpuNum == npuNumPerHccs || sNodeInf.rightNpuNum == npuNumPerHccs {
		// B group
		addPriNodeGroupFn(priNodeGroups[magicNumInt1], "B")
		return nil
	}

	if sNodeInf.leftNpuNum == magicNumInt3 || sNodeInf.rightNpuNum == magicNumInt3 {
		// C group
		addPriNodeGroupFn(priNodeGroups[magicNumInt2], "C")
		return nil
	}

	// no satisfy HCCS,8*N need whole nodes
	klog.V(logErrorLev).Infof("%s initTwoCardPriNodeGroups node(%s) (left :%d,right :%d) cannot fit 2",
		PluginName, sNodeInf.nodeName, sNodeInf.leftNpuNum, sNodeInf.rightNpuNum)

	return errors.New(nodeNoFitNPUWarning)
}

// the police four one need card
func initFourCardPriNodeGroups(
	sNodeInf selectNodeInf,
	priNodeGroups []map[string]*npuPriNodeInf,
	addPriNodeGroupFn initPriNodeGroupFn) error {
	// only can allocate 4
	if sNodeInf.leftNpuNum == npuNumPerHccs || sNodeInf.rightNpuNum == npuNumPerHccs {
		// A group
		addPriNodeGroupFn(priNodeGroups[0], "A")
		return nil
	}
	// no satisfy HCCS,8*N need whole nodes
	klog.V(logErrorLev).Infof("%s initFoureCardPriNodeGroups node(%s) (left :%d,right :%d) cannot fit 4",
		PluginName, sNodeInf.nodeName, sNodeInf.leftNpuNum, sNodeInf.rightNpuNum)
	return errors.New(nodeNoFitNPUWarning)
}

// the police four one need card
func initEightCardPriNodeGroups(
	sNodeInf selectNodeInf,
	priNodeGroups []map[string]*npuPriNodeInf,
	addPriNodeGroupFn initPriNodeGroupFn) error {

	if sNodeInf.allNpuNum == nodeNpuNumber {
		addPriNodeGroupFn(priNodeGroups[0], "A")
		return nil
	}

	klog.V(logErrorLev).Infof("%s initEightCardPriNodeGroups node(%s) (all:%d) cannot fit 8",
		PluginName, sNodeInf.nodeName, sNodeInf.allNpuNum)
	return errors.New(nodeNoFitNPUWarning)
}

// insert into group by police
func insertNodeInPriGroup(
	taskReqNPU int,
	sNodeInf selectNodeInf,
	priNodeGroups []map[string]*npuPriNodeInf,
	addPriNodeGroupFn initPriNodeGroupFn,
	taskModule string) error {

	var err error
	switch taskReqNPU {
	case 0:
		klog.V(logInfoLev).Infof("%s initPriNodeGroups task npu is 0", PluginName)
	case magicNumInt1:
		err = initOneCardPriNodeGroups(sNodeInf, priNodeGroups, addPriNodeGroupFn)
	case magicNumInt2:
		err = initTwoCardPriNodeGroups(sNodeInf, priNodeGroups, addPriNodeGroupFn, taskModule)
	case npuNumPerHccs:
		err = initFourCardPriNodeGroups(sNodeInf, priNodeGroups, addPriNodeGroupFn)
	case nodeNpuNumber:
		err = initEightCardPriNodeGroups(sNodeInf, priNodeGroups, addPriNodeGroupFn)
	default:
		// For normal,can not be here. The pre function validate job has done this.
		klog.V(logErrorLev).Infof("%s node(%s) (left :%d,right :%d) cannot fit %d,illegal task npu number",
			PluginName, sNodeInf.nodeName, sNodeInf.leftNpuNum, sNodeInf.rightNpuNum, taskReqNPU)
		err = errors.New("illegal request npu number " + strconv.Itoa(taskReqNPU))
	}

	return err
}

func initSelectNodeInf(node *api.NodeInfo) selectNodeInf {
	var sNodeInf selectNodeInf
	var leftHccsTop []int
	var rightHccsTop []int

	cardIds := getTopFromNode(node)
	klog.V(logDebugLev).Infof("%s initPriNodeGroups:%v", PluginName, cardIds)
	for _, cardID := range cardIds {
		if cardID < npuNumPerHccs {
			leftHccsTop = append(leftHccsTop, cardID)
		} else {
			rightHccsTop = append(rightHccsTop, cardID)
		}
	}
	sNodeInf.leftNpuNum = len(leftHccsTop)
	sNodeInf.rightNpuNum = len(rightHccsTop)
	sNodeInf.allNpuNum = sNodeInf.leftNpuNum + sNodeInf.rightNpuNum

	return sNodeInf
}

// init 4 pri node-list group
func initPriNodeGroups(taskReqNPU int, nodes []*api.NodeInfo, taskModule string) ([]map[string]*npuPriNodeInf, error) {
	var err error
	var priNodeGroups []map[string]*npuPriNodeInf

	// for pipelined state the node npu is nil
	if len(nodes) == 0 {
		return nil, errors.New("nodes is empty")
	}

	for i := 0; i < npuNumPerHccs; i++ {
		priNodeGroups = append(priNodeGroups, make(map[string]*npuPriNodeInf, 1))
	}

	// init pri Node group
	for _, node := range nodes {
		sNodeInf := initSelectNodeInf(node)
		// set the meet node into its pri-node-list group
		addPriNodeGroupFn := func(priNodeGroup map[string]*npuPriNodeInf, groupName string) {
			klog.V(logDebugLev).Infof("%s nodeName:%s,group:%v", PluginName, node.Name, priNodeGroup[node.Name])
			priNodeGroup[node.Name] = &npuPriNodeInf{
				Name:     groupName,
				nodeName: node.Name,
			}
			klog.V(logDebugLev).Infof("%s addPriNodeGroupFn node name:%s priNode:%v",
				PluginName, node.Name, priNodeGroup[node.Name])
		}

		// insert into group by police
		err = insertNodeInPriGroup(taskReqNPU, sNodeInf, priNodeGroups, addPriNodeGroupFn, taskModule)
		if err != nil {
			continue
		}
	}
	return priNodeGroups, nil
}

// According to need card number,get best node from 4 pri-node-list-group
func getBestNodesMap(needCards int,
	priNodeGroups []map[string]*npuPriNodeInf,
	taskModule string) (map[string]int, error) {
	var bestNodesMap = make(map[string]int, magicNumInt2)

	for i := 0; i < npuNumPerHccs; i++ {
		for nodeName, _ := range priNodeGroups[i] {
			bestNodesMap[nodeName] = i
		}
	}

	if len(bestNodesMap) == 0 {
		return nil, errors.New("none bestNodes")
	}

	return bestNodesMap, nil
}

func setSelectTopValue(needCards int, selectHCCS *hccsInf) ([]int, error) {
	var setTop []int
	i := 0
	// select whole node
	if needCards == nodeNpuNumber {
		for i = 0; i < nodeNpuNumber; i++ {
			setTop = append(setTop, i)
		}
		return setTop, nil
	}

	i = 0
	for _, cardID := range selectHCCS.top {
		if i < needCards {
			setTop = append(setTop, cardID)
			i++
		} else {
			break
		}
	}
	return setTop, nil
}

// get the best node
// map[string]int string:nodeName, int:priority[0-3]
func getNpuAffinityBestNodes(taskReqNPU int, nodes []*api.NodeInfo, taskModule string) (map[string]int, error) {

	// 1. inti 4 pri node-list group
	priNodeGroups, err := initPriNodeGroups(taskReqNPU, nodes, taskModule)
	if err != nil {
		klog.V(logErrorLev).Infof("%s initPriNodeGroups failed :%s", PluginName, err)
		return nil, err
	}
	// 2.get the bestNodes map by taskReqNPU
	bestNodesMap, err := getBestNodesMap(taskReqNPU, priNodeGroups, taskModule)
	if err != nil {
		klog.V(logErrorLev).Infof("%s getBestNodesMap failed :%s", PluginName, err)
		return nil, err
	}

	klog.V(logInfoLev).Infof("%s getNpuAffinityBestNodes %v", PluginName, bestNodesMap)
	return bestNodesMap, nil
}

func checkResourceNumMeet(task *api.TaskInfo, nodeInfo *api.NodeInfo) error {
	if !task.InitResreq.LessEqual(nodeInfo.Idle) {
		return errors.New("IN Pipeline")
	}
	return nil
}

func checkResourceReady(task *api.TaskInfo, nodeInfo *api.NodeInfo) error {
	// check resource number fit
	err := checkResourceNumMeet(task, nodeInfo)
	if err != nil {
		return err
	}

	return nil
}

// write the topology to pod
// to change the top kind :int[] to device-plugin point kind:string link with ','
func setNpuTopologyToPod(task *api.TaskInfo, nodeInfo *api.NodeInfo, top []int) error {
	/*	err := checkResourceReady(task, nodeInfo)
		if err != nil {
			return err
		}*/

	return doSetPodNpuTopology(top, task)
}

// select the best nodes for task using, score=healthCores*4-priority
func scoreBestNpuNodes(scoreMp map[string]float64,
	bestNodes map[string]int,
	nodes []*api.NodeInfo) (map[string]float64, error) {
	var nodeWeight = 1.0

	for nodeName, priority := range bestNodes {
		healthNpuNumber, err := getNodeHealthNpuNumberByName(nodeName, nodes)
		if err != nil {
			scoreMp[nodeName] = 0.0
			klog.V(logInfoLev).Infof("%s getNodeHealthNpuNumberByName error:%v", PluginName, err)
			continue
		}

		scoreMp[nodeName] = nodeWeight * (healthNpuNumber*4 - float64(priority))
	}

	return scoreMp, nil
}

// open-session register for batch Node Order fun
// for volcano frame,can not return error
func batchNodeOrderFn(task *api.TaskInfo, nodes []*api.NodeInfo) (map[string]float64, error) {
	var interPodAffinityScore schedulerApi.HostPriorityList
	var bestNodes map[string]int
	var errGet error

	klog.V(logInfoLev).Infof("%s Enter batchNodeOrderFn", PluginName)
	defer klog.V(logInfoLev).Infof("%s leaving batchNodeOrderFn", PluginName)

	// init score-map
	scoreMp := initScoreMp(nodes, interPodAffinityScore)

	// 1.Get the task's NPU request
	taskReqNPU, errGet := getTaskNpuNum(task)
	if errGet != nil {
		// cannot return error for task is no npu kind possible.
		klog.V(logErrorLev).Infof("%s batchNodeOrderFn task :%s,%v", PluginName, task.Name, errGet)
		// cannot return nil，will panic
		return scoreMp, nil
	}
	klog.V(logInfoLev).Infof("%s batchNodeOrderFn Get %s taskReqNPU:%v", PluginName, task.Name, taskReqNPU)

	taskModule := getTaskModule(task)
	// 2.get the best node and top by A,B,C,D rules and require numbers.
	bestNodes, errGet = getNpuAffinityBestNodes(taskReqNPU, nodes, taskModule)
	if errGet != nil || len(bestNodes) == 0 {
		// get suitable node failed
		klog.V(logErrorLev).Infof("%s batchNodeOrderFn task[%s],require npu:%d, failed[%v]",
			PluginName, task.Name, taskReqNPU, errGet)
		return scoreMp, nil
	}
	klog.V(logInfoLev).Infof("%s batchNodeOrderFn Get task for NPU number:%d, %+v",
		PluginName, taskReqNPU, bestNodes)

	// 3.scored the nodes and set top
	scoreMp, errGet = scoreBestNpuNodes(scoreMp, bestNodes, nodes)
	if errGet != nil {
		// get suitable node failed
		klog.V(logErrorLev).Infof("%s scoreBestNpuNodes get err:%v", PluginName, errGet)
		return scoreMp, errGet
	}

	klog.V(logInfoLev).Infof("%s Total Score for task %s/%s is: %v", PluginName,
		task.Namespace, task.Name, scoreMp)
	return scoreMp, nil
}
