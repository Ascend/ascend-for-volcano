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
Package half910x4 is using for HuaWei A800/9000 Ascend910 pin affinity schedule.
*/
package half910x4

import (
	"fmt"
	"reflect"

	"k8s.io/api/core/v1"
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/base"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/rescheduling"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// New return npu plugin
func New(name string) base.AscendHandler {
	m := &half910x4{}
	m.SetPluginName(name)
	m.SetAnnoName(util.NPU910CardName)
	m.SetAnnoPreVal(util.NPU910CardNamePre)
	m.SetDefaultJobSchedulerConfig(nil)
	m.SetMaxNodeNPUNum(npuNumPerHccs)
	m.netUnhealthyKey = networkUnhealthyNPU
	m.affScoreList = [][]int{
		{util.AffScore0, util.AffScore2, util.AffScore1, util.AffScore3},
		{util.AffScore4, util.AffScore0, util.AffScore2, util.AffScore1},
		{util.AffScore4, util.AffScore4, util.AffScore4, util.AffScore4},
		{util.AffScore4, util.AffScore4, util.AffScore4, util.AffScore0},
	}
	return m
}

// ValidNPUJob check job req npu num
func (tp *half910x4) ValidNPUJob() *api.ValidateResult {
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

	// 2.check job train mode:distribute and single.
	if vErr = tp.checkJobTrainMode(); vErr != nil {
		klog.V(util.LogErrorLev).Infof("checkJobTrainMode: %s.", vErr)
		return vResult
	}

	return nil
}

// PreStartAction pre-processing actions for rescheduling
func (tp *half910x4) PreStartAction(ssn *framework.Session) error {
	moduleFullName := util.NPU910CardName + util.HalfAcceleratorType
	klog.V(util.LogInfoLev).Infof("Entering PreStartAction of %s...", moduleFullName)
	defer klog.V(util.LogInfoLev).Infof("Leaving PreStartAction of %s", moduleFullName)
	if tp == nil || ssn == nil || tp.FrameAttr.KubeClient == nil {
		return fmt.Errorf("%s handler not enabled or ssn is nil: %s", moduleFullName, util.ArgumentError)
	}

	tp.reHandle = rescheduling.New(&tp.ScheduleEnv, rescheduling.CmFaultJob910x4Kind)
	if tp.reHandle == nil {
		klog.V(util.LogErrorLev).Infof("create new fault handler failed.")
		return fmt.Errorf("%s reSchedule not enabled: %s", moduleFullName, util.ArgumentError)
	}
	tp.reHandle.NewCommonReScheduler(rescheduling.CmFaultJob910x4Kind)
	tp.reHandle.SynCacheFaultNodeWithSession(util.NPU910CardName)
	tp.reHandle.AddFaultNodeWithSession(util.NPU910CardName)
	tp.reHandle.SynCacheFaultJobWithSession(ssn, util.NPU910CardName, util.NPU910CardNamePre)
	tp.reHandle.SynCacheNodeRankOccMapWithSession(ssn)
	// 1. restart Fault Jobs that are recorded in cache
	if restartErr := tp.reHandle.RestartNeedForceDeleteJobs(ssn); restartErr != nil {
		klog.V(util.LogErrorLev).Infof("%s RestartNeedForceDeleteJobs: %s", moduleFullName, restartErr.Error())
	}
	// 2. get all the new jobs in session
	runningJobs, getRunErr := tp.reHandle.GetRunningJobs(ssn, util.NPU910CardName, util.HalfAcceleratorType)
	if getRunErr != nil {
		klog.V(util.LogInfoLev).Infof("%s GetRunningJobs: %s", moduleFullName, getRunErr.Error())
	}
	// 3. get nodes of session and fault jobs
	err := tp.reHandle.AddFaultJobWithSession(runningJobs, util.NPU910CardName, util.NPU910CardNamePre)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s AddFaultJobWithSession", moduleFullName)
	}
	// 4. restart the fault jobs
	if restartErr := tp.reHandle.RestartFaultJobs(ssn); restartErr != nil {
		klog.V(util.LogErrorLev).Infof("%s RestartFaultJobs: %s", moduleFullName, restartErr.Error())
		return restartErr
	}
	// 5. save structure for later allocation process
	tp.reHandle.GenerateNodeRankIndexTaskMap()
	return nil
}

// PreStopAction post-processing actions for re-scheduling
func (tp *half910x4) PreStopAction(env *plugin.ScheduleEnv) error {
	moduleFullName := util.NPU910CardName + util.HalfAcceleratorType
	klog.V(util.LogInfoLev).Infof("enter PreStopAction %s...", moduleFullName)
	defer klog.V(util.LogInfoLev).Infof("leave PreStopAction %s...", moduleFullName)
	if tp == nil || tp.reHandle == nil || env == nil {
		return fmt.Errorf("%s reSchedule not enabled or nil env: %s", moduleFullName, util.ArgumentError)
	}
	if err := tp.reHandle.WriteReSchedulerCacheToEnvCache(env, rescheduling.CmFaultJob910x4Kind); err != nil {
		return err
	}
	return nil
}

