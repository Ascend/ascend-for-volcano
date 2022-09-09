/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package plugin is using for HuaWei Ascend pin affinity schedule.

*/
package plugin

import (
	"errors"
	"strings"

	v12 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"k8s.io/kube-scheduler/extender/v1"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/conf"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// RegisterNPUScheduler register the plugin,like factory.
func (sHandle *ScheduleHandler) RegisterNPUScheduler(name string, pc NPUBuilder) {
	if sHandle == nil || pc == nil {
		klog.V(util.LogInfoLev).Infof("RegisterNPUScheduler : %s.", objectNilError)
		return
	}
	if _, ok := sHandle.NPUPlugins[name]; ok {
		klog.V(util.LogInfoLev).Infof("NPU Scheduler[%#v] has been registered before.", name)
		return
	}

	sHandle.NPUPlugins[name] = pc(name)
	klog.V(util.LogInfoLev).Infof("NPU Scheduler[%#v] registered.", name)
}

// UnRegisterNPUScheduler unRegister the plugin
func (sHandle *ScheduleHandler) UnRegisterNPUScheduler(name string) error {
	if sHandle == nil {
		return errors.New(util.ArgumentError)
	}
	if _, ok := sHandle.NPUPlugins[name]; ok {
		sHandle.NPUPlugins[name] = nil
		delete(sHandle.NPUPlugins, name)
		klog.V(util.LogErrorLev).Infof("NPU Scheduler[%#v] delete.", name)
	}
	klog.V(util.LogDebugLev).Infof("NPU Scheduler[%#v] unRegistered.", name)
	return nil
}

// IsPluginRegistered Determine if the plug-in is registered.
func (sHandle *ScheduleHandler) IsPluginRegistered(name string) bool {
	if sHandle == nil {
		klog.V(util.LogErrorLev).Infof("IsPluginRegistered %s", objectNilError)
		return false
	}
	pNames := strings.Split(name, "-")
	if len(pNames) == 0 {
		klog.V(util.LogErrorLev).Infof("IsPluginRegistered %s %#v", name, pNames)
		return false
	}
	for k := range sHandle.NPUPlugins {
		if k == pNames[0] {
			return true
		}
	}
	klog.V(util.LogErrorLev).Infof("IsPluginRegistered %s not in NPUPlugins %+v", name, sHandle.NPUPlugins)
	return false
}

// checkSession check the ssn's parameters
func (sHandle *ScheduleHandler) checkSession(ssn *framework.Session) error {
	if sHandle == nil || ssn == nil || len(ssn.Jobs) == 0 || len(ssn.Nodes) == 0 {
		klog.V(util.LogInfoLev).Infof("%s nil session or no jobs/nodes hence doing nothing.", PluginName)
		return errors.New("nil ssn")
	}
	return nil
}

// InitJobsFromSsn init all jobs in ssn.
func (sHandle *ScheduleHandler) InitJobsFromSsn(ssn *framework.Session) {
	if sHandle == nil {
		klog.V(util.LogInfoLev).Infof("InitJobsFromSsn failed: %s.", util.ArgumentError)
		return
	}
	sHandle.Jobs = make(map[api.JobID]SchedulerJob, util.MapInitNum)
	for jobID, jobInfo := range ssn.Jobs {
		sJob := SchedulerJob{}
		if err := sJob.Init(jobInfo, sHandle); err != nil {
			klog.V(util.LogInfoLev).Infof("%s InitJobsFromSsn failed: %#v.", jobInfo.Name, err)
			continue
		}
		sHandle.Jobs[jobID] = sJob
	}
	return
}

// AddDefaultSchedulerSelectorConfig Merge default and customer custom tags.
func (vf *VolcanoFrame) AddDefaultSchedulerSelectorConfig() {
	if vf == nil {
		klog.V(util.LogInfoLev).Infof("AddDefaultSchedulerSelectorConfig failed: %s.", util.ArgumentError)
		return
	}
	defaultSchedulerConfig := make(map[string]string, util.MapInitNum)
	defaultSchedulerConfig[util.ArchSelector] = util.HuaweiArchArm + "|" + util.HuaweiArchX86
	defaultSchedulerConfig[util.Accelerator] = util.Accelerator910Value + "|" + util.Accelerator310Value
	defaultSchedulerConfig[util.AcceleratorType] = util.CardAcceleratorType + "|" + util.ModuleAcceleratorType +
		"|" + util.ChipAcceleratorType

	if len(vf.Conf) == 0 {
		vf.Conf = []conf.Configuration{{Name: util.CMSelectorKey, Arguments: defaultSchedulerConfig}}
		return
	}
	if len(vf.Conf[0].Arguments) == 0 {
		vf.Conf[0].Arguments = defaultSchedulerConfig
		return
	}

	for k, v := range defaultSchedulerConfig {
		value, ok := vf.Conf[0].Arguments[k]
		if ok {
			vf.Conf[0].Arguments[k] = value + "|" + v
			continue
		}
		vf.Conf[0].Arguments[k] = v
	}
	return
}

// InitVolcanoFrameFromSsn init frame parameter from ssn.
func (sHandle *ScheduleHandler) InitVolcanoFrameFromSsn(ssn *framework.Session) {
	if sHandle == nil || ssn == nil {
		klog.V(util.LogInfoLev).Infof("InitVolcanoFrameFromSsn failed: %s.", util.ArgumentError)
		return
	}
	sHandle.FrameAttr = VolcanoFrame{
		UID:        ssn.UID,
		Conf:       ssn.Configurations,
		KubeClient: ssn.KubeClient(),
	}
	sHandle.FrameAttr.AddDefaultSchedulerSelectorConfig()
}

// InitJobsPlugin init job by plugins.
func (sHandle *ScheduleHandler) InitJobsPlugin() {
	if sHandle == nil {
		klog.V(util.LogInfoLev).Infof("InitJobsPlugin failed: %s.", util.ArgumentError)
		return
	}
	for _, vcJob := range sHandle.Jobs {
		if vcJob.handler == nil {
			klog.V(util.LogErrorLev).Infof("InitJobsPlugin %s's plugin not register.", vcJob.JobName)
			continue
		}
		if err := vcJob.handler.InitMyJobPlugin(vcJob.SchedulerJobAttr, sHandle.ScheduleEnv); err != nil {
			return
		}
	}
}

// InitCache init ScheduleHandler's cache.
func (sHandle *ScheduleHandler) InitCache() {
	if sHandle == nil {
		klog.V(util.LogInfoLev).Infof("InitCache failed: %s.", util.ArgumentError)
		return
	}
	data := make(map[string]map[string]string, util.MapInitNum)
	data[util.RePropertyCacheName] = make(map[string]string, util.MapInitNum)
	data[util.JobRecovery] = make(map[string]string, util.MapInitNum)
	sHandle.Cache = ScheduleCache{Names: make(map[string]string, util.MapInitNum), Namespaces: make(map[string]string,
		util.MapInitNum), Data: data}
}

// PreStartPlugin preStart plugin action.
func (sHandle *ScheduleHandler) PreStartPlugin(ssn *framework.Session) {
	if sHandle == nil {
		klog.V(util.LogInfoLev).Infof("PreStartPlugin failed: %s.", util.ArgumentError)
		return
	}
	for name, plugin := range sHandle.NPUPlugins {
		if err := plugin.PreStartAction(ssn); err != nil {
			klog.V(util.LogErrorLev).Infof("PreStartPlugin %s %#v.", name, err)
		}
	}
}

func (sHandle *ScheduleHandler) saveCacheToCm() {
	for spName, cmName := range sHandle.ScheduleEnv.Cache.Names {
		nameSpace, okSp := sHandle.ScheduleEnv.Cache.Namespaces[spName]
		data, okData := sHandle.ScheduleEnv.Cache.Data[spName]
		if !okSp || !okData {
			klog.V(util.LogErrorLev).Infof("SaveCacheToCm %s no namespace or Data in cache.", spName)
			continue
		}
		var tmpCM = &v12.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cmName,
				Namespace: nameSpace,
			},
			Data: data,
		}
		if err := util.CreateOrUpdateConfigMap(sHandle.FrameAttr.KubeClient, tmpCM, cmName, nameSpace); err != nil {
			klog.V(util.LogErrorLev).Infof("CreateOrUpdateConfigMap : %#v.", err)
			continue
		}
	}
}

