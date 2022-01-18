/*
Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package plugin is using for HuaWei Ascend pin affinity schedule frame.

*/
package plugin

import (
	"context"
	"errors"
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	"strings"
	"volcano.sh/apis/pkg/apis/scheduling"
	"volcano.sh/volcano/pkg/cli/vjobs"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/conf"
	"volcano.sh/volcano/pkg/scheduler/framework"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/rescheduling"
	hwutil "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/util"
)

func getJobHandle(obj interface{}) *api.JobInfo {
	job, ok := obj.(*api.JobInfo)
	if !ok {
		klog.V(logErrorLev).Infof("job valid Failed to convert <%v> to *JobInfo.", obj)
		return nil
	}

	return job
}

func updatePodGroupPendingReason(ssn *framework.Session, job *api.JobInfo, reason string) {
	jc := scheduling.PodGroupCondition{
		Type:               scheduling.PodGroupUnschedulableType,
		Status:             v1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		TransitionID:       string(ssn.UID),
		Reason:             scheduling.NotEnoughResourcesReason,
		Message:            reason,
	}

	for k, value := range job.PodGroup.Status.Conditions {
		if value.Message == jc.Message {
			job.PodGroup.Status.Conditions[k].LastTransitionTime = jc.LastTransitionTime
			job.PodGroup.Status.Conditions[k].TransitionID = jc.TransitionID
			return
		}
	}

	job.PodGroup.Status.Conditions = append(job.PodGroup.Status.Conditions, jc)
}

func updatePodsPendingReason(job *api.JobInfo, reason string) {
	for _, task := range job.Tasks {
		updatePodPendingReason(task, reason)
	}
}

func updateJobPendingReason(ssn *framework.Session, job *api.JobInfo, reason interface{}) error {
	var flag = false
	var reasonTmp string

	// for set job not meet case
	jobError, jobOk := reason.(string)
	if jobOk {
		// job failed
		job.JobFitErrors = jobError
		reasonTmp = jobError
		flag = true
	}
	// for set nodes not meet case
	nodeErrors, nodeOk := reason.(map[api.TaskID]*api.FitErrors)
	if nodeOk {
		job.NodesFitErrors = nodeErrors
		for _, nodeErrors := range nodeErrors {
			reasonTmp += nodeErrors.Error()
		}
		flag = true
	}
	// other type are not allowed
	if !flag {
		return fmt.Errorf("assert reason(%T) failed", reason)
	}
	// for write pending reason into pod
	updatePodsPendingReason(job, reasonTmp)
	// for write pending reason into vcjob
	updatePodGroupPendingReason(ssn, job, reasonTmp)

	return nil
}

// SetJobPendingReason to set job failed and add failed reason
func SetJobPendingReason(ssn *framework.Session, obj interface{}, reason interface{}) error {
	job := getJobHandle(obj)
	if job == nil {
		message := fmt.Errorf("getJobHandle [%v] failed", obj)
		klog.V(logErrorLev).Infof("%v.", message)
		return message
	}

	if err := updateJobPendingReason(ssn, job, reason); err != nil {
		klog.V(logErrorLev).Infof("update job(%s) failed reason(%v),failed!", job.Name, reason)
	}

	klog.V(logInfoLev).Infof("set job(%s) to pending, reason:%v.", job.Name, reason)
	job.PodGroup.Status.Phase = scheduling.PodGroupPhase(vjobs.Pending)

	if err := rescheduling.ReleaseFaultJobTakeNodes(job); err != nil {
		return err
	}

	return nil
}

// ForceDeleteFaultPod Force Delete Pod by k8s.
func ForceDeleteFaultPod(ssn *framework.Session, nameSpace string, podName string, podUID types.UID) error {
	deleteOptions := metav1.DeleteOptions{
		GracePeriodSeconds: new(int64),
		Preconditions:      metav1.NewUIDPreconditions(string(podUID)),
	}

	err := ssn.KubeClient().CoreV1().Pods(nameSpace).Delete(context.TODO(), podName, deleteOptions)
	if err != nil {
		klog.V(logErrorLev).Infof("Failed to delete %s: %v", podUID, err)
		return err
	}

	klog.V(logInfoLev).Infof("%s==%v force terminated and removed from etcd", podName, podUID)
	return nil
}

func forceDeleteFaultJob(ssn *framework.Session, job *api.JobInfo) error {
	for _, task := range job.Tasks {
		pod := task.Pod
		if err := ForceDeleteFaultPod(ssn, task.Namespace, pod.Name, pod.UID); err != nil {
			updatePodPendingReason(task, err.Error())
			continue
		}
	}
	return nil
}

func graceDeleteFaultJob(ssn *framework.Session, job *api.JobInfo, reason string) error {
	for _, task := range job.Tasks {
		if err := ssn.Evict(task, reason); err != nil {
			klog.V(logErrorLev).Infof("%s Failed to restart %s : %v", PluginName, task.UID, err)
			updatePodPendingReason(task, err.Error())
			return err
		}
	}
	return nil
}

func evictFaultJob(ssn *framework.Session, job *api.JobInfo, reason string) error {
	// Write restart reason into vcJob.
	updatePodGroupPendingReason(ssn, job, reason)
	label, getErr := rescheduling.GetJobFaultRescheduleLabel(job)
	if getErr != nil {
		label = rescheduling.JobOffRescheduleLabelValue
	}

	var deleteErr error
	switch label {
	case rescheduling.JobForceRescheduleLabelValue:
		deleteErr = forceDeleteFaultJob(ssn, job)
	case rescheduling.JobGraceRescheduleLabelValue:
		deleteErr = graceDeleteFaultJob(ssn, job, reason)
	case rescheduling.JobOffRescheduleLabelValue:
		deleteErr = fmt.Errorf("job reschedule %s", label)
	default:
		deleteErr = fmt.Errorf("not support %s to reschedule job", label)
	}

	return deleteErr
}

