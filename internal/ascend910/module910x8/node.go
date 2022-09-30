/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package module910x8 is using for HuaWei Ascend pin affinity schedule.

*/
package module910x8

import (
	"fmt"

	"k8s.io/klog"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

func (tp *module910x8) getUsableTopFromNode(node plugin.NPUNode, disFlag bool) ([]int, error) {
	var resTop []int
	topStr, ok := node.Annotation[tp.GetAnnoName()]
	if !ok {
		err := fmt.Errorf("node<%s> don't have npu<%s>", node.Name, tp.GetAnnoName())
		klog.V(util.LogWarningLev).Infof("%s getUsableTopFromNode err: %s", tp.GetPluginName(), err.Error())
		return nil, err
	}
	nodeTop := util.ChangeTopToIntArray(topStr, tp.GetAnnoPreVal())
	if len(nodeTop) > tp.MaxNodeNPUNum {
		err := fmt.Errorf("node<%s> npu top<%v> is invalid", node.Name, nodeTop)
		klog.V(util.LogWarningLev).Infof("%s getUsableTopFromNode err: %s", tp.GetPluginName(), err.Error())
		return nil, err
	}
	if len(nodeTop) != 0 {
		resTop = append(resTop, nodeTop...)
	}
	if !disFlag {
		return resTop, nil
	}

	networkUnhealthyTopStr, ok := node.Annotation[tp.netUnhealthyKey]
	if !ok {
		err := fmt.Errorf("node<%s> don't have resource<%s>", node.Name, tp.netUnhealthyKey)
		klog.V(util.LogWarningLev).Infof("%s getUsableTopFromNode err: %s", tp.GetPluginName(), err.Error())
		return nil, err
	}
	networkUnhealthyTop := util.ChangeTopToIntArray(networkUnhealthyTopStr, tp.GetAnnoPreVal())
	if len(nodeTop) > tp.MaxNodeNPUNum {
		err := fmt.Errorf("node<%s> npu networkUnhealthy top<%v> is invalid", node.Name, networkUnhealthyTop)
		klog.V(util.LogWarningLev).Infof("%s getUsableTopFromNode err: %s", tp.GetPluginName(), err.Error())
		return nil, err
	}
	if len(networkUnhealthyTop) == 0 {
		return resTop, nil
	}
	var newTop []int
	for _, rId := range resTop {
		existFlag := false
		for _, nId := range networkUnhealthyTop {
			if rId == nId {
				existFlag = true
				break
			}
		}
		if !existFlag {
			newTop = append(newTop, rId)
		}
	}
	return newTop, nil
}

func initSelectNodeInf(npuTop []int) selectNodeInf {
	var sNodeInf selectNodeInf
	var leftHccsTop []int
	var rightHccsTop []int

	for _, cardID := range npuTop {
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

func (tp *module910x8) getNodeBestScore(taskNPUNum int, npuTop []int) (int, error) {
	var bestScore = affScore4

	sNodeInf := initSelectNodeInf(npuTop)
	if sNodeInf.allNPUNum < 1 ||
		sNodeInf.allNPUNum > tp.MaxNodeNPUNum ||
		sNodeInf.rightNPUNum > npuNumPerHccs ||
		sNodeInf.leftNPUNum > npuNumPerHccs {
		return bestScore, fmt.Errorf("node top<%v> is invalid", npuTop)
	}

	var err = fmt.Errorf("node top<%v> is not meet task req npu<%d>", npuTop, taskNPUNum)
	if taskNPUNum == nodeNPUNumber {
		if len(npuTop) == nodeNPUNumber {
			return 0, nil
		}
		return bestScore, err
	}
	if taskNPUNum < 1 || taskNPUNum > npuNumPerHccs {
		return bestScore, fmt.Errorf("task req npu num<%d> is invalid", taskNPUNum)
	}
	switch {
	case sNodeInf.rightNPUNum == 0:
		bestScore = tp.affScoreList[taskNPUNum-1][sNodeInf.leftNPUNum-1]
	case sNodeInf.leftNPUNum == 0:
		bestScore = tp.affScoreList[taskNPUNum-1][sNodeInf.rightNPUNum-1]
	default:
		bestScore = util.Min(tp.affScoreList[taskNPUNum-1][sNodeInf.rightNPUNum-1],
			tp.affScoreList[taskNPUNum-1][sNodeInf.leftNPUNum-1])
	}
	if bestScore == affScore4 {
		return bestScore, err
	}
	return bestScore, nil
}

// UpdateNodeInfo update node info
func (tp *module910x8) UpdateNodeInfo(node plugin.NPUNode, usedTop []int) *plugin.NPUNode {
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
	netUnhealthyAnno, err := node.GetNewNPUNodeAnnotation(usedTop, tp.netUnhealthyKey, tp.GetAnnoPreVal())
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s UpdateNodeInfo err: %s", tp.GetPluginName(), err.Error())
		return nil
	}
	node.Annotation[tp.GetAnnoName()] = healthyAnno
	node.Annotation[tp.netUnhealthyKey] = netUnhealthyAnno
	klog.V(util.LogDebugLev).Infof("%s after UpdateNodeInfo node<%s> Annotation: %#v",
		tp.GetPluginName(), node.Name, node.Annotation)
	return &node
}
