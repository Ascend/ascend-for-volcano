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

Package huaweinpu is using for HuaWei Ascend pin affinity schedule.

*/
package main

import (
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/apis/scheduling"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"
	npuapi "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/npuinterface"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/card910x2"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/cardv910x2"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/module910x8"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/modulev910x8"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/util"
)

var sHandler *plugin.ScheduleHandler

// This need by volcano frame init plugin.
func (tp *huaweiNPUPlugin) Name() string {
	return PluginName
}

// New return npu plugin.
func New(arguments framework.Arguments) framework.Plugin {
	return &huaweiNPUPlugin{pluginArguments: arguments}
}

func init() {
	sHandler = HandlerStart()
}

// HuaWei NPU Plugin's init session for frame.
func (tp *huaweiNPUPlugin) OnSessionOpen(ssn *framework.Session) {
	klog.V(logDebugLev).Infof("enter %s OnSessionOpen.", PluginName)
	defer klog.V(logDebugLev).Infof("leave %s OnSessionOpen.", PluginName)

	if ssn == nil {
		klog.V(logErrorLev).Infof("%s OnSessionOpen got a null session hence doing nothing.", PluginName)
		return
	}

	// Init npu plugin and nodes.
	sHandler.InitNPUSession(ssn)
	// Handle NPU fault chip.
	if err := sHandler.PreHandleFaultNPU(ssn); err != nil {
		klog.V(logDebugLev).Infof("PreHandleFaultNPU :%v .", err)
	}
	// check job npu resource, if illegal return failed
	ssn.AddJobValidFn(tp.Name(), func(obj interface{}) *api.ValidateResult {
		result := sHandler.ValidJobFn(obj, ssn.Configurations)
		if result != nil {
			if setErr := plugin.SetJobPendingReason(ssn, obj, result.Message); setErr != nil {
				klog.V(logErrorLev).Infof("%s setJobFailed err: %v.", PluginName, setErr)
			}
		}
		return result
	})
	// if npu no meet the task require,the task will failed.so need to intercept in advance
	ssn.AddPredicateFn(tp.Name(), func(taskInfo *api.TaskInfo, nodeInfo *api.NodeInfo) error {
		return sHandler.NodePredicate(taskInfo, nodeInfo, ssn.Configurations)
	})
	// The job who has below or equal 8 NPU,only has one pod. If over, every pod has 8s NPU.
	ssn.AddBatchNodeOrderFn(tp.Name(), func(task *api.TaskInfo, nodes []*api.NodeInfo) (map[string]float64, error) {
		score, err := sHandler.BatchNodeOrderFn(task, nodes)
		if err != nil {
			if setErr := plugin.SetJobPendingReason(ssn, ssn.Jobs[task.Job], err.Error()); setErr != nil {
				klog.V(logErrorLev).Infof("%s setJobFailed err:%v.", PluginName, setErr)
			}
		}
		return score, nil
	})
	// Register event handlers to update task info in PodLister & nodeMap
	// for support Concurrency
	ssn.AddEventHandler(&framework.EventHandler{
		AllocateFunc: func(event *framework.Event) {
			sHandler.NPUAllocateFunc(event, ssn.Nodes)
		},
		DeallocateFunc: func(event *framework.Event) {
			sHandler.NPUDeallocateFunc(event, ssn.Nodes)
		},
	})
}

// Close session by volcano frame.
func (tp *huaweiNPUPlugin) OnSessionClose(ssn *framework.Session) {
	// 1、Record job's unscheduled reason;
	// 2、Update job statue;
	// 3、Handle other post-dispatch issues.
	for _, job := range ssn.Jobs {
		// deal pending job
		if job.PodGroup.Status.Phase == scheduling.PodGroupInqueue {
			// if all nodes not meet job require failed
			plugin.SetJobPendReasonByNodesCase(ssn, ssn.Nodes, job)
		}
	}

	if err := util.WriteFaultNPUJobsToCM(ssn, util.ReSchedulerJobs); err != nil {
		klog.V(logInfoLev).Infof("%s writeFaultNPUJobsToCM %v.", PluginName, err)
		return
	}
}

// HandlerStart HuaWei NPU plugin start by frame.
func HandlerStart() *plugin.ScheduleHandler {
	scheduleHandler := &plugin.ScheduleHandler{
		HuaweiNPUs: map[string]plugin.HwNPUSchedulerPlugin{},
		// for object funcs
		InitNodesNPUAllocTopologyFns: map[string]npuapi.InitNodesNPUTopologyFn{},
		// Handle NPU fault chip functions.
		PreHandleFaultNPUFns: map[string]npuapi.PreHandleFaultNPUFn{},
	}

	// registor new npu scheduler strategy.
	scheduleHandler.RegisterNPUScheduler(card910x2.PluginName, card910x2.New)
	scheduleHandler.RegisterNPUScheduler(module910x8.PluginName, module910x8.New)
	scheduleHandler.RegisterNPUScheduler(cardv910x2.PluginName, cardv910x2.New)
	scheduleHandler.RegisterNPUScheduler(modulev910x8.PluginName, modulev910x8.New)
	// for npu scheduler start.
	for _, huaweiNPU := range scheduleHandler.HuaweiNPUs {
		huaweiNPU.OnHandlerStart(scheduleHandler)
	}

	return scheduleHandler
}
