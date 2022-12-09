/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package chip310x4 is using for HuaWei 310 Ascend pin affinity schedule.
*/
package chip310x4

import "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/base"

type chip310x4 struct {
	base.NPUHandler
}

const (
	// SchedulerName chip310 plugin name
	SchedulerName = "huawei.com/Ascend310chip"
	maxNodeNPUNum = 64
	maxCardNPUNum = 4
)
