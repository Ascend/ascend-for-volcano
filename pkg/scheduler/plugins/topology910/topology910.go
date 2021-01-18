/*
Copyright(C) 2020. Huawei Technologies Co.,Ltd. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/*

Package topology910 is using for HuaWei Ascend910 pin affinity schedule.

*/
package topology910

import (
	"errors"
	"fmt"
	"k8s.io/klog"
	schedulerapi "k8s.io/kubernetes/pkg/scheduler/api"
	"strconv"
	"strings"
	"time"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"
)

const (
	// PluginName indicates name of volcano scheduler plugin.
	PluginName = "topology910"
	// npuResourceName need to follow huawei's 910 device-plugin
	npuResourceName  = "huawei.com/Ascend910"
	podPredicateTime = "predicate-time"
	leftHccsName     = "leftHccs"
	rightHccsName    = "rightHccs"
	allHccsName      = "allHccs"
	npuCardNamePre   = "Ascend910-"
	archSelector     = "host-arch"
	npuHex           = 1000
	logErrorLev      = 1
	logInfoLev       = 3
	logDebugLev      = 4
	npuNumbPerNode   = 8
	priNpuGroupN     = 4
	npuNumPerHccs    = 4
	magicNumInt0     = 0
	magicNumInt1     = 1
	magicNumInt2     = 2
	magicNumInt3     = 3
)

type hccsInf struct {
	name     string
	availNum int
	top      []int
	nodeInf  *api.NodeInfo
}

type npuPriNodeInf struct {
	// the priority for TPU HCCS top
	Name        string
	allocateNum int
	idleNpuNum  int
	nodeName    string
	selHccsName string
	leftHccs    hccsInf
	rightHccs   hccsInf
}

type topology910plugin struct {
	// Arguments given for the plugin
	pluginArguments framework.Arguments
}

// New return gang plugin
func New(arguments framework.Arguments) framework.Plugin {
	return &topology910plugin{pluginArguments: arguments}
}

func (pp *topology910plugin) Name() string {
	return PluginName
}

type initPriNodeGroupFn func(priNodeGroup map[string]*npuPriNodeInf, groupName string, selectCards int)

// the police for one need card
func initOneCardPriNodeGroups(tmpLeftNpuAvaiNum int,
	tmpRightNpuAvaiNum int,
	priNodeGroups []map[string]*npuPriNodeInf,
	nodeName string,
	addPriNodeGroupFn initPriNodeGroupFn) error {
	// priority:1>3>2>4
	if tmpLeftNpuAvaiNum == magicNumInt1 || tmpRightNpuAvaiNum == magicNumInt1 {
		// A group
		addPriNodeGroupFn(priNodeGroups[magicNumInt0], "A", magicNumInt1)
		return nil
	}

	if tmpLeftNpuAvaiNum == magicNumInt3 || tmpRightNpuAvaiNum == magicNumInt3 {
		// B group
		addPriNodeGroupFn(priNodeGroups[magicNumInt1], "B", magicNumInt3)
		return nil
	}

	if tmpLeftNpuAvaiNum == magicNumInt2 || tmpRightNpuAvaiNum == magicNumInt2 {
		// C group
		addPriNodeGroupFn(priNodeGroups[magicNumInt2], "C", magicNumInt2)
		return nil
	}

	if tmpLeftNpuAvaiNum == npuNumPerHccs || tmpRightNpuAvaiNum == npuNumPerHccs {
		// D group
		addPriNodeGroupFn(priNodeGroups[magicNumInt3], "D", npuNumPerHccs)
		return nil
	}

	// no satisfy HCCS,8*N need whole nodes
	klog.V(logErrorLev).Infof("%s initOneCardPriNodeGroups node(%s) (left :%d,right :%d) cannot fit", PluginName, nodeName, tmpLeftNpuAvaiNum, tmpRightNpuAvaiNum)
	return errors.New("node no fit npu number")
}

// the police for Two need card
func initTwoCardPriNodeGroups(tmpLeftNpuAvaiNum int,
	tmpRightNpuAvaiNum int,
	priNodeGroups []map[string]*npuPriNodeInf,
	nodeName string,
	addPriNodeGroupFn initPriNodeGroupFn) error {

	// priority：2>npuNumPerHccs>3
	if tmpLeftNpuAvaiNum == magicNumInt2 || tmpRightNpuAvaiNum == magicNumInt2 {
		// A group
		addPriNodeGroupFn(priNodeGroups[magicNumInt0], "A", magicNumInt2)
		return nil
	}

	if tmpLeftNpuAvaiNum == npuNumPerHccs || tmpRightNpuAvaiNum == npuNumPerHccs {
		// B group
		addPriNodeGroupFn(priNodeGroups[magicNumInt1], "B", npuNumPerHccs)
		return nil
	}

	if tmpLeftNpuAvaiNum == magicNumInt3 || tmpRightNpuAvaiNum == magicNumInt3 {
		// C group
		addPriNodeGroupFn(priNodeGroups[magicNumInt2], "C", magicNumInt3)
		return nil
	}

	// no satisfy HCCS,8*N need whole nodes
	klog.V(logErrorLev).Infof("%s initTwoCardPriNodeGroups node(%s) (left :%d,right :%d) cannot fit 2", PluginName, nodeName, tmpLeftNpuAvaiNum, tmpRightNpuAvaiNum)
	return errors.New("node no fit npu number")
}

// the police four one need card
func initFourCardPriNodeGroups(tmpLeftNpuAvaiNum int,
	tmpRightNpuAvaiNum int,
	priNodeGroups []map[string]*npuPriNodeInf,
	nodeName string,
	addPriNodeGroupFn initPriNodeGroupFn) error {
	// only can allocate 4
	if tmpLeftNpuAvaiNum == npuNumPerHccs || tmpRightNpuAvaiNum == npuNumPerHccs {
		// A group
		addPriNodeGroupFn(priNodeGroups[magicNumInt0], "A", npuNumPerHccs)
		return nil
	}
	// no satisfy HCCS,8*N need whole nodes
	klog.V(logErrorLev).Infof("%s initFoureCardPriNodeGroups node(%s) (left :%d,right :%d) cannot fit 4", PluginName, nodeName, tmpLeftNpuAvaiNum, tmpRightNpuAvaiNum)
	return errors.New("node no fit npu number")
}

