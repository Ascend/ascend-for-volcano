/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package comvnpu is using for virtual HuaWei Ascend910 schedule.

*/
package comvnpu

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/conf"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/vnpu/modulev710"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/vnpu/modulev910"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/vnpu/vnpuutil"
)

// Name Get plugin name for frame init
func (tp *VNPU) Name() string {
	if tp == nil {
		return vnpuutil.PluginName
	}
	return tp.Attr.PluginName
}

// New returns a virtual npu plugin
func New(npuName string) plugin.HwNPUSchedulerPlugin {
	var npuPlugin = VNPU{}
	npuPlugin.Attr.PluginName = npuName
	return &npuPlugin
}

// OnHandlerStart Vnpu scheduler policy initial and common processing
func (tp *VNPU) OnHandlerStart(sHandler *plugin.ScheduleHandler) {
	klog.V(util.LogInfoLev).Infof("%v start Handler.", tp.Name())
	sHandler.AddInitNodesNPUAllocTopology(tp.Name(), tp.InitVNodesFn)
	sHandler.AddPreHandleVNPU(tp.Attr.AnnoName, tp.PreHandleVNPU)
	sHandler.AddVJobRunHandle(tp.Attr.AnnoName, tp.VJobRunHandle)
}

// UpdateRunningVNPJobIntoCache update running job into cache.
func (tp *VNPU) UpdateRunningVNPJobIntoCache(ssn *framework.Session) error {
	for jobID, jobInf := range ssn.Jobs {
		_, getERR := tp.GetVNPUAllocInfFromCacheByJobID(jobID)
		if getERR != nil {
			klog.V(util.LogErrorLev).Infof("%s %s DealVNPUSelectNodeAndChip %v.", tp.Name(), jobID, getERR)
			continue
		}
		// for num is one.
		if !vnpuutil.IsVJobRunning(jobInf) {
			continue
		}
		if updateErr := tp.UpdateVJobsCacheAllocChipByJobName(jobInf); updateErr != nil {
			continue
		}
	}
	return nil
}

// VJobRunHandle deal running vJobs.
func (tp *VNPU) VJobRunHandle(ssn *framework.Session) error {
	// 1.update running job into cache
	updateErr := tp.UpdateRunningVNPJobIntoCache(ssn)
	if updateErr != nil {
		klog.V(util.LogInfoLev).Infof("%v UpdateRunningVNPJobIntoCache %v.", tp.Name(), updateErr)
	}
	// 2.update cm
	cmErr := tp.writevNPUAllocInfIntoCm(ssn)
	if cmErr != nil {
		klog.V(util.LogInfoLev).Infof("%v writevNPUAllocInfIntoCm %v.", tp.Name(), cmErr)
	}
	// 3.Curing cache
	writeErr := vnpuutil.WriteVNPUAllocInfDataIntoCacheCM(ssn)
	if writeErr != nil {
		klog.V(util.LogInfoLev).Infof("%v WriteVNPUAllocInfDataIntoCacheCM %v.", tp.Name(), writeErr)
	}
	if updateErr == nil && cmErr == nil && writeErr == nil {
		return nil
	}
	return multierror.Append(updateErr, cmErr, writeErr)
}

// GetVNPUCacheFromCacheCM syn cache and cache cm.
func (tp *VNPU) GetVNPUCacheFromCacheCM(ssn *framework.Session) error {
	cache, getErr := vnpuutil.GetVNPUAllocInfDataFromCacheCM(ssn)
	if getErr != nil {
		klog.V(util.LogDebugLev).Infof("GetVNPUAllocInfDataFromCacheCM :%v.", getErr)
		return getErr
	}
	vnpuutil.VNPUAllocData = *cache
	klog.V(util.LogDebugLev).Infof("GetVNPUAllocInfDataFromCacheCM get cache from cm successes.")
	return nil
}

// PreHandleVNPU Only for abstract VNPU, not v910,v710 and so on.
func (tp *VNPU) PreHandleVNPU(ssn *framework.Session) error {
	if getErr := tp.GetVNPUCacheFromCacheCM(ssn); getErr != nil {
		klog.V(util.LogErrorLev).Infof("PreHandleVNPU :%v.", getErr)
	}

	if updateErr := tp.updateNodesOthersByVNPUCache(ssn.Nodes); updateErr != nil {
		klog.V(util.LogErrorLev).Infof("PreHandleVNPU :%v.", updateErr)
	}

	if err := tp.DealNewVNPUJob(ssn); err != nil {
		klog.V(util.LogInfoLev).Infof("PreHandleVNPU :%v.", err)
	}

	if npuErr := tp.DealFinishedVNPUJob(ssn); npuErr != nil {
		klog.V(util.LogInfoLev).Infof("PreHandleVNPU :%v.", npuErr)
	}
	// no need deal errors.
	return nil
}

