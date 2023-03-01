/*
Copyright(C)2020-2023. Huawei Technologies Co.,Ltd. All rights reserved.

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

// Determine if the two string has same element.
func isEachStringContainsSameElement(first, second, seq string) bool {
	if first == second {
		return true
	}
	fList := strings.Split(first, seq)
	sList := strings.Split(second, seq)
	for _, vFirst := range fList {
		for _, vSecond := range sList {
			if strings.EqualFold(vFirst, vSecond) {
				return true
			}
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
	resLabel := make(map[string]string, util.MapInitNum)
	for labelKey, labelValue := range job.PodGroup.Labels {
		resLabel[labelKey] = labelValue
	}
	for _, task := range job.Tasks {
		taskSelector := GetTaskLabels(task)
		for k, v := range taskSelector {
			label, ok := resLabel[k]
			if !ok {
				// no task selector
				resLabel[k] = v
				continue
			}
			if isSelectorContains(label, v) {
				// has task selector
				continue
			}
			// use '|' to join tasks
			resLabel[k] = label + "|" + v
		}
	}
	return resLabel
}

// GetVCJobReqNPUTypeFromJobInfo get job request resource, only NPU.
func GetVCJobReqNPUTypeFromJobInfo(vcJob *api.JobInfo) (string, int, error) {
	if vcJob == nil || vcJob.TotalRequest == nil {
		klog.V(util.LogInfoLev).Infof("GetVCJobReqNPUTypeFromJobInfo nil job's parameter.")
		return "", 0.0, errors.New("nil parameter")
	}

	for k, v := range vcJob.TotalRequest.ScalarResources {
		// must contain "huawei.com/"
		if strings.Contains(string(k), util.HwPreName) {
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
		// must contain "huawei.com/"
		if strings.Contains(string(k), util.HwPreName) {
			return string(k), int(v / util.NPUHexKilo)
		}
		continue
	}
	klog.V(util.LogErrorLev).Infof("GetVCTaskReqNPUTypeFromTaskInfo %+v.", vcTask.Resreq.ScalarResources)
	return "", 0
}

// GetJobNPUTasks get NPUTask from jobInfo.
func GetJobNPUTasks(vcJob *api.JobInfo) map[api.TaskID]util.NPUTask {
	if vcJob == nil || len(vcJob.Tasks) == 0 {
		return nil
	}
	resultMap := make(map[api.TaskID]util.NPUTask, util.MapInitNum)
	for taskID, taskInf := range vcJob.Tasks {
		name, num := GetVCTaskReqNPUTypeFromTaskInfo(taskInf)
		if num == 0 {
			resultMap[taskID] = util.NPUTask{}
			continue
		}
		resultMap[taskID] = util.NPUTask{Name: taskInf.Name, NameSpace: taskInf.Namespace, ReqNPUName: name,
			ReqNPUNum: num,
			Selector:  GetTaskSelectors(taskInf), Label: GetTaskLabels(taskInf), VTask: &util.VTask{}}
	}
	return resultMap
}

// InitSelfPluginByJobInfo init job's handler, the deal plugin.
func (sJob *SchedulerJob) InitSelfPluginByJobInfo(sHandle *ScheduleHandler) {
	if sJob == nil {
		return
	}

	pluginName := sJob.getPluginNameByReq()
	if pluginName == "" {
		return
	}

	plugin, ok := sHandle.NPUPlugins[pluginName]
	if !ok {
		return
	}

	sJob.handler = plugin
}

// IsJobInitial Determine if the task is ready.
func IsJobInitial(job *api.JobInfo) bool {
	return job.ValidTaskNum() >= job.MinAvailable
}

// IsJobRestarted used for rescheduling, judge if job restarted
func IsJobRestarted(job *api.JobInfo) bool {
	return IsJobInitial(job) && job.PodGroup.Status.Phase == scheduling.PodGroupRunning
}

// Init the SchedulerJob's init.
func (sJob *SchedulerJob) Init(vcJob *api.JobInfo, sHandle *ScheduleHandler) error {
	if sJob == nil || vcJob == nil {
		klog.V(util.LogInfoLev).Infof("SchedulerJob_Init: parameter is nil.")
		return errors.New("parameter is nil")
	}
	if initErr := sJob.initByJobInfo(vcJob); initErr != nil {
		klog.V(util.LogDebugLev).Infof("%s initByJobInfo %s", vcJob.UID, initErr)
		return initErr
	}

	if !sJob.IsJobSupportByPlugin(sHandle) {
		klog.V(util.LogInfoLev).Infof("%s IsJobSupportByPlugin not has suitable plugin.", sJob.Name)
		return fmt.Errorf("%s's plugin not regist", sJob.Name)
	}

	sJob.InitSelfPluginByJobInfo(sHandle)
	return nil
}

// setJobType get job type, used in vJob temporary.
func (sJob *SchedulerJob) initVTasks(vcJob *api.JobInfo) {
	for tID, t := range vcJob.Tasks {
		tmpTask, ok := sJob.Tasks[tID]
		if !ok {
			klog.V(util.LogErrorLev).Infof("%s not in frame tasks.", tID)
			continue
		}

		if initErr := tmpTask.InitVTask(t); initErr != nil {
			klog.V(util.LogErrorLev).Infof("Init vTask %s %s.", tID, initErr)
			continue
		}

		sJob.Tasks[tID] = tmpTask
	}
}

// initVJob get job type, used in vJob temporary.
func (sJob *SchedulerJob) initVJob(vcJob *api.JobInfo) {
	sJob.VJob = &util.VJob{}
	sJob.SetJobType()
	sJob.SetJobStatusByInf(vcJob)

	if !sJob.IsVJob() {
		klog.V(util.LogDebugLev).Infof("%s not VJob.", vcJob.Name)
		return
	}
	sJob.initVTasks(vcJob)
	return
}

func (sJob *SchedulerJob) initByJobInfo(vcJob *api.JobInfo) error {
	sJob.ComJob = util.ComJob{Name: vcJob.UID, NameSpace: vcJob.Namespace,
		Selector: GetJobSelectorFromVcJob(vcJob),
		Label:    GetJobLabelFromVcJob(vcJob)}
	sJob.NPUJob = nil
	sJob.handler = nil
	name, num, err := GetVCJobReqNPUTypeFromJobInfo(vcJob)
	if err != nil {
		return err
	}
	sJob.NPUJob = &util.NPUJob{ReqNPUName: name, ReqNPUNum: num, Tasks: GetJobNPUTasks(vcJob)}
	sJob.initVJob(vcJob)
	return nil
}

// IsNPUJob check SchedulerJob is npu job
func (sJob SchedulerJob) IsNPUJob() bool {
	return sJob.handler != nil
}

// ValidJobSelector validate the job selector.
func (sJob SchedulerJob) ValidJobSelector(vcFrame VolcanoFrame) error {
	if len(sJob.Selector) == 0 || len(vcFrame.Conf) == 0 || len(vcFrame.Conf[0].Arguments) == 0 {
		msg := fmt.Errorf("%s or vcFrame's selectors nil", sJob.Name)
		klog.V(util.LogErrorLev).Infof("%s.", msg.Error())
		return msg
	}

	// check the job selector
	if !util.IsSelectorMeetJob(sJob.Selector, vcFrame.Conf[0].Arguments) {
		meetErr := fmt.Errorf("job(%s) selector:%#v not meet scheduler conf:%#v", sJob.Name, sJob.Selector,
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
	if errPreCheck := sJob.preCheckJob(vcFrame); errPreCheck != nil {
		klog.V(util.LogErrorLev).Infof("%s %s, err: %#v.", PluginName, sJob.Name, errPreCheck)

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

	klog.V(util.LogInfoLev).Infof("%s valid ok.", sJob.Name)
	return nil
}

func updatePodsPendingReason(job *api.JobInfo, tID api.TaskID, reason string) {
	if tID != "" {
		if t, ok := job.Tasks[tID]; ok {
			updatePodPendingReason(t, reason)
			return
		}
		return
	}

	for _, task := range job.Tasks {
		updatePodPendingReason(task, reason)
	}
}

func (sHandle *ScheduleHandler) updatePodGroupPendingReason(job *api.JobInfo, reason string) {
	job.JobFitErrors = reason

	jc := scheduling.PodGroupCondition{
		Type:               scheduling.PodGroupUnschedulableType,
		Status:             v1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		TransitionID:       string(sHandle.FrameAttr.UID),
		Reason:             reason,
		Message:            reason,
	}

	for k, value := range job.PodGroup.Status.Conditions {
		if strings.Contains(value.Message, reason) {
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
		// for write pending reason into pod
		updatePodsPendingReason(vcJob, "", reasonTmp)
	case map[api.TaskID]*api.FitErrors:
		vcJob.NodesFitErrors = value
		for tID, nodeErrors := range value {
			// for write pending reason into pod
			updatePodsPendingReason(vcJob, tID, nodeErrors.Error())
			reasonTmp += nodeErrors.Error()
		}
	default:
		return fmt.Errorf("assert reason(%T) failed", reason)
	}
	// for write pending reason into vcjob
	sHandle.updatePodGroupPendingReason(vcJob, reasonTmp)
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
		reason := "job is not ready"
		klog.V(util.LogErrorLev).Infof("%s job(%s) not ready:%#v.", PluginName, job.Name,
			job.PodGroup.Status.Phase)
		return &api.ValidateResult{Pass: false, Reason: reason,
			Message: fmt.Sprintf("validJobFn [%#v] failed:%#v", obj, reason)}
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
	if int32(len(job.Tasks)-len(job.NodesFitErrors)) >= job.MinAvailable {
		klog.V(util.LogDebugLev).Infof("%s not block by nodes(tasks:%d -> jobMin:%d -> nodeErrs:%d).", job.Name,
			len(job.Tasks), job.MinAvailable, len(job.NodesFitErrors))
		return
	}
	if setErr := sHandle.SetJobPendingReason(job, job.NodesFitErrors); setErr != nil {
		klog.V(util.LogErrorLev).Infof("%s setJobFailed err:%s.", PluginName, setErr)
	}
}

// CheckNodeNum Check whether the number of cards on the node meets the task requirements.
func (sJob *SchedulerJob) CheckNodeNum(taskInfo *api.TaskInfo, vcNode NPUNode) error {
	if sJob == nil || taskInfo == nil {
		return errors.New(objectNilError)
	}
	vcTask, ok := sJob.NPUJob.Tasks[taskInfo.UID]
	if !ok {
		klog.V(util.LogErrorLev).Infof("CheckNodeNum %+v.", sJob.SchedulerJobAttr.NPUJob)
		return fmt.Errorf("no %s in SchedulerJob", taskInfo.Name)
	}
	if !vcTask.IsNPUTask() {
		return nil
	}
	nodeNPUNum, ok := vcNode.Idle[v1.ResourceName(vcTask.ReqNPUName)]
	if !ok {
		return fmt.Errorf("%s not have %s", vcNode.Name, vcTask.ReqNPUName)
	}
	if int(nodeNPUNum/util.NPUHexKilo) < vcTask.ReqNPUNum {
		return fmt.Errorf("%s not meet %s's %s:%#v",
			vcNode.Name, vcTask.Name, vcTask.ReqNPUName, vcTask.ReqNPUNum)
	}
	return nil
}

func (sJob SchedulerJob) getPluginNameByReq() string {
	name := sJob.ReqNPUName
	// 1. dynamic vJobs
	if strings.Contains(name, "npu-core") {
		label, ok := sJob.Label[util.JobKindKey]
		if !ok {
			klog.V(util.LogErrorLev).Infof("%s no has %s label in dyCut mode.", sJob.Name, util.JobKindKey)
			return ""
		}
		switch label {
		case util.JobKind910Value:
			name = util.NPU910CardName
		case util.JobKind310Value:
			name = util.NPU310CardName
		case util.JobKind310PValue:
			name = util.NPU310PCardName
		default:
			klog.V(util.LogErrorLev).Infof("%s unknown label: %s in dyCut mode.", sJob.Name, label)
			return ""
		}
	}
	// 2. static vJobs
	if strings.HasSuffix(name, "c") {
		nameSplit := strings.Split(name, "-")
		if len(nameSplit) < util.NPUIndex2 {
			return ""
		}
		return nameSplit[0]
	}
	return name
}

// IsJobSupportByPlugin judge job whether has it's plugin.
func (sJob SchedulerJob) IsJobSupportByPlugin(sHandle *ScheduleHandler) bool {
	name := sJob.getPluginNameByReq()
	if name == "" {
		return false
	}
	return sHandle.IsPluginRegistered(name)
}

// GetAnnoName get job AnnoName, include vNPU job.
func (sJob SchedulerJob) GetAnnoName() (string, error) {
	name := sJob.ReqNPUName
	if strings.Contains(name, "npu-core") {
		_, ok := sJob.Label[util.JobKindKey]
		if !ok {
			klog.V(util.LogErrorLev).Infof("%s no has %s label in dyCut mode.", sJob.Name, util.JobKindKey)
			return "", fmt.Errorf("no %s label in dyCut mode", util.JobKindKey)
		}
		return util.AscendNPUCore, nil
	}
	return sJob.handler.GetAnnoName(), nil
}

// GetReqCardNameFromRingController Get request card name from RingController.
func (sJob SchedulerJob) GetReqCardNameFromRingController() string {
	ringType, ok := sJob.Label[util.JobKindKey]
	if !ok {
		return ""
	}
	ringTypeSplit := strings.Split(ringType, "-")
	if len(ringTypeSplit) < util.NPUIndex2 {
		return ""
	}
	return util.NPUCardPreName + ringTypeSplit[util.NPUIndex1]
}
