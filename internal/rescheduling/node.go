/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package rescheduling is using for HuaWei Ascend pin fault rescheduling.

*/
package rescheduling

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"k8s.io/klog"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// createFaultCardHandlers initialise FaultCard struct == getInoperableNPUCards
func (fNode *FaultNode) createFaultCardHandlers(node *plugin.NPUNode) ([]FaultCard, error) {
	klog.V(util.LogInfoLev).Infof("create new fault card handlers for node %s", node.Name)
	var faultCards []FaultCard
	for _, card := range fNode.AllCards {
		faultCard := FaultCard{
			IsFaultCard: false,
			NPUName:     card,
			NodeName:    node.Name,
			FaultType:   CardHealthy,
		}

		if faultCard.isCardUnhealthy(fNode.UnhealthyNPU) {
			klog.V(util.LogDebugLev).Infof("card %s is unhealthy", faultCard.NPUName)
			faultCard.setIsFaultCard(true)
			faultCard.setFaultType(CardUnhealthy)
			faultCards = append(faultCards, faultCard)
			continue
		}
		if faultCard.isCardNetworkUnhealthy(fNode.NetworkUnhealthyNPU) {
			klog.V(util.LogDebugLev).Infof("card %s is network unhealthy", faultCard.NPUName)
			faultCard.setIsFaultCard(true)
			faultCard.setFaultType(CardNetworkUnhealthy)
			faultCards = append(faultCards, faultCard)
			continue
		}
		faultCards = append(faultCards, faultCard)
	}

	return faultCards, nil
}

// getNodeNPUsByKey get the npu list from node.DeviceInfo
func (fNode *FaultNode) getNodeNPUsByKey(node *plugin.NPUNode, deviceKey string) ([]string, error) {
	NPUStr, ok := node.Annotation[deviceKey]
	if !ok || len(NPUStr) == 0 {
		return nil, fmt.Errorf("%s get nil npus", node.Name)
	}
	NPUs := strings.Split(NPUStr, ",")

	return NPUs, nil
}

func (fNode *FaultNode) getNodeHeartbeatByKey(node *plugin.NPUNode, hbKey string) (string, error) {
	IntervalStr, ok := node.Annotation[hbKey]
	if !ok || len(IntervalStr) == 0 {
		klog.V(util.LogErrorLev).Infof("isNodeHealth %s no [%s].", node.Name, nodeHeartbeat)
		return "", fmt.Errorf("getFaultNodeState %s nil", node.Name)
	}
	return IntervalStr, nil
}

// getAllNPUCardsFromDeviceInfo get un-allocated healthy card from device info
func (fNode *FaultNode) getAllNPUCardsFromDeviceInfo(node *plugin.NPUNode, cardName string) ([]string, error) {
	var allCard []string
	healthyCard, err := fNode.getNodeNPUsByKey(node, cardName) // ["Ascend910-0", ...]
	allCard = append(allCard, healthyCard...)
	allCard = append(allCard, fNode.UnhealthyNPU...)
	allCard = append(allCard, fNode.NetworkUnhealthyNPU...)
	if err != nil {
		return allCard, err
	}
	return allCard, nil
}

// getUnhealthyCardsFromDeviceInfo get unhealthyCard from device info
func (fNode *FaultNode) getUnhealthyCardsFromDeviceInfo(node *plugin.NPUNode, cardName string) ([]string, error) {
	unhealthyCardName := fmt.Sprintf("%s-%s", cardName, CardUnhealthy) // ["Ascend910-1"]
	return fNode.getNodeNPUsByKey(node, unhealthyCardName)
}

// getNetworkUnhealthyCardsFromDeviceInfo get networkUnhealthyCard from device info
func (fNode *FaultNode) getNetworkUnhealthyCardsFromDeviceInfo(
	node *plugin.NPUNode, cardName string) ([]string, error) {
	networkUnhealthyCardName := fmt.Sprintf("%s-%s", cardName, CardNetworkUnhealthy) // ["Ascend910-1"]
	return fNode.getNodeNPUsByKey(node, networkUnhealthyCardName)
}

// getNodeHeartbeatIntervalFromDeviceInfo get nodeHeartbeatInterval from device info
func (fNode *FaultNode) getNodeHeartbeatIntervalFromDeviceInfo(node *plugin.NPUNode) (int, error) {
	var heartbeatInterval = nodeUpdateTime
	heartbeatIntervalStr, getErr := fNode.getNodeHeartbeatByKey(node, nodeHeartbeatInterval)
	if getErr != nil {
		return heartbeatInterval, getErr
	}
	var err error
	heartbeatInterval, err = strconv.Atoi(heartbeatIntervalStr)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s cover %s to int64 failed [%#v].",
			node.Name, heartbeatIntervalStr, err)
		return nodeUpdateTime, err
	}

	if heartbeatInterval > maxIntervalTime || heartbeatInterval < 1 {
		klog.V(util.LogErrorLev).Infof("%s's HeartbeatInterval %d over limit, will use %d.",
			node.Name, heartbeatInterval, nodeUpdateTime)
		return nodeUpdateTime, nil
	}
	klog.V(util.LogDebugLev).Infof("%s heartbeatTimeInterval: %d", node.Name, heartbeatInterval)
	return heartbeatInterval, nil
}

