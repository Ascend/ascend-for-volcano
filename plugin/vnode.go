/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package plugin is using for HuaWei Ascend pin affinity schedule.

*/
package plugin

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"k8s.io/api/core/v1"
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

func initTemplate() []util.VTemplate {
	nodeTemplate := make([]util.VTemplate, util.NPUIndex3)
	if len(nodeTemplate) < util.NPUIndex3 {
		return nodeTemplate
	}
	nodeTemplate[0] = util.VTemplate{
		ChipKind: Ascend310P,
		AICore:   util.NPUIndex8,
		AICPU:    util.NPUIndex7,
	}
	nodeTemplate[util.NPUIndex1] = util.VTemplate{
		ChipKind: Ascend910,
		AICore:   util.CoreNum32,
		AICPU:    util.CpuNum14,
	}
	nodeTemplate[util.NPUIndex2] = util.VTemplate{
		ChipKind: Ascend910,
		AICore:   util.CoreNum30,
		AICPU:    util.CpuNum14,
	}
	return nodeTemplate
}

func (n *NPUNode) setNodeVNPUInfo(ni *api.NodeInfo, jobTemplate map[string]map[string]util.VResource) error {
	n.VNode = VNode{
		Chips: make(map[int]*VChip, util.MapInitNum),
	}

	if !n.checkNodeResourceInitialized() {
		return fmt.Errorf("setNodeVNPUInfo %s: npuNode resource not initialized", n.Name)
	}

	// 1. get chipKind like Ascend910, chipLabel like Ascend310P-8
	if err := n.setChipPropertiesFromNPUNode(); err != nil {
		return fmt.Errorf("setNodeVNPUInfo %s: %v", n.Name, err)
	}

	// 2. get resource capacity, totalChipNum, freeChipNum
	if err := n.setTotalResAndChipNumByTemplates(); err != nil {
		return fmt.Errorf("setNodeVNPUInfo node %s: %v", n.Name, err)
	}

	// 3. create vChips on node and update vChip resource
	if err := n.initVChips(ni, jobTemplate); err != nil {
		return fmt.Errorf("setNodeVNPUInfo node %s: %v", n.Name, err)
	}

	n.ValidVNode = true
	return nil
}

func (n *NPUNode) checkNodeResourceInitialized() bool {
	return n.Capability[util.AscendNPUCore] > 0
}

// setChipPropertiesFromNPUNode returns chipKind, chipLabel, accType
func (n *NPUNode) setChipPropertiesFromNPUNode() error {
	chipKind, err := n.getChipKindFromNpuNode() // 1. set ChipKind(Ascend910/Ascend310/Ascend310P)
	if err != nil {
		return fmt.Errorf("setNodeVNPUInfo node %s: %v", n.Name, err)
	}
	n.VNode.ChipKind = chipKind

	chipLabel, ok := n.Label[util.ServerType] // 2. set ServerType(like Ascend310P-10-dual/Ascend910-30)
	if !ok {
		return fmt.Errorf("setNodeVNPUInfo node %s no node label <%s>", n.Name, util.ServerType)
	}
	n.VNode.ServerType = chipLabel

	nodeFreeChips, ok := n.Annotation[util.HwPreName+n.VNode.ChipKind] // 3. set free chip num from device-info
	if !ok {
		return errors.New("getFreeChipNum failed")
	}
	nodeFreeChipsSplit := strings.Split(nodeFreeChips, ",")
	n.VNode.FreeChipNum = len(nodeFreeChipsSplit)

	return nil
}

// getChipKindFromNpuNode input huawei-Ascend910 return Ascend910/Ascend310p/Ascend310
func (n NPUNode) getChipKindFromNpuNode() (string, error) {
	tempVal, ok := n.Label[util.Accelerator]
	if !ok {
		return "", fmt.Errorf("getChipKindFromNpuNode label %s absent", util.Accelerator)
	}
	chipKind := strings.Split(tempVal, "-")
	if len(chipKind) < util.NPUIndex2 {
		return "", fmt.Errorf("getChipKindFromNpuNode label %s value %s format incorrect", util.Accelerator,
			chipKind)
	}
	return chipKind[1], nil
}