// IsVNPUReqMeetActual Check whether the partition of node chips is consistent with the requirements in cache
func (tp *VNPU) IsVNPUReqMeetActual(reqCard, actCard string) bool {
	reqSlice := strings.Split(reqCard, "-")
	req := reqSlice[len(reqSlice)-1]
	actualTop := strings.Split(actCard, ",")
	for _, actCardString := range actualTop {
		actualSlice := strings.Split(actCardString, "-")
		actual := actualSlice[len(actualSlice)-1]
		if req == actual {
			return true
		}
	}

	return false
}

// DealVNPUSelectNodeAndChip check node whether has job require resource.
func (tp *VNPU) DealVNPUSelectNodeAndChip(task *api.TaskInfo, node *api.NodeInfo) error {
	data, getERR := tp.GetVNPUAllocInfFromCacheByJobID(task.Job)
	if getERR != nil {
		klog.V(util.LogErrorLev).Infof("%s %s DealVNPUSelectNodeAndChip %v.", tp.Name(), task.Name, getERR)
		return getERR
	}
	// check the node is record
	if data.NodeName != node.Name {
		nameErr := fmt.Errorf("select %s not %s", data.NodeName, node.Name)
		klog.V(util.LogErrorLev).Infof("%s %s DealVNPUSelectNodeAndChip %v.", tp.Name(), task.Name, nameErr)
		return nameErr
	}
	// check whether has the require chip
	value, covertErr := util.GetNPUAllocCardsFromNodeOthers(node, data.ReqNPUType)
	if covertErr != nil {
		klog.V(util.LogErrorLev).Infof("%s DealVNPUSelectNodeAndChip %v %v.", tp.Name(), task.Name, covertErr)
		return covertErr
	}
	// Check whether the partition of node chips is consistent with the requirements in cache
	if !tp.IsVNPUReqMeetActual(data.ReqCardName, value) {
		meetErr := fmt.Errorf("req %s,actual %s", data.ReqCardName, value)
		klog.V(util.LogErrorLev).Infof("%s %s DealVNPUSelectNodeAndChip %v.", tp.Name(), task.Name, meetErr)
		return meetErr
	}

	return nil
}

// PreCheckNodeFn check whether the node matches the tag requirements of the task.
func (tp *VNPU) PreCheckNodeFn(task *api.TaskInfo, node *api.NodeInfo, _ []conf.Configuration) error {
	klog.V(util.LogDebugLev).Infof("%s PreCheckNodeFn %s enter.", tp.Name(), task.Name)
	defer klog.V(util.LogDebugLev).Infof("%s PreCheckNodeFn %s leave.", tp.Name(), task.Name)
	// 1.init vnp
	pluginName, nameErr := tp.GetPluginNameByTaskInfo(task)
	if nameErr != nil {
		klog.V(util.LogErrorLev).Infof("%s CheckNodeNPUByTaskFn %s %v.", tp.Name(), task.Name, nameErr)
		return nameErr
	}
	if pluginErr := tp.InitVNPUPluginByType(pluginName); pluginErr != nil {
		klog.V(util.LogErrorLev).Infof("%s CheckNodeNPUByTaskFn :%v.", vnpuutil.PluginName, pluginErr)
		return pluginErr
	}

	// select node by architect
	if err := tp.IsSelectorMeetNode(task, node, tp.Attr.DefaultJobSchedulerConfig); err != nil {
		// get scheduler selector configure failed, but need continue
		klog.V(util.LogErrorLev).Infof("%s %s %s : %v.", tp.Name(), task.Name, node.Name, err)
		return fmt.Errorf("task(%s) in node(%s):%v", task.Name, node.Name, err)
	}

	if vNPUErr := tp.DealVNPUSelectNodeAndChip(task, node); vNPUErr != nil {
		klog.V(util.LogErrorLev).Infof("%s %s %s : %v.", tp.Name(), task.Name, node.Name, vNPUErr)
		return vNPUErr
	}
	return nil
}

// CheckNodeNPUByTaskFn check whether the requested resource exists on the node.The cored has been split.
func (tp *VNPU) CheckNodeNPUByTaskFn(vTask *api.TaskInfo, node *api.NodeInfo, _ bool) error {
	// has been done in pre-check.
	klog.V(util.LogErrorLev).Infof("%s CheckNodeNPUByTaskFn %s for %v,no need.", tp.Name(), vTask.Name, node.Name)
	return nil
}

