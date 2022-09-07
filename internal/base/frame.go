/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package base is using for HuaWei Ascend pin affinity schedule.

*/

package base

import (
	"errors"
	"fmt"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// InitMyJobPlugin set attr and env for plugin
func (tp *NPUHandler) InitMyJobPlugin(attr util.SchedulerJobAttr, env plugin.ScheduleEnv) error {
	if tp == nil {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("InitMyJobPlugin %s.", err.Error())
		return err
	}
	tp.SetSchedulerAttr(attr)
	tp.SetSchedulerEnv(env)
	return nil
}

// ValidNPUJob check job req npu num
func (tp *NPUHandler) ValidNPUJob() *api.ValidateResult {
	if tp == nil {
		err := errors.New(util.ArgumentError)
		return &api.ValidateResult{
			Pass:    false,
			Reason:  err.Error(),
			Message: err.Error(),
		}
	}
	klog.V(util.LogDebugLev).Infof("%s ValidNPUJob job(%s).", tp.GetPluginName(), tp.JobName)

	for _, task := range tp.Tasks {
		taskNPU := task.ReqNPUNum

		klog.V(util.LogDebugLev).Infof("%s check task<%s> require npu<%d>.",
			tp.GetPluginName(), task.TaskName, taskNPU)

		if taskNPU < 1 || taskNPU > tp.MaxNodeNPUNum {
			err := fmt.Errorf("task<%s-%s> req npu num<%d> is invalid", tp.JobName, task.TaskName, taskNPU)
			klog.V(util.LogErrorLev).Infof("%s ValidNPUJob err: %s", tp.GetPluginName(), err.Error())
			return &api.ValidateResult{
				Pass:    false,
				Reason:  "task req npu num is invalid",
				Message: err.Error(),
			}
		}
	}
	return nil
}

// CheckNodeNPUByTask check nod npu meet task req
func (tp *NPUHandler) CheckNodeNPUByTask(task *api.TaskInfo, node plugin.NPUNode) error {
	if tp == nil || task == nil || len(node.Annotation) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("CheckNodeNPUByTask err: %s.", err.Error())
		return err
	}
	klog.V(util.LogDebugLev).Infof("%s CheckNodeNPUByTask task<%s> node<%s>.",
		tp.GetPluginName(), task.Name, node.Name)
	taskNPUNum, err := tp.GetTaskReqNPUNum(task)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s CheckNodeNPUByTask err: %s", tp.GetPluginName(), err.Error())
		return err
	}

	nodeTop, err := tp.GetUsableTopFromNode(node)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s CheckNodeNPUByTask err: %s", tp.GetPluginName(), err.Error())
		return err
	}

	if err := tp.JudgeNodeAndTaskNPU(taskNPUNum, nodeTop); err != nil {
		klog.V(util.LogErrorLev).Infof("%s CheckNodeNPUByTask err: %s", tp.GetPluginName(), err.Error())
		return err
	}

	return nil
}

// ScoreBestNPUNodes score node by calculate task req npu num and node npu top
func (tp *NPUHandler) ScoreBestNPUNodes(task *api.TaskInfo, nodes []*api.NodeInfo, scoreMap map[string]float64) error {
	if tp == nil || task == nil || len(nodes) == 0 || len(scoreMap) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("ScoreBestNPUNodes err: %s.", err.Error())
		return err
	}
	return nil
}

// UseAnnotation select npu for task from node
func (tp *NPUHandler) UseAnnotation(task *api.TaskInfo, node plugin.NPUNode) *plugin.NPUNode {
	if tp == nil || task == nil || len(node.Annotation) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("UseAnnotation err: %s.", err.Error())
		return nil
	}
	selectedNPU, err := tp.SelectNPUFromNode(task, node)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s UseAnnotation err: %s.", tp.GetPluginName(), err.Error())
		return nil
	}
	klog.V(util.LogInfoLev).Infof("%s UseAnnotation task<%s> select npu <%v>.",
		tp.GetPluginName(), task.Name, selectedNPU)

	tp.SetNPUTopologyToPodFn(task, selectedNPU)
	return tp.UpdateNodeInfo(node, selectedNPU)
}

// PreStartAction do something before schedule
func (tp *NPUHandler) PreStartAction(ssn *framework.Session) error {
	if tp == nil || ssn == nil {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("PreStartAction err: %s.", err.Error())
		return nil
	}
	return nil
}

//PreStopAction do something after schedule
func (tp *NPUHandler) PreStopAction(env *plugin.ScheduleEnv) error {
	if tp == nil || env == nil {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("PreStopAction err: %s.", err.Error())
		return nil
	}
	return nil
}

// SetSchedulerAttr set scheduler attribute for plugin
func (tp *NPUHandler) SetSchedulerAttr(attr util.SchedulerJobAttr) {
	if tp == nil {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("SetSchedulerAttr err: %s.", err.Error())
		return
	}
	tp.SchedulerJobAttr = attr
}

// SetSchedulerEnv set scheduler env for plugin
func (tp *NPUHandler) SetSchedulerEnv(env plugin.ScheduleEnv) {
	if tp == nil {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("SetSchedulerEnv err: %s.", err.Error())
		return
	}
	tp.ScheduleEnv = env
}

// SetMaxNodeNPUNum set max npu num per node
func (tp *NPUHandler) SetMaxNodeNPUNum(num int) {
	if tp == nil || num < 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("SetMaxNodeNPUNum err: %s.", err.Error())
		return
	}
	tp.MaxNodeNPUNum = num
}

// SetMaxCardNPUNum set max npu num per card
func (tp *NPUHandler) SetMaxCardNPUNum(num int) {
	if tp == nil || num < 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("SetMaxCardNPUNum err: %s.", err.Error())
		return
	}
	tp.MaxCardNPUNum = num

}

// JudgeNodeAndTaskNPU judge node and task npu num
func (tp *NPUHandler) JudgeNodeAndTaskNPU(taskNPU int, nodeNPUTopology []int) error {
	if tp == nil {
		return errors.New(util.ArgumentError)
	}
	if taskNPU < 1 || taskNPU > tp.MaxNodeNPUNum {
		return fmt.Errorf("JudgeNodeAndTaskNPU task req num<%d> is invalid", taskNPU)
	}

	if len(nodeNPUTopology) < taskNPU {
		return fmt.Errorf("JudgeNodeAndTaskNPU node don't have enough resource, req<%d>, idle<%d>",
			taskNPU, len(nodeNPUTopology))
	}

	return nil
}

// SelectNPUFromNode select npu from node for task
func (tp *NPUHandler) SelectNPUFromNode(task *api.TaskInfo, node plugin.NPUNode) ([]int, error) {
	if tp == nil || task == nil || len(node.Annotation) == 0 {
		return nil, errors.New(util.ArgumentError)
	}
	taskNPUNum, err := tp.GetTaskReqNPUNum(task)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("selectNPUFromNode err: %s", err.Error())
		return nil, err
	}

	nodeTop, err := tp.GetUsableTopFromNode(node)
	if err != nil {
		return nil, fmt.Errorf("selectNPUFromNode err: %s", err.Error())
	}
	if len(nodeTop) < taskNPUNum {
		return nil, fmt.Errorf("selectNPUFromNode node<%s> npu<%v> not meet task req num<%d>",
			node.Name, nodeTop, taskNPUNum)
	}
	return nodeTop[:taskNPUNum], nil
}
