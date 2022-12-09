/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package vnpu is using for HuaWei Ascend pin vnpu allocation.
*/
package vnpu

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// CheckNodeNPUByTask check chip on node has enough resource, fault chips are not in list, unstable excluded
func (tp *DynamicVNPU) CheckNodeNPUByTask(task *api.TaskInfo, node plugin.NPUNode, taskResReq util.VResource) error {
	klog.V(util.LogInfoLev).Infof("dynamic vnpu task<%s> node<%s> CheckNodeNPUByTask", task.Name, node.Name)
	if !node.ValidVNode {
		klog.V(util.LogInfoLev).Infof("dynamic vnpu node<%s> not valid vNode", node.Name)
		return errors.New("CheckNodeNPUByTask invalid VNode")
	}
	if !node.IsNodeTotalResEnough(taskResReq) || !node.IsNodeChipResEnough(taskResReq) {
		// if node resource not enough, reduce task aiCPU
		if tp.taskAICPUCanBeDowngrade(taskResReq) {
			klog.V(util.LogInfoLev).Infof("dynamic vnpu task<%s> resource not enough, downgrade cpu", task.Name)
			tp.Cache[task.Name] = append(tp.Cache[task.Name], node.Name)
			return tp.CheckNodeNPUByTask(task, node, tp.downgradeTaskAICPU(taskResReq))
		}
		return fmt.Errorf("dynamic vnpu task<%s> CheckNodeNPUByTask node %s resource not enough",
			task.Name, node.Name)
	}
	klog.V(util.LogInfoLev).Infof("dynamic vnpu task<%s> CheckNodeNPUByTask node<%s> ok", task.Name, node.Name)
	return nil
}

// ScoreBestNPUNodes node with least free resource would be sorted to higher rank
func (tp *DynamicVNPU) ScoreBestNPUNodes(task *api.TaskInfo, nodes []*api.NodeInfo, scoreMap map[string]float64) error {
	klog.V(util.LogInfoLev).Infof("dynamic vnpu task<%s> ScoreBestNPUNodes", task.Name)
	// 1. sort nodes with free resource from low to high
	nodesSorted := tp.orderVNodesByFreeResource(nodes)
	if len(nodesSorted) == 0 {
		return fmt.Errorf("dynamic vnpu task<%s> ScoreBestNPUNodes err: sorted nodes len 0", task.Name)
	}

	downgradeNodes, ok := tp.Cache[task.Name]
	// 2. give the first node high score, none nodes are downgraded
	if !ok {
		scoreMap[nodesSorted[0].Name] += util.NPUIndex8
		return nil
	}

	// 3. if downgrade nodes exists, skip, util find none-downgraded nodes and add score
	for _, node := range nodesSorted {
		downgradeFlag := false
		for _, dNode := range downgradeNodes {
			if node.Name == dNode {
				downgradeFlag = true
				break
			}
		}
		if !downgradeFlag {
			scoreMap[node.Name] += util.NPUIndex8
			return nil
		}
	}

	return nil
}

// UseAnnotation write task use vnpu to pod annotation
func (tp *DynamicVNPU) UseAnnotation(task *api.TaskInfo, node plugin.NPUNode, taskResReq util.VResource) *plugin.NPUNode {
	klog.V(util.LogDebugLev).Infof("dynamic vnpu UseAnnotation task<%s> node<%s> Annotation: %#v", task.Name,
		node.Name, node.Annotation)

	allocChipID, err := node.VNode.SelectChipFromNode(taskResReq)
	if err != nil {
		// if chips cannot be found, reduce aiCPU
		if tp.taskAICPUCanBeDowngrade(taskResReq) {
			taskResReqDown := tp.downgradeTaskAICPU(taskResReq)
			return tp.UseAnnotation(task, node, taskResReqDown)
		}
		klog.V(util.LogErrorLev).Infof("dynamic vnpu task<%s> UseAnnotation err: %s", task.Name, err.Error())
		return nil
	}

	tp.SetNPUTopologyToPodFn(task, node, taskResReq, allocChipID)
	return tp.UpdateNodeInfo(node, allocChipID, taskResReq)
}

// taskAICPUCanBeDowngrade if task label is low, aicpu can be lower
func (tp *DynamicVNPU) taskAICPUCanBeDowngrade(taskResReq util.VResource) bool {
	if taskResReq.Aicore == util.NPUIndex2 && taskResReq.Aicpu == util.NPUIndex2 {
		return true
	}
	if taskResReq.Aicore == util.NPUIndex4 && taskResReq.Aicpu == util.NPUIndex4 && taskResReq.DVPP != plugin.
		AscendDVPPEnabledOn {
			return true
	}

	return false
}

