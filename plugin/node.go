/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package plugin is using for HuaWei Ascend pin affinity schedule frame.

*/
package plugin

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	v12 "k8s.io/kube-scheduler/extender/v1"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// NodeDeviceInfo like node annotation.
type NodeDeviceInfo struct {
	DeviceList map[string]string
	UpdateTime int64
}

// NodeDeviceInfoWithDevPlugin a node has one by cm.
type NodeDeviceInfoWithDevPlugin struct {
	DeviceInfo NodeDeviceInfo
	CheckCode  string
}

// NPUNode the plugin define node info.
type NPUNode struct {
	CommonNode
	VNode
}

// CommonNode common npu node properties
type CommonNode struct {
	Name       string
	Capability map[v1.ResourceName]float64
	Allocate   map[v1.ResourceName]float64
	Idle       map[v1.ResourceName]float64
	Annotation map[string]string
	Label      map[string]string
}

// VNode vnpu node class
type VNode struct {
	// Chips map chipID to VChip class
	Chips map[int]*VChip
	// ChipKind Ascend910/310/310p
	ChipKind string
	// ServerType Ascend310p-10-dual cardType-cardCoreNum-duo
	ServerType string
	// TotalChipNum num of total chips, get from capacity
	TotalChipNum int
	// FreeChipNum num of free chips get from device-info
	FreeChipNum int
	// TotalRes total resource on node
	TotalRes util.VResource
}

// VChip vnpu chip class
type VChip struct {
	PodMap map[string]*v1.Pod
	ID     []string
	// Name Ascend910-0
	Name string
	// Kind Ascend910/Ascend310/Ascend310P
	Kind        string
	isDual      bool
	CoreNum     int
	SegmentFlag bool
	TotalRes    util.VResource
	UsedRes     util.VResource
	FreeRes     util.VResource
}

// checkNodeDeviceInfo will be add more later
func checkNodeDeviceInfo(nodeData *NodeDeviceInfoWithDevPlugin) error {
	if nodeData == nil {
		return errors.New("nil parameters")
	}

	if nodeData.CheckCode == "" {
		return errors.New("checkCode is empty")
	}

	nodeDeviceInfo := nodeData.DeviceInfo
	if nodeData.CheckCode != MakeDataHash(nodeDeviceInfo) {
		return errors.New("checkCode is not match")
	}

	return nil
}

// MakeDataHash check code for configmap
func MakeDataHash(data interface{}) string {
	var dataBuffer []byte
	if dataBuffer = marshalData(data); len(dataBuffer) == 0 {
		return ""
	}
	h := sha256.New()
	if _, err := h.Write(dataBuffer); err != nil {
		klog.V(util.LogErrorLev).Infof("hash data error")
		return ""
	}
	sum := h.Sum(nil)
	return hex.EncodeToString(sum)
}

func marshalData(data interface{}) []byte {
	dataBuffer, err := json.Marshal(data)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("marshal data err: %#v", err)
		return nil
	}
	return dataBuffer
}

func getNodeDeviceInfoFromCM(kubeClient kubernetes.Interface, node *api.NodeInfo) (*NodeDeviceInfo, error) {
	cmData, getErr := util.GetConfigMapWithRetry(kubeClient, util.DevInfoNameSpace, util.DevInfoPreName+node.Name)
	if getErr != nil {
		klog.V(util.LogErrorLev).Infof("getNodeDeviceInfoFromCM :%#v.", getErr)
		return nil, getErr
	}

	devInf := &NodeDeviceInfoWithDevPlugin{}
	data, ok := cmData.Data[util.DevInfoCMKey]
	if !ok {
		return nil, fmt.Errorf("%s device-info no %s", node.Name, util.DevInfoCMKey)
	}
	if unmarshalErr := json.Unmarshal([]byte(data), &devInf); unmarshalErr != nil {
		klog.V(util.LogInfoLev).Infof("convertToReSchedulerJobsMapFromCM Unmarshal: %#v.", unmarshalErr)
		return nil, unmarshalErr
	}

	if checkErr := checkNodeDeviceInfo(devInf); checkErr != nil {
		klog.V(util.LogInfoLev).Infof("checkNodeDeviceInfo failed :%#v.", checkErr)
		return nil, checkErr
	}
	return &devInf.DeviceInfo, nil
}

