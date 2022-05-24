/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package card310x4 is using for HuaWei A300T Ascend pin affinity schedule.

*/
package card310x4

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/rescheduling"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/conf"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/common"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/util"
)

// Name This need by frame init plugin.
func (tp *card310x4) Name() string {
	return PluginName
}

// New return npu plugin.
func New(npuName string) plugin.HwNPUSchedulerPlugin {
	pluginCard := card310x4{}
	pluginCard.com = common.Scheduler{
		PluginName:                npuName,
		AnnoName:                  pluginCard.GetResourceName(),
		AnnoPreVal:                pluginCard.GetResourcePreVal(),
		DefaultJobSchedulerConfig: pluginCard.GetPluginDefaultJobSchedulerConfig(),
	}
	pluginCard.re = common.ReScheduler{AnnoUnHealthy: a310FaultNPUName,
		IsMyJob: pluginCard.com.IsMyJob, AnnoName: rescheduling.AscendNPUPodRealUse}
	return &pluginCard
}

// OnHandlerStart The npu scheduler policy initial and common processing.
func (tp *card310x4) OnHandlerStart(sHandler *plugin.ScheduleHandler) {
	klog.V(util.LogDebugLev).Infof("%v start handler.", PluginName)
	sHandler.AddInitNodesNPUAllocTopology(PluginName, initNodesNPUTopologyFn)
	sHandler.AddPreHandleFaultNPU(tp.com.AnnoName, tp.re.PreHandleFaultNPUFn)
}

// ValidNPUJobFn Check the compliance of the selector and resource request numbers of job.
func (tp *card310x4) ValidNPUJobFn(job *api.JobInfo) *api.ValidateResult {
	// 1.Validate npu job selector.
	if err := validNPUJobSelector(job); err != nil {
		klog.V(util.LogDebugLev).Infof("%s validNPUJobSelector err: %v.", PluginName, err)
		return &api.ValidateResult{
			Pass:    false,
			Reason:  err.Error(),
			Message: fmt.Sprintf("validNPUJob err: %v", err),
		}
	}

	// 2.Validate job npu number.
	if jobError := validJobNPUNum(job); jobError != nil {
		klog.V(util.LogDebugLev).Infof("%s validJobNPUNum err: %v.", PluginName, jobError)
		return &api.ValidateResult{
			Pass:    false,
			Reason:  "job require npu number illegal",
			Message: fmt.Sprintf("%s, err: %v", job.Name, jobError),
		}
	}
	// 3.Validate job scheduler-strategy.
	if errJob := validJobModel(job); errJob != nil {
		klog.V(util.LogDebugLev).Infof("%s validJobModel err: %v.", PluginName, errJob)
		return &api.ValidateResult{
			Pass:    false,
			Reason:  "job scheduler-strategy error",
			Message: fmt.Sprintf("%s, err: %v", job.Name, errJob),
		}
	}

	return nil
}

// PreCheckNodeFn 310 no need to Distinguish between architecture.
func (tp *card310x4) PreCheckNodeFn(task *api.TaskInfo, node *api.NodeInfo, confs []conf.Configuration) error {
	schedulerConf := util.GetSchedulerSelectorConfig(confs)
	if len(schedulerConf) == 0 {
		// get scheduler selector configure failed, but need continue
		klog.V(util.LogDebugLev).Infof("%s JobUID: %s get selector nil.", PluginName, task.Name)
		return fmt.Errorf("%s get scheduler selector nil", node.Name)
	}

	// select node by architect
	if err := util.IsSelectorMeetNode(task, node, schedulerConf, a310NPUCardName); err != nil {
		// get scheduler selector configure failed, but need continue
		klog.V(util.LogDebugLev).Infof("%s taskName: %s ,nodeName %s : %v.", PluginName, task.Name, node.Name, err)
		return err
	}
	return nil
}

