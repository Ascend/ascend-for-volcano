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
	"errors"
	"fmt"
	"reflect"

	"k8s.io/api/core/v1"
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/base"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// New return npu plugin
func New(name string) base.AscendHandler {
	m := &module910bx16{}
	m.SetPluginName(name)
	m.SetAnnoName(util.NPU910CardName)
	m.SetAnnoPreVal(util.NPU910CardNamePre)
	m.SetDefaultJobSchedulerConfig(nil)
	m.SetMaxNodeNPUNum(nodeNPUNumber)
	m.affScoreList = [][]int{
		{util.AffScore0, util.AffScore2, util.AffScore3, util.AffScore4, util.AffScore5, util.AffScore6,
			util.AffScore7, util.AffScore7},
		{util.AffScore8, util.AffScore0, util.AffScore1, util.AffScore2, util.AffScore3, util.AffScore4,
			util.AffScore5, util.AffScore6},
		{util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8,
			util.AffScore8, util.AffScore8},
		{util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore0, util.AffScore1, util.AffScore2,
			util.AffScore3, util.AffScore4},
		{util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8,
			util.AffScore8, util.AffScore8},
		{util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8,
			util.AffScore8, util.AffScore8},
		{util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8,
			util.AffScore8, util.AffScore8},
		{util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8,
			util.AffScore8, util.AffScore0},
	}
	return m
}

// ValidNPUJob check job req npu num and mode
func (tp *module910bx16) ValidNPUJob() *api.ValidateResult {
	vResult := &api.ValidateResult{}
	var vErr error
	defer func() {
		if vErr != nil {
			vResult.Pass = false
			vResult.Reason = vErr.Error()
			vResult.Message = vErr.Error()
			return
		}
	}()

	// 1. check parameter.
	if tp == nil {
		vErr = fmt.Errorf("nil plugin %s", SchedulerName)
		klog.V(util.LogErrorLev).Infof("ValidNPUJob err: %s.", vErr)
		return vResult
	}

	// 2.check ring-controller.atlas
	if vErr = tp.checkJobForm(); vErr != nil {
		klog.V(util.LogErrorLev).Infof("checkJobForm: %s.", vErr)
		return vResult
	}

	// 3.check host-arch for A+X not same as A+K
	if vErr = tp.checkJobArch(); vErr != nil {
		klog.V(util.LogErrorLev).Infof("checkJobArch: %s.", vErr)
		return vResult
	}

	// 4.check job train mode:distribute and single.
	if vErr = tp.checkJobTrainMode(); vErr != nil {
		klog.V(util.LogErrorLev).Infof("checkJobTrainMode: %s.", vErr)
		return vResult
	}

	return nil
}

// PreStartAction pre-processing actions for rescheduling
func (tp *module910bx16) PreStartAction(ssn *framework.Session) error {
	klog.V(util.LogInfoLev).Infof("Entering PreStartAction of %s...", tp.GetPluginName())
	defer klog.V(util.LogInfoLev).Infof("Leaving PreStartAction of %s", tp.GetPluginName())
	if tp == nil || ssn == nil || tp.FrameAttr.KubeClient == nil {
		return fmt.Errorf("%s handler not enabled or ssn is nil: %s", tp.GetPluginName(), util.ArgumentError)
	}
	return nil
}

// PreStopAction post-processing actions for re-scheduling
func (tp *module910bx16) PreStopAction(env *plugin.ScheduleEnv) error {
	klog.V(util.LogInfoLev).Infof("enter PreStopAction %s...", tp.GetPluginName())
	defer klog.V(util.LogInfoLev).Infof("leave PreStopAction %s...", tp.GetPluginName())
	if tp == nil || env == nil {
		return fmt.Errorf("%s reSchedule not enabled or nil env: %s", tp.GetPluginName(), util.ArgumentError)
	}
	return nil
}

// CheckNodeNPUByTask check nod npu meet task req
func (tp *module910bx16) CheckNodeNPUByTask(task *api.TaskInfo, node plugin.NPUNode) error {
	if tp == nil || task == nil || len(node.Annotation) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("CheckNodeNPUByTask err: %s", err)
		return err
	}

	taskNPUNum, err := tp.GetTaskReqNPUNum(task)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s GetTaskReqNPUNum %s: %s", tp.GetPluginName(), task.Name, err)
		return err
	}

	nodeTop, err := tp.getUsableTopFromNode(node)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s getUsableTopFromNode %s: %s", tp.GetPluginName(), task.Name, err)
		return err
	}

	if err = tp.judgeNodeAndTaskNPU(taskNPUNum, nodeTop); err != nil {
		klog.V(util.LogErrorLev).Infof("%s judgeNodeAndTaskNPU %s: %s", tp.GetPluginName(), task.Name, err)
		return fmt.Errorf("checkNodeNPUByTask %s err: %s", util.NodeNotMeetTopologyWarning, err)
	}

	return nil
}

