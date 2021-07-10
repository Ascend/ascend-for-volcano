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

Package module910x8 is using for HuaWei A800/9000 Ascend910 pin affinity schedule.

*/
package module910x8

import (
	"errors"
	"fmt"
	"k8s.io/klog"
	"reflect"
	"strconv"
	"time"
	"volcano.sh/volcano/pkg/scheduler/api"
	vapi "volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/conf"
	"volcano.sh/volcano/pkg/scheduler/framework"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	npuutil "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/util"
)

// This need by frame init plugin.
func (tp *module910x8) Name() string {
	return PluginName
}

// New return npu plugin.
func New(npuName string) plugin.HwNPUSchedulerPlugin {
	return &module910x8{name: npuName}
}

// The npu scheduler policy initial and common processing.
func (tp *module910x8) OnHandlerStart(sHandler *plugin.ScheduleHandler) {
	klog.V(logErrorLev).Infof("%v start Handler.", PluginName)
	sHandler.AddInitNodesNPUAllocTopology(PluginName, initNodesNPUTopologyFn)
	// Only used by deal fault npu.
	sHandler.AddPreHandleFaultNPU(PluginName, preHandleFaultNPUFn)
}

// Check the compliance of the selector and resource request numbers of job.
func (tp *module910x8) ValidNPUJobFn(job *vapi.JobInfo) *vapi.ValidateResult {
	// 1.validate npu job selector
	if err := validNPUJobSelector(job); err != nil {
		klog.V(logErrorLev).Infof("%s validNPUJobSelector err: %v.", PluginName, err)
		return &vapi.ValidateResult{
			Pass:    false,
			Reason:  err.Error(),
			Message: fmt.Sprintf("validNPUJob err: %v", err),
		}
	}
	// 2.validate job npu number
	if jobError := validJobNPUNum(job); jobError != nil {
		klog.V(logErrorLev).Infof("%s validJobNPUNum err: %v.", PluginName, jobError)
		return &vapi.ValidateResult{
			Pass:    false,
			Reason:  "job require npu number illegal",
			Message: fmt.Sprintf("%s, err: %v", job.Name, jobError),
		}
	}
	// 3.validate job scheduler-strategy
	if errJob := validJobModel(job); errJob != nil {
		klog.V(logErrorLev).Infof("%s validJobModel err: %v.", PluginName, errJob)
		return &vapi.ValidateResult{
			Pass:    false,
			Reason:  "job scheduler-strategy error",
			Message: fmt.Sprintf("%s, err: %v", job.Name, errJob),
		}
	}

	return nil
}

// Get the nodes that meet the task requirements.
func (tp *module910x8) PreCheckNodeFn(task *vapi.TaskInfo, node *vapi.NodeInfo, confs []conf.Configuration) error {
	schedulerConf := npuutil.GetSchedulerSelectorConfig(confs)
	if len(schedulerConf) == 0 {
		// get scheduler selector configure failed, but need continue
		klog.V(logErrorLev).Infof("%s JobName: %s get selector nil.", PluginName, task.Name)
		return fmt.Errorf("%s get scheduler selector nil", node.Name)
	}

	// select node by architect
	if err := npuutil.IsSelectorMeetNode(task, node, schedulerConf, npu800And9000CardName); err != nil {
		// get scheduler selector configure failed, but need continue
		klog.V(logErrorLev).Infof("%s taskName: %s ,nodeName %s : %v.", PluginName, task.Name, node.Name, err)
		return err
	}

	return nil
}

// Check whether the node's NPU resources are stable.
func (tp *module910x8) CheckNPUResourceStableFn(node *vapi.NodeInfo) error {
	// default is the npu task
	nodeNPUIdleNumFromTop, err := getNodeNPUNumFromAnnotation(node)
	if err != nil {
		return fmt.Errorf("getNodeNPUNumFromAnnotation %s : %s", nodesNoMeetNPUReqError, err)
	}

	nodeNPUIdleNumFromIdle, err := npuutil.GetNodeNPUNumFromIdle(node, npu800And9000CardName)
	if err != nil {
		return fmt.Errorf("getNodeNPUNumFromIdle %s : %s", nodesNoMeetNPUReqError, err)
	}

	if err = npuutil.CheckNodeNPUStabilize(nodeNPUIdleNumFromTop, nodeNPUIdleNumFromIdle); err != nil {
		return fmt.Errorf("%s : %s", nodeNotStableWarning, err)
	}

	return nil
}

