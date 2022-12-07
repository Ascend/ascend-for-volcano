/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

// Package plugin is using for HuaWei Ascend pin affinity schedule.
package plugin

import (
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/conf"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

const (
	// PluginName the HuaWei NPU 's plugin name.
	PluginName = "huaweiNPU"

	nodesNoMeetNPUReqError = "insufficient npus on the schedulable nodes in cluster"
	nodeNoFitSelectorError = "no matching label on this node"
	objectNilError         = "object or argument is nil"
	podRankIndex           = "hccl/rankIndex"

	// FormatIncorrectError format incorrect error
	FormatIncorrectError = "format incorrect"

	// AscendVNPULevel vnpu level
	AscendVNPULevel = "vnpu-level"
	// AscendVNPULevelLow low
	AscendVNPULevelLow = "low"
	// AscendVNPULevelHigh high
	AscendVNPULevelHigh = "high"
	// AscendVNPUPrefix vir
	AscendVNPUPrefix = "vir"
	// AscendVNPUDVPP dvpp enable
	AscendVNPUDVPP = "vnpu-dvpp"
	// AscendDVPPEnabledOff off
	AscendDVPPEnabledOff = "no"
	// AscendDVPPEnabledNull null
	AscendDVPPEnabledNull = "null"
	// AscendDVPPEnabledOn on
	AscendDVPPEnabledOn = "yes"
	// AscendNDVPPValue value
	AscendNDVPPValue = "ndvpp"
	// AscendDVPPValue value
	AscendDVPPValue = "dvpp"

	// Ascend310P 310P template name
	Ascend310P = "Ascend310P"
	// Ascend910 910 template name
	Ascend910 = "Ascend910"
	// RingController310P 310p ring controller name
	RingController310P = "ascend-310P"
	// RingController910 910 ring controller name
	RingController910 = "ascend-910"
)

// SchedulerJob the plugin define job info
type SchedulerJob struct {
	util.SchedulerJobAttr
	handler ISchedulerPlugin
}

// VolcanoFrame passed in by the volcano frame.
type VolcanoFrame struct {
	UID        types.UID
	Conf       []conf.Configuration
	KubeClient kubernetes.Interface
}

// ScheduleCache the plugin defined caches saving cm data
type ScheduleCache struct {
	// special, name, value
	Names, Namespaces map[string]string
	Data              map[string]map[string]string
}

// ScheduleEnv for job scheduler context.
type ScheduleEnv struct {
	Jobs      map[api.JobID]SchedulerJob
	Nodes     map[string]NPUNode
	FrameAttr VolcanoFrame
	Cache     ScheduleCache
}

// ScheduleHandler information for the current plugin
type ScheduleHandler struct {
	NPUPlugins map[string]ISchedulerPlugin
	ScheduleEnv
}
