/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package half910x4 is using for HuaWei A800/9000 Ascend910 pin affinity schedule.
*/
package half910x4

import (
	"fmt"

	"k8s.io/klog"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// CheckSingleTrainMode Single Train job has only one task.
func (tp *half910x4) checkSingleTrainMode() error {
	klog.V(util.LogDebugLev).Infof("checkSingleTrainMode job(%s) has %d tasks.", tp.Name, len(tp.Tasks))

	jobNPU := tp.ReqNPUNum
	if jobNPU == 1 || jobNPU == util.NPUIndex2 || jobNPU == npuNumPerHccs {
		return nil
	}
	return fmt.Errorf("%s checkSingleTrainMode job<%s> req npu num is not [1 or 2 or 4]",
		tp.GetPluginName(), tp.Name)
}

// If job requires more than 8 npu, every task need 8 npu.
func (tp *half910x4) checkModuleDistributeTrainMode() error {
	klog.V(util.LogDebugLev).Infof("half DistributeTrainMode %s has %d tasks.", tp.Name, len(tp.Tasks))

	for _, task := range tp.Tasks {
		taskNPU := task.ReqNPUNum
		klog.V(util.LogDebugLev).Infof("checkModuleDistributeTrainMode half DistributeTrain %s has %d npu.",
			task.Name, taskNPU)

		if taskNPU != tp.MaxNodeNPUNum {
			return fmt.Errorf("checkModuleDistributeTrainMode half distributeTrain task<%s> req npu[%d] "+
				"not equal [%d]", task.Name, taskNPU, tp.MaxNodeNPUNum)
		}
	}
	return nil
}

func (tp *half910x4) judgeNodeAndTaskNPU(taskNPU int, nodeTop []int) error {
	if taskNPU <= len(nodeTop) {
		return nil
	}
	meetErr := fmt.Errorf("%v not meet req npu(%d)", nodeTop, taskNPU)
	klog.V(util.LogErrorLev).Infof("cardIDs:<%v> not meet task reqNum<%d>.", nodeTop, taskNPU)
	return meetErr
}
