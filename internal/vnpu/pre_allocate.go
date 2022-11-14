/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package rescheduling is using for HuaWei Ascend pin fault rescheduling.

*/
package vnpu

import (
	"fmt"
	"k8s.io/klog"
	"time"

	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

func (action *Action) Initialize(template map[string]VResource) {
	action.template = template
}

// PreAlloc pre-allocate actions for vnpu jobs
func (action *Action) PreAlloc(ssn *framework.Session, env *plugin.ScheduleEnv, vCache *VCache,
	vNodes map[string]VNode) {
	klog.V(util.LogInfoLev).Infof("Enter VNPU action...")
	var vJobsToBePreAlloc []VJob
	klog.V(util.LogInfoLev).Infof("Deal with %s vJobs", VJobStatusNotHandled)
	for _, vJob := range vCache.vJobs {
		schedulerJob, ok := env.Jobs[vJob.jobUID]
		if !ok {
			klog.V(util.LogWarningLev).Infof("vJob %s not in session", vJob.jobUID)
		}
		if vJob.jobStatus == VJobStatusNotHandled && vJob.IsJobReadyForPreAlloc(schedulerJob){  // 1.
			// unhandled job transferred into notPreAllocated job
			vJob.setVJobStatus(VJobStatusNotPreSegmented)
			vCache.AddOrUpdateVJobToCache(vJob)
		}
		if vJob.jobStatus == VJobStatusNotPreSegmented {
			vJobsToBePreAlloc = append(vJobsToBePreAlloc, vJob)  // 2. put not preAllocated jobs into allocation queue
		}
	}
	// 3. deal with jobs of status VJobStatusNotPreSegmented
	klog.V(util.LogInfoLev).Infof("Deal with %s vJobs", VJobStatusNotPreSegmented)
	vJobsToBePreAllocSorted := vCache.sortNeedAllocVJobsByAllocTime(vJobsToBePreAlloc)
	for _, vJob := range vJobsToBePreAllocSorted {
		allocNode, allocChip, err := action.preAllocateVJobToNodeAndChip(&vJob, vNodes) // 3.1 find the node and chip
		if err != nil || allocChip == nil || allocNode == nil {
			klog.V(util.LogWarningLev).Infof("") // todo
			continue
		}
		allocNode.ExcludeResFromVNode(vJob.resourceReq) // 3.2 update node and chip resource
		allocChip.ExcludeResFromVNode(vJob.resourceReq)
		vJob.RecordVJobPreSegmentInfo(allocNode, allocChip) // 3.3 update alloc info to vJob
		vJob.setVJobStatus(VJobStatusPreSegmented)
		vCache.AddOrUpdateVJobToCache(vJob) // 3.4 update vJob to vCache
	}
	// 4. deal with jobs of status VJobStatusPreSegmented
	klog.V(util.LogInfoLev).Infof("Deal with %s vJobs", VJobStatusPreSegmented)
	for _, vJob := range vCache.vJobs {
		if vJob.jobStatus != VJobStatusPreSegmented {
			continue
		}
		_, ok := env.Jobs[vJob.jobUID]
		if !ok {
			klog.V(util.LogWarningLev).Infof("vJob %s not in session", vJob.jobUID)
		}
		if time.Now().Unix() - vJob.updateTime > 10 { // 4.1 job is not overtime(segmentation not yet failed)
			allocNode, ok := vNodes[vJob.reqNodeName]
			if !ok {
				klog.V(util.LogWarningLev).Infof("vJob %s's preAllocated node %s not in vNodes", vJob.jobUID, vJob.reqNodeName)
			}
			allocChip, ok := allocNode.nodeChips[vJob.reqCardName]
			if !ok {
				klog.V(util.LogWarningLev).Infof("vJob %s's preAllocated chip %s not in vNode %s's vChips",
					vJob.jobUID, vJob.reqCardName, vJob.reqNodeName)
			}
			allocNode.AddResToVNode(vJob.resourceReq)
			allocChip.AddResToVChip(vJob.resourceReq)
			vJob.ClearVJobPreSegmentInfo()
			vJob.setVJobStatus(VJobStatusNotPreSegmented)
			vCache.AddOrUpdateVJobToCache(vJob)
		}
	}
}

func (action *Action) preAllocateVJobToNodeAndChip(vJob *VJob, vNodes map[string]VNode) (*VNode, *VChip, error) {
	if !action.IsClusterResourceMeetJobReq(vJob, vNodes) {
		return nil, nil, fmt.Errorf("")  // todo
	}
	sortedVNodes := action.SortVNodes(vNodes)
	for _, vNode := range sortedVNodes {
		if !vNode.IsNodeStable() {
			continue
		}
		if !action.IsVNodeResourceMeetJobReq(vJob, vNode) {
			continue
		}
		for _, vChip := range vNode.nodeChips {
			if !vChip.IsChipStable() {
				continue
			}
			if !action.IsVChipResourceMeetJobReq(vJob, vChip) {
				continue
			}
			return &vNode, &vChip, nil
		}
	}
	return nil, nil, fmt.Errorf("") // todo
}

// IsClusterResourceMeetJobReq todo
func (action *Action) IsClusterResourceMeetJobReq(vJob *VJob, vNodes map[string]VNode) bool {
	return true
}

// IsVNodeResourceMeetJobReq todo
func (action *Action) IsVNodeResourceMeetJobReq(vJob *VJob, vNode VNode) bool {
	return true
}

// IsVChipResourceMeetJobReq todo
func (action *Action) IsVChipResourceMeetJobReq(vJob *VJob, vChip VChip) bool {
	return true
}

func (action *Action) SortVNodes(vNodes map[string]VNode) []VNode {

}