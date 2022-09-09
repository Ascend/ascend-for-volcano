/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package util is using for the total variable.

*/
package util

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"
)

// GetRealPodByTask get pod specified by task name and namespace from kubernetes
func (asTask *NPUTask) GetRealPodByTask(ssn *framework.Session) (*v1.Pod, error) {
	if asTask == nil {
		klog.V(LogErrorLev).Infof("GetRealPodByTask failed: %s.", ArgumentError)
		return nil, errors.New(ArgumentError)
	}
	taskInfo, getErr := GetTaskInfoByNameFromSSN(ssn, asTask.TaskName)
	if getErr != nil {
		klog.V(LogErrorLev).Infof("GetRealPodByTask %s: %#v", asTask.TaskName, getErr)
		return nil, getErr
	}

	pod, err := ssn.KubeClient().CoreV1().Pods(taskInfo.Namespace).Get(
		context.TODO(), asTask.TaskName, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			klog.V(LogErrorLev).Infof("Failed to get pod %s in %s: %#v",
				taskInfo.Namespace, asTask.TaskName, err)
			return nil, err
		}
		klog.V(LogErrorLev).Infof("pod %v in%v not found: %#v",
			taskInfo.Namespace, asTask.TaskName, err)
		return nil, err
	}
	return pod, nil
}

// DeleteRealPodByTask generally used by force deletion
func (asTask *NPUTask) DeleteRealPodByTask(ssn *framework.Session) error {
	if asTask == nil {
		klog.V(LogErrorLev).Infof("DeleteRealPodByTask failed: %s.", ArgumentError)
		return errors.New(ArgumentError)
	}
	taskInfo, getErr := GetTaskInfoByNameFromSSN(ssn, asTask.TaskName)
	if getErr != nil {
		klog.V(LogErrorLev).Infof("%s GetTaskInfoByNameFromSSN: %#v", asTask.TaskName, getErr)
	}

	deleteOptions := metav1.DeleteOptions{
		GracePeriodSeconds: new(int64),
		Preconditions:      metav1.NewUIDPreconditions(string(taskInfo.Pod.UID)),
	}

	err := ssn.KubeClient().CoreV1().Pods(taskInfo.Pod.Namespace).Delete(
		context.TODO(), taskInfo.Pod.Name, deleteOptions)
	if err != nil {
		klog.V(LogErrorLev).Infof("Failed to delete %s: %#v", taskInfo.Pod.UID, err)
		return err
	}

	klog.V(LogInfoLev).Infof("%s==%v force terminated and removed from etcd", taskInfo.Pod.Name, taskInfo.Pod.UID)
	return nil
}

// EvictJobByTask generally used by grace deletion
func (asTask *NPUTask) EvictJobByTask(ssn *framework.Session, reason string, taskName string) error {
	if asTask == nil {
		klog.V(LogErrorLev).Infof("EvictJobByTask failed: %s.", ArgumentError)
		return errors.New(ArgumentError)
	}
	if ssn == nil {
		klog.V(LogErrorLev).Infof("EvictJobByTask failed: %s.", ArgumentError)
		return errors.New(ArgumentError)
	}
	taskInfo, getErr := GetTaskInfoByNameFromSSN(ssn, taskName)
	if getErr != nil {
		klog.V(LogErrorLev).Infof("%s GetTaskInfoByNameFromSSN: %#v", taskName, getErr)
	}
	err := ssn.Evict(taskInfo, reason)
	if err != nil {
		klog.V(LogErrorLev).Infof("Failed to restart %s : %#v", taskName, err)
		if updateErr := asTask.UpdatePodPendingReason(taskInfo, err.Error()); updateErr != nil {
			return updateErr
		}
		return err
	}
	return nil
}

// UpdatePodPendingReason update pod pending reason.
func (asTask *NPUTask) UpdatePodPendingReason(taskInfo *api.TaskInfo, reasonTmp string) error {
	if asTask == nil {
		klog.V(LogErrorLev).Infof("UpdatePodPendingReason failed: %s.", ArgumentError)
		return errors.New(ArgumentError)
	}
	if asTask.TaskName != taskInfo.Name {
		return fmt.Errorf("NPUTask %s and TaskInfo %s does not match", asTask.TaskName, taskInfo.Name)
	}
	condition := v1.PodCondition{
		Type:    v1.PodScheduled,
		Status:  v1.ConditionFalse,
		Reason:  v1.PodReasonUnschedulable,
		Message: reasonTmp,
	}
	for _, tmp := range taskInfo.Pod.Status.Conditions {
		if reflect.DeepEqual(tmp, condition) {
			return nil
		}
	}
	taskInfo.Pod.Status.Conditions = append(taskInfo.Pod.Status.Conditions, condition)
	return nil
}

// GetTaskInfoByNameFromSSN get corresponding api.TaskInfo object by given taskName
func GetTaskInfoByNameFromSSN(ssn *framework.Session, taskName string) (*api.TaskInfo, error) {
	if ssn == nil {
		klog.V(LogErrorLev).Infof("UpdatePodPendingReason failed: %s.", ArgumentError)
		return nil, errors.New(ArgumentError)
	}
	if len(taskName) == 0 {
		klog.V(LogErrorLev).Infof("GetTaskInfoByNameFromSSN failed: taskName is empty")
		return nil, fmt.Errorf("getTaskInfoByNameFromSSN: taskName is empty")
	}
	for _, jobInfo := range ssn.Jobs {
		for _, taskInfo := range jobInfo.Tasks {
			if taskName == taskInfo.Name {
				return taskInfo, nil
			}
		}
	}
	return nil, fmt.Errorf("did not find task %s in session", taskName)
}