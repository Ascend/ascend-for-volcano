/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package chip310x4 is using for HuaWei Ascend pin affinity schedule.

*/

package chip310x4

import "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"

func (tp *chip310x4) getNPUIndex(cardNumGroups [][]int) map[int][]int {
	npuNumberIndex := make(map[int][]int, util.MapInitNum)
	for _, cardNumGroup := range cardNumGroups {
		index := len(cardNumGroup)
		npuNumberIndex[index] = append(npuNumberIndex[index], cardNumGroup...)
	}
	return npuNumberIndex
}
