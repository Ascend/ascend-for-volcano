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
	"strconv"
	"strings"

	"k8s.io/api/core/v1"
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

func newTemplate(kind, dvpp string, core, cpu int) util.VTemplate {
	return util.VTemplate{
		ChipKind:   kind,
		AICore:     core,
		AICPU:      cpu,
		DVPPEnable: dvpp,
	}
}

func initVJobTemplate() map[string]util.VTemplate {
	templateMap := make(map[string]util.VTemplate, util.MapInitNum)
	templateMap[RingController910+"-"+strconv.Itoa(util.NPUIndex2)+"-"+AscendVNPULevelLow] =
		newTemplate(Ascend910, AscendDVPPEnabledNull, util.NPUIndex2, util.NPUIndex1)
	templateMap[RingController910+"-"+strconv.Itoa(util.NPUIndex4)+"-"+AscendVNPULevelLow] =
		newTemplate(Ascend910, AscendDVPPEnabledNull, util.NPUIndex4, util.NPUIndex1)
	templateMap[RingController910+"-"+strconv.Itoa(util.NPUIndex8)+"-"+AscendVNPULevelLow] =
		newTemplate(Ascend910, AscendDVPPEnabledNull, util.NPUIndex8, util.NPUIndex3)
	templateMap[RingController910+"-"+strconv.Itoa(util.NPUIndex16)+"-"+AscendVNPULevelLow] =
		newTemplate(Ascend910, AscendDVPPEnabledNull, util.NPUIndex16, util.NPUIndex7)
	templateMap[RingController910+"-"+strconv.Itoa(util.CoreNum32)+"-"+AscendVNPULevelLow] =
		newTemplate(Ascend910, AscendDVPPEnabledNull, util.CoreNum32, util.CpuNum14)
	templateMap[RingController910+"-"+strconv.Itoa(util.CoreNum30)+"-"+AscendVNPULevelLow] =
		newTemplate(Ascend910, AscendDVPPEnabledNull, util.CoreNum30, util.CpuNum14)
	templateMap[RingController310P+"-"+strconv.Itoa(util.NPUIndex1)] = newTemplate(Ascend310P, AscendDVPPEnabledNull,
		util.NPUIndex1, util.NPUIndex1)
	templateMap[RingController310P+"-"+strconv.Itoa(util.NPUIndex2)] = newTemplate(Ascend310P, AscendDVPPEnabledNull,
		util.NPUIndex2, util.NPUIndex2)
	templateMap[RingController310P+"-"+strconv.Itoa(util.NPUIndex4)] = newTemplate(Ascend310P, AscendDVPPEnabledNull,
		util.NPUIndex4, util.NPUIndex4)
	templateMap[RingController310P+"-"+strconv.Itoa(util.NPUIndex2)+"-"+AscendVNPULevelLow] = newTemplate(Ascend310P,
		AscendDVPPEnabledNull, util.NPUIndex2, util.NPUIndex1)
	templateMap[RingController310P+"-"+strconv.Itoa(util.NPUIndex4)+"-"+AscendVNPULevelLow] = newTemplate(Ascend310P,
		AscendDVPPEnabledNull, util.NPUIndex4, util.NPUIndex3)
	templateMap[RingController310P+"-"+strconv.Itoa(util.NPUIndex4)+"-"+AscendVNPULevelLow] = newTemplate(Ascend310P,
		AscendDVPPEnabledNull, util.NPUIndex4, util.NPUIndex3)
	templateMap[RingController310P+"-"+strconv.Itoa(util.NPUIndex8)+"-"+AscendVNPULevelLow] = newTemplate(Ascend310P,
		AscendDVPPEnabledNull, util.NPUIndex8, util.NPUIndex7)
	return templateMap
}

// GetResourceFromRealStr vDeviceResourceStr like 4c.3cpu.ndvpp
func GetResourceFromRealStr(vDeviceResourceStr string) *util.VResource {
	klog.V(util.LogInfoLev).Infof("GetResourceFromRealStr parsing resource")
	resources := strings.Split(vDeviceResourceStr, ".") // like 4c.3cpu.ndvpp/2c.1cpu/8c

	// 1. get coreNum from template
	aicoreNum := getAicoreFromRealStr(resources) // like 4c
	if aicoreNum == util.ErrorInt {
		klog.V(util.LogErrorLev).Infof("%s aicore %s", vDeviceResourceStr, FormatIncorrectError)
		return nil
	}

	// 2. get aicpu from template
	aicpuNum := getAicpuFromRealStr(resources, aicoreNum) // like 4c.3cpu
	if aicpuNum == util.ErrorInt {
		klog.V(util.LogDebugLev).Infof("%s aicpu %s", vDeviceResourceStr, FormatIncorrectError)
		return nil
	}

	dvppValue := getDvppFromRealOrCoreStr(resources)
	return &util.VResource{
		Aicore: aicoreNum,
		Aicpu:  aicpuNum,
		DVPP:   dvppValue,
	}
}

