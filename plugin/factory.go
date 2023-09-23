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

/*
Package plugin is using for HuaWei Ascend pin affinity schedule.
*/
package plugin

import (
	"errors"
	"reflect"
	"strings"

	"gopkg.in/yaml.v2"
	v12 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"k8s.io/kube-scheduler/extender/v1"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/conf"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/config"
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

	sHandle.NPUPlugins[name] = pc
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

	defaultCfg := config.Configuration{Name: util.CMSelectorKey, Arguments: defaultSchedulerConfig}

	if len(vf.Confs) == 0 {
		vf.Confs = []config.Configuration{defaultCfg}
		return
	}

	var selectorConf config.Configuration
	var index int
	for idx, selectors := range vf.Confs {
		if selectors.Name == util.CMSelectorKey {
			selectorConf = selectors
			index = idx
			break
		}
	}

	if len(selectorConf.Arguments) == 0 {
		vf.Confs[index] = defaultCfg
		return
	}

	for k, v := range defaultSchedulerConfig {
		confs, ok := selectorConf.Arguments[k]
		if ok {
			selectorConf.Arguments[k] = addConf(confs, v)
			continue
		}
		selectorConf.Arguments[k] = v
	}
	vf.Confs[index] = selectorConf
}

// CheckVNPUSegmentEnableByConfig Check VNPU segmentEnable by init plugin parameters, return true if static
func (vf *VolcanoFrame) CheckVNPUSegmentEnableByConfig() bool {
	configuration, err := util.GetConfigFromSchedulerConfigMap(util.CMInitParamKey, vf.Confs)
	if err != nil {
		klog.V(util.LogDebugLev).Info("cannot get configuration, segmentEnable.")
		return false
	}
	// get segmentEnable by user configuration
	segmentEnable, ok := configuration.Arguments[util.SegmentEnable]
	if !ok {
		klog.V(util.LogDebugLev).Info("checkVNPUSegmentEnable doesn't exist presetVirtualDevice.")
		return false
	}
	if segmentEnable == "true" {
		return true
	}
	return false
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
			VNPUTempVir01:        {Aicore: 1, Aicpu: 1, DVPP: AscendDVPPEnabledNull},
			VNPUTempVir02:        {Aicore: util.NPUIndex2, Aicpu: util.NPUIndex2, DVPP: AscendDVPPEnabledNull},
			VNPUTempVir02C1:      {Aicore: util.NPUIndex2, Aicpu: 1, DVPP: AscendDVPPEnabledNull},
			VNPUTempVir04:        {Aicore: util.NPUIndex4, Aicpu: util.NPUIndex4, DVPP: AscendDVPPEnabledNull},
			VNPUTempVir04C3:      {Aicore: util.NPUIndex4, Aicpu: util.NPUIndex3, DVPP: AscendDVPPEnabledNull},
			VNPUTempVir04C3NDVPP: {Aicore: util.NPUIndex4, Aicpu: util.NPUIndex3, DVPP: AscendDVPPEnabledOff},
			VNPUTempVir04C4cDVPP: {Aicore: util.NPUIndex4, Aicpu: util.NPUIndex4, DVPP: AscendDVPPEnabledOn},
		},
		Ascend910: {
			VNPUTempVir02: {Aicore: util.NPUIndex2, Aicpu: 1, DVPP: AscendDVPPEnabledNull},
			VNPUTempVir04: {Aicore: util.NPUIndex4, Aicpu: 1, DVPP: AscendDVPPEnabledNull},
			VNPUTempVir08: {Aicore: util.NPUIndex8, Aicpu: util.NPUIndex3, DVPP: AscendDVPPEnabledNull},
			VNPUTempVir16: {Aicore: util.NPUIndex16, Aicpu: util.NPUIndex7, DVPP: AscendDVPPEnabledNull},
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
		Confs:        initConfsFromSsn(ssn.Configurations),
		KubeClient:   ssn.KubeClient(),
		VJobTemplate: sHandle.GetJobTemplate(),
	}

	sHandle.FrameAttr.AddDefaultSchedulerSelectorConfig()
}

