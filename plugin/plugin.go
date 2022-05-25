/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package plugin is using for HuaWei Ascend pin affinity schedule frame.

*/
package plugin

import (
	"errors"
	"reflect"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/conf"
	"volcano.sh/volcano/pkg/scheduler/framework"

	npuapi "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/api"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/rescheduling"
)

// NPUBuilder PluginBuilder plugin management
type NPUBuilder = func(string2 string) HwNPUSchedulerPlugin

// HwNPUSchedulerPlugin Define the npu scheduler policy which new npu scheduler need to realize.
type HwNPUSchedulerPlugin interface {
	// Name The unique name of npu scheduler policy, used by volcano frame.
	Name() string
	// GetResourceName get plugin NPU resource name.
	GetResourceName() string
	// GetResourcePreVal get plugin NPU resource name prefix.
	GetResourcePreVal() string
	// GetPluginDefaultJobSchedulerConfig get plugin default job scheduler config.
	GetPluginDefaultJobSchedulerConfig() map[string]string
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
	// PreCheckNodeFn Check whether the node matches the tag requirements of the task.
	PreCheckNodeFn(*api.TaskInfo, *api.NodeInfo, []conf.Configuration) error
	// CheckNPUResourceStableFn Check whether the resources on the node are stable.
	CheckNPUResourceStableFn(*api.NodeInfo) error
	// CheckNodeNPUByTaskFn Check whether the requested resource exists and are sufficient on the node.
	CheckNodeNPUByTaskFn(*api.TaskInfo, *api.NodeInfo, bool) error
	// GetNPUAffinityBestNodesFn Get all the nodes,witch meet the npu affinity condition.
	GetNPUAffinityBestNodesFn(*api.TaskInfo, []*api.NodeInfo, bool) (map[string]int, error)
	// ScoreBestNPUNodesFn Score the nodes.
	ScoreBestNPUNodesFn(map[string]float64, map[string]int, *api.TaskInfo, []*api.NodeInfo) (map[string]float64, error)
	// GetAllocatedNPUFromTopologyFn Get the pod's npu card to record in node others.
	GetAllocatedNPUFromTopologyFn(*api.TaskInfo, *api.NodeInfo, bool) (interface{}, error)
	// UpdateNPUNodeUsedCardFn Update node others.
	UpdateNPUNodeUsedCardFn(*api.NodeInfo, interface{}) error
	// SetNPUTopologyToPodFn Set the npu cared id into pod.
	SetNPUTopologyToPodFn(*api.TaskInfo, interface{}) error
	// UpdateReleaseNPUNodeTopologyFn Update the node using npu when release pod's npu.
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
	// Handle NPU fault chip/node functions.
	PreHandleFaultNPUFns map[string]npuapi.PreHandleFaultNPUFn
	// Nodes pre-select cluster processing
	ClusterNodePredicateFns map[string]npuapi.ClusterNodePredicateFn
	//  preHandleVNPUFn
	PreHandleVNPUFns map[string]npuapi.PreHandleVNPUFn
	// VJobRunHandleFn
	VJobRunHandleFns map[string]npuapi.VNPUJobRunningHandleFn
}

// AddPreHandleVNPU Add preHandle VNPU function at adapter pattern.
func (hwNPU *ScheduleHandler) AddPreHandleVNPU(pluginName string, fn npuapi.PreHandleVNPUFn) {
	hwNPU.PreHandleVNPUFns[pluginName] = fn
	klog.V(logDebugLev).Infof("PreHandleVNPUFns :%v add.", pluginName)
}

// AddVJobRunHandle Add vJob run handle function at adapter pattern.
func (hwNPU *ScheduleHandler) AddVJobRunHandle(pluginName string, fn npuapi.VNPUJobRunningHandleFn) {
	hwNPU.VJobRunHandleFns[pluginName] = fn
	klog.V(logDebugLev).Infof("VJobRunHandle :%v add.", pluginName)
}

// AddPreHandleFaultNPU add Pretreatment of NPU faults function
func (hwNPU *ScheduleHandler) AddPreHandleFaultNPU(pluginName string, fn npuapi.PreHandleFaultNPUFn) {
	hwNPU.PreHandleFaultNPUFns[pluginName] = fn
	klog.V(logDebugLev).Infof("PreHandleFaultNPUFns :%v add.", pluginName)
}

