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
