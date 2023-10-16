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
Package rescheduling is using for HuaWei Ascend pin fault rescheduling.
*/
package rescheduling

import (
	"context"
	"fmt"
	"strconv"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// GetJobFaultRescheduleLabel Get job's fault reschedule label.
func (fJob *FaultJob) GetJobFaultRescheduleLabel(job *plugin.SchedulerJob) string {
	if fJob == nil || job == nil {
		klog.V(util.LogErrorLev).Info(
			"GetJobFaultRescheduleLabel fJob or schedulerJob does not exist")
		return JobOffRescheduleLabelValue
	}
	value, ok := job.SchedulerJobAttr.Label[JobRescheduleLabelKey]
	if !ok {
		klog.V(util.LogInfoLev).Infof(
			"GetJobFaultRescheduleLabel %s. %s no job reschedule label", value, job.Name)
		return JobOffRescheduleLabelValue
	}
	klog.V(util.LogInfoLev).Infof("GetJobFaultRescheduleLabel job: %s, label: %s", job.Name, value)
	return value
}

// GetJobElasticSchedulingLabel get job's elastic scheduling label
func (fJob *FaultJob) GetJobElasticSchedulingLabel(job *plugin.SchedulerJob) string {
	if fJob == nil || job == nil {
		klog.V(util.LogErrorLev).Info(
			"GetJobElasticSchedulingLabel fJob or schedulerJob object does not exist")
		return JobOffElasticScheduling
	}
	value, ok := job.SchedulerJobAttr.Label[ElasticSchedulingKey]
	if !ok {
		klog.V(util.LogInfoLev).Infof(
			"GetJobElasticSchedulingLabel %s. %s no job reschedule label", value, job.Name)
		return JobOffRescheduleLabelValue
	}
	klog.V(util.LogInfoLev).Infof("GetJobElasticSchedulingLabel job: %s, label: %s", job.Name, value)
	return value
}

func (fJob *FaultJob) GetJobFaultNPUTaskNum() int {
	var count int
	for _, fTask := range fJob.FaultTasks {
		if len(fTask.UseCardName) > 0 {
			count++
		}
	}
	return count
}

func (fJob *FaultJob) isJobGraceDeleteSuccess(jobInfo *api.JobInfo) bool {
	restartNum := 0
	for _, fTask := range fJob.FaultTasks {
		podCreateTimeRecord := fTask.PodCreateTime // former pods create time
		npuTask, ok := jobInfo.Tasks[fTask.TaskUID]
		if !ok { // task not in job indicates it has been restarted
			restartNum++
			klog.V(util.LogInfoLev).Infof("%s in %s has been deleted", fTask.TaskName, jobInfo.Name)
			continue
		}
		podCreateTimeCur := npuTask.Pod.CreationTimestamp.Unix() // current pod create time
		if podCreateTimeCur != podCreateTimeRecord {
			klog.V(util.LogInfoLev).Infof("pod %s restart success[new:%v---old:%v]",
				npuTask.Pod.Name, podCreateTimeCur, podCreateTimeRecord)
			restartNum++
		}
	}

	klog.V(util.LogDebugLev).Infof("<%d/%d> pod of job restarted", restartNum, jobInfo.MinAvailable)
	return restartNum >= len(fJob.FaultTasks)
}

// CheckJobExistsInKubernetes check whether job recorded in cache can be traced in kubernetes
func (fJob *FaultJob) CheckJobExistsInKubernetes(ssn *framework.Session) bool {
	var existTaskNum int
	for _, fTask := range fJob.FaultTasks {
		klog.V(util.LogDebugLev).Infof("check task %s via client-go", fTask.TaskName)
		realPod, err := ssn.KubeClient().CoreV1().Pods(fTask.TaskNamespace).Get(
			context.TODO(), fTask.TaskName, metav1.GetOptions{})
		if err != nil || realPod == nil {
			klog.V(util.LogInfoLev).Infof("pod %s not in kubernetes", fTask.TaskName)
			continue
		}
		for _, ref := range realPod.GetOwnerReferences() {
			if ref.UID == fJob.UUID {
				existTaskNum += 1
			}
		}
		klog.V(util.LogDebugLev).Infof("task %s is in kubernetes", fTask.TaskName)
	}
	if existTaskNum > 0 {
		return true
	}
	return false
}

// ForceDeleteJob force delete jobs includes labelled force delete ones and grace delete failed ones
func (fJob *FaultJob) ForceDeleteJob(ssn *framework.Session, schedulerJob *plugin.SchedulerJob) error {
	klog.V(util.LogDebugLev).Infof("enter ForceDeleteJob")
	if fJob == nil || ssn == nil || schedulerJob == nil {
		return fmt.Errorf(
			"getJobFaultRescheduleLabel fJob object or ssn or schedulerJob does not exist")
	}
	for _, fTask := range fJob.FaultTasks {
		err := fTask.DeleteRealPodByTask(ssn, 0)
		if err != nil {
			klog.V(util.LogDebugLev).Infof("ForceDeleteFaultPod %s: %#v.", fTask.TaskName, err)
		}
	}
	return nil
}

// GraceDeleteJob grace delete jobs labelled to be deleted gracefully
func (fJob *FaultJob) GraceDeleteJob(ssn *framework.Session, npuJob *plugin.SchedulerJob) error {
	if fJob == nil {
		return fmt.Errorf("getJobFaultRescheduleLabel fJob object does not exist")
	}
	if ssn == nil {
		return fmt.Errorf("session does not exist")
	}
	if npuJob == nil {
		return fmt.Errorf("schedulerJob does not exist")
	}
	var reasonList []FaultReasonList
	for _, fTask := range fJob.FaultTasks {
		if fTask.Reason != nil {
			reasonList = append(reasonList, fTask.Reason...)
		}
	}
	reason := GetTaskRestartReason(reasonList)
	for _, fTask := range fJob.FaultTasks {
		npuTask, ok := npuJob.Tasks[fTask.TaskUID]
		if !ok {
			klog.V(util.LogDebugLev).Infof(
				"GraceDeleteJob: npuTask %s has been deleted in session.", fTask.TaskName)
			return fmt.Errorf("npuTask %s not in session", fTask.TaskName)
		}
		if delErr := npuTask.ForceDeletePodByTaskInf(ssn, reason, fTask.NodeName); delErr != nil {
			klog.V(util.LogErrorLev).Infof("ForceDeletePodByTaskInf %s: %s.", npuTask.Name, delErr)
		}
	}
	return nil
}

func (fJob *FaultJob) restartSingleFaultJob(ssn *framework.Session,
	kubeClient kubernetes.Interface, schedulerJob *plugin.SchedulerJob) error {

	// delete jobs
	var deleteErr error

	switch fJob.ReScheduleKey {
	case JobForceRescheduleLabelValue:
		deleteErr = fJob.ForceDeleteJob(ssn, schedulerJob)
	case JobGraceRescheduleLabelValue:
		deleteErr = fJob.GraceDeleteJob(ssn, schedulerJob)
	case JobOffRescheduleLabelValue:
		deleteErr = fmt.Errorf("job reschedule %s", fJob.ReScheduleKey)
	default:
		deleteErr = fmt.Errorf("not support %s to reschedule job", fJob.ReScheduleKey)
	}

	return deleteErr
}

func (fJob *FaultJob) isJobInSession(jobs map[api.JobID]plugin.SchedulerJob) bool {
	_, ok := jobs[fJob.JobUID]
	return ok
}

func (fJob *FaultJob) jobInfoInSession(jobs map[api.JobID]*api.JobInfo) *api.JobInfo {
	if fJob.ElasticScheduling == JobOnElasticScheduling {
		for _, job := range jobs {
			// consider elastic scheduling which lead job uid/name change
			if job.Namespace == fJob.JobNamespace &&
				(job.Name == fJob.JobName || util.ReferenceNameOfJob(job) == fJob.ReferenceName) {
				return job
			}
		}
		return nil
	}

	job, ok := jobs[fJob.JobUID]
	if ok && util.UuidOfJob(job) == fJob.UUID {
		return job
	}

	return nil
}

func (fJob *FaultJob) getJobUseNodes() []string {
	jobUseNodes := make([]string, len(fJob.FaultTasks))
	for i, fTask := range fJob.FaultTasks {
		jobUseNodes[i] = fTask.NodeName
	}
	return jobUseNodes
}

func (fJob *FaultJob) getIsFaultJob() bool {
	for _, task := range fJob.FaultTasks {
		if task.IsFaultTask {
			klog.V(util.LogDebugLev).Infof(
				"job %s isFaultJob set true because of having fault tasks", fJob.JobName)
			return true
		}
	}
	return false
}

func (fJob *FaultJob) checkJobNodeRankIndexValid() bool {
	for _, fTask := range fJob.FaultTasks {
		if fTask.NodeRankIndex == "" {
			return false
		}
	}
	return true
}

func (fJob *FaultJob) IsJobHasPreSeparateNPUKey() bool {
	if fJob == nil {
		return false
	}
	for _, fTask := range fJob.FaultTasks {
		for _, reason := range fTask.Reason {
			if reason.LargeModelFaultLevel == PreSeparateNPU {
				return true
			}
		}
	}
	return false
}

func (fJob *FaultJob) IsJobFaultTypeHasCardUnhealthy() bool {
	if fJob == nil {
		return false
	}
	for _, fTask := range fJob.FaultTasks {
		for _, reason := range fTask.Reason {
			if reason.FaultType == NodeCardUnhealthy {
				return true
			}
		}
	}
	return false
}

func (fJob *FaultJob) setJobFaultReScheduleLabel(value string) {
	fJob.ReScheduleKey = value
}

func (fJob *FaultJob) setJobElasticReScheduleLabel(value string) {
	fJob.ElasticScheduling = value
}

func (fJob *FaultJob) setFaultTasks(value []FaultTask) {
	fJob.FaultTasks = value
}

func (fJob *FaultJob) setJobRankIds(value []string) {
	fJob.JobRankIds = value
}

func (fJob *FaultJob) setNodeNames(value []string) {
	fJob.NodeNames = value
}

func (fJob *FaultJob) setIsFaultJob(value bool) {
	fJob.IsFaultJob = value
}

func newFaultJobDefault(job *api.JobInfo, updateTime int64) FaultJob {
	faultJob := FaultJob{
		ReScheduleKey:       JobOffRescheduleLabelValue, // off/grace/force
		IsInSession:         true,
		JobName:             job.Name,
		JobUID:              job.UID,
		JobNamespace:        job.Namespace,
		UpdateTime:          updateTime,
		JobRankIdCreateTime: updateTime,
		FaultTypes:          make([]string, 0),
		ElasticScheduling:   JobOffElasticScheduling,
		ReferenceName:       util.ReferenceNameOfJob(job),
		UUID:                util.UuidOfJob(job),
		FaultRetryTimes:     faultRetryTimeOfJob(job),
	}
	return faultJob
}

func faultRetryTimeOfJob(job *api.JobInfo) int {
	value, ok := job.PodGroup.Labels[FaultRetryTimesKey]
	if !ok {
		klog.V(util.LogInfoLev).Infof("get job<%s> label<%s> failed", job.UID, FaultRetryTimesKey)
		return 0
	}
	v, err := strconv.Atoi(value)
	if err != nil {
		klog.V(util.LogInfoLev).Infof("Failed to convert fault-retry-times <%s> of job <%s> into number",
			value, job.UID)
		return 0
	}
	klog.V(util.LogInfoLev).Infof("get job: %s, fault-retry-times: %s", job.Name, value)
	return v
}
