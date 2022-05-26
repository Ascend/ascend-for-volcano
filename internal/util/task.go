/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package util is using for HuaWei Ascend9 pin affinity schedule utilities.

*/
package util

import (
	"context"
	"fmt"
	"strings"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"
)

// GetTaskNPUNum Get task requires npu number.
func GetTaskNPUNum(task *api.TaskInfo, npuCardName string) (int, error) {
	tmpNPU, ok := task.Resreq.ScalarResources[v1.ResourceName(npuCardName)]
	if !ok || int(tmpNPU/NPUHex) == 0 {
		return 0, fmt.Errorf("not %s task", npuCardName)
	}

	taskNPU := int(tmpNPU / NPUHex)
	return taskNPU, nil
}

// IsNPUTask Judge the task whether is npu or not.
func IsNPUTask(task *api.TaskInfo, npuCardName string) error {
	tmpNPU, ok := GetTaskNPUNum(task, npuCardName)
	if ok != nil || tmpNPU == 0 {
		return fmt.Errorf("not npu task")
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
		klog.V(LogDebugLev).Infof("task(%s) has no selectors.", task.Name)
		return false
	}

	return ValidStringMap(taskSelectors, AcceleratorType, CardAcceleratorType)
}

// GetDeviceIDsFromAnnotations Get npu card ids from Annotations.
func GetDeviceIDsFromAnnotations(Annotations map[string]string, npuCardName string, npuCardPreName string) []int {
	tmpTopStr, ok := Annotations[npuCardName]
	if !ok {
		klog.V(LogDebugLev).Infof("%s getDeviceIDsFromAnnotations top nil.", npuCardName)
		return nil
	}

	tmpDeviceIDs := ChangeTopToIntArray(tmpTopStr, npuCardPreName)
	if tmpDeviceIDs == nil {
		klog.V(LogErrorLev).Infof("%s getDeviceIDsFromAnnotations to int failed.", npuCardName)
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

// GetRealPodByTask get pod by task from k8s directly.
func GetRealPodByTask(ssn *framework.Session, task *api.TaskInfo) (*v1.Pod, error) {
	pod, err := ssn.KubeClient().CoreV1().Pods(task.Namespace).Get(context.TODO(), task.Name, metav1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			klog.V(LogErrorLev).Infof("Failed to get pod %v in%v: %v",
				task.Namespace, task.Name, err)
			return nil, err
		}
	}
	return pod, nil
}

// GetPodUsedNPUNames get the task alloc NPUs.like set of Ascend910-1.
func GetPodUsedNPUNames(task *api.TaskInfo, resType string) []string {
	if task == nil {
		return nil
	}
	res, ok := task.Pod.Annotations[resType]
	if !ok {
		return nil
	}
	resSlice := strings.Split(res, ",")
	return resSlice
}

// GetReqResourceNameFromTask get task use npu name
func GetReqResourceNameFromTask(vTask *api.TaskInfo) (string, error) {
	if vTask == nil {
		return "", fmt.Errorf("nil parameter")
	}
	for k := range vTask.Resreq.ScalarResources {
		temp := string(k)
		// must contains "huawei.com/Ascend"
		if strings.Contains(temp, CommCardPreName) {
			return temp, nil
		}
		continue
	}
	klog.V(LogErrorLev).Infof("GetReqResourceNameFromTask %+v.", vTask.Resreq.ScalarResources)
	return "", fmt.Errorf("nil NPU")
}
