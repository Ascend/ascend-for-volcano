/*
Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.

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

/*

Package plugin is using for HuaWei Ascend pin affinity schedule frame.

*/
package plugin

import (
	"errors"
	"k8s.io/klog"
	"reflect"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/conf"
	"volcano.sh/volcano/pkg/scheduler/framework"
	npuapi "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/npuinterface"
)

// PluginBuilder plugin management
type NPUBuilder = func(string2 string) HwNPUSchedulerPlugin

// HwNPUSchedulerPlugin Define the npu scheduler policy which new npu scheduler need to realize.
type HwNPUSchedulerPlugin interface {
	// Name The unique name of npu scheduler policy, used by volcano frame.
	Name() string

	// OnHandlerStart The npu scheduler policy initial and common processing.
	OnHandlerStart(*ScheduleHandler)

	// IsMyTask For npu scheduler policy distinguish itself task.
	IsMyTask(*api.TaskInfo) error
	// IsMyNode For npu scheduler policy distinguish itself node.
	IsMyNode(*api.NodeInfo) error
	// IsMyJob For npu scheduler policy distinguish itself job.
	IsMyJob(*api.JobInfo) error
	// ValidNPUJobFn Valid the job part of npu scheduler policy, if not, disallowed.
	ValidNPUJobFn(*api.JobInfo) *api.ValidateResult
	// Check whether the node matches the tag requirements of the task.
	PreCheckNodeFn(*api.TaskInfo, *api.NodeInfo, []conf.Configuration) error
	// Check whether the resources on the node are stable.
	CheckNPUResourceStableFn(*api.NodeInfo) error
	// Check whether the requested resource exists and are sufficient on the node.
	CheckNodeNPUByTaskFn(*api.TaskInfo, *api.NodeInfo) error
	// Get all the nodes,witch meet the npu affinity condition.
	GetNPUAffinityBestNodesFn(*api.TaskInfo, []*api.NodeInfo) (map[string]int, error)
	// Score the nodes.
	ScoreBestNPUNodesFn(map[string]float64, map[string]int, *api.TaskInfo, []*api.NodeInfo) (map[string]float64, error)
	// Get the pod's npu card to record in node others.
	GetAllocatedNPUFromTopologyFn(*api.TaskInfo, *api.NodeInfo) (interface{}, error)
	// Update node others.
	UpdateNPUNodeUsedCardFn(*api.NodeInfo, interface{}) error
	// Set the npu cared id into pod.
	SetNPUTopologyToPodFn(*api.TaskInfo, interface{}) error
	// Update the node using npu when release pod's npu.
	UpdateReleaseNPUNodeTopologyFn(*api.NodeInfo, interface{}) error
	// GetReleaseNPUTopologyFn Get the release npu card id from task.
	GetReleaseNPUTopologyFn(*api.TaskInfo) (interface{}, error)
}

// ScheduleHandler information for the current plugin
type ScheduleHandler struct {
	// for object
	HuaweiNPUs map[string]HwNPUSchedulerPlugin
	// For object functions
	InitNodesNPUAllocTopologyFns map[string]npuapi.InitNodesNPUTopologyFn
}

// AddInitNodesNPUAllocTopology add node NPU assignable topology chip function
func (hwNPU *ScheduleHandler) AddInitNodesNPUAllocTopology(name string, fn npuapi.InitNodesNPUTopologyFn) {
	hwNPU.InitNodesNPUAllocTopologyFns[name] = fn
	klog.V(logDebugLev).Infof("InitNodesNPUAllocTopology :%v add.", name)
}

// RegisterNPUScheduler register the plugin
func (hwNPU *ScheduleHandler) RegisterNPUScheduler(name string, pc NPUBuilder) {
	hwNPU.HuaweiNPUs[name] = pc(name)
	klog.V(logInfoLev).Infof("NPU Scheduler[%v] registered.", name)
}

// GetNPUScheduler get the NPU scheduler by name
func (hwNPU *ScheduleHandler) GetNPUScheduler(name string) (HwNPUSchedulerPlugin, bool) {
	pb, found := hwNPU.HuaweiNPUs[name]
	if found && pb != nil {
		return pb, found
	}

	return nil, found
}

