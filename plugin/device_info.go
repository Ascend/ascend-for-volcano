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
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// GetResourceFromTemplate nodeType like Ascend310P, templateString like "vir04_3c_ndvpp"
func GetResourceFromTemplate(nodeType string, templateString string, taskTemplate map[string]map[string]util.VResource) *util.VResource {
	taskNodeTemplate, ok := taskTemplate[nodeType]
	if !ok {
		return nil
	}
	taskResource := taskNodeTemplate[templateString]
	if !ok {
		return nil
	}
	return &taskResource
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