// InitNPUNodeByNodeInf init NPU node from node info and cm.
func (n *NPUNode) InitNPUNodeByNodeInf(npuNode *api.NodeInfo, kubeClient kubernetes.Interface) error {
	if n == nil || npuNode == nil {
		klog.V(util.LogInfoLev).Infof("InitNPUNodeByNodeInf failed: %s.", util.ArgumentError)
		return errors.New(util.ArgumentError)
	}
	data, getErr := getNodeDeviceInfoFromCM(kubeClient, npuNode)
	if getErr != nil {
		klog.V(util.LogDebugLev).Infof("getNodeDeviceInfoFromCM %s %#v", npuNode.Name, getErr)
		return getErr
	}
	capability := npuNode.Capability.ScalarResources
	if !util.IsMapHasNPUResource(capability, util.NPUCardPreName) {
		return fmt.Errorf("%s not NPU node", npuNode.Name)
	}
	n.Name = npuNode.Name
	n.Capability = capability

	n.Allocate = npuNode.Allocatable.ScalarResources
	n.Idle = npuNode.Idle.ScalarResources
	n.Label = npuNode.Node.Labels
	n.Annotation = make(map[string]string, util.MapInitNum)
	for k, v := range npuNode.Node.Annotations {
		n.Annotation[k] = v
	}
	for k, v := range data.DeviceList {
		n.Annotation[k] = v
	}
	if setVNPUErr := n.setNodeVNPUInfo(npuNode); setVNPUErr != nil {
		klog.V(util.LogDebugLev).Infof("setNodeVNPUInfo %s %v", npuNode.Name, setVNPUErr)
	}
	return nil
}

// GetNewNPUNodeAnnotation get new annotation after allocate
func (n *NPUNode) GetNewNPUNodeAnnotation(usedTop []int, resourceName, resourceNamePre string) (string, error) {
	if n == nil || len(usedTop) == 0 || resourceName == "" || resourceNamePre == "" {
		klog.V(util.LogInfoLev).Infof("GetNewNPUNodeAnnotation failed: %s.", util.ArgumentError)
		return "", errors.New(util.ArgumentError)
	}
	annotation, ok := n.Annotation[resourceName]
	if !ok {
		err := fmt.Errorf("node<%s> not have resource<%s>", n.Name, resourceName)
		klog.V(util.LogErrorLev).Infof("GetNewNPUNodeAnnotation err: %s.", err.Error())
		return "", err
	}
	if annotation == "" {
		return "", nil
	}
	topStrArray := strings.Split(annotation, ",")
	var newTopStrArray []string
	for _, cardStr := range topStrArray {
		v := strings.TrimPrefix(cardStr, resourceNamePre)
		existFlag := false
		cardInt, err := strconv.Atoi(v)
		if err != nil {
			klog.V(util.LogErrorLev).Infof("ChangeTopToIntArray conv failed %v.", err)
			return "", err
		}
		for _, useId := range usedTop {
			if cardInt == useId {
				existFlag = true
				break
			}
		}
		if !existFlag {
			newTopStrArray = append(newTopStrArray, cardStr)
		}
	}
	newTopStr := strings.Join(newTopStrArray, ",")
	return newTopStr, nil
}

// CheckNPUResourceStable check resource stabilize.
func (n NPUNode) CheckNPUResourceStable(vcJob SchedulerJob) error {
	k := vcJob.handler.GetAnnoName()
	iNum, iOK := n.Idle[v1.ResourceName(k)]
	nodeA, aOK := n.Annotation[k]
	if iOK != true || aOK != true {
		return fmt.Errorf("%s not has(or not same) %s", n.Name, k)
	}

	sSlice := strings.Split(nodeA, ",")
	length := len(sSlice)
	if length == 1 && sSlice[0] == "" {
		length = 0
	}
	if length != int(iNum/util.NPUHexKilo) {
		return fmt.Errorf("%s's %s not statble:%d=>%d", n.Name, k, length, int(iNum/util.NPUHexKilo))
	}
	return nil
}

