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

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/vnpu/modulev310p"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/vnpu/modulev910"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/vnpu/vnpuutil"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
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

// PreHandleVNPU Only for abstract VNPU, not v910,v310P and so on.
func (tp *VNPU) PreHandleVNPU(ssn *framework.Session) error {
	err := vnpuutil.CheckVNPUSegmentEnable(ssn)
	klog.V(util.LogDebugLev).Infof("PreHandleVNPU :%v.", err)
	return err
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
func (tp *VNPU) PreCheckNodeFn(task *api.TaskInfo, node *api.NodeInfo, confs []conf.Configuration) error {
	klog.V(util.LogDebugLev).Infof("%s PreCheckNodeFn %s enter.", tp.Name(), task.Name)
	defer klog.V(util.LogDebugLev).Infof("%s PreCheckNodeFn %s leave.", tp.Name(), task.Name)
	if err := vnpuutil.CheckVNPUSegmentEnableByConfig(confs); err != nil {
		klog.V(util.LogErrorLev).Infof("%s PreCheckNodeFn %v.", vnpuutil.PluginName, err)
		return err
	}
	// presetVirtualDevice does not need follow actions, these will be delete
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
	var bestNodesMap = make(map[string]int, util.NPUIndex3)

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
	case vnpuutil.PluginNameBy310PVNPU:
		v310P := &modulev310p.ChipV310P{}
		if err := v310P.InitVNPUPlugin(); err != nil {
			break
		}
		tp.Attr = v310P.ComVNPU
		var f VNPUHandler = v310P
		tp.Plugin = f
	default:
		supErr := fmt.Errorf("not support plugin type :%s", reqNpuType)
		klog.V(util.LogErrorLev).Infof("%s InitVNPUPluginByType %v.", tp.Name(), supErr)
		return supErr
	}
	return nil
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
		case vnpuutil.PluginNameBy310PVNPU:
			v310PPlugin := &modulev310p.ChipV310P{}
			if pluginErr := v310PPlugin.InitVNPUPlugin(); pluginErr != nil {
				return pluginErr
			}
			tp.Attr = v310PPlugin.ComVNPU
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
	// reduce the select one from node other.
	if reduceErr := tp.reduceTheAllocChipFromNodeOther(chip, vJob, nodeInf); reduceErr != nil {
		klog.V(util.LogDebugLev).Infof("%s RecordVJobAllocInfoInCache %s %+v.", tp.Name(), vJob.UID, reduceErr)
		return reduceErr
	}
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

func (tp *VNPU) setVPUPluginToVNPUBack() {
	// set vnpu plugin to vnp back.
	tp.Attr.PluginName = PluginName
}
