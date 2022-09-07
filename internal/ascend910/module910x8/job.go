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

// CheckSingleTrainMode Single Train job has only one task.
func (tp *module910x8) checkSingleTrainMode() error {
	taskNum := len(tp.Tasks)
	klog.V(util.LogDebugLev).Infof("checkSingleTrainMode job(%s) has %d tasks.", tp.JobName, taskNum)
	if taskNum > 1 {
		return fmt.Errorf("%s checkSingleTrainMode job<%s> single trainning has too many task:%d",
			tp.GetPluginName(), tp.JobName, taskNum)
	}

	jobNPU := tp.ReqNPUNum
	if jobNPU == 1 || jobNPU == npuIndex2 || jobNPU == npuNumPerHccs || jobNPU == nodeNPUNumber {
		return nil
	}
	return fmt.Errorf("%s checkSingleTrainMode job<%s> req npu num is not [1 or 2 or 4 or 8]",
		tp.GetPluginName(), tp.JobName)
}

// If job requires more than 8 npu, every task need 8 npu.
func (tp *module910x8) checkModuleDistributeTrainMode() error {
	taskNum := len(tp.Tasks)

	klog.V(util.LogDebugLev).Infof("%s Module DistributeTrainMode %s has %d tasks.",
		tp.GetPluginName(), tp.JobName, taskNum)

	for _, task := range tp.Tasks {
		taskNPU := task.ReqNPUNum
		klog.V(util.LogDebugLev).Infof("%s  checkModuleDistributeTrainMode Module DistributeTrain %s has %d npu.",
			tp.GetPluginName(), task.TaskName, taskNPU)

		if taskNPU != tp.MaxNodeNPUNum {
			return fmt.Errorf("checkModuleDistributeTrainMode distributeTrain task<%s> req npu[%d] "+
				"not equal [%d]", task.TaskName, taskNPU, tp.MaxNodeNPUNum)
		}
	}
	return nil
}