// the police four one need card
func initEightCardPriNodeGroups(tmpAllNpuAvaiNum int,
	priNodeGroups []map[string]*npuPriNodeInf,
	nodeName string,
	addPriNodeGroupFn initPriNodeGroupFn) error {

	if tmpAllNpuAvaiNum == npuNumbPerNode {
		addPriNodeGroupFn(priNodeGroups[magicNumInt0], "A", npuNumbPerNode)
		return nil
	}

	klog.V(logErrorLev).Infof("%s initEightCardPriNodeGroups node(%s) (all:%d) cannot fit 8", PluginName, nodeName, tmpAllNpuAvaiNum)
	return errors.New("node no fit npu number")
}

// insert into group by police
func insertNodeInPriGroup(taskReqNPU int,
	tmpLeftNpuAvaiNum int,
	tmpRightNpuAvaiNum int,
	tmpAllNpuAvaiNum int,
	nodeName string,
	priNodeGroups []map[string]*npuPriNodeInf,
	addPriNodeGroupFn initPriNodeGroupFn) error {

	var err error
	switch taskReqNPU {
	case magicNumInt0:
		klog.V(logInfoLev).Infof("%s initPriNodeGroups task npu is 0", PluginName)
	case magicNumInt1:
		err = initOneCardPriNodeGroups(tmpLeftNpuAvaiNum, tmpRightNpuAvaiNum, priNodeGroups, nodeName, addPriNodeGroupFn)
	case magicNumInt2:
		err = initTwoCardPriNodeGroups(tmpLeftNpuAvaiNum, tmpRightNpuAvaiNum, priNodeGroups, nodeName, addPriNodeGroupFn)
	case npuNumPerHccs:
		err = initFourCardPriNodeGroups(tmpLeftNpuAvaiNum, tmpRightNpuAvaiNum, priNodeGroups, nodeName, addPriNodeGroupFn)
	case npuNumbPerNode:
		err = initEightCardPriNodeGroups(tmpAllNpuAvaiNum, priNodeGroups, nodeName, addPriNodeGroupFn)
	default:
		// For normal,can not be here. The pre function validate job has done this.
		klog.V(logErrorLev).Infof("%s addPriNodeGroupFn node(%s) (left :%d,right :%d) cannot fit %d,illegal task request npu number",
			PluginName, nodeName, tmpLeftNpuAvaiNum, tmpRightNpuAvaiNum, taskReqNPU)
		err = errors.New("illegal request npu number " + strconv.Itoa(taskReqNPU))
	}

	return err
}

// choose the right hccs ring in node
func selectDireTopFn(selectCardNum int, tmpLeftNpuAvaiNum int, tmpRightNpuAvaiNum int, tmpAllNpuAvaiNum int) string {
	if tmpLeftNpuAvaiNum == selectCardNum {
		return leftHccsName
	} else if tmpRightNpuAvaiNum == selectCardNum {
		return rightHccsName
	} else if tmpAllNpuAvaiNum == selectCardNum {
		return allHccsName
	} else {
		return ""
	}
}

// init 4 pri node-list group
func initPriNodeGroups(taskReqNPU int, nodes []*api.NodeInfo, priNodeGroups []map[string]*npuPriNodeInf) error {
	var err error
	for i := magicNumInt0; i < npuNumPerHccs; i++ {
		priNodeGroups[i] = make(map[string]*npuPriNodeInf)
	}

	// init pri Node group
	for _, node := range nodes {
		var leftHccsTop []int
		var rightHccsTop []int
		var tmpAllNpuAvaiNum int
		var tmpLeftNpuAvaiNum int
		var tmpRightNpuAvaiNum int

		// get node NPU top
		tmpAllNpuAvaiNum = 0
		tmpLeftNpuAvaiNum = 0
		tmpRightNpuAvaiNum = 0

		cardIds := getTopFromNode(node)
		klog.V(logDebugLev).Infof("%s initPriNodeGroups:%v", PluginName, cardIds)
		for _, cardID := range cardIds {
			tmpAllNpuAvaiNum++

			if cardID < npuNumPerHccs {
				tmpLeftNpuAvaiNum++
				leftHccsTop = append(leftHccsTop, cardID)
			} else {
				tmpRightNpuAvaiNum++
				rightHccsTop = append(rightHccsTop, cardID)
			}
		}
		// set the meet node into its pri-node-list group
		addPriNodeGroupFn := func(priNodeGroup map[string]*npuPriNodeInf, groupName string, selectCards int) {
			klog.V(logDebugLev).Infof("%s addPriNodeGroupFn nodeName:%s,group:%v", PluginName, node.Name, priNodeGroup[node.Name])
			priNodeGroup[node.Name] = &npuPriNodeInf{
				Name:        groupName,
				idleNpuNum:  tmpAllNpuAvaiNum,
				allocateNum: int(node.Allocatable.ScalarResources[npuResourceName]) / npuHex,
				selHccsName: selectDireTopFn(selectCards, tmpLeftNpuAvaiNum, tmpRightNpuAvaiNum, tmpAllNpuAvaiNum),
				leftHccs:    hccsInf{name: leftHccsName, availNum: tmpLeftNpuAvaiNum, top: leftHccsTop, nodeInf: node},
				rightHccs:   hccsInf{name: rightHccsName, availNum: tmpRightNpuAvaiNum, top: rightHccsTop, nodeInf: node},
				nodeName:    node.Name,
			}
			klog.V(logDebugLev).Infof("%s addPriNodeGroupFn node name:%s priNode:%v", PluginName, node.Name, priNodeGroup[node.Name])
		}
		// insert into group by police
		err = insertNodeInPriGroup(taskReqNPU, tmpLeftNpuAvaiNum, tmpRightNpuAvaiNum, tmpAllNpuAvaiNum, node.Name, priNodeGroups, addPriNodeGroupFn)
		if err != nil {
			continue
		}
	}
	return nil
}