// setTotalResAndChipNumByTemplates set totalRes, totalChipNum and serverType
func (n *NPUNode) setTotalResAndChipNumByTemplates() error {
	// 1. get and set total AiCore from capability like Capacity: huawei.com/npu-core: 56
	totalCore, ok := n.Capability[util.AscendNPUCore]
	if !ok {
		return fmt.Errorf("getTotalResFromNpuNode no resource <%s>", util.AscendNPUCore)
	}
	n.VNode.TotalRes.Aicore = int(totalCore)

	numCorePerChip, err := n.getVChipCoreNum()
	if err != nil {
		return fmt.Errorf("getTotalChipNum error: %v", err)
	}
	n.AiCorePerChip = numCorePerChip

	// 2.2 get totalChipNum use totalChipNum = totalCoreNum / coreNumPerChip
	totalChipNum, err := n.getTotalChipNum()
	if err != nil {
		return fmt.Errorf("getTotalResFromNpuNode failed: %v", err)
	}
	n.VNode.TotalChipNum = totalChipNum

	// 2.2 get cpuNum per chip use totalAiCpuNum = aicpuNumPerChip * totalChipNum
	templates := initTemplate()
	cpuPerChip := n.getCpuNumPerChip(templates)
	if cpuPerChip == util.ErrorInt {
		return errors.New("getTotalResFromNpuNode get aicpu from template failed")
	}
	n.VNode.TotalRes.Aicpu = cpuPerChip * totalChipNum
	n.VNode.TotalRes.DVPP = AscendDVPPEnabledOff

	return nil
}

func (n NPUNode) getCpuNumPerChip(templates []util.VTemplate) int {
	cpuPerChip := util.ErrorInt
	for _, temp := range templates {
		if temp.ChipKind != n.VNode.ChipKind || temp.AICore*n.TotalChipNum != n.VNode.TotalRes.Aicore {
			continue
		}
		cpuPerChip = temp.AICPU
	}
	return cpuPerChip
}

func (n *NPUNode) initVChips(ni *api.NodeInfo, taskTemplate map[string]map[string]util.VResource) error {
	chipCoreNum, err := n.VNode.getVChipCoreNum()
	if err != nil {
		return fmt.Errorf("vNode %s get vChip coreNum failed: %v", n.Name, err)
	} // 1. get core number on each chip

	chipTotalRes := n.VNode.getVChipTotalRes()

	if err := n.initFreeWholeVChips(chipCoreNum, chipTotalRes); err != nil {
		return fmt.Errorf("vNode %s initFreeWholeVChips failed", n.Name)
	} // 3. create new VChip by freeCardID whole card

	for _, ti := range ni.Tasks {
		n.VNode.addNPUResource(ti.Pod, chipTotalRes, taskTemplate)
	} // 4. update VChips and create VChips for chips being occupied

	return nil
}

func (n NPUNode) initFreeWholeVChips(chipCoreNum int, chipTotalRes util.VResource) error {
	freeCardIDs := n.getFreeCardIDsFromDeviceInfo()
	if len(freeCardIDs) == 0 {
		return fmt.Errorf("vNode %s getFreeCardIDsFromDeviceInfo failed", n.Name)
	}
	for _, freeCardID := range freeCardIDs {
		n.VNode.Chips[freeCardID] = n.VNode.NewVChip(freeCardID, chipTotalRes)
	}
	return nil
}

func (n *NPUNode) getFreeCardIDsFromDeviceInfo() []int {
	var freeCardIDs []int
	freeChipStr, ok := n.Annotation[util.HwPreName+n.ChipKind] // like Ascend910-0,Ascend910-1
	if !ok {
		return freeCardIDs
	}
	freeChips := strings.Split(freeChipStr, ",")
	for _, freeChip := range freeChips {
		id, err := GetWholeCardIDFromAscendReal(freeChip)
		if err != nil {
			klog.V(util.LogWarningLev).Infof("GetWholeCardIDFromAscendReal error: %v", err)
			continue
		}
		freeCardIDs = append(freeCardIDs, id)
	}
	return freeCardIDs
}

