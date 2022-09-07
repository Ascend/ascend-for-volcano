/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package util is using for the total variable.

*/
package util

import (
	"strings"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/framework"
)

// IsSelectorMeetJob check the selectors
func IsSelectorMeetJob(jobSelectors, conf map[string]string) bool {
	for jobKey, jobValue := range jobSelectors {
		confValue, confOk := conf[jobKey]
		if !confOk {
			klog.V(LogErrorLev).Infof("conf has no job selector key:%s.", jobKey)
			return false
		}

		if !strings.Contains(confValue, jobValue) {
			klog.V(LogErrorLev).Infof("conf has no job selector value:%s.", jobValue)
			return false
		}
	}
	return true
}

// ValidStringMap valid the map key and value.
func ValidStringMap(tmpMap map[string]string, first, second string) bool {
	tmpValue, ok := tmpMap[first]
	if !ok {
		// no AcceleratorType means module
		return false
	}

	if tmpValue == second {
		return true
	}

	klog.V(LogDebugLev).Infof("valid ok .")
	return false
}

// IsJobOfCardMode Return job's mode: true is card, false is module.
func (asJob *SchedulerJobAttr) IsJobOfCardMode() bool {
	if len(asJob.Selector) == 0 {
		klog.V(LogErrorLev).Infof("job(%s) has no selectors.", asJob.JobName)
		return false
	}

	klog.V(LogDebugLev).Infof("job(%s) is module type.", asJob.JobName)
	return ValidStringMap(asJob.Selector, AcceleratorType, CardAcceleratorType)
}

// CheckJobExistsInKubernetes Check whether the jobs exists in K8S.
func (asJob *SchedulerJobAttr) CheckJobExistsInKubernetes(kubeClient kubernetes.Interface) bool {
	for _, npuTask := range asJob.Tasks {
		realPod, err := npuTask.GetRealPodByTask(nil)
		if realPod != nil { // job中的pod在k8s中能找到, 不能删, job还在
			return true
		}
		if err != nil { // pod返回的不是不存在错误, 不能删, job还在
			if !errors.IsNotFound(err) {
				klog.V(LogErrorLev).Infof("CheckJobExistsInKubernetes  getRealPodByTask %v %v.", npuTask.TaskName, err)
				return true
			}
		}
	}
	klog.V(LogInfoLev).Infof("%v CheckJobExistsInKubernetes, all pods not exist in kubernetes, job can be deleted", asJob.JobName)
	return false
}

// CheckJobPodStatusOK check the pod is running.
func (asJob *SchedulerJobAttr) CheckJobPodStatusOK(ssn *framework.Session) bool {
	for _, npuTask := range asJob.Tasks {
		realPod, err := npuTask.GetRealPodByTask(ssn)
		if err != nil {
			klog.V(LogErrorLev).Infof("checkJobPodStatusOK  getRealPodByTask %v %v.", npuTask.TaskName, err)
			return false
		}
		if realPod.Status.Phase == v1.PodRunning {
			continue
		}
		klog.V(LogErrorLev).Infof("checkJobPodStatusOK %v not ok %v.", npuTask.TaskName, realPod.Status.Phase)
		return false
	}
	klog.V(LogInfoLev).Infof("%v checkJobPodStatusOK.", asJob.JobName)
	return true
}
