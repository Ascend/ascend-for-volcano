/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package module910x8 is using for HuaWei Ascend pin affinity schedule.

*/

package module910x8

import (
	"fmt"

	"k8s.io/klog"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

func (tp *module910x8) judgeNodeAndTaskNPU(taskNPU int, nodeTop []int) error {
	var reFlag = false

	sNodeInf := initSelectNodeInf(nodeTop)

	switch taskNPU {
	case 1, npuIndex2, npuNumPerHccs:
		reFlag = (sNodeInf.leftNPUNum >= taskNPU) || (sNodeInf.rightNPUNum >= taskNPU)
	case nodeNPUNumber:
		reFlag = sNodeInf.allNPUNum == nodeNPUNumber
	default:
		// single pod(task) cannot require npu not belong to mode
		// this kind job has been deal with job logical
		klog.V(util.LogErrorLev).Infof("judgeNodeAndTaskNPU err: task req npu is invalid.")
	}

	if reFlag {
		return nil
	}
	meetErr := fmt.Errorf("%v not meet req npu(%d)", nodeTop, taskNPU)
	klog.V(util.LogErrorLev).Infof("cardIDs:<%v> not meet task reqNum<%d>.", nodeTop, taskNPU)
	return meetErr
}

func getNPUAllocPriorityArray(taskNPUNumber int) ([]int, error) {
	var priorityArray []int
	var err error

	switch taskNPUNumber {
	case 1:
		// priority:1>3>2>4
		priorityArray = []int{1, npuIndex3, npuIndex2, npuNumPerHccs}
	case npuIndex2:
		// priority：2>npuNumPerHccs>3
		priorityArray = []int{npuIndex2, npuNumPerHccs, npuIndex3}
	case npuNumPerHccs:
		// priority：4
		priorityArray = []int{npuNumPerHccs}
	case nodeNPUNumber:
		priorityArray = []int{nodeNPUNumber}
	default:
		// For normal,can not be here. The pre function validate job has done this.
		err = fmt.Errorf("illegal request npu number: %d", taskNPUNumber)
	}

	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s %s.", SchedulerName, err.Error())
		return priorityArray, err
	}

	return priorityArray, nil
}
