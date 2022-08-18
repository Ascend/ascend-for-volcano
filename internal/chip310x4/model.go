/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package chip310x4 is using for HuaWei A300T Ascend pin affinity schedule.

*/
package chip310x4

import (
	"k8s.io/klog"
)

func (tp *chip310x4) getNPUAllocPriorityArray() ([cardNPUNumber]int, error) {
	var priorityArray [cardNPUNumber]int
	var err = error(nil)
	// priority:1>2>3>4
	priorityArray = [cardNPUNumber]int{1, constIntNum2, constIntNum3, cardNPUNumber}
	if err != nil {
		klog.V(logErrorLev).Infof("%s %s.", PluginName, err.Error())
		return priorityArray, err
	}
	return priorityArray, nil
}

// getNPUIndex get NPU index by cardNumGroups
func (tp *chip310x4) getNPUIndex(cardNumGroups [][]int) map[int][]int {
	npuNumberIndex := make(map[int][]int, constIntNum5)
	for _, cardNumGroup := range cardNumGroups {
		index := len(cardNumGroup)
		npuNumberIndex[index] = append(npuNumberIndex[index], cardNumGroup...)
	}
	return npuNumberIndex
}
