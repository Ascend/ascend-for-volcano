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

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/vnpu/vnpuutil"
)

// GetReleaseNPUTopologyFn obtain allocated device info from Pod
func (tp *VNPU) GetReleaseNPUTopologyFn(vTask *api.TaskInfo) (interface{}, error) {
	klog.V(util.LogDebugLev).Infof("%s GetReleaseNPUTopologyFn %s, no need.", vnpuutil.PluginName, vTask.Name)
	return nil, nil
}

// IsMyTask used for identify Vnpu task, need to be implemented by vNPU plugins
func (tp *VNPU) IsMyTask(vTask *api.TaskInfo) error {
	klog.V(util.LogDebugLev).Infof("%s IsMyTask %s, no need", tp.Name(), vTask.Name)
	return nil
}

// SetNPUTopologyToPodFn write the name of the allocated devices to Pod
func (tp *VNPU) SetNPUTopologyToPodFn(task *api.TaskInfo, _ interface{}) error {
	klog.V(util.LogDebugLev).Infof("%s SetNPUTopologyToPodFn %s, no need", tp.Name(), task.Name)
	return nil
}
