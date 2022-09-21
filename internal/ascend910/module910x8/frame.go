/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package module910x8 is using for HuaWei A800/9000 Ascend910 pin affinity schedule.

*/
package module910x8

import (
	"errors"
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
	m := &module910x8{}
	m.SetPluginName(name)
	m.SetAnnoName(util.NPU910CardName)
	m.SetAnnoPreVal(util.NPU910CardNamePre)
	m.SetDefaultJobSchedulerConfig(nil)
	m.SetMaxNodeNPUNum(nodeNPUNumber)
	m.netUnhealthyKey = networkUnhealthyNPU
	m.affScoreList = [][]int{
		{affScore0, affScore2, affScore1, affScore3},
		{affScore4, affScore0, affScore2, affScore1},
		{affScore4, affScore4, affScore4, affScore4},
		{affScore4, affScore4, affScore4, affScore0},
	}
	return m
}

// ValidNPUJob check job req npu num and mode
func (tp *module910x8) ValidNPUJob() *api.ValidateResult {
	if tp == nil {
		err := fmt.Errorf("nil plugin %s", SchedulerName)
		klog.V(util.LogErrorLev).Infof("ValidNPUJob err: %s.", err.Error())
		return &api.ValidateResult{
			Pass:    false,
			Reason:  err.Error(),
			Message: err.Error(),
		}
	}

	jobNPU := tp.ReqNPUNum
	if jobNPU <= nodeNPUNumber {
		if err := tp.checkSingleTrainMode(); err != nil {
			klog.V(util.LogErrorLev).Infof("%s ValidNPUJob err: %s", tp.GetPluginName(), err.Error())
			return &api.ValidateResult{
				Pass:    false,
				Reason:  "job req npu is invalid",
				Message: err.Error(),
			}
		}
		return nil
	}

	if err := tp.checkModuleDistributeTrainMode(); err != nil {
		klog.V(util.LogErrorLev).Infof("%s ValidNPUJob err: %s", tp.GetPluginName(), err.Error())
		return &api.ValidateResult{
			Pass:    false,
			Reason:  "job req npu is invalid",
			Message: err.Error(),
		}
	}
	return nil
}

// PreStartAction pre-processing actions for rescheduling
func (tp *module910x8) PreStartAction(ssn *framework.Session) error {
	moduleFullName := util.NPU910CardName + util.ModuleAcceleratorType
	klog.V(util.LogInfoLev).Infof("Entering PreStartAction of %s...", moduleFullName)
	defer klog.V(util.LogInfoLev).Infof("Leaving PreStartAction of %s", moduleFullName)
	if tp == nil {
		return fmt.Errorf("%s handler not enabled: %s", moduleFullName, util.ArgumentError)
	}
	if ssn == nil {
		return fmt.Errorf("%s session is nil: %s", moduleFullName, util.ArgumentError)
	}
	reschEnable, ok := tp.SchedulerJobAttr.Label[rescheduling.JobRescheduleLabelKey]
	if !ok {
		klog.V(util.LogErrorLev).Infof("%s no re-scheduler key 910x8", moduleFullName)
		return nil
	}
	if reschEnable == rescheduling.JobOffRescheduleLabelValue {
		klog.V(util.LogInfoLev).Infof("%s RescheduleLabel not enabled", moduleFullName)
		return nil
	}
	tp.reHandle = rescheduling.New(&tp.ScheduleEnv, rescheduling.CmFaultJob910x8Kind)
	if tp.reHandle == nil {
		return fmt.Errorf("%s reSchedule not enabled: %s", moduleFullName, util.ArgumentError)
	}
	tp.reHandle.New910ReScheduler()
	tp.reHandle.SynCacheFaultNodeWithSession(util.NPU910CardName)
	tp.reHandle.AddFaultNodeWithSession(util.NPU910CardName)
	tp.reHandle.SynCacheFaultJobWithSession(ssn, util.NPU910CardName, util.NPU910CardNamePre)
	tp.reHandle.SynCacheNodeRankOccMapWithSession(ssn)
	// 1. restart Fault Jobs that are recorded in cache
	if restartErr := tp.reHandle.RestartNeedForceDeleteJobs(ssn); restartErr != nil {
		klog.V(util.LogErrorLev).Infof("%s RestartNeedForceDeleteJobs: %s", moduleFullName, restartErr.Error())
	}
	// 2. get all the new 910x8 jobs in session
	runningJobs910x8, getRunErr := tp.reHandle.GetRunningJobs(ssn, util.NPU910CardName, util.ModuleAcceleratorType)
	if getRunErr != nil {
		klog.V(util.LogErrorLev).Infof("%s GetRunningJobs: %s", moduleFullName, getRunErr.Error())
	}
	// 3. get nodes of session and fault jobs of 910x8
	err := tp.reHandle.AddFaultJobWithSession(runningJobs910x8, util.NPU910CardName, util.NPU910CardNamePre)
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
func (tp *module910x8) PreStopAction(env *plugin.ScheduleEnv) error {
	moduleFullName := util.NPU910CardName + util.ModuleAcceleratorType
	klog.V(util.LogInfoLev).Infof("enter PreStopAction %s...", moduleFullName)
	defer klog.V(util.LogInfoLev).Infof("leave PreStopAction %s...", moduleFullName)
	if tp == nil || tp.reHandle == nil {
		return fmt.Errorf("%s reSchedule not enabled: %s", moduleFullName, util.ArgumentError)
	}
	if env == nil {
		return fmt.Errorf("%s env is nil: %s", moduleFullName, util.ArgumentError)
	}
	if err := tp.reHandle.WriteReSchedulerCacheToEnvCache(env, rescheduling.CmFaultJob910x8Kind); err != nil {
		return err
	}
	return nil
}

// CheckNodeNPUByTask check nod npu meet task req
func (tp *module910x8) CheckNodeNPUByTask(task *api.TaskInfo, node plugin.NPUNode) error {
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
	job, ok := tp.Jobs[task.Job]
	if !ok {
		err = fmt.Errorf("task<%s> is not npu task", task.Name)
		klog.V(util.LogErrorLev).Infof("%s CheckNodeNPUByTask err: %s", tp.GetPluginName(), err.Error())
		return err
	}
	if taskNPUNum < tp.MaxNodeNPUNum && len(job.Tasks) > 1 {
		err = fmt.Errorf("distribute task<%s> req num<%d> is invalid", task.Name, taskNPUNum)
		klog.V(util.LogErrorLev).Infof("%s CheckNodeNPUByTask err: %s", tp.GetPluginName(), err.Error())
		return err
	}
	nodeTop, err := tp.getUsableTopFromNode(node, len(tp.Tasks) > 1)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s CheckNodeNPUByTask err: %s", tp.GetPluginName(), err.Error())
		return err
	}

	if err = tp.judgeNodeAndTaskNPU(taskNPUNum, nodeTop); err != nil {
		klog.V(util.LogErrorLev).Infof("%s CheckNodeNPUByTask err: %s", tp.GetPluginName(), err.Error())
		return fmt.Errorf("checkNodeNPUByTask %s err: %s", util.NodeNotMeetTopologyWarning, err.Error())
	}

	if tp.reHandle != nil {
		if reErr := tp.reHandle.CheckNodeNPUByTask(task, node); reErr != nil {
			return fmt.Errorf("rescheduling CheckNodeNPUByTask %s", reErr.Error())
		}
	}
	return nil
}

