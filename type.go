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

// PluginName use in frame
var PluginName = "volcano-npu-v2.0.4"

const (
	logErrorLev = 1
	logInfoLev  = 3
	logDebugLev = 4
)

type huaweiNPUPlugin struct {
	*plugin.ScheduleHandler
	// Arguments given for the plugin
	pluginArguments framework.Arguments
}
