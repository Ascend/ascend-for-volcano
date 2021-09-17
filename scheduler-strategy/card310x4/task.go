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

Package card310x4 is using for HuaWei A300T Ascend pin affinity schedule.

*/
package card310x4

import (
	"errors"
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/util"
)

// According to need card number, get best node from 4 pri-node-list-group.
func getBestNodesMap(priNodeGroups []map[string]*npuPriNodeInf) (map[string]int, error) {
	var bestNodesMap = make(map[string]int, cardNPUNumber)

	for i := 0; i < cardNPUNumber; i++ {
		for nodeName := range priNodeGroups[i] {
			tmpName := nodeName
			bestNodesMap[tmpName] = i
		}
	}

	if len(bestNodesMap) == 0 {
		return nil, errors.New("none bestNodes")
	}

	return bestNodesMap, nil
}

func IsTaskOfCardModeFromLabel(task *api.TaskInfo) bool {
	taskSelectors := util.GetTaskLabels(task)
	if len(taskSelectors) == 0 {
		klog.V(logDebugLev).Infof("task(%s) has no selectors.", task.Name)
		return false
	}

	return util.ValidStringMapKeyAndValue(taskSelectors, acceleratorType, cardAcceleratorType)
}
