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
	// The vNPU chip divide name. Like huawei.com/Ascend910-16c,huawei.com/Ascend910-8c and so on.
	DivideKinds []string
	// divide vNPU coefficient for each chip.
	Coefficients map[string]int
	// the source of NPU cores in node annotation. like huawei.com/Ascend910-spec.
	NPUCardCoreKey string
}

// CardVNPUs record virtual NPU Chip allocation information in one card.
type CardVNPUs struct {
	// like Ascend910-7
	CardName string
	// like huawei.com/Ascend910-2c
	Req []string
	// like Ascend910-2c-212-7
	Alloc []string
}

// NodeVNPUs record virtual NPU Chip allocation information in one node. Will be solidified into cm.
type NodeVNPUs struct {
	// NodeName preallocate node name.
	NodeName string
	// VNPUs record VNPUs Cutting and requirements in k8s.
	Cards []CardVNPUs
}

// VNPUCM record virtual NPU Chip allocation information in k8s. Will be solidified into cm.
type VNPUCM struct {
	// VNPUs record VNPUs Cutting and requirements in k8s.
	Nodes []NodeVNPUs
	// UpdateTime data update time.
	UpdateTime int64
	// CheckCode
	CheckCode uint32
}
