/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package plugin is using for HuaWei Ascend pin affinity schedule frame.

*/
package plugin

import (
	"errors"
	"fmt"
	"strings"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"volcano.sh/apis/pkg/apis/scheduling"
	"volcano.sh/volcano/pkg/cli/vjobs"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// Determine if the selectors are exactly equal.
func isSelectorContains(defValue, jobValue string) bool {
	for _, v := range strings.Split(defValue, "|") {
		if strings.EqualFold(v, jobValue) {
			return true
		}
	}

	return false
}

// GetTaskSelectors get task's selector.
func GetTaskSelectors(task *api.TaskInfo) map[string]string {
	if task == nil {
		klog.V(util.LogErrorLev).Infof("GetTaskSelectors task nil.")
		return nil
	}
	return task.Pod.Spec.NodeSelector
}

// GetTaskLabels get task's Labels.
func GetTaskLabels(task *api.TaskInfo) map[string]string {
	if task == nil {
		klog.V(util.LogErrorLev).Infof("GetTaskLabels task nil.")
		return nil
	}
	return task.Pod.Labels
}

// GetJobSelectorFromVcJob get job selector.
func GetJobSelectorFromVcJob(job *api.JobInfo) map[string]string {
	var jobLabel = make(map[string]string, util.MapInitNum)
	for _, task := range job.Tasks {
		taskSelectors := task.Pod.Spec.NodeSelector
		for k, v := range taskSelectors {
			label, ok := jobLabel[k]
			if !ok {
				// no task selector
				jobLabel[k] = v
				continue
			}
			if isSelectorContains(label, v) {
				// has task selector
				continue
			}
			// use '|' to join tasks
			jobLabel[k] = label + "|" + v
		}
	}
	return jobLabel
}

// GetJobLabelFromVcJob get job's label, not task's.
func GetJobLabelFromVcJob(job *api.JobInfo) map[string]string {
	if job == nil {
		klog.V(util.LogErrorLev).Infof("GetJobLabelFromVcJob job nil.")
		return nil
	}
	jobLabel := job.PodGroup.Labels
	if jobLabel == nil {
		jobLabel = make(map[string]string, util.MapInitNum)
	}
	for _, task := range job.Tasks {
		taskSelector := GetTaskLabels(task)
		for k, v := range taskSelector {
			label, ok := jobLabel[k]
			if !ok {
				// no task selector
				jobLabel[k] = v
				continue
			}
			if isSelectorContains(label, v) {
				// has task selector
				continue
			}
			// use '|' to join tasks
			jobLabel[k] = label + "|" + v
		}
	}
	return jobLabel
}

// GetVCJobReqNPUTypeFromJobInfo get job request resource, only NPU.
func GetVCJobReqNPUTypeFromJobInfo(vcJob *api.JobInfo) (string, int, error) {
	if vcJob == nil || vcJob.TotalRequest == nil {
		klog.V(util.LogInfoLev).Infof("GetVCJobReqNPUTypeFromJobInfo nil job's parameter.")
		return "", 0.0, errors.New("nil parameter")
	}

	for k, v := range vcJob.TotalRequest.ScalarResources {
		// must contains "huawei.com/Ascend"
		if strings.Contains(string(k), util.NPUCardPreName) {
			return string(k), int(v / util.NPUHexKilo), nil
		}
	}
	klog.V(util.LogErrorLev).Infof("GetVCJobReqNPUTypeFromJobInfo %+v.", vcJob.TotalRequest.ScalarResources)
	return "", 0.0, errors.New("nil NPU")
}

// GetVCTaskReqNPUTypeFromTaskInfo get task request resource, only NPU.
func GetVCTaskReqNPUTypeFromTaskInfo(vcTask *api.TaskInfo) (string, int) {
	if vcTask == nil || vcTask.Resreq == nil {
		klog.V(util.LogInfoLev).Infof("GetVCTaskReqNPUTypeFromTaskInfo nil job's parameter.")
		return "", 0
	}
	for k, v := range vcTask.Resreq.ScalarResources {
		// must contains "huawei.com/Ascend"
		if strings.Contains(string(k), util.NPUCardPreName) {
			return string(k), int(v / util.NPUHexKilo)
		}
		continue
	}
	klog.V(util.LogErrorLev).Infof("GetVCTaskReqNPUTypeFromTaskInfo %+v.", vcTask.Resreq.ScalarResources)
	return "", 0
}

// GetJobNPUTasks get NPUTask from jobInfo.
func GetJobNPUTasks(vcJob *api.JobInfo) map[string]util.NPUTask {
	if vcJob == nil || len(vcJob.Tasks) == 0 {
		return nil
	}
	resultMap := make(map[string]util.NPUTask, util.MapInitNum)
	for taskID, taskInf := range vcJob.Tasks {
		name, num := GetVCTaskReqNPUTypeFromTaskInfo(taskInf)
		if num == 0 {
			resultMap[string(taskID)] = util.NPUTask{}
			continue
		}
		resultMap[string(taskID)] = util.NPUTask{TaskName: taskInf.Name, ReqNPUName: name, ReqNPUNum: num,
			Selector: GetTaskSelectors(taskInf), Label: GetTaskLabels(taskInf)}
	}
	return resultMap
}

// InitSelfPluginByJobInfo init job's handler, the deal plugin.
func (sJob *SchedulerJob) InitSelfPluginByJobInfo(sHandle *ScheduleHandler) {
	if sJob == nil {
		return
	}
	pluginNames := strings.Split(sJob.ReqNPUName, "-")
	if len(pluginNames) == 0 {
		return
	}
	plugin, ok := sHandle.NPUPlugins[pluginNames[0]]
	if !ok {
		return
	}
	sJob.handler = plugin
}

// IsJobInitial Determine if the task is ready.
func IsJobInitial(job *api.JobInfo) bool {
	if job.ValidTaskNum() < job.MinAvailable {
		return false
	}

	if job.PodGroup.Status.Phase != scheduling.PodGroupRunning {
		klog.V(util.LogInfoLev).Infof("%s not running %#v", job.UID, job.PodGroup.Status.Phase)
		return false
	}

	return true
}

// Init the SchedulerJob's init.
func (sJob *SchedulerJob) Init(vcJob *api.JobInfo, sHandle *ScheduleHandler) error {
	if sJob == nil || vcJob == nil {
		klog.V(util.LogInfoLev).Infof("SchedulerJob_Init: parameter is nil.")
		return errors.New("parameter is nil")
	}
	sJob.ComJob = util.ComJob{JobName: vcJob.UID, NameSpace: vcJob.Namespace,
		Selector: GetJobSelectorFromVcJob(vcJob),
		Label:    GetJobLabelFromVcJob(vcJob)}
	sJob.NPUJob = nil
	sJob.handler = nil
	name, num, err := GetVCJobReqNPUTypeFromJobInfo(vcJob)
	if err != nil {
		return err
	}
	if !sHandle.IsPluginRegistered(name) {
		return fmt.Errorf("%s's plugin not regist", name)
	}

	sJob.NPUJob = &util.NPUJob{ReqNPUName: name, ReqNPUNum: num, Tasks: GetJobNPUTasks(vcJob)}
	sJob.InitSelfPluginByJobInfo(sHandle)
	return nil
}

// IsNPUJob check SchedulerJob is npu job
func (sJob SchedulerJob) IsNPUJob() bool {
	return sJob.handler != nil
}

// ValidJobSelector validate the job selector.
func (sJob SchedulerJob) ValidJobSelector(vcFrame VolcanoFrame) error {
	if len(sJob.Selector) == 0 || len(vcFrame.Conf) == 0 || len(vcFrame.Conf[0].Arguments) == 0 {
		msg := fmt.Errorf("%s or vcFrame's selectors nil", sJob.JobName)
		klog.V(util.LogErrorLev).Infof("%s.", msg.Error())
		return msg
	}

	// check the job selector
	if !util.IsSelectorMeetJob(sJob.Selector, vcFrame.Conf[0].Arguments) {
		meetErr := fmt.Errorf("job(%s) selector:%#v not meet scheduler conf:%#v", sJob.JobName, sJob.Selector,
			vcFrame.Conf[0].Arguments)
		klog.V(util.LogErrorLev).Infof(meetErr.Error())
		return meetErr
	}
	return nil
}

func (sJob SchedulerJob) preCheckJob(vcFrame VolcanoFrame) error {
	return sJob.ValidJobSelector(vcFrame)
}

// ValidJobFn valid job.
func (sJob SchedulerJob) ValidJobFn(vcFrame VolcanoFrame) *api.ValidateResult {
	// Validate job selector, for all kinds of job.
	if errPreCheck := sJob.preCheckJob(vcFrame); errPreCheck != nil {
		klog.V(util.LogErrorLev).Infof("%s %s, err: %#v.", PluginName, sJob.JobName, errPreCheck)

		msg := "Job selector error"
		return &api.ValidateResult{
			Pass:    false,
			Reason:  msg,
			Message: fmt.Sprintf("%s", errPreCheck.Error()),
		}
	}
	if result := sJob.handler.ValidNPUJob(); result != nil {
		klog.V(util.LogErrorLev).Infof("%s validNPUJob failed:%#v.", PluginName, result.Message)
		return result
	}

	klog.V(util.LogInfoLev).Infof("check ok, %s, reqNPU(%#v:%#v).", sJob.JobName, sJob.ReqNPUName, sJob.ReqNPUNum)
	return nil
}

func updatePodsPendingReason(job *api.JobInfo, reason string) {
	for _, task := range job.Tasks {
		updatePodPendingReason(task, reason)
	}
}

func (sHandle *ScheduleHandler) updatePodGroupPendingReason(job *api.JobInfo, reason string) {
	jc := scheduling.PodGroupCondition{
		Type:               scheduling.PodGroupUnschedulableType,
		Status:             v1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		TransitionID:       string(sHandle.FrameAttr.UID),
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

// SetJobPendingReason set the pod and podGroup pending reason.
func (sHandle *ScheduleHandler) SetJobPendingReason(vcJob *api.JobInfo, reason interface{}) error {
	if sHandle == nil || vcJob == nil {
		klog.V(util.LogErrorLev).Infof("SetJobPendingReason not init jobs.")
		return errors.New(util.ArgumentError)
	}
	var reasonTmp string

	switch value := reason.(type) {
	case string:
		// job failed
		vcJob.JobFitErrors = value
		reasonTmp = value
	case map[api.TaskID]*api.FitErrors:
		vcJob.NodesFitErrors = value
		for _, nodeErrors := range value {
			reasonTmp += nodeErrors.Error()
		}
	default:
		return fmt.Errorf("assert reason(%T) failed", reason)
	}
	// for write pending reason into pod
	updatePodsPendingReason(vcJob, reasonTmp)
	// for write pending reason into vcjob
	sHandle.updatePodGroupPendingReason(vcJob, reasonTmp)
	vcJob.PodGroup.Status.Phase = scheduling.PodGroupPhase(vjobs.Pending)
	return nil
}

// JobValid the job valid, used by volcano frame.
func (sHandle *ScheduleHandler) JobValid(obj interface{}) *api.ValidateResult {
	klog.V(util.LogInfoLev).Infof("enter job valid")
	defer klog.V(util.LogInfoLev).Infof("leave job valid")

	if sHandle == nil {
		return &api.ValidateResult{Pass: false, Reason: objectNilError,
			Message: fmt.Sprintf("validJobFn [%#v] failed:%#v", obj, objectNilError)}
	}

	job, ok := obj.(*api.JobInfo)
	if !ok {
		reason := "job convert failed"
		klog.V(util.LogErrorLev).Infof("%s :%#v.", reason, obj)
		return &api.ValidateResult{Pass: false, Reason: reason,
			Message: fmt.Sprintf("validJobFn [%#v] failed:%#v", obj, reason)}
	}
	if !IsJobInitial(job) {
		klog.V(util.LogDebugLev).Infof("%s job(%s) not ready:%#v.", PluginName, job.Name,
			job.PodGroup.Status.Phase)
		return nil
	}
	vcJob, ok := sHandle.Jobs[job.UID]
	if !ok {
		klog.V(util.LogDebugLev).Infof("%s %s not support or init", PluginName, job.Name)
		return nil
	}

	result := vcJob.ValidJobFn(sHandle.FrameAttr)
	if result != nil {
		if setErr := sHandle.SetJobPendingReason(job, result.Message); setErr != nil {
			klog.V(util.LogErrorLev).Infof("%s setJobFailed err: %#v.", PluginName, setErr)
		}
		return result
	}
	return nil
}

// SetJobPendReasonByNodesCase In nodes select case, set node failed and add failed reason.
func (sHandle ScheduleHandler) SetJobPendReasonByNodesCase(job *api.JobInfo) {
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
				klog.V(util.LogInfoLev).Infof("%s %s : %#v", PluginName, task.Name, msg)
			}
		}
	}
	availableNodes := len(sHandle.Nodes) - errorNodeCount
	needNodes := len(job.Tasks)
	klog.V(util.LogDebugLev).Infof("%s %d:%d %#v", job.Name, availableNodes, needNodes, job.NodesFitErrors)
	if availableNodes < needNodes {
		klog.V(util.LogErrorLev).Infof("%s %s req (%d)nodes but has (%d)nodes, will be pending.",
			PluginName, job.Name, needNodes, availableNodes)
		if setErr := sHandle.SetJobPendingReason(job, job.NodesFitErrors); setErr != nil {
			klog.V(util.LogErrorLev).Infof("%s setJobFailed err:%#v.", PluginName, setErr)
		}
	}
}

// CheckNodeNum Check whether the number of cards on the node meets the task requirements.
func (sJob *SchedulerJob) CheckNodeNum(taskInfo *api.TaskInfo, vcNode NPUNode) error {
	if sJob == nil || taskInfo == nil {
		return errors.New(objectNilError)
	}
	vcTask, ok := sJob.NPUJob.Tasks[string(taskInfo.UID)]
	if !ok {
		klog.V(util.LogErrorLev).Infof("CheckNodeNum %+v.", sJob.SchedulerJobAttr.NPUJob)
		return fmt.Errorf("no %s in SchedulerJob", taskInfo.Name)
	}
	nodeNPUNum, ok := vcNode.Idle[v1.ResourceName(vcTask.ReqNPUName)]
	if !ok {
		return fmt.Errorf("%s not have %s", vcNode.Name, vcTask.ReqNPUName)
	}
	if int(nodeNPUNum/util.NPUHexKilo) < vcTask.ReqNPUNum {
		return fmt.Errorf("%s not meet %s's %s:%#v", vcNode.Name, vcTask.TaskName, vcTask.ReqNPUName, vcTask.ReqNPUNum)
	}
	return nil
}
