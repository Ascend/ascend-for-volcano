/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package vnpu is using for HuaWei Ascend pin fault rescheduling.

*/
package vnpu

import (
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/base"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

const (
	// VJobStatusUnhandled not handled
	VJobStatusUnhandled = "Unhandled"
	// VJobStatusNotPreSegmented not pre-allocated
	VJobStatusNotPreSegmented = "NotPreSegmented"
	// VJobStatusPreSegmented pre-allocated
	VJobStatusPreSegmented = "PreSegmented"
	// VJobStatusSegmented segmented
	VJobStatusSegmented = "Segmented"
	// VJobStatusAllocated allocated
	VJobStatusAllocated = "Allocated"
	// VJobStatusDestroying destroying
	VJobStatusDestroying = "Destroying"
	// VJobStatusDestroyed destroyed
	VJobStatusDestroyed = "Destroyed"

	// VNPUNodeLabelValue for select vnpu node label value.
	VNPUNodeLabelValue = "vnpu"
	// DeleteOverTime over time for job finish deal.
	DeleteOverTime = 5
	// JobPendingWaitTime The time wait for device-plugin create vnpu.
	JobPendingWaitTime = 300
	// VNPUScoreWeight for volcano select vnpu node core.
	VNPUScoreWeight = 64
	// PreAllocateFailureWaitTime wait time to judge pre-allocation failure
	PreAllocateFailureWaitTime = 10
)

// ComVNPU vnpu struct
type ComVNPU struct {
	*ComVNPUHandler
	// The vNPU chip divide name. Like huawei.com/Ascend910-16c,huawei.com/Ascend910-8c and so on.
	DivideKinds []string
	// divide vNPU coefficient for each chip.
	Coefficients map[string]int
	// the source of NPU cores in node annotation. like huawei.com/Ascend910-spec.
	NPUCardCoreKey string
}

// ComVNPUHandler vnpu handler
type ComVNPUHandler struct {
	*Action
	vNodes map[string]plugin.VNode
	base.NPUHandler
}

type staticVNPU struct {
}

type dynamicVNPU struct {
}

// Action vnpu actions
type Action struct {
	template map[string]util.VResource
}