func getAicoreFromRealStr(resources []string) int {
	if len(resources) < 1 {
		klog.V(util.LogErrorLev).Infof("%v resource %s", resources, FormatIncorrectError)
		return util.ErrorInt
	}

	if !strings.HasSuffix(resources[0], "c") {
		klog.V(util.LogErrorLev).Infof("%s aicore %s", resources[0], FormatIncorrectError)
		return util.ErrorInt
	}

	aicoreNum, err := strconv.Atoi(strings.TrimSuffix(resources[0], "c")) // like 4c
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s aicore %s", resources[0], FormatIncorrectError)
		return util.ErrorInt
	}
	return aicoreNum
}

func getAicpuFromRealStr(resources []string, aicoreNum int) int {
	if len(resources) < util.NPUIndex2 { // 2.1 cpu==core
		klog.V(util.LogDebugLev).Infof("high cpu requirements")
		return aicoreNum
	}
	aicpuNum, err := strconv.Atoi(strings.TrimSuffix(resources[1], util.AICPU)) // 2.2 cpu<core
	if err != nil {
		klog.V(util.LogDebugLev).Infof("aicpu format error")
		return util.ErrorInt
	}

	return aicpuNum
}

func getDvppFromRealOrCoreStr(resources []string) string {
	if len(resources) < util.NPUIndex3 {
		return AscendDVPPEnabledNull
	}

	// 3. get dvpp from template
	var dvppValue string
	switch resources[util.NPUIndex2] {
	case AscendDVPPValue:
		dvppValue = AscendDVPPEnabledOn
	case AscendNDVPPValue:
		dvppValue = AscendDVPPEnabledOff
	default:
		dvppValue = AscendDVPPEnabledNull
	}
	return dvppValue
}

// GetResourceFromCoreStr like vir04_3c_ndvpp
func GetResourceFromCoreStr(coreStr string) *util.VResource {
	resources := strings.Split(coreStr, "_")

	// 1. get coreNum from template
	aicoreNum := getAiCoreFromCoreStr(resources) // like vir04
	if aicoreNum == util.ErrorInt {
		klog.V(util.LogErrorLev).Infof("%s aicore %s", coreStr, FormatIncorrectError)
		return nil
	}

	// 2. get aicpu from template
	aicpuNum := getAiCpuFromCoreStr(resources, aicoreNum) // like vir04_3c
	if aicpuNum == util.ErrorInt {
		klog.V(util.LogDebugLev).Infof("%s aicpu %s", coreStr, FormatIncorrectError)
		return nil
	}

	dvppValue := getDvppFromRealOrCoreStr(resources) // like vir04_3c_ndvpp
	return &util.VResource{
		Aicore: aicoreNum,
		Aicpu:  aicpuNum,
		DVPP:   dvppValue,
	}
}

func getAiCoreFromCoreStr(resources []string) int {
	if len(resources) < 1 {
		klog.V(util.LogErrorLev).Infof("%v resource %s", resources, FormatIncorrectError)
		return util.ErrorInt
	}

	if !strings.HasPrefix(resources[0], AscendVNPUPrefix) {
		klog.V(util.LogErrorLev).Infof("%s aicore %s", resources[0], FormatIncorrectError)
		return util.ErrorInt
	}

	aicoreNum, err := strconv.Atoi(strings.TrimPrefix(resources[0], AscendVNPUPrefix)) // like 4c
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s aicore %s", resources[0], FormatIncorrectError)
		return util.ErrorInt
	}
	return aicoreNum
}

func getAiCpuFromCoreStr(resources []string, aicoreNum int) int {
	if len(resources) < util.NPUIndex2 { // 2.1 cpu==core
		klog.V(util.LogDebugLev).Infof("high cpu requirements")
		return aicoreNum
	}
	aicpuNum, err := strconv.Atoi(strings.TrimSuffix(resources[1], "c")) // 2.2 cpu<core
	if err != nil {
		klog.V(util.LogDebugLev).Infof("aicpu format error")
		return util.ErrorInt
	}

	return aicpuNum
}

// IsPodWholeCardFromAscendCore judge if card is whole card 0,1/0-vir04
func IsPodWholeCardFromAscendCore(coreCardName string) bool {
	temp := strings.Split(coreCardName, ",")
	for _, cardName := range temp {
		singleCardTemp := strings.Split(cardName, "-")
		if len(singleCardTemp) == 1 {
			return true
		}
	}
	return false
}

// GetWholeCardIDFromAscendReal get card physics id from Ascend910-0
func GetWholeCardIDFromAscendReal(cardNameStr string) (int, error) {
	idStr := strings.Split(cardNameStr, "-")
	if len(idStr) < util.NPUIndex2 {
		return util.ErrorInt, fmt.Errorf("getCardIDFromCardNameStr %s format incorrect", cardNameStr)
	}
	id, err := strconv.Atoi(idStr[util.NPUIndex1])
	if err != nil {
		return util.ErrorInt, fmt.Errorf("getCardIDFromCardNameStr %s %v", cardNameStr, err)
	}
	return id, nil
}

