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
	"strings"

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

// PreCheckNodeFn check whether the node matches the tag requirements of the task.
func (tp *VNPU) PreCheckNodeFn(task *api.TaskInfo, _ *api.NodeInfo, confs []conf.Configuration) error {
	klog.V(util.LogDebugLev).Infof("%s PreCheckNodeFn %s enter.", tp.Name(), task.Name)
	defer klog.V(util.LogDebugLev).Infof("%s PreCheckNodeFn %s leave.", tp.Name(), task.Name)

	err := vnpuutil.CheckVNPUSegmentEnableByConfig(confs)
	klog.V(util.LogErrorLev).Infof("%s PreCheckNodeFn %v.", vnpuutil.PluginName, err)
	return err
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

func (tp *VNPU) setVPUPluginToVNPUBack() {
	// set vnpu plugin to vnp back.
	tp.Attr.PluginName = PluginName
}
