/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package card310x4 is using for HuaWei A300T Ascend pin affinity schedule.
*/
package card310x4

import (
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/base"
)

type card310x4 struct {
	base.NPUHandler
	affScoreList [][]int
}

const (
	affScore0 = iota
	affScore1
	affScore2
	affScore3
	affScore4

	// SchedulerName card310 plugin name
	SchedulerName  = "huawei.com/Ascend310card"
	maxNodeNPUNum  = 64
	maxCardNPUNum  = 4
	constNPUWeight = 8.0
)
