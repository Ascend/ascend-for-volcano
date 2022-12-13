/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package util is using for the total variable.
*/
package util

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"
)

// for task status
const (
	TaskStatusUnknown = -1
	TaskStatusInit    = iota
	TaskStatusAllocate
	TaskStatusWrBack
	TaskStatusRunning
	TaskStatusFailed
)

type TaskAllocated struct {
	// like ubuntu
	NodeName string
	// element like 1
	CardName []int
	// element like Ascend310P-2c-100-1
	PhysicsName []string
}

type VTask struct {
	// TASK_STATUS_INIT...
	Status int
	// type: JobTypeWhole, JobTypeDycut, JobTypeStcut.
	Type      int
	Allocated TaskAllocated
}

// NPUTask for npu task need.
type NPUTask struct {
	Name       string
	NameSpace  string
	ReqNPUName string
	ReqNPUNum  int
	// Selector the same as job.
	Selector map[string]string
	Label    map[string]string
	*VTask
}

// GetRealPodByTask get pod specified by task name and namespace from kubernetes
func (asTask *NPUTask) GetRealPodByTask(ssn *framework.Session) (*v1.Pod, error) {
	if asTask == nil {
		klog.V(LogErrorLev).Infof("GetRealPodByTask failed: %s.", ArgumentError)
		return nil, fmt.Errorf(ArgumentError)
	}
	taskInfo, getErr := GetTaskInfoByNameFromSSN(ssn, asTask.Name)
	if getErr != nil {
		klog.V(LogErrorLev).Infof("GetRealPodByTask %s: %#v", asTask.Name, getErr)
		return nil, getErr
	}

	pod, err := ssn.KubeClient().CoreV1().Pods(taskInfo.Namespace).Get(
		context.TODO(), asTask.Name, metav1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			klog.V(LogErrorLev).Infof("Failed to get pod %s in %s: %#v",
				taskInfo.Namespace, asTask.Name, err)
			return nil, err
		}
		klog.V(LogErrorLev).Infof("pod %v in%v not found: %#v",
			taskInfo.Namespace, asTask.Name, err)
		return nil, err
	}
	return pod, nil
}

