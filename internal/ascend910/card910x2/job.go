/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package card910x2 is using for HuaWei Ascend pin affinity schedule.

*/
package card910x2

import (
	"fmt"

	"k8s.io/klog"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

func (tp *card910x2) checkSingleTrainMode() error {
	taskNum := len(tp.Tasks)

	klog.V(util.LogDebugLev).Infof("%s checkSingleTrainMode job<%s> has <%d> tasks.",
		tp.GetPluginName(), tp.JobName, taskNum)

	if taskNum > 1 {
		return fmt.Errorf("%s checkSingleTrainMode single trainning job<%s> has too many task: <%d>",
			tp.GetPluginName(), tp.JobName, taskNum)
	}

	return nil
}

func (tp *card910x2) checkCardDistributeTrainMode() error {
	taskNum := len(tp.Tasks)

	klog.V(util.LogDebugLev).Infof("%s check Card Mode %s has %d tasks.",
		tp.GetPluginName(), tp.JobName, taskNum)

	for _, task := range tp.Tasks {
		taskNPU := task.ReqNPUNum

		klog.V(util.LogDebugLev).Infof("%s check Card Mode %s require %d npu.",
			tp.GetPluginName(), task.TaskName, taskNPU)

		if taskNPU != tp.MaxNodeNPUNum {
			return fmt.Errorf("%s checkCardDistributeTrainMode distributed card train job<%s> req npu<%d>"+
				" not equal<%d>", tp.GetPluginName(), task.TaskName, taskNPU, tp.MaxNodeNPUNum)
		}
	}

	return nil
}
