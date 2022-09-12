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
	"strconv"
	"time"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// GetTaskReqNPUNum get task require npu num
func (tp *NPUHandler) GetTaskReqNPUNum(task *api.TaskInfo) (int, error) {
	if tp == nil || task == nil {
		return 0, errors.New(util.ArgumentError)
	}
	nTask, ok := tp.Tasks[string(task.UID)]
	if !ok {
		err := fmt.Errorf("task<%s> is not npu task", task.Name)
		klog.V(util.LogErrorLev).Infof("GetTaskReqNPUNum err: %s", err.Error())
		return 0, err
	}
	klog.V(util.LogDebugLev).Infof("GetTaskReqNPUNum task req npu<%s>-<%d> ", nTask.ReqNPUName, nTask.ReqNPUNum)
	return nTask.ReqNPUNum, nil
}

// SetNPUTopologyToPodFn set task select npu to pod annotation
func (tp *NPUHandler) SetNPUTopologyToPodFn(task *api.TaskInfo, top []int) {
	if tp == nil || task == nil || len(top) == 0 {
		return
	}
	topologyStr := util.ChangeIntArrToStr(top, tp.GetAnnoPreVal())
	task.Pod.Annotations[tp.GetAnnoName()] = topologyStr
	// to device-plugin judge pending pod.
	tmp := strconv.FormatInt(time.Now().UnixNano(), util.Base10)
	task.Pod.Annotations[util.PodPredicateTime] = tmp
	klog.V(util.LogInfoLev).Infof("%s setNPUTopologyToPod %s==%v top:%s.", tp.GetPluginName(),
		task.Name, tmp, topologyStr)
}
