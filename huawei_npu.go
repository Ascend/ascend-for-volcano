/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package main is using for HuaWei Ascend pin affinity schedule.

*/
package main

import (
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
	klog.V(util.LogInfoLev).Infof("enter %s OnSessionOpen %#v.", PluginName, tp.Scheduler.NPUPlugins)
	defer klog.V(util.LogInfoLev).Infof("leave %s OnSessionOpen.", PluginName)
	if tp == nil || ssn == nil {
		klog.V(util.LogInfoLev).Infof("OnSessionOpen failed: %s.", util.ArgumentError)
		return
	}
	// Init npu plugin and nodes.
	if err := tp.Scheduler.InitNPUSession(ssn); err != nil {
		return
	}
	// check job npu resource, if illegal return failed
	ssn.AddJobValidFn(tp.Name(), func(obj interface{}) *api.ValidateResult {
		return tp.Scheduler.JobValid(obj)
	})
	// if npu no meet the task require,the task will failed.so need to intercept in advance
	ssn.AddPredicateFn(tp.Name(), func(taskInfo *api.TaskInfo, nodeInfo *api.NodeInfo) error {
		return tp.Scheduler.NodePredicate(taskInfo, nodeInfo)
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
	tp.Scheduler.BeforeCloseHandler()
}

// HandlerStart HuaWei NPU plugin start by frame.
func HandlerStart() *plugin.ScheduleHandler {
	scheduleHandler := &plugin.ScheduleHandler{
		NPUPlugins: map[string]plugin.ISchedulerPlugin{},
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
