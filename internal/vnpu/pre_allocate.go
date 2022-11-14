/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package vnpu is using for HuaWei Ascend pin fault rescheduling.

*/
package vnpu

import (
	"fmt"
	"k8s.io/klog"
	"time"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// Initialize initialize template
func (action *Action) Initialize(template map[string]VResource) {
	action.template = template
}

// PreAlloc pre-allocate actions for vnpu jobs
func (action *Action) PreAlloc(env *plugin.ScheduleEnv, vCache *VCache, vNodes map[string]VNode) {
	klog.V(util.LogInfoLev).Infof("Enter VNPU action preAlloc...")
	defer klog.V(util.LogInfoLev).Infof("Leave VNPU action preAlloc...")
	action.dealUnhandled(vCache, env)
	action.dealNotPreSegmented(vCache, vNodes, env)
	action.dealPreSegmented(vCache, env, vNodes)
}

func (action *Action) dealUnhandled(vCache *VCache, env *plugin.ScheduleEnv) {
	klog.V(util.LogInfoLev).Infof("Start dealing with %s vJobs", VJobStatusUnhandled)
	defer klog.V(util.LogInfoLev).Infof("Finish dealing with %s vJobs", VJobStatusUnhandled)
	for jobUID, vJob := range vCache.vJobs {
		if !action.stateTransferUnhandled(vJob, env) { // unhandled job doesn't meet state transfer condition
			klog.V(util.LogDebugLev).Infof("vJob %s not meet state transfer forward conditions", jobUID)
			continue
		}
		klog.V(util.LogDebugLev).Infof("update vJob %s job status", jobUID)
		vJob.setVJobStatus(VJobStatusNotPreSegmented) // update job state
		vCache.addOrUpdateVJobToCache(vJob)           // update cache
	}
}

func (action *Action) stateTransferUnhandled(vJob VJob, env *plugin.ScheduleEnv) bool {
	schedulerJob, ok := env.Jobs[vJob.jobUID]
	if !ok {
		klog.V(util.LogWarningLev).Infof("vJob %s not in session", vJob.jobUID)
		return false
	}
	return vJob.jobStatus == VJobStatusUnhandled && vJob.IsJobReadyForPreAlloc(schedulerJob)
}

func (action *Action) dealNotPreSegmented(vCache *VCache, vNodes map[string]VNode, env *plugin.ScheduleEnv) {
	// deal with jobs of status VJobStatusNotPreSegmented
	klog.V(util.LogInfoLev).Infof("Start dealing with %s vJobs", VJobStatusNotPreSegmented)
	defer klog.V(util.LogInfoLev).Infof("Finish dealing with %s vJobs", VJobStatusNotPreSegmented)
	// 1. get and sort notPreSegmented jobs
	vJobsToBePreAllocSorted := vCache.GetAndSortNotPreAllocVJobsFromCache(env)
	// 2. alloc vJobs to node and chip
	for _, vJob := range vJobsToBePreAllocSorted {
		allocNode, allocChip, err := action.preAllocateVJob(&vJob, vNodes) // 2.1 find the node and chip
		if err != nil {
			klog.V(util.LogWarningLev).Infof("pre-allocate vJob %s failed: %s", vJob.jobUID, err.Error())
			continue
		}
		action.subtractJobResFromDevice(vJob, allocNode, allocChip)
		vJob.recordVJobPreSegmentInfo(allocNode, allocChip) // 2.3 update alloc info to vJob
		vJob.setVJobStatus(VJobStatusPreSegmented)
		vCache.addOrUpdateVJobToCache(vJob) // 2.4 update vJob to vCache
	}
}

func (action *Action) subtractJobResFromDevice(vJob VJob, allocNode *VNode, allocChip *VChip) {
	allocNode.ExcludeResFromVNode(vJob.resourceReq) // 2.2 update node and chip resource
	allocChip.ExcludeResFromVChip(vJob.resourceReq)
}

func (action *Action) dealPreSegmented(vCache *VCache, env *plugin.ScheduleEnv, vNodes map[string]VNode) {
	// 4. deal with jobs of status VJobStatusPreSegmented
	klog.V(util.LogInfoLev).Infof("Start dealing with %s (backward) vJobs", VJobStatusPreSegmented)
	defer klog.V(util.LogInfoLev).Infof("Finish dealing with %s (backward) vJobs", VJobStatusPreSegmented)
	for jobUID, vJob := range vCache.vJobs {
		if !action.stateTransferPreSegmentedBackward(vJob, env) {
			klog.V(util.LogInfoLev).Infof("vJob %s not in %s state or not in session or not overtime, skip",
				vJob.jobUID, VJobStatusPreSegmented)
			continue
		}
		// unset the pre-alloc info
		if err := action.addFailedPreAllocJobResBackToDevice(vJob, vNodes); err != nil {
			klog.V(util.LogErrorLev).Infof("vJob %s device resource return failed: {%s}", jobUID, err.Error())
		}
		vJob.clearVJobPreSegmentInfo()
		vJob.setVJobStatus(VJobStatusNotPreSegmented)
		vCache.addOrUpdateVJobToCache(vJob)
	}
}

func (action *Action) addFailedPreAllocJobResBackToDevice(vJob VJob, vNodes map[string]VNode) error {
	// todo: if vNode or vChip not found, strategy?
	allocNode, ok := vNodes[vJob.reqNodeName]
	if !ok {
		return fmt.Errorf("vJob %s's preAllocated node %s not in vNodes", vJob.jobUID, vJob.reqNodeName)
	}
	allocNode.AddResToVNode(vJob.resourceReq)
	allocChip, ok := allocNode.nodeChips[vJob.reqCardName]
	if !ok {
		return fmt.Errorf("vJob %s's preAllocated chip %s not in vNode %s's vChips", vJob.jobUID,
			vJob.reqCardName, vJob.reqNodeName)
	}
	allocChip.AddResToVChip(vJob.resourceReq)
	return nil
}

func (action *Action) stateTransferPreSegmentedBackward(vJob VJob, env *plugin.ScheduleEnv) bool {
	_, ok := env.Jobs[vJob.jobUID]
	elapsedTime := time.Now().Unix() - vJob.updateTime
	return ok && vJob.jobStatus == VJobStatusSegmented && elapsedTime > PreAllocateFailureWaitTime
}

func (action *Action) preAllocateVJob(vJob *VJob, vNodes map[string]VNode) (*VNode, *VChip, error) {
	if !action.IsClusterResourceMeetJobReq(vJob, vNodes) { // 1. judge cluster resource enough
		return nil, nil, fmt.Errorf("cluster resource does not meet vJob %s resource requirements", vJob.jobUID)
	}
	sortedVNodes := action.SortVNodes(vNodes) // 2.1 sort vNodes
	for _, vNode := range sortedVNodes {
		if !vNode.IsNodeStable() { // 2.2 judge vNode stable
			klog.V(util.LogDebugLev).Infof("node %s resource not stable", vNode.nodeName)
			continue
		}
		if !action.IsVNodeResourceMeetJobReq(vJob, vNode) { // 2.3 judge vNode resource enough
			klog.V(util.LogDebugLev).Infof("node %s resource does not meet requirements", vNode.nodeName)
			continue
		}
		for _, vChip := range vNode.nodeChips {
			if !vChip.IsChipStable() { // 2.4.1 judge vChip stable
				klog.V(util.LogDebugLev).Infof("chip %s resource not stable", vChip.cardName)
				continue
			}
			if !action.IsVChipResourceMeetJobReq(vJob, vChip) { // 2.4.2 judge vChip resource enough
				klog.V(util.LogDebugLev).Infof("chip %s resource does not meet requirements", vChip.cardName)
				continue
			}
			klog.V(util.LogInfoLev).Infof("vJob %s device allocated success, vNode {%s}, vChip {%s}",
				vJob.jobUID, vNode.nodeName, vChip.cardName)
			return &vNode, &vChip, nil
		}
	}
	return nil, nil, fmt.Errorf("non device to be allocated to vJov %s", vJob.jobUID)
}

// IsClusterResourceMeetJobReq todo: function to be completed
func (action *Action) IsClusterResourceMeetJobReq(vJob *VJob, vNodes map[string]VNode) bool {
	return true
}

// IsVNodeResourceMeetJobReq todo: function to be completed
func (action *Action) IsVNodeResourceMeetJobReq(vJob *VJob, vNode VNode) bool {
	return true
}

// IsVChipResourceMeetJobReq todo: function to be completed
func (action *Action) IsVChipResourceMeetJobReq(vJob *VJob, vChip VChip) bool {
	return true
}

// SortVNodes todo: function to be completed
func (action *Action) SortVNodes(vNodes map[string]VNode) []VNode {
	var sortedVNodes []VNode
	return sortedVNodes
}