// GetNPUAffinityBestNodesFn initialize a mapping between nodes and priorities
func (tp *VNPU) GetNPUAffinityBestNodesFn(task *api.TaskInfo, nodes []*api.NodeInfo, _ bool) (map[string]int, error) {
	klog.V(util.LogDebugLev).Infof("Get NPU affinity best node for task %s.", task.Name)
	var bestNodesMap = make(map[string]int, util.ConstIntNum3)

	for _, node := range nodes {
		if node == nil {
			continue
		}
		bestNodesMap[node.Name] = 0
	}

	return bestNodesMap, nil
}

// ScoreTheVJobSelectNode sorce the select node in cache.
func (tp *VNPU) ScoreTheVJobSelectNode(vTask *api.TaskInfo, nodes []*api.NodeInfo, sm map[string]float64) error {
	if vTask == nil || len(nodes) == 0 || len(sm) == 0 {
		return errors.New("error parameter")
	}
	// get vJob select nodeName
	data, getERR := tp.GetVNPUAllocInfFromCacheByJobID(vTask.Job)
	if getERR != nil {
		klog.V(util.LogErrorLev).Infof("%s GetVNPUAllocInfFromCacheByJobID %v.", vTask.Name, getERR)
		return getERR
	}
	nodeName := data.NodeName
	// check nodeName whether in nodes or not
	existFlag := false
	for _, nodeInf := range nodes {
		if nodeInf.Name == nodeName {
			existFlag = true
			break
		}
	}
	if !existFlag {
		existErr := fmt.Errorf("%s req %s not input", vTask.Job, nodeName)
		klog.V(util.LogErrorLev).Infof("ScoreTheVJobSelectNode %v.", existErr)
		return existErr
	}
	// force score the map,
	sm[nodeName] = vnpuutil.VNPUScoreWeight
	return nil
}

// ScoreBestNPUNodesFn used for score candidate nodes
func (tp *VNPU) ScoreBestNPUNodesFn(scoreMap map[string]float64, bestNodes map[string]int, vTask *api.TaskInfo,
	nodes []*api.NodeInfo) (map[string]float64, error) {
	if len(scoreMap) == 0 || reflect.ValueOf(scoreMap).IsNil() {
		err := errors.New("scoreBestNPUNodes's scoreMap is nil")
		klog.V(util.LogInfoLev).Infof("%s %s %v.", nodes, tp.Name(), err)
		return nil, err
	}
	for nodeName, priority := range bestNodes {
		if _, ok := scoreMap[nodeName]; ok {
			scoreMap[nodeName] = float64(priority)
			continue
		}

		scoreMap[nodeName] = 0.0
	}
	// if the vJob select node, give the max score.
	if vJobScoreErr := tp.ScoreTheVJobSelectNode(vTask, nodes, scoreMap); vJobScoreErr == nil {
		klog.V(util.LogInfoLev).Infof("%s ScoreBestNPUNodesFn %v.", tp.Name(), vJobScoreErr)
		return scoreMap, nil
	}
	return scoreMap, nil
}

// GetAllocatedNPUFromTopologyFn obtain the name of the allocated devices, VNPU only has one chip.
func (tp *VNPU) GetAllocatedNPUFromTopologyFn(vTask *api.TaskInfo, node *api.NodeInfo, _ bool) (interface{}, error) {
	var allocVNPUChip []string
	// get the VJob req NPU chip(Ascend910-1) and VNPU type(huawei.com/Ascend910-2c).
	data, getERR := tp.GetVNPUAllocInfFromCacheByJobID(vTask.Job)
	if getERR != nil {
		klog.V(util.LogErrorLev).Infof("%s GetAllocatedNPUFromTopologyFn %v.", vTask.Name, getERR)
		return nil, getERR
	}
	// like: huawei.com/Ascend910-4c
	reqCardName := data.ReqCardName
	// like: huawei.com/Ascend910-0
	reqVNPUType := data.ReqNPUType

	nodeCards, err := util.GetNPUAllocCardsFromNodeOthers(node, reqVNPUType)
	if err != nil {
		covertErr := fmt.Errorf("%s other %s not string", node.Name, reqVNPUType)
		klog.V(util.LogErrorLev).Infof("%s GetAllocatedNPUFromTopologyFn %v.", tp.Name(), covertErr)
		return nil, covertErr
	}
	// find the card id
	reqIDStrings := strings.Split(reqCardName, "-")
	reqIDStr := reqIDStrings[len(reqIDStrings)-1]
	nodeCardSlices := strings.Split(nodeCards, ",")
	for _, chip := range nodeCardSlices {
		// chip like :Ascend910-4c-100-0
		actIDStrings := strings.Split(chip, "-")
		actIDStr := actIDStrings[len(actIDStrings)-1]
		if actIDStr == reqIDStr {
			allocVNPUChip = append(allocVNPUChip, chip)
			break
		}
	}

	if len(allocVNPUChip) == 0 {
		allocErr := fmt.Errorf("%s no meet VNPU in %s", vTask.Job, node.Name)
		klog.V(util.LogErrorLev).Infof("%s GetAllocatedNPUFromTopologyFn %+v %+v.",
			tp.Name(), nodeCards, allocErr)
		return nil, allocErr
	}
	klog.V(util.LogInfoLev).Infof("GetAllocatedNPUFromTopologyFn %s get %v in %s.", vTask.Job, allocVNPUChip,
		node.Name)
	return allocVNPUChip[0], nil
}