// DeleteRealPodByTask generally used by force deletion
func (asTask *NPUTask) DeleteRealPodByTask(ssn *framework.Session, waitTime int64) error {
	if asTask == nil {
		klog.V(LogErrorLev).Infof("DeleteRealPodByTask failed: %s.", ArgumentError)
		return fmt.Errorf(ArgumentError)
	}
	taskInfo, getErr := GetTaskInfoByNameFromSSN(ssn, asTask.Name)
	if getErr != nil {
		klog.V(LogErrorLev).Infof("%s GetTaskInfoByNameFromSSN: %#v", asTask.Name, getErr)
	}
	if taskInfo == nil || taskInfo.Pod == nil {
		klog.V(LogInfoLev).Infof("DeleteRealPodByTask pod does not exist")
		return fmt.Errorf("%s: taskInfo does not exist", ArgumentError)
	}

	deleteOptions := metav1.DeleteOptions{
		GracePeriodSeconds: &waitTime,
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
	klog.V(LogDebugLev).Infof("enter EvictJobByTask...")
	if asTask == nil {
		klog.V(LogErrorLev).Infof("EvictJobByTask failed: %s.", ArgumentError)
		return fmt.Errorf(ArgumentError)
	}
	if ssn == nil {
		klog.V(LogErrorLev).Infof("EvictJobByTask failed: %s.", ArgumentError)
		return fmt.Errorf(ArgumentError)
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
	klog.V(LogInfoLev).Infof("Evict %s : %#v", taskName, taskInfo.UID)
	if updateErr := asTask.UpdatePodPendingReason(taskInfo, reason); updateErr != nil {
		return updateErr
	}
	return nil
}

// UpdatePodPendingReason update pod pending reason.
func (asTask *NPUTask) UpdatePodPendingReason(taskInfo *api.TaskInfo, reasonTmp string) error {
	if asTask == nil {
		klog.V(LogErrorLev).Infof("UpdatePodPendingReason failed: %s.", ArgumentError)
		return fmt.Errorf(ArgumentError)
	}
	if asTask.Name != taskInfo.Name {
		return fmt.Errorf("NPUTask %s and TaskInfo %s does not match", asTask.Name, taskInfo.Name)
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
		return nil, fmt.Errorf(ArgumentError)
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

func (asTask *NPUTask) ForceDeletePodByTaskInf(ssn *framework.Session, reason string) error {
	if !asTask.IsTaskInItsNode(ssn) {
		klog.V(LogErrorLev).Infof("%s not in %s, need force delete.", asTask.Name,
			asTask.VTask.Allocated.NodeName)
		deleteErr := asTask.DeleteRealPodByTask(ssn, 0)
		if deleteErr != nil {
			klog.V(LogErrorLev).Infof("GraceDeleteFaultJob %s: %#v.", asTask.Name, deleteErr)
		}
		return deleteErr
	}
	if err := asTask.EvictJobByTask(ssn, reason, asTask.Name); err != nil {
		return err
	}
	return nil
}

// IsTaskInItsNode check if task is on the node
func (asTask *NPUTask) IsTaskInItsNode(ssn *framework.Session) bool {
	if ssn == nil {
		klog.V(LogErrorLev).Infof("isTaskInItsNode has no node.")
		return false
	}
	nodeName := asTask.VTask.Allocated.NodeName
	nodeInf, ok := ssn.Nodes[nodeName]
	if !ok {
		klog.V(LogErrorLev).Infof("session has no node %v.", nodeName)
		return false
	}
	_, taskOK := nodeInf.Tasks[api.TaskID(asTask.Name)]
	_, taskFullNameOK := nodeInf.Tasks[api.TaskID(asTask.NameSpace+"/"+asTask.Name)]
	klog.V(LogDebugLev).Infof("node %s has tasks: %#v", nodeInf.Name, nodeInf.Tasks)
	if !taskOK && !taskFullNameOK {
		klog.V(LogErrorLev).Infof("node %s has no task %s.", nodeInf.Name, asTask.Name)
		return false
	}
	return true
}

func (asTask *NPUTask) setVTaskType() {
	taskType := JobTypeUnknown
	names := strings.Split(asTask.ReqNPUName, "-")
	if len(names) == 1 {
		taskType = JobTypeWhole
	}
	if strings.Contains(asTask.ReqNPUName, "vir") {
		taskType = JobTypeStCut
	}
	if strings.Contains(asTask.ReqNPUName, "npu-core") {
		taskType = JobTypeDyCut
	}
	asTask.Type = taskType
}

func getVTaskUsePhysicsNamesByInfo(taskInf *api.TaskInfo) []string {
	value, ok := taskInf.Pod.Annotations[AscendNPUPodRealUse]
	if !ok {
		klog.V(LogErrorLev).Infof("%s's %#v has no %s.",
			taskInf.Name, taskInf.Pod.Annotations, AscendNPUPodRealUse)
		return nil
	}
	return strings.Split(value, ",")
}

func (vt *VTask) setVTaskUseCardIDs() {
	if len(vt.Allocated.PhysicsName) == 0 {
		klog.V(LogErrorLev).Infof("%#v nil PhysicsName.", vt.Allocated)
		return
	}
	var ids []int
	for _, value := range vt.Allocated.PhysicsName {
		// value like Ascend310P-2c-100-1_1
		tmps := strings.Split(value, "-")
		realV := strings.Split(tmps[len(tmps)-1], "_")
		if len(realV) == 0 {
			klog.V(LogErrorLev).Infof("get card id from %s==>%#v error.", value, tmps)
			continue
		}
		vInt, err := strconv.Atoi(realV[0])
		if err != nil {
			klog.V(LogErrorLev).Infof("setVTaskUseCardIDs %s.", err)
			continue
		}
		ids = append(ids, vInt)
	}
	vt.Allocated.CardName = ids
}

func (asTask *NPUTask) setVTaskAllocated(taskInf *api.TaskInfo) {
	switch asTask.Status {
	case TaskStatusRunning, TaskStatusWrBack, TaskStatusFailed:
		asTask.VTask.Allocated.NodeName = taskInf.NodeName
		asTask.VTask.Allocated.PhysicsName = getVTaskUsePhysicsNamesByInfo(taskInf)
		asTask.VTask.setVTaskUseCardIDs()
	default:
		klog.V(LogErrorLev).Infof("setVTaskAllocated %s status %v.", asTask.Name, asTask.Status)
		return
	}
	return
}

func (asTask *NPUTask) setVTaskStatusFromInfo(taskInf *api.TaskInfo) error {
	if _, ok := taskInf.Pod.Annotations[PodAssignKey]; !ok {
		asTask.Status = TaskStatusInit
		return nil
	}
	if _, ok := taskInf.Pod.Annotations[AscendNPUPodRealUse]; !ok {
		asTask.Status = TaskStatusAllocate
		return nil
	}
	if taskInf.Status == api.Running {
		asTask.Status = TaskStatusRunning
		return nil
	}
	if taskInf.Status == api.Failed || taskInf.Status == api.Releasing {
		asTask.Status = TaskStatusFailed
		return nil
	}
	asTask.Status = TaskStatusUnknown
	return fmt.Errorf("unkown %s status: %s", taskInf.Name, taskInf.Status)
}

func (asTask *NPUTask) InitVTask(taskInf *api.TaskInfo) error {
	asTask.setVTaskType()
	if setErr := asTask.setVTaskStatusFromInfo(taskInf); setErr != nil {
		return setErr
	}
	asTask.setVTaskAllocated(taskInf)
	return nil
}

// IsVNPUTask Determine whether is the NPU virtual task.
// Dynamic segmentation: huawei.com/npu-core.
// static segmentation: huawei.com/Ascend910-Y.
// no segmentation: huawei.com/Ascend910.
func (asTask *NPUTask) IsVNPUTask() bool {
	if asTask == nil {
		return false
	}
	if len(strings.Split(asTask.ReqNPUName, "-")) > 1 {
		return true
	}
	return false
}
