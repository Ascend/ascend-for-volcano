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

Package card310x4 is using for HuaWei A300T Ascend pin affinity schedule.

*/
package card310x4

import (
	"errors"
	"fmt"
	"k8s.io/klog"
	"reflect"
	"strconv"
	"time"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/conf"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	hwutil "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/card310x4/util"
)

// This need by frame init plugin.
func (tp *card310x4) Name() string {
	return PluginName
}

// New return npu plugin.
func New(npuName string) plugin.HwNPUSchedulerPlugin {
	return &card310x4{name: npuName}
}

// The npu scheduler policy initial and common processing.
func (tp *card310x4) OnHandlerStart(sHandler *plugin.ScheduleHandler) {
	klog.V(logDebugLev).Infof("%v start handler.", PluginName)
	sHandler.AddInitNodesNPUAllocTopology(PluginName, initNodesNPUTopologyFn)
}

// ValidNPUJobFn Check the compliance of the selector and resource request numbers of job.
func (tp *card310x4) ValidNPUJobFn(job *api.JobInfo) *api.ValidateResult {
	// 1.Validate npu job selector.
	if err := validNPUJobSelector(job); err != nil {
		klog.V(logErrorLev).Infof("%s validNPUJobSelector err: %v.", PluginName, err)
		return &api.ValidateResult{
			Pass:    false,
			Reason:  err.Error(),
			Message: fmt.Sprintf("validNPUJob err: %v", err),
		}
	}

	// 2.Validate job npu number.
	if jobError := validJobNPUNum(job); jobError != nil {
		klog.V(logErrorLev).Infof("%s validJobNPUNum err: %v.", PluginName, jobError)
		return &api.ValidateResult{
			Pass:    false,
			Reason:  "job require npu number illegal",
			Message: fmt.Sprintf("%s, err: %v", job.Name, jobError),
		}
	}
	// 3.Validate job scheduler-strategy.
	if errJob := validJobModel(job); errJob != nil {
		klog.V(logErrorLev).Infof("%s validJobModel err: %v.", PluginName, errJob)
		return &api.ValidateResult{
			Pass:    false,
			Reason:  "job scheduler-strategy error",
			Message: fmt.Sprintf("%s, err: %v", job.Name, errJob),
		}
	}

	return nil
}

// PreCheckNodeFn Get the nodes that meet the task requirements.
func (tp *card310x4) PreCheckNodeFn(task *api.TaskInfo, node *api.NodeInfo, confs []conf.Configuration) error {
	schedulerConf := hwutil.GetSchedulerSelectorConfig(confs)
	if len(schedulerConf) == 0 {
		// get scheduler selector configure failed, but need continue
		klog.V(logErrorLev).Infof("%s JobName: %s get selector nil", PluginName, task.Name)
		return fmt.Errorf("%s get scheduler selector nil", node.Name)
	}

	defaultSchedulerConfig := getCardNPUNodeDefaultSelectorConfig()
	klog.V(logDebugLev).Infof("%s card selector: %v default:%v.", node.Name, schedulerConf, defaultSchedulerConfig)

	if err := hwutil.IsSelectorMeetNode(task, node, defaultSchedulerConfig, schedulerConf, a310NPUCardName); err != nil {
		// get scheduler selector configure failed, but need continue
		klog.V(logErrorLev).Infof("%s taskName: %s ,nodeName %s : %v.", PluginName, task.Name, node.Name, err)
		return err
	}

	return nil
}

// CheckNPUResourceStableFn Check whether the node's NPU resources are stable.
func (tp *card310x4) CheckNPUResourceStableFn(node *api.NodeInfo) error {
	// default is the npu task
	nodeNPUIdleNumFromTop, err := getNodeNPUNumFromAnnotation(node)
	if err != nil {
		return fmt.Errorf("getNodeNPUNumFromAnnotation %s : %s", nodesNoMeetNPUReqError, err)
	}

	nodeNPUIdleNumFromIdle, err := hwutil.GetNodeNPUNumFromIdle(node, a310NPUCardName)
	if err != nil {
		return fmt.Errorf("getNodeNPUNumFromIdle %s : %s", nodesNoMeetNPUReqError, err)
	}

	if err = hwutil.CheckNodeNPUStabilize(nodeNPUIdleNumFromTop, nodeNPUIdleNumFromIdle); err != nil {
		return fmt.Errorf("%s : %s", nodeNotStableWarning, err)
	}

	return nil
}

// CheckNodeNPUByTaskFn Check whether the requested resource exists and are sufficient on the node.
func (tp *card310x4) CheckNodeNPUByTaskFn(task *api.TaskInfo, node *api.NodeInfo) error {
	taskNPU, taskError := hwutil.GetTaskNPUNum(task, a310NPUCardName)
	if taskError != nil {
		return fmt.Errorf("getTaskNPUNum %s : %s", nodesNoMeetNPUReqError, taskError)
	}

	nodeNPUTopology := hwutil.GetTopFromNode(node, a310NPUCardName, a310NPUCardPreName)
	if len(nodeNPUTopology) == 0 {
		// node has none npu
		klog.V(logInfoLev).Infof("%s checkNodeNPUByTask nil,node name:%s(top:%v),task req npu:%d",
			PluginName, node.Name, nodeNPUTopology, taskNPU)
		return fmt.Errorf("%s:get npu nil", nodeNotEnoughNPUWarning)
	}
	klog.V(logInfoLev).Infof("%s %s top:%v,req %d", PluginName, node.Name, nodeNPUTopology, taskNPU)

	err := judgeNodeAndTaskNPU(taskNPU, nodeNPUTopology)
	if err != nil {
		return fmt.Errorf("judgeNodeAndTaskNPU %s : %v", nodeNotMeetTopologyWarning, err)
	}

	return nil
}

