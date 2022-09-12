/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package ascend910 is using for HuaWei Ascend pin affinity schedule.

*/
package ascend910

import (
	"errors"
	"fmt"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/ascend910/card910x2"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/ascend910/module910x8"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/base"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// Name This need by frame init plugin.
func (tp *ascend910) Name() string {
	return PluginName
}

// New return npu plugin.
func New(npuName string) plugin.ISchedulerPlugin {
	var npuPlugin = &ascend910{}
	npuPlugin.SetPluginName(npuName)
	npuPlugin.SetAnnoName(util.NPU910CardName)
	npuPlugin.SetAnnoPreVal(util.NPU910CardNamePre)
	npuPlugin.SetDefaultJobSchedulerConfig(nil)

	npuPlugin.Kind = map[string]base.AscendHandler{}
	npuPlugin.Kind[card910x2.SchedulerName] = card910x2.New(card910x2.SchedulerName)
	npuPlugin.Kind[module910x8.SchedulerName] = module910x8.New(module910x8.SchedulerName)
	return npuPlugin
}

// InitMyJobPlugin init job handle plugin
func (tp *ascend910) InitMyJobPlugin(attr util.SchedulerJobAttr, env plugin.ScheduleEnv) error {
	if tp == nil {
		err := fmt.Errorf("nil plugin %s", PluginName)
		klog.V(util.LogErrorLev).Infof("InitMyJobPlugin err: %s.", err.Error())
		return err
	}
	tp.SetSchedulerAttr(attr)
	tp.SetSchedulerEnv(env)
	v, ok := attr.Selector[Accelerator910Key]
	if !ok {
		v = Module910AcceleratorValue
	}
	value, ok := tp.Kind[attr.ReqNPUName+v]
	if !ok {
		err := fmt.Errorf("not support %s", attr.ReqNPUName+v)
		klog.V(util.LogErrorLev).Infof("%s InitMyJobPlugin err: %s", tp.GetPluginName(), err.Error())
		return err
	}
	if err := value.InitMyJobPlugin(attr, env); err != nil {
		klog.V(util.LogErrorLev).Infof("%s InitMyJobPlugin err: %s", tp.GetPluginName(), err.Error())
		return err
	}

	tp.handle = value

	return nil
}

// ValidNPUJob check job req npu num and mode
func (tp *ascend910) ValidNPUJob() *api.ValidateResult {
	if tp == nil {
		err := fmt.Errorf("nil plugin %s", PluginName)
		klog.V(util.LogErrorLev).Infof("ValidNPUJob err: %s.", err.Error())
		return &api.ValidateResult{
			Pass:    false,
			Reason:  err.Error(),
			Message: err.Error(),
		}
	}
	if tp.handle != nil {
		return tp.handle.ValidNPUJob()
	}
	klog.V(util.LogDebugLev).Infof("%s ValidNPUJob handle is nil", tp.GetPluginName())
	return nil
}

// CheckNodeNPUByTask check node npu meet task request
func (tp *ascend910) CheckNodeNPUByTask(task *api.TaskInfo, node plugin.NPUNode) error {
	if tp == nil || task == nil || len(node.Annotation) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("CheckNodeNPUByTask err: %s", err.Error())
		return err
	}
	if tp.handle != nil {
		return tp.handle.CheckNodeNPUByTask(task, node)
	}
	klog.V(util.LogDebugLev).Infof("%s CheckNodeNPUByTask handle is nil", tp.GetPluginName())
	return nil
}

// ScoreBestNPUNodes score nodes which meet task req
func (tp *ascend910) ScoreBestNPUNodes(task *api.TaskInfo, nodes []*api.NodeInfo, scoreMap map[string]float64) error {
	if tp == nil || task == nil || len(nodes) == 0 || len(scoreMap) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("ScoreBestNPUNodes %v.", err.Error())
		return err
	}
	if tp.handle != nil {
		return tp.handle.ScoreBestNPUNodes(task, nodes, scoreMap)
	}
	klog.V(util.LogDebugLev).Infof("%s ScoreBestNPUNodes handle is nil", tp.GetPluginName())
	return nil
}

// UseAnnotation select npu for task from node
func (tp *ascend910) UseAnnotation(task *api.TaskInfo, node plugin.NPUNode) *plugin.NPUNode {
	if tp == nil || task == nil || len(node.Annotation) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("UseAnnotation %s.", err.Error())
		return nil
	}
	if tp.handle != nil {
		return tp.handle.UseAnnotation(task, node)
	}
	klog.V(util.LogDebugLev).Infof("%s UseAnnotation handle is nil", tp.GetPluginName())
	return nil
}

// PreStartAction pre-processing actions for rescheduling
func (tp *ascend910) PreStartAction(ssn *framework.Session) error {
	if tp == nil || tp.handle == nil {
		return fmt.Errorf(util.ArgumentError)
	}
	if ssn == nil {
		return fmt.Errorf("session is nil: %s", util.ArgumentError)
	}
	klog.V(util.LogInfoLev).Infof("Enter PreStartAction for %s", tp.GetPluginName())
	defer klog.V(util.LogInfoLev).Infof("Leave PreStartAction for %s", tp.GetPluginName())
	for name, handler := range tp.Kind {
		klog.V(util.LogInfoLev).Infof("preStartAction for %s", name)
		if err := handler.PreStartAction(ssn); err != nil {
			klog.V(util.LogErrorLev).Infof("preStartAction %s error: %v", name, err)
		}
	}
	return nil
}

// PreStopAction post-processing actions for re-scheduling
func (tp *ascend910) PreStopAction(env *plugin.ScheduleEnv) error {
	if tp == nil || tp.handle == nil {
		return fmt.Errorf(util.ArgumentError)
	}
	if env == nil {
		return fmt.Errorf("env is nil: %s", util.ArgumentError)
	}
	for name, handler := range tp.Kind {
		klog.V(util.LogInfoLev).Infof("preStopAction for %s", name)
		if err := handler.PreStopAction(env); err != nil {
			klog.V(util.LogErrorLev).Infof("preStopAction %s error: %v", name, err)
		}
	}
	return nil
}
