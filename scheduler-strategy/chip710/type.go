/*
Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package chip710 is using for HuaWei A300T Ascend pin affinity schedule.

*/
package chip710

import "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/common"

const (
	archSelector  = "host-arch"
	huaweiArchArm = "huawei-arm"
	huaweiArchX86 = "huawei-x86"

	// PluginName the chip710's plugin name.
	PluginName         = "A710"
	a710NPUChipName    = "huawei.com/Ascend710"
	a710NPUCardPreName = "Ascend710-"
	a710FaultNPUName   = "huawei.com/Ascend710-Unhealthy"
)

type chip710 struct {
	com common.Scheduler
	re  common.ReScheduler
}
