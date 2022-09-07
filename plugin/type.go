/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package plugin is using for HuaWei Ascend pin affinity schedule.

*/

package plugin

import (
	v1 "k8s.io/api/core/v1"
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
)

type SchedulerJob struct {
	util.SchedulerJobAttr
	handler ISchedulerPlugin
}

// NPUNode the plugin define node info.
type NPUNode struct {
	Name       string
	Capability map[v1.ResourceName]float64
	Allocate   map[v1.ResourceName]float64
	Idle       map[v1.ResourceName]float64
	Annotation map[string]string
	Label      map[string]string
}

// VolcanoFrame passed in by the volcano frame.
type VolcanoFrame struct {
	UID        types.UID
	Conf       []conf.Configuration
	KubeClient kubernetes.Interface
}

type ScheduleCache struct {
	// special name, value
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
