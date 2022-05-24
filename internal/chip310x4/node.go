/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package chip310x4 is using for HuaWei A300T Ascend pin affinity schedule.

*/
package chip310x4

import "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/util"

// getCardNumGroupsFromTop get the chip for each card from nodeTop
func (tp *chip310x4) getCardNumGroupsFromTop(nodeNPUTopology []int) [][]int {
	maxCardNum := 0
	for _, v := range nodeNPUTopology {
		maxCardNum = max(maxCardNum, v)
	}
	cardNumGroups := make([][]int, maxCardNum/util.NPUIndex4+1, maxCardNum/util.NPUIndex4+1)
	for _, v := range nodeNPUTopology {
		cardNumGroups[v/util.NPUIndex4] = append(cardNumGroups[v/util.NPUIndex4], v)
	}
	return cardNumGroups
}

func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}
