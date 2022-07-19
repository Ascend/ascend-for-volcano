/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package comvnpu is using for virtual HuaWei Ascend910 schedule.

*/
package comvnpu

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/common"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/vnpu/vnpuutil"
)

// GetReleaseNPUTopologyFn obtain allocated device info from Pod
func (tp *VNPU) GetReleaseNPUTopologyFn(vTask *api.TaskInfo) (interface{}, error) {
	klog.V(util.LogDebugLev).Infof("%s GetReleaseNPUTopologyFn %s, no need.", vnpuutil.PluginName, vTask.Name)
	return nil, nil
}

// GetVTaskReqNPUType get task require npu type.
func (tp *VNPU) GetVTaskReqNPUType(vTask *api.TaskInfo) (string, error) {
	tmp, getErr := util.GetReqResourceNameFromTask(vTask)
	if getErr != nil {
		klog.V(util.LogErrorLev).Infof("%s GetVJobReqNPUType %v.", tp.Name(), getErr)
		return "", getErr
	}

	return tp.GetNPUTypeByResourceName(tmp)
}

func (tp *VNPU) getVNPUPluginNameByReqType(reqNpuType string) string {
	var pluginName string
	switch reqNpuType {
	case vnpuutil.NPU310PCardName:
		pluginName = vnpuutil.PluginNameBy310PVNPU
	case vnpuutil.NPU910CardName:
		pluginName = vnpuutil.PluginNameBy910VNPU
	default:
		pluginName = vnpuutil.UNKnownPluginName
	}
	return pluginName
}

// GetPluginNameByTaskInfo Get plugin name by taskInfo
func (tp *VNPU) GetPluginNameByTaskInfo(vTask *api.TaskInfo) (string, error) {
	reqNpuType, typeErr := tp.GetVTaskReqNPUType(vTask)
	if typeErr != nil {
		klog.V(util.LogErrorLev).Infof("%s GetPluginNameByJobInfo %s %v.", tp.Name(), vTask.Name, typeErr)
		return "", typeErr
	}

	return tp.getVNPUPluginNameByReqType(reqNpuType), nil
}

// IsMyTask used for identify Vnpu task, need to be implemented by vNPU plugins
func (tp *VNPU) IsMyTask(vTask *api.TaskInfo) error {
	klog.V(util.LogDebugLev).Infof("%s IsMyTask %s.", tp.Name(), vTask.Name)
	return nil
}

// SetNPUTopologyToPodFn write the name of the allocated devices to Pod
func (tp *VNPU) SetNPUTopologyToPodFn(task *api.TaskInfo, top interface{}) error {
	if task == nil {
		return errors.New("nil parameters")
	}
	topInstance, ok := top.(string)
	if !ok {
		return errors.New("set NPU topology to pod gets invalid argument")
	}

	var vType string
	for _, vt := range tp.Attr.DivideKinds {
		v := strings.TrimPrefix(vt, tp.Attr.AnnoPreVal)
		if strings.HasPrefix(topInstance, v) {
			vType = vt
			break
		}
	}
	if vType == "" {
		return fmt.Errorf("%s no requets %v", task.Name, top)
	}

	klog.V(util.LogInfoLev).Infof("%s setNPUTopologyToPod begin top:%v.", tp.Name(), topInstance)
	task.Pod.Annotations[vType] = topInstance
	tmpTime := strconv.FormatInt(time.Now().UnixNano(), util.Base10)
	task.Pod.Annotations[common.PodPredicateTime] = tmpTime
	klog.V(util.LogInfoLev).Infof("%s setNPUTopologyToPod %s at %v.", tp.Name(), task.Name, tmpTime)

	return nil
}