func getBestPriNodeGroup(priNodeGroups []map[string]*npuPriNodeInf) (map[string]*npuPriNodeInf, int, error) {
	var i int
	var selectPriGroup map[string]*npuPriNodeInf
	capacity := npuNumbPerNode
	loopFlg := true

	findNodeInPriGroupFn := func() bool {
		for _, nodeTpu := range priNodeGroups[i] {
			// if capacity suitable,the node is meet the required.
			if nodeTpu.allocateNum == capacity {
				// right one
				selectPriGroup = priNodeGroups[i]
				loopFlg = false
				break
			}
		}
		return loopFlg
	}
	// loop capacity
	for (capacity != magicNumInt0) && loopFlg {
		// loop 4 pri-node-list-group
		i = magicNumInt0
		for i < npuNumPerHccs && loopFlg {
			// to find node
			loopFlg = findNodeInPriGroupFn()
			// next pri-node-list-group
			i++
		}
		// next capacity
		capacity--
	}

	if loopFlg {
		return selectPriGroup, capacity, errors.New("pri group nil")
	}

	return selectPriGroup, capacity + 1, nil
}

// get best node in priNodeGroup
func getBestNodeFromPriNodeGroup(selectPriGroup map[string]*npuPriNodeInf) (*npuPriNodeInf, bool) {
	var tmpPriNodeInf *npuPriNodeInf
	var capacity int
	var loopFlg bool

	capacity = npuNumbPerNode
	loopFlg = true
	tmpPriNodeInf = &npuPriNodeInf{
		idleNpuNum: npuNumbPerNode + 1,
	}

	getBetNodeFromGroupFn := func(capacity int, priNodeInf *npuPriNodeInf) (*npuPriNodeInf, bool) {
		// only need find least one,it is best
		if priNodeInf.allocateNum == capacity {
			if tmpPriNodeInf.idleNpuNum > priNodeInf.idleNpuNum {
				tmpPriNodeInf = priNodeInf
				// capacity is right
				loopFlg = false
			}
		}
		return tmpPriNodeInf, loopFlg
	}

	// loop capacity
	for (capacity != magicNumInt0) && loopFlg {
		for _, priNodeInf := range selectPriGroup {
			tmpPriNodeInf, loopFlg = getBetNodeFromGroupFn(capacity, priNodeInf)
		}
		capacity--
	}

	return tmpPriNodeInf, loopFlg
}

func getBestHccsFromSelectNode(needCards int, loopFlg bool, tmpPriNodeInf *npuPriNodeInf) (hccsInf, error) {
	var selectHCCS hccsInf
	selectHCCS = hccsInf{}

	if (needCards <= npuNumPerHccs) && loopFlg {
		// for normal can not be here.the preceded function has hold on
		klog.V(logErrorLev).Infof("%s getBestNodeAndTop failed :get none node", PluginName)
		return selectHCCS, errors.New("none matched")
	}

	if (needCards <= npuNumPerHccs) && !loopFlg {
		// choose bestHCCS
		if tmpPriNodeInf.selHccsName == leftHccsName {
			return tmpPriNodeInf.leftHccs, nil
		}

		return tmpPriNodeInf.rightHccs, nil
	}

	if (needCards == npuNumbPerNode) && !loopFlg {
		// need whole node
		return hccsInf{}, nil
	}
	// the request number illegal
	return selectHCCS, errors.New("error require cards")
}

// According to need card number,get best node from 4 pri-node-list-group
func getBestNodeAndTop(needCards int, priNodeGroups []map[string]*npuPriNodeInf) (string, *hccsInf, error) {
	var selectPriGroup map[string]*npuPriNodeInf
	var tmpPriNodeInf *npuPriNodeInf
	var selectHCCS hccsInf
	var err error
	var capacity int
	var loopFlg bool
	// get best pri-node-list-group
	selectPriGroup, capacity, err = getBestPriNodeGroup(priNodeGroups)
	if err != nil {
		// cannot be here,for predicate function has hold up.
		klog.V(logErrorLev).Infof("%s getBestNodeAndTop failed capacity:%d", PluginName, capacity)
		return "", nil, err
	}
	klog.V(logInfoLev).Infof("%s getBestNodeAndTop capacity:%d", PluginName, capacity)

	// Find best node and hccs in selected pri-node-list-group
	// for best node
	tmpPriNodeInf, loopFlg = getBestNodeFromPriNodeGroup(selectPriGroup)

	// select hccs
	selectHCCS, err = getBestHccsFromSelectNode(needCards, loopFlg, tmpPriNodeInf)
	if err != nil {
		return "", nil, err
	}

	klog.V(logDebugLev).Infof("%s getBestNodeAndTop nodeInf:%s hccs:%v", PluginName, tmpPriNodeInf.nodeName, selectHCCS)
	return tmpPriNodeInf.nodeName, &selectHCCS, nil
}

func setSelectTopValue(needCards int, selectHCCS *hccsInf) ([]int, error) {
	var setTop []int
	i := magicNumInt0
	// select whole node
	if needCards == npuNumbPerNode {
		for i = magicNumInt0; i < npuNumbPerNode; i++ {
			setTop = append(setTop, i)
		}
		return setTop, nil
	}

	i = magicNumInt0
	for _, cardID := range selectHCCS.top {
		if i < needCards {
			setTop = append(setTop, cardID)
			i++
		} else {
			break
		}
	}
	return setTop, nil
}

