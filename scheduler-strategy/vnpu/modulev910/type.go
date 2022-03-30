/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package modulev910x8 is using for virtual HuaWei Ascend910 schedule.

*/
package modulev910

import (
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/vnpu/vnpuutil"
)

const (
	// PluginName the ChipV910's plugin name.
	PluginName         = vnpuutil.PluginNameBy910VNPU
	npu910CardName     = vnpuutil.NPU910CardName
	npuV910CardName2c  = "huawei.com/Ascend910-2c"
	npuV910CardName4c  = "huawei.com/Ascend910-4c"
	npuV910CardName8c  = "huawei.com/Ascend910-8c"
	npuV910CardName16c = "huawei.com/Ascend910-16c"
	npuV910CardCoef2c  = 2
	npuV910CardCoef4c  = 4
	npuV910CardCoef8c  = 8
	npuV910CardCoef16c = 16
)

// ChipV910 910 VNPU plugin struct. Include 910x8 and 910x2.
type ChipV910 struct {
	vnpuutil.ComVNPU
}
