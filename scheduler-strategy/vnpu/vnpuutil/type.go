/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package vnpuutil is using for virtual HuaWei Ascend910 schedule.

*/
package vnpuutil

import (
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/util"
)

const (
	// PluginName the vNPU's plugin name.
	PluginName = "Vnpu"
	// PluginNameBy910VNPU 910 vnpu plugin.
	PluginNameBy910VNPU = "910-VNPU"
	// PluginNameBy710VNPU 710 vnpu plugin.
	PluginNameBy710VNPU = "710-VNPU"
	// NPUCardNamePrefix huawei.com/ for judge npu resource.
	NPUCardNamePrefix = "huawei.com/"
	// NPUIdentifyName to identify the NPU
	NPUIdentifyName = util.CommCardPreName
	// NPU910CardName for judge 910 npu resource.
	NPU910CardName = "huawei.com/Ascend910"
	// NPU710CardName for judge 710 npu resource.
	NPU710CardName = "huawei.com/Ascend710"
	// UNKnownPluginName unknown vnpu plugin
	UNKnownPluginName = "unknown"
	// NPU910CardCoreKey for npu card core. like chipId-allCores-freeCores(4-32c-4c)
	NPU910CardCoreKey = "huawei.com/Ascend910-spec"
	// NPU710CardCoreKey for npu card core. like chipId-allCores-notCutCores(4-8c-4c)
	NPU710CardCoreKey = "huawei.com/Ascend710-spec"
	// NodesNoMeetNPUReqError error for insufficient npus on the schedulable nodes.
	NodesNoMeetNPUReqError = "insufficient npus on the schedulable nodes in cluster"
	// NodeNotStableWarning error for unstable npu node.
	NodeNotStableWarning = "the npus on this node are unstable"
	// PluginUninitializedError plugin not initialized Error.
	PluginUninitializedError = "uninitialized plugin"

	// VNPUCMNameSpace for uninstall volcano, also delete cm
	VNPUCMNameSpace = "volcano-system"
	// VNPUCMName the cm intercommunicate to device-plugin.
	VNPUCMName = "mindx-dl-vnpu-manager"
	// VNPCMDataKey cm date key
	VNPCMDataKey = "VNPUCfg"
	// VNPUCacheCMName solidified the vnpu pre-alloc cache.
	VNPUCacheCMName = "mindx-dl-vnpu-cache"
	// VNPUNodeLabelKey for select vnpu node label key.
	VNPUNodeLabelKey = "npu-spec"
	// VNPUNodeLabelValue for select vnpu node label value.
	VNPUNodeLabelValue = "vnpu"
	// DeleteOverTime over time for job finish deal.
	DeleteOverTime = 5
	// JobPendingWaitTime The time wait for device-plugin create vnpu.
	JobPendingWaitTime = 300
	// VNPUScoreWeight for volcano select vnpu node core.
	VNPUScoreWeight = 64
)

// ComVNPU common type
type ComVNPU struct {
	// vNPU chip name. Like cardV910x2,chip710,moduleV910x8 and so on.
	plugin.HwEntity
	// The vNPU chip divide name. Like huawei.com/Ascend910-16c,huawei.com/Ascend910-8c and so on.
	DivideKinds []string
	// divide vNPU coefficient for each chip.
	Coefficients map[string]int
	// the source of NPU cores in node annotation. like huawei.com/Ascend910-spec.
	NPUCardCoreKey string
}

// VNPUAllocInf record virtual NPU Chip allocation information. Will be solidified into cm.
type VNPUAllocInf struct {
	// JobUID preallocate task name.
	JobUID api.JobID
	// ReqNPUType for job req huawei.com/Ascend910-2c change to Ascend910-2c
	ReqNPUType string
	// NodeName preallocate node name.
	NodeName string
	// ReqCardName preallocate NPU chip name(Ascend910-0,Ascend710-0), for one job one pod one card.
	ReqCardName string
	// AllocCardName preallocate NPU chip name(Ascend710-2c-100-1), for one job one pod one card.
	AllocCardName string
	// AllocFlag preallocate deal flag :after preprocess success.
	AllocFlag bool
	// UpdateTime data update time for every job.
	UpdateTime int64
}

// VNPUAllocInfCache record virtual NPU Chip allocation information. Will be solidified into cm.
type VNPUAllocInfCache struct {
	Cache []VNPUAllocInf
	// CheckCode
	CheckCode uint32
}

// VNPUAllocData is the automatic cache cutting and destruction for VNPU.
var VNPUAllocData = VNPUAllocInfCache{}

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