// AddOrUpdateVNPUAllocInfIntoCache  add or update cache element.
func (tp *VNPU) AddOrUpdateVNPUAllocInfIntoCache(data *vnpuutil.VNPUAllocInf) error {
	if data == nil {
		return errors.New("nil data")
	}
	flag := false
	for k, cacheData := range vnpuutil.VNPUAllocData.Cache {
		if data.JobUID == cacheData.JobUID {
			vnpuutil.VNPUAllocData.Cache[k] = *data
			flag = true
			break
		}
	}
	if !flag {
		vnpuutil.VNPUAllocData.Cache = append(vnpuutil.VNPUAllocData.Cache, *data)
	}
	vnpuutil.VNPUAllocData.CheckCode = util.MakeDataHash(vnpuutil.VNPUAllocData)
	klog.V(util.LogDebugLev).Infof("%s AddOrUpdateVNPUAllocInfIntoCache %+v.", tp.Name(),
		vnpuutil.VNPUAllocData.Cache)
	return nil
}

// InitVNPUPluginByType init vnpu plugin.Add new vnpu plugin must do.
func (tp *VNPU) InitVNPUPluginByType(reqNpuType string) error {
	switch reqNpuType {
	case vnpuutil.PluginNameBy910VNPU:
		v910x8 := &modulev910.ChipV910{}
		if err := v910x8.InitVNPUPlugin(); err != nil {
			break
		}
		tp.Attr = v910x8.ComVNPU
		var s VNPUHandler = v910x8
		tp.Plugin = s
	case vnpuutil.PluginNameBy710VNPU:
		v710 := &modulev710.ChipV710{}
		if err := v710.InitVNPUPlugin(); err != nil {
			break
		}
		tp.Attr = v710.ComVNPU
		var f VNPUHandler = v710
		tp.Plugin = f
	default:
		supErr := fmt.Errorf("not support plugin type :%s", reqNpuType)
		klog.V(util.LogErrorLev).Infof("%s InitVNPUPluginByType %v.", tp.Name(), supErr)
		return supErr
	}
	return nil
}

// GetAndRecordNewVNPUJobsFromSsn record new vJob in cache.
func (tp *VNPU) GetAndRecordNewVNPUJobsFromSsn(ssn *framework.Session) error {
	getNum := 0
	for _, job := range ssn.Jobs {
		// for not ready job can pre handle
		if !vnpuutil.IsVJobPending(job) {
			klog.V(util.LogDebugLev).Infof("%s GetAndRecordNewVNPUJobsFromSsn %s ready:%v.",
				vnpuutil.PluginName, job.Name, job.PodGroup.Status.Phase)
			continue
		}
		if !tp.IsVNPUJob(job) {
			klog.V(util.LogDebugLev).Infof("%s GetAndRecordNewVNPUJobsFromSsn %s not VNPU job.", tp.Name(),
				job.Name)
			continue
		}
		if err := tp.RecordNewVNPUJobInCache(job); err != nil {
			klog.V(util.LogErrorLev).Infof("%s RecordNewVNPUJobInCache :%v.", vnpuutil.PluginName, err)
			continue
		}
		getNum++
	}
	if getNum == 0 {
		return errors.New("none jobs need deal")
	}
	return nil
}

// GetAllNeedAllocVJobsFromCache get unAlloc vJobs from cache.
func (tp *VNPU) GetAllNeedAllocVJobsFromCache(ssn *framework.Session) ([]*api.JobInfo, error) {
	var vJobs []*api.JobInfo
	for _, inf := range vnpuutil.VNPUAllocData.Cache {
		if inf.AllocFlag {
			continue
		}
		tmpJob, ok := ssn.Jobs[inf.JobUID]
		if !ok {
			err := fmt.Errorf("%s not in ssn", inf.JobUID)
			klog.V(util.LogErrorLev).Infof("%s GetAllNeedAllocVJobsFromCache :%v.", vnpuutil.PluginName, err)
			return nil, err
		}
		vJobs = append(vJobs, tmpJob)
	}
	if len(vJobs) == 0 {
		return nil, errors.New("none vJobs need alloc")
	}
	return vJobs, nil
}

