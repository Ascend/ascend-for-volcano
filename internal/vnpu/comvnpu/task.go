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
	var vType string
	var taskTopArr []string
	var err error
	// 1.init vnp
	pluginName, nameErr := tp.GetPluginNameByTaskInfo(vTask)
	if nameErr != nil {
		klog.V(util.LogErrorLev).Infof("%s GetPluginNameByJobInfo %s %v.", tp.Name(), vTask.Name, nameErr)
		return nil, nameErr
	}
	if pluginErr := tp.InitVNPUPluginByType(pluginName); pluginErr != nil {
		klog.V(util.LogErrorLev).Infof("%s InitVNPUPluginByType :%v.", vnpuutil.PluginName, pluginErr)
		return nil, pluginErr
	}
	for _, vType = range tp.Attr.DivideKinds {
		taskTopArr, err = tp.GetNPUsFromNodeAnnotation(vTask.Pod.Annotations, vType)
		if err != nil {
			continue
		}
		break
	}

	if len(taskTopArr) == 0 {
		klog.V(util.LogErrorLev).Infof("%s getReleaseVnpuTopology pod annotation for %s is empty.", tp.Name(), vType)
		return taskTopArr, errors.New("task pod annotation is empty")
	}

	return taskTopArr, nil
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
	if vTask == nil {
		return errors.New("nil task")
	}
	// 1.init vnp
	pluginName, nameErr := tp.GetPluginNameByTaskInfo(vTask)
	if nameErr != nil {
		klog.V(util.LogErrorLev).Infof("%s GetPluginNameByJobInfo %s %v.", tp.Name(), vTask.Name, nameErr)
		return nameErr
	}
	if pluginErr := tp.InitVNPUPluginByType(pluginName); pluginErr != nil {
		klog.V(util.LogErrorLev).Infof("%s InitVNPUPluginByType :%v.", vnpuutil.PluginName, pluginErr)
		return pluginErr
	}
	// 2.vnp job.
	reqNpuType, getErr := util.GetReqResourceNameFromTask(vTask)
	if getErr != nil {
		klog.V(util.LogErrorLev).Infof("%s GetVJobReqNPUType %s %v.", tp.Name(), vTask.Name, getErr)
		return getErr
	}

	for _, kind := range tp.Attr.DivideKinds {
		if kind == reqNpuType {
			klog.V(util.LogErrorLev).Infof("%s IsMyTask %s is %+v kind.", tp.Name(), vTask.Name, kind)
			tp.setVPUPluginToVNPUBack()
			return nil
		}
	}
	kindErr := fmt.Errorf("%s: %s not in %+v", vTask.Name, reqNpuType, tp.Attr)
	klog.V(util.LogErrorLev).Infof("%s IsVNPUJob %+v.", tp.Name(), kindErr)
	return kindErr
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
