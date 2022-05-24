/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
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
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/rescheduling"
	npuutil "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/util"
)

// Name This need by frame init plugin.
func (tp *module910x8) Name() string {
	return PluginName
}

// New return npu plugin.
func New(npuName string) plugin.HwNPUSchedulerPlugin {
	var npuPlugin = module910x8{}
	npuPlugin.PluginName = npuName
	npuPlugin.AnnoName = npuPlugin.GetResourceName()
	npuPlugin.AnnoPreVal = npuPlugin.GetResourcePreVal()
	npuPlugin.DefaultJobSchedulerConfig = npuPlugin.GetPluginDefaultJobSchedulerConfig()
	return &npuPlugin
}

// OnHandlerStart The npu scheduler policy initial and common processing.
func (tp *module910x8) OnHandlerStart(sHandler *plugin.ScheduleHandler) {
	klog.V(logErrorLev).Infof("%v start Handler.", PluginName)
	sHandler.AddInitNodesNPUAllocTopology(PluginName, initNodesNPUTopologyFn)
	// Only used by deal fault npu.
	sHandler.AddPreHandleFaultNPU(PluginName, preHandleFaultNPUFn)
	// Pre-select cluster processing.
	sHandler.AddClusterNodePredicateFn(PluginName, clusterNodePredicateFn)
}

// ValidNPUJobFn Check the compliance of the selector and resource request numbers of job.
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