// GetClusterAllResourceFromSsn get cluster all core used.
func (tp *VNPU) GetClusterAllResourceFromSsn(ssn *framework.Session) (map[string]int, error) {
	var allocatableResource = make(map[string]int, util.ConstIntNum8)
	for _, nodInf := range ssn.Nodes {
		nodeCoresInf, coresErr := tp.GetNodeNPUCoreInfoMap(nodInf)
		if coresErr != nil {
			klog.V(util.LogDebugLev).Infof("%s GetClusterAllResourceFromSsn %v.", tp.Name(), coresErr)
			continue
		}
		var nodeAllUsableCores = 0
		for _, tmp := range nodeCoresInf {
			nodeAllUsableCores += tmp.UnCutCore
		}
		allocatableResource[nodInf.Name] = nodeAllUsableCores
	}
	if len(allocatableResource) == 0 {
		return nil, errors.New("nil npu node")
	}
	return allocatableResource, nil
}

// AllocCacheVJobsIntoCache Allocate the resources required by the vJobs to the cache.
func (tp *VNPU) AllocCacheVJobsIntoCache(jobs []*api.JobInfo, res map[string]int, ssn *framework.Session) error {
	var allocErrores error
	for _, vJob := range jobs {
		pluginName, nameErr := tp.GetPluginNameByJobInfo(vJob)
		if nameErr != nil {
			klog.V(util.LogErrorLev).Infof("%s AllocCacheVJobsIntoCache %s %v.", tp.Name(), vJob.Name, nameErr)
			return nameErr
		}
		switch pluginName {
		case vnpuutil.PluginNameBy910VNPU:
			v910Plugin := &modulev910.ChipV910{}
			if pluginErr := v910Plugin.InitVNPUPlugin(); pluginErr != nil {
				return pluginErr
			}
			tp.Attr = v910Plugin.ComVNPU
		case vnpuutil.PluginNameBy710VNPU:
			v710Plugin := &modulev710.ChipV710{}
			if pluginErr := v710Plugin.InitVNPUPlugin(); pluginErr != nil {
				return pluginErr
			}
			tp.Attr = v710Plugin.ComVNPU
		default:
			klog.V(util.LogErrorLev).Infof("unknown VNPU chip type %s.", pluginName)
			return nil
		}
		if allocErr := tp.AllocCacheVJobIntoCache(vJob, res, ssn); allocErr != nil {
			if tmpErr := multierror.Append(allocErrores, allocErr); tmpErr != nil {
				return tmpErr
			}
		}
	}
	return allocErrores
}

// AllocNewVNPUJobsFromCache alloc vJobs after GetAndRecordNewVNPUJobsFromSsn.
func (tp *VNPU) AllocNewVNPUJobsFromCache(ssn *framework.Session) error {
	// 1.Get all need allocate jobs
	vJobs, getErr := tp.GetAllNeedAllocVJobsFromCache(ssn)
	if getErr != nil {
		klog.V(util.LogErrorLev).Infof("%s GetAllNeedAllocVJobsFromCache :%v.", vnpuutil.PluginName, getErr)
		return getErr
	}
	// 2.Order the jobs by create time(all kinds vJobs are in one list).
	vJobsList, orderErr := tp.OrderVJobsByCreateTime(vJobs)
	if orderErr != nil {
		klog.V(util.LogErrorLev).Infof("%s OrderVjobByTimeAndKinds :%v.", vnpuutil.PluginName, orderErr)
		return orderErr
	}
	klog.V(util.LogDebugLev).Infof("%s AllocCacheVJobsIntoCache get %d-%d jobs.",
		vnpuutil.PluginName, len(vJobs), len(vJobsList))
	// 3.Get cluster resource(node,chip).
	clusterResource, resErr := tp.GetClusterAllResourceFromSsn(ssn)
	if resErr != nil {
		klog.V(util.LogErrorLev).Infof("%s GetClusterAllResourceFromSsn :%v.", vnpuutil.PluginName, resErr)
		return resErr
	}
	klog.V(util.LogDebugLev).Infof("%s AllocNewVNPUJobsFromCache:%+v.", vnpuutil.PluginName, clusterResource)
	// 4.Alloc vJobs
	if allocErr := tp.AllocCacheVJobsIntoCache(vJobsList, clusterResource, ssn); allocErr != nil {
		klog.V(util.LogErrorLev).Infof("%s AllocCacheVJobsIntoCache :%v.", vnpuutil.PluginName, allocErr)
		return allocErr
	}
	klog.V(util.LogInfoLev).Infof("%s AllocCacheVJobsIntoCache success.", vnpuutil.PluginName)
	return nil
}

