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

// IsSelectorMeetNode Check whether the selector of the node matches that of the task.
func (tp *VNPU) IsSelectorMeetNode(task *api.TaskInfo, node *api.NodeInfo, conf map[string]string) error {
	// Get node selectors of task
	taskSelectors := util.GetTaskSelectors(task)
	if len(taskSelectors) == 0 {
		for _, v := range tp.Attr.DivideKinds {
			if err := util.IsNPUTask(task, v); err == nil {
				// Vnpu task should have selector
				klog.V(util.LogErrorLev).Infof("isSelectorMeetNode %s no selector by %+v.", task.Name, v)
				return errors.New(util.NodeNoFitSelectorError)
			}
		}
		klog.V(util.LogErrorLev).Infof("isSelectorMeetNode %s no selector .", task.Name)
		return nil
	}

	// Get node selectors of node
	nodeSelector, errNode := util.GetNodeSelector(node)
	if errNode != nil {
		klog.V(util.LogErrorLev).Infof("task[%s] on node(%s) %v.", task.Name, node.Name, errNode)
		return errors.New(util.NodeNoFitSelectorError)
	}

	if err := util.CheckTaskAndNodeSelectorMeet(taskSelectors, nodeSelector, conf); err != nil {
		klog.V(util.LogErrorLev).Infof("isSelectorMeetNode %s not meet %s err:%v.", task.Name, node.Name, err)
		return err
	}

	return nil
}

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
		klog.V(util.LogErrorLev).Infof("%s GetVJobReqNPUType %s %v.", tp.Name(), vTask.Name, getErr)
		return "", getErr
	}

	return tp.GetNPUTypeByResourceName(tmp)
}

func (tp *VNPU) getVNPUPluginNameByReqType(reqNpuType string) string {
	var pluginName string
	switch reqNpuType {
	case vnpuutil.NPU710CardName:
		pluginName = vnpuutil.PluginNameBy710VNPU
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
	var vType string

	topInstance, ok := top.(string)
	if !ok {
		return errors.New("set NPU topology to pod gets invalid argument")
	}

	for _, vt := range tp.Attr.DivideKinds {
		v := strings.TrimPrefix(vt, tp.Attr.AnnoPreVal)
		if strings.HasPrefix(topInstance, v) {
			vType = vt
			break
		}
	}

	klog.V(util.LogInfoLev).Infof("%s setNPUTopologyToPod begin top:%v.", tp.Name(), topInstance)
	task.Pod.Annotations[vType] = topInstance
	tmpTime := strconv.FormatInt(time.Now().UnixNano(), util.Base10)
	task.Pod.Annotations[common.PodPredicateTime] = tmpTime
	klog.V(util.LogInfoLev).Infof("%s setNPUTopologyToPod %s at %v.", tp.Name(), task.Name, tmpTime)

	return nil
}
