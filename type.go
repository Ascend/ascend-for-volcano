/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package main is using for HuaWei Ascend pin affinity schedule.

*/
package main

import (
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

// PluginName use in frame build.
var PluginName = "volcano-npu-v3.0.0"

type huaweiNPUPlugin struct {
	// Scheduler for plugin args and its handler.
	Scheduler *plugin.ScheduleHandler
	// Arguments given for the plugin
	Arguments framework.Arguments
}
