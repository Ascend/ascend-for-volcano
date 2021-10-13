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

Package module910x8 is using for HuaWei A800/9000 Ascend910 pin affinity schedule.

*/
package module910x8

import (
	"errors"
	"fmt"
	"k8s.io/klog"
	"reflect"
	"volcano.sh/volcano/pkg/scheduler/api"
	vapi "volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/rescheduling"
	hwutil "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/util"
)

type selectNodeInf struct {
	nodeName    string
	allNPUNum   int
	leftNPUNum  int
	rightNPUNum int
}

func initSelectNodeInf(node *api.NodeInfo, disFlag bool) selectNodeInf {
	var sNodeInf selectNodeInf
	var leftHccsTop []int
	var rightHccsTop []int

	cardIds := getUsableTopFromNode(node, disFlag)
	klog.V(logDebugLev).Infof("%s initPriNodeGroups:%v.", PluginName, cardIds)
	for _, cardID := range cardIds {
		if cardID < npuNumPerHccs {
			leftHccsTop = append(leftHccsTop, cardID)
		} else {
			rightHccsTop = append(rightHccsTop, cardID)
		}
	}
	sNodeInf.leftNPUNum = len(leftHccsTop)
	sNodeInf.rightNPUNum = len(rightHccsTop)
	sNodeInf.allNPUNum = sNodeInf.leftNPUNum + sNodeInf.rightNPUNum

	return sNodeInf
}

// Initializes the priority group of the node.
func initPriNodeGroups(task *api.TaskInfo, nodes []*api.NodeInfo, disFlag bool) ([]map[string]*npuPriNodeInf, error) {
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
		if reflect.ValueOf(node).IsNil() {
			continue
		}

		sNodeInf := initSelectNodeInf(node, disFlag)
		// set the meet node into its pri-node-list group
		addPriNodeGroupFn := func(priNodeGroup map[string]*npuPriNodeInf, groupName string) {
			klog.V(logDebugLev).Infof("%s nodeName:%s,group:%v.", PluginName, node.Name, priNodeGroup[node.Name])
			priNodeGroup[node.Name] = &npuPriNodeInf{
				Name:     groupName,
				nodeName: node.Name,
			}
			klog.V(logDebugLev).Infof("%s addPriNodeGroupFn node name:%s priNode:%v.",
				PluginName, node.Name, priNodeGroup[node.Name])
		}

		// insert into group by policy
		err = insertNodeInPriGroup(task, sNodeInf, priNodeGroups, addPriNodeGroupFn)
		if err != nil {
			continue
		}
	}
	return priNodeGroups, nil
}

func getNodeHccsArray(nodeTop []int) ([]int, []int) {
	var leftHccsArray []int
	var rightHccsArray []int

	for _, v := range nodeTop {
		if v < npuNumPerHccs {
			leftHccsArray = append(leftHccsArray, v)
			continue
		}
		rightHccsArray = append(rightHccsArray, v)
	}

	return leftHccsArray, rightHccsArray
}

func getNodeNPUNumFromOthers(nodeInfo *api.NodeInfo) (int, error) {
	top := hwutil.GetTopFromNodeOthers(nodeInfo, npu800And9000CardName, npu910CardPreName)
	if top == nil {
		return 0, fmt.Errorf("nil node(%s) top", nodeInfo.Name)
	}

	nodeNPUIdleNumFromTop := len(top)
	if nodeNPUIdleNumFromTop > maxNPUNum {
		return 0, fmt.Errorf("amount of npus exceeded the limitation, maximum(%d), actual(%d)",
			maxNPUNum, nodeNPUIdleNumFromTop)
	}

	return nodeNPUIdleNumFromTop, nil
}

func initNodesNPUTopologyFn(nodes map[string]*api.NodeInfo) error {
	for key := range nodes {
		if hwutil.IsCardModeNode(nodes[key]) {
			continue
		}

		topStr, err := hwutil.GetNPUAllocCardsFromNodeAnnotations(nodes[key], npu800And9000CardName)
		if err != nil {
			klog.V(logDebugLev).Infof("%s initNodesFn :%v.", PluginName, err)
			return nil
		}

		err = hwutil.SaveTopologyInMap(nodes[key].Others, topStr, npu800And9000CardName)
		if err != nil {
			return err
		}
	}
	return nil
}