// CheckNPUResourceStableFn Check whether the node's NPU resources are stable.
func (tp *card310x4) CheckNPUResourceStableFn(node *api.NodeInfo) error {
	// default is the npu task
	nodeNPUIdleNumFromTop, err := getNodeNPUNumFromOthers(node)
	if err != nil {
		return fmt.Errorf("getNodeNPUNumFromOthers %s : %v", nodesNoMeetNPUReqError, err)
	}

	nodeNPUIdleNumFromIdle, err := util.GetNodeNPUNumFromIdle(node, a310NPUCardName)
	if err != nil {
		return fmt.Errorf("getNodeNPUNumFromIdle %s : %v", nodesNoMeetNPUReqError, err)
	}

	if err = util.CheckNodeNPUStabilize(nodeNPUIdleNumFromTop, nodeNPUIdleNumFromIdle); err != nil {
		return fmt.Errorf("%s : %v", nodeNotStableWarning, err)
	}

	return nil
}

// CheckNodeNPUByTaskFn Check whether the requested resource exists and are sufficient on the node.
func (tp *card310x4) CheckNodeNPUByTaskFn(task *api.TaskInfo, node *api.NodeInfo, _ bool) error {
	taskNPU, taskError := util.GetTaskNPUNum(task, a310NPUCardName)
	if taskError != nil {
		return fmt.Errorf("getTaskNPUNum %s : %v", nodesNoMeetNPUReqError, taskError)
	}

	nodeNPUTopology := util.GetTopFromNodeOthers(node, a310NPUCardName, a310NPUCardPreName)
	if nodeNPUTopology == nil {
		// node has none npu
		klog.V(util.LogInfoLev).Infof("%s checkNodeNPUByTask nil,node name:%s(top:%v),task req npu:%d",
			PluginName, node.Name, nodeNPUTopology, taskNPU)
		return fmt.Errorf("%s:get npu nil", nodeNotEnoughNPUWarning)
	}
	klog.V(util.LogInfoLev).Infof("%s %s top:%v,req %d", PluginName, node.Name, nodeNPUTopology, taskNPU)

	err := judgeNodeAndTaskNPU(taskNPU, nodeNPUTopology)
	if err != nil {
		return fmt.Errorf("judgeNodeAndTaskNPU %s : %v", nodeNotMeetTopologyWarning, err)
	}

	return nil
}

// GetNPUAffinityBestNodesFn Initialize a mapping between nodes and priorities.
func (tp *card310x4) GetNPUAffinityBestNodesFn(task *api.TaskInfo,
	nodes []*api.NodeInfo, _ bool) (map[string]int, error) {
	// 1. init 4 prioritized node-list array.
	priNodeGroups, err := initPriNodeGroups(task, nodes)
	if err != nil {
		klog.V(util.LogDebugLev).Infof("%s initPriNodeGroups failed :%v", PluginName, err)
		return nil, err
	}
	// 2.get the bestNodes map by taskReqNPU
	bestNodesMap, err := getBestNodesMap(priNodeGroups)
	if err != nil {
		klog.V(util.LogDebugLev).Infof("%s getBestNodesMap failed :%v", PluginName, err)
		return nil, err
	}

	klog.V(util.LogInfoLev).Infof("%s getNPUAffinityBestNodes %s:%v", PluginName, task.Name, bestNodesMap)
	return bestNodesMap, nil
}

