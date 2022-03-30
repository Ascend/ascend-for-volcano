/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
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

// IsTaskOfCardModeFromLabel judge task is card mode or card mod by label.
func IsTaskOfCardModeFromLabel(task *api.TaskInfo) bool {
	taskSelectors := util.GetTaskLabels(task)
	if len(taskSelectors) == 0 {
		klog.V(util.LogDebugLev).Infof("task(%s) has no selectors.", task.Name)
		return false
	}

	return util.ValidStringMapKeyAndValue(taskSelectors, acceleratorType, cardAcceleratorType)
}
