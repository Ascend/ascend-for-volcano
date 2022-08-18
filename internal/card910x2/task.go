/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package card910x2 is using for HuaWei A300T Ascend pin affinity schedule.

*/
package card910x2

import (
	"errors"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/util"
)

// According to need card number, get best node from 4 pri-node-list-group.
func getBestNodesMap(priNodeGroups []map[string]*npuPriNodeInf) (map[string]int, error) {
	var bestNodesMap = make(map[string]int, util.NPUIndex2)

	for i := 0; i < len(priNodeGroups); i++ {
		for nodeName := range priNodeGroups[i] {
			tmpName := nodeName
			bestNodesMap[tmpName] = i
		}
	}

	if len(bestNodesMap) == 0 {
		return nil, errors.New("none bestNodes")
	}

	return bestNodesMap, nil
}
