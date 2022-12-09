/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package plugin is using for HuaWei Ascend pin affinity schedule frame.
*/
package plugin

import (
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// NPUBuilder PluginBuilder plugin management
type NPUBuilder = func(string2 string) ISchedulerPlugin

// ISchedulerPluginBase the frame plugin need implement.
type ISchedulerPluginBase interface {
	GetPluginName() string
	SetPluginName(string)
	GetAnnoPreVal() string
	SetAnnoPreVal(string)
	GetAnnoName() string
	SetAnnoName(string)
	GetDefaultJobSchedulerConfig() map[string]string
	SetDefaultJobSchedulerConfig(map[string]string)
}

// ISchedulerPluginNeed The interface that the specific plug-in needs to implement.
type ISchedulerPluginNeed interface {
	// ValidNPUJob Valid the job part of npu scheduler policy, if not, disallowed.
	ValidNPUJob() *api.ValidateResult
	CheckNodeNPUByTask(*api.TaskInfo, NPUNode) error
	ScoreBestNPUNodes(*api.TaskInfo, []*api.NodeInfo, map[string]float64) error
	UseAnnotation(*api.TaskInfo, NPUNode) *NPUNode
	PreStartAction(ssn *framework.Session) error
	PreStopAction(*ScheduleEnv) error
	InitMyJobPlugin(util.SchedulerJobAttr, ScheduleEnv) error
}

// ISchedulerPlugin for volcano-npu plugin has function.
type ISchedulerPlugin interface {
	ISchedulerPluginBase
	ISchedulerPluginNeed
}

// SchedulerPlugin for all volcano-npu plugin.
type SchedulerPlugin struct {
	// the new func add name
	pluginName string
	// in k8s annotation huawei.com/Ascend310,huawei.com/Ascend910
	annoName string
	// huawei.com/
	annoPreVal string
	// config like arm x86
	defaultJobSchedulerConfig map[string]string
}

// GetPluginName get PluginName.
func (sp SchedulerPlugin) GetPluginName() string {
	return sp.pluginName
}

// SetPluginName set PluginName.
func (sp *SchedulerPlugin) SetPluginName(name string) {
	if sp == nil {
		klog.V(util.LogInfoLev).Infof("SetPluginName failed: %s.", util.ArgumentError)
		return
	}
	sp.pluginName = name
}

// GetAnnoPreVal get AnnoPreVal.
func (sp SchedulerPlugin) GetAnnoPreVal() string {
	return sp.annoPreVal
}

// SetAnnoPreVal set AnnoPreVal.
func (sp *SchedulerPlugin) SetAnnoPreVal(value string) {
	if sp == nil {
		klog.V(util.LogInfoLev).Infof("SetAnnoPreVal failed: %s.", util.ArgumentError)
		return
	}
	sp.annoPreVal = value
}

// GetAnnoName get AnnoName.
func (sp SchedulerPlugin) GetAnnoName() string {
	return sp.annoName
}

// SetAnnoName set AnnoName.
func (sp *SchedulerPlugin) SetAnnoName(annoName string) {
	if sp == nil {
		klog.V(util.LogInfoLev).Infof("SetAnnoName failed: %s.", util.ArgumentError)
		return
	}
	sp.annoName = annoName
}

// GetDefaultJobSchedulerConfig get DefaultJobSchedulerConfig.
func (sp SchedulerPlugin) GetDefaultJobSchedulerConfig() map[string]string {
	return sp.defaultJobSchedulerConfig
}

// SetDefaultJobSchedulerConfig set DefaultJobSchedulerConfig.
func (sp *SchedulerPlugin) SetDefaultJobSchedulerConfig(conf map[string]string) {
	if sp == nil {
		klog.V(util.LogInfoLev).Infof("SetDefaultJobSchedulerConfig failed: %s.", util.ArgumentError)
		return
	}

	if len(sp.defaultJobSchedulerConfig) == 0 {
		sp.defaultJobSchedulerConfig = make(map[string]string, util.MapInitNum)
	}
	sp.defaultJobSchedulerConfig = conf
}
