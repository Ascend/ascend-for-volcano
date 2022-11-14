/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package vnpu is using for HuaWei Ascend pin fault rescheduling.

*/
package vnpu

import (
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// BackFill backFill actions for VNPU jobs
func (action *Action) BackFill(vCache *VCache, ssn *framework.Session, env *plugin.ScheduleEnv) {
	action.dealPreSegmentedForward(vCache, env, ssn)
}

func (action *Action) dealPreSegmentedForward(vCache *VCache, env *plugin.ScheduleEnv, ssn *framework.Session) {
	klog.V(util.LogInfoLev).Infof("Start dealing with %s (forward) vJobs", VJobStatusPreSegmented)
	defer klog.V(util.LogInfoLev).Infof("Finish dealing with %s (forward) vJobs", VJobStatusPreSegmented)
	for jobUID, vJob := range vCache.vJobs {
		if !action.stateTransferPreSegmentedBackward(vJob, env) {
			continue
		}
		ssnJob, ok := ssn.Jobs[jobUID]
		if !ok {
			klog.V(util.LogWarningLev).Infof("vJob %s not in session, skip", vJob.jobUID)
		}
		allocCardName, err := vJob.getVJobUsedNPUNames(ssnJob)
		if err != nil || allocCardName == "" {
			klog.V(util.LogWarningLev).Infof("vJob %s alloc card read from pod annotation failed", vJob.jobUID)
		}
		vJob.setVJobAllocInfo(allocCardName)
		vJob.setVJobStatus(VJobStatusAllocated)
		vCache.addOrUpdateVJobToCache(vJob)
	}
}

func (action *Action) stateTransferPreSegmentedForward(vJob VJob, env *plugin.ScheduleEnv) bool {
	schedulerJob, ok := env.Jobs[vJob.jobUID]
	if !ok {
		klog.V(util.LogWarningLev).Infof("vJob %s not in session, skip", vJob.jobUID)
		return false
	}
	if (vJob.jobStatus != VJobStatusPreSegmented) && (vJob.jobStatus != VJobStatusSegmented) {
		klog.V(util.LogInfoLev).Infof("vJob %s not in %s or %s status, skip read pod annotation",
			vJob.jobUID, VJobStatusPreSegmented, VJobStatusSegmented)
		return false
	}
	if !vJob.IsJobAllocated(schedulerJob) {
		klog.V(util.LogInfoLev).Infof("vJob %s is not running, skip read pod annotation", vJob.jobUID)
		return false
	}
	return true
}
