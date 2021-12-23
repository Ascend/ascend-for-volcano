/*
Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package chip310x4 is using for HuaWei A300T Ascend pin affinity schedule.

*/
package chip310x4

import "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/common"

const (
	// PluginName the chip310x4's plugin name.
	PluginName         = "A310-chip"
	a310NPUChipName    = "huawei.com/Ascend310"
	a310NPUCardPreName = "Ascend310-"
	a310FaultNPUName   = "huawei.com/Ascend310-Unhealthy"

	archSelector  = "host-arch"
	huaweiArchArm = "huawei-arm"
	huaweiArchX86 = "huawei-x86"

	acceleratorType     = "npu-310-strategy"
	cardAcceleratorType = "card"
	chipAcceleratorType = "chip"

	constIntNum2  = 2
	constIntNum3  = 3
	cardNPUNumber = 4
	constIntNum5  = 5

	logErrorLev = 1
	logInfoLev  = 3
	logDebugLev = 4

	modeNotChip = "no chip mode npu"
)

type chip310x4 struct {
	com common.Scheduler
	re  common.ReScheduler
}
