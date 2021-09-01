/*
Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/*

Package modulev910x8 is using for virtual HuaWei Ascend910 schedule.

*/
package modulev910x8

import (
	"errors"
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	v910 "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/commonv910"
	hwutil "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/util"
)

var mType []string

func init() {
	mType = v910.GetVnpuType()
}

// Name get plugin name of 910-A800 for frame init
func (tp *modulev910x8) Name() string {
	return PluginName
}

// New returns a 910-A800 npu plugin
func New(npuName string) plugin.HwNPUSchedulerPlugin {
	return &modulev910x8{
		name: npuName,
		Vnpu: v910.Vnpu{
			MaxNPUNum: maxNPUNum,
		},
	}
}

// IsMyTask determine whether the task is a 910-A800 task
func (tp *modulev910x8) IsMyTask(task *api.TaskInfo) error {
	var vModuleExist bool

	for _, vType := range mType {
		_, err := hwutil.GetTaskNPUNum(task, vType)
		if err != nil {
			continue
		}
		vModuleExist = true
		break
	}

	if vModuleExist && !hwutil.IsTaskOfCardMode(task) {
		klog.V(logDebugLev).Info("determined as A800 Vnpu Task.")
		return nil
	}

	return errors.New("task doesn't use module type Vnpu")
}

// IsMyNode determine whether the node is a 910-A800 node
func (tp *modulev910x8) IsMyNode(node *api.NodeInfo) error {
	var vModuleExist bool

	for _, vType := range mType {
		topStr, err := hwutil.GetNodeNPUAllocCards(node, vType)
		// IsMyNode is called only in node predict phase, fields of vNPU in Annotation cannot be empty at this phase
		if err != nil || topStr == "" {
			continue
		}
		vModuleExist = true
		break
	}

	if vModuleExist && !hwutil.IsCardModeNode(node) {
		klog.V(logDebugLev).Info("determined as A800 Vnpu Node.")
		return nil
	}

	return errors.New("node doesn't have module type Vnpu")
}

// IsMyJob determine whether the job is a 910-A800 job
func (tp *modulev910x8) IsMyJob(job *api.JobInfo) error {
	var vModuleExist bool

	for _, vType := range mType {
		_, err := hwutil.GetJobReqNPUNum(job, vType)
		if err != nil {
			continue
		}
		vModuleExist = true
		break
	}

	if vModuleExist && !hwutil.IsJobOfCardMode(job) {
		klog.V(logDebugLev).Info("determined as A800 Vnpu Job.")
		return nil
	}

	return errors.New("job doesn't use module type Vnpu")
}