// GetNPUAffinityBestNodesFn Initialize a mapping between nodes and priorities.
func (tp *card310x4) GetNPUAffinityBestNodesFn(task *api.TaskInfo, nodes []*api.NodeInfo) (map[string]int, error) {
	// 1. init 4 prioritized node-list array.
	priNodeGroups, err := initPriNodeGroups(task, nodes)
	if err != nil {
		klog.V(logErrorLev).Infof("%s initPriNodeGroups failed :%s", PluginName, err)
		return nil, err
	}
	// 2.get the bestNodes map by taskReqNPU
	bestNodesMap, err := getBestNodesMap(priNodeGroups)
	if err != nil {
		klog.V(logErrorLev).Infof("%s getBestNodesMap failed :%s", PluginName, err)
		return nil, err
	}

	klog.V(logInfoLev).Infof("%s getNPUAffinityBestNodes %s:%v", PluginName, task.Name, bestNodesMap)
	return bestNodesMap, nil
}

// ScoreBestNPUNodesFn Used for score candidate nodes.
func (tp *card310x4) ScoreBestNPUNodesFn(scoreMap map[string]float64,
	bestNodes map[string]int,
	task *api.TaskInfo,
	nodes []*api.NodeInfo) (map[string]float64, error) {
	var nodeWeight = 1.0

	// parameters check
	if reflect.ValueOf(scoreMap).IsNil() {
		err := errors.New("scoreBestNPUNodes's scoreMap is nil")
		klog.V(logInfoLev).Infof("%s %v", PluginName, err)
		return nil, err
	}

	// the score value ranges from 1 to 4 (4 is best)
	for nodeName, priority := range bestNodes {
		scoreMap[nodeName] = nodeWeight * (cardNPUNumber - float64(priority))
	}

	return scoreMap, nil
}

// UpdateNPUNodeUsedCardFn Update used npu resources on node.
func (tp *card310x4) UpdateNPUNodeUsedCardFn(node *api.NodeInfo, top interface{}) error {
	useTop, ok := top.([]int)
	if !ok {
		return errors.New(argumentError)
	}

	// get node available top
	nodeDeviceIDs := hwutil.GetDeviceIDsFromNodeOther(node.Others, a310NPUCardName, a310NPUCardPreName)
	if nodeDeviceIDs == nil {
		klog.V(logErrorLev).Infof("%s useAnnotation node(%s) top nil.", PluginName, node.Name)
		return errors.New("nodeDeviceIDs nil")
	}

	// delete the use top
	klog.V(logInfoLev).Infof("%s useAnnotation %s:%v , will use: %v.", PluginName, node.Name, nodeDeviceIDs, useTop)
	newNodeTopStr := hwutil.GetRealTopAfterAlloc(nodeDeviceIDs, useTop, a310NPUCardPreName)
	if newNodeTopStr == "" {
		klog.V(logDebugLev).Infof("%s getRealTopAfterAlloc all top has allocated .", PluginName)
	}

	err := hwutil.ReloadNewTopToNodeOther(node, newNodeTopStr, a310NPUCardName)
	if err != nil {
		klog.V(logErrorLev).Infof("%s reloadNewTopToNode failed.", PluginName)
		return err
	}

	klog.V(logInfoLev).Infof("%s ReloadNewTopToNode %s to %s successes.", PluginName, newNodeTopStr, node.Name)
	return nil
}

// GetReleaseNPUTopologyFn Get the release npu card id from task(pod).
func (tp *card310x4) GetReleaseNPUTopologyFn(task *api.TaskInfo) (interface{}, error) {
	// get task use top
	taskDeviceIDs := hwutil.GetDeviceIDsFromAnnotations(task.Pod.Annotations, a310NPUCardName, a310NPUCardPreName)
	if taskDeviceIDs == nil {
		klog.V(logErrorLev).Infof("%s releaseAnnotation failed task:%s", PluginName, task.Name)
		return nil, fmt.Errorf("%s get npu nil", task.Name)
	}

	return taskDeviceIDs, nil
}

