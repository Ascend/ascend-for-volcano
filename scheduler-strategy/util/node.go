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

Package util is using for HuaWei Ascend9 pin affinity schedule utilities.

*/
package util

import (
	"errors"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog"
	"sort"
	"volcano.sh/volcano/pkg/scheduler/api"
)

const (
	npuNumberUpperBoundary = 7
	npuNumberLowerBoundary = 0
)

// GetNPUAllocCardsFromNodeAnnotation Get node distributable npu card ids.
func GetNPUAllocCardsFromNodeAnnotation(node *api.NodeInfo, npuCardName string) (string, error) {
	topStr1, ok := node.Node.Annotations[npuCardName]
	if !ok {
		msg := fmt.Errorf("%s npu card nil", node.Name)
		klog.V(logDebugLev).Infof("%s :%v :%v", npuCardName, msg.Error(), node.Node.Annotations[npuCardName])
		return "", msg
	}
	return topStr1, nil
}

// GetTopFromNodeOthers Get npu card ids from node other.
func GetTopFromNodeOthers(node *api.NodeInfo, npuCardName string, npuCardPreName string) []int {
	mapStr, err := GetNPUAllocCardsFromNodeOthers(node, npuCardName)
	if err != nil {
		klog.V(logErrorLev).Infof("%s GetNPUAllocCardsFromNodeOthers top nil:%v.", npuCardName, node.Name)
		return nil
	}

	tmpDeviceIDs := ChangeTopToIntArray(mapStr, npuCardPreName)
	if tmpDeviceIDs == nil {
		klog.V(logErrorLev).Infof("%s ChangeTopToIntArray to int failed.", npuCardName)
		return nil
	}

	if !CheckTopValidity(tmpDeviceIDs) {
		klog.V(logErrorLev).Infof("%s CheckTopValidity %s got invalid top", npuCardName, node.Name)
		return nil
	}

	klog.V(logDebugLev).Infof("%s GetTopFromNodeOthers int: %v, s: %s.", npuCardName, tmpDeviceIDs, mapStr)

	return tmpDeviceIDs
}

// GetNPUAllocCardsFromNodeOthers Get node distributable npu card ids from node other.
func GetNPUAllocCardsFromNodeOthers(node *api.NodeInfo, npuCardName string) (string, error) {
	valueTmp, ok := node.Others[npuCardName]
	if !ok {
		msg := fmt.Errorf("%s npu card nil", node.Name)
		klog.V(logErrorLev).Infof("%s :%v :%v", npuCardName, msg.Error(), node.Others[npuCardName])
		return "", msg
	}

	mapStr, ok := valueTmp.(string)
	if !ok {
		msg := fmt.Errorf("type is not match")
		klog.V(logErrorLev).Infof("%s :%v :%v.", npuCardName, valueTmp, msg.Error())
		return "", msg
	}

	return mapStr, nil
}

// CheckTopValidity Checks the validity of npu card ids which are read from annotations
func CheckTopValidity(top []int) bool {
	tmp := make([]int, len(top))
	for index, v := range top {
		tmp[index] = v
	}
	sort.Ints(tmp)
	for i, num := range tmp {
		if i != 0 && num == tmp[i-1] {
			klog.V(logErrorLev).Infof("duplicated npu(%d)", num)
			return false
		}

		if num < npuNumberLowerBoundary || num > npuNumberUpperBoundary {
			klog.V(logErrorLev).Infof("got npu number(%d) out of range [%d, %d]",
				num, npuNumberLowerBoundary, npuNumberUpperBoundary)
			return false
		}
	}

	return true
}

// GetNodeNPUNumFromIdle Get npu top like（int[]） from node idle.
func GetNodeNPUNumFromIdle(nodeInfo *api.NodeInfo, npuCardName string) (int, error) {
	nodeNPUIdleNumFromIdle, ok := nodeInfo.Idle.ScalarResources[v1.ResourceName(npuCardName)]
	if !ok {
		klog.V(logErrorLev).Infof("%s getNodeNPUNumFromIdle failed.", npuCardName)
		return 0, errors.New("get node idle npu failed")
	}

	return int(nodeNPUIdleNumFromIdle / npuHex), nil
}

// GetNodeHealthNPUNumberByName Get health npu number from node.
func GetNodeHealthNPUNumberByName(nodeName string, nodes []*api.NodeInfo, npuCardName string) (float64, error) {
	for _, nodeInf := range nodes {
		if nodeInf == nil {
			continue
		}

		if nodeName == nodeInf.Name {
			nodeNPUIdleNumFromIdle, ok := nodeInf.Allocatable.ScalarResources[v1.ResourceName(npuCardName)]
			if !ok {
				klog.V(logErrorLev).Infof("%s getNodeNPUNumFromIdle failed.", npuCardName)
				return 0, errors.New("get node capability npu failed")
			}
			return nodeNPUIdleNumFromIdle / npuHex, nil
		}
	}
	return 0, errors.New("no found")
}