// getTotalChipNum used after aicorePerChip set
func (vNode *VNode) getTotalChipNum() (int, error) {
	totalChipNum := vNode.TotalRes.Aicore / vNode.AiCorePerChip
	if vNode.TotalRes.Aicore%vNode.AiCorePerChip != 0 {
		return 0, errors.New("getTotalChipNum error: total resource cannot be divided by coreNumPerChip")
	}
	if totalChipNum == 0 {
		return 0, errors.New("getTotalChipNum error: total chip number zero")
	}
	return totalChipNum, nil
}

// NewVChip create new vChip
func (vNode *VNode) NewVChip(id int, totalRes util.VResource) *VChip {
	chipName := vNode.ChipKind + "-" + strconv.Itoa(id)
	vChip := VChip{
		PodMap:   make(map[string]*v1.Pod, util.MapInitNum),
		Name:     chipName,
		Kind:     vNode.ChipKind,
		CoreNum:  vNode.AiCorePerChip,
		TotalRes: totalRes,
		FreeRes:  totalRes,
	}
	vChip.TotalRes.DVPP = AscendDVPPEnabledOff
	vChip.UsedRes.DVPP = AscendDVPPEnabledOff
	vChip.FreeRes.DVPP = AscendDVPPEnabledOn

	if strings.HasPrefix(vNode.ServerType, util.ServerTypeDual) {
		vChip.setIsDual(true)
	}

	return &vChip
}

// addNPUResource update all pod resource to node
func (vNode *VNode) addNPUResource(pod *v1.Pod, chipTotalRes util.VResource, taskTemplate map[string]map[string]util.VResource) {
	coreNameStr, ok := pod.Annotations[util.AscendNPUCore]
	if !ok {
		klog.V(util.LogErrorLev).Infof("addNPUResource pod %s %s no value", pod.Name, util.AscendNPUCore)
		return
	}

	if IsPodWholeCardFromAscendCore(coreNameStr) {
		vNode.addNPUResourceWholeCard(pod)
		return
	}
	vNode.addNPUResourceVNPUCard(pod, chipTotalRes, taskTemplate)
}

func (vNode *VNode) getVChipCoreNum() (int, error) {
	serverTypeSplit := strings.Split(vNode.ServerType, "-")
	if len(serverTypeSplit) < util.NPUIndex2 {
		return 0, fmt.Errorf("getVChipCoreNum serverType %s format error", vNode.ServerType)
	}
	coreNum, err := strconv.Atoi(serverTypeSplit[1])
	if err != nil {
		return 0, fmt.Errorf("getVChipCoreNum serverType %s split error", vNode.ServerType)
	}
	return coreNum, nil
}

func (vNode *VNode) getVChipTotalRes() util.VResource {
	AiCore := vNode.TotalRes.Aicore / vNode.TotalChipNum
	AiCpu := vNode.TotalRes.Aicpu / vNode.TotalChipNum
	return util.VResource{
		Aicore: AiCore,
		Aicpu:  AiCpu,
		DVPP:   AscendDVPPEnabledOff,
	}
}

func (vNode *VNode) getPodUsedRes(pod *v1.Pod, taskTemplate map[string]map[string]util.VResource) *util.VResource {
	realStr, ok := pod.Annotations[util.AscendNPUCore]
	if !ok {
		klog.V(util.LogErrorLev).Infof("getPodUsedRes get pod<%s> %s value failed", pod.Name,
			util.AscendNPUCore)
		return nil
	}
	ascendRealSplit := strings.Split(realStr, "-")
	if len(ascendRealSplit) != util.NPUIndex2 {
		klog.V(util.LogErrorLev).Infof("getPodUsedRes get pod<%s> %s format error", pod.Name, realStr)
		return nil
	}
	return GetResourceFromTemplate(vNode.ChipKind, ascendRealSplit[1], taskTemplate)
}

