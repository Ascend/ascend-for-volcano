/*
Copyright(C)2020-2023. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package module910bx16 is using for HuaWei A800/9000 Ascend910B A+X pin affinity schedule.
*/
package module910bx16

import (
	"fmt"

	"k8s.io/klog"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/ascend910/ascend910b"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

func (tp *module910bx16) getUsableTopFromNode(node plugin.NPUNode) ([]int, error) {
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

	return nodeTop, nil
}

func initSelectNodeInf(npuTop []int) ascend910b.SelectNodeInf {
	var sNodeInf ascend910b.SelectNodeInf
	var leftHccsTop []int
	var rightHccsTop []int

	for _, cardID := range npuTop {
		if cardID < nodeNPUNumber {
			leftHccsTop = append(leftHccsTop, cardID)
		} else {
			rightHccsTop = append(rightHccsTop, cardID)
		}
	}
	sNodeInf.LeftNPUNum = len(leftHccsTop)
	sNodeInf.RightNPUNum = len(rightHccsTop)
	sNodeInf.AllNPUNum = sNodeInf.LeftNPUNum + sNodeInf.RightNPUNum

	return sNodeInf
}

func getNodeHccsArray(nodeTop []int) ([]int, []int) {
	var leftHccsArray []int
	var rightHccsArray []int

	for _, v := range nodeTop {
		if v < util.NPUIndex8 {
			leftHccsArray = append(leftHccsArray, v)
			continue
		}
		rightHccsArray = append(rightHccsArray, v)
	}

	return leftHccsArray, rightHccsArray
}

func (tp *module910bx16) getNodeBestScore(taskNPUNum int, npuTop []int) (int, error) {
	var bestScore = util.AffScore8

	sNodeInf := initSelectNodeInf(npuTop)
	if sNodeInf.AllNPUNum < 1 ||
		sNodeInf.AllNPUNum > tp.MaxNodeNPUNum ||
		sNodeInf.RightNPUNum > util.NPUIndex8 ||
		sNodeInf.LeftNPUNum > util.NPUIndex8 {
		return bestScore, fmt.Errorf("node top %#v is invalid", npuTop)
	}

	var err = fmt.Errorf("node %#v is not meet task req %d", npuTop, taskNPUNum)
	if taskNPUNum == tp.MaxNodeNPUNum {
		if len(npuTop) == tp.MaxNodeNPUNum {
			return 0, nil
		}
		return 0, err
	}
	if taskNPUNum < 1 || taskNPUNum > tp.MaxNodeNPUNum {
		return bestScore, fmt.Errorf("task req npu num<%d> is invalid", taskNPUNum)
	}
	switch {
	case sNodeInf.RightNPUNum == 0:
		bestScore = tp.affScoreList[taskNPUNum-1][sNodeInf.LeftNPUNum-1]
	case sNodeInf.LeftNPUNum == 0:
		bestScore = tp.affScoreList[taskNPUNum-1][sNodeInf.RightNPUNum-1]
	default:
		bestScore = util.Min(tp.affScoreList[taskNPUNum-1][sNodeInf.RightNPUNum-1],
			tp.affScoreList[taskNPUNum-1][sNodeInf.LeftNPUNum-1])
	}
	if bestScore == util.AffScore8 {
		return 0, err
	}
	return bestScore, nil
}

// UpdateNodeInfo update node info
func (tp *module910bx16) UpdateNodeInfo(node plugin.NPUNode, usedTop []int) *plugin.NPUNode {
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
