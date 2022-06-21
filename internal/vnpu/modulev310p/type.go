/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package modulev310p is using for virtual HuaWei 310P chips schedule.

*/
package modulev310p

import (
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/vnpu/vnpuutil"
)

const (
	// PluginName the CardV310Px2's plugin name.
	PluginName             = vnpuutil.PluginNameBy310PVNPU
	npu310PCardName        = vnpuutil.NPU310PCardName
	npuV310PCardName1c     = "huawei.com/Ascend310P-1c"
	npuV310PCardName2c     = "huawei.com/Ascend310P-2c"
	npuV310PCardName4c     = "huawei.com/Ascend310P-4c"
	npuV310PCardName4C3Cpu = "huawei.com/Ascend310P-4c.3cpu"
	npuV310PCardName2C1Cpu = "huawei.com/Ascend310P-2c.1cpu"
	npuV310PCardCoef1c     = 1
	npuV310PCardCoef2c     = 2
	npuV310PCardCoef4c     = 4
)

// ChipV310P 310P VNPU plugin struct
type ChipV310P struct {
	vnpuutil.ComVNPU
}