// CheckNodeNPUByTask check nod npu meet task req
func (tp *half910x4) CheckNodeNPUByTask(task *api.TaskInfo, node plugin.NPUNode) error {
	if len(node.Annotation) == 0 {
		err := fmt.Errorf("node<%s> annotation is empty", node.Name)
		klog.V(util.LogErrorLev).Infof("CheckNodeNPUByTask err: %s", err)
		return err
	}
	taskNPUNum, err := tp.GetTaskReqNPUNum(task)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s CheckNodeNPUByTask err: %s", tp.GetPluginName(), err)
		return err
	}
	nTaskNum := tp.GetNPUTaskNumInJob()
	nodeTop, err := tp.getUsableTopFromNode(node, nTaskNum > 1)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s CheckNodeNPUByTask err: %s", tp.GetPluginName(), err)
		return err
	}

	if err = tp.judgeNodeAndTaskNPU(taskNPUNum, nodeTop); err != nil {
		klog.V(util.LogErrorLev).Infof("%s CheckNodeNPUByTask err: %s", tp.GetPluginName(), err)
		return fmt.Errorf("checkNodeNPUByTask %s err: %s", util.NodeNotMeetTopologyWarning, err)
	}

	if tp.reHandle != nil {
		if reErr := tp.reHandle.CheckNodeNPUByTask(task, node); reErr != nil {
			return fmt.Errorf("rescheduling CheckNodeNPUByTask %s", reErr)
		}
	}
	return nil
}

// ScoreBestNPUNodes core node by calculate task req npu num and node npu top
func (tp *half910x4) ScoreBestNPUNodes(task *api.TaskInfo, nodes []*api.NodeInfo, scoreMap map[string]float64) error {
	if len(scoreMap) == 0 {
		err := fmt.Errorf("score map is empty")
		klog.V(util.LogErrorLev).Infof("ScoreBestNPUNodes %v.", err)
		return err
	}
	taskNPUNum, err := tp.GetTaskReqNPUNum(task)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s ScoreBestNPUNodes err: %s", tp.GetPluginName(), err)
		return err
	}
	for _, node := range nodes {
		if reflect.ValueOf(node).IsNil() {
			continue
		}
		nNode, ok := tp.Nodes[node.Name]
		if !ok {
			klog.V(util.LogWarningLev).Infof("%s ScoreBestNPUNodes node<%s> is not npu node",
				tp.GetPluginName(), node.Name)
			continue
		}
		nTaskNum := tp.GetNPUTaskNumInJob()
		cardIds, err := tp.getUsableTopFromNode(nNode, nTaskNum > 1)
		if err != nil {
			klog.V(util.LogWarningLev).Infof("%s ScoreBestNPUNodes err: %s", tp.GetPluginName(), err)
			continue
		}
		bestScore, err := tp.getNodeBestScore(taskNPUNum, cardIds)
		if err != nil {
			klog.V(util.LogWarningLev).Infof("%s ScoreBestNPUNodes err: %s", tp.GetPluginName(), err)
			continue
		}
		healthyNPUNum, ok := nNode.Allocate[v1.ResourceName(tp.GetAnnoName())]
		if !ok {
			klog.V(util.LogWarningLev).Infof("%s ScoreBestNPUNodes node<%s> get allocate npu failed",
				tp.GetPluginName(), node.Name)
			continue
		}
		scoreMap[node.Name] = nodeWeight * float64(int(healthyNPUNum/util.NPUHexKilo)*npuNumPerHccs-bestScore)
	}
	reErr := tp.reHandle.ScoreBestNPUNodes(task, scoreMap)
	if reErr != nil {
		klog.V(util.LogErrorLev).Infof(
			"%s rescheduling ScoreBestNPUNodes failed :%s.", SchedulerName, reErr)
	}
	klog.V(util.LogInfoLev).Infof("%s ScoreBestNPUNodes task<%s> scoreMap<%v>", tp.GetPluginName(),
		task.Name, scoreMap)
	return nil
}

// UseAnnotation select npu for task from node
func (tp *half910x4) UseAnnotation(task *api.TaskInfo, node plugin.NPUNode) *plugin.NPUNode {
	if len(node.Annotation) == 0 {
		err := fmt.Errorf("node annotation is empty")
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
	klog.V(util.LogInfoLev).Infof("%s UseAnnotation task<%s> select npu <%v>.",
		tp.GetPluginName(), task.Name, selectedNPU)

	tp.SetNPUTopologyToPodFn(task, selectedNPU)
	newNode := tp.UpdateNodeInfo(node, selectedNPU)
	if tp.reHandle != nil {
		if reErr := tp.reHandle.UseAnnotation(task, newNode); reErr != nil {
			klog.V(util.LogErrorLev).Infof("%s rescheduling UseAnnotation: %s", SchedulerName, reErr.Error())
			return nil
		}
	}
	return newNode
}

func (tp *half910x4) selectNPUFromNode(task *api.TaskInfo, node plugin.NPUNode) ([]int, error) {
	taskNPUNum, err := tp.GetTaskReqNPUNum(task)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("ScoreBestNPUNodes err: %s", err)
		return nil, err
	}
	nTaskNum := tp.GetNPUTaskNumInJob()
	nodeTop, err := tp.getUsableTopFromNode(node, nTaskNum > 1)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("ScoreBestNPUNodes err: %s", err)
		return nil, err
	}
	if len(nodeTop) < taskNPUNum {
		err = fmt.Errorf("node<%s> top<%v> can not meet task req<%d>", node.Name, len(nodeTop), taskNPUNum)
		klog.V(util.LogErrorLev).Infof("ScoreBestNPUNodes err: %s", err)
		return nil, err
	}
	return nodeTop[:taskNPUNum], nil
}

// ReleaseAnnotation Release used resource.
func (tp *half910x4) ReleaseAnnotation(_ *api.TaskInfo, node plugin.NPUNode) *plugin.NPUNode {
	return &node
}
