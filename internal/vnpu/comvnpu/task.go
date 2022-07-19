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
