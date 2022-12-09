/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package card910x2 is using for HuaWei Ascend pin affinity schedule.
*/
package card910x2

import (
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/base"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/rescheduling"
)

type card910x2 struct {
	base.NPUHandler
	affScoreList [][]int
	reHandle     *rescheduling.ReScheduler
}

const (
	// SchedulerName name of scheduler
	SchedulerName = "huawei.com/Ascend910card"
	maxNodeNPUNum = 2
	npuNumPerHccs = 4
	nodeWeight    = 8.0
)

const (
	affScore0 = iota
	affScore1
	affScore2
)
