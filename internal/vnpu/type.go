/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package vnpu is using for HuaWei Ascend pin vnpu allocation.
*/
package vnpu

import (
	"volcano.sh/volcano/pkg/scheduler/api"

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

	// PodEventMsgAllocateFailed dp pod segment failed msg
	PodEventMsgAllocateFailed = "NoNPUAffinity"
	// PodEventReasonAllocateFailed dp pod segment failed reason
	PodEventReasonAllocateFailed = "UnexpectedAdmissionError"
	// PodEventTypeAllocateFailed dp pod segment failed type
	PodEventTypeAllocateFailed = "Warning"
	podObjectType              = "Pod"
)

// VTemplate vNPU template
type VTemplate struct {
	Data map[string]util.VResource
}

// VNPU vnpu struct
type VNPU struct {
	VT VTemplate
	StaticVNPU
	DynamicVNPU
}

type StaticVNPU struct {
	vnpuHandler
}

type DynamicVNPU struct {
	vnpuHandler
}

type vnpuHandler interface {
	CheckNodeNPUByTask(*api.TaskInfo, plugin.NPUNode, util.VResource) error
	ScoreBestNPUNodes(*api.TaskInfo, []*api.NodeInfo, map[string]float64) error
	UseAnnotation(*api.TaskInfo, plugin.NPUNode, util.VResource) *plugin.NPUNode
}

// Action vnpu actions
type Action struct {
	template map[string]util.VResource
}
