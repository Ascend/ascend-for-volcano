/*
Copyright(C) 2020. Huawei Technologies Co.,Ltd. All rights reserved.

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

Package topology910 is using for HuaWei Ascend910 pin affinity schedule.

*/
package topology910

import (
	"errors"
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"strings"
	"volcano.sh/volcano/pkg/apis/scheduling"
	"volcano.sh/volcano/pkg/apis/scheduling/v1beta1"
	"volcano.sh/volcano/pkg/cli/vjobs"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/conf"
)

func getJobHandle(obj interface{}) *api.JobInfo {
	job, ok := obj.(*api.JobInfo)
	if !ok {
		klog.V(logErrorLev).Infof("%s job valid Failed to convert <%v> to *JobInfo", PluginName, obj)
		return nil
	}

	return job
}

func getJobReqNpuNum(job *api.JobInfo) (int, error) {
	jobNpu, ok := job.TotalRequest.ScalarResources[npu910CardName]
	if !ok || int(jobNpu/npuHex) == 0 {
		return 0, errors.New("job no use npu")
	}

	return int(jobNpu / npuHex), nil
}

func getNpuJobDefaultSelector() map[string]string {
	var defaultSchedulerConfig map[string]string
	defaultSchedulerConfig = make(map[string]string, magicNumInt3)

	defaultSchedulerConfig[archSelector] = huaweiArchArm + "|" + huaweiArchX86

	return defaultSchedulerConfig
}

// for verify npu job must config selector
func validNpuJobSelector(job *api.JobInfo) error {
	jobSelectors, errJob := getJobSelectors(job)
	if errJob != nil {
		klog.V(logErrorLev).Infof("%s %s, err: %v", PluginName, job.Name, errJob)
		return fmt.Errorf("job(%s) selector error:%v", job.Name, errJob)
	}

	defaultSchedulerConfig := getNpuJobDefaultSelector()

	for defKey, defValue := range defaultSchedulerConfig {
		jobValue, jobOk := jobSelectors[defKey]
		if !jobOk {
			msg := fmt.Errorf("%s has no selector:%s", job.Name, defKey)
			klog.V(logErrorLev).Infof("%s : %v", PluginName, msg)
			return msg
		}

		if !strings.Contains(defValue, jobValue) {
			msg := fmt.Errorf("%s selector[%s]:[%s] not in [%s]", job.Name, defKey, jobValue, defValue)
			klog.V(logErrorLev).Infof("%s : %v", PluginName, msg)
			return msg
		}
	}

	return nil
}

func isNpuJob(job *api.JobInfo) error {
	if _, err := getJobReqNpuNum(job); err != nil {
		return err
	}

	return nil
}

func getJobSelectors(job *api.JobInfo) (map[string]string, error) {
	var jobSelector map[string]string
	jobSelector = make(map[string]string, magicNumInt3)

	for _, task := range job.Tasks {
		taskSelector := getTaskSelectors(task)
		for k, v := range taskSelector {
			selector, ok := jobSelector[k]
			if !ok {
				// no task selector
				jobSelector[k] = v
				continue
			}
			if strings.Contains(selector, v) {
				// has task selector
				continue
			}
			// use '|' to join tasks
			jobSelector[k] = selector + "|" + v
		}
	}
	return jobSelector, nil
}

func isSelectorMeetJob(jobSelectors, schedulerConf map[string]string) bool {
	for jobKey, jobValue := range jobSelectors {
		confValue, confOk := schedulerConf[jobKey]
		if !confOk {
			klog.V(logErrorLev).Infof("conf has no job selector key:%s", jobKey)
			return false
		}

		if !strings.Contains(confValue, jobValue) {
			klog.V(logErrorLev).Infof("conf has no job selector value:%s", jobValue)
			return false
		}
	}
	return true
}

func validJobSelector(job *api.JobInfo, confs []conf.Configuration) error {
	jobSelectors, errJob := getJobSelectors(job)
	if errJob != nil {
		klog.V(logErrorLev).Infof("%s JobName: %s, err: %v,", PluginName, job.Name, errJob)
		return fmt.Errorf("job(%s) selector error (%v)", job.Name, errJob)
	}

	schedulerConf := getSchedulerSelectorConfig(confs)
	if schedulerConf == nil || len(schedulerConf) == 0 {
		// get scheduler selector configure failed, including default
		msg := "scheduler selector get nil"
		klog.V(logErrorLev).Infof("%s : %s", PluginName, msg)
		return errors.New(msg)
	}

	// check the job selector
	if !isSelectorMeetJob(jobSelectors, schedulerConf) {
		klog.V(logErrorLev).Infof("%s job(%s) selector:%v not meet scheduler conf:%v",
			PluginName, job.Name, jobSelectors, schedulerConf)
		return fmt.Errorf("job(%s) selector:%v not meet scheduler conf:%v", job.Name, jobSelectors, schedulerConf)
	}
	return nil
}

