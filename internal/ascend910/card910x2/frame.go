/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package card910x2 is using for HuaWei Ascend pin affinity schedule.
*/
package card910x2

import (
	"errors"
	"fmt"

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
	c := &card910x2{}
	c.SetPluginName(name)
	c.SetAnnoName(util.NPU910CardName)
	c.SetAnnoPreVal(util.NPU910CardNamePre)
	c.SetDefaultJobSchedulerConfig(nil)
	c.SetMaxNodeNPUNum(maxNodeNPUNum)
	c.affScoreList = [][]int{
		{affScore0, affScore1},
		{affScore2, affScore0},
	}
	return c
}

// ValidNPUJob check job req npu num and mode
func (tp *card910x2) ValidNPUJob() *api.ValidateResult {
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
func (tp *card910x2) PreStartAction(ssn *framework.Session) error {
	cardFullName := util.NPU910CardName + util.CardAcceleratorType
	klog.V(util.LogInfoLev).Infof("Entering PreStartAction of %s", cardFullName)
	defer klog.V(util.LogInfoLev).Infof("Leaving PreStartAction of %s", cardFullName)
	if tp == nil || ssn == nil || tp.FrameAttr.KubeClient == nil {
		return fmt.Errorf("%s handler not enabled or ssn is nil: %s", cardFullName, util.ArgumentError)
	}
	tp.reHandle = rescheduling.New(&tp.ScheduleEnv, rescheduling.CmFaultJob910x2Kind)
	if tp.reHandle == nil {
		return fmt.Errorf("%s reSchedule not enabled: %s", cardFullName, util.ArgumentError)
	}
	tp.reHandle.NewCommonReScheduler(rescheduling.CmFaultJob910x2Kind)
	tp.reHandle.SynCacheFaultNodeWithSession(util.NPU910CardName)
	tp.reHandle.AddFaultNodeWithSession(util.NPU910CardName)
	tp.reHandle.SynCacheFaultJobWithSession(ssn, util.NPU910CardName, util.NPU910CardNamePre)
	// 1. restart Fault Jobs that are recorded in cache
	if restartErr := tp.reHandle.RestartNeedForceDeleteJobs(ssn); restartErr != nil {
		klog.V(util.LogErrorLev).Infof("%s RestartNeedForceDeleteJobs: %s", cardFullName, restartErr.Error())
	}
	// 2. get all the new 910x2 jobs in session
	runningJobs910x2, getRunErr := tp.reHandle.GetRunningJobs(ssn, util.NPU910CardName, util.CardAcceleratorType)
	if getRunErr != nil {
		klog.V(util.LogErrorLev).Infof("%s GetRunningJobs: %s", cardFullName, getRunErr.Error())
	}
	// 3. get nodes of session and fault jobs of 910x2
	err := tp.reHandle.AddFaultJobWithSession(runningJobs910x2, util.NPU910CardName, util.NPU910CardNamePre)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s AddFaultJobWithSession", cardFullName)
	}
	// 4. restart the fault jobs
	if restartErr := tp.reHandle.RestartFaultJobs(ssn); restartErr != nil {
		klog.V(util.LogErrorLev).Infof("%s RestartFaultJobs: %s", cardFullName, restartErr.Error())
		return restartErr
	}
	return nil
}

// PreStopAction post-processing actions for re-scheduling
func (tp *card910x2) PreStopAction(env *plugin.ScheduleEnv) error {
	cardFullName := util.NPU910CardName + util.CardAcceleratorType
	klog.V(util.LogInfoLev).Infof("enter PreStopAction of %s...", cardFullName)
	defer klog.V(util.LogInfoLev).Infof("leave PreStopAction of %s...", cardFullName)
	if tp == nil || tp.reHandle == nil || env == nil {
		return fmt.Errorf("%s reSchedule not enabled or nil env: %s", cardFullName, util.ArgumentError)
	}
	if err := tp.reHandle.WriteReSchedulerCacheToEnvCache(env, rescheduling.CmFaultJob910x2Kind); err != nil {
		return err
	}
	return nil
}

// CheckNodeNPUByTask check nod npu meet task req
func (tp *card910x2) CheckNodeNPUByTask(task *api.TaskInfo, node plugin.NPUNode) error {
	klog.V(util.LogDebugLev).Infof("CheckNodeNPUByTask %v.", tp.GetPluginName())
	if tp == nil || task == nil || len(node.Annotation) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("CheckNodeNPUByTask err: %s", err.Error())
		return err
	}
	taskNPUNum, err := tp.GetTaskReqNPUNum(task)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s CheckNodeNPUByTask err: %s", tp.GetPluginName(), err.Error())
		return err
	}
	_, ok := tp.Jobs[task.Job]
	if !ok {
		err = fmt.Errorf("task<%s> is not npu task", task.Name)
		klog.V(util.LogErrorLev).Infof("%s CheckNodeNPUByTask err: %s", tp.GetPluginName(), err.Error())
		return err
	}
	nodeTop, err := tp.GetUsableTopFromNode(node)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s CheckNodeNPUByTask err: %s", tp.GetPluginName(), err.Error())
		return err
	}

	if len(nodeTop) < taskNPUNum {
		return fmt.Errorf("node <%s> don't have enough resource <%s>, req<%d>, idle<%d>",
			node.Name, tp.GetAnnoName(), taskNPUNum, len(nodeTop))
	}
	return nil
}

// ScoreBestNPUNodes score node by calculate task req npu num and node npu top
func (tp *card910x2) ScoreBestNPUNodes(task *api.TaskInfo, nodes []*api.NodeInfo, scoreMap map[string]float64) error {
	if tp == nil || task == nil || len(nodes) == 0 || len(scoreMap) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("ScoreBestNPUNodes %v.", err.Error())
		return err
	}
	taskNPUNum, err := tp.GetTaskReqNPUNum(task)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s ScoreBestNPUNodes err: %s", tp.GetPluginName(), err.Error())
		return err
	}
	if taskNPUNum < 1 || taskNPUNum > tp.MaxNodeNPUNum {
		err = fmt.Errorf("task<%s> req npu num<%d> is invalid", task.Name, taskNPUNum)
		klog.V(util.LogErrorLev).Infof("%s ScoreBestNPUNodes err: %s", tp.GetPluginName(), err.Error())
		return err
	}
	for _, node := range nodes {
		nNode, ok := tp.Nodes[node.Name]
		if !ok {
			continue
		}
		nodeTop, err := tp.GetUsableTopFromNode(nNode)
		if err != nil {
			klog.V(util.LogErrorLev).Infof("%s ScoreBestNPUNodes err: %s", tp.GetPluginName(), err.Error())
			continue
		}
		if len(nodeTop) > tp.MaxNodeNPUNum {
			continue
		}
		bestScore := tp.affScoreList[taskNPUNum-1][len(nodeTop)-1]
		if bestScore == affScore2 {
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
	return nil
}

// ReleaseAnnotation Release used resource.
func (tp *card910x2) ReleaseAnnotation(_ *api.TaskInfo, node plugin.NPUNode) *plugin.NPUNode {
	return &node
}
