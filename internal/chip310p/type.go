/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package chip310p is using for HuaWei A300T Ascend pin affinity schedule.

*/
package chip310p

import "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/common"

const (
	archSelector  = "host-arch"
	huaweiArchArm = "huawei-arm"
	huaweiArchX86 = "huawei-x86"

	// PluginName the chip310p's plugin name.
	PluginName          = "A310P"
	a310PNPUChipName    = "huawei.com/Ascend310P"
	a310PNPUCardPreName = "Ascend310P-"
	a310PFaultNPUName   = "huawei.com/Ascend310P-Unhealthy"
)

type chip310P struct {
	com common.Scheduler
	re  common.ReScheduler
}
