/*
Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package common is using for HuaWei common infer Ascend pin affinity schedule.

*/
package common

import (
	"fmt"
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/util"
)

func (cn *Scheduler) initNodesNPUTopologyFn(nodes map[string]*api.NodeInfo) error {
	for key := range nodes {
		topStr, err := util.GetNPUAllocCardsFromNodeAnnotations(nodes[key], cn.AnnoName)
		if err != nil {
			klog.V(LogDebugLev).Infof("%s initNodesFn :%v", cn.PluginName, err)
			return nil
		}
		err = util.SaveTopologyInMap(nodes[key].Others, topStr, cn.AnnoName)
		if err != nil {
			return err
		}
	}
	klog.V(LogDebugLev).Infof("All nodes are initialized successfully")
	return nil
}

func (cn *Scheduler) getNodeNPUNumFromOthers(nodeInfo *api.NodeInfo) (int, error) {
	top := util.GetTopFromNodeOthers(nodeInfo, cn.AnnoName, cn.AnnoPreVal)
	if top == nil {
		return 0, fmt.Errorf("nil node(%s) top", nodeInfo.Name)
	}

	nodeNPUIdleNumFromTop := len(top)
	if nodeNPUIdleNumFromTop > NodeNPUNumber {
		return 0, fmt.Errorf("amount of npus exceeded limitation, maximum(%d), actual(%d)",
			NodeNPUNumber, nodeNPUIdleNumFromTop)
	}

	return nodeNPUIdleNumFromTop, nil
}
