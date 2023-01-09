/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package vnpu is using for HuaWei Ascend pin vnpu allocation.
*/
package vnpu

import (
	"errors"
	"fmt"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// GetTaskResource get vTask used resource.
func (tp *VirtualNPU) GetTaskResource(task *api.TaskInfo, node plugin.NPUNode) (util.VResource, error) {
	klog.V(util.LogDebugLev).Infof("enter task<%s> GetTaskResource", task.Name)
	coreNum, err := getAiCoreNumFromTask(task)
	if err != nil {
		return util.VResource{}, fmt.Errorf("task %s AscendNPUCore read failed", task.Name)
	}
	klog.V(util.LogDebugLev).Infof("get coreNum %d", coreNum)

	if node.IsResourceWholeCard(coreNum) {
		res := util.VResource{
			Aicore: coreNum,
			Aicpu:  coreNum * node.TotalRes.Aicpu / node.TotalRes.Aicore,
			DVPP:   plugin.AscendDVPPEnabledNull,
		}
		return res, nil
	}

	dvpp, err := tp.getVTaskDVPP(task)
	if err != nil {
		return util.VResource{}, err
	}

	cpuLevel := tp.getVTaskLevel(task)

	virTemplate := getResTemplateFromTaskSetting(coreNum, cpuLevel, dvpp)
	klog.V(util.LogDebugLev).Infof("vnpu template string for cur task:<%s>", virTemplate)
	taskReqRes := tp.VT.Data[virTemplate]
	return taskReqRes, nil
}

func (tp *VirtualNPU) getVTaskDVPP(task *api.TaskInfo) (string, error) {
	dvpp, ok := task.Pod.Labels[plugin.AscendVNPUDVPP]
	if !ok {
		klog.V(util.LogWarningLev).Infof("%s not set VNPU dvpp, use default null.", task.Name)
		return plugin.AscendDVPPEnabledNull, nil
	}
	switch dvpp {
	case plugin.AscendDVPPEnabledOff, plugin.AscendDVPPEnabledNull, plugin.AscendDVPPEnabledOn:
		break
	default:
		klog.V(util.LogWarningLev).Infof("%s set wrong dvpp %s.", task.Name, dvpp)
		return "", fmt.Errorf("err dvpp value:%s", dvpp)
	}
	return dvpp, nil
}

func (tp *VirtualNPU) getVTaskLevel(task *api.TaskInfo) string {
	cpuLevel, ok := task.Pod.Labels[plugin.AscendVNPULevel]
	if !ok {
		klog.V(util.LogWarningLev).Infof("%s not set VNPU level, use default low.", task.Name)
		return plugin.AscendVNPULevelLow
	}
	switch cpuLevel {
	case plugin.AscendVNPULevelLow, plugin.AscendVNPULevelHigh:
		break
	default:
		klog.V(util.LogWarningLev).Infof("%s set wrong VNPU level %s, use default low.", task.Name, cpuLevel)
		cpuLevel = plugin.AscendVNPULevelLow
	}
	return cpuLevel
}

func getAiCoreNumFromTask(task *api.TaskInfo) (int, error) {
	for _, container := range task.Pod.Spec.Containers {
		coreNum, ok := container.Resources.Requests[util.AscendNPUCore]
		if !ok {
			return 0, errors.New("getAiCoreNumFromTask get resource requests failed")
		}
		return int(coreNum.Value()), nil
	}
	return 0, fmt.Errorf("getAiCoreNumFromTask get resource requests failed")
}

// getResTemplateFromTaskSetting get like vir04_3c_ndvpp
func getResTemplateFromTaskSetting(coreNum int, cpuLevel, dvpp string) string {
	var virTemplate string
	switch coreNum {
	case util.NPUIndex1:
		virTemplate = plugin.VNPUTempVir01
	case util.NPUIndex2:
		virTemplate = plugin.VNPUTempVir02
		if cpuLevel == plugin.AscendVNPULevelLow {
			virTemplate = virTemplate + "_1c"
		}
	case util.NPUIndex4:
		switch dvpp {
		case plugin.AscendDVPPEnabledOn:
			virTemplate = plugin.VNPUTempVir04C4cDVPP
		case plugin.AscendDVPPEnabledOff:
			virTemplate = plugin.VNPUTempVir04C3NDVPP
		default:
			virTemplate = plugin.VNPUTempVir04
			if cpuLevel == plugin.AscendVNPULevelLow {
				virTemplate = virTemplate + "_3c"
			}
		}
	default:
		klog.V(util.LogErrorLev).Infof("wrong number %d", coreNum)
		return ""
	}
	return virTemplate
}