// get the best node
func getNpuAffinityBestNodeAndTop(taskReqNPU int, nodes []*api.NodeInfo) (string, []int, error) {
	var top []int
	priNodeGroups := make([]map[string]*npuPriNodeInf, priNpuGroupN)
	// 1. inti 4 pri node-list group
	err := initPriNodeGroups(taskReqNPU, nodes, priNodeGroups)
	if err != nil {
		klog.V(logErrorLev).Infof("%s getNpuAffinityBestNodeAndTop initPriNodeGroups failed :%s", PluginName, err)
		return "", nil, err
	}
	// 2.get the right node and its hccs
	selectNodeName, selectHCCS, err := getBestNodeAndTop(taskReqNPU, priNodeGroups)
	if err != nil {
		klog.V(logErrorLev).Infof("%s getNpuAffinityBestNodeAndTop getBestNodeAndTop failed :%s", PluginName, err)
		return "", nil, err
	}
	// 3.set the selected top,selectHCCS.top
	top, err = setSelectTopValue(taskReqNPU, selectHCCS)
	if err != nil {
		klog.V(logErrorLev).Infof("%s getNpuAffinityBestNodeAndTop setSelectTopValue failed :%s", PluginName, err)
		return "", nil, err
	}

	klog.V(logInfoLev).Infof("%s setSelectTopValue top :%v", PluginName, top)
	return selectNodeName, top, nil
}

// select the best node for task using
func scoreAndSetSelectNodes(task *api.TaskInfo, nodes []*api.NodeInfo, scoreMp map[string]float64, bestNodeName string, top []int) (map[string]float64, error) {
	var nodeWeight float64
	nodeWeight = 1.0

	for _, nodeInfo := range nodes {
		if nodeInfo.Name != bestNodeName {
			scoreMp[nodeInfo.Name] = 0.0
			continue
		}

		// write the topology to task's pod
		err := setTpuTopologyToPod(task, nodeInfo, top)
		if err != nil {
			return nil, err
		}

		scoreMp[nodeInfo.Name] = nodeWeight
		break
	}

	return scoreMp, nil
}

func changeIntArrToStr(top []int) string {
	var tmp int
	var str string

	i := 0
	for i, tmp = range top {
		str += npuCardNamePre + strconv.Itoa(tmp)
		if i+1 < len(top) {
			str += ","
		}
	}

	return str
}

func checkResourceNumMeet(task *api.TaskInfo, nodeInfo *api.NodeInfo) error {
	if !task.InitResreq.LessEqual(nodeInfo.Idle) {
		return errors.New("IN Pipeline")
	}
	return nil
}

func getNodeNpuNumFromAnnotation(nodeInfo *api.NodeInfo) (int, error) {
	top := getTopFromNode(nodeInfo)
	if top == nil {
		return 0, fmt.Errorf("nil node(%s) top", nodeInfo.Name)
	}
	nodeNpuIdleNumFromTop := len(top)

	return nodeNpuIdleNumFromTop, nil
}

func checkNpuResourceStable(nodeInfo *api.NodeInfo) error {
	nodeNpuIdleNumFromTop, err := getNodeNpuNumFromAnnotation(nodeInfo)
	if err != nil {
		return err
	}

	nodeNpuIdleNumFromIdle, err := getNodeNpuNumFromIdle(nodeInfo)
	if err != nil {
		return err
	}

	return checkNodeNpuStabilize(nodeNpuIdleNumFromTop, nodeNpuIdleNumFromIdle)
}

func checkNodeNpuStabilize(nodeNpuIdleNumFromTop int, nodeNpuIdleNumFromIdle int) error {
	if nodeNpuIdleNumFromTop != nodeNpuIdleNumFromIdle {
		return fmt.Errorf("node annotations(%d) not same node idle(%d)", nodeNpuIdleNumFromTop, nodeNpuIdleNumFromIdle)
	}

	return nil
}

func getNodeNpuNumFromIdle(nodeInfo *api.NodeInfo) (int, error) {
	nodeNpuIdleNumFromIdle, ok := nodeInfo.Idle.ScalarResources[npuResourceName]
	if !ok {
		klog.V(logErrorLev).Infof("%s getNodeNpuNumFromIdle failed", PluginName)
		return 0, errors.New("get node idle npu failed")
	}

	return int(nodeNpuIdleNumFromIdle / npuHex), nil
}

func checkResourceReady(task *api.TaskInfo, nodeInfo *api.NodeInfo) error {
	// check resource number fit
	err := checkResourceNumMeet(task, nodeInfo)
	if err != nil {
		return err
	}

	return nil
}

func doSetPodNpuTopology(top []int, task *api.TaskInfo) error {
	var topologyStr string

	klog.V(logDebugLev).Infof("%s setTpuTopologyToPod begin top:%v", PluginName, top)
	topologyStr = changeIntArrToStr(top)
	task.Pod.Annotations[npuResourceName] = topologyStr
	// to device-plugin judge pending pod.
	task.Pod.Annotations[podPredicateTime] = strconv.FormatInt(time.Now().UnixNano(), 10)
	klog.V(logInfoLev).Infof("%s setTpuTopologyToPod task:%s top:%s", PluginName, task.Name, topologyStr)

	return nil
}

// write the topology to pod
// to change the top kind :int[] to device-plugin point kind:string link with ','
func setTpuTopologyToPod(task *api.TaskInfo, nodeInfo *api.NodeInfo, top []int) error {
	err := checkResourceReady(task, nodeInfo)
	if err != nil {
		return err
	}

	return doSetPodNpuTopology(top, task)
}

func getTopToIntArray(topStr string) []int {
	var topInt []int
	var cardInt int
	var cardStr string
	var err error
	var topStrArray []string

	if topStr == "" {
		return []int{}
	}

	topStrArray = strings.Split(topStr, ",")
	for _, cardStr = range topStrArray {
		klog.V(logDebugLev).Infof("%s getTopFromNode cardStr %s", PluginName, cardStr)
		// cannot use strings.Trim()
		v := strings.TrimPrefix(cardStr, npuCardNamePre)
		klog.V(logDebugLev).Infof("%s getTopFromNode cardStr2 %s", PluginName, v)
		cardInt, err = strconv.Atoi(v)
		if err != nil {
			klog.V(logErrorLev).Infof("%s getTopFromNode conv failed %v", PluginName, err)
			return nil
		}

		topInt = append(topInt, cardInt)
	}

	return topInt
}

