/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package base is using for HuaWei Ascend pin affinity schedule.

*/

package base

import (
	"errors"
	"fmt"

	"k8s.io/klog"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// GetUsableTopFromNode get node usable npu from annotation
func (tp *NPUHandler) GetUsableTopFromNode(node plugin.NPUNode) ([]int, error) {
	if tp == nil || len(node.Annotation) == 0 {
		return nil, errors.New(util.ArgumentError)
	}
	topStr, ok := node.Annotation[tp.GetAnnoName()]
	if !ok || len(topStr) == 0 {
		return nil, fmt.Errorf("getUsableTopFromNode node<%s> don't have npu<%s>", node.Name, tp.GetAnnoName())
	}

	nodeTop := util.ChangeTopToIntArray(topStr, tp.GetAnnoPreVal())
	if len(nodeTop) == 0 {
		return nil, fmt.Errorf("getUsableTopFromNode err: top string<%s> convert faild", topStr)
	}
	return nodeTop, nil
}

// GetCardNumGroupsFromTop get the chip for each card from nodeTop
func (tp *NPUHandler) GetCardNumGroupsFromTop(nodeNPUTopology []int) [][]int {
	if tp == nil || tp.MaxCardNPUNum == 0 {
		return nil
	}
	maxCardNum := 0
	for _, v := range nodeNPUTopology {
		maxCardNum = util.Max(maxCardNum, v)
	}
	cardNumGroups := make([][]int, maxCardNum/tp.MaxCardNPUNum+1)
	for _, v := range nodeNPUTopology {
		index := v / tp.MaxCardNPUNum
		if index > len(cardNumGroups)-1 {
			continue
		}
		cardNumGroups[index] = append(cardNumGroups[index], v)
	}
	return cardNumGroups
}

// UpdateNodeInfo update node info
func (tp *NPUHandler) UpdateNodeInfo(node plugin.NPUNode, usedTop []int) *plugin.NPUNode {
	if tp == nil || len(node.Annotation) == 0 || len(usedTop) == 0 {
		return nil
	}
	if len(usedTop) > tp.MaxNodeNPUNum {
		klog.V(util.LogErrorLev).Infof("%s UpdateNodeInfo err: used npu num<%d> is invalid",
			tp.GetPluginName(), len(usedTop))
		return nil
	}
	klog.V(util.LogDebugLev).Infof("%s before UpdateNodeInfo node<%s> Annotation: %#v",
		tp.GetPluginName(), node.Name, node.Annotation)
	healthyAnno, err := node.GetNewNPUNodeAnnotation(usedTop, tp.GetAnnoName(), tp.GetAnnoPreVal())
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s UpdateNodeInfo err: %s", tp.GetPluginName(), err.Error())
		return nil
	}
	node.Annotation[tp.GetAnnoName()] = healthyAnno
	klog.V(util.LogDebugLev).Infof("%s after UpdateNodeInfo node<%s> Annotation: %#v",
		tp.GetPluginName(), node.Name, node.Annotation)
	return &node
}
