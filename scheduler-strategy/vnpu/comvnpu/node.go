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
	"sort"
	"strconv"
	"strings"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/vnpu/vnpuutil"
)

// InitVNodesFn init node.
func (tp *VNPU) InitVNodesFn(nodes map[string]*api.NodeInfo) error {
	for key := range nodes {
		for _, vType := range tp.Attr.DivideKinds {
			nTopStr, err := getResourceFromAnnotationFn(nodes[key].Node.Annotations, vType)
			if err != nil {
				continue
			}
			err = util.SaveTopologyInMap(nodes[key].Others, nTopStr, vType)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func getResourceFromAnnotationFn(Annotations map[string]string, resourceName string) (string, error) {
	topStr, ok := Annotations[resourceName]
	// In case of kubernetes doesn't have some kind of resource type, but name of that type was written in
	// node annotation with a value of empty string. If topStr is empty, an error should be returned so that that type
	// of resource will be ignored.
	if !ok || topStr == "" {
		return "", fmt.Errorf("requested %s does not exist", resourceName)
	}

	return topStr, nil
}

// GetNPUsFromNodeAnnotation get the node annotation.
func (tp *VNPU) GetNPUsFromNodeAnnotation(annotations map[string]string, resourceName string) ([]string, error) {
	topStr, err := getResourceFromAnnotationFn(annotations, resourceName)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("getNPUsFromNodeAnnotation failed to get annotation value")
		return nil, err
	}

	prefix := strings.TrimPrefix(resourceName, tp.Attr.AnnoPreVal)
	tops := strings.Split(topStr, ",")
	sort.Strings(tops)
	for i, top := range tops {
		if !strings.HasPrefix(top, prefix) {
			klog.V(util.LogErrorLev).Infof("getNPUsFromNodeAnnotation: vnpu name(%s) did not match its type(%s)",
				top, prefix)
			return nil, fmt.Errorf("vnpu name(%s) did not match its type(%s)", top, prefix)
		}

		if i > 0 && top == tops[i-1] {
			klog.V(util.LogErrorLev).Infof("getNPUsFromNodeAnnotation: got duplicated npu(%s)", top)
			return nil, fmt.Errorf("got duplicated npu(%s)", top)
		}
	}

	return tops, nil
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

	topArr = strings.Split(mapStr, ",")
	return topArr, nil
}

// Update occupied resource info after allocate
func updateTopStrOfNodeOtherAlloc(nodeTopStrArr []string, top []string) string {
	var tmpTopStrArr []string
	var existFlag bool

	for _, nTop := range nodeTopStrArr {
		existFlag = false
		for _, tTop := range top {
			if nTop == tTop {
				existFlag = true
				break
			}
		}
		if !existFlag {
			tmpTopStrArr = append(tmpTopStrArr, nTop)
		}
	}
	klog.V(util.LogDebugLev).Infof("updateTopStrOfNodeOtherAlloc : %v.", tmpTopStrArr)
	newNodeTopStr := strings.Join(tmpTopStrArr, ",")

	return newNodeTopStr
}

// Update occupied resource info after release
func updateTopStrOfNodeOtherRelease(nodeTopStrArr []string, top []string) string {
	var tmpTopStrArr []string

	tmpTopMap := make(map[string]int, util.ConstIntNum3)
	// add tops that already exist in node.Others to tmp map
	for _, nTop := range nodeTopStrArr {
		tmpTopMap[nTop] = 0
	}
	// add tops that been released to tmp map
	for _, tTop := range top {
		if _, ok := tmpTopMap[tTop]; ok {
			klog.V(util.LogInfoLev).Infof("updateTopStrOfNodeOtherRelease card exists: %s.", tTop)
			continue
		}
		tmpTopMap[tTop] = 0
	}

	for k := range tmpTopMap {
		tmpTopStrArr = append(tmpTopStrArr, k)
	}

	klog.V(util.LogDebugLev).Infof("updateTopStrOfNodeOtherRelease : %v.", tmpTopStrArr)
	newNodeTopStr := strings.Join(tmpTopStrArr, ",")

	return newNodeTopStr
}

// UpdateNPUNodeTopology Update node info to node.Others
func (tp *VNPU) UpdateNPUNodeTopology(node *api.NodeInfo, top interface{}, updateFn func([]string, []string) string) error {
	var vType string

	topArr, ok := top.([]string)
	if !ok {
		return errors.New("invalid argument")
	}

	topInstance := topArr[0]
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
	newNodeTopStr := updateFn(nodeTopStrArr, topArr)
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

// IsVNPUNodeMeetReqResource for VNPU card only need one for one job.
// jobNeedNPUType is like huawei.com/Ascend910-2c
func (tp *VNPU) IsVNPUNodeMeetReqResource(jobNeedNPUType string, tmpNode *api.NodeInfo) bool {
	if tp == nil {
		return false
	}
	nodeCoresInf, coresErr := tp.GetNodeNPUCoreInfoMap(tmpNode)
	if coresErr != nil {
		klog.V(util.LogErrorLev).Infof("%s IsVNPUNodeMeetReqResource %v.", tp.Name(), coresErr)
		return false
	}
	// get need cores
	tmpSlice := strings.Split(jobNeedNPUType, "-")
	if len(tmpSlice) != util.ConstIntNum2 {
		klog.V(util.LogErrorLev).Infof("%s IsVNPUNodeMeetReqResource %s error.", tp.Name(), jobNeedNPUType)
		return false
	}
	chipCoreStr := tmpSlice[1]
	chipCoreStr = strings.TrimRight(chipCoreStr, "c")
	chipCore, covAllErr := strconv.Atoi(chipCoreStr)
	if covAllErr != nil {
		klog.V(util.LogErrorLev).Infof("%s IsVNPUNodeMeetReqResource convert %v.", tp.Name(), covAllErr)
		return false
	}
	for cardID, v := range nodeCoresInf {
		if v.UnCutCore == v.AllCore {
			if !tp.isWholeCardInNodeAnnotation(cardID, tmpNode) {
				// whole card has been used.
				continue
			}
		}

		if v.UnCutCore > chipCore {
			return true
		}
	}
	klog.V(util.LogInfoLev).Infof("%s IsVNPUNodeMeetReqResource %v no %s===%+v.",
		tp.Name(), tmpNode.Name, jobNeedNPUType, nodeCoresInf)
	return false
}

// deal 1-32c-32c
func (tp *VNPU) parseStringToVNPUCoreInfo(coreString string) (string, vNPUCoreInfo, error) {
	// deal 1-32c-32c
	tmpSlice := strings.Split(coreString, "-")
	if len(tmpSlice) != util.ConstIntNum3 {
		coreErr := fmt.Errorf("%s error format", coreString)
		return "", vNPUCoreInfo{}, coreErr
	}
	// get chip id
	chipIDStr := tmpSlice[0]
	chipID, covIDErr := strconv.Atoi(chipIDStr)
	if covIDErr != nil {
		return "", vNPUCoreInfo{}, covIDErr
	}
	// get chip all core.deal 32c
	chipAllCoreStr := tmpSlice[1]
	chipAllCoreStr = strings.TrimRight(chipAllCoreStr, "c")
	chipAllCore, covAllErr := strconv.Atoi(chipAllCoreStr)
	if covAllErr != nil {
		return "", vNPUCoreInfo{}, covAllErr
	}
	// get not cut core
	chipNoCutCoreStr := tmpSlice[util.ConstIntNum2]
	chipNoCutCoreStr = strings.TrimRight(chipNoCutCoreStr, "c")
	chipNoCutCore, covNoCutErr := strconv.Atoi(chipNoCutCoreStr)
	if covNoCutErr != nil {
		return "", vNPUCoreInfo{}, covNoCutErr
	}
	chipCoreInfo := vNPUCoreInfo{
		ChipID:    chipID,
		AllCore:   chipAllCore,
		UnCutCore: chipNoCutCore,
	}
	tmp := strings.TrimLeft(tp.Attr.AnnoName, tp.Attr.AnnoPreVal)
	return tmp + "-" + chipIDStr, chipCoreInfo, nil
}

// GetNodeNPUCoreInfoMap in node annotation like, huawei.com/Ascend910-spec:1-32c,2-30c;
// the key is Ascend910-1;
func (tp *VNPU) GetNodeNPUCoreInfoMap(vNode *api.NodeInfo) (map[string]vNPUCoreInfo, error) {
	if tp == nil {
		return nil, errors.New(vnpuutil.PluginUninitializedError)
	}
	// get the all cores.
	coreString, ok := vNode.Node.Annotations[tp.Attr.NPUCardCoreKey]
	if !ok {
		coreErr := fmt.Errorf("%s has no %s", vNode.Name, tp.Attr.NPUCardCoreKey)
		return nil, coreErr
	}
	coreMap := make(map[string]vNPUCoreInfo, util.ConstIntNum3)
	coreSlice := strings.Split(coreString, ",")
	for _, coreShortInf := range coreSlice {
		card, tmp, parseErr := tp.parseStringToVNPUCoreInfo(coreShortInf)
		if parseErr != nil {
			return nil, parseErr
		}
		coreMap[card] = tmp
	}
	if len(coreMap) == 0 {
		return nil, fmt.Errorf("%s nil core information", vNode.Name)
	}
	return coreMap, nil
}

// cardID like Ascend910-0
func (tp *VNPU) isWholeCardInNodeAnnotation(cardID string, nodeInf *api.NodeInfo) bool {
	cardsStr, ok := nodeInf.Node.Annotations[tp.Attr.AnnoName]
	if !ok {
		klog.V(util.LogInfoLev).Infof("%s isWholeCardInNodeAnnotation %v no %s.",
			tp.Name(), nodeInf.Name, tp.Attr.AnnoName)
		return false
	}
	cardsSlice := strings.Split(cardsStr, ",")
	for _, cardStr := range cardsSlice {
		if cardStr == cardID {
			return true
		}
	}
	return false
}

// get the cut cores
func (tp *VNPU) getNodeUsedNPUCores(nodeInf *api.NodeInfo) int {
	nodeCoresInf, coresErr := tp.GetNodeNPUCoreInfoMap(nodeInf)
	if coresErr != nil {
		return 0
	}
	var allUsedCores = 0
	// has cut
	for _, v := range nodeCoresInf {
		tmp := v.AllCore - v.UnCutCore
		allUsedCores += tmp
	}
	// whole card.
	for cardID, v := range nodeCoresInf {
		if v.AllCore == v.UnCutCore {
			if !tp.isWholeCardInNodeAnnotation(cardID, nodeInf) {
				continue
			}
			allUsedCores += v.AllCore
		}
	}
	return allUsedCores
}

func (tp *VNPU) getNodeHealThyNPUNum(nodeInf *api.NodeInfo) int {
	healthNum := 0
	unhealthyString, ok := nodeInf.Node.Annotations[tp.Attr.AnnoUnhealthyName]
	if ok {
		tmpUnhealthy := strings.Split(unhealthyString, ",")
		healthNum = len(tmpUnhealthy)
	}
	coreString, ok := nodeInf.Node.Annotations[tp.Attr.NPUCardCoreKey]
	if !ok {
		return 0
	}
	coreSlice := strings.Split(coreString, ",")
	allNPUNum := len(coreSlice)
	return allNPUNum - healthNum
}

func (tp *VNPU) intSortNodeListByPriorityAndCores(nodeList []*api.NodeInfo) ([]priorNodes, error) {
	var tmpList []priorNodes
	for _, nodeInf := range nodeList {
		tmp := priorNodes{
			NodeName:    nodeInf.Name,
			HealThyCard: tp.getNodeHealThyNPUNum(nodeInf),
			UsedCores:   tp.getNodeUsedNPUCores(nodeInf),
		}
		tmpList = append(tmpList, tmp)
	}
	if len(tmpList) == 0 {
		return nil, errors.New("none node init into struct")
	}
	return tmpList, nil
}

// GetVJobMeetNode sort by follows:
// 1. Prioritize by chip failure.
// 2. Total number of cores used.
func (tp *VNPU) GetVJobMeetNode(nodeList []*api.NodeInfo) (*api.NodeInfo, error) {
	if tp == nil {
		return nil, errors.New("nil plugin")
	}
	// 1.init sort struct
	tmpList, initErr := tp.intSortNodeListByPriorityAndCores(nodeList)
	if initErr != nil {
		return nil, fmt.Errorf("init PriorNodes %v", initErr)
	}
	// 2.sort node list
	sort.Slice(tmpList, func(i, j int) bool {
		if tmpList[i].HealThyCard < tmpList[j].HealThyCard {
			return true
		}
		if tmpList[i].HealThyCard == tmpList[j].HealThyCard {
			return tmpList[i].UsedCores < tmpList[j].UsedCores
		}
		return false
	})
	// 3.get node information.
	nodeName := tmpList[0].NodeName
	for _, node := range nodeList {
		if node.Name == nodeName {
			return node, nil
		}
	}
	return nil, errors.New("none node find")
}

func (tp *VNPU) getMeetChipsInNode(needCores int, nodeInf *api.NodeInfo) (map[int]int, error) {
	var chipCores = make(map[int]int, util.ConstIntNum8)
	// Get from whole card.
	idsMap, numErr := util.GetNodeAvailNPUIdsFromAnno(nodeInf, tp.Attr.AnnoName)
	if numErr != numErr {
		return nil, numErr
	}
	// Get split cards.
	nodeCoresInf, coresErr := tp.GetNodeNPUCoreInfoMap(nodeInf)
	if coresErr != nil {
		return nil, coresErr
	}
	for _, v := range nodeCoresInf {
		// whole card
		if _, ok := idsMap[v.ChipID]; ok {
			chipCores[v.ChipID] = v.UnCutCore
			continue
		}
		// split card
		if v.UnCutCore > needCores {
			chipCores[v.ChipID] = v.UnCutCore
		}
	}
	if len(chipCores) == 0 {
		return nil, fmt.Errorf("%s get none fit %v", nodeInf.Name, needCores)
	}
	return chipCores, nil
}

// GetVNPUUsedChipByReq First occupancy principle.
// needNPU like huawei.com/Ascend910-16c
func (tp *VNPU) GetVNPUUsedChipByReq(needNPU string, nodeInf *api.NodeInfo) (string, error) {
	core, err := vnpuutil.ChangeReqVNPUToCores(needNPU)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s GetVNPUUsedChipByReq %s %v.", tp.Name(), nodeInf.Name, err)
		return "", err
	}
	chipList, listErr := tp.getMeetChipsInNode(core, nodeInf)
	if listErr != nil {
		klog.V(util.LogErrorLev).Infof("%s GetVNPUUsedChipByReq %v.", tp.Name(), listErr)
		return "", listErr
	}
	return tp.getTheFillOneFromList(chipList)
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
			return fmt.Errorf("getNodeNPUNumFromIdle %s : %s", vnpuutil.NodesNoMeetNPUReqError, err)
		}

		if err = util.CheckNodeNPUStabilize(len(nodeNPUIdleNumFromTop), nodeNPUIdleNumFromIdle); err != nil {
			return fmt.Errorf("node %s %s : %s", node.Name, vnpuutil.NodeNotStableWarning, err)
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