// get top kind like（int[]） from node inf
func getTopFromNode(node *api.NodeInfo) []int {
	var topInt []int

	topStr, ok := node.Node.Annotations[npuResourceName]
	if !ok {
		klog.V(logErrorLev).Infof("%s getTopFromNode top nil:%v", PluginName, node.Name)
		return nil
	}

	topInt = getTopToIntArray(topStr)
	if topInt == nil {
		klog.V(logInfoLev).Infof("%s getTopFromNode %s nil(%s)", PluginName, node.Name, topStr)
		return nil
	}

	klog.V(logDebugLev).Infof("%s getTopFromNode int: %v, s: %s", PluginName, topInt, topStr)
	return topInt
}

func getJobReqNpuNum(job *api.JobInfo) (int, error) {
	jobNpu, ok := job.TotalRequest.ScalarResources[npuResourceName]
	if !ok {
		return 0, errors.New("not npu job")
	}

	return int(jobNpu / npuHex), nil
}

func validJobNpuTotalNumber(job *api.JobInfo) error {
	var jobNpu int
	var err error

	if jobNpu, err = getJobReqNpuNum(job); err != nil {
		return err
	}

	if jobNpu == magicNumInt1 ||
		jobNpu == magicNumInt2 ||
		jobNpu == npuNumPerHccs ||
		jobNpu%npuNumbPerNode == 0 {
		return nil
	}

	return fmt.Errorf("illegal req_npu num:%d", jobNpu)
}

func getTaskNpuNum(task *api.TaskInfo) (int, error) {
	tmpNpu, ok := task.Resreq.ScalarResources[npuResourceName]
	if !ok {
		return 0, errors.New("not npu task")
	}

	taskNpu := int(tmpNpu / npuHex)
	return taskNpu, nil
}

// less 8 npu, can only one task.
func checkSingleTrainMode(job *api.JobInfo) error {
	taskNum := len(job.Tasks)

	klog.V(logDebugLev).Infof("%s checkSingleTrainMode job(%s) has %d tasks", PluginName, job.Name, taskNum)

	if taskNum > magicNumInt1 {
		return fmt.Errorf("%s single trainning has too many task:%d", job.Name, taskNum)
	}

	return nil
}

// more 8 npu required,every task need 8 npu.
func checkDistributeTrainMode(job *api.JobInfo) error {
	taskNum := len(job.Tasks)

	klog.V(logDebugLev).Infof("%s checkDistributeTrainMode job(%s) has %d tasks", PluginName, job.Name, taskNum)

	for _, task := range job.Tasks {
		taskNpu, taskError := getTaskNpuNum(task)
		if taskError != nil {
			return errors.New("no npu task")
		}

		klog.V(logDebugLev).Infof("%s checkDistributeTrainMode task(%s) has %d npu", PluginName, task.Name, taskNpu)

		if taskNpu != npuNumbPerNode {
			return fmt.Errorf("DistributeTrain Job: %s  has %d tasks, and req npu illegal: %d", job.Name, taskNum, taskNpu)
		}
	}

	return nil
}

func validJobTaskModel(job *api.JobInfo) error {
	var jobNpu int
	var err error

	if jobNpu, err = getJobReqNpuNum(job); err != nil {
		return err
	}

	if jobNpu <= npuNumbPerNode {
		if err = checkSingleTrainMode(job); err != nil {
			return err
		}
	}

	if jobNpu > npuNumbPerNode {
		if err = checkDistributeTrainMode(job); err != nil {
			return err
		}
	}

	return nil
}

func getJobHandle(obj interface{}) *api.JobInfo {
	job, ok := obj.(*api.JobInfo)
	if !ok {
		klog.V(logErrorLev).Infof("%s job valid Failed to convert <%v> to *JobInfo", PluginName, obj)
		return nil
	}

	return job
}

func isNpuJob(job *api.JobInfo) error {
	if _, err := getJobReqNpuNum(job); err != nil {
		return err
	}

	return nil
}

func validNpuJob(job *api.JobInfo) *api.ValidateResult {
	if errJob := validJobNpuTotalNumber(job); errJob != nil {
		klog.V(logErrorLev).Infof("%s JobName: %s, err: %v,", PluginName, job.Name, errJob)

		return &api.ValidateResult{
			Pass:    false,
			Reason:  "illegal job parameter",
			Message: fmt.Sprintf("JobName: %s, err:%v", job.Name, errJob),
		}
	}

	if errTask := validJobTaskModel(job); errTask != nil {
		klog.V(logErrorLev).Infof("%s err: %v", PluginName, errTask)
		return &api.ValidateResult{
			Pass:    false,
			Reason:  "illegal task number",
			Message: fmt.Sprintf("JobName: %s, err: %v", job.Name, errTask),
		}
	}

	return nil
}

// open-session register for job validate fun
// in this function cannot check nodes npu,base on job validate function
func validJobFn(obj interface{}) *api.ValidateResult {
	klog.V(logInfoLev).Infof("%s enter job valid", PluginName)
	defer klog.V(logInfoLev).Infof("%s leave job valid", PluginName)

	job := getJobHandle(obj)
	if job == nil {
		klog.V(logErrorLev).Infof("%s validJobFn convert <%v> failed", PluginName, obj)
		return &api.ValidateResult{
			Pass:    false,
			Message: fmt.Sprintf("%s validJobFn convert <%v> failed", PluginName, obj),
		}
	}

	if err := isNpuJob(job); err != nil {
		klog.V(logInfoLev).Infof("%s job(%s),err: %v", PluginName, job.Name, err)
		// to be Compatible with CPU scenarios ,cannot return error
		return nil
	}

	result := validNpuJob(job)
	if result != nil {
		klog.V(logErrorLev).Infof("%s validNpuJob failed:%v", PluginName, result.Message)

		return result
	}

	klog.V(logInfoLev).Infof("%s npu affinity check ok, JobName: %s, req:%v", PluginName, job.Name, job.TotalRequest)

	return nil
}

