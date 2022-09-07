/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package ascend310P is using for HuaWei Ascend pin affinity schedule.

*/

package ascend310P

import (
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/base"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/rescheduling"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

type ascend310P struct {
	// base event handler
	base.NPUHandler
	reHandle *rescheduling.ReScheduler
}

const (
	// PluginName ascend31P plugin name
	PluginName    = util.NPU310PCardName
	maxNodeNPUNum = 64
)