// Check whether the requested resource exists and are sufficient on the node.
func (tp *module910x8) CheckNodeNPUByTaskFn(task *vapi.TaskInfo, node *vapi.NodeInfo) error {
	taskNPU, taskError := npuutil.GetTaskNPUNum(task, npu800And9000CardName)
	if taskError != nil {
		return fmt.Errorf("getTaskNPUNum %s : %s", nodesNoMeetNPUReqError, taskError)
	}

	nodeNPUTopology := npuutil.GetTopFromNode(node, npu800And9000CardName, npu910CardPreName)
	if len(nodeNPUTopology) == 0 {
		// node has none npu
		klog.V(logInfoLev).Infof("%s checkNodeNPUByTask nil,node name:%s(top:%v),task req npu:%d.",
			PluginName, node.Name, nodeNPUTopology, taskNPU)
		return fmt.Errorf("%s:get npu nil", nodeNotEnoughNPUWarning)
	}
	klog.V(logInfoLev).Infof("%s %s top:%v,req %d.", PluginName, node.Name, nodeNPUTopology, taskNPU)

	err := judgeNodeAndTaskNPU(taskNPU, nodeNPUTopology)
	if err != nil {
		return fmt.Errorf("judgeNodeAndTaskNPU %s : %v", nodeNotMeetTopologyWarning, err)
	}

	faultJobErr := checkFaultJobNode(task, node)
	if faultJobErr != nil {
		return fmt.Errorf("checkFaultJobNode %s : %v", nodesNoMeetNPUReqError, faultJobErr)
	}

	return nil
}

// Initialize a mapping between nodes and priorities.
func (tp *module910x8) GetNPUAffinityBestNodesFn(
	task *vapi.TaskInfo,
	nodes []*vapi.NodeInfo) (map[string]int, error) {
	// 1. init 4 prioritized node-list array.
	priNodeGroups, err := initPriNodeGroups(task, nodes)
	if err != nil {
		klog.V(logErrorLev).Infof("%s initPriNodeGroups failed :%s.", PluginName, err)
		return nil, err
	}
	// 2.get the bestNodes map by taskReqNPU
	bestNodesMap, err := getBestNodesMap(priNodeGroups)
	if err != nil {
		klog.V(logErrorLev).Infof("%s getBestNodesMap failed :%s.", PluginName, err)
		return nil, err
	}

	klog.V(logInfoLev).Infof("%s getNPUAffinityBestNodes %s:%v.", PluginName, task.Name, bestNodesMap)
	return bestNodesMap, nil
}

// ScoreBestNPUNodesFn Used for score candidate nodes.
func (tp *module910x8) ScoreBestNPUNodesFn(scoreMap map[string]float64,
	bestNodes map[string]int,
	task *api.TaskInfo,
	nodes []*vapi.NodeInfo) (map[string]float64, error) {
	var nodeWeight = 1.0

	// parameters check
	if reflect.ValueOf(scoreMap).IsNil() {
		err := errors.New("scoreBestNPUNodes's scoreMap is nil")
		klog.V(logInfoLev).Infof("%s %v.", PluginName, err)
		return nil, err
	}

	for nodeName, priority := range bestNodes {
		healthNPUNumber, err := npuutil.GetNodeHealthNPUNumberByName(nodeName, nodes, npu800And9000CardName)
		if err != nil {
			scoreMap[nodeName] = 0.0
			klog.V(logInfoLev).Infof("%s getNodeHealthNPUNumberByName error:%v.", PluginName, err)
			continue
		}

		scoreMap[nodeName] = nodeWeight * (healthNPUNumber*npuNumPerHccs - float64(priority))
	}

	if scoreMap, err := npuutil.AddScoreByFaultNPUTask(task, scoreMap); err != nil {
		klog.V(logErrorLev).Infof("%s : %v.", PluginName, err)
		return scoreMap, err
	}
	return scoreMap, nil
}