func judgeNodeAndTaskNpu(taskNpu int, nodeNpuTopology []int) error {
	var reFlag bool
	reFlag = false

	// record the npu card number of HCCS rings
	leftCardNum, rightCardNum := getNodeHccsCardNum(nodeNpuTopology, taskNpu)

	switch taskNpu {
	case magicNumInt0:
		return nil
	case magicNumInt1:
		reFlag = (leftCardNum > magicNumInt0) || (rightCardNum > magicNumInt0)
	case magicNumInt2:
		reFlag = (leftCardNum > magicNumInt1) || (rightCardNum > magicNumInt1)
	case npuNumPerHccs:
		reFlag = (leftCardNum == npuNumPerHccs) || (rightCardNum == npuNumPerHccs)
	case npuNumbPerNode:
		reFlag = (leftCardNum + rightCardNum) == npuNumbPerNode
	default:
		// single pod(task) cannot require more than 8 npu
		klog.V(logErrorLev).Infof("%s Predicate jobs req more than 8 NPUs :%d", PluginName, taskNpu)
	}

	if reFlag {
		return nil
	}

	return errors.New("no meet")
}

func isNpuTask(task *api.TaskInfo) error {
	_, ok := task.Resreq.ScalarResources[npuResourceName]
	if !ok {
		return errors.New("not npu task")
	}

	return nil
}

func isNpuNode(node *api.NodeInfo) error {
	_, ok := node.Node.Annotations[npuResourceName]
	if !ok {
		return errors.New("no npu node")
	}

	return nil
}

func checkNodeNpuByTask(task *api.TaskInfo, node *api.NodeInfo) error {

	taskNpu, taskError := getTaskNpuNum(task)
	if taskError != nil {
		return taskError
	}

	nodeNpuTopology := getTopFromNode(node)
	if nodeNpuTopology == nil {
		// node has none npu
		klog.V(logInfoLev).Infof("%s checkNodeNpuByTask nil,node name:%s,task req npu:%d", PluginName, node.Name, taskNpu)
		return errors.New("node no available npu")
	}
	klog.V(logInfoLev).Infof("%s checkNodeNpuByTask node(%s)top:%v", PluginName, node.Name, nodeNpuTopology)

	err := judgeNodeAndTaskNpu(taskNpu, nodeNpuTopology)
	if err != nil {
		return err
	}

	return nil
}

func getNodeHccsCardNum(nodeNpuTopology []int, taskNpu int) (int, int) {
	var leftCardNum int
	var rightCardNum int
	var carID int
	leftCardNum = 0
	rightCardNum = 0
	for _, carID = range nodeNpuTopology {
		// 0~3 is a hccs ring,4~7 is an other one
		if carID < npuNumPerHccs {
			leftCardNum++
		} else {
			rightCardNum++
		}
	}
	klog.V(logInfoLev).Infof("%s Predicate get Top NPU req num:%v,left num:%d,right num:%d", PluginName, taskNpu, leftCardNum, rightCardNum)
	return leftCardNum, rightCardNum
}

func getTaskSelector(task *api.TaskInfo) (string, error) {
	taskSelector, ok := task.Pod.Spec.NodeSelector[archSelector]
	if !ok {
		return "", errors.New("not Selector")
	}

	return taskSelector, nil
}

func getNodeSelector(node *api.NodeInfo) (string, error) {
	nodeSelector, ok := node.Node.Labels[archSelector]
	if !ok {
		return "", errors.New("not Selector")
	}

	return nodeSelector, nil
}

// task or node has no selector will failed
func isSelectorMeetNode(task *api.TaskInfo, node *api.NodeInfo) error {
	taskSelector, errTask := getTaskSelector(task)
	if errTask != nil {
		klog.V(logErrorLev).Infof("%s task(%s) has no selector", PluginName, task.Name)
		return fmt.Errorf("task(%s) no selector", task.Name)
	}

	nodeSelector, errNode := getNodeSelector(node)
	if errNode != nil {
		klog.V(logErrorLev).Infof("%s node(%s) has no selector", PluginName, task.Name)
		return fmt.Errorf("node(%s) no selector", node.Name)
	}

	if taskSelector != nodeSelector {
		klog.V(logInfoLev).Infof("%s selector: task：%v, node %v", PluginName, taskSelector, nodeSelector)
		return fmt.Errorf("selector not equal: task(%s) node(%s) ", taskSelector, nodeSelector)
	}

	return nil
}

// open-session register for node predicate fun
func npuPredicate(task *api.TaskInfo, node *api.NodeInfo) error {
	if err := isNpuTask(task); err != nil {
		// task no need npu, and other predicateFn need work, so cannot return err
		klog.V(logInfoLev).Infof("%s isNpuTask %s : %v", PluginName, task.Name, err)
		return nil
	}

	if err := isNpuNode(node); err != nil {
		// node no npu or k8s is not the match, cannot be selected
		klog.V(logInfoLev).Infof("%s isNpuNode %s : %v ,cannot be selected.", PluginName, node.Name, err)
		return err
	}

	// select node by architect
	if err := isSelectorMeetNode(task, node); err != nil {
		// node no npu or k8s is not the match, cannot be selected
		klog.V(logInfoLev).Infof("%s selector Node %s : %v ,cannot be selected.", PluginName, node.Name, err)
		return err
	}

	// check resource stabilize
	if err := checkNpuResourceStable(node); err != nil {
		// npu not be Stable by k8s,cannot select.
		klog.V(logInfoLev).Infof("%s checkNpuResourceStable %s : %v ,cannot be selected.", PluginName, node.Name, err)
		return err
	}

	if err := checkNodeNpuByTask(task, node); err != nil {
		// npu not be Stable by k8s,cannot select.
		klog.V(logInfoLev).Infof("%s checkNodeNpuByTask :%v ,cannot be selected.", PluginName, err)
		return err
	}

	return nil
}

