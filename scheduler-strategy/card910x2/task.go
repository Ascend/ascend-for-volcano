/*
Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/*

Package card910x2 is using for HuaWei A300T Ascend pin affinity schedule.

*/
package card910x2

import "errors"

// According to need card number, get best node from 4 pri-node-list-group.
func getBestNodesMap(priNodeGroups []map[string]*npuPriNodeInf) (map[string]int, error) {
	var bestNodesMap = make(map[string]int, constIntNum2)

	for i := 0; i < constIntNum2; i++ {
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