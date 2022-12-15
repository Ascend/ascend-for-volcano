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