// AddInitNodesNPUAllocTopology add node NPU assignable topology chip function
func (hwNPU *ScheduleHandler) AddInitNodesNPUAllocTopology(name string, fn npuapi.InitNodesNPUTopologyFn) {
	hwNPU.InitNodesNPUAllocTopologyFns[name] = fn
	klog.V(logDebugLev).Infof("InitNodesNPUAllocTopology :%v add.", name)
}

// AddClusterNodePredicateFn add pre-select cluster processing function.
func (hwNPU *ScheduleHandler) AddClusterNodePredicateFn(pluginName string, fn npuapi.ClusterNodePredicateFn) {
	hwNPU.ClusterNodePredicateFns[pluginName] = fn
	klog.V(logDebugLev).Infof("ClusterNodePredicate :%v add.", pluginName)
}

// RegisterNPUScheduler register the plugin
func (hwNPU *ScheduleHandler) RegisterNPUScheduler(name string, pc NPUBuilder) {
	if _, ok := hwNPU.HuaweiNPUs[name]; ok {
		klog.V(logInfoLev).Infof("NPU Scheduler[%v] has been registered before.", name)
		return
	}

	temp := pc(name)
	hwNPU.HuaweiNPUs[name] = temp
	klog.V(logInfoLev).Infof("NPU Scheduler[%v] registered.", name)
	// for npu scheduler start.
	temp.OnHandlerStart(hwNPU)
}

// UnRegisterNPUScheduler unRegister the plugin
func (hwNPU *ScheduleHandler) UnRegisterNPUScheduler(name string) error {
	if hwNPU == nil {
		return errors.New("nil parameters")
	}
	if _, ok := hwNPU.HuaweiNPUs[name]; ok {
		hwNPU.HuaweiNPUs[name] = nil
		delete(hwNPU.HuaweiNPUs, name)
		klog.V(logErrorLev).Infof("NPU Scheduler[%v] delete.", name)
	}
	klog.V(logDebugLev).Infof("NPU Scheduler[%v] unRegistered.", name)
	return nil
}

// IsPluginRegistered Determine if the plug-in is registered.
func (hwNPU *ScheduleHandler) IsPluginRegistered(name string) bool {
	if hwNPU == nil {
		klog.V(logErrorLev).Infof("IsPluginRegistered nil parameters.")
		return false
	}
	_, ok := hwNPU.HuaweiNPUs[name]
	return ok
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
	// Handle NPU fault chip/node/job.
	if err := hwNPU.initHandleFaultNPUInf(ssn); err != nil {
		klog.V(logDebugLev).Infof("init handle fault NPU :%v .", err)
	}
	// Handle VNPU 910,710
	if err := hwNPU.preHandleVNPUFn(ssn); err != nil {
		klog.V(logDebugLev).Infof("preprocess virtual NPU :%v.", err)
	}
}

func (hwNPU *ScheduleHandler) getHWPluginByTask(task *api.TaskInfo) (HwNPUSchedulerPlugin, error) {
	var err = errors.New("nil plugin")
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
		err = errors.New("nil nodes assert")
		for _, node := range nodes {
			if reflect.ValueOf(node).IsNil() || myNPUPlugin == nil {
				continue
			}
			if err = myNPUPlugin.IsMyNode(node); err == nil {
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

// BeforeCloseHandler  the function handler before session close.
func (hwNPU *ScheduleHandler) BeforeCloseHandler(ssn *framework.Session) {
	// deal fault ReScheduler
	if err := rescheduling.WriteReSchedulerDataToCM(ssn, rescheduling.ReSchedulerCache); err != nil {
		klog.V(logInfoLev).Infof("%s BeforeCloseHandler %v.", PluginName, err)
	}
	// deal VNPU Jobs
	if vJobErr := hwNPU.vJobRunHandleFn(ssn); vJobErr != nil {
		klog.V(logInfoLev).Infof("%s BeforeCloseHandler %v.", PluginName, vJobErr)
	}
	return
}
