/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package modulev710 is using for virtual HuaWei 710 chips schedule.

*/
package modulev710

import (
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/vnpu/vnpuutil"
)

const (
	// PluginName the CardV710x2's plugin name.
	PluginName        = vnpuutil.PluginNameBy710VNPU
	npu710CardName    = vnpuutil.NPU710CardName
	npuV710CardName1c = "huawei.com/Ascend710-1c"
	npuV710CardName2c = "huawei.com/Ascend710-2c"
	npuV710CardName4c = "huawei.com/Ascend710-4c"
	npuV710CardCoef1c = 1
	npuV710CardCoef2c = 2
	npuV710CardCoef4c = 4
)

// ChipV710 710 VNPU plugin struct
type ChipV710 struct {
	vnpuutil.ComVNPU
}
