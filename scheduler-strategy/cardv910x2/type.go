/*
Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package cardv910x2 is using for virtual HuaWei A300T schedule.

*/
package cardv910x2

import (
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/commonv910"
)

const (
	// PluginName the cardv910x2's plugin name.
	PluginName = "A300T-Vnpu"

	mapInitLen  = 3
	logDebugLev = 4

	archSelector          = "host-arch"
	huaweiArchArm         = "huawei-arm"
	huaweiArchX86         = "huawei-x86"
	acceleratorType       = "accelerator-type"
	cardAcceleratorType   = "card"
	moduleAcceleratorType = "module"

	maxNPUNum = 2
)

type cardv910x2 struct {
	name string
	commonv910.Vnpu
}
