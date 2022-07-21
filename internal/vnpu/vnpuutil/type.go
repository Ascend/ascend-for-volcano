/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package vnpuutil is using for virtual HuaWei Ascend910 schedule.

*/
package vnpuutil

import (
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

const (
	// PluginName the vNPU's plugin name.
	PluginName = "Vnpu"
	// NPUCardNamePrefix huawei.com/ for judge npu resource.
	NPUCardNamePrefix = "huawei.com/"
	// NPUIdentifyName to identify the NPU
	NPUIdentifyName = util.CommCardPreName
	// NPU910CardName for judge 910 npu resource.
	NPU910CardName = "huawei.com/Ascend910"
	// NPU310PCardName for judge 310P npu resource.
	NPU310PCardName = "huawei.com/Ascend310P"
	// NPU910CardCoreKey for npu card core. like chipId-allCores-freeCores example:4-32c-4c
	NPU910CardCoreKey = "huawei.com/Ascend910-spec"

	// VNPCMDataKey cm date key
	VNPCMDataKey = "VNPUCfg"
)

// ComVNPU common type
type ComVNPU struct {
	// vNPU chip name. Like cardV910x2,chip310p,moduleV910x8 and so on.
	plugin.HwEntity
}
