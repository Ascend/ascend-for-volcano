/*
Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package modulev910x8 is using for virtual HuaWei Ascend910 schedule.

*/
package modulev910x8

import "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/commonv910"

const (
	// PluginName the modulev910x8's plugin name.
	PluginName = "A800-9000-VNpu"

	maxNPUNum   = 8
	logDebugLev = 4
)

type modulev910x8 struct {
	name string
	commonv910.Vnpu
}
