/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package plugin is using for HuaWei Ascend pin affinity schedule.
*/
package plugin

import (
	"fmt"
	"strconv"
	"strings"

	"k8s.io/api/core/v1"
	"k8s.io/klog"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

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
