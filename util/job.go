/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package util is using for the total variable.
*/
package util

import (
	"strings"

	"k8s.io/klog"
	"volcano.sh/apis/pkg/apis/scheduling"
	"volcano.sh/volcano/pkg/scheduler/api"
)

// For Job type
const (
	// JobTypeWhole whole card  for NPU  Job.
	JobTypeWhole = iota
	// JobTypeDyCut Dynamic force cutting NPU vJob.
	JobTypeDyCut
	// JobTypeStCut Static force segmentation NPU Job.
	JobTypeStCut
	// JobTypeUnknown unknown NPU Job type.
	JobTypeUnknown
)

// VJob for dynamic NPU Job.
type VJob struct {
	// type: JobTypeWhole, JobTypeDycut, JobTypeStcut.
	Type   int
	Status scheduling.PodGroupPhase
}

// NPUJob only npu vcJob have.
type NPUJob struct {
	// the mapKey is taskID, not Name.
	Tasks      map[api.TaskID]NPUTask
	ReqNPUName string
	ReqNPUNum  int
	*VJob
}

// ComJob all vcJob has.
type ComJob struct {
	Name      api.JobID
	NameSpace string
	Selector  map[string]string
	Label     map[string]string
}

// SchedulerJobAttr vcJob's attribute.
type SchedulerJobAttr struct {
	ComJob
	*NPUJob
}

// IsSelectorMeetJob check the selectors
func IsSelectorMeetJob(jobSelectors, conf map[string]string) bool {
	for jobKey, jobValue := range jobSelectors {
		confValue, confOk := conf[jobKey]
		if !confOk {
			klog.V(LogErrorLev).Infof("conf has no job selector key:%s.", jobKey)
			return false
		}

		if !strings.Contains(confValue, jobValue) {
			klog.V(LogErrorLev).Infof("conf has no job selector value:%s.", jobValue)
			return false
		}
	}
	return true
}

// IsVJob Determine whether is the NPU virtual job.
// Dynamic segmentation: huawei.com/npu-core.
// static segmentation: huawei.com/Ascend910-Y.
// no segmentation: huawei.com/Ascend910.
func (nJob *NPUJob) IsVJob() bool {
	if len(strings.Split(nJob.ReqNPUName, "-")) > 1 {
		return true
	}
	return false
}

func (nJob *NPUJob) setVJobType() {
	tmpTypes := make(map[int]struct{}, MapInitNum)
	for _, vTask := range nJob.Tasks {
		nJob.Type = vTask.Type
		if len(nJob.Tasks) == 1 {
			return
		}
		tmpTypes[vTask.Type] = struct{}{}
	}
	// all vTask type in job must be same.
	if len(tmpTypes) != 1 {
		nJob.Type = JobTypeUnknown
		return
	}
}

// SetVJobType set virtual job type
// must be used after isVJob.
// for all chips: 910, 310P
func (nJob *NPUJob) SetVJobType() {
	if nJob == nil {
		return
	}
	nJob.setVJobType()
}

func (nJob *NPUJob) SetVJobStatusByInf(vcJob *api.JobInfo) {
	if nJob == nil {
		return
	}
	nJob.Status = vcJob.PodGroup.Status.Phase
}
