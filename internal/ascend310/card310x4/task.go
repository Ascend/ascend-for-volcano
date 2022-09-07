package card310x4

import (
	"fmt"

	"k8s.io/klog"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

func getNPUAllocPriorityArray(taskNPUNumber int) ([]int, error) {
	var priorityArray []int
	var err = error(nil)
	switch taskNPUNumber {
	case 0:
		klog.V(util.LogInfoLev).Infof("%s task req npu is 0.", SchedulerName)
	case 1:
		// priority:1>3>2>4
		priorityArray = []int{1, util.NPUIndex3, util.NPUIndex2, maxCardNPUNum}
	case util.NPUIndex2:
		// priority：2>3>4
		priorityArray = []int{util.NPUIndex2, util.NPUIndex3, maxCardNPUNum}
	case util.NPUIndex3:
		// priority：3>4
		priorityArray = []int{util.NPUIndex3, maxCardNPUNum}
	case maxCardNPUNum:
		priorityArray = []int{maxCardNPUNum}
	default:
		// For normal,can not be here. The pre function validate job has done this.
		err = fmt.Errorf("illegal request npu number: %d", taskNPUNumber)
	}
	if err != nil {
		klog.V(util.LogDebugLev).Infof("%s %s.", SchedulerName, err.Error())
		return priorityArray, err
	}
	return priorityArray, nil
}

// JudgeNodeAndTaskNPU judge node topology meet task require
func (tp *card310x4) JudgeNodeAndTaskNPU(taskNPUNum int, nodeTop []int) error {
	cardNumGroups := tp.GetCardNumGroupsFromTop(nodeTop)

	for _, cardNumGroup := range cardNumGroups {
		if len(cardNumGroup) >= taskNPUNum {
			return nil
		}
	}

	var meetErr = fmt.Errorf("req npu(%d) illegal not meet node top<%v>", taskNPUNum, nodeTop)
	klog.V(util.LogErrorLev).Infof("%s judgeNodeAndTaskNPU err: %v.", tp.GetPluginName(), meetErr)
	return meetErr
}