func (tp *VNPU) writevNPUAllocInfIntoCm(ssn *framework.Session) error {
	var vNPUCM = &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      vnpuutil.VNPUCMName,
			Namespace: vnpuutil.VNPUCMNameSpace,
		},
		Data: vnpuutil.GetVNPUCMData(vnpuutil.VNPUAllocData),
	}

	klog.V(util.LogDebugLev).Infof("Write vNPU alloc inf into cm: %+v.", vNPUCM)
	if err := util.CreateOrUpdateConfigMap(ssn.KubeClient(),
		vNPUCM, vnpuutil.VNPUCMName, vnpuutil.VNPUCMNameSpace); err != nil {
		klog.V(util.LogErrorLev).Infof("writevNPUAllocInfIntoCm : %v.", err)
		return err
	}

	return nil
}

// DealNewVNPUJob Deal ssn new vJob.
func (tp *VNPU) DealNewVNPUJob(ssn *framework.Session) error {
	if err := tp.GetAndRecordNewVNPUJobsFromSsn(ssn); err != nil {
		klog.V(util.LogDebugLev).Infof("%s GetAndRecordNewVNPUJobsFromSsn :%v.", vnpuutil.PluginName, err)
		return err
	}

	if err := tp.AllocNewVNPUJobsFromCache(ssn); err != nil {
		klog.V(util.LogErrorLev).Infof("%s AllocNewVNPUJobsFromCache :%v.", vnpuutil.PluginName, err)
		return err
	}

	return tp.writevNPUAllocInfIntoCm(ssn)
}

// GetVJobNamesFromCache key is vJob name,value is pod name
func (tp *VNPU) GetVJobNamesFromCache() ([]string, error) {
	var jobNames []string
	for _, data := range vnpuutil.VNPUAllocData.Cache {
		name := data.JobUID
		jobNames = append(jobNames, string(name))
	}
	if len(jobNames) == 0 {
		return nil, errors.New("nil jobs in cache")
	}

	return jobNames, nil
}

// CheckVJobCanBeDeleteByName for one vJob has one task.
func (tp *VNPU) CheckVJobCanBeDeleteByName(vJobName string, ssn *framework.Session) bool {
	jobUID := api.JobID(vJobName)
	jobInf, ok := ssn.Jobs[jobUID]
	if !ok {
		klog.V(util.LogErrorLev).Infof("CheckVJobCanBeDeleteByName %s not exist", vJobName)
		return true
	}

	if jobInf.PodGroup == nil {
		// no pg means failed or complete.
		klog.V(util.LogErrorLev).Infof("CheckVJobCanBeDeleteByName %s pg not exist", vJobName)
		return true
	}

	for _, task := range jobInf.Tasks {
		if task.Status == api.Succeeded || task.Status == api.Failed {
			return true
		}
		klog.V(util.LogDebugLev).Infof("CheckVJobCanBeDeleteByName %s task status %v", task.Job, task.Status)
		return false
	}
	klog.V(util.LogErrorLev).Infof("CheckVJobCanBeDeleteByName %s no tasks", vJobName)
	// no task maybe enqueue,can not delete.
	return false
}

// IsDeleteVJobOverTime judge vJob whether is over the wait time.
func (tp *VNPU) IsDeleteVJobOverTime(vJobName string) bool {
	for _, data := range vnpuutil.VNPUAllocData.Cache {
		if string(data.JobUID) == vJobName {
			now := time.Now().Unix()
			if now-data.UpdateTime > vnpuutil.DeleteOverTime {
				klog.V(util.LogErrorLev).Infof("IsDeleteVJobOverTime %v==%v", now, data.UpdateTime)
				return true
			}
			return false
		}
	}
	return true
}

// IsVJobNeedDeleteBySSN judge the vJob whether need delete from cache.
func (tp *VNPU) IsVJobNeedDeleteBySSN(vJobName string, ssn *framework.Session) bool {
	if !tp.CheckVJobCanBeDeleteByName(vJobName, ssn) {
		return false
	}

	// over time
	return tp.IsDeleteVJobOverTime(vJobName)
}

