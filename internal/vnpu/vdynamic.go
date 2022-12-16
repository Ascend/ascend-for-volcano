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
	"time"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// GetTemplateByResReq get template by resource request.
func (tp *DynamicVNPU) GetTemplateByResReq(taskResReq util.VResource, vt VTemplate) (string, error) {
	name := ""
	for tName, value := range vt.Data {
		if value.Aicore != taskResReq.Aicore {
			continue
		}
		if value.Aicpu != taskResReq.Aicpu {
			continue
		}
		if value.DVPP != taskResReq.DVPP {
			continue
		}
		name = tName
	}
	if name == "" {
		return "", fmt.Errorf("%#v not get template", taskResReq)
	}
	return name, nil
}

// IsNodeHasDifferentUnFinishedTask judge the node wither has the different template unfinished job.
func (tp *VirtualNPU) IsNodeHasDifferentUnFinishedTask(taskInfo *api.TaskInfo, nodeInf plugin.NPUNode,
	taskResReq util.VResource) error {
	klog.V(util.LogDebugLev).Infof("%s IsNodeHasDifferentUnFinishedTask cache :%#v", taskInfo.Name, tp.ConCache)
	nodeTempMap, ok := tp.ConCache[nodeInf.Name]
	if !ok {
		klog.V(util.LogDebugLev).Infof("%s IsNodeHasDifferentUnFinishedTask cache no node %s, ok.",
			taskInfo.Name, nodeInf.Name)
		return nil
	}
	template, getErr := tp.GetTemplateByResReq(taskResReq, tp.VT)
	if getErr != nil {
		klog.V(util.LogDebugLev).Infof("IsNodeHasDifferentUnFinishedTask %s", getErr)
		return getErr
	}
	if len(nodeTempMap) == 1 {
		_, tOK := nodeTempMap[template]
		if tOK {
			klog.V(util.LogDebugLev).Infof("%s IsNodeHasDifferentUnFinishedTask cache no template:%s, ok.",
				taskInfo.Name, template)
			return nil
		}
	}

	return fmt.Errorf("%s is using %s, and not rewrite", taskInfo.Name, nodeInf.Name)
}

// CheckNodeNPUByDyTask check chip on node has enough resource, fault chips are not in list, unstable excluded
func (tp *VirtualNPU) CheckNodeNPUByDyTask(task *api.TaskInfo, node plugin.NPUNode, taskResReq util.VResource) error {
	klog.V(util.LogDebugLev).Infof("check dynamic vNPU %s on %s", task.Name, node.Name)
	if !node.ValidVNode {
		klog.V(util.LogInfoLev).Infof("dynamic vNPU node<%s> not valid vNode", node.Name)
		return errors.New("checkNodeNPUByDyTask invalid VNode")
	}
	if node.IsNodeMeetRes(taskResReq) {
		// if node resource not enough, reduce task aiCPU
		if tp.taskAICPUCanBeDowngrade(taskResReq) {
			klog.V(util.LogInfoLev).Infof("dynamic vnpu task<%s> resource not enough, downgrade cpu", task.Name)
			tp.DowngradeCache[task.Name] = append(tp.DowngradeCache[task.Name], node.Name)
			return tp.CheckNodeNPUByDyTask(task, node, tp.downgradeTaskAICPU(taskResReq))
		}
		return fmt.Errorf("dynamic vnpu task<%s> CheckNodeNPUByDyTask node %s resource not enough",
			task.Name, node.Name)
	}
	if diffErr := tp.IsNodeHasDifferentUnFinishedTask(task, node, taskResReq); diffErr != nil {
		return diffErr
	}
	klog.V(util.LogInfoLev).Infof("dynamic vnpu task<%s> CheckNodeNPUByDyTask node<%s> ok", task.Name, node.Name)
	return nil
}

