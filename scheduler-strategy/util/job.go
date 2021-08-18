/*
Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.

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

Package util is using for HuaWei Ascend9 pin affinity schedule utilities.

*/
package util

import (
	"errors"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/conf"
)

// ValidJobSelector Verify npu job selector, which must be config.
func ValidJobSelector(job *api.JobInfo, confs []conf.Configuration) error {
	jobSelectors := GetJobSelectors(job)
	if len(jobSelectors) == 0 {
		msg := fmt.Errorf("%s GetJobSelectors nil", job.Name)
		klog.V(logErrorLev).Infof("%s.", msg.Error())
		return msg
	}

	schedulerConf := GetSchedulerSelectorConfig(confs)
	if len(schedulerConf) == 0 {
		// get scheduler selector configure failed, including default
		msg := "scheduler selector get nil"
		klog.V(logErrorLev).Infof("%s : %s.", job.Name, msg)
		return errors.New(msg)
	}

	// check the job selector
	if !isSelectorMeetJob(jobSelectors, schedulerConf) {
		klog.V(logErrorLev).Infof("job(%s) selector:%v not meet scheduler conf:%v.",
			job.Name, jobSelectors, schedulerConf)
		return fmt.Errorf("job(%s) selector:%v not meet scheduler conf:%v", job.Name, jobSelectors, schedulerConf)
	}
	return nil
}

// GetJobSelectors Get job selectors.
func GetJobSelectors(job *api.JobInfo) map[string]string {
	var jobSelector = make(map[string]string, constIntNum3)

	for _, task := range job.Tasks {
		taskSelector := getTaskSelectors(task)
		for k, v := range taskSelector {
			selector, ok := jobSelector[k]
			if !ok {
				// no task selector
				jobSelector[k] = v
				continue
			}
			if isSelectorContains(selector, v) {
				// has task selector
				continue
			}
			// use '|' to join tasks
			jobSelector[k] = selector + "|" + v
		}
	}
	return jobSelector
}

// GetJobReqNPUNum Get job's request npu numbers.
func GetJobReqNPUNum(job *api.JobInfo, npuCardName string) (int, error) {
	jobNPU, ok := job.TotalRequest.ScalarResources[v1.ResourceName(npuCardName)]
	if !ok || int(jobNPU/npuHex) == 0 {
		klog.V(logDebugLev).Infof("%s npu:%v %+v .", job.Name, v1.ResourceName(npuCardName), job.TotalRequest)
		return 0, errors.New("job no use npu")
	}

	return int(jobNPU / npuHex), nil
}

// IsJobOfCardMode Return job's mode: true is card, false is module.
func IsJobOfCardMode(job *api.JobInfo) bool {
	jobSelectors := GetJobSelectors(job)
	if len(jobSelectors) == 0 {
		klog.V(logErrorLev).Infof("job(%s) has no selectors.", job.Name)
		return false
	}

	acceleratorValue, ok := jobSelectors[acceleratorType]
	if !ok {
		// no acceleratorType means module
		klog.V(logDebugLev).Infof("job(%s) is module type.", job.Name)
		return false
	}

	if acceleratorValue == cardAcceleratorType {
		klog.V(logDebugLev).Infof("job(%s) is card type.", job.Name)
		return true
	}

	klog.V(logDebugLev).Infof("job(%s) is module type.", job.Name)
	return false
}

// CheckSingleTrainMode Single Train job has only one task.
func CheckSingleTrainMode(job *api.JobInfo) error {
	taskNum := len(job.Tasks)

	klog.V(logDebugLev).Infof("checkSingleTrainMode job(%s) has %d tasks.", job.Name, taskNum)

	if taskNum > 1 {
		return fmt.Errorf("%s single trainning has too many task:%d", job.Name, taskNum)
	}

	return nil
}