func initScoreMp(nodes []*api.NodeInfo, interPodAffinityScore schedulerapi.HostPriorityList) map[string]float64 {
	for _, node := range nodes {
		interPodAffinityScore = append(interPodAffinityScore, schedulerapi.HostPriority{
			Host:  node.Name,
			Score: 0,
		})
	}
	scoreMp := make(map[string]float64, len(interPodAffinityScore))
	for _, host := range interPodAffinityScore {
		scoreMp[host.Host] = float64(host.Score)
	}
	return scoreMp
}

// open-session register for batch Node Order fun
// for volcano frame,can not return error
func batchNodeOrderFn(task *api.TaskInfo, nodes []*api.NodeInfo) (map[string]float64, error) {
	var interPodAffinityScore schedulerapi.HostPriorityList
	var bestNodeName string
	var top []int
	var errGet error

	klog.V(logInfoLev).Infof("%s Enter batchNodeOrderFn", PluginName)
	defer klog.V(logInfoLev).Infof("%s leaving batchNodeOrderFn", PluginName)

	// init score-map
	scoreMp := initScoreMp(nodes, interPodAffinityScore)

	// 1.Get the task's NPU request
	taskReqNPU, taskError := getTaskNpuNum(task)
	if taskError != nil {
		// cannot return error for task is no npu kind possible.
		klog.V(logErrorLev).Infof("%s batchNodeOrderFn task :%s,%v", PluginName, task.Name, taskError)
		// cannot return nil，will panic
		return scoreMp, nil
	}
	klog.V(logInfoLev).Infof("%s batchNodeOrderFn Get %s taskReqNPU:%v", PluginName, task.Name, taskReqNPU)

	// 2.get the best node and top by A,B,C,D rules and require numbers.
	bestNodeName, top, errGet = getNpuAffinityBestNodeAndTop(taskReqNPU, nodes)
	if errGet != nil {
		// get suitable node failed
		klog.V(logErrorLev).Infof("%s batchNodeOrderFn find none fit node for task:%s,require npu:%d", PluginName, task.Name, taskReqNPU)
		return scoreMp, nil
	}
	klog.V(logInfoLev).Infof("%s batchNodeOrderFn Get task for TPU number:%d,node top:%v", PluginName, taskReqNPU, top)

	// 3.scored the nodes and set top
	scoreMp, errGet = scoreAndSetSelectNodes(task, nodes, scoreMp, bestNodeName, top)
	if errGet != nil {
		// get suitable node failed
		klog.V(logErrorLev).Infof("%s scoreAndSetSelectNodes get err:%v", PluginName, errGet)
		return scoreMp, nil
	}

	klog.V(logInfoLev).Infof("%s Total Score for task %s/%s is: %v", PluginName, task.Namespace, task.Name, scoreMp)
	return scoreMp, nil
}

func getTopIntFromAnnotations(Annotations map[string]string) []int {
	tmpTopStr, ok := Annotations[npuResourceName]
	if !ok {
		klog.V(logErrorLev).Infof("%s getTopIntFromAnnotations top nil", PluginName)
		return nil
	}

	tmpTopInt := getTopToIntArray(tmpTopStr)
	if tmpTopInt == nil {
		klog.V(logErrorLev).Infof("%s getTopIntFromAnnotations to int failed", PluginName)
		return nil
	}

	return tmpTopInt
}

// reduce nodeTop from taskTop
func getRealTopAfterAlloc(nodeTopInt []int, taskTopInt []int) string {
	var tmpTopInt []int
	var existFlag bool

	for _, nTopI := range nodeTopInt {
		existFlag = false
		for _, tTopI := range taskTopInt {
			if nTopI == tTopI {
				existFlag = true
				break
			}
		}

		if !existFlag {
			tmpTopInt = append(tmpTopInt, nTopI)
		}
	}
	klog.V(logDebugLev).Infof("%s getRealTopAfterAlloc ：%v ", PluginName, tmpTopInt)
	// change int to string
	return changeIntArrToStr(tmpTopInt)
}

// set node top by new one
func reloadNewTopToNodeOther(node *api.NodeInfo, newNodeTopStr string) error {

	node.Others[npuResourceName] = newNodeTopStr

	return nil
}

func getNodeNpuStrFromOther(mapInter map[string]interface{}) string {
	valueTmp, ok := mapInter[npuResourceName]
	if !ok {
		klog.V(logErrorLev).Infof("%s getNodeNpuStrFromOther nil", PluginName)
		return ""
	}

	mapStr, ok := valueTmp.(string)
	if !ok {
		klog.V(logErrorLev).Infof("%s getNodeNpuStrFromOther no string type", PluginName)
		return ""
	}

	return mapStr
}

func getTopIntFromNodeOther(mapInter map[string]interface{}) []int {

	mapStr := getNodeNpuStrFromOther(mapInter)
	if mapStr == "" {
		klog.V(logErrorLev).Infof("%s getNodeNpuStrFromOther other nil", PluginName)
		return nil
	}

	tmpTopInt := getTopToIntArray(mapStr)
	if tmpTopInt == nil {
		klog.V(logErrorLev).Infof("%s getTopIntFromAnnotations to int failed", PluginName)
		return nil
	}

	return tmpTopInt
}

