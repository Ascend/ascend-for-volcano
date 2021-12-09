/*
Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package util is using for HuaWei Ascend9 pin affinity schedule utilities.

*/
package util

import (
	"errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
)

// GetTaskNPUNum Get task requires npu number.
func GetTaskNPUNum(task *api.TaskInfo, npuCardName string) (int, error) {
	tmpNPU, ok := task.Resreq.ScalarResources[v1.ResourceName(npuCardName)]
	if !ok || int(tmpNPU/npuHex) == 0 {
		return 0, errors.New("not npu task")
	}

	taskNPU := int(tmpNPU / npuHex)
	return taskNPU, nil
}

// IsNPUTask Judge the task whether is npu or not.
func IsNPUTask(task *api.TaskInfo, npuCardName string) error {
	tmpNPU, ok := GetTaskNPUNum(task, npuCardName)
	if ok != nil || tmpNPU == 0 {
		return errors.New("not npu task")
	}

	return nil
}

// GetTaskSelectors Get task's selector.
func GetTaskSelectors(task *api.TaskInfo) map[string]string {
	return getTaskSelectors(task)
}

// IsTaskOfCardMode Determine if the task is in card mode.
func IsTaskOfCardMode(task *api.TaskInfo) bool {
	taskSelectors := getTaskSelectors(task)
	if len(taskSelectors) == 0 {
		klog.V(logDebugLev).Infof("task(%s) has no selectors.", task.Name)
		return false
	}

	return ValidStringMapKeyAndValue(taskSelectors, acceleratorType, cardAcceleratorType)
}

// GetDeviceIDsFromAnnotations Get npu card ids from Annotations.
func GetDeviceIDsFromAnnotations(Annotations map[string]string, npuCardName string, npuCardPreName string) []int {
	tmpTopStr, ok := Annotations[npuCardName]
	if !ok {
		klog.V(logDebugLev).Infof("%s getDeviceIDsFromAnnotations top nil.", npuCardName)
		return nil
	}

	tmpDeviceIDs := ChangeTopToIntArray(tmpTopStr, npuCardPreName)
	if tmpDeviceIDs == nil {
		klog.V(logErrorLev).Infof("%s getDeviceIDsFromAnnotations to int failed.", npuCardName)
		return nil
	}

	return tmpDeviceIDs
}

func getTaskSelectors(task *api.TaskInfo) map[string]string {
	return task.Pod.Spec.NodeSelector
}

// GetTaskLabels Get task labels from pod label.
func GetTaskLabels(task *api.TaskInfo) map[string]string {
	return task.Pod.Labels
}