func (tp *VNPU) getNeedKeepJobsInCache(vJobNames []string, ssn *framework.Session) ([]string, error) {
	var vJobKeepNames []string
	for _, jobName := range vJobNames {
		if tp.IsVJobNeedDeleteBySSN(jobName, ssn) {
			continue
		}
		vJobKeepNames = append(vJobKeepNames, jobName)
	}
	// Do not judge whether vJobKeepNames is nil.
	return vJobKeepNames, nil
}

// excludeVJobsByJobNamesInCache only the parameters vJob can be keep in cache.
func (tp *VNPU) excludeVJobsByJobNamesInCache(jobNames []string) error {
	var tmp []vnpuutil.VNPUAllocInf
	for _, data := range vnpuutil.VNPUAllocData.Cache {
		for _, jobName := range jobNames {
			if string(data.JobUID) == jobName {
				tmp = append(tmp, data)
				break
			}
		}
	}
	vnpuutil.VNPUAllocData.Cache = tmp
	vnpuutil.VNPUAllocData.CheckCode = util.MakeDataHash(vnpuutil.VNPUAllocData.Cache)
	klog.V(util.LogDebugLev).Infof("%s excludeVJobsByJobNamesInCache result %+v.", tp.Name(),
		vnpuutil.VNPUAllocData.Cache)
	return nil
}

// DealFinishedVNPUJob deal finished VNPU job from job, include completed, failed , nonexistent.
func (tp *VNPU) DealFinishedVNPUJob(ssn *framework.Session) error {
	// 1.get all jobNames from cache
	vJobNames, nameErr := tp.GetVJobNamesFromCache()
	if nameErr != nil {
		klog.V(util.LogInfoLev).Infof("GetVJobNamesFromCache :%v.", nameErr)
		return nameErr
	}
	// 2.get all need keep jobNames
	keepJobNames, delErr := tp.getNeedKeepJobsInCache(vJobNames, ssn)
	if delErr != nil {
		klog.V(util.LogErrorLev).Infof("getNeedKeepJobsInCache :%v.", delErr)
		return delErr
	}
	// 3.update info into cache
	if updateErr := tp.excludeVJobsByJobNamesInCache(keepJobNames); updateErr != nil {
		klog.V(util.LogErrorLev).Infof("excludeVJobsByJobNamesInCache :%v.", updateErr)
		return updateErr
	}
	// 4.write new cache into cm
	return tp.writevNPUAllocInfIntoCm(ssn)
}

// GetNodeListByReqVNPU consider node,card,chip
func (tp *VNPU) GetNodeListByReqVNPU(jobNeedNPUType string, ssn *framework.Session) ([]*api.NodeInfo, error) {
	if tp == nil {
		return nil, errors.New(vnpuutil.PluginUninitializedError)
	}
	var nodeList []*api.NodeInfo
	for _, nodeInf := range ssn.Nodes {
		if !vnpuutil.IsNPUCardNodeByCardName(tp.Attr.AnnoName, nodeInf) {
			klog.V(util.LogDebugLev).Infof("%s GetNodeListByReqVNPU no %v in %v.",
				tp.Name(), tp.Attr.AnnoName, nodeInf.Name)
			continue
		}
		if selectorErr := tp.IsNodeHasVNPUSelector(nodeInf); selectorErr != nil {
			klog.V(util.LogDebugLev).Infof("%s GetNodeListByReqVNPU %v.", tp.Name(), selectorErr)
			continue
		}
		if !tp.IsVNPUNodeMeetReqResource(jobNeedNPUType, nodeInf) {
			klog.V(util.LogDebugLev).Infof("%s GetNodeListByReqVNPU %v no in %s.",
				tp.Name(), jobNeedNPUType, nodeInf.Name)
			continue
		}
		if !vnpuutil.IsNPUResourceStableInNode(tp.Attr.AnnoName, nodeInf) {
			klog.V(util.LogDebugLev).Infof("%s GetNodeListByReqVNPU %v unstable in %s.",
				tp.Name(), tp.Attr.AnnoName, nodeInf.Name)
			continue
		}
		nodeList = append(nodeList, nodeInf)
	}
	if len(nodeList) == 0 {
		listErr := fmt.Errorf("none node list meet %s", jobNeedNPUType)
		return nil, listErr
	}
	return nodeList, nil
}

func (tp *VNPU) getChipNameByID(id int) string {
	name := tp.Attr.AnnoPreVal + "-" + strconv.Itoa(id)
	return name
}