// getNodeHeartbeatFromDeviceInfo get nodeHeartbeat from device info
func (fNode *FaultNode) getNodeHeartbeatFromDeviceInfo(node *plugin.NPUNode) (int64, error) {
	heartbeatTimeStr, getErr := fNode.getNodeHeartbeatByKey(node, nodeHeartbeat)
	if getErr != nil {
		return 0, getErr
	}
	heartbeatTime, err := strconv.ParseInt(heartbeatTimeStr, util.Base10, util.BitSize64)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s cover %s to int64 failed [%#v].", node.Name, heartbeatTimeStr, err)
		return 0, err
	}
	klog.V(util.LogDebugLev).Infof("%s heartbeatTime: %d", node.Name, heartbeatTime)
	return heartbeatTime, nil
}

func (fCard *FaultCard) isCardUnhealthy(unHealthyList []string) bool {
	return util.IsSliceContain(fCard.NPUName, unHealthyList)
}

func (fCard *FaultCard) isCardNetworkUnhealthy(networkUnhealthyList []string) bool {
	return util.IsSliceContain(fCard.NPUName, networkUnhealthyList)
}

func (fNode *FaultNode) updateFaultNodesFromDeviceInfo(node *plugin.NPUNode, cardName string) {
	klog.V(util.LogInfoLev).Infof("update information from device info for node %s", node.Name)

	tmpHBTime, err := fNode.getNodeHeartbeatFromDeviceInfo(node)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("getNodeHeartbeatFromDeviceInfo: %#v", err)
	}

	klog.V(util.LogDebugLev).Infof(
		"getNodeHeartbeatFromDeviceInfo: former heartbeat time %d, new heartbeat time %d",
		fNode.OldHeartbeatTime, tmpHBTime)
	if fNode.OldHeartbeatTime != tmpHBTime {
		fNode.UpdateHeartbeatTime = time.Now().Unix()
	}
	fNode.setNewNodeHeartbeatTime(tmpHBTime)

	tmpHBIntervalTime, err := fNode.getNodeHeartbeatIntervalFromDeviceInfo(node)
	if err != nil {
		klog.V(util.LogDebugLev).Infof("getNodeHeartbeatIntervalFromDeviceInfo: %#v", err)
	}
	fNode.setNodeHeartbeatInterval(tmpHBIntervalTime)

	tmpUnhealthyNPUs, err := fNode.getUnhealthyCardsFromDeviceInfo(node, cardName)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("getUnhealthyCardsFromDeviceInfo: %#v", err)
	}
	fNode.setUnhealthyNPUList(tmpUnhealthyNPUs)

	tmpNetworkUnhealthyNPUs, err := fNode.getNetworkUnhealthyCardsFromDeviceInfo(node, cardName)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("getNetworkUnhealthyCardsFromDeviceInfo: %#v", err)
	}
	fNode.setNetworkUnhealthyNPUList(tmpNetworkUnhealthyNPUs)

	tmpAllCardsList, err := fNode.getAllNPUCardsFromDeviceInfo(node, cardName)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("getAllNPUCardsFromDeviceInfo: %#v", err)
	}
	fNode.setAllCardList(tmpAllCardsList)
}

// updateFaultNodesAttr update Information from device Info
func (fNode *FaultNode) updateFaultNodesAttr(node *plugin.NPUNode) error {
	klog.V(util.LogInfoLev).Infof("Update node %s attributes", node.Name)
	// 1. create fault Card Object
	tmpFaultCards, err := fNode.createFaultCardHandlers(node)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("Getting node card failed: %#v", err)
	}
	fNode.setFaultCards(tmpFaultCards)

	// 2. judge if node is unhealthy because of card unhealthy
	for _, card := range tmpFaultCards {
		if !card.IsFaultCard {
			continue
		}
		klog.V(util.LogInfoLev).Infof("node %s is fault node for having fault cards", node.Name)
		fNode.setIsFaultNodeValue(true)
		switch card.FaultType {
		case CardUnhealthy:
			fNode.setNodeHealthStateValue(NodeCardUnhealthy)
		case CardNetworkUnhealthy:
			fNode.setNodeHealthStateValue(NodeCardNetworkUnhealthy)
		default:
			klog.V(util.LogErrorLev).Infof("card health state not legal")
		}
		return nil
	}
	fNode.setNodeHealthStateValue(NodeHealthy)
	fNode.setIsFaultNodeValue(false)
	klog.V(util.LogDebugLev).Infof("Node %s health state set to %s", node.Name, NodeHealthy)

	// 3. Ensure nodeD enabled to get node heartbeat
	if !fNode.isNodeDEnabled(node) {
		klog.V(util.LogInfoLev).Infof("node %s nodeD not enabled", node.Name)
		fNode.setNodeDValue(false)
		return nil
	}
	fNode.setNodeDValue(true)

	// 4. last node heartbeat update time until now being greater than maxInterval indicates unhealthy
	if fNode.isNodeHealthyByHeartbeat() {
		fNode.setIsFaultNodeValue(false)
		fNode.setNodeHealthStateValue(NodeHealthy)
		return nil
	}
	klog.V(util.LogInfoLev).Infof("node %s is fault node for having wrong heartbeat", node.Name)
	fNode.setIsFaultNodeValue(true)
	fNode.setNodeHealthStateValue(NodeUnhealthy)
	return nil
}