// true is card ,false is model
func isJobCardModel(job *api.JobInfo) bool {
	// one task is card module, the job is
	for _, task := range job.Tasks {
		ok := isTaskOfCardMode(task)
		if !ok {
			klog.V(logDebugLev).Infof("task(%s) is module mode", task.Name)
			return false
		}
	}

	return true
}

func validJobNpuNum(job *api.JobInfo, jobType string) error {
	jobNpu, err := getJobReqNpuNum(job)
	if err != nil {
		klog.V(logDebugLev).Infof("job(%s) get npu number failed", job.Name)
		return err
	}

	if jobType == cardAcceleratorType {
		// only support 1,2,2*n
		if jobNpu == magicNumInt1 || jobNpu%magicNumInt2 == 0 {
			return nil
		}
		return fmt.Errorf("illegal req_npu num: %d in %s mode", jobNpu, cardAcceleratorType)
	}

	if jobNpu == magicNumInt1 ||
		jobNpu == magicNumInt2 ||
		jobNpu == npuNumPerHccs ||
		jobNpu%nodeNpuNumber == 0 {
		return nil
	}

	return fmt.Errorf("illegal req_npu num:%d in %s mode", jobNpu, moduleAcceleratorType)
}

// less 8 npu, can only one task.
func checkSingleTrainMode(job *api.JobInfo) error {
	taskNum := len(job.Tasks)

	klog.V(logDebugLev).Infof("%s checkSingleTrainMode job(%s) has %d tasks", PluginName, job.Name, taskNum)

	if taskNum > magicNumInt1 {
		return fmt.Errorf("%s single trainning has too many task:%d", job.Name, taskNum)
	}

	return nil
}

// more 8 npu required,every task need 8 npu.
func checkDistributeTrainMode(job *api.JobInfo, nodeNpu int) error {
	taskNum := len(job.Tasks)

	klog.V(logDebugLev).Infof("%s checkDistributeTrainMode job(%s) has %d tasks", PluginName, job.Name, taskNum)

	for _, task := range job.Tasks {
		taskNpu, taskError := getTaskNpuNum(task)
		if taskError != nil {
			return errors.New("no npu task")
		}

		klog.V(logDebugLev).Infof("%s checkDistributeTrainMode task(%s) has %d npu", PluginName, task.Name, taskNpu)

		if taskNpu != nodeNpu {
			return fmt.Errorf("DistributeTrain %s req npu [%d] but node [%d]", task.Name, taskNpu, nodeNpu)
		}
	}

	return nil
}

func validJobNpuMode(job *api.JobInfo, jobType string) error {
	var jobNpu int
	var err error
	var nodeNpu = nodeNpuNumber

	if jobNpu, err = getJobReqNpuNum(job); err != nil {
		return err
	}

	if jobType == cardAcceleratorType {
		nodeNpu = magicNumInt2
	}

	if jobNpu <= nodeNpu {
		if err = checkSingleTrainMode(job); err != nil {
			return err
		}
		return nil
	}

	if err = checkDistributeTrainMode(job, nodeNpu); err != nil {
		return err
	}

	return nil
}

func validModuleJob(job *api.JobInfo) error {
	if jobError := validJobNpuNum(job, moduleAcceleratorType); jobError != nil {
		return jobError
	}

	if jobError := validJobNpuMode(job, moduleAcceleratorType); jobError != nil {
		return jobError
	}

	return nil
}

func validCardJob(job *api.JobInfo) error {
	if jobError := validJobNpuNum(job, cardAcceleratorType); jobError != nil {
		return jobError
	}

	if jobError := validJobNpuMode(job, cardAcceleratorType); jobError != nil {
		return jobError
	}

	return nil
}

func validJobModel(job *api.JobInfo) error {
	if isJobCardModel(job) {
		// card mode job
		klog.V(logDebugLev).Infof("job(%s) is card mode", job.Name)
		if errJob := validCardJob(job); errJob != nil {
			return errJob
		}
		return nil
	}
	// module mode
	klog.V(logDebugLev).Infof("job(%s) is module mode", job.Name)
	if errJob := validModuleJob(job); errJob != nil {
		return errJob
	}
	return nil
}

func validNpuJob(job *api.JobInfo, confs []conf.Configuration) *api.ValidateResult {
	// 1.validate npu job selector
	if err := validNpuJobSelector(job); err != nil {
		klog.V(logErrorLev).Infof("%s err: %v", PluginName, err)
		return &api.ValidateResult{
			Pass:    false,
			Reason:  err.Error(),
			Message: fmt.Sprintf("validNpuJob err: %v", err),
		}
	}
	// 2.validate job model
	if errJob := validJobModel(job); errJob != nil {
		klog.V(logErrorLev).Infof("%s err: %v", PluginName, errJob)
		return &api.ValidateResult{
			Pass:    false,
			Reason:  "job model error",
			Message: fmt.Sprintf("%s, err: %v", job.Name, errJob),
		}
	}

	return nil
}

