/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package comvnpu is using for virtual HuaWei vnpu schedule.

*/
package comvnpu

import (
	"errors"
	"fmt"
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

// Update occupied resource info after release
func updateTopStrOfNodeOtherRelease(nodeTopStrArr []string, top string) string {
	var tmpTopStrArr []string

	tmpTopMap := make(map[string]struct{}, util.NPUIndex3)
	// add tops that already exist in node.Others to tmp map
	for _, nTop := range nodeTopStrArr {
		tmpTopMap[nTop] = struct{}{}
	}
	// add tops that been released to tmp map
	tmpTopMap[top] = struct{}{}

	for k := range tmpTopMap {
		tmpTopStrArr = append(tmpTopStrArr, k)
	}

	klog.V(util.LogDebugLev).Infof("updateTopStrOfNodeOtherRelease : %v.", tmpTopStrArr)
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
	if ok := tp.UpdateNPUNodeTopology(node, top, updateTopStrOfNodeOtherAlloc); ok != nil {
		return errors.New("update npu node topology failed")
	}

	return nil
}

// IsNodeHasVNPUSelector judge the node vnpu label.
func (tp *VNPU) IsNodeHasVNPUSelector(vNode *api.NodeInfo) error {
	nodeSelectors, nodeErr := util.GetNodeSelector(vNode)
	if nodeErr != nil {
		klog.V(util.LogDebugLev).Infof("node(%s) %v.", vNode.Name, nodeErr)
		return nodeErr
	}

	acceleratorValue, ok := nodeSelectors[vnpuutil.VNPUNodeLabelKey]
	if !ok {
		selectErr := fmt.Errorf("%s has no %s", vNode.Name, vnpuutil.VNPUNodeLabelKey)
		klog.V(util.LogDebugLev).Infof("%s IsNodeHasVNPUSelector %v.", tp.Name(), selectErr)
		return selectErr
	}

	if acceleratorValue != vnpuutil.VNPUNodeLabelValue {
		valueErr := fmt.Errorf("%s has %s::%s", vNode.Name, acceleratorValue, vnpuutil.VNPUNodeLabelKey)
		klog.V(util.LogDebugLev).Infof("%s IsNodeHasVNPUSelector %v.", tp.Name(), valueErr)
		return valueErr
	}

	klog.V(util.LogInfoLev).Infof("%s is vnpu node.", vNode.Name)
	return nil
}

// IsMyNode used for identify Vnpu node, need to be implemented by vNPU plugins
func (tp *VNPU) IsMyNode(vNode *api.NodeInfo) error {
	// 1、has npu card
	if nodeErr := util.IsNPUNNode(vNode); nodeErr != nil {
		return nodeErr
	}
	// 2、has npu selector
	if selectorErr := tp.IsNodeHasVNPUSelector(vNode); selectorErr != nil {
		return selectorErr
	}
	return nil
}

// CheckNPUResourceStableFn check whether the resources on the node are stable
func (tp *VNPU) CheckNPUResourceStableFn(node *api.NodeInfo) error {
	for _, vType := range tp.Attr.DivideKinds {
		nodeNPUIdleNumFromTop, getErr := getTopStrFromNodeOther(node.Others, vType)
		if getErr != nil {
			klog.V(util.LogDebugLev).Infof("getNodeNPUNumFromOthers %s %v.", node.Name, getErr)
			continue
		}

		nodeNPUIdleNumFromIdle, err := util.GetNodeNPUNumFromIdle(node, vType)
		if err != nil {
			idleErr := fmt.Errorf("getNodeNPUNumFromIdle %s : %s", vnpuutil.NodesNoMeetNPUReqError, err)
			klog.V(util.LogErrorLev).Infof("%s CheckNPUResourceStableFn %v.", tp.Name(), idleErr)
			return idleErr
		}

		if err = util.CheckNodeNPUStabilize(len(nodeNPUIdleNumFromTop), nodeNPUIdleNumFromIdle); err != nil {
			checkErr := fmt.Errorf("%s %s %s : %v", node.Name, vType, vnpuutil.NodeNotStableWarning, err)
			klog.V(util.LogErrorLev).Infof("%s CheckNPUResourceStableFn %v.", tp.Name(), checkErr)
			return checkErr
		}
	}

	return nil
}

// UpdateReleaseNPUNodeTopologyFn update node others after release
func (tp *VNPU) UpdateReleaseNPUNodeTopologyFn(node *api.NodeInfo, top interface{}) error {
	if ok := tp.UpdateNPUNodeTopology(node, top, updateTopStrOfNodeOtherRelease); ok != nil {
		return errors.New("update npu node topology after release failed")
	}

	return nil
}