// ScoreBestNPUNodes core node by calculate task req npu num and node npu top
func (tp *module910bx16) ScoreBestNPUNodes(task *api.TaskInfo, nodes []*api.NodeInfo, sMap map[string]float64) error {
	if tp == nil || task == nil || len(nodes) == 0 || len(sMap) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("ScoreBestNPUNodes %s.", err)
		return err
	}
	taskNPUNum, getErr := tp.GetTaskReqNPUNum(task)
	if getErr != nil {
		klog.V(util.LogErrorLev).Infof("%s GetTaskReqNPUNum %s: %s", tp.GetPluginName(), task.Name, getErr)
		return getErr
	}
	for _, node := range nodes {
		if reflect.ValueOf(node).IsNil() {
			continue
		}
		nNode, ok := tp.Nodes[node.Name]
		if !ok {
			klog.V(util.LogWarningLev).Infof("%s %s ScoreBestNPUNodes %s is not npu node",
				tp.GetPluginName(), task.Name, node.Name)
			continue
		}
		cardIds, err := tp.getUsableTopFromNode(nNode)
		if err != nil {
			klog.V(util.LogWarningLev).Infof("%s ScoreBestNPUNodes getErr: %s", tp.GetPluginName(), err)
			continue
		}
		bestScore, err := tp.getNodeBestScore(taskNPUNum, cardIds)
		if err != nil {
			klog.V(util.LogWarningLev).Infof("%s ScoreBestNPUNodes getErr: %s", tp.GetPluginName(), err)
			continue
		}
		healthyNPUNum, ok := nNode.Allocate[v1.ResourceName(tp.GetAnnoName())]
		if !ok {
			klog.V(util.LogWarningLev).Infof("%s ScoreBestNPUNodes node<%s> get allocate npu failed",
				tp.GetPluginName(), node.Name)
			continue
		}
		sMap[node.Name] = nodeWeight * float64(int(healthyNPUNum/util.NPUHexKilo)*tp.MaxNodeNPUNum-bestScore)
	}
	klog.V(util.LogInfoLev).Infof("%s ScoreBestNPUNodes task<%s> sMap<%v>", tp.GetPluginName(),
		task.Name, sMap)
	return nil
}

// UseAnnotation select npu for task from node
func (tp *module910bx16) UseAnnotation(task *api.TaskInfo, node plugin.NPUNode) *plugin.NPUNode {
	if tp == nil || task == nil || len(node.Annotation) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("UseAnnotation %s.", err)
		return nil
	}
	klog.V(util.LogDebugLev).Infof("%s UseAnnotation task<%s> node<%s> resource<%s> Annotation: %#v",
		tp.GetPluginName(), task.Name, node.Name, tp.GetAnnoName(), node.Annotation)
	selectedNPU, err := tp.selectNPUFromNode(task, node)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s UseAnnotation err:%s.", tp.GetPluginName(), err)
		return nil
	}
	klog.V(util.LogInfoLev).Infof("%s UseAnnotation %s select %v.", tp.GetPluginName(), task.Name, selectedNPU)

	tp.SetNPUTopologyToPodFn(task, selectedNPU)
	newNode := tp.UpdateNodeInfo(node, selectedNPU)
	return newNode
}

func (tp *module910bx16) selectNPUFromNode(task *api.TaskInfo, node plugin.NPUNode) ([]int, error) {
	taskNPUNum, err := tp.GetTaskReqNPUNum(task)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s ScoreBestNPUNodes err: %s", tp.GetPluginName(), err.Error())
		return nil, err
	}
	nodeTop, err := tp.getUsableTopFromNode(node)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s ScoreBestNPUNodes err: %s", tp.GetPluginName(), err.Error())
		return nil, err
	}
	if taskNPUNum == tp.MaxNodeNPUNum {
		if len(nodeTop) == tp.MaxNodeNPUNum {
			return nodeTop, nil
		}
		err = fmt.Errorf("%s %#v can not meet task req:%d", node.Name, nodeTop, taskNPUNum)
		klog.V(util.LogErrorLev).Infof("%s ScoreBestNPUNodes err: %s", tp.GetPluginName(), err.Error())
		return nil, err
	}
	priorityArray, err := getNPUAllocPriorityArray(taskNPUNum)
	if err != nil {
		klog.V(util.LogErrorLev).Info(err.Error())
		return nil, err
	}
	klog.V(util.LogInfoLev).Infof("%s selectNPUFromNode %s[%d] priority:%v in %v.", tp.GetPluginName(),
		task.Name, taskNPUNum, priorityArray, nodeTop)

	leftHCCSArray, rightHCCSArray := getNodeHccsArray(nodeTop)
	for _, priority := range priorityArray {
		if priority == len(leftHCCSArray) {
			return leftHCCSArray[:taskNPUNum], nil
		}
		if priority == len(rightHCCSArray) {
			return rightHCCSArray[:taskNPUNum], nil
		}
	}
	err = fmt.Errorf("node<%s> top<%v> can not meet task req<%d>", node.Name, len(nodeTop), taskNPUNum)
	klog.V(util.LogErrorLev).Infof("%s ScoreBestNPUNodes err: %s", tp.GetPluginName(), err.Error())
	return nil, err
}

// ReleaseAnnotation Release used resource.
func (tp *module910bx16) ReleaseAnnotation(_ *api.TaskInfo, node plugin.NPUNode) *plugin.NPUNode {
	return &node
}
