/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package common is using for HuaWei infer common Ascend pin affinity schedule.

*/
package common

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/conf"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

// Name This need by frame init plugin.
func (cn *Scheduler) Name() string {
	return cn.PluginName
}

// OnHandlerStart The npu scheduler policy initial and Scheduler processing.
func (cn *Scheduler) OnHandlerStart(sHandler *plugin.ScheduleHandler) {
	klog.V(util.LogDebugLev).Infof("%s start handler.", cn.PluginName)
	sHandler.AddInitNodesNPUAllocTopology(cn.PluginName, cn.initNodesNPUTopologyFn)
}

// ValidNPUJobFn Check the compliance of the selector and resource request numbers of job.
func (cn *Scheduler) ValidNPUJobFn(job *api.JobInfo) *api.ValidateResult {
	// 1.Validate npu job selector.
	if err := cn.validNPUJobSelector(job); err != nil {
		klog.V(util.LogDebugLev).Infof("%s validNPUJobSelector err: %v.", cn.PluginName, err)
		return &api.ValidateResult{
			Pass:    false,
			Reason:  err.Error(),
			Message: fmt.Sprintf("validNPUJob err: %v", err),
		}
	}

	// 2.Validate job npu number.
	if jobError := cn.validJobNPUNum(job); jobError != nil {
		klog.V(util.LogDebugLev).Infof("%s validJobNPUNum err: %v.", cn.PluginName, jobError)
		return &api.ValidateResult{
			Pass:    false,
			Reason:  "job require npu number illegal",
			Message: fmt.Sprintf("%s, err: %v", job.Name, jobError),
		}
	}
	// 3.Validate job scheduler-strategy.
	if errJob := cn.validJobModel(job); errJob != nil {
		klog.V(util.LogDebugLev).Infof("%s validJobModel err: %v.", cn.PluginName, errJob)
		return &api.ValidateResult{
			Pass:    false,
			Reason:  "job scheduler-strategy error",
			Message: fmt.Sprintf("%s, err: %v", job.Name, errJob),
		}
	}

	return nil
}

// PreCheckNodeFn 310 no need to Distinguish between architecture.
func (cn *Scheduler) PreCheckNodeFn(task *api.TaskInfo, node *api.NodeInfo, confs []conf.Configuration) error {
	schedulerConf := util.GetSchedulerSelectorConfig(confs)
	if len(schedulerConf) == 0 {
		// get scheduler selector configure failed, but need continue
		klog.V(util.LogDebugLev).Infof("%s JobUID: %s get selector nil.", cn.PluginName, task.Name)
		return fmt.Errorf("%s get scheduler selector nil", node.Name)
	}

	// select node by architect
	if err := util.IsSelectorMeetNode(task, node, schedulerConf, cn.AnnoName); err != nil {
		// get scheduler selector configure failed, but need continue
		klog.V(util.LogDebugLev).Infof("%s taskName: %s ,nodeName %s : %v.", cn.PluginName, task.Name, node.Name, err)
		return err
	}
	return nil
}

// CheckNPUResourceStableFn Check whether the node's NPU resources are stable.
func (cn *Scheduler) CheckNPUResourceStableFn(node *api.NodeInfo) error {
	// default is the npu task
	nodeNPUIdleNumFromTop, err := cn.getNodeNPUNumFromOthers(node)
	if err != nil {
		return fmt.Errorf("getNodeNPUNumFromOthers %s : %s", NodesNoMeetNPUReqError, err)
	}

	nodeNPUIdleNumFromIdle, err := util.GetNodeNPUNumFromIdle(node, cn.AnnoName)
	if err != nil {
		return fmt.Errorf("getNodeNPUNumFromIdle %s : %s", NodesNoMeetNPUReqError, err)
	}

	if err = util.CheckNodeNPUStabilize(nodeNPUIdleNumFromTop, nodeNPUIdleNumFromIdle); err != nil {
		return fmt.Errorf("%s : %s", NodeNotStableWarning, err)
	}

	return nil
}

