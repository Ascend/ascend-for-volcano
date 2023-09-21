/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package main is using for HuaWei Ascend pin affinity schedule.
*/
package main

import (
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/klog"
	"volcano.sh/apis/pkg/apis/scheduling"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/ascend310"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/ascend310p"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/ascend910"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

var sHandler *plugin.ScheduleHandler

func init() {
	sHandler = HandlerStart()
}

// Name This need by volcano frame init plugin.
func (tp *huaweiNPUPlugin) Name() string {
	return PluginName
}

// New return npu plugin.
func New(arguments framework.Arguments) framework.Plugin {
	return &huaweiNPUPlugin{Scheduler: sHandler, Arguments: arguments}
}

// OnSessionOpen HuaWei NPU Action's init session for frame.
func (tp *huaweiNPUPlugin) OnSessionOpen(ssn *framework.Session) {
	klog.V(util.LogInfoLev).Infof("enter %s OnSessionOpen.", PluginName)
	defer klog.V(util.LogInfoLev).Infof("leave %s OnSessionOpen.", PluginName)
	if tp == nil || ssn == nil {
		klog.V(util.LogInfoLev).Infof("OnSessionOpen : %s.", util.ArgumentError)
		return
	}
	// Init npu plugin and nodes.
	if err := tp.Scheduler.InitNPUSession(ssn); err != nil {
		klog.V(util.LogErrorLev).Infof("InitNPUSession : %s, npu plugin will not be initialized.", err)
		return
	}
	// check job npu resource, if illegal return failed
	ssn.AddJobValidFn(tp.Name(), func(obj interface{}) *api.ValidateResult {
		return tp.Scheduler.JobValid(obj)
	})
	// if node not meet the task require, the task will be failed. so need to intercept in advance
	ssn.AddPredicateFn(tp.Name(), func(taskInfo *api.TaskInfo, nodeInfo *api.NodeInfo) error {
		err := tp.Scheduler.NodePredicate(taskInfo, nodeInfo)
		if err != nil {
			klog.V(util.LogErrorLev).Infof("NodePredicate failed for task %s err:%s", taskInfo.Name, err)
		}
		return err
	})

	ssn.AddBatchNodeOrderFn(tp.Name(), func(task *api.TaskInfo, nodes []*api.NodeInfo) (map[string]float64, error) {
		score, err := tp.Scheduler.BatchNodeOrderFn(task, nodes)
		if err != nil {
			if setErr := tp.Scheduler.SetJobPendingReason(ssn.Jobs[task.Job], err.Error()); setErr != nil {
				klog.V(util.LogErrorLev).Infof("%s setJobFailed err:%#v.", PluginName, setErr)
			}
		}
		return score, nil
	})

	ssn.AddJobReadyFn(tp.Name(), func(obj interface{}) bool {
		ji := obj.(*api.JobInfo)
		k, ok := ji.PodGroup.Labels[plugin.TorAffinityKey]
		if !ok || k == plugin.NullTag {
			return true
		}
		if tp.Scheduler.Tors == nil {
			return false
		}
		job, ok := tp.Scheduler.Jobs[ji.UID]
		if !ok {
			return true
		}
		return job.JobReadyTag
	})
	ssn.AddJobEnqueueableFn(tp.Name(), func(job interface{}) int {
		if tp.Scheduler.NPUPlugins == nil {
			klog.V(util.LogErrorLev).Infof("AddJobEnqueueableFn : %s", util.ArgumentError)
			return util.JobEnqueueSkip
		}
		vcjob, ok := job.(*api.JobInfo)
		if !ok {
			return util.JobEnqueueSkip
		}
		npuName, rNpuNum, _ := plugin.GetVCJobReqNPUTypeFromJobInfo(vcjob)
		if _, ok := tp.Scheduler.NPUPlugins[npuName]; !ok {
			return util.JobEnqueueSkip
		}
		var tNpuNum int
		for _, node := range ssn.Nodes {
			vcNode, ok := tp.Scheduler.Nodes[node.Name]
			if !ok {
				klog.V(util.LogErrorLev).Infof("AddJobEnqueueableFn add node failed,%s is not in cache", node.Name)
				continue
			}
			deviceInfo, ok := vcNode.Annotation[npuName]
			if !ok {
				klog.V(util.LogErrorLev).Infof("AddJobEnqueueableFn add node failed,"+
					"%s deviceList is empty", node.Name)
				continue
			}
			deviceList := strings.Split(deviceInfo, ",")
			klog.V(util.LogInfoLev).Infof("Add enqueue node %s deviceList is: %#v", vcNode.Name, deviceList)
			npuNum, ok := vcNode.Idle[v1.ResourceName(npuName)]
			if !ok || len(deviceList) != int(npuNum/util.NPUHexKilo) {
				continue
			}
			tNpuNum += len(deviceList)
		}
		if tNpuNum < rNpuNum {
			klog.V(util.LogWarningLev).Infof("Add enqueue failed, require npu num is %v "+
				"but cluster npu num is %v", rNpuNum, tNpuNum)
			return util.JobNotEnqueue
		}
		return util.JobEnqueue
	})
	// Register event handlers to update task info in PodLister & nodeMap
	// for support Concurrency
	ssn.AddEventHandler(&framework.EventHandler{
		AllocateFunc: func(event *framework.Event) {
			if event == nil {
				klog.V(util.LogErrorLev).Infof("AllocateFunc event nil.")
				return
			}
			tp.Scheduler.NPUAllocateFunc(event.Task)
		},
		DeallocateFunc: func(event *framework.Event) {
			if event == nil {
				klog.V(util.LogErrorLev).Infof("DeallocateFunc event nil.")
				return
			}
			tp.Scheduler.NPUDeallocateFunc(event.Task)
		},
	})
}

// OnSessionClose Close session by volcano frame.
func (tp *huaweiNPUPlugin) OnSessionClose(ssn *framework.Session) {
	klog.V(util.LogInfoLev).Infof("enter %s OnSessionClose.", PluginName)
	defer klog.V(util.LogInfoLev).Infof("leave %s OnSessionClose.", PluginName)
	if tp == nil || ssn == nil {
		klog.V(util.LogInfoLev).Infof("OnSessionClose failed: %s.", util.ArgumentError)
		return
	}
	// 1、Record job's unscheduled reason;
	// 2、Update job statue;
	// 3、Handle other post-dispatch issues.
	for _, job := range ssn.Jobs {
		// deal pending job
		if job.PodGroup.Status.Phase == scheduling.PodGroupInqueue ||
			job.PodGroup.Status.Phase == scheduling.PodGroupPending {
			// if all nodes not meet job require failed
			tp.Scheduler.SetJobPendReasonByNodesCase(job)
		}
	}
	tp.Scheduler.BeforeCloseHandler(ssn)
}

// HandlerStart HuaWei NPU plugin start by frame.
func HandlerStart() *plugin.ScheduleHandler {
	scheduleHandler := &plugin.ScheduleHandler{
		NPUPlugins: map[string]plugin.NPUBuilder{},
		ScheduleEnv: plugin.ScheduleEnv{
			Jobs:      map[api.JobID]plugin.SchedulerJob{},
			Nodes:     map[string]plugin.NPUNode{},
			FrameAttr: plugin.VolcanoFrame{},
		},
	}

	// Register new npu scheduler strategy.
	scheduleHandler.RegisterNPUScheduler(ascend310.PluginName, ascend310.New)
	scheduleHandler.RegisterNPUScheduler(ascend310p.PluginName, ascend310p.New)
	scheduleHandler.RegisterNPUScheduler(ascend910.PluginName, ascend910.New)
	klog.V(util.LogInfoLev).Infof("HandlerStart %#v.", scheduleHandler.NPUPlugins)
	return scheduleHandler
}