// RestartJob set the job restart and the reason.
func RestartJob(ssn *framework.Session, job *api.JobInfo, obj interface{}) error {
	var reason string

	// for set job not meet case
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

	if job.PodGroup.Status.Phase == scheduling.PodGroupRunning {
		if err := evictFaultJob(ssn, job, reason); err != nil {
			klog.V(logErrorLev).Infof("%s Failed to restart %s : %v", PluginName, job.UID, err)
			return err
		}
	}
	return nil
}

func (hwNPU *ScheduleHandler) preCheckJob(job *api.JobInfo, confs []conf.Configuration) error {
	return hwutil.ValidJobSelector(job, confs)
}

func (hwNPU *ScheduleHandler) isHwNPUJob(job *api.JobInfo) error {
	curNPUPlugin := hwNPU.getNPUPlugin(job)
	if curNPUPlugin == nil {
		return errors.New(noneNPUPlugin)
	}

	return curNPUPlugin.IsMyJob(job)
}

func (hwNPU *ScheduleHandler) validNPUJob(job *api.JobInfo) *api.ValidateResult {
	curNPUPlugin := hwNPU.getNPUPlugin(job)
	if curNPUPlugin == nil {
		return nil
	}

	return curNPUPlugin.ValidNPUJobFn(job)
}

// SetJobPendReasonByNodesCase In nodes select case, set node failed and add failed reason.
func SetJobPendReasonByNodesCase(ssn *framework.Session, nodes map[string]*api.NodeInfo, job *api.JobInfo) {
	var msgString string
	var errorNodeCount int

	for _, task := range job.Tasks {
		nodeErr, ok := job.NodesFitErrors[task.UID]
		if !ok {
			continue
		}

		msgString = nodeErr.Error()
		errorNodeCount = 0
		msgs := strings.Split(msgString, ", ")
		for _, msg := range msgs {
			// only error need failed, warning will pending
			if strings.Contains(msg, nodeNoFitSelectorError) || strings.Contains(msg, nodesNoMeetNPUReqError) {
				errorNodeCount++
				klog.V(logInfoLev).Infof("%s %s : %v", PluginName, task.Name, msg)
			}
		}

		if _, ok := task.Resreq.ScalarResources[a310NPUCardName]; ok {
			return
		}

		availableNodes := len(nodes) - errorNodeCount
		needNodes := len(job.Tasks)
		klog.V(logDebugLev).Infof("%s %d %d %v", PluginName, availableNodes, needNodes, job.NodesFitErrors)
		if availableNodes < needNodes {
			klog.V(logErrorLev).Infof("%s %s req (%d)nodes but has (%d)nodes, will be pending.",
				PluginName, job.Name, needNodes, availableNodes)
			if setErr := SetJobPendingReason(ssn, job, job.NodesFitErrors); setErr != nil {
				klog.V(logErrorLev).Infof("%s setJobFailed err:%v.", PluginName, setErr)
			}
		}
	}
}

func isJobInitial(job *api.JobInfo) bool {
	if job.ValidTaskNum() < job.MinAvailable {
		return false
	}

	if job.PodGroup.Status.Phase != scheduling.PodGroupRunning {
		klog.V(logInfoLev).Infof("%s not running", job.UID)
		return false
	}

	return true
}

// ValidJobFn For job preconception, used by volcano frame.
func (hwNPU *ScheduleHandler) ValidJobFn(obj interface{}, confs []conf.Configuration) *api.ValidateResult {
	klog.V(logInfoLev).Infof("enter job valid")
	defer klog.V(logInfoLev).Infof("leave job valid")

	job := getJobHandle(obj)
	if job == nil {
		klog.V(logErrorLev).Infof(" validJobFn convert <%v> failed.", obj)
		reason := "job convert failed"
		return &api.ValidateResult{
			Pass:    false,
			Reason:  reason,
			Message: fmt.Sprintf("validJobFn [%v] failed:%v", obj, reason),
		}
	}

	if !isJobInitial(job) {
		klog.V(logDebugLev).Infof("%s job(%s) is not ready.", PluginName, job.Name)
		return nil
	}

	// Validate job selector, for all kinds of job.
	if errPrecheck := hwNPU.preCheckJob(job, confs); errPrecheck != nil {
		klog.V(logErrorLev).Infof("%s %s, err: %v.", PluginName, job.Name, errPrecheck)

		msg := "Job selector error"
		return &api.ValidateResult{
			Pass:    false,
			Reason:  msg,
			Message: fmt.Sprintf("%v", errPrecheck),
		}
	}

	if err := hwNPU.isHwNPUJob(job); err != nil {
		klog.V(logDebugLev).Infof("%s job(%s) : %v.", PluginName, job.Name, err)
		// to be Compatible with CPU scenarios ,cannot return error
		return nil
	}

	result := hwNPU.validNPUJob(job)
	if result != nil {
		klog.V(logErrorLev).Infof("%s validNPUJob failed:%v.", PluginName, result.Message)
		return result
	}

	klog.V(logInfoLev).Infof("check ok, Job(%s), reqNPU(%v).", job.Name, job.TotalRequest)

	return nil
}
