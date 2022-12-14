/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package ascend310p is using for HuaWei 310P Ascend pin affinity schedule.
*/
package ascend310p

import (
	"errors"
	"fmt"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/rescheduling"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// New return npu plugin.
func New(npuName string) plugin.ISchedulerPlugin {
	npuPlugin := &ascend310P{}
	npuPlugin.SetPluginName(npuName)
	npuPlugin.SetAnnoName(util.NPU310PCardName)
	npuPlugin.SetAnnoPreVal(util.NPU310PCardNamePre)
	npuPlugin.SetDefaultJobSchedulerConfig(nil)
	npuPlugin.SetMaxNodeNPUNum(maxNodeNPUNum)
	npuPlugin.InitVNPU()
	return npuPlugin
}

// PreStartAction pre-processing actions for rescheduling
func (tp *ascend310P) PreStartAction(ssn *framework.Session) error {
	klog.V(util.LogDebugLev).Infof("Entering PreStartAction of %s", util.NPU310PCardName)
	defer klog.V(util.LogDebugLev).Infof("Leaving PreStartAction of %s", util.NPU310PCardName)
	if tp == nil || ssn == nil || tp.FrameAttr.KubeClient == nil {
		return fmt.Errorf("%s handler not enabled or ssn is nil: %s", util.NPU310PCardName, util.ArgumentError)
	}

	reErr := tp.preStartRescheduling(ssn)
	vErr := tp.preStartVNPU(ssn)
	if reErr == nil && vErr == nil {
		return nil
	}

	return fmt.Errorf("%s %s", reErr, vErr)
}

// PreStopAction post-processing actions for re-scheduling
func (tp *ascend310P) PreStopAction(env *plugin.ScheduleEnv) error {
	klog.V(util.LogDebugLev).Infof("enter PreStopAction of %s...", util.NPU310PCardName)
	defer klog.V(util.LogDebugLev).Infof("leave PreStopAction of %s...", util.NPU310PCardName)
	if tp == nil || tp.reHandle == nil || env == nil {
		return fmt.Errorf("%s reSchedule not enabled or nil env: %s", util.NPU310PCardName, util.ArgumentError)
	}
	if err := tp.reHandle.WriteReSchedulerCacheToEnvCache(env, rescheduling.CmFaultJob310PKind); err != nil {
		return err
	}
	return nil
}

// ValidNPUJob check job req npu num and mode
func (tp *ascend310P) ValidNPUJob() *api.ValidateResult {
	var err error
	if tp == nil {
		err := errors.New(util.ArgumentError)
		return &api.ValidateResult{Pass: false, Reason: err.Error(), Message: err.Error()}
	}
	klog.V(util.LogDebugLev).Infof("%s ValidNPUJob job(%s).", tp.GetPluginName(), tp.Name)

	switch tp.Type {
	case util.JobTypeWhole:
		return tp.NPUHandler.ValidNPUJob()
	case util.JobTypeStCut:
		return tp.validStVNPUJob()
	case util.JobTypeDyCut:
		return tp.validDyVNPUJob()
	default:
		err = fmt.Errorf("%s no type %d", tp.Name, tp.Type)
		klog.V(util.LogDebugLev).Infof("%s ValidNPUJob %s %s.", tp.GetPluginName(), tp.Name, err)
	}

	return &api.ValidateResult{Pass: false, Reason: err.Error(), Message: err.Error()}
}

// CheckNodeNPUByTask check nod npu meet task req
func (tp *ascend310P) CheckNodeNPUByTask(task *api.TaskInfo, node plugin.NPUNode) error {
	klog.V(util.LogDebugLev).Infof("%s CheckNodeNPUByTask job(%s).", tp.GetPluginName(), tp.Name)
	var err error
	taskRes, err := tp.vHandle.GetTaskResource(task, node)
	if err != nil {
		return err
	}

	switch tp.Type {
	case util.JobTypeWhole:
		if err = tp.NPUHandler.CheckNodeNPUByTask(task, node); err != nil {
			return err
		}
	case util.JobTypeStCut:
		if err = tp.vHandle.StaticVNPU.CheckNodeNPUByTask(task, node, taskRes); err != nil {
			return err
		}
	case util.JobTypeDyCut:
		if err = tp.vHandle.DynamicVNPU.CheckNodeNPUByTask(task, node, taskRes); err != nil {
			return err
		}
	default:
		err = fmt.Errorf("%s no type %d", tp.Name, tp.Type)
		klog.V(util.LogDebugLev).Infof("%s CheckNodeNPUByTask %s %s.", tp.GetPluginName(), tp.Name, err)
		return err
	}

	if reErr := tp.reHandle.CheckNodeNPUByTask(task, node); reErr != nil {
		return fmt.Errorf("rescheduling CheckNodeNPUByTask %s", reErr.Error())
	}
	return nil
}

// ScoreBestNPUNodes score node by calculate task req npu num and node npu top
func (tp *ascend310P) ScoreBestNPUNodes(task *api.TaskInfo, nodes []*api.NodeInfo, scoreMap map[string]float64) error {
	klog.V(util.LogDebugLev).Infof("%s ScoreBestNPUNodes job(%s).", tp.GetPluginName(), tp.Name)

	switch tp.Type {
	case util.JobTypeWhole:
		if err := tp.NPUHandler.ScoreBestNPUNodes(task, nodes, scoreMap); err != nil {
			return err
		}
	case util.JobTypeStCut:
		if err := tp.vHandle.StaticVNPU.ScoreBestNPUNodes(task, nodes, scoreMap); err != nil {
			return err
		}
	case util.JobTypeDyCut:
		if err := tp.vHandle.DynamicVNPU.ScoreBestNPUNodes(task, nodes, scoreMap); err != nil {
			return err
		}
	default:
		err := fmt.Errorf("%s no type %d", tp.Name, tp.Type)
		klog.V(util.LogDebugLev).Infof("%s ScoreBestNPUNodes %s %s.", tp.GetPluginName(), tp.Name, err)
		return err
	}

	if reErr := tp.reHandle.ScoreBestNPUNodes(task, scoreMap); reErr != nil {
		klog.V(util.LogErrorLev).Infof("%s rescheduling ScoreBestNPUNodes failed :%s.",
			tp.GetPluginName(), reErr.Error())
	}
	klog.V(util.LogInfoLev).Infof("%s ScoreBestNPUNodes task<%s> scoreMap<%v>", tp.GetPluginName(),
		task.Name, scoreMap)
	return nil
}

// UseAnnotation select npu for task from node
func (tp *ascend310P) UseAnnotation(task *api.TaskInfo, node plugin.NPUNode) *plugin.NPUNode {
	klog.V(util.LogDebugLev).Infof("%s UseAnnotation job(%s).", tp.GetPluginName(), tp.Name)
	var err error
	taskRes, err := tp.vHandle.GetTaskResource(task, node)
	klog.V(util.LogDebugLev).Infof("task<%s> require resource<%#v>", task.Name, taskRes)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s UseAnnotation job(%s) get require task resource failed: %s",
			tp.GetPluginName(), tp.Name, err)
	}
	switch tp.Type {
	case util.JobTypeWhole:
		return tp.NPUHandler.UseAnnotation(task, node)
	case util.JobTypeStCut:
		return tp.vHandle.StaticVNPU.UseAnnotation(task, node, taskRes, tp.vHandle.VT)
	case util.JobTypeDyCut:
		return tp.vHandle.DynamicVNPU.UseAnnotation(task, node, taskRes, tp.vHandle.VT)
	default:
		err = fmt.Errorf("%s no type %d", tp.Name, tp.Type)
		klog.V(util.LogDebugLev).Infof("%s CheckNodeNPUByTask %s %s.", tp.GetPluginName(), tp.Name, err)
	}

	return nil
}

// ReleaseAnnotation release select npu for task to node
func (tp *ascend310P) ReleaseAnnotation(task *api.TaskInfo, node plugin.NPUNode) *plugin.NPUNode {
	klog.V(util.LogDebugLev).Infof("%s UseAnnotation job(%s).", tp.GetPluginName(), tp.Name)
	var err error
	switch tp.Type {
	case util.JobTypeWhole, util.JobTypeStCut:
		return &node
	case util.JobTypeDyCut:
		return tp.vHandle.DynamicVNPU.ReleaseAnnotation(task, node)
	default:
		err = fmt.Errorf("%s no type %d", tp.Name, tp.Type)
		klog.V(util.LogDebugLev).Infof("%s CheckNodeNPUByTask %s %s.", tp.GetPluginName(), tp.Name, err)
	}

	return &node
}