// addNPUResourceWholeCard Ascend910-0,Ascend910-1
func (vNode *VNode) addNPUResourceWholeCard(pod *v1.Pod) {
	physicsID, err := GetCardPhysicsIDFromAscendCore(pod, true)
	if err != nil || len(physicsID) == 0 {
		return
	}
	for _, id := range physicsID {
		// 1. get resource of pod, which is chip total resource
		podVResource := vNode.getVChipTotalRes()

		// 2. get chip id
		curVChip, ok := vNode.Chips[id]
		if !ok {
			curVChip = vNode.NewVChip(id, podVResource)
			vNode.Chips[id] = curVChip
		}

		// 3. update node
		curVChip.Unstable = curVChip.IsPodResUnstable(pod) || curVChip.Unstable
		curVChip.addRealCardID(strconv.Itoa(id))
		curVChip.addPodToPodMap(pod)
		curVChip.UsedRes.Add(podVResource)
		curVChip.FreeRes.Sub(podVResource)
	}
}

// addNPUResourceVNPUCard ascendStr Ascend310P-4c.3cpu.ndvpp-100(VNPUID)-1(physic ID)_1（vgroupID）
func (vNode *VNode) addNPUResourceVNPUCard(pod *v1.Pod, chipTotalRes util.VResource,
	taskTemplate map[string]map[string]util.VResource) {
	// 1. get physics id
	physicsID, err := GetCardPhysicsIDFromAscendCore(pod, false)
	if err != nil || len(physicsID) != util.NPUIndex1 {
		klog.V(util.LogErrorLev).Infof("addNPUResourceVNPUCard get pod<%s> card physics id failed", pod.Name)
		return
	}

	// 2. add chip to node
	curVChip, ok := vNode.Chips[physicsID[0]]
	if !ok {
		curVChip = vNode.NewVChip(physicsID[0], chipTotalRes)
		vNode.Chips[physicsID[0]] = curVChip
	}
	curVChip.Unstable = curVChip.IsPodResUnstable(pod) || curVChip.Unstable
	curVChip.addRealCardID(pod.Annotations[util.AscendNPUPodRealUse])
	curVChip.addPodToPodMap(pod)
	curVChip.setSegmentFlag(true)

	// 3. get resource of pod
	podVResource := vNode.getPodUsedRes(pod, taskTemplate)
	if podVResource == nil {
		klog.V(util.LogErrorLev).Infof("addNPUResource resolving pod<%s> resource failed", pod.Name)
		return
	}
	// 4. update node properties
	curVChip.UsedRes.Add(*podVResource)
	curVChip.FreeRes.Sub(*podVResource)
	curVChip.UpdateDVPP(podVResource.DVPP)
}

// IsPodResUnstable return true if chip stable
func (vChip *VChip) IsPodResUnstable(pod *v1.Pod) bool {
	realStr, ok := pod.Annotations[util.AscendNPUPodRealUse]
	return !ok || realStr == ""
}

func (vChip *VChip) setIsDual(value bool) {
	vChip.IsDual = value
}

func (vChip *VChip) setSegmentFlag(value bool) {
	vChip.SegmentFlag = value
}

func (vChip *VChip) addRealCardID(id string) {
	if id == "" {
		return
	}
	vChip.ID = append(vChip.ID, id)
}

func (vChip *VChip) addPodToPodMap(pod *v1.Pod) {
	vChip.PodMap[string(pod.UID)] = pod
}

// UpdateDVPP update dvpp according to pod resource
func (vChip *VChip) UpdateDVPP(podResDVPP string) {
	if podResDVPP == AscendDVPPEnabledOn {
		vChip.UsedRes.DVPP = AscendDVPPEnabledOn
		vChip.FreeRes.DVPP = AscendDVPPEnabledOff
	}
	if podResDVPP == AscendDVPPEnabledOff && vChip.UsedRes.DVPP == AscendDVPPEnabledOff {
		vChip.UsedRes.DVPP = AscendDVPPEnabledOff
		vChip.FreeRes.DVPP = AscendDVPPEnabledOn
	}
	if podResDVPP == AscendDVPPEnabledNull && vChip.UsedRes.DVPP != AscendDVPPEnabledOn {
		vChip.UsedRes.DVPP = AscendDVPPEnabledNull
		vChip.FreeRes.DVPP = AscendDVPPEnabledNull
	}
}

