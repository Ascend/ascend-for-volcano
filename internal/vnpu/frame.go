/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package vnpu is using for HuaWei Ascend pin fault rescheduling.

*/
package vnpu

import (
	"fmt"
	"strconv"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/conf"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// CheckVNPUSegmentEnableByConfig Check VNPU segmentEnable by init plugin parameters.
func CheckVNPUSegmentEnableByConfig(configurations []conf.Configuration) bool {
	configuration, err := util.GetConfigFromSchedulerConfigMap(util.CMInitParamKey, configurations)
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
	if segmentEnable == "false" {
		return true
	}
	return false
}

// CheckVNPUSegmentEnable Check VNPU segmentEnable by init plugin parameters.
func CheckVNPUSegmentEnable(ssn *framework.Session) bool {
	if len(ssn.Configurations) == 0 {
		klog.V(util.LogDebugLev).Info("no configurations, segmentEnable will not be changed.")
		return false
	}

	return CheckVNPUSegmentEnableByConfig(ssn.Configurations)
}

// CheckNodeNPUByTask todo: deal with fault chips
func (tp *ComVNPU) CheckNodeNPUByTask(task *api.TaskInfo, node plugin.NPUNode) error {
	taskResReq, err := plugin.TransferTaskLabelToResReq(task)
	if err != nil {
		return fmt.Errorf("%s task<%s> CheckNodeNPUByTask err: %s", tp.GetPluginName(), task.Name, err.Error())
	}

	if !node.IsNodeTotalResEnough(taskResReq) {
		return fmt.Errorf("%s task<%s> CheckNodeNPUByTask err: node resource not enough", tp.GetPluginName(),
			task.Name)
	}

	if !node.IsNodeChipResEnough(taskResReq) {
		return fmt.Errorf("%s task<%s> CheckNodeNPUByTask err: chip resource not enough", tp.GetPluginName(),
			task.Name)
	}

	return nil
}

// ScoreBestNPUNodes node with least free resource would be sorted to higher rank
func (tp *ComVNPU) ScoreBestNPUNodes(task *api.TaskInfo, nodes []*api.NodeInfo, scoreMap map[string]float64) error {
	// 1. sort nodes with free resource from low to high
	nodesSorted := tp.orderVNodesByFreeResource(nodes)
	if len(nodesSorted) == 0 {
		return fmt.Errorf("%s task<%s> ScoreBestNPUNodes err: sorted nodes len 0", tp.GetPluginName(), task.Name)
	}

	// 2. give the first node high score
	scoreMap[nodesSorted[0].Name] += util.NPUIndex8
	return nil
}

// UseAnnotation write task use vnpu to pod annotation
func (tp *ComVNPU) UseAnnotation(task *api.TaskInfo, node plugin.NPUNode) *plugin.NPUNode {
	klog.V(util.LogDebugLev).Infof("%s UseAnnotation task<%s> node<%s> resource<%s> Annotation: %#v",
		tp.GetPluginName(), task.Name, node.Name, tp.GetAnnoName(), node.Annotation)
	taskResReq, err := plugin.TransferTaskLabelToResReq(task)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s task<%s> UseAnnotation err: %s", tp.GetPluginName(), task.Name, err.Error())
		return nil
	}

	allocChipID, err := node.VNode.SelectChipFromNode(taskResReq)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s task<%s> UseAnnotation err: %s", tp.GetPluginName(), task.Name, err.Error())
		return nil
	}

	tp.SetNPUTopologyToPodFn(task, node, taskResReq, allocChipID)
	return tp.UpdateNodeInfo(node, allocChipID, taskResReq)
}

// SetNPUTopologyToPodFn write chip to pod annotation AscendNPUCore
func (tp *ComVNPU) SetNPUTopologyToPodFn(task *api.TaskInfo, node plugin.NPUNode, taskResReq util.VResource,
	allocChipID string) {
	// 1. whole card
	if node.IsResourceWholeCard(taskResReq) {
		task.Pod.Annotations[util.AscendNPUCore] = allocChipID
		klog.V(util.LogInfoLev).Infof("%s setNPUTopologyToPod %s top:%s.", tp.GetPluginName(),
			task.Name, allocChipID)
		return
	}

	// 2.1 like vir04
	segmentAnnotation := fmt.Sprintf("%s-%s", allocChipID, getAiCoreNumStr(taskResReq.Aicore))
	if node.ChipKind != plugin.Ascend310P || taskResReq.Aicore == taskResReq.Aicpu {
		task.Pod.Annotations[util.AscendNPUCore] = segmentAnnotation
		klog.V(util.LogInfoLev).Infof("%s setNPUTopologyToPod %s top:%s.", tp.GetPluginName(),
			task.Name, segmentAnnotation)
		return
	}

	// 2.2 like vir04_3c
	segmentAnnotation = fmt.Sprintf("%s_%sc", segmentAnnotation, strconv.Itoa(taskResReq.Aicpu))

	// 2.3 like vir04_3c_ndvpp
	task.Pod.Annotations[util.AscendNPUCore] = getDVPPValue(taskResReq.DVPP, segmentAnnotation)
	klog.V(util.LogInfoLev).Infof("%s setNPUTopologyToPod %s top:%s.", tp.GetPluginName(),
		task.Name, segmentAnnotation)
	return
}

func getAiCoreNumStr(AiCoreNum int) string {
	coreNumStr := strconv.Itoa(AiCoreNum)
	if len(coreNumStr) < util.NPUIndex2 {
		coreNumStr = "0" + coreNumStr
	}
	return plugin.AscendVNPUPrefix + coreNumStr
}

func getDVPPValue(DVPPEnable string, preAnno string) string {
	switch DVPPEnable {
	case plugin.AscendDVPPEnabledNull:
		klog.V(util.LogDebugLev).Infof("null dvpp")
		return ""
	case plugin.AscendDVPPEnabledOff:
		return fmt.Sprintf("%s_%s", preAnno, plugin.AscendNDVPPValue)
	case plugin.AscendDVPPEnabledOn:
		return fmt.Sprintf("%s_%s", preAnno, plugin.AscendDVPPValue)
	}
	return ""
}

// UpdateNodeInfo update npuNode after allocation
func (tp *ComVNPU) UpdateNodeInfo(node plugin.NPUNode, allocChipID string, taskResReq util.VResource) *plugin.NPUNode {
	for chipID, chip := range node.Chips {
		if strconv.Itoa(chipID) != allocChipID {
			continue
		}
		chip.UsedRes.Add(taskResReq)
		chip.FreeRes.Sub(taskResReq)
		if !node.IsResourceWholeCard(taskResReq) {
			chip.SegmentFlag = true
		}
	}
	klog.V(util.LogInfoLev).Infof("%s UpdateNodeInfo node <%s> chip resource updated", tp.GetPluginName(),
		node.Name)
	return &node
}
