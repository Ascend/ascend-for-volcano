/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package comvnpu is using for virtual HuaWei vnpu schedule.

*/
package comvnpu

import (
	"strings"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/vnpu/vnpuutil"
)

// InitVNodesFn init node.
func (tp *VNPU) InitVNodesFn(nodes map[string]*api.NodeInfo) error {
	for _, tmpNode := range nodes {
		anno := tmpNode.Node.Annotations
		for typeKey := range anno {
			if !strings.Contains(typeKey, vnpuutil.NPUIdentifyName) {
				continue
			}
			nTopStr, err := util.GetResourceFromAnnotationFn(anno, typeKey)
			if err != nil {
				nTopStr = ""
			}
			err = util.SaveTopologyInMap(tmpNode.Others, nTopStr, typeKey)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// UpdateNPUNodeUsedCardFn update node others after allocate
func (tp *VNPU) UpdateNPUNodeUsedCardFn(node *api.NodeInfo, _ interface{}) error {
	klog.V(util.LogDebugLev).Infof("%s UpdateNPUNodeUsedCardFn %s, no need.", tp.Name(), node.Name)
	return nil
}

// IsMyNode used for identify Vnpu node, need to be implemented by vNPU plugins
func (tp *VNPU) IsMyNode(vNode *api.NodeInfo) error {
	klog.V(util.LogDebugLev).Infof("%s IsMyNode %s, no need.", tp.Name(), vNode.Name)
	return nil
}

// CheckNPUResourceStableFn check whether the resources on the node are stable
func (tp *VNPU) CheckNPUResourceStableFn(node *api.NodeInfo) error {
	klog.V(util.LogDebugLev).Infof("%s CheckNPUResourceStableFn %s, no need.", tp.Name(), node.Name)
	return nil
}

// UpdateReleaseNPUNodeTopologyFn update node others after release
func (tp *VNPU) UpdateReleaseNPUNodeTopologyFn(node *api.NodeInfo, _ interface{}) error {
	klog.V(util.LogDebugLev).Infof("%s UpdateReleaseNPUNodeTopologyFn %s, no need.", tp.Name(), node.Name)
	return nil
}
