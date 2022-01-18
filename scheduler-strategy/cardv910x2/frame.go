/*
Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package cardv910x2 is using for virtual HuaWei A300T schedule.

*/
package cardv910x2

import (
	"errors"
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	v910 "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/commonv910"
	hwutil "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/util"
)

var cType []string

func init() {
	cType = v910.GetVnpuType()
}

// Name Get plugin name of 910-300T for frame init
func (tp *cardv910x2) Name() string {
	return PluginName
}

// New returns a 910-300T npu plugin
func New(npuName string) plugin.HwNPUSchedulerPlugin {
	return &cardv910x2{
		name: npuName,
		Vnpu: v910.Vnpu{
			MaxNPUNum: maxNPUNum,
		},
	}
}

// GetNpuJobDefaultSelectorConfig get selector configs of 910-300T
func (tp *cardv910x2) GetNpuJobDefaultSelectorConfig() map[string]string {
	var defaultSchedulerConfig map[string]string
	defaultSchedulerConfig = make(map[string]string, mapInitLen)

	defaultSchedulerConfig[archSelector] = huaweiArchArm + "|" + huaweiArchX86
	defaultSchedulerConfig[acceleratorType] = cardAcceleratorType + "|" + moduleAcceleratorType

	return defaultSchedulerConfig
}

// IsMyTask determine whether the task is a 910-300T task
func (tp *cardv910x2) IsMyTask(task *api.TaskInfo) error {
	var vCardExist bool

	for _, vType := range cType {
		_, err := hwutil.GetTaskNPUNum(task, vType)
		if err != nil {
			continue
		}
		vCardExist = true
		break
	}

	if vCardExist && hwutil.IsTaskOfCardMode(task) {
		klog.V(logDebugLev).Info("determined as 300T Vnpu Task.")
		return nil
	}

	return errors.New("task doesn't use card type Vnpu")
}

// IsMyNode determine whether the node is a 910-300T node
func (tp *cardv910x2) IsMyNode(node *api.NodeInfo) error {
	var vCardExist bool

	for _, vType := range cType {
		topStr, err := hwutil.GetNPUAllocCardsFromNodeOthers(node, vType)
		// IsMyNode is called only in node predict phase, fields of vNPU in Annotation cannot be empty at this phase
		if err != nil || topStr == "" {
			continue
		}
		vCardExist = true
		break
	}

	if vCardExist && hwutil.IsCardModeNode(node) {
		klog.V(logDebugLev).Info("determined as 300T Vnpu Node.")
		return nil
	}

	return errors.New("node doesn't have card type Vnpu")
}

// IsMyJob determine whether the job is a 910-300T job
func (tp *cardv910x2) IsMyJob(job *api.JobInfo) error {
	var vCardExist bool

	for _, vType := range cType {
		_, err := hwutil.GetJobReqNPUNum(job, vType)
		if err != nil {
			continue
		}
		vCardExist = true
		break
	}

	if vCardExist && hwutil.IsJobOfCardMode(job) {
		klog.V(logDebugLev).Info("determined as 300T Vnpu Job.")
		return nil
	}

	return errors.New("job doesn't use card type Vnpu")
}