// BeforeCloseHandler do the action before ssn close.
func (sHandle *ScheduleHandler) BeforeCloseHandler() {
	if sHandle == nil {
		klog.V(util.LogInfoLev).Infof("BeforeCloseHandler failed: %s.", util.ArgumentError)
		return
	}
	for name, plugin := range sHandle.NPUPlugins {
		if err := plugin.PreStopAction(&sHandle.ScheduleEnv); err != nil {
			klog.V(util.LogErrorLev).Infof("PreStopPlugin %s %#v.", name, err)
			continue
		}
	}
	sHandle.saveCacheToCm()
}

// InitNPUSession init npu plugin and nodes.
func (sHandle *ScheduleHandler) InitNPUSession(ssn *framework.Session) error {
	klog.V(util.LogInfoLev).Infof("enter %s InitNPUSession.", PluginName)
	defer klog.V(util.LogInfoLev).Infof("leave %s InitNPUSession.", PluginName)

	if err := sHandle.checkSession(ssn); err != nil {
		klog.V(util.LogErrorLev).Infof("%s checkSession : %s.", PluginName, err)
		return err
	}
	sHandle.InitVolcanoFrameFromSsn(ssn)
	sHandle.InitNodesFromSsn(ssn)
	sHandle.InitJobsFromSsn(ssn)

	sHandle.InitJobsPlugin()
	sHandle.InitCache()
	sHandle.PreStartPlugin(ssn)
	return nil
}

