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

Package chip310x4 is using for HuaWei A300T Ascend pin affinity schedule.

*/
package chip310x4

import (
	"fmt"
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/util"
)

func initNodesNPUTopologyFn(nodes map[string]*api.NodeInfo) error {
	for key := range nodes {
		topStr, err := util.GetNPUAllocCardsFromNodeAnnotations(nodes[key], a310NPUChipName)
		if err != nil {
			klog.V(logDebugLev).Infof("%s initNodesFn :%v", PluginName, err)
			return nil
		}
		err = util.SaveTopologyInMap(nodes[key].Others, topStr, a310NPUChipName)
		if err != nil {
			return err
		}
	}
	klog.V(logDebugLev).Infof("All nodes are initialized successfully")
	return nil
}

func getNodeNPUNumFromOthers(nodeInfo *api.NodeInfo) (int, error) {
	top := util.GetTopFromNodeOthers(nodeInfo, a310NPUChipName, a310NPUCardPreName)
	if top == nil {
		return 0, fmt.Errorf("nil node(%s) top", nodeInfo.Name)
	}

	nodeNPUIdleNumFromTop := len(top)
	if nodeNPUIdleNumFromTop > nodeNPUNumber {
		return 0, fmt.Errorf("amount of npus exceeded limitation, maximum(%d), actual(%d)",
			nodeNPUNumber, nodeNPUIdleNumFromTop)
	}

	return nodeNPUIdleNumFromTop, nil
}

// getCardNumGroupsFromTop get the chip for each card from nodeTop
func getCardNumGroupsFromTop(nodeNPUTopology []int) [][]int {
	maxCardNum := 0
	for _, v := range nodeNPUTopology {
		maxCardNum = max(maxCardNum, v)
	}
	cardNumGroups := make([][]int, maxCardNum/4+1, maxCardNum/4+1)
	for _, v := range nodeNPUTopology {
		cardNumGroups[v/4] = append(cardNumGroups[v/4], v)
	}
	return cardNumGroups
}

func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}