func initConfsFromSsn(confs []conf.Configuration) []config.Configuration {
	var out []byte
	var err error
	newConfs := make([]config.Configuration, len(confs))
	for idx, cfg := range confs {
		newCfg := &config.Configuration{}
		out, err = yaml.Marshal(cfg)
		if err != nil {
			klog.V(util.LogInfoLev).Infof("Marshal configuration failed: %s.", err)
			continue
		}
		if err = yaml.Unmarshal(out, newCfg); err != nil {
			klog.V(util.LogInfoLev).Infof("Unmarshal configuration failed: %s.", err)
			continue
		}
		newConfs[idx] = *newCfg
	}
	return newConfs
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
	sHandle.Cache = ScheduleCache{
		Names:           make(map[string]string, util.MapInitNum),
		Namespaces:      make(map[string]string, util.MapInitNum),
		FaultConfigMaps: map[api.JobID]*FaultRankIdData{},
		Data:            data}
}

// PreStartPlugin preStart plugin action.
func (sHandle *ScheduleHandler) PreStartPlugin(ssn *framework.Session) {
	if sHandle == nil || ssn == nil {
		klog.V(util.LogInfoLev).Infof("PreStartPlugin failed: %s.", util.ArgumentError)
		return
	}
	for _, job := range sHandle.Jobs {
		if err := job.handler.PreStartAction(ssn); err != nil {
			if strings.Contains(err.Error(), util.ArgumentError) {
				continue
			}
			klog.V(util.LogErrorLev).Infof("PreStartPlugin %s %s.", job.Name, err)
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
		}
	}

	for _, faultConfig := range sHandle.ScheduleEnv.Cache.FaultConfigMaps {
		data, err := util.UpdateConfigmapIncrementally(sHandle.FrameAttr.KubeClient, faultConfig.Namespace,
			faultConfig.Name, faultConfig.Data)
		if err != nil {
			klog.V(util.LogInfoLev).Infof("get old %s configmap failed: %v", faultConfig.Name, err)
			continue
		}
		var tmpCM = &v12.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      faultConfig.Name,
				Namespace: faultConfig.Namespace,
			},
			Data: data,
		}
		if err = util.CreateOrUpdateConfigMap(sHandle.FrameAttr.KubeClient, tmpCM, faultConfig.Name,
			faultConfig.Namespace); err != nil {
			klog.V(util.LogErrorLev).Infof("CreateOrUpdateConfigMap : %#v.", err)
		}
	}
}

// BeforeCloseHandler do the action before ssn close.
func (sHandle *ScheduleHandler) BeforeCloseHandler(ssn *framework.Session) {
	if sHandle == nil {
		klog.V(util.LogInfoLev).Infof("BeforeCloseHandler failed: %s.", util.ArgumentError)
		return
	}
	for _, job := range sHandle.Jobs {
		if err := job.handler.PreStopAction(&sHandle.ScheduleEnv); err != nil {
			if strings.Contains(err.Error(), util.ArgumentError) {
				continue
			}
			klog.V(util.LogErrorLev).Infof("PreStopPlugin %s %#v.", job.Name, err)
		}
	}
	sHandle.saveCacheToCm()
}

// InitNPUSession init npu plugin and nodes.
func (sHandle *ScheduleHandler) InitNPUSession(ssn *framework.Session) error {
	klog.V(util.LogDebugLev).Infof("enter %s InitNPUSession.", PluginName)
	defer klog.V(util.LogDebugLev).Infof("leave %s InitNPUSession.", PluginName)

	if err := sHandle.checkSession(ssn); err != nil {
		klog.V(util.LogErrorLev).Infof("%s checkSession : %s.", PluginName, err)
		return err
	}

	sHandle.InitVolcanoFrameFromSsn(ssn)
	sHandle.InitNodesFromSsn(ssn)
	sHandle.InitJobsFromSsn(ssn)

	sHandle.InitTorNodeInfo(ssn)
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
		return pb(name), found
	}

	return nil, found
}