// PreCheckNodeFn Get the nodes that meet the task requirements.
func (tp *module910x8) PreCheckNodeFn(task *vapi.TaskInfo, node *vapi.NodeInfo, confs []conf.Configuration) error {
	if rescheduling.IsNodeInFaultNodeList(node) {
		return fmt.Errorf("PreCheckNodeFn %s in fault node list", node.Name)
	}

	schedulerConf := npuutil.GetSchedulerSelectorConfig(confs)
	if len(schedulerConf) == 0 {
		// get scheduler selector configure failed, but need continue
		klog.V(logErrorLev).Infof("%s JobUID: %s get selector nil.", PluginName, task.Name)
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

// CheckNPUResourceStableFn Check whether the node's NPU resources are stable.
func (tp *module910x8) CheckNPUResourceStableFn(node *vapi.NodeInfo) error {
	return checkNPUResourceStable(node)
}

// CheckNodeNPUByTaskFn Check whether the requested resource exists and are sufficient on the node.
func (tp *module910x8) CheckNodeNPUByTaskFn(task *vapi.TaskInfo, node *vapi.NodeInfo, distributeFlag bool) error {
	taskNPU, taskError := npuutil.GetTaskNPUNum(task, npu800And9000CardName)
	if taskError != nil {
		return fmt.Errorf("getTaskNPUNum %s : %s", nodesNoMeetNPUReqError, taskError)
	}

	nodeNPUTopology := getUsableTopFromNode(node, distributeFlag)
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

	faultJobErr := rescheduling.CheckFaultJobNode(task, node)
	if faultJobErr != nil {
		return fmt.Errorf("checkFaultJobNode %s : %v", nodesNoMeetNPUReqError, faultJobErr)
	}

	return nil
}

// GetNPUAffinityBestNodesFn Initialize a mapping between nodes and priorities.
func (tp *module910x8) GetNPUAffinityBestNodesFn(
	task *vapi.TaskInfo,
	nodes []*vapi.NodeInfo,
	disFlag bool) (map[string]int, error) {
	// 1. init 4 prioritized node-list array.
	priNodeGroups, err := initPriNodeGroups(task, nodes, disFlag)
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
	if len(scoreMap) == 0 || reflect.ValueOf(scoreMap).IsNil() {
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

	tmpScoreMap, err := rescheduling.AddScoreByFaultNPUTask(task, scoreMap)
	if err != nil {
		klog.V(logErrorLev).Infof("%s ScoreBestNPUNodesFn: %v.", PluginName, err)
	}
	return tmpScoreMap, nil
}

// GetAllocatedNPUFromTopologyFn Get the pod's npu card to record in node others.
func (tp *module910x8) GetAllocatedNPUFromTopologyFn(task *vapi.TaskInfo,
	node *vapi.NodeInfo, disFlag bool) (interface{}, error) {
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

	nodeTop := getUsableTopFromNode(node, disFlag)
	if len(nodeTop) == 0 {
		klog.V(logErrorLev).Infof("module910x8 not npu node[%s], no need to continue.", node.Name)
		return allocTopologyHccl, errors.New("failed to get npu topology from node")
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

// SetNPUTopologyToPodFn Set the npu card ids into pod.
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
	tmp := strconv.FormatInt(time.Now().UnixNano(), 10)
	task.Pod.Annotations[podPredicateTime] = tmp
	klog.V(logInfoLev).Infof("%s setNPUTopologyToPod %s==%v top:%s.", PluginName, task.Name, tmp, topologyStr)
	return nil
}

// UpdateNPUNodeUsedCardFn Update used npu resources on node.
func (tp *module910x8) UpdateNPUNodeUsedCardFn(node *vapi.NodeInfo, top interface{}) error {
	useTop, ok := top.([]int)
	if !ok {
		return errors.New(argumentError)
	}

	// get node available top
	nodeDeviceIDs := npuutil.GetTopFromNodeOthers(node, npu800And9000CardName, npu910CardPreName)
	if len(nodeDeviceIDs) == 0 {
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

// GetReleaseNPUTopologyFn Get the release npu card id from task(pod).
func (tp *module910x8) GetReleaseNPUTopologyFn(task *vapi.TaskInfo) (interface{}, error) {
	// get task use top
	taskDevIDs := npuutil.GetDeviceIDsFromAnnotations(task.Pod.Annotations, npu800And9000CardName, npu910CardPreName)
	if taskDevIDs == nil {
		klog.V(logErrorLev).Infof("%s releaseAnnotation failed task:%s.", PluginName, task.Name)
		return nil, fmt.Errorf("%s get npu nil", task.Name)
	}

	return taskDevIDs, nil
}

// UpdateReleaseNPUNodeTopologyFn Update the node using npu when release pod's npu.
func (tp *module910x8) UpdateReleaseNPUNodeTopologyFn(node *vapi.NodeInfo, top interface{}) error {
	taskDeviceIDs, ok := top.([]int)
	if !ok {
		return errors.New(argumentError)
	}

	// get node available top
	nodeDeviceIDs := npuutil.GetTopFromNodeOthers(node, npu800And9000CardName, npu910CardPreName)
	if nodeDeviceIDs == nil {
		klog.V(logErrorLev).Infof("%s useAnnotation node(%s) top nil.", PluginName, node.Name)
		return fmt.Errorf("%s has nil npu", node.Name)
	}
	// delete the use top
	newNodeTopStr := npuutil.GetRealTopAfterRelease(nodeDeviceIDs, taskDeviceIDs, npu910CardPreName)
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

// IsMyTask Determine if it is the NPU task of your plug-in.
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

// IsMyNode Determine if it is the NPU node of your plug-in.
func (tp *module910x8) IsMyNode(node *vapi.NodeInfo) error {
	return isMyNode(node)
}

// IsMyJob Determine if it is the NPU job of your plug-in.
func (tp *module910x8) IsMyJob(job *vapi.JobInfo) error {
	return isMyJob(job)
}

func forceDeleteJobsByPods(ssn *framework.Session, forceJobs []*api.JobInfo) error {
	for _, job := range forceJobs {
		jobPodsTime, jobPodsUID, err := rescheduling.GetRecordJobPods(job)
		if err != nil {
			klog.V(logErrorLev).Infof("GetRecordJobPods: %v.", err)
			continue
		}
		klog.V(logDebugLev).Infof("GetRecordJobPods: %v=%v.", jobPodsTime, jobPodsUID)
		for podName := range jobPodsTime {
			klog.V(logDebugLev).Infof("ForceDeleteFaultPod: %v.", podName)
			deleteErr := plugin.ForceDeleteFaultPod(ssn, job.Namespace, podName, jobPodsUID[podName])
			if deleteErr != nil {
				klog.V(logErrorLev).Infof("ForceDeleteFaultPod %s: %v.", podName, deleteErr)
			}
		}
	}
	return nil
}

func forceDeleteTimeOutGraceTask(ssn *framework.Session) error {
	// 1.Get grace delete jobs from cache.
	dJobs, delayErr := rescheduling.GetGraceDeleteJobsFromCache()
	if delayErr != nil {
		klog.V(logDebugLev).Infof("GetGraceDeleteJobsFromCache %v.", delayErr)
		return delayErr
	}

	// 2.Get force delete jobs.
	klog.V(logDebugLev).Infof("GetCanGraceDeleteJobs: %v.", len(dJobs))
	forceJobs, forceErr := rescheduling.GetNeedForceDeleteDelayingJobs(ssn, dJobs)
	if forceErr != nil {
		return forceErr
	}
	klog.V(logDebugLev).Infof("will force delete: %v.", len(forceJobs))

	// 3.Get force delete jobs.
	if deleteErr := forceDeleteJobsByPods(ssn, forceJobs); deleteErr != nil {
		klog.V(logDebugLev).Infof("ForceDeleteJobsByPods: %v.", deleteErr)
		return deleteErr
	}
	return nil
}

func preHandleFaultNPUFn(ssn *framework.Session) error {
	klog.V(logDebugLev).Infof("%s enter preHandleFaultNPUFn.", PluginName)
	defer klog.V(logDebugLev).Infof("%s leave preHandleFaultNPUFn.", PluginName)

	// 0.init param
	if err := setGraceOverTime(ssn); err != nil {
		klog.V(logErrorLev).Infof("%s setGraceOverTime %v.", PluginName, err)
	}
	// 1.record fault information.
	if err := rescheduling.RecordFaultInfInCache(ssn, nodeNPUNumber); err != nil {
		klog.V(logDebugLev).Infof("%s preHandleFaultNPUFn %v.", PluginName, err)
	}
	// 2.Force delete tasks that gracefully delete cannot be deleted. For job not running
	if err := forceDeleteTimeOutGraceTask(ssn); err != nil {
		klog.V(logInfoLev).Infof("%s ForceDeleteGraceTask %v.", PluginName, err)
	}
	// 3.Determine if it is a 910 jobs.
	jobs, jobGetErr := get910x8RunningJobs(ssn.Jobs)
	if jobGetErr != nil {
		klog.V(logDebugLev).Infof("%s get910x8Jobs: %v.", PluginName, jobGetErr)
		return nil
	}
	// 4.Get fault vcjobs and its node-rankIndex.
	faultNPUJobs, getFaultErr := rescheduling.GetFaultNPUJobs(jobs)
	if getFaultErr != nil {
		klog.V(logDebugLev).Infof("%s getFaultNPUJobs %v.", PluginName, getFaultErr)
		return nil
	}
	// 5.Sets the fault jobs and its index.
	if err := rescheduling.SetFaultInNodeAndJobs(ssn, faultNPUJobs, jobs); err != nil {
		klog.V(logErrorLev).Infof("%s setFaultInNodeAndJobs %v.", PluginName, err)
		return err
	}
	// 6.Restart vcJobs.
	if err := restartFaultJob(ssn, faultNPUJobs, jobs); err != nil {
		klog.V(logErrorLev).Infof("%s restartFaultJob %v.", PluginName, err)
		return err
	}

	return nil
}

func setGraceOverTime(ssn *framework.Session) error {
	if len(ssn.Configurations) == 0 {
		klog.V(logDebugLev).Info("no configurations, GraceOverTime will not be changed.")
		return nil
	}

	configuration, err := npuutil.GetConfigFromSchedulerConfigMap(npuutil.CMInitParamKey, ssn.Configurations)
	if err != nil {
		klog.V(logDebugLev).Info("cannot get configuration, GraceOverTime will not be changed.")
		return err
	}
	// get grace over time by user configuration
	overTimeStr, ok := configuration.Arguments[rescheduling.GraceOverTimeKey]
	if !ok {
		klog.V(logDebugLev).Info("set GraceOverTime failed and will not be changed, " +
			"key grace-over-time doesn't exists.")
		return nil
	}
	const (
		constNumber10    = 10
		constNumber64    = 64
		minGraceOverTime = 2
		maxGraceOverTime = 3600
	)

	overTime, err := strconv.ParseInt(overTimeStr, constNumber10, constNumber64)
	if err != nil {
		klog.V(logDebugLev).Infof("set GraceOverTime failed and will not be changed, "+
			"grace-over-time is invalid [%#v].", overTimeStr)
		return err
	}
	if overTime < minGraceOverTime || overTime > maxGraceOverTime {
		klog.V(logDebugLev).Infof("GraceOverTime value should be range [2, 3600], configured is [%#v], "+
			"GraceOverTime will not be changed", overTimeStr)
		return errors.New("graceOverTime is out of range")
	}
	// use user's configuration to set grace over time
	rescheduling.GraceOverTime = overTime
	klog.V(logDebugLev).Infof("set GraceOverTime to new value [%d].", overTime)

	return nil
}

// GetResourceName get plugin NPU resource name.
func (tp *module910x8) GetResourceName() string {
	return npu800And9000CardName
}

// GetResourcePreVal get plugin NPU resource name prefix.
func (tp *module910x8) GetResourcePreVal() string {
	return npu800And9000CardName
}

// GetPluginDefaultJobSchedulerConfig get plugin default job scheduler config.
func (tp *module910x8) GetPluginDefaultJobSchedulerConfig() map[string]string {
	defaultSchedulerConfig := make(map[string]string, npuutil.NPUIndex1)
	defaultSchedulerConfig[archSelector] = huaweiArchArm + "|" + huaweiArchX86
	return defaultSchedulerConfig
}
