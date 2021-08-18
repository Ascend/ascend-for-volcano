/*
Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.

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

Package commonv910 is using for virtual HuaWei Ascend910 schedule.

*/
package commonv910

import (
	"errors"
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	hwutil "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/util"
)

// Check whether the selector of the node matches that of the task.
func isSelectorMeetNode(task *api.TaskInfo, node *api.NodeInfo, conf map[string]string) error {
	// Get node selectors of task
	taskSelectors := hwutil.GetTaskSelectors(task)
	if taskSelectors == nil || len(taskSelectors) == 0 {
		for _, v := range VnpuType {
			if err := hwutil.IsNPUTask(task, v); err == nil {
				// Vnpu task should have node selector
				klog.V(logErrorLev).Infof("task[%s] no selector in select node[%s].", task.Name, node.Name)
				return errors.New(nodeNoFitSelectorError)
			}
		}
		return nil
	}

	// Get node selectors of node
	nodeSelector, errNode := hwutil.GetNodeSelector(node)
	if errNode != nil {
		klog.V(logErrorLev).Infof("task[%s] on node(%s) %v.", task.Name, node.Name, errNode)
		return errors.New(nodeNoFitSelectorError)
	}

	if err := hwutil.CheckTaskAndNodeSelectorMeet(taskSelectors, nodeSelector, conf); err != nil {
		klog.V(logErrorLev).Infof("isSelectorMeetNode %s not meet %s err:%v.", task.Name, node.Name, err)
		return err
	}

	return nil
}
