/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.

# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
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
	klog.V(util.LogDebugLev).Infof("checkSingleTrainMode job(%s) has %d tasks.", tp.JobName, len(tp.Tasks))

	jobNPU := tp.ReqNPUNum
	if jobNPU == 1 || jobNPU == util.NPUIndex2 || jobNPU == npuNumPerHccs {
		return nil
	}
	return fmt.Errorf("%s checkSingleTrainMode job<%s> req npu num is not [1 or 2 or 4]",
		tp.GetPluginName(), tp.JobName)
}

// If job requires more than 8 npu, every task need 8 npu.
func (tp *half910x4) checkModuleDistributeTrainMode() error {
	klog.V(util.LogDebugLev).Infof("half DistributeTrainMode %s has %d tasks.", tp.JobName, len(tp.Tasks))

	for _, task := range tp.Tasks {
		taskNPU := task.ReqNPUNum
		klog.V(util.LogDebugLev).Infof("checkModuleDistributeTrainMode half DistributeTrain %s has %d npu.",
			task.TaskName, taskNPU)

		if taskNPU != tp.MaxNodeNPUNum {
			return fmt.Errorf("checkModuleDistributeTrainMode half distributeTrain task<%s> req npu[%d] "+
				"not equal [%d]", task.TaskName, taskNPU, tp.MaxNodeNPUNum)
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