// UpdateReleaseNPUNodeTopologyFn Update the node using npu when release pod's npu.
func (tp *card310x4) UpdateReleaseNPUNodeTopologyFn(node *api.NodeInfo, top interface{}) error {
	taskDeviceIDs, ok := top.([]int)
	if !ok {
		return errors.New(argumentError)
	}

	// get node available top
	nodeDeviceIDs := hwutil.GetDeviceIDsFromNodeOther(node.Others, a310NPUCardName, a310NPUCardPreName)
	if nodeDeviceIDs == nil {
		klog.V(logErrorLev).Infof("%s useAnnotation node(%s) top nil", PluginName, node.Name)
		return fmt.Errorf("%s has nil npu", node.Name)
	}
	// delete the use top
	newNodeTopStr := hwutil.GetRealTopAfterRelease(nodeDeviceIDs, taskDeviceIDs, a310NPUCardPreName)
	if newNodeTopStr == "" {
		klog.V(logErrorLev).Infof("%s getRealTopAfterRelease top failed", PluginName)
		return fmt.Errorf("%s release nil npu", node.Name)
	}

	err := hwutil.ReloadNewTopToNodeOther(node, newNodeTopStr, a310NPUCardName)
	if err != nil {
		klog.V(logErrorLev).Infof("%s reloadNewTopToNode failed", PluginName)
		return err
	}

	klog.V(logInfoLev).Infof("%s useAnnotation node(%s) top(%s) successes", PluginName, node.Name, newNodeTopStr)

	return nil
}

// GetAllocatedNPUFromTopologyFn Get the pod's npu card to record in node others.
func (tp *card310x4) GetAllocatedNPUFromTopologyFn(task *api.TaskInfo, node *api.NodeInfo) (interface{}, error) {
	var allocTopologyHccl []int
	var allocTopologyNPUs []int

	taskNPUNumber, taskError := hwutil.GetTaskNPUNum(task, a310NPUCardName)
	if taskError != nil {
		return nil, errors.New("no npu task")
	}

	priorityArray, err := getNPUAllocPriorityArray(taskNPUNumber)
	if err != nil {
		return allocTopologyHccl, err
	}

	nodeTop := hwutil.GetTopFromNode(node, a310NPUCardName, a310NPUCardPreName)
	if nodeTop == nil {
		klog.V(logErrorLev).Infof("not npu node[%s], no need to continue.", node.Name)
		return allocTopologyHccl, err
	}
	klog.V(logInfoLev).Infof("%s %s[%d] priority:%v in %v.", PluginName,
		task.Name, taskNPUNumber, priorityArray, nodeTop)

	allocTopologyHccl, err = getFitCardFromNodeByPriority(nodeTop, priorityArray)
	if err != nil {
		err = fmt.Errorf("node %v not meet req: %d", nodeTop, taskNPUNumber)
		klog.V(logErrorLev).Infof("%s %s.", PluginName, err.Error())
		return allocTopologyHccl, err
	}
	klog.V(logDebugLev).Infof("%s %s get top %v.", PluginName, task.Name, allocTopologyHccl)

	allocTopologyNPUs, err = hwutil.GetNPUTopFromHccs(taskNPUNumber, allocTopologyHccl)
	if err != nil {
		return allocTopologyNPUs, err
	}
	klog.V(logInfoLev).Infof("%s %s req:%d alloc %v.", PluginName, task.Name, taskNPUNumber, allocTopologyNPUs)
	return allocTopologyNPUs, nil
}

// SetNPUTopologyToPodFn Set the npu card ids into pod.
func (tp *card310x4) SetNPUTopologyToPodFn(task *api.TaskInfo, top interface{}) error {
	var topologyStr string

	klog.V(logInfoLev).Infof("%s setNPUTopologyToPod begin top:%v", PluginName, top)
	intTop, ok := top.([]int)
	if !ok {
		return errors.New(argumentError)
	}

	topologyStr = hwutil.ChangeIntArrToStr(intTop, a310NPUCardPreName)
	task.Pod.Annotations[a310NPUCardName] = topologyStr
	// to device-plugin judge pending pod.
	task.Pod.Annotations[podPredicateTime] = strconv.FormatInt(time.Now().UnixNano(), 10)
	klog.V(logInfoLev).Infof("%s setNPUTopologyToPod %s top:%s", PluginName, task.Name, topologyStr)

	return nil
}

// IsMyTask Determine if it is the NPU task of your plug-in.
func (tp *card310x4) IsMyTask(task *api.TaskInfo) error {
	_, err := hwutil.GetTaskNPUNum(task, a310NPUCardName)
	if err != nil {
		return errors.New(jobNoNPUCard)
	}

	if !hwutil.IsTaskOfCardMode(task) {
		return errors.New(modeNotCard)
	}

	return nil
}

// IsMyNode Determine if it is the NPU node of your plug-in.
func (tp *card310x4) IsMyNode(node *api.NodeInfo) error {
	_, err := hwutil.GetNodeNPUAllocCards(node, a310NPUCardName)
	if err != nil {
		return errors.New(jobNoNPUCard + err.Error())
	}

	return nil
}

// IsMyJob Determine if it is the NPU job of your plug-in.
func (tp *card310x4) IsMyJob(job *api.JobInfo) error {
	_, err := hwutil.GetJobReqNPUNum(job, a310NPUCardName)
	if err != nil {
		return errors.New(jobNoNPUCard)
	}

	if !hwutil.IsJobOfCardMode(job) {
		return errors.New(modeNotCard)
	}

	return nil
}
