/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package comvnpu is using for virtual HuaWei vnpu schedule.

*/
package comvnpu

import (
	"errors"
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

// JudgeResourceTypeByTopInfo Judge resource type.
func (tp *VNPU) JudgeResourceTypeByTopInfo(instance string) string {
	var vType string
	for _, vt := range tp.Attr.DivideKinds {
		v := strings.TrimPrefix(vt, tp.Attr.AnnoPreVal)
		if strings.HasPrefix(instance, v) {
			vType = vt
			break
		}
	}

	return vType
}

func getTopStrFromNodeOther(othersMap map[string]interface{}, npuCardName string) ([]string, error) {
	var topArr []string

	valueTmp, ok := othersMap[npuCardName]
	if !ok {
		klog.V(util.LogDebugLev).Infof("getNodeNPUStrFromOther %s not in node other.", npuCardName)
		return nil, errors.New("nodeTopStrArr nil")
	}

	mapStr, ok := valueTmp.(string)
	if !ok {
		klog.V(util.LogErrorLev).Infof("%s getNodeNPUStrFromOther not string type.", npuCardName)
		return nil, errors.New("nodeTopStrArr nil")
	}
	if mapStr == "" {
		return nil, nil
	}
	topArr = strings.Split(mapStr, ",")
	return topArr, nil
}

// Update occupied resource info after allocate, for only one chip.
func updateTopStrOfNodeOtherAlloc(nodeTopStrArr []string, top string) string {
	var tmpTopStrArr []string

	for _, nTop := range nodeTopStrArr {
		if nTop == top {
			continue
		}
		tmpTopStrArr = append(tmpTopStrArr, nTop)
	}
	klog.V(util.LogDebugLev).Infof("updateTopStrOfNodeOtherAlloc : %v.", tmpTopStrArr)
	newNodeTopStr := strings.Join(tmpTopStrArr, ",")

	return newNodeTopStr
}

// UpdateNPUNodeTopology Update node info to node.Others
func (tp *VNPU) UpdateNPUNodeTopology(node *api.NodeInfo, top interface{}, updateFn func([]string,
	string) string) error {
	var vType string

	topInstance, ok := top.(string)
	if !ok {
		return errors.New("invalid argument")
	}

	vType = tp.JudgeResourceTypeByTopInfo(topInstance)
	if vType == "" {
		return errors.New("invalid top content")
	}

	// get node available top from node.Others
	nodeTopStrArr, err := getTopStrFromNodeOther(node.Others, vType)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("updateNPUNodeTopology node(%s) top nil.", node.Name)
		return err
	}
	// update to node.Others
	newNodeTopStr := updateFn(nodeTopStrArr, topInstance)
	err = util.ReloadNewTopToNodeOther(node, newNodeTopStr, vType)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("reloadNewTopToNode failed.")
		return err
	}

	klog.V(util.LogInfoLev).Infof("ReloadNewTopToNode %s to %s successes.", newNodeTopStr, node.Name)

	return nil
}

// UpdateNPUNodeUsedCardFn update node others after allocate
func (tp *VNPU) UpdateNPUNodeUsedCardFn(node *api.NodeInfo, top interface{}) error {
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
func (tp *VNPU) UpdateReleaseNPUNodeTopologyFn(node *api.NodeInfo, top interface{}) error {
	klog.V(util.LogDebugLev).Infof("%s UpdateReleaseNPUNodeTopologyFn %s, no need.", tp.Name(), node.Name)
	return nil
}
