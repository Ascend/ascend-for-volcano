/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package main is using for HuaWei Ascend pin affinity schedule.

*/
package main

import (
	"errors"

	"k8s.io/klog"
	"volcano.sh/apis/pkg/apis/scheduling"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	npuapi "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/npuinterface"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/card310x4"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/card910x2"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/chip310x4"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/chip710"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/module910x8"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/vnpu/comvnpu"
)

var sHandler *plugin.ScheduleHandler

// Name This need by volcano frame init plugin.
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

func checkSession(ssn *framework.Session) error {
	if ssn == nil {
		klog.V(logErrorLev).Infof("%s OnSessionOpen got a null session hence doing nothing.", PluginName)
		return errors.New("nil ssn")
	}
	return nil
}

// OnSessionOpen HuaWei NPU Plugin's init session for frame.
func (tp *huaweiNPUPlugin) OnSessionOpen(ssn *framework.Session) {
	klog.V(logInfoLev).Infof("enter %s OnSessionOpen.", PluginName)
	defer klog.V(logInfoLev).Infof("leave %s OnSessionOpen.", PluginName)

	if err := checkSession(ssn); err != nil {
		klog.V(logErrorLev).Infof("%s checkSession : %#v.", PluginName, err)
		return
	}

	// Init npu plugin and nodes.
	sHandler.InitNPUSession(ssn)
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
		if err := sHandler.ClusterNodePredicate(taskInfo, ssn); err != nil {
			klog.V(logDebugLev).Infof("%s clusterNodePredicate : %v.", PluginName, err)
			return err
		}

		return sHandler.NodePredicate(taskInfo, nodeInfo, ssn)
	})
	// The job who has below or equal 8 NPU,only has one pod. If over, every pod has 8s NPU.
	ssn.AddBatchNodeOrderFn(tp.Name(), func(task *api.TaskInfo, nodes []*api.NodeInfo) (map[string]float64, error) {
		score, err := sHandler.BatchNodeOrderFn(task, nodes, plugin.IsDistributeTask(task, ssn))
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
			sHandler.NPUAllocateFunc(event, ssn)
		},
		DeallocateFunc: func(event *framework.Event) {
			sHandler.NPUDeallocateFunc(event, ssn.Nodes)
		},
	})
}

// OnSessionClose Close session by volcano frame.
func (tp *huaweiNPUPlugin) OnSessionClose(ssn *framework.Session) {
	// 1、Record job's unscheduled reason;
	// 2、Update job statue;
	// 3、Handle other post-dispatch issues.
	for _, job := range ssn.Jobs {
		// deal pending job
		if job.PodGroup.Status.Phase == scheduling.PodGroupInqueue ||
			job.PodGroup.Status.Phase == scheduling.PodGroupPending {
			// if all nodes not meet job require failed
			plugin.SetJobPendReasonByNodesCase(ssn, ssn.Nodes, job)
		}
	}

	sHandler.BeforeCloseHandler(ssn)
}

// HandlerStart HuaWei NPU plugin start by frame.
func HandlerStart() *plugin.ScheduleHandler {
	scheduleHandler := &plugin.ScheduleHandler{
		HuaweiNPUs:       map[string]plugin.HwNPUSchedulerPlugin{},
		PluginEntity:     map[string]plugin.HwEntity{},
		PreHandleVNPUFns: map[string]npuapi.PreHandleVNPUFn{},
		VJobRunHandleFns: map[string]npuapi.VNPUJobRunningHandleFn{},
		// for object funcs
		InitNodesNPUAllocTopologyFns: map[string]npuapi.InitNodesNPUTopologyFn{},
		// Handle NPU fault chip functions.
		PreHandleFaultNPUFns: map[string]npuapi.PreHandleFaultNPUFn{},
		// Nodes pre-select cluster processing
		ClusterNodePredicateFns: map[string]npuapi.ClusterNodePredicateFn{},
	}

	// Register new npu scheduler strategy.
	scheduleHandler.RegisterNPUScheduler(card910x2.PluginName, card910x2.New)
	scheduleHandler.RegisterNPUScheduler(module910x8.PluginName, module910x8.New)
	scheduleHandler.RegisterNPUScheduler(card310x4.PluginName, card310x4.New)
	scheduleHandler.RegisterNPUScheduler(chip310x4.PluginName, chip310x4.New)
	scheduleHandler.RegisterNPUScheduler(chip710.PluginName, chip710.New)
	scheduleHandler.RegisterNPUScheduler(comvnpu.PluginName, comvnpu.New)
	// for npu scheduler start.
	for _, huaweiNPU := range scheduleHandler.HuaweiNPUs {
		huaweiNPU.OnHandlerStart(scheduleHandler)
	}

	return scheduleHandler
}
