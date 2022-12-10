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
	if len(pNames) > 1 {
		// vnpu support
		pNames[0] = pNames[0] + "-"
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
	if sHandle == nil || ssn == nil {
		klog.V(util.LogInfoLev).Infof("%s nil session hence doing nothing.", PluginName)
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
		confs, ok := vf.Conf[0].Arguments[k]
		if ok {
			vf.Conf[0].Arguments[k] = addConf(confs, v)
			continue
		}
		vf.Conf[0].Arguments[k] = v
	}
}

func addConf(configs, value string) string {
	for _, cfg := range strings.Split(value, "|") {
		if !isSelectorContains(configs, cfg) {
			configs += "|" + cfg
		}
	}
	return configs
}

// GetJobTemplate get template of all possible segmentation jobs
func (sHandle *ScheduleHandler) GetJobTemplate() map[string]map[string]util.VResource {
	jobTemplate := map[string]map[string]util.VResource{
		Ascend310P: {
			"vir01":          {Aicore: 1, Aicpu: 1, DVPP: "null"},
			"vir02":          {Aicore: 2, Aicpu: 2, DVPP: "null"},
			"vir02_1c":       {Aicore: 2, Aicpu: 1, DVPP: "null"},
			"vir04":          {Aicore: 4, Aicpu: 4, DVPP: "null"},
			"vir04_3c":       {Aicore: 4, Aicpu: 3, DVPP: "null"},
			"vir04_3c_ndvpp": {Aicore: 4, Aicpu: 3, DVPP: "no"},
			"vir04_4c_dvpp":  {Aicore: 4, Aicpu: 4, DVPP: "yes"},
		},
		Ascend910: {
			"vir02": {Aicore: 2, Aicpu: 1, DVPP: "null"},
			"vir04": {Aicore: 4, Aicpu: 1, DVPP: "null"},
			"vir08": {Aicore: 8, Aicpu: 3, DVPP: "null"},
			"vir16": {Aicore: 16, Aicpu: 7, DVPP: "null"},
		},
	}
	return jobTemplate
}

// InitVolcanoFrameFromSsn init frame parameter from ssn.
func (sHandle *ScheduleHandler) InitVolcanoFrameFromSsn(ssn *framework.Session) {
	if sHandle == nil || ssn == nil {
		klog.V(util.LogInfoLev).Infof("InitVolcanoFrameFromSsn failed: %s.", util.ArgumentError)
		return
	}
	sHandle.FrameAttr = VolcanoFrame{
		UID:          ssn.UID,
		Conf:         ssn.Configurations,
		KubeClient:   ssn.KubeClient(),
		VJobTemplate: sHandle.GetJobTemplate(),
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
			klog.V(util.LogErrorLev).Infof("InitJobsPlugin %s's plugin not register.", vcJob.Name)
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
	if sHandle == nil || ssn == nil {
		klog.V(util.LogInfoLev).Infof("PreStartPlugin failed: %s.", util.ArgumentError)
		return
	}
	for name, plugin := range sHandle.NPUPlugins {
		if err := plugin.PreStartAction(ssn); err != nil {
			if strings.Contains(err.Error(), util.ArgumentError) {
				continue
			}
			klog.V(util.LogErrorLev).Infof("PreStartPlugin %s %s.", name, err)
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
		data, err := util.UpdateConfigmapIncrementally(sHandle.FrameAttr.KubeClient, nameSpace, cmName, data)
		if err != nil {
			klog.V(util.LogInfoLev).Infof("get old %s configmap failed: %v, write new data into cm", spName, err)
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
			if strings.Contains(err.Error(), util.ArgumentError) {
				continue
			}
			klog.V(util.LogErrorLev).Infof("PreStopPlugin %s %#v.", name, err)
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
