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
	"k8s.io/klog"
	"strings"
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
	if !ok {
		return 0, errors.New("not npu job")
	}

	return int(jobNpu / npuHex), nil
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
		return fmt.Errorf("job(%s) selector error", job.Name)
	}

	schedulerConf := getSchedulerSelectorConfig(confs)
	if schedulerConf == nil || len(schedulerConf) == 0 {
		// get scheduler selector configure failed, but need continue
		klog.V(logErrorLev).Infof("%s JobName: %s get selector nil", PluginName, job.Name)
		return errors.New("get scheduler selector nil")
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
		ok := validCardModule(task)
		if !ok {
			klog.V(logDebugLev).Infof("task(%s) is module mode", task.Name)
			return false
		}
	}

	return true
}

func validMouldeJobNpuNum(job *api.JobInfo) error {
	jobNpu, err := getJobReqNpuNum(job)
	if err != nil {
		klog.V(logDebugLev).Infof("job(%s) get npu number failed", job.Name)
		return err
	}

	if jobNpu == magicNumInt1 ||
		jobNpu == magicNumInt2 ||
		jobNpu == npuNumPerHccs ||
		jobNpu%nodeNpuNumber == 0 {
		return nil
	}

	return fmt.Errorf("illegal req_npu num:%d", jobNpu)
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
func checkDistributeTrainMode(job *api.JobInfo) error {
	taskNum := len(job.Tasks)

	klog.V(logDebugLev).Infof("%s checkDistributeTrainMode job(%s) has %d tasks", PluginName, job.Name, taskNum)

	for _, task := range job.Tasks {
		taskNpu, taskError := getTaskNpuNum(task)
		if taskError != nil {
			return errors.New("no npu task")
		}

		klog.V(logDebugLev).Infof("%s checkDistributeTrainMode task(%s) has %d npu", PluginName, task.Name, taskNpu)

		if taskNpu != nodeNpuNumber {
			return fmt.Errorf("DistributeTrain Job: %s  has %d tasks, and req npu illegal: %d", job.Name, taskNum, taskNpu)
		}
	}

	return nil
}

func validJobNpuMode(job *api.JobInfo) error {
	var jobNpu int
	var err error

	if jobNpu, err = getJobReqNpuNum(job); err != nil {
		return err
	}

	if jobNpu <= nodeNpuNumber {
		if err = checkSingleTrainMode(job); err != nil {
			return err
		}
	}

	if jobNpu > nodeNpuNumber {
		if err = checkDistributeTrainMode(job); err != nil {
			return err
		}
	}

	return nil
}

func validModuleJob(job *api.JobInfo) error {
	if jobError := validMouldeJobNpuNum(job); jobError != nil {
		return jobError
	}

	if jobError := validJobNpuMode(job); jobError != nil {
		return jobError
	}

	return nil
}

func validJobModel(job *api.JobInfo) error {
	if isJobCardModel(job) {
		// card mode job
		// valid nothing
		klog.V(logDebugLev).Infof("job(%s) is card mode", job.Name)
		return nil
	}
	// module mode
	klog.V(logDebugLev).Infof("job(%s) is module mode", job.Name)
	errJob := validModuleJob(job)
	if errJob != nil {
		return errJob
	}
	return nil
}

func validNpuJob(job *api.JobInfo, confs []conf.Configuration) *api.ValidateResult {
	// 1.validate job selector
	if errSelector := validJobSelector(job, confs); errSelector != nil {
		klog.V(logErrorLev).Infof("%s JobName: %s, err: %v,", PluginName, job.Name, errSelector)
		return &api.ValidateResult{
			Pass:    false,
			Reason:  "Job selector error",
			Message: fmt.Sprintf("JobName: %s, err:%v", job.Name, errSelector),
		}
	}
	// 2.validate job model
	if errTask := validJobModel(job); errTask != nil {
		klog.V(logErrorLev).Infof("%s err: %v", PluginName, errTask)
		return &api.ValidateResult{
			Pass:    false,
			Reason:  "job model error",
			Message: fmt.Sprintf("JobName: %s, err: %v", job.Name, errTask),
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
		return &api.ValidateResult{
			Pass:    false,
			Message: fmt.Sprintf("%s validJobFn convert <%v> failed", PluginName, obj),
		}
	}

	if err := isNpuJob(job); err != nil {
		klog.V(logInfoLev).Infof("%s job(%s) : %v", PluginName, job.Name, err)
		// to be Compatible with CPU scenarios ,cannot return error
		return nil
	}

	result := validNpuJob(job, confs)
	if result != nil {
		klog.V(logErrorLev).Infof("%s validNpuJob failed:%v", PluginName, result.Message)

		return result
	}

	klog.V(logInfoLev).Infof("%s validJobFn check ok, JobName: %s, reqNpu:%v", PluginName, job.Name, job.TotalRequest)

	return nil
}
