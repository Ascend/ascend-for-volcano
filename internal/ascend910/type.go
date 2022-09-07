/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package ascend910 is using for HuaWei Ascend pin affinity schedule.

*/
package ascend910

import (
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/base"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/rescheduling"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

type ascend910 struct {
	// base event handler
	base.NPUHandler
	// 910 support scheduler kinds.
	Kind map[string]base.AscendHandler
	// specific job use.
	handle   base.AscendHandler
	reHandle *rescheduling.ReScheduler
}

const (
	// PluginName name of plugin
	PluginName = util.NPU910CardName

	Accelerator910Key         = "accelerator-type"
	Module910AcceleratorValue = "module"
	Card910AcceleratorValue   = "card"
)
