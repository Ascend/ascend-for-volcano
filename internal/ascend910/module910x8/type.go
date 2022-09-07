/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package module910x8 is using for HuaWei Ascend pin affinity schedule.

*/

package module910x8

import (
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/base"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/rescheduling"
)

type module910x8 struct {
	base.NPUHandler
	netUnhealthyKey string
	affScoreList    [][]int
	reHandle        *rescheduling.ReScheduler
}

const (
	// SchedulerName module910x8 plugin name
	SchedulerName = "huawei.com/Ascend910module"
	npuIndex2     = 2
	npuIndex3     = 3
	npuNumPerHccs = 4
	nodeNPUNumber = 8
	nodeWeight    = 8.0

	networkUnhealthyNPU = "huawei.com/Ascend910-NetworkUnhealthy"
)

const (
	affScore0 = iota
	affScore1
	affScore2
	affScore3
	affScore4
)

type selectNodeInf struct {
	nodeName    string
	allNPUNum   int
	leftNPUNum  int
	rightNPUNum int
}