// IsNodeTotalResEnough judge node total resource enough
func (n NPUNode) IsNodeTotalResEnough(vRes util.VResource) bool {
	var nodeResFree util.VResource
	for _, chip := range n.VNode.Chips {
		if chip.Unstable {
			klog.V(util.LogDebugLev).Infof("chip <%s> unstable, resource exempted", chip.Name)
		}
		nodeResFree.Add(chip.FreeRes)
	}
	return nodeResFree.BeGreater(vRes)
}

// IsNodeChipResEnough judge if chip on node can be allocated to job
func (n NPUNode) IsNodeChipResEnough(vRes util.VResource) bool {
	if n.IsResourceWholeCard(vRes.Aicore) {
		return n.VNode.isNodeChipResEnoughWholeCard(vRes)
	}
	for _, vChip := range n.Chips {
		if !vChip.IsChipMeetResReq(vRes) || vChip.Unstable {
			klog.V(util.LogDebugLev).Infof("vChip %s does not meet resource requirements", vChip.Name)
			continue
		}
		return true
	}
	return false
}

func (vNode VNode) isNodeChipResEnoughWholeCard(vRes util.VResource) bool {
	freeWholeCard := 0
	for _, vChip := range vNode.Chips {
		if vChip.SegmentFlag {
			continue
		}
		freeWholeCard += 1
	}
	return vRes.Aicore/vNode.AiCorePerChip <= freeWholeCard
}

// IsChipMeetResReq check chip resource can be allocated as the task requires
func (vChip *VChip) IsChipMeetResReq(vRes util.VResource) bool {
	if !vChip.isChipResourceEnough(vRes) {
		klog.V(util.LogDebugLev).Infof("vChip %s resource <%#v> not enough", vChip.Name, vChip.FreeRes)
		return false
	}
	if !vChip.isChipVGroupValid(vRes) {
		klog.V(util.LogDebugLev).Infof("vChip %s vGroup not enough", vChip.Name)
		return false
	}
	if !vChip.isChipDVPPValid(vRes) {
		klog.V(util.LogDebugLev).Infof("vChip %s DVPP not enough", vChip.Name)
		return false
	}
	return true
}

// isChipResourceEnough check if core is enough for task
func (vChip *VChip) isChipResourceEnough(vRes util.VResource) bool {
	return vChip.FreeRes.BeGreater(vRes)
}

// isChipVGroupValid check if vGroup is valid
func (vChip *VChip) isChipVGroupValid(vRes util.VResource) bool {
	if vChip.Kind != Ascend310P {
		klog.V(util.LogDebugLev).Infof("not %s task, no need to check vGroup", Ascend310P)
		return true
	}

	if !vChip.SegmentFlag {
		klog.V(util.LogDebugLev).Info("whole card, no need to check vGroup")
		return true
	}

	vGroups := vChip.getVGroups()

	if len(vGroups) == util.NPUIndex3 && vRes.Aicore >= util.NPUIndex4 {
		klog.V(util.LogDebugLev).Infof("%d vGroups, only support 1,2 core", len(vGroups))
		return false
	}

	if len(vGroups) == util.NPUIndex4 && vRes.Aicore >= util.NPUIndex2 {
		klog.V(util.LogDebugLev).Infof("%d vGroups, only support 1 core", len(vGroups))
		return false
	}

	return true
}

func (vChip *VChip) getVGroups() []int {
	var vGroups []int
	for _, realChip := range vChip.ID {
		realChipSplit := strings.Split(realChip, "-")
		if len(realChipSplit) < util.NPUIndex5 {
			continue
		}
		vGroupStr := realChipSplit[util.NPUIndex4]
		vGroup, err := strconv.Atoi(vGroupStr)
		if err != nil {
			continue
		}
		var existFlag bool
		for _, v := range vGroups {
			if vGroup == v {
				existFlag = true
				break
			}
		}
		if !existFlag {
			vGroups = append(vGroups, vGroup)
		}
	}
	return vGroups
}