func useAnnotation(node *api.NodeInfo, task *api.TaskInfo) {
	// get task use top
	taskTopInt := getTopIntFromAnnotations(task.Pod.Annotations)
	if taskTopInt == nil {
		klog.V(logErrorLev).Infof("%s getTopIntFromAnnotations failed task:%s", PluginName, task.Name)
		return
	}
	// get node available top
	nodeTopInt := getTopIntFromNodeOther(node.Others)
	if nodeTopInt == nil {
		klog.V(logErrorLev).Infof("%s useAnnotation node(%s) top nil", PluginName, node.Name)
		return
	}

	// delete the use top
	klog.V(logInfoLev).Infof("%s useAnnotation top nod：%v , task: %v", PluginName, nodeTopInt, taskTopInt)
	newNodeTopStr := getRealTopAfterAlloc(nodeTopInt, taskTopInt)
	if newNodeTopStr == "" {
		klog.V(logDebugLev).Infof("%s getRealTopAfterAlloc all top has allocated ", PluginName)
	}

	klog.V(logInfoLev).Infof("%s useAnnotation top(%s) to node[%s] successes", PluginName, newNodeTopStr, node.Name)
	err := reloadNewTopToNodeOther(node, newNodeTopStr)
	if err != nil {
		klog.V(logErrorLev).Infof("%s reloadNewTopToNode failed", PluginName)
		return
	}
	return
}

func getRealTopAfterRelease(nodeTopInt []int, taskTopInt []int) string {
	var tmpTopInt []int
	tmpTopMap := make(map[int]int, npuNumbPerNode)
	// add node topology to tmp map
	for _, nTopI := range nodeTopInt {
		tmpTopMap[nTopI] = 0
	}
	// add task topology to tmp map, Deduplicate the same topology
	for _, tTopI := range taskTopInt {
		if _, ok := tmpTopMap[tTopI]; ok {
			klog.V(logInfoLev).Infof("%s getRealTopAfterRelease already has cardId: %d", PluginName, tTopI)
			continue
		}
		tmpTopMap[tTopI] = 0
	}
	// change tmp map to slice
	for k := range tmpTopMap {
		tmpTopInt = append(tmpTopInt, k)
	}
	// change int to string
	return changeIntArrToStr(tmpTopInt)
}

func releaseAnnotation(node *api.NodeInfo, task *api.TaskInfo) {
	// get task use top
	taskTopInt := getTopIntFromAnnotations(task.Pod.Annotations)
	if taskTopInt == nil {
		klog.V(logErrorLev).Infof("%s getTopIntFromAnnotations failed task:%s", PluginName, task.Name)
		return
	}
	// get node available top
	nodeTopInt := getTopIntFromNodeOther(node.Others)
	if nodeTopInt == nil {
		klog.V(logErrorLev).Infof("%s useAnnotation node(%s) top nil", PluginName, node.Name)
		return
	}
	// delete the use top
	newNodeTopStr := getRealTopAfterRelease(nodeTopInt, taskTopInt)
	if newNodeTopStr == "" {
		klog.V(logErrorLev).Infof("%s getRealTopAfterRelease top failed", PluginName)
		return
	}

	err := reloadNewTopToNodeOther(node, newNodeTopStr)
	if err != nil {
		klog.V(logErrorLev).Infof("%s reloadNewTopToNode failed", PluginName)
		return
	}

	klog.V(logInfoLev).Infof("%s useAnnotation top(%s) to node(%s) successes", PluginName, newNodeTopStr, node.Name)
	return
}

// update node annotation of npu
func npuAllocateFunc(event *framework.Event, nodeMap map[string]*api.NodeInfo) {
	nodeName := event.Task.NodeName
	node, found := nodeMap[nodeName]
	if !found {
		klog.Warningf("%s npuAllocateFunc NOT EXIST node [%s]", PluginName, nodeName)
	} else {
		useAnnotation(node, event.Task)
		klog.V(logDebugLev).Infof("%s useAnnotation node [%s]'s top", PluginName, nodeName)
	}
}

// release node annotation of npu
func npuDeallocateFunc(event *framework.Event, nodeMap map[string]*api.NodeInfo) {
	nodeName := event.Task.NodeName
	node, found := nodeMap[nodeName]
	if !found {
		klog.Warningf("%s npuDeallocateFunc from NOT EXIST node [%s]", PluginName, nodeName)
	} else {
		releaseAnnotation(node, event.Task)
		klog.V(logDebugLev).Infof("%s releaseAnnotation node [%s]'s top", PluginName, nodeName)
	}
}

func setNodeAnnotations(Annotations map[string]interface{}, srcStr string) {
	Annotations[npuResourceName] = srcStr
}

func initNodeOther(Annotations map[string]interface{}) map[string]interface{} {
	Annotations = make(map[string]interface{})
	return Annotations
}

func initNodeNpuUseState(ssn *framework.Session) {
	nodeTopInt := []int{0, 1, 2, 3, 4, 5, 6, 7}

	for _, node := range ssn.Nodes {
		node.Others = initNodeOther(node.Others)
		setNodeAnnotations(node.Others, changeIntArrToStr(nodeTopInt))
	}
}

func (pp *topology910plugin) OnSessionOpen(ssn *framework.Session) {
	// init nodes npu top, for concurrency-oriented Allocation
	initNodeNpuUseState(ssn)

	// check the job require npu number,if illegal return filed
	ssn.AddJobValidFn(pp.Name(), validJobFn)

	// if npu no meet the task require,the task will failed.so need to intercept in advance
	ssn.AddPredicateFn(pp.Name(), npuPredicate)

	// The job who has below or equal 8 NPU,only has one pod. If over, every pod has 8s NPU.
	ssn.AddBatchNodeOrderFn(pp.Name(), batchNodeOrderFn)

	// Register event handlers to update task info in PodLister & nodeMap
	// for support Concurrency
	ssn.AddEventHandler(&framework.EventHandler{
		AllocateFunc: func(event *framework.Event) {
			npuAllocateFunc(event, ssn.Nodes)
		},
		DeallocateFunc: func(event *framework.Event) {
			npuDeallocateFunc(event, ssn.Nodes)
		},
	})
}

func (pp *topology910plugin) OnSessionClose(ssn *framework.Session) {
	// for recording job's unscheduled reason; and update job status
}
