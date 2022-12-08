/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package vnpu is using for HuaWei Ascend pin vnpu allocation.

*/
package vnpu

import (
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// CheckNodeNPUByTask pass for static
func (tp *StaticVNPU) CheckNodeNPUByTask(task *api.TaskInfo, node plugin.NPUNode) error {
	klog.V(util.LogInfoLev).Infof("static vnpu task<%s> node<%s> CheckNodeNPUByTask pass", task.Name, node.Name)
	return nil
}

// ScoreBestNPUNodes pass for static
func (tp *StaticVNPU) ScoreBestNPUNodes(task *api.TaskInfo, nodes []*api.NodeInfo, scoreMap map[string]float64) error {
	klog.V(util.LogInfoLev).Infof("static vnpu task<%s> scoreMap<%#v> ScoreBestNPUNodes pass",
		task.Name, scoreMap)
	return nil
}

// UseAnnotation pass for static
func (tp *StaticVNPU) UseAnnotation(task *api.TaskInfo, node plugin.NPUNode) *plugin.NPUNode {
	klog.V(util.LogInfoLev).Infof("static vnpu task<%s> node<%s> UseAnnotation pass", task.Name, node.Name)
	return &node
}
