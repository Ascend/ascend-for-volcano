/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package rescheduling is using for HuaWei Ascend pin fault rescheduling.

*/
package rescheduling

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// GetJobFaultRescheduleLabel Get job's fault reschedule label.
func (fJob *FaultJob) GetJobFaultRescheduleLabel(job *plugin.SchedulerJob) string {
	if fJob == nil {
		klog.V(util.LogErrorLev).Info(
			"GetJobFaultRescheduleLabel fJob object does not exist")
		return JobOffRescheduleLabelValue
	}
	if job == nil {
		klog.V(util.LogErrorLev).Info(
			"GetJobFaultRescheduleLabel SchedulerJob does not exist")
		return JobOffRescheduleLabelValue
	}
	value, ok := job.SchedulerJobAttr.Label[JobRescheduleLabelKey]
	if !ok {
		klog.V(util.LogErrorLev).Infof(
			"GetJobFaultRescheduleLabel %s. %s no job reschedule label", value, job.JobName)
		return JobOffRescheduleLabelValue
	}

	return value
}

func (fJob *FaultJob) isJobGraceDeleteSuccess(jobInfo *api.JobInfo) bool {
	// 1. judge if job task exists
	if len(jobInfo.Tasks) == 0 {
		// old pod has been deleted.
		klog.V(util.LogInfoLev).Infof("isJobGraceDeletedSuccess: %v pods has been deleted.", jobInfo.Name)
		return true
	}
	// 2. judge if the create time of pods in cache and current pod differs
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
			klog.V(util.LogInfoLev).Infof("pod restart success[new:%v---old:%v]",
				podCreateTimeCur, podCreateTimeRecord)
			restartNum++
		}
	}

	if restartNum == len(jobInfo.Tasks) {
		klog.V(util.LogInfoLev).Infof("job all pod %d restart success.", restartNum)
		return true
	}
	return false
}

// ForceDeleteJob force delete jobs includes labelled force delete ones and grace delete failed ones
func (fJob *FaultJob) ForceDeleteJob(ssn *framework.Session, schedulerJob *plugin.SchedulerJob) error {
	klog.V(util.LogDebugLev).Infof("enter ForceDeleteJob")
	if fJob == nil {
		return fmt.Errorf(
			"getJobFaultRescheduleLabel fJob object does not exist")
	}
	if ssn == nil {
		return fmt.Errorf("session does not exist")
	}
	if schedulerJob == nil {
		return fmt.Errorf("schedulerJob does not exist")
	}
	for _, fTask := range fJob.FaultTasks {
		npuTask, ok := schedulerJob.Tasks[string(fTask.TaskUID)]
		if !ok {
			klog.V(util.LogDebugLev).Infof(
				"ForceDeleteJob: npuTask %s has been deleted in session.", fTask.TaskName)
		}
		err := npuTask.DeleteRealPodByTask(ssn)
		if err != nil {
			klog.V(util.LogDebugLev).Infof("ForceDeleteFaultPod %s: %#v.", npuTask.TaskName, err)
		}
	}
	return nil
}

// GraceDeleteJob grace delete jobs labelled to be deleted gracefully
func (fJob *FaultJob) GraceDeleteJob(ssn *framework.Session, npuJob *plugin.SchedulerJob, reason string) error {
	if fJob == nil {
		return fmt.Errorf(
			"getJobFaultRescheduleLabel fJob object does not exist")
	}
	if ssn == nil {
		return fmt.Errorf("session does not exist")
	}
	if npuJob == nil {
		return fmt.Errorf("schedulerJob does not exist")
	}
	for _, fTask := range fJob.FaultTasks {
		npuTask, ok := npuJob.Tasks[string(fTask.TaskUID)]
		if !ok {
			klog.V(util.LogDebugLev).Infof(
				"GraceDeleteJob: npuTask %s has been deleted in session.", fTask.TaskName)
			return fmt.Errorf("npuTask %s not in session", fTask.TaskName)
		}
		if err := npuTask.EvictJobByTask(ssn, reason, fTask.TaskName); err != nil {
			return err
		}
	}
	return nil
}

func (fJob *FaultJob) restartSingleFaultJob(ssn *framework.Session,
	kubeClient kubernetes.Interface, schedulerJob *plugin.SchedulerJob, obj interface{}) error {
	// 1. get pod pending reason
	var reason string
	switch para := obj.(type) {
	case string:
		reason = para
	case map[api.TaskID]*api.FitErrors:
		for _, nodeErrors := range para {
			reason += nodeErrors.Error()
		}
	case error:
		reason = para.Error()
	default:
		// other type are not allowed
		return fmt.Errorf("aseert reason(%T) failed", reason)
	}

	// delete jobs
	var deleteErr error
	switch fJob.ReScheduleKey {
	case JobForceRescheduleLabelValue:
		deleteErr = fJob.ForceDeleteJob(ssn, schedulerJob)
	case JobGraceRescheduleLabelValue:
		deleteErr = fJob.GraceDeleteJob(ssn, schedulerJob, reason)
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

func (fJob *FaultJob) setJobFaultReScheduleLabel(value string) {
	fJob.ReScheduleKey = value
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

func newFaultJobDefault(jobName, jobNamespace string, jobUID api.JobID, updateTime int64) FaultJob {
	faultJob := FaultJob{
		ReScheduleKey:       JobOffRescheduleLabelValue, // off/grace/force
		IsFaultJob:          false,
		IsInSession:         true,
		JobName:             jobName,
		JobUID:              jobUID,
		JobNamespace:        jobNamespace,
		JobRankIds:          nil,
		NodeNames:           nil,
		FaultTasks:          nil,
		UpdateTime:          updateTime,
		JobRankIdCreateTime: updateTime,
		FaultTypes:          nil,
		DeleteExecutedFlag:  false,
	}
	return faultJob
}
