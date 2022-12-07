/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package plugin is using for HuaWei Ascend pin affinity schedule frame.
*/
package plugin

import (
	"fmt"
	"reflect"
	"strings"

	"k8s.io/api/core/v1"
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// NPUAllocateFunc Allocate npu and called by volcano frame.
func (sHandle ScheduleHandler) NPUAllocateFunc(task *api.TaskInfo) {
	if task == nil {
		klog.V(util.LogErrorLev).Infof("NPUAllocateFunc %s.", util.ArgumentError)
		return
	}
	vcJob, ok := sHandle.Jobs[task.Job]
	if !ok {
		klog.V(util.LogDebugLev).Infof("NPUAllocateFunc %s not req npu.", task.Name)
		return
	}
	nodeName := task.NodeName
	node, found := sHandle.Nodes[nodeName]
	if !found {
		klog.V(util.LogWarningLev).Infof("%s npuAllocateFunc %s not exist.", PluginName, nodeName)
		return
	}
	vcNode := vcJob.handler.UseAnnotation(task, node)
	if vcNode != nil {
		// update node.
		sHandle.Nodes[nodeName] = *vcNode
	}
	klog.V(util.LogDebugLev).Infof("%s %#v useAnnotation node [%s]'s top.", PluginName, task.Name, nodeName)
}

func (sHandle *ScheduleHandler) releaseAnnotation(task *api.TaskInfo, vcJob SchedulerJob, vcNode NPUNode) {
	rankIndex, ok := task.Pod.Annotations[podRankIndex]
	klog.V(util.LogInfoLev).Infof("task %s node %s rankIndex annotation: %s", task.Name, vcNode.Name, rankIndex)
	if ok { // if pod rankIndex has been written, delete it
		klog.V(util.LogInfoLev).Infof("node %s bind failed, release rankIndex annotation: %s",
			vcNode.Name, rankIndex)
		delete(task.Pod.Annotations, podRankIndex)
	}
	vcTask, ok := vcJob.Tasks[task.Name]
	if !ok {
		klog.V(util.LogInfoLev).Infof("task %s not in vcjob %s", vcTask.Name, vcJob.Name)
		return
	}
	reqStr, ok := task.Pod.Annotations[util.AscendNPUPodRealUse]
	if !ok {
		return
	}
	reqSlice := strings.Split(reqStr, ",")
	if len(reqSlice) != vcTask.ReqNPUNum {
		return
	}
	value, ok := vcNode.Annotation[vcTask.ReqNPUName]
	if !ok {
		return
	}
	vcNode.Annotation[vcTask.ReqNPUName] = reqStr
	if value != "" {
		// if failed, reset by next session.
		if isEachStringContainsSameElement(value, reqStr, ",") {
			annErr := fmt.Errorf("%s:%s has same NPU used %s:%s", vcNode.Name, value, vcTask.Name, reqStr)
			klog.V(util.LogErrorLev).Infof("releaseAnnotation %s", annErr)
			return
		}
		vcNode.Annotation[vcTask.ReqNPUName] = reqStr + "," + value
	}
	sHandle.Nodes[vcNode.Name] = vcNode
	klog.V(util.LogDebugLev).Infof("%s releaseAnnotation %s's %s on %s,new top:[%s].", PluginName, task.Name,
		reqStr, vcNode.Name, reqStr+","+value)
}

// NPUDeallocateFunc Free assigned npu, if allocate failed by volcano frame.
func (sHandle *ScheduleHandler) NPUDeallocateFunc(task *api.TaskInfo) {
	if sHandle == nil || task == nil {
		klog.V(util.LogInfoLev).Infof("NPUDeallocateFunc failed: %s.", util.ArgumentError)
		return
	}
	vcJob, ok := sHandle.Jobs[task.Job]
	if !ok {
		klog.V(util.LogDebugLev).Infof("NPUDeallocateFunc %s not req npu.", task.Name)
		return
	}
	nodeName := task.NodeName
	node, found := sHandle.Nodes[nodeName]
	if !found {
		klog.V(util.LogWarningLev).Infof("%s npuAllocateFunc NOT EXIST node [%s].", PluginName, nodeName)
		return
	}
	sHandle.releaseAnnotation(task, vcJob, node)
	klog.V(util.LogDebugLev).Infof("%s %#v NPUDeallocateFunc node [%s]'s top.", PluginName, task.Name, nodeName)
}

func updatePodPendingReason(task *api.TaskInfo, reasonTmp string) {
	condition := v1.PodCondition{
		Type:    v1.PodScheduled,
		Status:  v1.ConditionFalse,
		Reason:  v1.PodReasonUnschedulable,
		Message: reasonTmp,
	}
	for _, tmp := range task.Pod.Status.Conditions {
		if reflect.DeepEqual(tmp, condition) {
			return
		}
	}
	task.Pod.Status.Conditions = append(task.Pod.Status.Conditions, condition)
}