// CheckNodeNPUByTaskFn Check whether the requested resource exists and are sufficient on the node.
func (cn *Scheduler) CheckNodeNPUByTaskFn(task *api.TaskInfo, node *api.NodeInfo, _ bool) error {
	taskNPU, taskError := util.GetTaskNPUNum(task, cn.AnnoName)
	if taskError != nil {
		return fmt.Errorf("getTaskNPUNum %s : %s", NodesNoMeetNPUReqError, taskError)
	}

	nodeNPUTopology := util.GetTopFromNodeOthers(node, cn.AnnoName, cn.AnnoPreVal)
	if nodeNPUTopology == nil {
		// node has none npu
		klog.V(util.LogInfoLev).Infof("%s checkNodeNPUByTask nil,node PluginName:%s(top:%v),task req npu:%d",
			cn.PluginName, node.Name, nodeNPUTopology, taskNPU)
		return fmt.Errorf("%s:get npu nil", NodeNotEnoughNPUWarning)
	}
	klog.V(util.LogInfoLev).Infof("%s %s top:%v,req %d", cn.PluginName, node.Name, nodeNPUTopology, taskNPU)

	return cn.judgeNodeAndTaskNPU(taskNPU, nodeNPUTopology)
}

// UpdateNPUNodeUsedCardFn Update used npu resources on node.
func (cn *Scheduler) UpdateNPUNodeUsedCardFn(node *api.NodeInfo, top interface{}) error {
	useTop, ok := top.([]int)
	if !ok {
		return errors.New(ArgumentError)
	}

	// get node available top
	nodeDeviceIDs := util.GetTopFromNodeOthers(node, cn.AnnoName, cn.AnnoPreVal)
	if len(nodeDeviceIDs) == 0 {
		klog.V(util.LogDebugLev).Infof("%s useAnnotation node(%s) top nil.", cn.PluginName, node.Name)
		return errors.New("nodeDeviceIDs nil")
	}

	// delete the use top
	klog.V(util.LogInfoLev).Infof("%s useAnnotation %s:%v , will use: %v.", cn.PluginName, node.Name,
		nodeDeviceIDs, useTop)
	newNodeTopStr := util.GetRealTopAfterAlloc(nodeDeviceIDs, useTop, cn.AnnoPreVal)
	if newNodeTopStr == "" {
		klog.V(util.LogDebugLev).Infof("%s getRealTopAfterAlloc all top has allocated .", cn.PluginName)
	}

	err := util.ReloadNewTopToNodeOther(node, newNodeTopStr, cn.AnnoName)
	if err != nil {
		klog.V(util.LogDebugLev).Infof("%s reloadNewTopToNode failed.", cn.PluginName)
		return err
	}

	klog.V(util.LogInfoLev).Infof("%s ReloadNewTopToNode %s to %s successes.", cn.PluginName, newNodeTopStr, node.Name)
	return nil
}

// GetReleaseNPUTopologyFn Get the release npu card id from task(pod).
func (cn *Scheduler) GetReleaseNPUTopologyFn(task *api.TaskInfo) (interface{}, error) {
	// get task use top
	taskDeviceIDs := util.GetDeviceIDsFromAnnotations(task.Pod.Annotations, cn.AnnoName, cn.AnnoPreVal)
	if taskDeviceIDs == nil {
		klog.V(util.LogDebugLev).Infof("%s releaseAnnotation failed task:%s", cn.PluginName, task.Name)
		return nil, fmt.Errorf("%s get npu nil", task.Name)
	}

	return taskDeviceIDs, nil
}

