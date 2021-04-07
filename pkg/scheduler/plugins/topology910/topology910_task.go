/*
Copyright(C) 2020. Huawei Technologies Co.,Ltd. All rights reserved.

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

Package topology910 is using for HuaWei Ascend910 pin affinity schedule.

*/
package topology910

import (
	"errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog"
	"strings"
	"volcano.sh/volcano/pkg/scheduler/api"
)

func getTaskSelectors(task *api.TaskInfo) map[string]string {
	return task.Pod.Spec.NodeSelector
}

// true is card, false is module
func isTaskOfCardMode(task *api.TaskInfo) bool {
	taskSelectors := getTaskSelectors(task)
	if taskSelectors == nil || len(taskSelectors) == 0 {
		klog.V(logDebugLev).Infof("task(%s) has no selectors", task.Name)
		return false
	}

	acceleratorValue, ok := taskSelectors[acceleratorType]
	if !ok {
		// no acceleratorType means module
		klog.V(logDebugLev).Infof("task(%s) is module type", task.Name)
		return false
	}

	if acceleratorValue == cardAcceleratorType {
		klog.V(logDebugLev).Infof("task(%s) is card type", task.Name)
		return true
	}

	klog.V(logDebugLev).Infof("task(%s) is module type", task.Name)
	return false
}

func getTaskNpuNum(task *api.TaskInfo) (int, error) {
	tmpNpu, ok := task.Resreq.ScalarResources[npu910CardName]
	if !ok || int(tmpNpu/npuHex) == 0 {
		return 0, errors.New("not npu task")
	}

	taskNpu := int(tmpNpu / npuHex)
	return taskNpu, nil
}

func isNpuTask(task *api.TaskInfo) error {
	tmpNpu, ok := task.Resreq.ScalarResources[npu910CardName]
	if !ok || int(tmpNpu/npuHex) == 0 {
		return errors.New("not npu task")
	}

	return nil
}

func updatePodsFailedReason(job *api.JobInfo, reasonTmp string) {
	for _, task := range job.Tasks {
		condition := v1.PodCondition{
			Type:    v1.PodScheduled,
			Status:  v1.ConditionFalse,
			Reason:  v1.PodReasonUnschedulable,
			Message: reasonTmp,
		}

		task.Pod.Status.Conditions = append(task.Pod.Status.Conditions, condition)
	}
}

func getTaskModule(task *api.TaskInfo) string {
	var taskModule = moduleAcceleratorType

	taskSelector := getTaskSelectors(task)
	for _, selector := range taskSelector {
		if strings.Contains(selector, cardAcceleratorType) {
			taskModule = cardAcceleratorType
			break
		}
	}

	return taskModule
}