// Get the pod's npu card to record in node others.
func (tp *module910x8) GetAllocatedNPUFromTopologyFn(task *vapi.TaskInfo, node *vapi.NodeInfo) (interface{}, error) {
	var allocTopologyHccl []int
	var allocTopologyNPUs []int

	taskNPUNumber, taskError := npuutil.GetTaskNPUNum(task, npu800And9000CardName)
	if taskError != nil {
		return nil, errors.New("no npu task")
	}

	priorityArray, err := getNPUAllocPriorityArray(taskNPUNumber)
	if err != nil {
		return allocTopologyHccl, err
	}

	nodeTop := npuutil.GetTopFromNode(node, npu800And9000CardName, npu910CardPreName)
	if nodeTop == nil {
		klog.V(logErrorLev).Infof("module910x8 not npu node[%s], no need to continue.", node.Name)
		return allocTopologyHccl, err
	}
	klog.V(logInfoLev).Infof("module910x8 %s %s[%d] priority:%v in %v.", PluginName,
		task.Name, taskNPUNumber, priorityArray, nodeTop)

	allocTopologyHccl, err = getHccsFromNodeByPriority(nodeTop, priorityArray)
	if err != nil {
		err = fmt.Errorf("node %v not meet req: %d", nodeTop, taskNPUNumber)
		klog.V(logErrorLev).Infof("module910x8 %s %s.", PluginName, err.Error())
		return allocTopologyHccl, err
	}
	klog.V(logDebugLev).Infof("module910x8 %s %s get top %v.", PluginName, task.Name, allocTopologyHccl)

	allocTopologyNPUs, err = npuutil.GetNPUTopFromHccs(taskNPUNumber, allocTopologyHccl)
	if err != nil {
		return allocTopologyNPUs, err
	}
	klog.V(logInfoLev).Infof("%s %s req:%d alloc %v.", PluginName, task.Name, taskNPUNumber, allocTopologyNPUs)
	return allocTopologyNPUs, nil
}

// Set the npu card ids into pod.
func (tp *module910x8) SetNPUTopologyToPodFn(task *vapi.TaskInfo, top interface{}) error {
	var topologyStr string

	intTop, ok := top.([]int)
	if !ok {
		klog.V(logErrorLev).Infof("%s setNPUTopologyToPod top:%v.", PluginName, top)
		return errors.New(argumentError)
	}

	topologyStr = npuutil.ChangeIntArrToStr(intTop, npu910CardPreName)
	task.Pod.Annotations[npu800And9000CardName] = topologyStr
	// to device-plugin judge pending pod.
	task.Pod.Annotations[podPredicateTime] = strconv.FormatInt(time.Now().UnixNano(), 10)
	klog.V(logInfoLev).Infof("%s setNPUTopologyToPod %s top:%s.", PluginName, task.Name, topologyStr)

	tmpValue, ok := npuutil.ReSchedulerJobs[task.Job]
	if !ok {
		klog.V(logInfoLev).Infof("%s setNPUTopologyToPod %s not npu fault task.", PluginName, task.Name)
		return nil
	}

	klog.V(logDebugLev).Infof("%s %s setNPUTopologyToPodFn from buffer: %v.", PluginName, task.Job, tmpValue)
	for taskName, rankIndex := range tmpValue.RankIndexes {
		if taskName == task.Name {
			klog.V(logInfoLev).Infof("%s set %s rankIndex %v.", PluginName, task.Pod.Name, rankIndex)
			task.Pod.Annotations[podRankIndex] = rankIndex
			break
		}
	}
	return nil
}

// Update used npu resources on node.
func (tp *module910x8) UpdateNPUNodeUsedCardFn(node *vapi.NodeInfo, top interface{}) error {
	useTop, ok := top.([]int)
	if !ok {
		return errors.New(argumentError)
	}

	// get node available top
	nodeDeviceIDs := npuutil.GetDeviceIDsFromNodeOther(node.Others, npu800And9000CardName, npu910CardPreName)
	if nodeDeviceIDs == nil {
		klog.V(logErrorLev).Infof("%s useAnnotation node(%s) top nil.", PluginName, node.Name)
		return errors.New("nodeDeviceIDs nil")
	}

	// delete the use top
	klog.V(logInfoLev).Infof("%s useAnnotation %s:%v , will use: %v.", PluginName, node.Name, nodeDeviceIDs, useTop)
	newNodeTopStr := npuutil.GetRealTopAfterAlloc(nodeDeviceIDs, useTop, npu910CardPreName)
	if newNodeTopStr == "" {
		klog.V(logDebugLev).Infof("%s getRealTopAfterAlloc all top has allocated .", PluginName)
	}

	err := npuutil.ReloadNewTopToNodeOther(node, newNodeTopStr, npu800And9000CardName)
	if err != nil {
		klog.V(logErrorLev).Infof("%s reloadNewTopToNode failed.", PluginName)
		return err
	}

	klog.V(logInfoLev).Infof("%s ReloadNewTopToNode %s to %s successes.", PluginName, newNodeTopStr, node.Name)

	return nil
}

// Get the release npu card id from task(pod).
func (tp *module910x8) GetReleaseNPUTopologyFn(task *vapi.TaskInfo) (interface{}, error) {
	// get task use top
	taskDevIDs := npuutil.GetDeviceIDsFromAnnotations(task.Pod.Annotations, npu800And9000CardName, npu910CardPreName)
	if taskDevIDs == nil {
		klog.V(logErrorLev).Infof("%s releaseAnnotation failed task:%s.", PluginName, task.Name)
		return nil, fmt.Errorf("%s get npu nil", task.Name)
	}

	return taskDevIDs, nil
}