// GetCardPhysicsIDFromAscendCore get card physics id from 0,1/0-vir04
func GetCardPhysicsIDFromAscendCore(pod *v1.Pod, isWholeCard bool) ([]int, error) {
	var physicsIDs []int
	coreNameStr, ok := pod.Annotations[util.AscendNPUCore]
	if !ok {
		return physicsIDs, fmt.Errorf("GetCardPhysicsIDFromAscendCore vnpu device <%s> get %s value failed",
			pod.Name, util.AscendNPUCore)
	}

	if !isWholeCard {
		phyCardID, err := getVNPUCardIDFromAscendCore(coreNameStr)
		if err != nil {
			return physicsIDs, fmt.Errorf("GetCardPhysicsIDFromAscendCore vnpu device <%s> get id failed",
				coreNameStr)
		}
		physicsIDs = append(physicsIDs, phyCardID)
		return physicsIDs, nil
	}
	coreNameSplit := strings.Split(coreNameStr, ",")
	for _, id := range coreNameSplit {
		phyCardID, err := strconv.Atoi(id)
		if err != nil {
			return physicsIDs, fmt.Errorf("GetCardPhysicsIDFromAscendCore device <%s> get physics id failed",
				coreNameStr)
		}
		physicsIDs = append(physicsIDs, phyCardID)
	}
	return physicsIDs, nil
}

func getVNPUCardIDFromAscendCore(coreNameStr string) (int, error) {
	coreNameSplit := strings.Split(coreNameStr, "-")
	if len(coreNameSplit) != util.NPUIndex2 {
		return 0, fmt.Errorf("getVNPUCardIDFromAscendCore vnpu real device <%s> format error", coreNameStr)
	}
	phyCardID, err := strconv.Atoi(coreNameSplit[0])
	if err != nil {
		return 0, fmt.Errorf("getVNPUCardIDFromAscendCore vnpu device <%s> get physics id failed", coreNameStr)
	}
	return phyCardID, nil
}

// TransferTaskLabelToResReq transfer 4c.3cpu.ndvpp to resource
func (vNode *VNode) TransferTaskLabelToResReq(task *api.TaskInfo) (util.VResource, error) {
	resReq := util.VResource{
		DVPP: AscendDVPPEnabledNull,
	}
	coreNum, err := getAiCoreNumFromTask(task)
	if err != nil {
		return resReq, fmt.Errorf("task %s AscendNPUCore read failed", task.Name)
	}

	cpuNum, err := vNode.getAiCpuNum(task, coreNum)
	if err != nil {
		return resReq, fmt.Errorf("task %s AscendCPUNum get failed", task.Name)
	}

	dvppVal := vNode.getDVPPEnable(task, coreNum)

	resReq = util.VResource{
		Aicore: coreNum,
		Aicpu:  cpuNum,
		DVPP:   dvppVal,
	}

	return resReq, nil
}

func getAiCoreNumFromTask(task *api.TaskInfo) (int, error) {
	for _, container := range task.Pod.Spec.Containers {
		coreNum, ok := container.Resources.Requests[util.AscendNPUCore]
		if !ok {
			return 0, errors.New("getAiCoreNumFromTask get resource requests failed")
		}
		return int(coreNum.Value()), nil
	}
	return 0, fmt.Errorf("getAiCoreNumFromTask get resource requests failed")
}

func (vNode *VNode) getDVPPEnable(task *api.TaskInfo, coreNum int) string {
	ringController := task.Pod.Labels[util.JobKindKey]
	dvppVal, ok := task.Pod.Labels[AscendVNPUDVPP]
	if !ok || ringController != RingController310P || vNode.IsResourceWholeCard(coreNum) {
		return AscendDVPPEnabledNull
	}
	return dvppVal
}

func (vNode *VNode) getAiCpuNum(task *api.TaskInfo, coreNum int) (int, error) {
	if vNode.IsResourceWholeCard(coreNum) {
		return coreNum * vNode.TotalRes.Aicpu / vNode.TotalRes.Aicore, nil
	}
	ringControllerType, ok := task.Pod.Labels[util.JobKindKey] // ascend-910/ascend-310P/ascend-310
	if !ok {
		return util.ErrorInt, fmt.Errorf("getAiCpuNum get label %s failed", util.JobKindKey)
	}

	vnpuLevel, ok := task.Pod.Labels[AscendVNPULevel]
	if !ok || ringControllerType != RingController310P {
		vnpuLevel = AscendVNPULevelLow
	}
	key := fmt.Sprintf("%s-%s", ringControllerType, strconv.Itoa(coreNum))
	if vnpuLevel == AscendVNPULevelLow {
		key = fmt.Sprintf("%s-%s", key, vnpuLevel)
	}
	vJobTemplate := initVJobTemplate()
	vTemplate, ok := vJobTemplate[key]
	if !ok {
		return util.ErrorInt, fmt.Errorf("getAiCpuNum get %s from vJob template failed", key)
	}
	return vTemplate.AICPU, nil
}
