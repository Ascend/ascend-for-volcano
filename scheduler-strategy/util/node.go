/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package util is using for HuaWei Ascend9 pin affinity schedule utilities.

*/
package util

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"k8s.io/api/core/v1"
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
)

// GetNodeIdleNPUNum Get node npu idle number
func GetNodeIdleNPUNum(node *api.NodeInfo, npuCardName string) (int, error) {
	nodeNPUIdleNumber, ok := node.Idle.ScalarResources[v1.ResourceName(npuCardName)]
	if !ok {
		return 0, errors.New("not npu node")
	}

	nodeNPU := int(nodeNPUIdleNumber / NPUHex)
	return nodeNPU, nil
}

// GetNodeAvailNPUIdsFromAnno Get node npu available number
func GetNodeAvailNPUIdsFromAnno(node *api.NodeInfo, npuCardName string) (map[int]struct{}, error) {
	idsMap := make(map[int]struct{}, ConstIntNum8)
	chipStr, getErr := GetNPUAllocCardsFromNodeOthers(node, npuCardName)
	if getErr != nil || chipStr == "" {
		klog.V(LogDebugLev).Infof("GetNodeAvailNPUIdsFromAnno %s %s %v", node.Name, npuCardName, getErr)
		return idsMap, nil
	}
	chipSlice := strings.Split(chipStr, ",")
	for _, chip := range chipSlice {
		// chip like Ascend710-4
		cardSlice := strings.Split(chip, "-")
		cardStr := cardSlice[len(cardSlice)-1]
		cardID, covErr := strconv.Atoi(cardStr)
		if covErr != nil {
			return nil, covErr
		}
		idsMap[cardID] = struct{}{}
	}
	return idsMap, nil
}

// GetTopFromNodeOthers Get npu card ids like（int[]） from node info.
func GetTopFromNodeOthers(node *api.NodeInfo, npuCardName string, npuCardPreName string) []int {
	topStr, err := GetNPUAllocCardsFromNodeOthers(node, npuCardName)
	if err != nil {
		klog.V(LogErrorLev).Infof("%s getTopFromNode top nil:%v.", npuCardName, node.Name)
		return nil
	}

	// cannot judge len(topInt) is 0, for pipelined state
	tmpDeviceIDs := ChangeTopToIntArray(topStr, npuCardPreName)
	if tmpDeviceIDs == nil {
		klog.V(LogInfoLev).Infof("%s getTopFromNode %s nil(%s).", npuCardName, node.Name, topStr)
		return nil
	}

	klog.V(LogDebugLev).Infof("%s getTopFromNode int: %v, s: %s.", npuCardName, tmpDeviceIDs, topStr)
	return tmpDeviceIDs
}

// GetNodeNPUNumFromIdle Get npu top like（int[]） from node idle.
func GetNodeNPUNumFromIdle(nodeInfo *api.NodeInfo, npuCardName string) (int, error) {
	nodeNPUIdleNumFromIdle, ok := nodeInfo.Idle.ScalarResources[v1.ResourceName(npuCardName)]
	if !ok {
		klog.V(LogErrorLev).Infof("%s getNodeNPUNumFromIdle failed.", npuCardName)
		return 0, errors.New("get node idle npu failed")
	}

	return int(nodeNPUIdleNumFromIdle / NPUHex), nil
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
				klog.V(LogErrorLev).Infof("%s getNodeNPUNumFromIdle failed.", npuCardName)
				return 0, errors.New("get node capability npu failed")
			}
			return nodeNPUIdleNumFromIdle / NPUHex, nil
		}
	}
	return 0, errors.New("no found")
}

// GetRealTopAfterAlloc Get npu card ids after alloc.
func GetRealTopAfterAlloc(nodeDeviceIDs []int, useDeviceIDs []int, npuCardName string) string {
	if len(nodeDeviceIDs) == 0 {
		return ""
	}

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
	klog.V(LogDebugLev).Infof("%s GetRealTopAfterAlloc : %v .", npuCardName, tmpDeviceIDs)
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
	_, ok := node.Node.Labels[ArchSelector]
	if !ok {
		return nil, errors.New("selector is nil")
	}

	return node.Node.Labels, nil
}

// GetNPUAllocCardsFromNodeOthers Get node distributable npu card ids.
func GetNPUAllocCardsFromNodeOthers(node *api.NodeInfo, npuCardName string) (string, error) {
	valueTmp, ok := node.Others[npuCardName]
	if !ok {
		msg := fmt.Errorf("%s npu card %s nil", node.Name, npuCardName)
		klog.V(LogDebugLev).Infof("GetNPUAllocCardsFromNodeOthers :%v.", msg.Error())
		return "", msg
	}

	mapStr, mapOk := valueTmp.(string)
	if !mapOk {
		msg := fmt.Errorf("type is not match")
		klog.V(LogErrorLev).Infof("%s :%v :%v.", npuCardName, valueTmp, msg.Error())
		return "", msg
	}
	return mapStr, nil
}

// GetNPUAllocCardsFromNodeAnnotations Get node distributable npu card ids.
func GetNPUAllocCardsFromNodeAnnotations(node *api.NodeInfo, npuCardName string) (string, error) {
	valueTmp, ok := node.Node.Annotations[npuCardName]
	if !ok {
		msg := fmt.Errorf("%s npu card nil", node.Name)
		klog.V(LogDebugLev).Infof("%s :%v.", npuCardName, msg.Error())
		return "", msg
	}

	return valueTmp, nil
}

// IsCardModeNode Judge the node mode:card or not
func IsCardModeNode(node *api.NodeInfo) bool {
	nodeSelectors, errNode := GetNodeSelector(node)
	if errNode != nil {
		klog.V(LogErrorLev).Infof("node(%s) %v.", node.Name, errNode)
		return false
	}

	acceleratorValue, ok := nodeSelectors[AcceleratorType]
	if !ok {
		// no AcceleratorType means module
		klog.V(LogDebugLev).Infof("node(%s) is module type.", node.Name)
		return false
	}

	if acceleratorValue == CardAcceleratorType {
		klog.V(LogDebugLev).Infof("node(%s) is card type.", node.Name)
		return true
	}

	klog.V(LogDebugLev).Infof("node(%s) is module type.", node.Name)
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
		klog.V(LogErrorLev).Infof("%v.", err)
		return nil, err
	}

	return allocTopologyNPUs, nil
}

// IsNPUNNode to judge the node wither has NPU card or not.
func IsNPUNNode(tmpNode *api.NodeInfo) error {
	for key, value := range tmpNode.Node.Annotations {
		if strings.Contains(key, CommCardPreName) {
			tmp := strings.Split(value, ",")
			if len(tmp) != 0 {
				return nil
			}
			continue
		}
	}
	return fmt.Errorf("%s has no %s", tmpNode.Name, CommCardPreName)
}

// GetReqResourceNameFromNode get node has npu name
func GetReqResourceNameFromNode(tmpNode *api.NodeInfo) (string, error) {
	if tmpNode == nil {
		return "", errors.New("nil parameter")
	}
	for k, num := range tmpNode.Allocatable.ScalarResources {
		temp := string(k)
		// must contains "huawei.com/Ascend"
		if !strings.Contains(temp, CommCardPreName) {
			continue
		}
		if num != 0 {
			return temp, nil
		}
	}
	klog.V(LogErrorLev).Infof("GetReqResourceNameFromNode %+v.", tmpNode.Allocatable.ScalarResources)
	return "", errors.New("nil NPU")
}
