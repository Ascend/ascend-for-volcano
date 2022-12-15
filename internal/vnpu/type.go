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
	// PodEventMsgAllocateFailed dp pod segment failed msg
	PodEventMsgAllocateFailed = "NoNPUAffinity"
	// PodEventReasonAllocateFailed dp pod segment failed reason
	PodEventReasonAllocateFailed = "UnexpectedAdmissionError"
	// DyCutFailedError for device-plugin cut failed error.
	DyCutFailedError = "chipDyCutErr"
	// PodEventTypeAllocateFailed dp pod segment failed type
	PodEventTypeAllocateFailed = "Warning"
	podObjectType              = "Pod"
)

// VTemplate vNPU template
type VTemplate struct {
	Data map[string]util.VResource
}

// VirtualNPU vnpu struct
type VirtualNPU struct {
	DynamicByConf bool
	VT            VTemplate
	StaticVNPU
	DynamicVNPU
}

// StaticVNPU Static VNPU struct.
type StaticVNPU struct {
	vnpuHandler
}

// DynamicVNPU dynamic VNPU struct.
type DynamicVNPU struct {
	vnpuHandler
	DowngradeCache map[string][]string // taskName: nodes
	// for Concurrent task. not same core request task only has one on a node in same time.
	// nodeName: templateName:taskUID
	ConCache map[string]map[string]map[api.TaskID]struct{}
}

type vnpuHandler interface {
	CheckNodeNPUByTask(*api.TaskInfo, plugin.NPUNode, util.VResource) error
	ScoreBestNPUNodes(*api.TaskInfo, []*api.NodeInfo, map[string]float64) error
	UseAnnotation(*api.TaskInfo, plugin.NPUNode, util.VResource, VTemplate) *plugin.NPUNode
}

// Action vnpu actions
type Action struct {
	template map[string]util.VResource
}
