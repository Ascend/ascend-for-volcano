/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
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
