/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package ascend310p is using for HuaWei 310P Ascend pin affinity schedule.
*/
package ascend310p

import (
	"fmt"
	"k8s.io/klog"

	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/rescheduling"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

func (tp *ascend310P) preStartRescheduling(ssn *framework.Session) error {
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