// InitNPUSession init npu plugin and nodes
func (hwNPU *ScheduleHandler) InitNPUSession(ssn *framework.Session) {
	if err := hwNPU.initNodesNPUAllocTopology(ssn.Nodes); err != nil {
		klog.V(logErrorLev).Infof("%s InitNPUSession failed :%v.", PluginName, err)
	}
}

func (hwNPU *ScheduleHandler) getHWPluginByTask(task *api.TaskInfo) (HwNPUSchedulerPlugin, error) {
	var err = error(nil)
	var hwNPUPlugin HwNPUSchedulerPlugin

	for _, myNPUPlugin := range hwNPU.HuaweiNPUs {
		if myNPUPlugin == nil {
			continue
		}

		if err = myNPUPlugin.IsMyTask(task); err != nil {
			continue
		}
		hwNPUPlugin = myNPUPlugin
		err = nil
		break
	}

	return hwNPUPlugin, err
}

func (hwNPU *ScheduleHandler) getHWPluginByJob(job *api.JobInfo) (HwNPUSchedulerPlugin, error) {
	var err = error(nil)
	var hwNPUPlugin HwNPUSchedulerPlugin

	for _, myNPUPlugin := range hwNPU.HuaweiNPUs {
		if myNPUPlugin == nil {
			continue
		}

		if err = myNPUPlugin.IsMyJob(job); err != nil {
			continue
		}
		hwNPUPlugin = myNPUPlugin
		err = nil
		break
	}

	return hwNPUPlugin, err
}

func (hwNPU *ScheduleHandler) getHWPluginByNode(node *api.NodeInfo) (HwNPUSchedulerPlugin, error) {
	var err = error(nil)
	var hwNPUPlugin HwNPUSchedulerPlugin

	for _, myNPUPlugin := range hwNPU.HuaweiNPUs {
		if myNPUPlugin == nil {
			continue
		}

		if err = myNPUPlugin.IsMyNode(node); err != nil {
			continue
		}
		hwNPUPlugin = myNPUPlugin
		err = nil
		break
	}

	return hwNPUPlugin, err
}

func (hwNPU *ScheduleHandler) getHWPluginByNodes(nodes []*api.NodeInfo) (HwNPUSchedulerPlugin, error) {
	var err = error(nil)
	var hwNPUPlugin HwNPUSchedulerPlugin

	for _, myNPUPlugin := range hwNPU.HuaweiNPUs {
		if myNPUPlugin == nil {
			continue
		}

		err = errors.New("nil nodes assert")
		for _, node := range nodes {
			if reflect.ValueOf(node).IsNil() {
				continue
			}

			if err = myNPUPlugin.IsMyNode(node); err == nil && myNPUPlugin != nil {
				hwNPUPlugin = myNPUPlugin
				break
			}
		}
		if err == nil {
			break
		}
	}

	return hwNPUPlugin, err
}

func (hwNPU *ScheduleHandler) getNPUPlugin(obj interface{}) HwNPUSchedulerPlugin {
	var err = error(nil)
	var hwNPUPlugin HwNPUSchedulerPlugin

	switch para := obj.(type) {
	case *api.TaskInfo:
		hwNPUPlugin, err = hwNPU.getHWPluginByTask(para)
		if err == nil {
			break
		}
	case *api.JobInfo:
		hwNPUPlugin, err = hwNPU.getHWPluginByJob(para)
		if err == nil {
			break
		}
	case *api.NodeInfo:
		hwNPUPlugin, err = hwNPU.getHWPluginByNode(para)
		if err == nil {
			break
		}
	case []*api.NodeInfo:
		hwNPUPlugin, err = hwNPU.getHWPluginByNodes(para)
		if err == nil {
			break
		}
	default:
		klog.V(logErrorLev).Infof("unknown npu scheduler by %T.", para)
		return nil
	}

	if err == nil {
		return hwNPUPlugin
	}

	klog.V(logDebugLev).Infof("%T : %v.", obj, err)
	return nil
}
