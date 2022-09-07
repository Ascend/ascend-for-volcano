/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package chip310p is using for HuaWei 310P Ascend pin affinity schedule.

*/

package ascend310P

import (
	"fmt"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/rescheduling"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// New return npu plugin.
func New(npuName string) plugin.ISchedulerPlugin {
	npuPlugin := &ascend310P{}
	npuPlugin.SetPluginName(npuName)
	npuPlugin.SetAnnoName(util.NPU310PCardName)
	npuPlugin.SetAnnoPreVal(util.NPU310PCardNamePre)
	npuPlugin.SetDefaultJobSchedulerConfig(nil)
	npuPlugin.SetMaxNodeNPUNum(maxNodeNPUNum)

	return npuPlugin
}

// PreStartAction pre-processing actions for rescheduling
func (tp *ascend310P) PreStartAction(ssn *framework.Session) error {
	klog.V(util.LogInfoLev).Infof("Entering PreStartAction of %s", util.NPU310PCardName)
	defer klog.V(util.LogInfoLev).Infof("Leaving PreStartAction of %s", util.NPU310PCardName)
	if tp == nil {
		return fmt.Errorf("%s handler not enabled: %s", util.NPU310PCardName, util.ArgumentError)
	}
	if ssn == nil {
		return fmt.Errorf("%s session is nil: %s", util.NPU310PCardName, util.ArgumentError)
	}
	reschEnable, ok := tp.SchedulerJobAttr.Label[rescheduling.JobRescheduleLabelKey]
	if !ok {
		klog.V(util.LogErrorLev).Infof("%s no re-scheduler key", util.NPU310PCardName)
		return nil
	}
	if reschEnable == rescheduling.JobOffRescheduleLabelValue {
		klog.V(util.LogInfoLev).Infof("%s RescheduleLabel not enabled", util.NPU310PCardName)
		return nil
	}
	tp.reHandle = rescheduling.New(&tp.ScheduleEnv, rescheduling.CmFaultJob310PKind)
	if tp.reHandle == nil {
		return fmt.Errorf("%s reSchedule not enabled: %s", util.NPU310PCardName, util.ArgumentError)
	}
	tp.reHandle.NewCommonReScheduler(rescheduling.CmFaultJob310PKind)
	tp.reHandle.SynCacheFaultNodeWithSession(util.NPU310PCardName)
	tp.reHandle.AddFaultNodeWithSession(util.NPU310PCardName)
	tp.reHandle.SynCacheFaultJobWithSession(ssn, util.NPU310PCardName, util.NPU310PCardNamePre)
	// 1. restart Fault Jobs that are recorded in cache
	if restartErr := tp.reHandle.RestartNeedForceDeleteJobs(ssn); restartErr != nil {
		klog.V(util.LogInfoLev).Infof("%s RestartNeedForceDeleteJobs: %s",
			util.NPU310PCardName, restartErr.Error())
	}
	// 2. get all the new 310P jobs in session
	runningJobs, getRunErr := tp.reHandle.GetRunningJobs(ssn, util.NPU310PCardName, util.ChipAcceleratorType)
	if getRunErr != nil {
		klog.V(util.LogInfoLev).Infof("%s GetRunningJobs: %s", util.NPU310PCardName, getRunErr.Error())
	}
	// 3. get nodes of session and fault jobs of 310P
	err := tp.reHandle.AddFaultJobWithSession(runningJobs, util.NPU310PCardName, util.NPU310PCardNamePre)
	if err != nil {
		klog.V(util.LogInfoLev).Infof("%s AddFaultJobWithSession", util.NPU310PCardName)
	}
	// 4. restart the fault jobs
	if restartErr := tp.reHandle.RestartFaultJobs(ssn); restartErr != nil {
		klog.V(util.LogInfoLev).Infof("%s RestartFaultJobs: %s", util.NPU310PCardName, restartErr.Error())
		return restartErr
	}
	return nil
}

// PreStopAction post-processing actions for re-scheduling
func (tp *ascend310P) PreStopAction(env *plugin.ScheduleEnv) error {
	klog.V(util.LogInfoLev).Infof("enter PreStopAction of %s...", util.NPU310PCardName)
	defer klog.V(util.LogInfoLev).Infof("leave PreStopAction of %s...", util.NPU310PCardName)
	if tp == nil || tp.reHandle == nil {
		return fmt.Errorf("%s reSchedule not enabled: %s", util.NPU310PCardName, util.ArgumentError)
	}
	if env == nil {
		return fmt.Errorf("%s env is nil: %s", util.NPU310PCardName, util.ArgumentError)
	}
	if err := tp.reHandle.WriteReSchedulerCacheToEnvCache(env, rescheduling.CmFaultJob310PKind); err != nil {
		return err
	}
	return nil
}