// CheckNPUResourceStableReScheduling check resource stabilize.
func (n NPUNode) CheckNPUResourceStableReScheduling(vcJob SchedulerJob) error {
	k := vcJob.handler.GetAnnoName()
	iNum, iOK := n.Idle[v1.ResourceName(k)]
	nodeA, aOK := n.Annotation[k]
	if iOK != true || aOK != true {
		return fmt.Errorf("%s not has(or not same) %s", n.Name, k)
	}

	sSlice := strings.Split(nodeA, ",")
	length := len(sSlice)
	if length == 1 && sSlice[0] == "" {
		length = 0
	}
	iNumInt := int(iNum / util.NPUHexKilo)
	if length != iNumInt && iNumInt >= 0 {
		return fmt.Errorf("%s's %s not stable:%d=>%d", n.Name, k, length, iNumInt)
	}
	return nil
}

// InitNodesFromSsn init all nodes in ssn.
func (sHandle *ScheduleHandler) InitNodesFromSsn(ssn *framework.Session) {
	if sHandle == nil || sHandle.FrameAttr.KubeClient == nil {
		return
	}
	sHandle.Nodes = make(map[string]NPUNode, util.MapInitNum)
	for nodeName, nodeInf := range ssn.Nodes {
		npuNode := NPUNode{}
		if err := npuNode.InitNPUNodeByNodeInf(nodeInf, sHandle.FrameAttr.KubeClient); err != nil {
			continue
		}
		sHandle.Nodes[nodeName] = npuNode
	}
	return
}

// NodePredicate Predicate nodes.
func (sHandle *ScheduleHandler) NodePredicate(taskInfo *api.TaskInfo, nodeInfo *api.NodeInfo) error {
	if sHandle == nil || taskInfo == nil || nodeInfo == nil {
		klog.V(util.LogErrorLev).Infof("NodePredicate got null parameter(s), which is invalid.")
		return fmt.Errorf("got null parameter(s)")
	}
	klog.V(util.LogInfoLev).Infof("enter node(%s) predicate", nodeInfo.Name)
	defer klog.V(util.LogInfoLev).Infof("leave node(%s) predicate", nodeInfo.Name)

	vcJob, ok := sHandle.Jobs[taskInfo.Job]
	if !ok {
		klog.V(util.LogInfoLev).Infof("NodePredicate not support job:%#v.", taskInfo.Job)
		return nil
	}
	// check vcjob is npu job
	if !vcJob.IsNPUJob() {
		klog.V(util.LogInfoLev).Infof("NodePredicate vc-job:%#v is not npu job.", vcJob)
		return nil
	}
	vcNode, ok := sHandle.Nodes[nodeInfo.Name]
	if !ok {
		klog.V(util.LogInfoLev).Infof("NodePredicate %s not in.", nodeInfo.Name)
		return nil
	}
	// task and frame conf has check before in job valid.
	if !util.IsSelectorMeetJob(vcJob.Selector, vcNode.Label) {
		meetErr := fmt.Errorf("job(%s) selector:%#v not meet node<%s> label or selector:%#v", vcJob.JobName,
			vcJob.Selector, vcNode.Name, vcNode.Label)
		klog.V(util.LogErrorLev).Infof(meetErr.Error())
		return meetErr
	}
	if err := vcNode.CheckNPUResourceStable(vcJob); err != nil {
		return err
	}
	if err := vcJob.CheckNodeNum(taskInfo, vcNode); err != nil {
		return err
	}
	if err := vcJob.handler.CheckNodeNPUByTask(taskInfo, vcNode); err != nil {
		// node doesn't have enough npu for the task
		klog.V(util.LogInfoLev).Infof("checkNodeNPUByTask %s:%#v ,cannot be selected.", vcNode.Name, err)
		return fmt.Errorf("checkNodeNPUByTask  %s : %s %#v", vcNode.Name, nodesNoMeetNPUReqError, err)
	}
	klog.V(util.LogInfoLev).Infof("%s NodePredicate %s select successes.", PluginName, vcNode.Name)
	return nil
}

func initScoreMap(nodes []*api.NodeInfo, interPodAffinityScore v12.HostPriorityList) map[string]float64 {
	for _, node := range nodes {
		if reflect.ValueOf(node).IsNil() {
			continue
		}

		interPodAffinityScore = append(interPodAffinityScore, v12.HostPriority{
			Host:  node.Name,
			Score: 0,
		})
	}
	scoreMap := make(map[string]float64, len(interPodAffinityScore))
	for _, host := range interPodAffinityScore {
		scoreMap[host.Host] = float64(host.Score)
	}

	return scoreMap
}