// GetRealTopAfterAlloc Get npu card ids after alloc.
func GetRealTopAfterAlloc(nodeDeviceIDs []int, useDeviceIDs []int, npuCardName string) string {
	var tmpDeviceIDs []int
	var existFlag bool

	for _, nTopI := range nodeDeviceIDs {
		existFlag = false
		for _, tTopI := range useDeviceIDs {
			if nTopI == tTopI {
				existFlag = true
				break
			}
		}

		if !existFlag {
			tmpDeviceIDs = append(tmpDeviceIDs, nTopI)
		}
	}
	klog.V(logDebugLev).Infof("%s getRealTopAfterAlloc ：%v .", npuCardName, tmpDeviceIDs)
	// change int to string
	return ChangeIntArrToStr(tmpDeviceIDs, npuCardName)
}

// ReloadNewTopToNodeOther Set node npu card ids by new one.
func ReloadNewTopToNodeOther(node *api.NodeInfo, newNodeTopStr string, npuCardName string) error {
	if _, ok := node.Others[npuCardName]; !ok {
		return fmt.Errorf("%s node.Others nil", node.Name)
	}

	node.Others[npuCardName] = newNodeTopStr

	return nil
}

// GetNodeSelector Get node selector.
func GetNodeSelector(node *api.NodeInfo) (map[string]string, error) {
	_, ok := node.Node.Labels[archSelector]
	if !ok {
		return nil, errors.New("selector is nil")
	}

	return node.Node.Labels, nil
}

// IsCardModeNode Judge the node mode:card or not
func IsCardModeNode(node *api.NodeInfo) bool {
	nodeSelectors, errNode := GetNodeSelector(node)
	if errNode != nil {
		klog.V(logErrorLev).Infof("node(%s) %v.", node.Name, errNode)
		return false
	}

	acceleratorValue, ok := nodeSelectors[acceleratorType]
	if !ok {
		// no acceleratorType means module
		klog.V(logDebugLev).Infof("node(%s) is module type.", node.Name)
		return false
	}

	if acceleratorValue == cardAcceleratorType {
		klog.V(logDebugLev).Infof("node(%s) is card type.", node.Name)
		return true
	}

	klog.V(logDebugLev).Infof("node(%s) is module type.", node.Name)
	return false
}

// GetNodeHccsCardNum Get node npu card number according by hccs.
func GetNodeHccsCardNum(nodeTop []int) (int, int) {
	var leftHccsArray []int
	var rightHccsArray []int

	for _, v := range nodeTop {
		if v < npuNumPerHccs {
			leftHccsArray = append(leftHccsArray, v)
			continue
		}
		rightHccsArray = append(rightHccsArray, v)
	}

	return len(leftHccsArray), len(rightHccsArray)
}

// GetNPUTopFromHccs Get npu card ids from hccs.
func GetNPUTopFromHccs(taskNPUNumber int, allocTopologyHccl []int) ([]int, error) {
	var allocTopologyNPUs []int
	var npuNum = taskNPUNumber
	var err error
	err = nil

	for _, npu := range allocTopologyHccl {
		if npuNum > 0 {
			npuNum = npuNum - 1
			allocTopologyNPUs = append(allocTopologyNPUs, npu)
		}
	}

	if npuNum != 0 {
		err = fmt.Errorf("%v-%v not meet request number[%d]", allocTopologyHccl, allocTopologyNPUs, taskNPUNumber)
		klog.V(logErrorLev).Infof("%v.", err)
		return nil, err
	}

	return allocTopologyNPUs, nil
}

// AddScoreByFaultNPUTask returns nil to indicate that it is selected, and anything else to indicate that it is not.
func AddScoreByFaultNPUTask(task *api.TaskInfo, scoreMap map[string]float64) (map[string]float64, error) {
	tmpValue, ok := ReSchedulerJobs[task.Job]
	if !ok {
		klog.V(logInfoLev).Infof("not fault %s,no need add score.", task.Name)
		return scoreMap, nil
	}

	if len(scoreMap) == 0 {
		mgs := fmt.Errorf("AddScoreByFaultNPUTask scoreMap is nil")
		klog.V(logInfoLev).Infof("%v.", mgs)
		return scoreMap, mgs
	}

	klog.V(logInfoLev).Infof("AddScoreByFaultNPUTask get %v.", tmpValue)
	for taskName, nodeName := range tmpValue.NodeNames {
		if taskName != task.Name {
			continue
		}
		if _, ok := scoreMap[nodeName]; !ok {
			return scoreMap, nil
		}
		scoreMap[nodeName] += nodeNPUNumber * nodeNPUNumber
	}

	return scoreMap, nil
}