// BatchNodeOrderFn Score the selected nodes.
func (sHandle *ScheduleHandler) BatchNodeOrderFn(task *api.TaskInfo, nodes []*api.NodeInfo) (map[string]float64, error) {
	klog.V(util.LogInfoLev).Infof("Enter batchNodeOrderFn")
	defer klog.V(util.LogInfoLev).Infof("leaving batchNodeOrderFn")

	if sHandle == nil || task == nil || len(nodes) == 0 {
		klog.V(util.LogErrorLev).Infof("%s batchNodeOrderFn %s.", PluginName, util.ArgumentError)
		return nil, errors.New(util.ArgumentError)
	}
	if !IsNPUTask(task) {
		return nil, nil
	}
	// init score-map
	var interPodAffinityScore v1.HostPriorityList
	scoreMap := initScoreMap(nodes, interPodAffinityScore)
	vcJob, ok := sHandle.Jobs[task.Job]
	if !ok {
		klog.V(util.LogDebugLev).Infof("BatchNodeOrderFn %s not req npu.", task.Name)
		return scoreMap, nil
	}

	k, ok := vcJob.Label[TorAffinityKey]
	if ok && (k == LargeModelTag || k == NormalSchema) {
		klog.V(util.LogInfoLev).Infof("validNPUJob job is not use tor affinity")
		return sHandle.SetTorAffinityJobNodesScore(task, nodes, vcJob, k, scoreMap)
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

func (sHandle *ScheduleHandler) SetTorAffinityJobNodesScore(task *api.TaskInfo, nodes []*api.NodeInfo,
	vcJob SchedulerJob, label string, scoreMap map[string]float64) (map[string]float64, error) {
	if sHandle == nil || task == nil || len(nodes) == 0 || len(scoreMap) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("ScoreBestNPUNodes %s.", err)
		return scoreMap, err
	}
	result := CheckNetSliceIsMeetJobRequire(vcJob, sHandle, nodes)
	vcJob = sHandle.Jobs[task.Job]
	if result != nil {
		klog.V(util.LogErrorLev).Infof("check job %s tor affinity failed: %s,"+
			"used servers is %s", vcJob.Name, result, vcJob.SelectServers)
		switch label {
		case LargeModelTag:
			vcJob.JobReadyTag = false
		case NormalSchema:
			vcJob.SetNormalJobServerList(sHandle)
		default:
			return scoreMap, nil
		}
		sHandle.Jobs[task.Job] = vcJob
	}
	errGet := sHandle.ScoreBestNPUNodes(task, nodes, scoreMap)
	if errGet != nil {
		// get suitable node failed
		klog.V(util.LogErrorLev).Infof("batchNodeOrderFn task[%s] failed[%#v].", task.Name, errGet)
	}
	klog.V(util.LogInfoLev).Infof("batchNodeOrderFn Get %s for NPU %+v.", task.Name, scoreMap)
	return scoreMap, result
}

func (sHandle *ScheduleHandler) ScoreBestNPUNodes(task *api.TaskInfo, nodes []*api.NodeInfo, sMap map[string]float64) error {
	if sHandle == nil || task == nil || len(nodes) == 0 || len(sMap) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("ScoreBestNPUNodes %s.", err)
		return err
	}
	vcjob, ok := sHandle.ScheduleEnv.Jobs[task.Job]
	if !ok {
		return errors.New(util.ArgumentError)
	}
	vcjob.ServerList = vcjob.SortJobServerListBySliceId()
	for nodeName, index := range vcjob.HealthTorRankIndex {
		if index == task.Pod.Annotations[podRankIndex] {
			sMap[nodeName] = maxTorAffinityNodeScore
			return nil
		}
	}
	vcjob.SetJobRankIndex()
	for _, sl := range vcjob.ServerList {
		if reflect.ValueOf(sl).IsNil() {
			continue
		}
		for _, server := range sl.Servers {
			if reflect.ValueOf(server).IsNil() {
				continue
			}
			for _, node := range nodes {
				if server.Name != node.Name || server.NodeRank != task.Pod.Annotations[podRankIndex] {
					continue
				}
				sMap[server.Name] = maxTorAffinityNodeScore
				break
			}
		}
	}

	klog.V(util.LogInfoLev).Infof("ScoreBestNPUNodes task<%s> sMap<%v>", task.Name, sMap)
	return nil
}
