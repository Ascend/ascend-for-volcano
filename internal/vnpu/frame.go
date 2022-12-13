/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
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

func (tp *VNPU) GetTaskResource(task *api.TaskInfo, node plugin.NPUNode) (util.VResource, error) {
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

	cpuLevel, ok := task.Pod.Labels[plugin.AscendVNPULevel]
	if !ok {
		cpuLevel = plugin.AscendVNPULevelLow
	}

	dvpp, ok := task.Pod.Labels[plugin.AscendVNPUDVPP]
	if !ok {
		dvpp = plugin.AscendDVPPEnabledNull
	}

	virTemplate := getResTemplateFromTaskSetting(coreNum, cpuLevel, dvpp)
	klog.V(util.LogDebugLev).Infof("vnpu template string for cur task:<%s>", virTemplate)
	taskReqRes := tp.VT.Data[virTemplate]
	return taskReqRes, nil
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
			virTemplate = virTemplate + "_c"
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
	}
	return virTemplate
}