func (fNode *FaultNode) isNodeDEnabled(node *plugin.NPUNode) bool {
	value, ok := node.Label[nodeDEnableKey]
	if !ok {
		return false
	}

	switch value {
	case nodeDEnableOnValue:
		return true
	case nodeDEnableOffValue:
		return false
	default:
		klog.V(util.LogErrorLev).Infof("isEnableFaultNode not support %s.", value)
		return false
	}
}

func (fNode *FaultNode) isNodeHealthyByHeartbeat() bool {
	maxInterval := int64(fNode.HeartbeatInterval) * util.MapInitNum
	nowTime := time.Now().Unix()
	latestInterval := nowTime - fNode.UpdateHeartbeatTime

	klog.V(util.LogDebugLev).Infof(
		"node %s latestInterval: %d, nowTime: %d", fNode.NodeName, latestInterval, nowTime)
	if latestInterval < 0 {
		klog.V(util.LogErrorLev).Infof(" isNodeHealth %s cache Time is newer[%d-%d], confused, skip.",
			fNode.NodeName, nowTime, fNode.UpdateHeartbeatTime)
	}
	if latestInterval > maxInterval {
		klog.V(util.LogErrorLev).Infof(" %s Time over %d [%d-%d],not health.",
			fNode.NodeName, maxInterval, nowTime, fNode.UpdateHeartbeatTime)
		return false
	}
	return true
}

func (fNode *FaultNode) getFaultCardIds(cardName string) ([]int, error) {
	if fNode.UnhealthyNPU == nil && fNode.NetworkUnhealthyNPU == nil {
		return nil, fmt.Errorf("no fault card on node")
	}
	allFaultCards := append(fNode.UnhealthyNPU, fNode.NetworkUnhealthyNPU...)
	faultCardIds := util.ChangeTopToIntArray(strings.Join(allFaultCards, ","), cardName)
	return faultCardIds, nil
}

// isNodeInSessionByNpuNodes judge if node is sent in session
func (fNode *FaultNode) isNodeInSessionByNpuNodes(nodes map[string]plugin.NPUNode) bool {
	_, ok := nodes[fNode.NodeName]
	return ok
}

func (fNode *FaultNode) setNodeDValue(value bool) {
	fNode.NodeDEnable = value
}

func (fNode *FaultNode) setIsFaultNodeValue(value bool) {
	fNode.IsFaultNode = value
}

func (fNode *FaultNode) setNodeHealthStateValue(nodeHealthState string) {
	fNode.NodeHealthState = nodeHealthState
}

func (fNode *FaultNode) setAllCardList(value []string) {
	fNode.AllCards = value
}

func (fNode *FaultNode) setUnhealthyNPUList(value []string) {
	fNode.UnhealthyNPU = value
}

func (fNode *FaultNode) setNetworkUnhealthyNPUList(value []string) {
	fNode.NetworkUnhealthyNPU = value
}

func (fNode *FaultNode) setUpdateTime(value int64) {
	fNode.UpdateTime = value
}

func (fNode *FaultNode) setFaultCards(value []FaultCard) {
	fNode.FaultCards = value
}

func (fNode *FaultNode) setOldNodeHeartbeatTime(value int64) {
	fNode.OldHeartbeatTime = value
}

func (fNode *FaultNode) setNewNodeHeartbeatTime(value int64) {
	fNode.NewHeartbeatTime = value
}

func (fCard *FaultCard) setFaultType(value string) {
	fCard.FaultType = value
}

func (fCard *FaultCard) setIsFaultCard(value bool) {
	fCard.IsFaultCard = value
}

func (fNode *FaultNode) setNodeHeartbeatInterval(value int) {
	fNode.HeartbeatInterval = value
}

func newFaultNodeDefault(nodeName string, updateTime int64) FaultNode {
	faultNode := FaultNode{
		NodeName:            nodeName,
		UpdateTime:          updateTime,
		UnhealthyNPU:        nil,
		NetworkUnhealthyNPU: nil,
		IsFaultNode:         false,
		NodeDEnable:         false,
		NodeHealthState:     NodeHealthy,
		AllCards:            nil,
		FaultCards:          nil,
		HeartbeatInterval:   0,
		OldHeartbeatTime:    0,
		UpdateHeartbeatTime: 0,
	}
	return faultNode
}