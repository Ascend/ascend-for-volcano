/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package comvnpu is using for virtual HuaWei Ascend910 schedule.

*/
package comvnpu

import (
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/conf"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/vnpu/vnpuutil"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

// Name Get plugin name for frame init
func (tp *VNPU) Name() string {
	if tp == nil {
		return vnpuutil.PluginName
	}
	return tp.Attr.PluginName
}

// New returns a virtual npu plugin
func New(npuName string) plugin.HwNPUSchedulerPlugin {
	var npuPlugin = VNPU{}
	npuPlugin.Attr.PluginName = npuName
	return &npuPlugin
}

// OnHandlerStart Vnpu scheduler policy initial and common processing
func (tp *VNPU) OnHandlerStart(sHandler *plugin.ScheduleHandler) {
	klog.V(util.LogInfoLev).Infof("%v start Handler.", tp.Name())
	sHandler.AddInitNodesNPUAllocTopology(tp.Name(), tp.InitVNodesFn)
	sHandler.AddPreHandleVNPU(tp.Name(), tp.PreHandleVNPU)
}

// PreHandleVNPU Only for abstract VNPU, not v910,v310P and so on.
func (tp *VNPU) PreHandleVNPU(ssn *framework.Session) error {
	err := vnpuutil.CheckVNPUSegmentEnable(ssn)
	klog.V(util.LogDebugLev).Infof("PreHandleVNPU :%v.", err)
	return err
}

// PreCheckNodeFn check whether the node matches the tag requirements of the task.
func (tp *VNPU) PreCheckNodeFn(task *api.TaskInfo, _ *api.NodeInfo, confs []conf.Configuration) error {
	klog.V(util.LogDebugLev).Infof("%s PreCheckNodeFn %s enter.", tp.Name(), task.Name)
	defer klog.V(util.LogDebugLev).Infof("%s PreCheckNodeFn %s leave.", tp.Name(), task.Name)

	err := vnpuutil.CheckVNPUSegmentEnableByConfig(confs)
	klog.V(util.LogErrorLev).Infof("%s PreCheckNodeFn %v.", vnpuutil.PluginName, err)
	return err
}

// CheckNodeNPUByTaskFn check whether the requested resource exists on the node.The cored has been split.
func (tp *VNPU) CheckNodeNPUByTaskFn(vTask *api.TaskInfo, node *api.NodeInfo, _ bool) error {
	// has been done in pre-check.
	klog.V(util.LogDebugLev).Infof("%s CheckNodeNPUByTaskFn %s for %s, no need.", tp.Name(), vTask.Name, node.Name)
	return nil
}

// GetNPUAffinityBestNodesFn initialize a mapping between nodes and priorities
func (tp *VNPU) GetNPUAffinityBestNodesFn(task *api.TaskInfo, _ []*api.NodeInfo, _ bool) (map[string]int, error) {
	klog.V(util.LogDebugLev).Infof("%s GetNPUAffinityBestNodesFn %s, no need.", tp.Name(), task.Name)
	return nil, nil
}

// ScoreBestNPUNodesFn used for score candidate nodes
func (tp *VNPU) ScoreBestNPUNodesFn(_ map[string]float64, _ map[string]int, vTask *api.TaskInfo,
	_ []*api.NodeInfo) (map[string]float64, error) {
	klog.V(util.LogDebugLev).Infof("%s ScoreBestNPUNodesFn %s, no need.", tp.Name(), vTask.Name)
	return nil, nil
}

// GetAllocatedNPUFromTopologyFn obtain the name of the allocated devices, VNPU only has one chip.
func (tp *VNPU) GetAllocatedNPUFromTopologyFn(vTask *api.TaskInfo, node *api.NodeInfo, _ bool) (interface{}, error) {
	klog.V(util.LogDebugLev).Infof("%s GetAllocatedNPUFromTopologyFn %s for %s, no need.",
		tp.Name(), node.Name, vTask.Name)
	return nil, nil
}

func (tp *VNPU) setVPUPluginToVNPUBack() {
	// set vnpu plugin to vnp back.
	tp.Attr.PluginName = PluginName
}
