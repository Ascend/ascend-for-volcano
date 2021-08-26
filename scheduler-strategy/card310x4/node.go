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

Package card310x4 is using for HuaWei A300T Ascend pin affinity schedule.

*/
package card310x4

import (
	"errors"
	"fmt"
	"k8s.io/klog"
	"reflect"
	"volcano.sh/volcano/pkg/scheduler/api"
	hwutil "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/card310x4/util"
)

func initNodesNPUTopologyFn(nodes map[string]*api.NodeInfo) error {
	for _, node := range nodes {
		topStr, err := hwutil.GetNPUAllocCardsFromNodeAnnotation(node, a310NPUCardName)
		if err != nil {
			klog.V(logDebugLev).Infof("%s initNodesFn :%v", PluginName, err)
			return nil
		}
		if node.Others == nil {
			node.Others = make(map[string]interface{}, 1)
		}
		err = hwutil.SaveTopologyInMap(node.Others, topStr, a310NPUCardName)
		if err != nil {
			return err
		}
	}
	klog.V(logDebugLev).Infof("All nodes are initialized successfully")
	return nil
}

func getNodeNPUNumFromOthers(nodeInfo *api.NodeInfo) (int, error) {
	top := hwutil.GetTopFromNodeOthers(nodeInfo, a310NPUCardName, a310NPUCardPreName)
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

// Initializes the priority group of the node.
func initPriNodeGroups(task *api.TaskInfo, nodes []*api.NodeInfo) ([]map[string]*npuPriNodeInf, error) {
	var err error
	var priNodeGroups []map[string]*npuPriNodeInf
	// for pipelined state the node npu is nil
	if len(nodes) == 0 {
		return nil, errors.New("nodes is empty")
	}
	for i := 0; i < cardNPUNumber; i++ {
		priNodeGroups = append(priNodeGroups, make(map[string]*npuPriNodeInf, 1))
	}
	// init pri Node group
	for _, node := range nodes {
		if reflect.ValueOf(node).IsNil() {
			continue
		}
		cardIds := hwutil.GetTopFromNodeOthers(node, a310NPUCardName, a310NPUCardPreName)
		if cardIds == nil {
			klog.V(logDebugLev).Infof("%s initPriNodeGroups [%s] get node top nil.", PluginName, node.Name)
			continue
		}
		cardNumGroups := getCardNumGroupsFromTop(cardIds)
		// set the meet node into its pri-node-list group
		addPriNodeGroupFn := func(priNodeGroup map[string]*npuPriNodeInf, groupName string) {
			klog.V(logDebugLev).Infof("%s [%s],group:%v.", PluginName, node.Name, priNodeGroup[node.Name])
			priNodeGroup[node.Name] = &npuPriNodeInf{
				Name:     groupName,
				nodeName: node.Name,
			}
			klog.V(logDebugLev).Infof("%s addPriNodeGroupFn node name:%s priNode:%v.",
				PluginName, node.Name, priNodeGroup[node.Name])
		}
		// insert into group by policy
		err = insertNodeInPriGroup(task, cardNumGroups, priNodeGroups, addPriNodeGroupFn)
		if err != nil {
			continue
		}
	}
	return priNodeGroups, nil
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

// Max returns the larger of a and b.
func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

// Min returns the smaller of a and b.
func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