// Update the node using npu when release pod's npu.
func (tp *module910x8) UpdateReleaseNPUNodeTopologyFn(node *vapi.NodeInfo, top interface{}) error {
	taskDeviceIDs, ok := top.([]int)
	if !ok {
		return errors.New(argumentError)
	}

	// get node available top
	nodeDeviceIDs := npuutil.GetDeviceIDsFromNodeOther(node.Others, npu800And9000CardName, npu910CardPreName)
	if nodeDeviceIDs == nil {
		klog.V(logErrorLev).Infof("%s useAnnotation node(%s) top nil.", PluginName, node.Name)
		return fmt.Errorf("%s has nil npu", node.Name)
	}
	// delete the use top
	newNodeTopStr := npuutil.GetRealTopAfterRelease(nodeDeviceIDs, taskDeviceIDs, npu800And9000CardName)
	if newNodeTopStr == "" {
		klog.V(logErrorLev).Infof("%s getRealTopAfterRelease top failed.", PluginName)
		return fmt.Errorf("%s release nil npu", node.Name)
	}

	err := npuutil.ReloadNewTopToNodeOther(node, newNodeTopStr, npu800And9000CardName)
	if err != nil {
		klog.V(logErrorLev).Infof("%s reloadNewTopToNode failed.", PluginName)
		return err
	}

	klog.V(logInfoLev).Infof("%s useAnnotation node(%s) top(%s) successes.", PluginName, node.Name, newNodeTopStr)

	return nil
}

// Determine if it is the NPU task of your plug-in.
func (tp *module910x8) IsMyTask(task *vapi.TaskInfo) error {
	_, err := npuutil.GetTaskNPUNum(task, npu800And9000CardName)
	if err != nil {
		return errors.New(jobNoNPUCard)
	}

	if npuutil.IsTaskOfCardMode(task) {
		return errors.New("task is card mode")
	}

	return nil
}

// Determine if it is the NPU node of your plug-in.
func (tp *module910x8) IsMyNode(node *vapi.NodeInfo) error {
	_, err := npuutil.GetNodeNPUAllocCards(node, npu800And9000CardName)
	if err != nil {
		return fmt.Errorf("%s %s", node.Name, jobNoNPUCard)
	}

	if npuutil.IsCardModeNode(node) {
		return fmt.Errorf("%s is card mode", node.Name)
	}

	return nil
}

// Determine if it is the NPU job of your plug-in.
func (tp *module910x8) IsMyJob(job *vapi.JobInfo) error {
	return isMyJob(job)
}

func preHandleFaultNPUFn(ssn *framework.Session) error {
	var faultNPUs []nodeFaultNPUs
	var faultNPUJobs []faultNPUJob

	klog.V(logDebugLev).Infof("%s enter preHandleFaultNPUFn.", PluginName)
	defer klog.V(logDebugLev).Infof("%s leave preHandleFaultNPUFn.", PluginName)

	// 1.Determine if it is a 910 jobs.
	jobs, err := get910x8Jobs(ssn.Jobs)
	if err != nil {
		klog.V(logDebugLev).Infof("%s get910x8Jobs: %v.", PluginName, err)
		return nil
	}
	// 2.Get fault npus and its nodes from running vcjob.
	faultNPUs, err = getInoperableNPUs(ssn.Nodes)
	if err != nil {
		klog.V(logDebugLev).Infof("%s getInoperableNPUs %v.", PluginName, err)
		return nil
	}
	// 3.Get fault vcjobs and its node-rankIndex.
	faultNPUJobs, err = getFaultNPUJobs(jobs, faultNPUs)
	if err != nil {
		klog.V(logErrorLev).Infof("%s getFaultNPUJobs %v.", PluginName, err)
		return err
	}
	// 4.Sets the node reschedule occupancy flag.
	err = setFaultLabelOnNodeAndJob(faultNPUJobs, jobs)
	if err != nil {
		klog.V(logErrorLev).Infof("%s setFaultLabelOnNodeAndJob %v.", PluginName, err)
		return err
	}
	// 5.Restart vcjobs.
	err = restartNPUFaultJob(ssn, faultNPUJobs, jobs)
	if err != nil {
		klog.V(logErrorLev).Infof("%s restartNPUFaultJob %v.", PluginName, err)
		return err
	}

	return nil
}
