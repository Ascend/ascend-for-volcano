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

	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

func (tp *VNPU) GetTaskResource(task *api.TaskInfo, node plugin.NPUNode) (util.VResource, error) {
	coreNum, err := getAiCoreNumFromTask(task)
	if err != nil {
		return util.VResource{}, fmt.Errorf("task %s AscendNPUCore read failed", task.Name)
	}

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
	return tp.VT.Data[virTemplate], nil
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
		virTemplate = "vir01"
	case util.NPUIndex2:
		virTemplate = "vir02"
		if cpuLevel == plugin.AscendVNPULevelLow {
			virTemplate = virTemplate + "_c"
		}
	case util.NPUIndex4:
		switch dvpp {
		case plugin.AscendDVPPEnabledOn:
			virTemplate = "vir04_4c_dvpp"
		case plugin.AscendDVPPEnabledOff:
			virTemplate = "vir04_3c_ndvpp"
		default:
			virTemplate = "vir04"
			if cpuLevel == plugin.AscendVNPULevelLow {
				virTemplate = virTemplate + "_3c"
			}
		}
	}
	return virTemplate
}
