/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package card310x4 is using for HuaWei A300T Ascend pin affinity schedule.

*/
package card310x4

import (
	"errors"
	"fmt"
	"reflect"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/util"
)

func initNodesNPUTopologyFn(nodes map[string]*api.NodeInfo) error {
	for _, tmpNode := range nodes {
		topStr, err := util.GetNPUAllocCardsFromNodeDeviceInfoCache(tmpNode, a310NPUCardName)
		if err != nil {
			klog.V(util.LogDebugLev).Infof("%s initNodesFn :%v", PluginName, err)
			return nil
		}
		err = util.SaveTopologyInMap(tmpNode.Others, topStr, a310NPUCardName)
		if err != nil {
			return err
		}
	}
	klog.V(util.LogDebugLev).Infof("All nodes are initialized successfully")
	return nil
}

func getNodeNPUNumFromOthers(nodeInfo *api.NodeInfo) (int, error) {
	top := util.GetTopFromNodeOthers(nodeInfo, a310NPUCardName, a310NPUCardPreName)
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

// initPriNodeGroups the priority group of the node.
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
			klog.V(util.LogDebugLev).Infof("%s initPriNodeGroups get node nil.", PluginName)
			continue
		}
		cardIds := util.GetTopFromNodeOthers(node, a310NPUCardName, a310NPUCardPreName)
		if cardIds == nil {
			klog.V(util.LogDebugLev).Infof("%s initPriNodeGroups [%s] get node top nil.", PluginName, node.Name)
			continue
		}
		cardNumGroups := getCardNumGroupsFromTop(cardIds)
		// set the meet node into its pri-node-list group
		addPriNodeGroupFn := func(priNodeGroup map[string]*npuPriNodeInf, groupName string) {
			klog.V(util.LogDebugLev).Infof("%s [%s],group:%v.", PluginName, node.Name, priNodeGroup[node.Name])
			priNodeGroup[node.Name] = &npuPriNodeInf{
				Name:     groupName,
				nodeName: node.Name,
			}
			klog.V(util.LogDebugLev).Infof("%s addPriNodeGroupFn node name:%s priNode:%v.",
				PluginName, node.Name, priNodeGroup[node.Name])
		}
		// insert into group by policy
		err = insertNodeInPriGroup(task, cardNumGroups, priNodeGroups, addPriNodeGroupFn)
		if err != nil {
			klog.V(util.LogDebugLev).Infof("%s insertNodeInPriGroup %v.", PluginName, err)
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
	cardNumGroups := make([][]int, maxCardNum/util.NPUIndex4+1, maxCardNum/util.NPUIndex4+1)
	for _, v := range nodeNPUTopology {
		index := v / util.NPUIndex4
		if index > len(cardNumGroups)-1 {
			continue
		}
		cardNumGroups[index] = append(cardNumGroups[index], v)
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
