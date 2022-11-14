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

func (action *Action) Destroy(vCache *VCache, ssn *framework.Session, env *plugin.ScheduleEnv) {
	action.dealAllocated(vCache, env, ssn)
	action.dealDestroying(vCache, env)
}

func (action *Action) dealAllocated(vCache *VCache, env *plugin.ScheduleEnv, ssn *framework.Session) {
	klog.V(util.LogInfoLev).Infof("Start dealing with %s vJobs", VJobStatusAllocated)
	defer klog.V(util.LogInfoLev).Infof("Finish dealing with %s vJobs", VJobStatusAllocated)
	for jobUID, vJob := range vCache.vJobs {
		if !action.stateTransferAllocated(vJob, env, ssn) {
			klog.V(util.LogDebugLev).Infof("vJob %s not meet state transfer forward conditions", jobUID)
		}
		vJob.setVJobStatus(VJobStatusDestroying)
		vCache.addOrUpdateVJobToCache(vJob)
	}
}

func (action *Action) stateTransferAllocated(vJob VJob, env *plugin.ScheduleEnv, ssn *framework.Session) bool {
	schedulerJob, ok := env.Jobs[vJob.jobUID]
	if !ok || (vJob.jobStatus != VJobStatusAllocated) {
		klog.V(util.LogDebugLev).Infof("vJob %s not in session or status not %s",
			vJob.jobUID, VJobStatusAllocated)
		return false
	}
	if vJob.IsJobNeedDeleting(schedulerJob, ssn) {
		return true
	}
	return false
}

func (action *Action) dealDestroying(vCache *VCache, env *plugin.ScheduleEnv) {
	klog.V(util.LogInfoLev).Infof("Start dealing with %s vJobs", VJobStatusDestroying)
	defer klog.V(util.LogInfoLev).Infof("Finish dealing with %s vJobs", VJobStatusDestroying)
	for jobUID, vJob := range vCache.vJobs {
		if !action.stateTransferDestroying(vJob, env) {
			klog.V(util.LogDebugLev).Infof("vJob %s not meet state transfer forward conditions", jobUID)
			continue
		}
		delete(vCache.vJobs, jobUID)
	}
}

func (action *Action) stateTransferDestroying(vJob VJob, env *plugin.ScheduleEnv) bool {
	npuNode, ok := env.Nodes[vJob.reqNodeName]
	if !ok || (vJob.jobStatus != VJobStatusDestroying) {
		klog.V(util.LogDebugLev).Infof("vJob %s used node %s not in session or vJob status not %s", vJob.jobUID,
			vJob.reqNodeName, VJobStatusDestroying)
		return false
	}
	if !vJob.isVJobInDeviceInfo(npuNode) {
		klog.V(util.LogInfoLev).Infof("vJob %s node %s not in device list, vJob can be removed", vJob.jobUID,
			vJob.reqNodeName)
		return true
	}
	return false
}
