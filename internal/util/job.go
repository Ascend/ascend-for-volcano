/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package util is using for HuaWei Ascend9 pin affinity schedule utilities.

*/
package util

import (
	"errors"
	"fmt"
	"strings"

	"k8s.io/api/core/v1"
	"k8s.io/klog"
	"volcano.sh/apis/pkg/apis/scheduling"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/conf"
)

// ValidJobSelector Verify npu job selector, which must be config.
func ValidJobSelector(job *api.JobInfo, confs []conf.Configuration) error {
	jobSelectors := GetJobSelectors(job)
	if len(jobSelectors) == 0 {
		msg := fmt.Errorf("%s GetJobSelectors nil", job.Name)
		klog.V(LogErrorLev).Infof("%s.", msg.Error())
		return msg
	}

	schedulerConf := GetSchedulerSelectorConfig(confs)
	if len(schedulerConf) == 0 {
		// get scheduler selector configure failed, including default
		msg := "scheduler selector get nil"
		klog.V(LogErrorLev).Infof("%s : %s.", job.Name, msg)
		return errors.New(msg)
	}

	// check the job selector
	if !isSelectorMeetJob(jobSelectors, schedulerConf) {
		klog.V(LogErrorLev).Infof("job(%s) selector:%v not meet scheduler conf:%v.",
			job.Name, jobSelectors, schedulerConf)
		return fmt.Errorf("job(%s) selector:%v not meet scheduler conf:%v", job.Name, jobSelectors, schedulerConf)
	}
	return nil
}

// GetJobSelectors Get job selectors.
func GetJobSelectors(job *api.JobInfo) map[string]string {
	var jobSelector = make(map[string]string, NPUIndex3)

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
	if !ok || int(jobNPU/NPUHex) == 0 {
		klog.V(LogDebugLev).Infof("%s no npu(%v) total:[%+v].", job.Name, npuCardName, job.TotalRequest)
		return 0, errors.New("job no use npu")
	}

	return int(jobNPU / NPUHex), nil
}

// IsJobOfCardMode Return job's mode: true is card, false is module.
func IsJobOfCardMode(job *api.JobInfo) bool {
	jobSelectors := GetJobSelectors(job)
	if len(jobSelectors) == 0 {
		klog.V(LogErrorLev).Infof("job(%s) has no selectors.", job.Name)
		return false
	}

	klog.V(LogDebugLev).Infof("job(%s) is module type.", job.Name)
	return ValidStringMap(jobSelectors, AcceleratorType, CardAcceleratorType)
}

// CheckSingleTrainMode Single Train job has only one task.
func CheckSingleTrainMode(job *api.JobInfo) error {
	taskNum := len(job.Tasks)

	klog.V(LogDebugLev).Infof("checkSingleTrainMode job(%s) has %d tasks.", job.Name, taskNum)

	if taskNum > 1 {
		return fmt.Errorf("%s single trainning has too many task:%d", job.Name, taskNum)
	}

	return nil
}

// GetJobLabels Get job labels.
func GetJobLabels(job *api.JobInfo) map[string]string {
	var jobLabel = make(map[string]string, NPUIndex3)

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

// IsJobInitial Determine if the task is ready.
func IsJobInitial(job *api.JobInfo) bool {
	if job.ValidTaskNum() < job.MinAvailable {
		return false
	}

	if job.PodGroup.Status.Phase != scheduling.PodGroupRunning {
		klog.V(LogInfoLev).Infof("%s not running %v", job.UID, job.PodGroup.Status.Phase)
		return false
	}

	return true
}

// GetReqResourceNameFromJob get job require npu resource.
func GetReqResourceNameFromJob(vJob *api.JobInfo) (string, error) {
	if vJob == nil {
		return "", errors.New("nil parameter")
	}
	reqReses := api.NewResource(*vJob.PodGroup.Spec.MinResources)
	for k := range reqReses.ScalarResources {
		temp := string(k)
		// must contains "huawei.com/Ascend"
		if strings.Contains(temp, CommCardPreName) {
			return temp, nil
		}
		continue
	}
	klog.V(LogErrorLev).Infof("GetReqResourceNameFromJob %+v.", vJob.PodGroup.Spec.MinResources)
	return "", errors.New("nil NPU")
}

// IsJobRunningByInfo judge job whether is running or not.
func IsJobRunningByInfo(vJob *api.JobInfo) bool {
	if vJob == nil {
		return false
	}
	if len(vJob.Tasks) == 0 {
		return false
	}
	for _, task := range vJob.Tasks {
		if task.Pod == nil {
			return false
		}
		if task.Pod.Status.Phase != v1.PodRunning {
			return false
		}
	}
	return true
}

// GetJobReqResourceNumFromJobPG Get job require resource number from job podGroup.
func GetJobReqResourceNumFromJobPG(tmpJob *api.JobInfo, reqNpuType string) (int, error) {
	if tmpJob == nil {
		return 0, errors.New("nil parameter")
	}
	reqResource := api.NewResource(*tmpJob.PodGroup.Spec.MinResources)
	value, ok := reqResource.ScalarResources[v1.ResourceName(reqNpuType)]
	if !ok {
		return 0, fmt.Errorf("%s no %s", tmpJob.Name, reqNpuType)
	}
	return int(value / NPUHex), nil
}
