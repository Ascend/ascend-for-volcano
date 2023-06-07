/*
Copyright(C)2023. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package card910bx2 is using for HuaWei Ascend 910B(Atlas 300T A2) card pin affinity schedule.
*/
package card910bx2

import (
	"errors"
	"fmt"

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
	m := &card910bx2{}
	m.SetPluginName(name)
	m.SetAnnoName(util.NPU910CardName)
	m.SetAnnoPreVal(util.NPU910CardNamePre)
	m.SetDefaultJobSchedulerConfig(nil)
	m.SetMaxNodeNPUNum(nodeNPUNumber)
	m.SetAcceleratorValue(util.JobKind910BValue)
	m.SetArch(util.HuaweiArchX86 + util.HuaweiArchArm)
	m.AffScoreList = [][]int{
		{util.AffScore0, util.AffScore1},
		{util.AffScore2, util.AffScore0},
	}
	return m
}

// ValidNPUJob check job req npu num and mode
func (tp *card910bx2) ValidNPUJob() *api.ValidateResult {
	return tp.Valid910bNPUJob()
}

// PreStartAction pre-processing actions for rescheduling
func (tp *card910bx2) PreStartAction(ssn *framework.Session) error {
	moduleFullName := util.NPU910CardName + util.Card910bx2AcceleratorType
	klog.V(util.LogInfoLev).Infof("Entering PreStartAction of %s...", moduleFullName)
	if tp == nil || ssn == nil || tp.FrameAttr.KubeClient == nil {
		return fmt.Errorf("%s handler not enabled or ssn is nil: %s", moduleFullName, util.ArgumentError)
	}
	tp.reHandle = rescheduling.New(&tp.ScheduleEnv, rescheduling.CmFaultJob910bx2Kind)
	if tp.reHandle == nil {
		klog.V(util.LogErrorLev).Infof("create new fault handler failed.")
		return fmt.Errorf("%s reSchedule not enabled: %s", moduleFullName, util.ArgumentError)
	}
	tp.reHandle.NewCommonReScheduler(rescheduling.CmFaultJob910bx2Kind)
	tp.reHandle.SynCacheFaultNodeWithSession(util.NPU910CardName)
	tp.reHandle.AddFaultNodeWithSession(util.NPU910CardName)
	tp.reHandle.SynCacheFaultJobWithSession(ssn, util.NPU910CardName, util.NPU910CardNamePre)
	tp.reHandle.SynCacheNodeRankOccMapWithSession(ssn)
	// 1. restart Fault Jobs that are recorded in cache
	if restartErr := tp.reHandle.RestartNeedForceDeleteJobs(ssn); restartErr != nil {
		klog.V(util.LogErrorLev).Infof("%s RestartNeedForceDeleteJobs: %s", moduleFullName, restartErr.Error())
	}
	// 2. get all the new 910bx2 jobs in session
	runningJobs910bx2, getRunErr := tp.reHandle.GetRunningJobs(ssn, util.Ascend910bName, util.Card910bx2AcceleratorType)
	if getRunErr != nil {
		klog.V(util.LogInfoLev).Infof("%s GetRunningJobs: %s", moduleFullName, getRunErr.Error())
	}
	// 3. get nodes of session and fault jobs of 910bx2
	err := tp.reHandle.AddFaultJobWithSession(runningJobs910bx2, util.NPU910CardName, util.NPU910CardNamePre)
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
	klog.V(util.LogInfoLev).Infof("Leaving PreStartAction of %s", moduleFullName)
	return nil
}

// PreStopAction post-processing actions for re-scheduling
func (tp *card910bx2) PreStopAction(env *plugin.ScheduleEnv) error {
	moduleFullName := util.NPU910CardName + util.Card910bx2AcceleratorType
	klog.V(util.LogInfoLev).Infof("enter PreStopAction %s...", moduleFullName)
	if tp == nil || tp.reHandle == nil || env == nil || tp.FrameAttr.KubeClient == nil {
		return fmt.Errorf("%s reSchedule not enabled or nil env: %s", moduleFullName, util.ArgumentError)
	}
	if err := tp.reHandle.WriteReSchedulerCacheToEnvCache(env, rescheduling.CmFaultJob910bx2Kind); err != nil {
		return err
	}
	klog.V(util.LogInfoLev).Infof("leave PreStopAction %s...", moduleFullName)
	return nil
}

// CheckNodeNPUByTask check nod npu meet task req
func (tp *card910bx2) CheckNodeNPUByTask(task *api.TaskInfo, node plugin.NPUNode) error {
	if tp == nil || task == nil || len(node.Annotation) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("CheckNodeNPUByTask err: %s", err)
		return err
	}
	return nil
}

// ScoreBestNPUNodes core node by calculate task req npu num and node npu top
func (tp *card910bx2) ScoreBestNPUNodes(task *api.TaskInfo, nodes []*api.NodeInfo, sMap map[string]float64) error {
	return tp.ScoreAscendNPUNodes(task, nodes, sMap)
}

// UseAnnotation select npu for task from node
func (tp *card910bx2) UseAnnotation(task *api.TaskInfo, node plugin.NPUNode) *plugin.NPUNode {
	return tp.Use910bAnnotation(task, node)
}

// ReleaseAnnotation Release used resource.
func (tp *card910bx2) ReleaseAnnotation(_ *api.TaskInfo, node plugin.NPUNode) *plugin.NPUNode {
	return &node
}