// ScoreBestNPUNodesFn Used for score candidate nodes.
func (tp *card310x4) ScoreBestNPUNodesFn(scoreMap map[string]float64,
	bestNodes map[string]int,
	_ *api.TaskInfo,
	_ []*api.NodeInfo) (map[string]float64, error) {

	// parameters check
	if len(scoreMap) == 0 || reflect.ValueOf(scoreMap).IsNil() {
		err := errors.New("ScoreBestNPUNodesFn scoreMap is nil")
		klog.V(util.LogInfoLev).Infof("%s %v", PluginName, err)
		return nil, err
	}

	// the score value ranges from 1 to 32 (4 is best)
	for nodeName, priority := range bestNodes {
		scoreMap[nodeName] = constNPUWeight * (cardNPUNumber - float64(priority))
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
	nodeDeviceIDs := util.GetTopFromNodeOthers(node, a310NPUCardName, a310NPUCardPreName)
	if len(nodeDeviceIDs) == 0 {
		klog.V(util.LogDebugLev).Infof("%s useAnnotation node(%s) top nil.", PluginName, node.Name)
		return errors.New("nodeDeviceIDs nil")
	}

	// delete the use top
	klog.V(util.LogInfoLev).Infof("%s useAnnotation %s:%v , will use: %v.", PluginName, node.Name, nodeDeviceIDs, useTop)
	newNodeTopStr := util.GetRealTopAfterAlloc(nodeDeviceIDs, useTop, a310NPUCardPreName)
	if newNodeTopStr == "" {
		klog.V(util.LogDebugLev).Infof("%s getRealTopAfterAlloc all top has allocated .", PluginName)
	}

	err := util.ReloadNewTopToNodeOther(node, newNodeTopStr, a310NPUCardName)
	if err != nil {
		klog.V(util.LogDebugLev).Infof("%s reloadNewTopToNode failed.", PluginName)
		return err
	}

	klog.V(util.LogInfoLev).Infof("%s ReloadNewTopToNode %s to %s successes.", PluginName, newNodeTopStr, node.Name)
	return nil
}

// GetReleaseNPUTopologyFn Get the release npu card id from task(pod).
func (tp *card310x4) GetReleaseNPUTopologyFn(task *api.TaskInfo) (interface{}, error) {
	// get task use top
	taskDeviceIDs := util.GetDeviceIDsFromAnnotations(task.Pod.Annotations, a310NPUCardName, a310NPUCardPreName)
	if taskDeviceIDs == nil {
		klog.V(util.LogDebugLev).Infof("%s GetReleaseNPUTopologyFn failed task:%s", PluginName, task.Name)
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
	nodeDeviceIDs := util.GetTopFromNodeOthers(node, a310NPUCardName, a310NPUCardPreName)
	if nodeDeviceIDs == nil {
		klog.V(util.LogDebugLev).Infof("%s useAnnotation node(%s) top nil", PluginName, node.Name)
		return fmt.Errorf("%s has nil npu", node.Name)
	}
	// delete the use top
	newNodeTopStr := util.GetRealTopAfterRelease(nodeDeviceIDs, taskDeviceIDs, a310NPUCardPreName)
	if newNodeTopStr == "" {
		klog.V(util.LogDebugLev).Infof("%s getRealTopAfterRelease top failed", PluginName)
		return fmt.Errorf("%s release nil npu", node.Name)
	}

	err := util.ReloadNewTopToNodeOther(node, newNodeTopStr, a310NPUCardName)
	if err != nil {
		klog.V(util.LogDebugLev).Infof("%s reloadNewTopToNode failed", PluginName)
		return err
	}

	klog.V(util.LogInfoLev).Infof("%s useAnnotation node(%s) top(%s) successes", PluginName, node.Name, newNodeTopStr)

	return nil
}

// GetAllocatedNPUFromTopologyFn Get the pod's npu card to record in node others.
func (tp *card310x4) GetAllocatedNPUFromTopologyFn(
	task *api.TaskInfo, node *api.NodeInfo, _ bool) (interface{}, error) {
	var allocTopologyHccl []int
	var allocTopologyNPUs []int

	taskNPUNumber, taskError := util.GetTaskNPUNum(task, a310NPUCardName)
	if taskError != nil {
		return nil, fmt.Errorf("%s %v", task.Name, taskError)
	}

	priorityArray, err := getNPUAllocPriorityArray(taskNPUNumber)
	if err != nil {
		return allocTopologyHccl, err
	}

	nodeTop := util.GetTopFromNodeOthers(node, a310NPUCardName, a310NPUCardPreName)
	if nodeTop == nil {
		klog.V(util.LogDebugLev).Infof("not npu node[%s], no need to continue.", node.Name)
		return allocTopologyHccl, err
	}
	klog.V(util.LogInfoLev).Infof("%s %s[%d] priority:%v in %v.", PluginName,
		task.Name, taskNPUNumber, priorityArray, nodeTop)

	allocTopologyHccl, err = getFitCardFromNodeByPriority(nodeTop, priorityArray)
	if err != nil {
		err = fmt.Errorf("node %v not meet req: %d", nodeTop, taskNPUNumber)
		klog.V(util.LogDebugLev).Infof("%s %s.", PluginName, err.Error())
		return allocTopologyHccl, err
	}
	klog.V(util.LogDebugLev).Infof("%s %s get alloc %v.", PluginName, task.Name, allocTopologyHccl)

	allocTopologyNPUs, err = util.GetNPUTopFromHccs(taskNPUNumber, allocTopologyHccl)
	if err != nil {
		return allocTopologyNPUs, err
	}
	klog.V(util.LogInfoLev).Infof("%s %s req:%d alloc %v.", PluginName, task.Name, taskNPUNumber, allocTopologyNPUs)
	return allocTopologyNPUs, nil
}

// SetNPUTopologyToPodFn Set the npu card ids into pod.
func (tp *card310x4) SetNPUTopologyToPodFn(task *api.TaskInfo, top interface{}) error {
	var topologyStr string

	klog.V(util.LogInfoLev).Infof("%s setNPUTopologyToPod begin top:%v", PluginName, top)
	intTop, ok := top.([]int)
	if !ok {
		return errors.New(argumentError)
	}

	topologyStr = util.ChangeIntArrToStr(intTop, a310NPUCardPreName)
	task.Pod.Annotations[a310NPUCardName] = topologyStr
	// to device-plugin judge pending pod.
	task.Pod.Annotations[podPredicateTime] = strconv.FormatInt(time.Now().UnixNano(), util.ConstIntNum10)
	klog.V(util.LogInfoLev).Infof("%s setNPUTopologyToPod %s top:%s", PluginName, task.Name, topologyStr)

	return nil
}

// IsMyTask Determine if it is the NPU task of your plug-in.
func (tp *card310x4) IsMyTask(task *api.TaskInfo) error {
	_, err := util.GetTaskNPUNum(task, a310NPUCardName)
	if err != nil {
		return errors.New(jobNoNPUCard)
	}

	if !IsTaskOfCardModeFromLabel(task) {
		return errors.New(modeNotCard)
	}

	return nil
}

// IsMyNode Determine if it is the NPU node of your plug-in.
func (tp *card310x4) IsMyNode(node *api.NodeInfo) error {
	_, err := util.GetNPUAllocCardsFromNodeOthers(node, a310NPUCardName)
	if err != nil {
		return errors.New(jobNoNPUCard + err.Error())
	}

	return nil
}

// IsMyJob Determine if it is the NPU job of your plug-in.
func (tp *card310x4) IsMyJob(job *api.JobInfo) error {
	_, err := util.GetJobReqNPUNum(job, a310NPUCardName)
	if err != nil {
		return errors.New(jobNoNPUCard)
	}

	if !IsJobOfCardModeFromLabel(job) {
		return errors.New(modeNotCard)
	}

	return nil
}

// GetResourceName get plugin NPU resource name.
func (tp *card310x4) GetResourceName() string {
	return a310NPUCardName
}

// GetResourcePreVal get plugin NPU resource name prefix.
func (tp *card310x4) GetResourcePreVal() string {
	return a310NPUCardPreName
}

// GetPluginDefaultJobSchedulerConfig get plugin default job scheduler config.
func (tp *card310x4) GetPluginDefaultJobSchedulerConfig() map[string]string {
	defaultSchedulerConfig := make(map[string]string, util.NPUIndex1)
	defaultSchedulerConfig[archSelector] = huaweiArchArm + "|" + huaweiArchX86
	defaultSchedulerConfig[acceleratorType] = cardAcceleratorType + "|" + chipAcceleratorType
	return defaultSchedulerConfig
}
