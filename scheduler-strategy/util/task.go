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

	acceleratorValue, ok := taskSelectors[acceleratorType]
	if !ok {
		// no acceleratorType means module
		klog.V(logDebugLev).Infof("task(%s) is module type.", task.Name)
		return false
	}

	if acceleratorValue == cardAcceleratorType {
		klog.V(logDebugLev).Infof("task(%s) is card type.", task.Name)
		return true
	}

	klog.V(logDebugLev).Infof("task(%s) is module type.", task.Name)
	return false
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
