/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package rescheduling is using for HuaWei Ascend pin fault rescheduling.
*/
package rescheduling

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
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
		klog.V(util.LogErrorLev).Infof(
			"GetJobElasticSchedulingLabel %s. %s no job reschedule label", value, job.Name)
		return JobOffRescheduleLabelValue
	}
	klog.V(util.LogInfoLev).Infof("GetJobElasticSchedulingLabel job: %s, label: %s", job.Name, value)
	return value
}

func (fJob *FaultJob) isJobGraceDeleteSuccess(jobInfo *api.JobInfo) bool {
	if jobInfo == nil {
		klog.V(util.LogErrorLev).Infof("jobInfo is nil: %#v", jobInfo)
		return false
	}
	if !plugin.IsJobInitial(jobInfo) {
		klog.V(util.LogInfoLev).Infof("isJobGraceDeletedSuccess: job %s not initialised.", jobInfo.Name)
		return false
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
			klog.V(util.LogInfoLev).Infof("pod %s restart success[new:%v---old:%v]",
				npuTask.Pod.Name, podCreateTimeCur, podCreateTimeRecord)
			restartNum++
		}
	}
	klog.V(util.LogDebugLev).Infof("<%d/%d> pod of job restarted", restartNum, len(jobInfo.Tasks))
	if restartNum >= len(jobInfo.Tasks) && plugin.IsJobInitial(
		jobInfo) { // minAvailable must equal to number of replicas
		klog.V(util.LogInfoLev).Infof("job %s grace delete success", jobInfo.Name)
		return true
	}
	klog.V(util.LogInfoLev).Infof("job %s not yet restarted", jobInfo.Name)
	return false
}

// CheckJobExistsInKubernetes check whether job recorded in cache can be traced in kubernetes
func (fJob *FaultJob) CheckJobExistsInKubernetes(ssn *framework.Session) bool {
	var existTaskNum int
	for _, fTask := range fJob.FaultTasks {
		klog.V(util.LogDebugLev).Infof("check task %s via client-go", fTask.TaskName)
		realPod, err := ssn.KubeClient().CoreV1().Pods(fTask.TaskNamespace).Get(
			context.TODO(), fTask.TaskName, v1.GetOptions{})
		if err != nil || realPod == nil {
			klog.V(util.LogInfoLev).Infof("pod %s not in kubernetes", fTask.TaskName)
			continue
		}
		existTaskNum += 1
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
func (fJob *FaultJob) GraceDeleteJob(ssn *framework.Session, npuJob *plugin.SchedulerJob, reason string) error {
	if fJob == nil {
		return fmt.Errorf("getJobFaultRescheduleLabel fJob object does not exist")
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
		if delErr := npuTask.ForceDeletePodByTaskInf(ssn); delErr != nil {
			klog.V(util.LogErrorLev).Infof("ForceDeletePodByTaskInf %s: %s.", npuTask.Name, delErr)
			continue
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

func (fJob *FaultJob) checkJobNodeRankIndexValid() bool {
	for _, fTask := range fJob.FaultTasks {
		if fTask.NodeRankIndex == "" {
			return false
		}
	}
	return true
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
		ElasticScheduling:   JobOffElasticScheduling,
	}
	return faultJob
}