// GetNPUScheduler get the NPU scheduler by name
func (sHandle *ScheduleHandler) GetNPUScheduler(name string) (ISchedulerPlugin, bool) {
	if sHandle == nil {
		klog.V(util.LogInfoLev).Infof("GetNPUScheduler failed: %s.", util.ArgumentError)
		return nil, false
	}
	pb, found := sHandle.NPUPlugins[name]
	if found && pb != nil {
		return pb, found
	}

	return nil, found
}

// BatchNodeOrderFn Score the selected nodes.
func (sHandle *ScheduleHandler) BatchNodeOrderFn(task *api.TaskInfo, nodes []*api.NodeInfo) (map[string]float64,
	error) {
	klog.V(util.LogInfoLev).Infof("Enter batchNodeOrderFn")
	defer klog.V(util.LogInfoLev).Infof("leaving batchNodeOrderFn")

	if sHandle == nil || task == nil || len(nodes) == 0 {
		klog.V(util.LogErrorLev).Infof("%s batchNodeOrderFn %s.", PluginName, util.ArgumentError)
		return nil, errors.New(util.ArgumentError)
	}

	// init score-map
	var interPodAffinityScore v1.HostPriorityList
	scoreMap := initScoreMap(nodes, interPodAffinityScore)
	vcJob, ok := sHandle.Jobs[task.Job]
	if !ok {
		klog.V(util.LogDebugLev).Infof("BatchNodeOrderFn %s not req npu.", task.Name)
		return scoreMap, nil
	}
	// 2.Get the best node and top by A,B,C,D rules and require numbers.
	errGet := vcJob.handler.ScoreBestNPUNodes(task, nodes, scoreMap)
	if errGet != nil {
		// get suitable node failed
		klog.V(util.LogErrorLev).Infof("batchNodeOrderFn task[%s] failed[%#v].", task.Name, errGet)
		return scoreMap, nil
	}
	klog.V(util.LogInfoLev).Infof("batchNodeOrderFn Get %s for NPU %+v.", task.Name, scoreMap)

	return scoreMap, nil
}