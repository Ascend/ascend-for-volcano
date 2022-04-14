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

	"github.com/hashicorp/go-multierror"
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/vnpu/vnpuutil"
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

// GetNPUsFromNodeAnnotation get the node annotation.
func (tp *VNPU) GetNPUsFromNodeAnnotation(annotations map[string]string, resourceName string) ([]string, error) {
	topStr, err := util.GetResourceFromAnnotationFn(annotations, resourceName)
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

	tmpTopMap := make(map[string]struct{}, util.ConstIntNum3)
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

func (tp *VNPU) coverReqNPUTypeToCoreNum(jobNeedNPUType string) (int, error) {
	tmpSlice := strings.Split(jobNeedNPUType, "-")
	if len(tmpSlice) != util.ConstIntNum2 {
		klog.V(util.LogErrorLev).Infof("%s IsVNPUNodeMeetReqResource %s error.", tp.Name(), jobNeedNPUType)
		return 0, fmt.Errorf("error format %v", jobNeedNPUType)
	}
	chipCoreStr := tmpSlice[1]
	chipCoreStr = strings.TrimRight(chipCoreStr, "c")
	chipCore, covAllErr := strconv.Atoi(chipCoreStr)
	if covAllErr != nil {
		klog.V(util.LogErrorLev).Infof("%s IsVNPUNodeMeetReqResource convert %v.", tp.Name(), covAllErr)
		return 0, covAllErr
	}
	return chipCore, nil
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
	chipCore, coverErr := tp.coverReqNPUTypeToCoreNum(jobNeedNPUType)
	if coverErr != nil {
		klog.V(util.LogErrorLev).Infof("%s IsVNPUNodeMeetReqResource %v.", tp.Name(), coresErr)
		return false
	}
	for cardID, v := range nodeCoresInf {
		if v.UnCutCore == v.AllCore {
			if !tp.isWholeCardInNodeAnnotation(cardID, tmpNode) {
				// whole card has been used.
				continue
			}
		}

		if v.UnCutCore >= chipCore {
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
	coreString, getErr := util.GetNPUAllocCardsFromNodeOthers(vNode, tp.Attr.NPUCardCoreKey)
	if getErr != nil {
		klog.V(util.LogDebugLev).Infof("GetNodeNPUCoreInfoMap :%v", getErr)
		return nil, getErr
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

func (tp *VNPU) updateNodeOtherCardCoresInf(nodeInf *api.NodeInfo, nodeCoresInf map[string]vNPUCoreInfo) error {
	var allCards []string
	for _, tmp := range nodeCoresInf {
		idStr := strconv.Itoa(tmp.ChipID)
		allStr := strconv.Itoa(tmp.AllCore)
		unCutStr := strconv.Itoa(tmp.UnCutCore)
		chipStr := idStr + "-" + allStr + "c-" + unCutStr + "c"
		allCards = append(allCards, chipStr)
	}
	writeString := strings.Join(allCards, ",")
	if saveErr := util.SaveTopologyInMap(nodeInf.Others, writeString, tp.Attr.NPUCardCoreKey); saveErr != nil {
		return saveErr
	}
	return nil
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
	if numErr != nil {
		klog.V(util.LogErrorLev).Infof("%s getMeetChipsInNode %s %v.", tp.Name(), nodeInf.Name, numErr)
		return nil, numErr
	}
	// Get split cards.
	nodeCoresInf, coresErr := tp.GetNodeNPUCoreInfoMap(nodeInf)
	if coresErr != nil {
		klog.V(util.LogErrorLev).Infof("%s getMeetChipsInNode %s %v.", tp.Name(), nodeInf.Name, coresErr)
		return nil, coresErr
	}
	klog.V(util.LogDebugLev).Infof("%s getMeetChipsInNode %s other:%+v idsMap:%+v %+v", tp.Name(), nodeInf.Name,
		nodeInf.Others, idsMap, nodeCoresInf)
	for _, v := range nodeCoresInf {
		// whole card
		if _, ok := idsMap[v.ChipID]; ok {
			chipCores[v.ChipID] = v.UnCutCore
			continue
		}
		// split card exclude no cut chip.
		if v.UnCutCore >= needCores && v.UnCutCore != v.AllCore {
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
	klog.V(util.LogDebugLev).Infof("%s GetVNPUUsedChipByReq %s %d chipList:%+v", tp.Name(), nodeInf.Name, core,
		chipList)
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

func (tp *VNPU) reduceTheAllocWholeChipToNodeOther(chip string, nodeInf *api.NodeInfo) error {
	topStr, getErr := util.GetNPUAllocCardsFromNodeOthers(nodeInf, tp.Attr.AnnoName)
	if getErr != nil {
		klog.V(util.LogErrorLev).Infof("%s reduceTheAllocChipFromNodeOther :+%v.", tp.Name(), getErr)
		return getErr
	}
	var allCards []string
	topSlice := strings.Split(topStr, ",")
	for _, chipStr := range topSlice {
		if chipStr == chip {
			continue
		}
		allCards = append(allCards, chipStr)
	}

	if len(allCards) != len(topSlice) {
		// whole card deal over.need to deal cores continue.
		nodeInf.Others[tp.Attr.AnnoName] = strings.Join(allCards, ",")
	}
	return nil
}

// GetVJobReqNPUCoreNum get the job require npu core number.
func (tp *VNPU) GetVJobReqNPUCoreNum(vJob *api.JobInfo) (int, error) {
	tmp, getErr := util.GetReqResourceNameFromJob(vJob)
	if getErr != nil {
		klog.V(util.LogErrorLev).Infof("%s GetVJobReqNPUType %s %v.", tp.Name(), vJob.Name, getErr)
		return 0, getErr
	}
	// tmp like huawei.com/Ascend910-2c
	tmpSlice := strings.Split(tmp, "-")
	if len(tmpSlice) != util.ConstIntNum2 {
		return 0, fmt.Errorf("%s err require %s", vJob.Name, tmp)
	}
	tmpStr := strings.TrimRight(tmpSlice[1], "c")
	coreNum, conErr := strconv.Atoi(tmpStr)
	if conErr != nil {
		return 0, conErr
	}
	return coreNum, nil
}

// reduceTheAllocChipFromNodeOther chip is Ascend710-0
func (tp *VNPU) reduceTheAllocChipFromNodeOther(chip string, vJob *api.JobInfo, nodeInf *api.NodeInfo) error {
	chipSlice := strings.Split(chip, "-")
	if len(chipSlice) != util.ConstIntNum2 {
		return fmt.Errorf("%s is wrong chip", chip)
	}
	// deal the whole card.
	if wholeErr := tp.reduceTheAllocWholeChipToNodeOther(chip, nodeInf); wholeErr != nil {
		return wholeErr
	}
	// deal the core.
	jobReqCoreNum, jobCoreErr := tp.GetVJobReqNPUCoreNum(vJob)
	if jobCoreErr != nil {
		return jobCoreErr
	}
	coreString, getErr := util.GetNPUAllocCardsFromNodeOthers(nodeInf, tp.Attr.NPUCardCoreKey)
	if getErr != nil {
		klog.V(util.LogDebugLev).Infof("%s reduceTheAllocChipFromNodeOther :%v", tp.Name(), getErr)
		return getErr
	}
	var allCors []string
	coreSlice := strings.Split(coreString, ",")
	for _, coreStr := range coreSlice {
		// coreStr is 4-8c-6c
		tmpSlice := strings.Split(coreStr, "-")
		if len(tmpSlice) != util.ConstIntNum3 {
			return fmt.Errorf("wrong core string %s", coreStr)
		}
		if chipSlice[1] != tmpSlice[0] {
			allCors = append(allCors, coreStr)
			continue
		}
		tempChar := tmpSlice[util.ConstIntNum2]
		desChar := strings.TrimRight(tempChar, "c")
		tmpNum, tmpErr := strconv.Atoi(desChar)
		if tmpErr != nil {
			return tmpErr
		}
		resNum := tmpNum - jobReqCoreNum
		if resNum < 0 {
			return fmt.Errorf("wrong core node(%s),job(%v)", coreStr, jobReqCoreNum)
		}
		resChar := strconv.Itoa(resNum)
		resStr := tmpSlice[0] + "-" + tmpSlice[1] + "-" + resChar + "c"
		allCors = append(allCors, resStr)
	}
	nodeInf.Others[tp.Attr.NPUCardCoreKey] = strings.Join(allCors, ",")
	return nil
}

// getNodeUseInfoFromNode get from node cores(NPUCardCoreKey).
func (tp *VNPU) getNodeUseInfoFromNode(nodeInf *api.NodeInfo) (map[string]int, error) {
	tmp := make(map[string]int, util.ConstIntNum3)
	nodeCoresInf, coresErr := tp.GetNodeNPUCoreInfoMap(nodeInf)
	if coresErr != nil {
		klog.V(util.LogDebugLev).Infof("%s getNodeUseInfoFromNode :%v", tp.Name(), coresErr)
		return nil, coresErr
	}
	for cardName, value := range nodeCoresInf {
		tmp[cardName] = value.AllCore - value.UnCutCore
	}
	if len(tmp) == 0 {
		return nil, fmt.Errorf("%s's other no need change", nodeInf.Name)
	}
	return tmp, nil
}

// getNodeUseInfoFromVNPUCache the format key like Ascend710-0.
func (tp *VNPU) getNodeUseInfoFromVNPUCache(nodeInf *api.NodeInfo) (map[string]int, error) {
	tmp := make(map[string]int, util.ConstIntNum3)
	for _, value := range vnpuutil.VNPUAllocData.Cache {
		if value.NodeName != nodeInf.Name {
			continue
		}
		if !value.AllocFlag {
			continue
		}
		chipCore, coverErr := tp.coverReqNPUTypeToCoreNum(value.ReqNPUType)
		if coverErr != nil {
			klog.V(util.LogErrorLev).Infof("%s getNodeUseInfoFromVNPUCache %v.", tp.Name(), coverErr)
			continue
		}
		tmp[value.ReqCardName] += chipCore
	}
	if len(tmp) == 0 {
		return nil, fmt.Errorf("%s's other no need change", nodeInf.Name)
	}
	return tmp, nil
}

// updateNodeOtherWholeCardByUseMap for node and useMap is corresponding, must use getNodeUseInfoFromVNPUCache before.
func (tp *VNPU) updateNodeOtherWholeCardByUseMap(nodeInf *api.NodeInfo, useMap map[string]int) error {
	for cardName := range useMap {
		// cardName is Ascend710-0
		tmpSlice := strings.Split(cardName, "-")
		if len(tmpSlice) < util.ConstIntNum2 {
			return fmt.Errorf("%s err card name %s", nodeInf.Name, cardName)
		}
		resName := vnpuutil.NPUCardNamePrefix + tmpSlice[0]
		topStr, getErr := util.GetNPUAllocCardsFromNodeOthers(nodeInf, resName)
		if getErr != nil {
			klog.V(util.LogDebugLev).Infof("%s updateNodeOtherWholeCardByUseMap:%v.", tp.Name(), getErr)
			continue
		}
		if !strings.Contains(topStr, cardName) {
			continue
		}
		var allCards []string
		topSlice := strings.Split(topStr, ",")
		for _, chipStr := range topSlice {
			if chipStr == cardName {
				continue
			}
			allCards = append(allCards, chipStr)
		}
		writeString := strings.Join(allCards, ",")
		if saveErr := util.SaveTopologyInMap(nodeInf.Others, writeString, resName); saveErr != nil {
			return saveErr
		}
	}
	return nil
}

// updateNodeOtherCardCoresByUseMap for node and useMap is corresponding, must use getNodeUseInfoFromVNPUCache before.
func (tp *VNPU) updateNodeOtherCardCoresByUseMap(nodeInf *api.NodeInfo, useMap map[string]int) error {
	if tp == nil {
		return fmt.Errorf("%s nil parameter", nodeInf.Name)
	}
	for cardName, useCores := range useMap {
		// cardName is Ascend710-0
		nodeCoresInf, coresErr := tp.GetNodeNPUCoreInfoMap(nodeInf)
		if coresErr != nil {
			klog.V(util.LogErrorLev).Infof("%s IsVNPUNodeMeetReqResource %v.", tp.Name(), coresErr)
			continue
		}
		coreInfo, ok := nodeCoresInf[cardName]
		if !ok {
			klog.V(util.LogErrorLev).Infof("%s updateNodeOtherCardCoresByUseMap %s no %v.", tp.Name(),
				nodeInf.Name, cardName)
			continue
		}
		if useCores > coreInfo.UnCutCore {
			klog.V(util.LogErrorLev).Infof("%s updateNodeOtherCardCoresByUseMap %s %s %v over %v.", tp.Name(),
				nodeInf.Name, cardName, useCores, coreInfo.UnCutCore)
			continue
		}
		coreInfo.UnCutCore -= useCores
		nodeCoresInf[cardName] = coreInfo
		if upErr := tp.updateNodeOtherCardCoresInf(nodeInf, nodeCoresInf); upErr != nil {
			klog.V(util.LogErrorLev).Infof("%s updateNodeOtherCardCoresByUseMap %v.", tp.Name(), upErr)
			continue
		}
	}
	return nil
}

// updateNPUInfInNodeOtherByUseMap for node and useMap is corresponding, must use getNodeUseInfoFromVNPUCache before.
func (tp *VNPU) updateNPUInfInNodeOtherByUseMap(nodeInf *api.NodeInfo, useMap map[string]int) error {
	klog.V(util.LogDebugLev).Infof("%s updateNPUInfInNodeOtherByUseMap before:%v", tp.Name(), nodeInf.Others)
	var returnErr error
	// deal whole card
	if upErr := tp.updateNodeOtherWholeCardByUseMap(nodeInf, useMap); upErr != nil {
		klog.V(util.LogErrorLev).Infof("%s updateNPUInfInNodeOtherByUseMap :%v", tp.Name(), upErr)
	}
	// deal chip cores
	if upErr := tp.updateNodeOtherCardCoresByUseMap(nodeInf, useMap); upErr != nil {
		klog.V(util.LogErrorLev).Infof("%s updateNPUInfInNodeOtherByUseMap :%v", tp.Name(), upErr)
		returnErr = multierror.Append(returnErr, upErr)
	}
	klog.V(util.LogDebugLev).Infof("%s updateNPUInfInNodeOtherByUseMap after:%v", tp.Name(), nodeInf.Others)
	return returnErr
}

// InitVNPUPluginByNodeInfo Init VNPU plugin by nodeInfo.
func (tp *VNPU) InitVNPUPluginByNodeInfo(nodeInf *api.NodeInfo) error {
	pluginName, nameErr := tp.GetPluginNameByNodeInfo(nodeInf)
	if nameErr != nil {
		klog.V(util.LogErrorLev).Infof("InitVNPUPluginByNodeInfo :%v.", nameErr)
		return nameErr
	}
	if pluginErr := tp.InitVNPUPluginByType(pluginName); pluginErr != nil {
		klog.V(util.LogErrorLev).Infof("InitVNPUPluginByNodeInfo :%v.", pluginErr)
		return pluginErr
	}
	klog.V(util.LogDebugLev).Infof("InitVNPUPluginByNodeInfo init %v.", tp.Name())
	return nil
}

func (tp *VNPU) getNodePreUseInfo(nodeInf *api.NodeInfo) (map[string]int, error) {
	preUseMap := make(map[string]int, util.ConstIntNum3)
	cacheUseMap, getErr := tp.getNodeUseInfoFromVNPUCache(nodeInf)
	if getErr != nil {
		klog.V(util.LogErrorLev).Infof("%s getNodePreUseInfo :%v", tp.Name(), getErr)
		return nil, getErr
	}
	nodeUseMap, getErr := tp.getNodeUseInfoFromNode(nodeInf)
	if getErr != nil {
		klog.V(util.LogErrorLev).Infof("%s getNodePreUseInfo :%v", tp.Name(), getErr)
		return cacheUseMap, nil
	}
	for cardName, value := range cacheUseMap {
		useValue, ok := nodeUseMap[cardName]
		if !ok {
			preUseMap[cardName] = value
			continue
		}
		re := value - useValue
		if re < 0 {
			valueErr := fmt.Errorf("%s cache pre-use %v less than core cut %d", nodeInf.Name, value, useValue)
			klog.V(util.LogErrorLev).Infof("%s getNodePreUseInfo :%v", tp.Name(), valueErr)
			return nil, valueErr
		}
		preUseMap[cardName] = re
	}
	return preUseMap, nil
}

// updateNodesOthersByVNPUCache must do after cache update.
func (tp *VNPU) updateNodesOthersByVNPUCache(ssnNodes map[string]*api.NodeInfo) error {
	var returnErr error
	if len(ssnNodes) == 0 {
		klog.V(util.LogErrorLev).Infof("%s updateNodesOthersByVNPUCache nil nodes.", tp.Name())
		return errors.New("nil nodes")
	}
	for _, nodeInf := range ssnNodes {
		if pluginErr := tp.InitVNPUPluginByNodeInfo(nodeInf); pluginErr != nil {
			klog.V(util.LogErrorLev).Infof("%s updateNodesOthersByVNPUCache :%v.", tp.Name(), pluginErr)
			continue
		}
		useMap, getErr := tp.getNodePreUseInfo(nodeInf)
		if getErr != nil {
			klog.V(util.LogDebugLev).Infof("%s updateNodesOthersByVNPUCache :%v", tp.Name(), getErr)
			continue
		}
		klog.V(util.LogErrorLev).Infof("%s updateNodesOthersByVNPUCache %v:%+v", tp.Name(), nodeInf.Name, useMap)
		if upErr := tp.updateNPUInfInNodeOtherByUseMap(nodeInf, useMap); upErr != nil {
			klog.V(util.LogErrorLev).Infof("%s updateNodesOthersByVNPUCache :%v", tp.Name(), upErr)
			returnErr = multierror.Append(returnErr, upErr)
			continue
		}
	}
	return returnErr
}

// GetVNodeNPUType get node resource npu type.
func (tp *VNPU) GetVNodeNPUType(nodeInf *api.NodeInfo) (string, error) {
	tmp, getErr := util.GetReqResourceNameFromNode(nodeInf)
	if getErr != nil {
		klog.V(util.LogErrorLev).Infof("%s GetVNodeNPUType %s %v.", tp.Name(), nodeInf.Name, getErr)
		return "", getErr
	}

	return tp.GetNPUTypeByResourceName(tmp)
}

// GetPluginNameByNodeInfo Get plugin name by nodeInfo
func (tp *VNPU) GetPluginNameByNodeInfo(nodeInf *api.NodeInfo) (string, error) {
	reqNpuType, typeErr := tp.GetVNodeNPUType(nodeInf)
	if typeErr != nil {
		klog.V(util.LogErrorLev).Infof("%s GetPluginNameByNodeInfo %s %v.", tp.Name(), nodeInf.Name, typeErr)
		return "", typeErr
	}

	return tp.getVNPUPluginNameByReqType(reqNpuType), nil
}