// getTheFillOneFromList Obtain the chip that meets the core requirement. get the smallest one.
func (tp *VNPU) getTheFillOneFromList(chipCores map[int]int) (string, error) {
	var min = maxNPUChipCores
	var selectID = 0
	for cardID, num := range chipCores {
		if min < num {
			continue
		}
		min = num
		selectID = cardID
	}
	if min == 0 {
		return "", errors.New("get nil")
	}
	selectCard := strings.TrimLeft(tp.Attr.AnnoName, tp.Attr.AnnoPreVal) + "-" + strconv.Itoa(selectID)
	return selectCard, nil
}

// GetVJobMeetChip get the chip name meet vJob require in nodeInf.
func (tp *VNPU) GetVJobMeetChip(nodeInf *api.NodeInfo, vJob *api.JobInfo) (string, error) {
	// 1.Get vJob need VNPU.
	needNPU, getErr := tp.GetVJobNeedVNPU(vJob)
	if getErr != nil {
		klog.V(util.LogErrorLev).Infof("%s GetVJobMeetChip %s %v.", tp.Name(), vJob.Name, getErr)
		return "", getErr
	}
	selectChip, nodeErr := tp.GetVNPUUsedChipByReq(needNPU, nodeInf)
	if nodeErr != nil {
		return "", nodeErr
	}
	return selectChip, nil
}

// RecordVJobAllocInfoInCache record alloc info in cache and node.other,cache must be exist
func (tp *VNPU) RecordVJobAllocInfoInCache(nodeInf *api.NodeInfo, chip string, vJob *api.JobInfo) error {
	changeFlag := false
	for k, data := range vnpuutil.VNPUAllocData.Cache {
		if vJob.UID == data.JobUID {
			tmp := data
			tmp.NodeName = nodeInf.Name
			tmp.ReqCardName = chip
			tmp.AllocFlag = true
			tmp.UpdateTime = time.Now().Unix()
			vnpuutil.VNPUAllocData.Cache[k] = tmp
			changeFlag = true
			break
		}
	}
	if !changeFlag {
		noErr := fmt.Errorf("%s not in cache", vJob.UID)
		klog.V(util.LogErrorLev).Infof("%s RecordVJobAllocInfoInCache %v.", tp.Name(), noErr)
		return noErr
	}
	klog.V(util.LogDebugLev).Infof("%s RecordVJobAllocInfoInCache %s record %+v.", tp.Name(), vJob.UID,
		vnpuutil.VNPUAllocData.Cache)
	klog.V(util.LogDebugLev).Infof("%s haha-1 record %v %+v.", nodeInf.Name, chip, nodeInf.Others)
	// reduce the select one from node other.
	if reduceErr := tp.reduceTheAllocChipFromNodeOther(chip, vJob, nodeInf); reduceErr != nil {
		klog.V(util.LogDebugLev).Infof("%s RecordVJobAllocInfoInCache %s %+v.", tp.Name(), vJob.UID, reduceErr)
		return reduceErr
	}
	klog.V(util.LogDebugLev).Infof("%s haha-2 record %+v.", nodeInf.Name, nodeInf.Others)
	return nil
}

// AllocCacheVJobIntoCache alloc vJob into cache by resources from session
func (tp *VNPU) AllocCacheVJobIntoCache(vJob *api.JobInfo, res map[string]int, ssn *framework.Session) error {
	if tp == nil {
		return errors.New(vnpuutil.PluginUninitializedError)
	}
	// 1.Get meet node list
	nodeList, listErr := tp.GetVJobMeetNodeList(vJob, res, ssn)
	if listErr != nil {
		klog.V(util.LogErrorLev).Infof("%s GetVJobMeetNodeList %s %v.", tp.Name(), vJob.Name, listErr)
		return listErr
	}
	// 2. Get meet node
	tmpNode, nodeErr := tp.GetVJobMeetNode(nodeList)
	if nodeErr != nil {
		klog.V(util.LogErrorLev).Infof("%s GetVJobMeetNode %s %v.", tp.Name(), vJob.Name, nodeErr)
		return nodeErr
	}
	// 3. Get meet chip.
	tmpChip, chipErr := tp.GetVJobMeetChip(tmpNode, vJob)
	if chipErr != nil {
		klog.V(util.LogErrorLev).Infof("%s GetVJobMeetChip %s in %s %v.", tp.Name(),
			vJob.Name, tmpNode.Name, chipErr)
		return chipErr
	}
	// 5 record alloc info in cache and node.other
	recordErr := tp.RecordVJobAllocInfoInCache(tmpNode, tmpChip, vJob)
	if recordErr != nil {
		klog.V(util.LogErrorLev).Infof("%s AllocCacheVJobIntoCache %v.", tp.Name(), recordErr)
		return nodeErr
	}
	return nil
}