func (vChip *VChip) isChipDVPPValid(vRes util.VResource) bool {
	// 1. if task dvpp on, the node's free resource must support dvpp
	if vRes.DVPP == AscendDVPPEnabledOn && vChip.FreeRes.DVPP != AscendDVPPEnabledOn {
		return false
	}
	// 2. if task dvpp null, the node's free resource cannot be off
	if vRes.DVPP == AscendDVPPEnabledNull && vChip.FreeRes.DVPP == AscendDVPPEnabledOff {
		return false
	}
	// 3. if task dvpp no, the node's free resource can be any
	return true
}

// SelectChipFromNode get chip with least resource that meets vRes requirements
func (vNode *VNode) SelectChipFromNode(vRes util.VResource) (string, error) {
	var Chips []*VChip
	for _, Chip := range vNode.Chips {
		Chips = append(Chips, Chip)
	}

	tempVChips := vChipsList(Chips)
	sort.Sort(tempVChips)
	if len(tempVChips) == 0 {
		return "", fmt.Errorf("SelectChipFromNode sorted chips len 0")
	}

	if vNode.IsResourceWholeCard(vRes.Aicore) {
		return vNode.selectChipFromNodeWhole(tempVChips, vRes)
	}
	return vNode.selectChipFromNodeSegment(tempVChips, vRes)
}

func (vNode *VNode) selectChipFromNodeSegment(vChip []*VChip, vRes util.VResource) (string, error) {
	for _, chip := range vChip {
		if !chip.IsChipMeetResReq(vRes) || chip.Unstable {
			klog.V(util.LogDebugLev).Infof("chip %s does not meet resource requirements", chip.Name)
			continue
		}
		chipID, err := GetWholeCardIDFromAscendReal(chip.Name)
		if err != nil {
			return "", fmt.Errorf("selectChipFromNodeSegment chip name <%s> err: %s", chip.Name, err)
		}
		return strconv.Itoa(chipID), nil
	}

	return "", fmt.Errorf("selectChipFromNodeSegment available chip not found for req <%d>", vRes.Aicore)
}

func (vNode *VNode) selectChipFromNodeWhole(vChips []*VChip, vRes util.VResource) (string, error) {
	reqCardNum := vRes.Aicore / vNode.AiCorePerChip
	allocCardNum := 0
	vResChip := util.VResource{
		Aicore: vRes.Aicore / reqCardNum,
		Aicpu:  vRes.Aicpu / reqCardNum,
		DVPP:   AscendDVPPEnabledNull,
	}
	var cardNames []string
	for _, chip := range vChips {
		if !chip.IsChipMeetResReq(vResChip) || chip.SegmentFlag {
			klog.V(util.LogDebugLev).Infof("chip %s does not meet whole card resource requirements", chip.Name)
			continue
		}
		chipID, err := GetWholeCardIDFromAscendReal(chip.Name)
		if err != nil {
			return "", fmt.Errorf("selectChipFromNodeWhole chip name <%s> err: %s", chip.Name, err)
		}
		cardNames = append(cardNames, strconv.Itoa(chipID))
		allocCardNum += 1
		if allocCardNum == reqCardNum {
			return strings.Join(cardNames, ","), nil
		}
	}
	return "", fmt.Errorf("selectChipFromNodeWhole free whole chip <%d> not enough for req <%d>", allocCardNum,
		reqCardNum)
}

// IsResourceWholeCard judge if resource is whole card by node total resource
func (vNode *VNode) IsResourceWholeCard(aiCore int) bool {
	chipCoreNum, err := vNode.getVChipCoreNum()
	if err != nil {
		klog.V(util.LogWarningLev).Infof("IsResourceWholeCard get chipCoreNum failed")
		return false
	}
	return aiCore%chipCoreNum == 0
}

type vChipsList []*VChip

// Len for order.
func (vChips vChipsList) Len() int {
	return len(vChips)
}

// Less for order.
func (vChips vChipsList) Less(i, j int) bool {
	if i > vChips.Len() || j > vChips.Len() {
		return false
	}
	return !vChips[i].FreeRes.BeGreater(vChips[j].FreeRes)
}

// Swap for order.
func (vChips vChipsList) Swap(i, j int) {
	if i > vChips.Len() || j > vChips.Len() {
		return
	}
	vChips[i], vChips[j] = vChips[j], vChips[i]
}