func (tp *DynamicVNPU) downgradeTaskAICPU(taskResReq util.VResource) util.VResource {
	if taskResReq.Aicore == util.NPUIndex2 {
		return util.VResource{
			Aicore: taskResReq.Aicore,
			Aicpu:  util.NPUIndex1,
			DVPP:   taskResReq.DVPP,
		}
	}
	if taskResReq.Aicore == util.NPUIndex4 {
		return util.VResource{
			Aicore: taskResReq.Aicore,
			Aicpu:  util.NPUIndex3,
			DVPP:   taskResReq.DVPP,
		}
	}
	return taskResReq
}

// SetNPUTopologyToPodFn write chip to pod annotation AscendNPUCore
func (tp *DynamicVNPU) SetNPUTopologyToPodFn(task *api.TaskInfo, node plugin.NPUNode, taskResReq util.VResource,
	allocChipID string) {
	// 1. whole card
	if node.IsResourceWholeCard(taskResReq.Aicore) {
		task.Pod.Annotations[util.AscendNPUCore] = allocChipID
		klog.V(util.LogInfoLev).Infof("dynamic vnpu setNPUTopologyToPod %s top:%s.", task.Name, allocChipID)
		return
	}

	// 2.1 like vir04
	segmentAnnotation := fmt.Sprintf("%s-%s", allocChipID, getAiCoreNumStr(taskResReq.Aicore))
	if node.ChipKind != plugin.Ascend310P || taskResReq.Aicore == taskResReq.Aicpu {
		task.Pod.Annotations[util.AscendNPUCore] = segmentAnnotation
		klog.V(util.LogInfoLev).Infof("dynamic vnpu setNPUTopologyToPod %s top:%s.", task.Name, segmentAnnotation)
		return
	}

	// 2.2 like vir04_3c
	segmentAnnotation = fmt.Sprintf("%s_%sc", segmentAnnotation, strconv.Itoa(taskResReq.Aicpu))

	// 2.3 like vir04_3c_ndvpp
	task.Pod.Annotations[util.AscendNPUCore] = getDVPPValue(taskResReq.DVPP, segmentAnnotation)
	klog.V(util.LogInfoLev).Infof("dynamic vnpu setNPUTopologyToPod %s top:%s.", task.Name, segmentAnnotation)
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

// UpdateNodeInfo vnpu update npuNode after allocation
func (tp *DynamicVNPU) UpdateNodeInfo(node plugin.NPUNode, allocChipID string, taskResReq util.VResource) *plugin.NPUNode {
	if node.IsResourceWholeCard(taskResReq.Aicore) {
		return tp.UpdateNodeInfoWhole(node, allocChipID)
	}
	return tp.UpdateNodeInfoSegment(node, allocChipID, taskResReq)
}

// UpdateNodeInfoSegment vnpu update npuNode after allocation for segmentation tasks
func (tp *DynamicVNPU) UpdateNodeInfoSegment(node plugin.NPUNode, allocChipID string,
	taskResReq util.VResource) *plugin.NPUNode {
	for chipID, chip := range node.Chips {
		if strconv.Itoa(chipID) != allocChipID {
			continue
		}
		chip.UsedRes.Add(taskResReq)
		chip.FreeRes.Sub(taskResReq)
		if !node.IsResourceWholeCard(taskResReq.Aicore) {
			chip.SegmentFlag = true
		}
		chip.UpdateDVPP(taskResReq.DVPP)
	}
	klog.V(util.LogInfoLev).Infof("dynamic vnpu UpdateNodeInfo node <%s> chip resource updated", node.Name)
	return &node
}

// UpdateNodeInfoWhole vnpu update npuNode after allocation for whole card tasks
func (tp *DynamicVNPU) UpdateNodeInfoWhole(node plugin.NPUNode, allocChipIDs string) *plugin.NPUNode {
	chipRes := util.VResource{
		Aicore: node.AiCorePerChip,
		Aicpu:  node.TotalRes.Aicpu / node.TotalChipNum,
		DVPP:   plugin.AscendDVPPEnabledNull,
	}
	allocChipIDList := strings.Split(allocChipIDs, ",")
	for _, allocChipID := range allocChipIDList {
		for chipID, chip := range node.Chips {
			if strconv.Itoa(chipID) != allocChipID {
				continue
			}
			chip.UsedRes.Add(chipRes)
			chip.FreeRes.Sub(chipRes)
			chip.UpdateDVPP(chipRes.DVPP)
		}
	}
	return &node
}