// ScoreBestNPUNodes core node by calculate task req npu num and node npu top
func (tp *module910x8) ScoreBestNPUNodes(task *api.TaskInfo, nodes []*api.NodeInfo, scoreMap map[string]float64) error {
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
		cardIds, err := tp.getUsableTopFromNode(nNode, len(tp.Tasks) > 1)
		if err != nil {
			klog.V(util.LogWarningLev).Infof("%s ScoreBestNPUNodes err: %s", tp.GetPluginName(), err.Error())
			continue
		}
		bestScore, err := tp.getNodeBestScore(taskNPUNum, cardIds)
		if err != nil {
			klog.V(util.LogWarningLev).Infof("%s ScoreBestNPUNodes err: %s", tp.GetPluginName(), err.Error())
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
			"%s rescheduling ScoreBestNPUNodes failed :%s.", SchedulerName, reErr.Error())
	}
	klog.V(util.LogInfoLev).Infof("%s ScoreBestNPUNodes task<%s> scoreMap<%v>", tp.GetPluginName(),
		task.Name, scoreMap)
	return nil
}

// UseAnnotation select npu for task from node
func (tp *module910x8) UseAnnotation(task *api.TaskInfo, node plugin.NPUNode) *plugin.NPUNode {
	if tp == nil || task == nil || len(node.Annotation) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("UseAnnotation %s.", err.Error())
		return nil
	}
	klog.V(util.LogDebugLev).Infof("%s UseAnnotation task<%s> node<%s> resource<%s> Annotation: %#v",
		tp.GetPluginName(), task.Name, node.Name, tp.GetAnnoName(), node.Annotation)
	selectedNPU, err := tp.selectNPUFromNode(task, node)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s UseAnnotation err:%s.", tp.GetPluginName(), err.Error())
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

func (tp *module910x8) selectNPUFromNode(task *api.TaskInfo, node plugin.NPUNode) ([]int, error) {
	taskNPUNum, err := tp.GetTaskReqNPUNum(task)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s ScoreBestNPUNodes err: %s", tp.GetPluginName(), err.Error())
		return nil, err
	}
	nodeTop, err := tp.getUsableTopFromNode(node, len(tp.Tasks) > 1)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s ScoreBestNPUNodes err: %s", tp.GetPluginName(), err.Error())
		return nil, err
	}
	if taskNPUNum == nodeNPUNumber {
		if len(nodeTop) == nodeNPUNumber {
			return nodeTop, nil
		}
		err = fmt.Errorf("node<%s> top<%v> can not meet task req<%d>", node.Name, nodeTop, taskNPUNum)
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

	leftHccsArray, rightHccsArray := getNodeHccsArray(nodeTop)
	for _, priority := range priorityArray {
		if priority == len(leftHccsArray) {
			return leftHccsArray[:taskNPUNum], nil
		}
		if priority == len(rightHccsArray) {
			return rightHccsArray[:taskNPUNum], nil
		}
	}
	err = fmt.Errorf("node<%s> top<%v> can not meet task req<%d>", node.Name, len(nodeTop), taskNPUNum)
	klog.V(util.LogErrorLev).Infof("%s ScoreBestNPUNodes err: %s", tp.GetPluginName(), err.Error())
	return nil, err
}
