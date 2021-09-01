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
	npuNumberUpperBoundaryOf310 = 63
	npuNumberLowerBoundary      = 0
)

// GetTopFromNode Get npu card ids like（int[]） from node info.
func GetTopFromNode(node *api.NodeInfo, npuCardName string, npuCardPreName string) []int {
	var topInt []int

	topStr, err := GetNodeNPUAllocCards(node, npuCardName)
	if err != nil {
		klog.V(logErrorLev).Infof("%s getTopFromNode top nil:%v.", npuCardName, node.Name)
		return nil
	}

	// cannot judge len(topInt) is 0, for pipelined state
	topInt = ChangeTopToIntArray(topStr, npuCardPreName)
	if topInt == nil {
		klog.V(logInfoLev).Infof("%s getTopFromNode %s nil(%s).", npuCardName, node.Name, topStr)
		return nil
	}

	if !CheckTopValidity(topInt) {
		klog.V(logErrorLev).Infof("%s getTopFromNode %s got invalid top", npuCardName, node.Name)
		return nil
	}

	klog.V(logDebugLev).Infof("%s getTopFromNode int: %v, s: %s.", npuCardName, topInt, topStr)
	return topInt
}

// CheckTopValidity Checks the validity of npu card ids which are read from annotations
func CheckTopValidity(top []int) bool {
	sort.Ints(top) // sort ascend
	for i, num := range top {
		if i != 0 && num == top[i-1] {
			klog.V(logErrorLev).Infof("duplicated npu(%d)", num)
			return false
		}
		// npuNumberUpperBoundary在310中应该为3
		if num < npuNumberLowerBoundary || num > npuNumberUpperBoundaryOf310 {
			klog.V(logErrorLev).Infof("got npu number(%d) out of range [%d, %d]",
				num, npuNumberLowerBoundary, npuNumberUpperBoundaryOf310)
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

func getNodeNPUStrFromOther(mapInter map[string]interface{}, npuCardName string) string {
	valueTmp, ok := mapInter[npuCardName]
	if !ok {
		klog.V(logErrorLev).Infof("%s getNodeNPUStrFromOther nil.", npuCardName)
		return ""
	}

	mapStr, ok := valueTmp.(string)
	if !ok {
		klog.V(logErrorLev).Infof("%s getNodeNPUStrFromOther no string type.", npuCardName)
		return ""
	}

	return mapStr
}

// GetDeviceIDsFromNodeOther Get npu card ids from node other.
func GetDeviceIDsFromNodeOther(mapInter map[string]interface{}, npuCardName string, npuCardPreName string) []int {
	mapStr := getNodeNPUStrFromOther(mapInter, npuCardName)
	if mapStr == "" {
		klog.V(logErrorLev).Infof("%s getNodeNPUStrFromOther other nil.", npuCardName)
		return nil
	}

	tmpDeviceIDs := ChangeTopToIntArray(mapStr, npuCardPreName)
	if tmpDeviceIDs == nil {
		klog.V(logErrorLev).Infof("%s getDeviceIDsFromAnnotations to int failed.", npuCardName)
		return nil
	}

	return tmpDeviceIDs
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

// GetNodeNPUAllocCards Get node distributable npu card ids.
func GetNodeNPUAllocCards(node *api.NodeInfo, npuCardName string) (string, error) {
	topStr, ok := node.Others[npuCardName]
	if ok {
		resStr, ok := topStr.(string)
		if !ok {
			msg := fmt.Errorf("type not match")
			klog.V(logDebugLev).Infof("%s :%v :%v", npuCardName, msg.Error(), node.Others[npuCardName])
			return "", msg
		}
		return resStr, nil
	}
	topStr1, ok := node.Node.Annotations[npuCardName]
	if !ok {
		msg := fmt.Errorf("%s npu card nil", node.Name)
		klog.V(logDebugLev).Infof("%s :%v :%v", npuCardName, msg.Error(), node.Others[npuCardName])
		return "", msg
	}
	return topStr1, nil
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
