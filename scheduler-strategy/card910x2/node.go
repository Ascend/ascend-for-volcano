/*
Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package card910x2 is using for HuaWei A300T Ascend pin affinity schedule.

*/
package card910x2

import (
	"errors"
	"fmt"
	"k8s.io/klog"
	"reflect"
	"volcano.sh/volcano/pkg/scheduler/api"
	hwutil "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/util"
)

func initNodesNPUTopologyFn(nodes map[string]*api.NodeInfo) error {
	for key := range nodes {
		if !hwutil.IsCardModeNode(nodes[key]) {
			continue
		}

		topStr, err := hwutil.GetNPUAllocCardsFromNodeAnnotations(nodes[key], a300TNPUCardName)
		if err != nil {
			klog.V(logDebugLev).Infof("%s initNodesFn :%v", PluginName, err)
			return nil
		}

		err = hwutil.SaveTopologyInMap(nodes[key].Others, topStr, a300TNPUCardName)
		if err != nil {
			return err
		}
	}

	return nil
}

func getNodeNPUNumFromOthers(nodeInfo *api.NodeInfo) (int, error) {
	top := hwutil.GetTopFromNodeOthers(nodeInfo, a300TNPUCardName, a300tNPUCardPreName)
	if top == nil {
		return 0, fmt.Errorf("nil node(%s) top", nodeInfo.Name)
	}

	nodeNPUIdleNumFromTop := len(top)
	if nodeNPUIdleNumFromTop > maxNPUNum {
		return 0, fmt.Errorf("amount of npus exceeded limitation, maximum(%d), actual(%d)",
			maxNPUNum, nodeNPUIdleNumFromTop)
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

	for i := 0; i <= constIntNum2; i++ {
		priNodeGroups = append(priNodeGroups, make(map[string]*npuPriNodeInf, 1))
	}

	// init pri Node group
	for _, node := range nodes {
		if reflect.ValueOf(node).IsNil() {
			continue
		}

		cardIds := hwutil.GetTopFromNodeOthers(node, a300TNPUCardName, a300tNPUCardPreName)
		if cardIds == nil {
			klog.V(logDebugLev).Infof("%s initPriNodeGroups [%s] get node top nil.", PluginName, node.Name)
			continue
		}
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
		err = insertNodeInPriGroup(task, cardIds, priNodeGroups, addPriNodeGroupFn)
		if err != nil {
			continue
		}
	}
	return priNodeGroups, nil
}