func checkNPUResourceStable(node *vapi.NodeInfo) error {
	// default is the npu task
	nodeNPUIdleNumFromTop, err := getNodeNPUNumFromOthers(node)
	if err != nil {
		return fmt.Errorf("getNodeNPUNumFromOthers %s : %s", nodesNoMeetNPUReqError, err)
	}

	nodeNPUIdleNumFromIdle, err := hwutil.GetNodeNPUNumFromIdle(node, npu800And9000CardName)
	if err != nil {
		return fmt.Errorf("getNodeNPUNumFromIdle %s : %s", nodesNoMeetNPUReqError, err)
	}

	if err = hwutil.CheckNodeNPUStabilize(nodeNPUIdleNumFromTop, nodeNPUIdleNumFromIdle); err != nil {
		return fmt.Errorf("%s : %s", nodeNotStableWarning, err)
	}

	return nil
}

// Pre-select cluster processing.
// The reason for not considering nodes passed in by the current framework is
// fault node
func clusterNodePredicateFn(task *api.TaskInfo, ssn *framework.Session) error {
	klog.V(logDebugLev).Infof("%s enter clusterNodePredicateFn.", PluginName)
	defer klog.V(logDebugLev).Infof("%s leave clusterNodePredicateFn.", PluginName)

	// 1.Determine if it is a 910 jobs.
	if err := isMyTask(task); err != nil {
		klog.V(logDebugLev).Infof("%s get910x8Jobs: %v.", PluginName, err)
		return nil
	}
	// 2.Determine if it is a NPUFault jobs.
	if !rescheduling.IsNPUFaultTask(task) {
		klog.V(logDebugLev).Infof("%s %s is not npu fault job.", PluginName, task.Name)
		return nil
	}
	// 3.Get the task uses node.
	node, err := rescheduling.GetFaultTaskUseNodeInfo(task, ssn)
	if err != nil {
		klog.V(logErrorLev).Infof("%s %s %v.", PluginName, task.Name, err)
		return nil
	}
	// 4.check node NPU Resource Stable
	stableErr := checkNPUResourceStable(node)
	if stableErr == nil {
		klog.V(logDebugLev).Infof("%s %s NPU Resource Stable.", PluginName, node.Name)
		return nil
	}
	klog.V(logInfoLev).Infof("%s %s %v.", PluginName, node.Name, stableErr)
	// 5.Instability requires a decision on whether to continue to wait this node..
	if isUnstableNodeMeetTaskReqNPUSource(task, node) {
		return fmt.Errorf("%s is meet npu fault task %s, need continue using this node", node.Name, task.Name)
	}
	klog.V(logInfoLev).Infof("%s not meet %s req.", PluginName, node.Name, task.Name)

	return nil
}

func isMyNode(node *vapi.NodeInfo) error {
	_, err := hwutil.GetNPUAllocCardsFromNodeOthers(node, npu800And9000CardName)
	if err != nil {
		return fmt.Errorf("%s %s", node.Name, nodeNoFitNPUWarning)
	}

	if hwutil.IsCardModeNode(node) {
		return fmt.Errorf("%s is card mode", node.Name)
	}

	return nil
}

func getUsableTopFromNode(node *api.NodeInfo, distributeFlag bool) []int {
	nodeNPUTopology := hwutil.GetTopFromNodeOthers(node, npu800And9000CardName, npu910CardPreName)
	if !distributeFlag {
		return nodeNPUTopology
	}
	// Network unhealthy Cards affect only distribution job.
	netUnhealthyCards := rescheduling.GetNetworkUnhealthyCards(node)
	if netUnhealthyCards == nil {
		return nodeNPUTopology
	}

	return rescheduling.GetDistributeUsableNPUTop(nodeNPUTopology, netUnhealthyCards)
}