func validJobFn(obj interface{}, confs []conf.Configuration) *api.ValidateResult {
	klog.V(logInfoLev).Infof("%s enter job valid", PluginName)
	defer klog.V(logInfoLev).Infof("%s leave job valid", PluginName)

	job := getJobHandle(obj)
	if job == nil {
		klog.V(logErrorLev).Infof("%s validJobFn convert <%v> failed", PluginName, obj)
		reason := "job convert failed"
		return &api.ValidateResult{
			Pass:    false,
			Reason:  reason,
			Message: fmt.Sprintf("%s validJobFn [%v] failed:%v", PluginName, obj, reason),
		}
	}

	// validate job selector, for all kinds
	if errSelector := validJobSelector(job, confs); errSelector != nil {
		klog.V(logErrorLev).Infof("%s %s, err: %v", PluginName, job.Name, errSelector)
		if setErr := setJobFailed(job, errSelector.Error()); setErr != nil {
			klog.V(logErrorLev).Infof("%s setJobFailed err: %v", PluginName, setErr)
		}

		msg := "Job selector error"
		return &api.ValidateResult{
			Pass:    false,
			Reason:  msg,
			Message: fmt.Sprintf("%v", errSelector),
		}
	}

	if err := isNpuJob(job); err != nil {
		klog.V(logDebugLev).Infof("%s job(%s) : %v", PluginName, job.Name, err)
		// to be Compatible with CPU scenarios ,cannot return error
		return nil
	}

	result := validNpuJob(job, confs)
	if result != nil {
		klog.V(logErrorLev).Infof("%s validNpuJob failed:%v", PluginName, result.Message)
		if setErr := setJobFailed(job, result.Message); setErr != nil {
			klog.V(logErrorLev).Infof("%s setJobFailed err: %v", PluginName, setErr)
		}
		return result
	}

	klog.V(logInfoLev).Infof("%s check ok, Job(%s), reqNpu(%v)", PluginName, job.Name, job.TotalRequest)

	return nil
}

func initPgLabels(jobs map[api.JobID]*api.JobInfo) {
	for _, jobIn := range jobs {
		jobIn.PodGroup.Labels = make(map[string]string, magicNumInt3)
	}
	klog.V(logDebugLev).Infof("%s pd init ok", PluginName)
}

func updatePgLabels(job *api.JobInfo, key, value string) error {
	if job.PodGroup.Labels == nil {
		return errors.New("nil pg labels")
	}
	job.PodGroup.Labels[key] = value

	return nil
}

func setJobStatusByScheduler(job *api.JobInfo, status scheduling.PodGroupPhase) error {
	job.PodGroup.Status.Phase = status
	return updatePgLabels(job, string(status), "scheduler")
}

func updateJobFailedReason(job *api.JobInfo, reason interface{}) error {
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
		return fmt.Errorf("aseert reason(%T) failed", reason)
	}
	// for write failed reason into pod
	updatePodsFailedReason(job, reasonTmp)
	// for write failed reason into vcjob
	updatePodGroupFailedReason(job, reasonTmp)

	return nil
}

func updatePodGroupFailedReason(job *api.JobInfo, reasonTmp string) {
	jc := scheduling.PodGroupCondition{
		Type:               scheduling.PodGroupConditionType("ScheduledFailed"),
		Status:             v1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		TransitionID:       "",
		Reason:             v1beta1.PodFailedReason,
		Message:            reasonTmp,
	}
	job.PodGroup.Status.Conditions = append(job.PodGroup.Status.Conditions, jc)
}

func setJobFailed(job *api.JobInfo, reason interface{}) error {
	if err := updateJobFailedReason(job, reason); err != nil {
		klog.V(logErrorLev).Infof("update job(%s) failed reason(%v),failed!", job.Name, reason)
	}

	klog.V(logInfoLev).Infof("set job(%s) to failed, reason:%v", job.Name, reason)

	return setJobStatusByScheduler(job, scheduling.PodGroupPhase(vjobs.Failed))
}

func setJobFailedByNodesCase(nodes map[string]*api.NodeInfo, job *api.JobInfo) {
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
				klog.V(logInfoLev).Infof("%s %s[%v]", PluginName, task.Name, msg)
				errorNodeCount++
			}
		}

		availableNodes := len(nodes) - errorNodeCount
		needNodes := len(job.Tasks)
		if availableNodes < needNodes {
			klog.V(logErrorLev).Infof("%s %s req (%d)nodes but has (%d)nodes, will be failed",
				PluginName, job.Name, needNodes, availableNodes)
			if setErr := setJobFailed(job, job.NodesFitErrors); setErr != nil {
				klog.V(logErrorLev).Infof("%s setJobFailed err:%v", PluginName, setErr)
			}
		}
	}
}