// UpdateReleaseNPUNodeTopologyFn Update the node using npu when release pod's npu.
func (cn *Scheduler) UpdateReleaseNPUNodeTopologyFn(node *api.NodeInfo, top interface{}) error {
	taskDeviceIDs, ok := top.([]int)
	if !ok {
		return errors.New(ArgumentError)
	}

	// get node available top
	nodeDeviceIDs := util.GetTopFromNodeOthers(node, cn.AnnoName, cn.AnnoPreVal)
	if nodeDeviceIDs == nil {
		klog.V(util.LogDebugLev).Infof("%s useAnnotation node(%s) top nil", cn.PluginName, node.Name)
		return fmt.Errorf("%s has nil npu", node.Name)
	}
	// delete the use top
	newNodeTopStr := util.GetRealTopAfterRelease(nodeDeviceIDs, taskDeviceIDs, cn.AnnoPreVal)
	if newNodeTopStr == "" {
		klog.V(util.LogDebugLev).Infof("%s getRealTopAfterRelease top failed", cn.PluginName)
		return fmt.Errorf("%s release nil npu", node.Name)
	}

	err := util.ReloadNewTopToNodeOther(node, newNodeTopStr, cn.AnnoName)
	if err != nil {
		klog.V(util.LogDebugLev).Infof("%s reloadNewTopToNode failed", cn.PluginName)
		return err
	}

	klog.V(util.LogInfoLev).Infof("%s useAnnotation node(%s) top(%s) successes", cn.PluginName, node.Name, newNodeTopStr)

	return nil
}

// GetAllocatedNPUFromTopologyFn Get the pod's npu card to record in node others.
func (cn *Scheduler) GetAllocatedNPUFromTopologyFn(task *api.TaskInfo,
	node *api.NodeInfo, _ bool) (interface{}, error) {
	var allocTopologyNPUs []int

	taskNPUNumber, taskError := util.GetTaskNPUNum(task, cn.AnnoName)
	if taskError != nil {
		return nil, errors.New("no npu task")
	}

	nodeTop := util.GetTopFromNodeOthers(node, cn.AnnoName, cn.AnnoPreVal)
	if nodeTop == nil {
		klog.V(util.LogDebugLev).Infof("not npu node[%s], no need to continue.", node.Name)
		return nodeTop, fmt.Errorf("nodeTop is nil")
	}
	klog.V(util.LogInfoLev).Infof("%s %s[%d] in %v.", cn.PluginName,
		task.Name, taskNPUNumber, nodeTop)

	klog.V(util.LogDebugLev).Infof("%s %s get top %v.", cn.PluginName, task.Name, nodeTop)

	allocTopologyNPUs, err := util.GetNPUTopFromHccs(taskNPUNumber, nodeTop)
	if err != nil {
		return allocTopologyNPUs, err
	}
	klog.V(util.LogInfoLev).Infof("%s %s req:%d alloc %v.", cn.PluginName, task.Name, taskNPUNumber, allocTopologyNPUs)
	return allocTopologyNPUs, nil
}

// SetNPUTopologyToPodFn Set the npu card ids into pod.
func (cn *Scheduler) SetNPUTopologyToPodFn(task *api.TaskInfo, top interface{}) error {
	var topologyStr string

	klog.V(util.LogInfoLev).Infof("%s setNPUTopologyToPod begin top:%v", cn.PluginName, top)
	intTop, ok := top.([]int)
	if !ok {
		return errors.New(ArgumentError)
	}

	topologyStr = util.ChangeIntArrToStr(intTop, cn.AnnoPreVal)
	task.Pod.Annotations[cn.AnnoName] = topologyStr
	// to device-plugin judge pending pod.
	task.Pod.Annotations[PodPredicateTime] = strconv.FormatInt(time.Now().UnixNano(), util.Base10)
	klog.V(util.LogInfoLev).Infof("%s setNPUTopologyToPod %s top:%s", cn.PluginName, task.Name, topologyStr)

	return nil
}

// IsMyTask Determine if it is the NPU task of your plug-in.
func (cn *Scheduler) IsMyTask(task *api.TaskInfo) error {
	_, err := util.GetTaskNPUNum(task, cn.AnnoName)
	if err != nil {
		return errors.New(JobNoNPUCard)
	}

	return nil
}

// IsMyNode Determine if it is the NPU node of your plug-in.
func (cn *Scheduler) IsMyNode(node *api.NodeInfo) error {
	_, err := util.GetNPUAllocCardsFromNodeOthers(node, cn.AnnoName)
	if err != nil {
		return errors.New(JobNoNPUCard)
	}

	return nil
}

// IsMyJob Determine if it is the NPU job of your plug-in.
func (cn *Scheduler) IsMyJob(job *api.JobInfo) error {
	_, err := util.GetJobReqNPUNum(job, cn.AnnoName)
	if err != nil {
		return errors.New(JobNoNPUCard)
	}

	return nil
}