// ScoreBestNPUNodes node with the least free resource would be sorted to higher rank
func (tp *DynamicVNPU) ScoreBestNPUNodes(task *api.TaskInfo, nodes []*api.NodeInfo, scoreMap map[string]float64) error {
	klog.V(util.LogInfoLev).Infof("dynamic vnpu task<%s> ScoreBestNPUNodes", task.Name)
	if len(scoreMap) == 0 {
		return errors.New(util.ArgumentError)
	}
	// 1. sort nodes with free resource from low to high
	nodesSorted := tp.orderVNodesByFreeResource(nodes)
	if len(nodesSorted) == 0 {
		return fmt.Errorf("dynamic vnpu task<%s> ScoreBestNPUNodes err: sorted nodes len 0", task.Name)
	}

	downgradeNodes, ok := tp.DowngradeCache[task.Name]
	// 2. give the first node high score, none nodes are downgraded
	if !ok {
		value, sOK := scoreMap[nodesSorted[0].Name]
		if !sOK {
			scoreMap[nodesSorted[0].Name] = value
		}
		scoreMap[nodesSorted[0].Name] = value + util.NPUIndex8
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

func (tp *DynamicVNPU) releaseTaskInConCache(task *api.TaskInfo, node plugin.NPUNode) error {
	template, getErr := util.GetVTaskUseTemplate(task)
	if getErr != nil {
		return getErr
	}
	temp, ok := tp.ConCache[node.Name]
	if !ok {
		return fmt.Errorf("node %s not in ConCache", node.Name)
	}
	tIDs, ok := temp[template]
	if !ok {
		return fmt.Errorf("template %s not in %s ConCache", template, node.Name)
	}
	if _, ok := tIDs[task.UID]; !ok {
		return fmt.Errorf("tID %s not in %s %s ConCache", task.UID, template, node.Name)
	}
	delete(tIDs, task.UID)
	if len(tIDs) == 0 {
		delete(temp, template)
		if len(temp) == 0 {
			delete(tp.ConCache, node.Name)
			return nil
		}
		tp.ConCache[node.Name] = temp
		return nil
	}
	temp[template] = tIDs
	tp.ConCache[node.Name] = temp
	return nil
}

// ReleaseAnnotation release Annotation, in dy is release ConCache.
func (tp *DynamicVNPU) ReleaseAnnotation(task *api.TaskInfo, node plugin.NPUNode) *plugin.NPUNode {
	if releaseERR := tp.releaseTaskInConCache(task, node); releaseERR != nil {
		klog.V(util.LogErrorLev).Infof("dynamic %s UseAnnotation UpdateNodeInfo:%s.", task.Name, releaseERR)
	}
	return &node
}

func (tp *DynamicVNPU) addTaskInConCache(task *api.TaskInfo, node plugin.NPUNode, taskResReq util.VResource,
	chipVTemplate VTemplate) error {
	template, getErr := tp.GetTemplateByResReq(taskResReq, chipVTemplate)
	if getErr != nil {
		klog.V(util.LogDebugLev).Infof("IsNodeHasDifferentUnFinishedTask %s", getErr)
		return getErr
	}
	date, ok := tp.ConCache[node.Name]
	if !ok {
		date = make(map[string]map[api.TaskID]struct{}, util.MapInitNum)
	}
	temp, ok := date[template]
	if !ok {
		temp = make(map[api.TaskID]struct{}, util.MapInitNum)
	}
	_, ok = temp[task.UID]
	if ok {
		return nil
	}
	temp[task.UID] = struct{}{}
	date[template] = temp
	tp.ConCache[node.Name] = date
	klog.V(util.LogDebugLev).Infof("addTaskInConCache %s %s ConCache: %#v", node.Name, task.Name, tp.ConCache)
	return nil
}

// UseAnnotation write task use vnpu to pod annotation
func (tp *DynamicVNPU) UseAnnotation(task *api.TaskInfo, node plugin.NPUNode, taskResReq util.VResource,
	chipVTemplate VTemplate) *plugin.NPUNode {
	klog.V(util.LogDebugLev).Infof("dynamic vnpu UseAnnotation node<%s> task<%s> Labels: %#v\n",
		node.Name, task.Name, task.Pod.Labels)

	taskDowngradeNodes, ok := tp.DowngradeCache[task.Name]
	if ok && util.IsSliceContain(node.Name, taskDowngradeNodes) {
		taskResReq = tp.downgradeTaskAICPU(taskResReq)
	}

	allocChipID, err := node.VNode.SelectChipFromNode(taskResReq)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("UseAnnotation dynamic %s on %s err: %s", task.Name, node.Name, err)
		return &node
	}
	klog.V(util.LogDebugLev).Infof("dynamic vnpu UseAnnotation allocChipID:<%s>", allocChipID)

	tp.SetNPUTopologyToPodFn(task, node, taskResReq, allocChipID, chipVTemplate)
	upNode := tp.UpdateNodeInfo(node, allocChipID, taskResReq)
	if addErr := tp.addTaskInConCache(task, *upNode, taskResReq, chipVTemplate); addErr != nil {
		klog.V(util.LogErrorLev).Infof("dynamic vnpu %s UseAnnotation addTaskInConCache:%s", task.Name, addErr)
	}
	return upNode
}

// GetTaskResource get task resource.
func (tp *DynamicVNPU) GetTaskResource(task *api.TaskInfo, node plugin.NPUNode,
	chipVTemplate map[string]util.VResource) (util.VResource,
	error) {
	coreNum, err := getAiCoreNumFromTask(task)
	if err != nil {
		return util.VResource{}, fmt.Errorf("task %s AscendNPUCore read failed", task.Name)
	}

	if node.IsResourceWholeCard(coreNum) {
		res := util.VResource{
			Aicore: coreNum,
			Aicpu:  coreNum * node.TotalRes.Aicpu / node.TotalRes.Aicore,
			DVPP:   plugin.AscendDVPPEnabledNull,
		}
		return res, nil
	}

	cpuLevel, ok := task.Pod.Labels[plugin.AscendVNPULevel]
	if !ok {
		cpuLevel = plugin.AscendVNPULevelLow
	}

	dvpp, ok := task.Pod.Labels[plugin.AscendVNPUDVPP]
	if !ok {
		dvpp = plugin.AscendDVPPEnabledNull
	}

	virTemplate := getResTemplateFromTaskSetting(coreNum, cpuLevel, dvpp)
	return chipVTemplate[virTemplate], nil
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
	allocChipID string, chipVTemplate VTemplate) {
	tmp := strconv.FormatInt(time.Now().UnixNano(), util.Base10)
	task.Pod.Annotations[util.PodPredicateTime] = tmp
	// 1. whole card
	if node.IsResourceWholeCard(taskResReq.Aicore) {
		task.Pod.Annotations[util.AscendNPUCore] = allocChipID
		klog.V(util.LogInfoLev).Infof("dynamic vnpu setNPUTopologyToPod %s top:%s.", task.Name, allocChipID)
		return
	}

	var segmentAnnotation string
	for curTemplate, jobVResource := range chipVTemplate.Data {
		if taskResReq != jobVResource {
			continue
		}
		segmentAnnotation = curTemplate
	}

	task.Pod.Annotations[util.AscendNPUCore] = fmt.Sprintf("%s-%s", allocChipID,
		segmentAnnotation)
	klog.V(util.LogInfoLev).Infof("dynamic vnpu setNPUTopologyToPod %s top:%s.", task.Name,
		task.Pod.Annotations[util.AscendNPUCore])
	return
}

// UpdateNodeInfo vnpu update npuNode after allocation
func (tp *DynamicVNPU) UpdateNodeInfo(node plugin.NPUNode, allocChipID string,
	taskResReq util.VResource) *plugin.NPUNode {
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
