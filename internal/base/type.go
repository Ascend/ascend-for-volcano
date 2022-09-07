/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package base is using for HuaWei Ascend pin affinity schedule.

*/
package base

import (
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// AscendHandler ascend npu event handler
type AscendHandler interface {
	plugin.ISchedulerPlugin
	SetSchedulerAttr(util.SchedulerJobAttr)
	SetSchedulerEnv(plugin.ScheduleEnv)
	SetMaxNodeNPUNum(int)
	SetMaxCardNPUNum(int)
}

// NPUHandler base npu handler
type NPUHandler struct {
	plugin.SchedulerPlugin
	util.SchedulerJobAttr
	plugin.ScheduleEnv
	MaxNodeNPUNum int
	MaxCardNPUNum int
}
