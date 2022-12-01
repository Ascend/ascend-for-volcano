/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package half910x4 is using for HuaWei A800/9000 Ascend910 pin affinity schedule.

*/
package half910x4

import (
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/base"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/rescheduling"
)

const (
	// SchedulerName Ascend910half plugin name
	SchedulerName       = "huawei.com/Ascend910half"
	npuNumPerHccs       = 4
	networkUnhealthyNPU = "huawei.com/Ascend910-NetworkUnhealthy"
	nodeWeight          = 8.0
)

const (
	affScore0 = iota
	affScore1
	affScore2
	affScore3
	affScore4
)

type half910x4 struct {
	base.NPUHandler
	netUnhealthyKey string
	affScoreList    [][]int
	reHandle        *rescheduling.ReScheduler
}
