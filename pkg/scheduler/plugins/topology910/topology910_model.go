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
	"k8s.io/klog"
	schedulerApi "k8s.io/kube-scheduler/extender/v1"
	"strconv"
	"volcano.sh/volcano/pkg/scheduler/api"
)

const (
	leftHccsName        = "leftHccs"
	rightHccsName       = "rightHccs"
	allHccsName         = "allHccs"
	nodeNoFitNPUWarning = "node no fit npu number"
)

type hccsInf struct {
	nodeInf  *api.NodeInfo
	name     string
	availNum int
	top      []int
}

type npuPriNodeInf struct {
	// the priority for NPU HCCS top
	Name        string
	allocateNum int
	idleNpuNum  int
	nodeName    string
	selHccsName string
	leftHccs    hccsInf
	rightHccs   hccsInf
}

type selectNodeInf struct {
	nodeName    string
	allNpuNum   int
	leftNpuNum  int
	rightNpuNum int
}

type initPriNodeGroupFn func(priNodeGroup map[string]*npuPriNodeInf, groupName string, selectCards int)

func initScoreMp(nodes []*api.NodeInfo, interPodAffinityScore schedulerApi.HostPriorityList) map[string]float64 {
	for _, node := range nodes {
		interPodAffinityScore = append(interPodAffinityScore, schedulerApi.HostPriority{
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

// choose the right hccs ring in node
func selectDireTopFn(selectCardNum int, sNodeInf selectNodeInf) string {
	if sNodeInf.leftNpuNum == selectCardNum {
		return leftHccsName
	} else if sNodeInf.rightNpuNum == selectCardNum {
		return rightHccsName
	} else if sNodeInf.allNpuNum == selectCardNum {
		return allHccsName
	} else {
		return ""
	}
}

// the police for one need card
func initOneCardPriNodeGroups(
	sNodeInf selectNodeInf,
	priNodeGroups []map[string]*npuPriNodeInf,
	addPriNodeGroupFn initPriNodeGroupFn) error {
	// priority:1>3>2>4
	if sNodeInf.leftNpuNum == magicNumInt1 || sNodeInf.rightNpuNum == magicNumInt1 {
		// A group
		addPriNodeGroupFn(priNodeGroups[0], "A", magicNumInt1)
		return nil
	}

	if sNodeInf.leftNpuNum == magicNumInt3 || sNodeInf.rightNpuNum == magicNumInt3 {
		// B group
		addPriNodeGroupFn(priNodeGroups[magicNumInt1], "B", magicNumInt3)
		return nil
	}

	if sNodeInf.leftNpuNum == magicNumInt2 || sNodeInf.rightNpuNum == magicNumInt2 {
		// C group
		addPriNodeGroupFn(priNodeGroups[magicNumInt2], "C", magicNumInt2)
		return nil
	}

	if sNodeInf.leftNpuNum == npuNumPerHccs || sNodeInf.rightNpuNum == npuNumPerHccs {
		// D group
		addPriNodeGroupFn(priNodeGroups[magicNumInt3], "D", npuNumPerHccs)
		return nil
	}

	// no satisfy HCCS,8*N need whole nodes
	klog.V(logErrorLev).Infof("%s initOneCardPriNodeGroups node(%s) (left :%d,right :%d) cannot fit",
		PluginName, sNodeInf.nodeName, sNodeInf.leftNpuNum, sNodeInf.rightNpuNum)
	return errors.New(nodeNoFitNPUWarning)
}

// the police for Two need card
func initTwoCardPriNodeGroups(
	sNodeInf selectNodeInf,
	priNodeGroups []map[string]*npuPriNodeInf,
	addPriNodeGroupFn initPriNodeGroupFn) error {

	// priority：2>npuNumPerHccs>3
	if sNodeInf.leftNpuNum == magicNumInt2 || sNodeInf.rightNpuNum == magicNumInt2 {
		// A group
		addPriNodeGroupFn(priNodeGroups[0], "A", magicNumInt2)
		return nil
	}

	if sNodeInf.leftNpuNum == npuNumPerHccs || sNodeInf.rightNpuNum == npuNumPerHccs {
		// B group
		addPriNodeGroupFn(priNodeGroups[magicNumInt1], "B", npuNumPerHccs)
		return nil
	}

	if sNodeInf.leftNpuNum == magicNumInt3 || sNodeInf.rightNpuNum == magicNumInt3 {
		// C group
		addPriNodeGroupFn(priNodeGroups[magicNumInt2], "C", magicNumInt3)
		return nil
	}

	// no satisfy HCCS,8*N need whole nodes
	klog.V(logErrorLev).Infof("%s initTwoCardPriNodeGroups node(%s) (left :%d,right :%d) cannot fit 2",
		PluginName, sNodeInf.nodeName, sNodeInf.leftNpuNum, sNodeInf.rightNpuNum)
	return errors.New(nodeNoFitNPUWarning)
}

// the police four one need card
func initFourCardPriNodeGroups(
	sNodeInf selectNodeInf,
	priNodeGroups []map[string]*npuPriNodeInf,
	addPriNodeGroupFn initPriNodeGroupFn) error {
	// only can allocate 4
	if sNodeInf.leftNpuNum == npuNumPerHccs || sNodeInf.rightNpuNum == npuNumPerHccs {
		// A group
		addPriNodeGroupFn(priNodeGroups[0], "A", npuNumPerHccs)
		return nil
	}
	// no satisfy HCCS,8*N need whole nodes
	klog.V(logErrorLev).Infof("%s initFoureCardPriNodeGroups node(%s) (left :%d,right :%d) cannot fit 4",
		PluginName, sNodeInf.nodeName, sNodeInf.leftNpuNum, sNodeInf.rightNpuNum)
	return errors.New(nodeNoFitNPUWarning)
}

// the police four one need card
func initEightCardPriNodeGroups(
	sNodeInf selectNodeInf,
	priNodeGroups []map[string]*npuPriNodeInf,
	addPriNodeGroupFn initPriNodeGroupFn) error {

	if sNodeInf.allNpuNum == nodeNpuNumber {
		addPriNodeGroupFn(priNodeGroups[0], "A", nodeNpuNumber)
		return nil
	}

	klog.V(logErrorLev).Infof("%s initEightCardPriNodeGroups node(%s) (all:%d) cannot fit 8",
		PluginName, sNodeInf.nodeName, sNodeInf.allNpuNum)
	return errors.New(nodeNoFitNPUWarning)
}

// insert into group by police
func insertNodeInPriGroup(
	taskReqNPU int,
	sNodeInf selectNodeInf,
	priNodeGroups []map[string]*npuPriNodeInf,
	addPriNodeGroupFn initPriNodeGroupFn) error {

	var err error
	switch taskReqNPU {
	case 0:
		klog.V(logInfoLev).Infof("%s initPriNodeGroups task npu is 0", PluginName)
	case magicNumInt1:
		err = initOneCardPriNodeGroups(sNodeInf, priNodeGroups, addPriNodeGroupFn)
	case magicNumInt2:
		err = initTwoCardPriNodeGroups(sNodeInf, priNodeGroups, addPriNodeGroupFn)
	case npuNumPerHccs:
		err = initFourCardPriNodeGroups(sNodeInf, priNodeGroups, addPriNodeGroupFn)
	case nodeNpuNumber:
		err = initEightCardPriNodeGroups(sNodeInf, priNodeGroups, addPriNodeGroupFn)
	default:
		// For normal,can not be here. The pre function validate job has done this.
		klog.V(logErrorLev).Infof("%s node(%s) (left :%d,right :%d) cannot fit %d,illegal task npu number",
			PluginName, sNodeInf.nodeName, sNodeInf.leftNpuNum, sNodeInf.rightNpuNum, taskReqNPU)
		err = errors.New("illegal request npu number " + strconv.Itoa(taskReqNPU))
	}

	return err
}

// init 4 pri node-list group
func initPriNodeGroups(taskReqNPU int, nodes []*api.NodeInfo) ([]map[string]*npuPriNodeInf, error) {
	var err error
	var priNodeGroups []map[string]*npuPriNodeInf

	// for pipelined state the node npu is nil
	if len(nodes) == 0 {
		return nil, errors.New("nodes is empty")
	}

	for i := 0; i < npuNumPerHccs; i++ {
		priNodeGroups = append(priNodeGroups, make(map[string]*npuPriNodeInf, 1))
	}

	// init pri Node group
	for _, node := range nodes {
		var leftHccsTop []int
		var rightHccsTop []int
		var sNodeInf selectNodeInf

		cardIds := getTopFromNode(node)
		klog.V(logDebugLev).Infof("%s initPriNodeGroups:%v", PluginName, cardIds)
		for _, cardID := range cardIds {
			sNodeInf.allNpuNum++

			if cardID < npuNumPerHccs {
				sNodeInf.leftNpuNum++
				leftHccsTop = append(leftHccsTop, cardID)
			} else {
				sNodeInf.rightNpuNum++
				rightHccsTop = append(rightHccsTop, cardID)
			}
		}
		// set the meet node into its pri-node-list group
		addPriNodeGroupFn := func(priNodeGroup map[string]*npuPriNodeInf, groupName string, selectCards int) {
			klog.V(logDebugLev).Infof("%s nodeName:%s,group:%v", PluginName, node.Name, priNodeGroup[node.Name])
			priNodeGroup[node.Name] = &npuPriNodeInf{
				Name:        groupName,
				idleNpuNum:  sNodeInf.allNpuNum,
				allocateNum: int(node.Allocatable.ScalarResources[npu910CardName]) / npuHex,
				selHccsName: selectDireTopFn(selectCards, sNodeInf),
				leftHccs:    hccsInf{name: leftHccsName, availNum: sNodeInf.leftNpuNum, top: leftHccsTop, nodeInf: node},
				rightHccs:   hccsInf{name: rightHccsName, availNum: sNodeInf.rightNpuNum, top: rightHccsTop, nodeInf: node},
				nodeName:    node.Name,
			}
			klog.V(logDebugLev).Infof("%s addPriNodeGroupFn node name:%s priNode:%v",
				PluginName, node.Name, priNodeGroup[node.Name])
		}

		// insert into group by police
		err = insertNodeInPriGroup(taskReqNPU, sNodeInf, priNodeGroups, addPriNodeGroupFn)
		if err != nil {
			continue
		}
	}
	return priNodeGroups, nil
}

func getBestPriNodeGroup(priNodeGroups []map[string]*npuPriNodeInf) (map[string]*npuPriNodeInf, int, error) {
	var i int
	var selectPriGroup map[string]*npuPriNodeInf
	capacity := nodeNpuNumber
	loopFlg := true

	findNodeInPriGroupFn := func() bool {
		for _, nodeNpu := range priNodeGroups[i] {
			// if capacity suitable,the node is meet the required.
			if nodeNpu.allocateNum == capacity {
				// right one
				selectPriGroup = priNodeGroups[i]
				loopFlg = false
				break
			}
		}
		return loopFlg
	}
	// loop capacity
	for (capacity != 0) && loopFlg {
		// loop 4 pri-node-list-group
		i = 0
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

	capacity = nodeNpuNumber
	loopFlg = true
	tmpPriNodeInf = &npuPriNodeInf{
		idleNpuNum: nodeNpuNumber + 1,
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
	for (capacity != 0) && loopFlg {
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

	if (needCards == nodeNpuNumber) && !loopFlg {
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
	i := 0
	// select whole node
	if needCards == nodeNpuNumber {
		for i = 0; i < nodeNpuNumber; i++ {
			setTop = append(setTop, i)
		}
		return setTop, nil
	}

	i = 0
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
	// 1. inti 4 pri node-list group
	priNodeGroups, err := initPriNodeGroups(taskReqNPU, nodes)
	if err != nil {
		klog.V(logErrorLev).Infof("%s initPriNodeGroups failed :%s", PluginName, err)
		return "", nil, err
	}
	// 2.get the right node and its hccs
	selectNodeName, selectHCCS, err := getBestNodeAndTop(taskReqNPU, priNodeGroups)
	if err != nil {
		klog.V(logErrorLev).Infof("%s getBestNodeAndTop failed :%s", PluginName, err)
		return "", nil, err
	}
	// 3.set the selected top,selectHCCS.top
	top, err = setSelectTopValue(taskReqNPU, selectHCCS)
	if err != nil {
		klog.V(logErrorLev).Infof("%s setSelectTopValue failed :%s", PluginName, err)
		return "", nil, err
	}

	klog.V(logInfoLev).Infof("%s setSelectTopValue top :%v", PluginName, top)
	return selectNodeName, top, nil
}

func checkResourceNumMeet(task *api.TaskInfo, nodeInfo *api.NodeInfo) error {
	if !task.InitResreq.LessEqual(nodeInfo.Idle) {
		return errors.New("IN Pipeline")
	}
	return nil
}

func checkResourceReady(task *api.TaskInfo, nodeInfo *api.NodeInfo) error {
	// check resource number fit
	err := checkResourceNumMeet(task, nodeInfo)
	if err != nil {
		return err
	}

	return nil
}

// write the topology to pod
// to change the top kind :int[] to device-plugin point kind:string link with ','
func setNpuTopologyToPod(task *api.TaskInfo, nodeInfo *api.NodeInfo, top []int) error {
	err := checkResourceReady(task, nodeInfo)
	if err != nil {
		return err
	}

	return doSetPodNpuTopology(top, task)
}

// select the best node for task using
func scoreAndSetSelectNodes(task *api.TaskInfo,
	nodes []*api.NodeInfo,
	scoreMp map[string]float64,
	bestNodeName string,
	top []int) (map[string]float64, error) {
	var nodeWeight float64
	nodeWeight = 1.0

	for _, nodeInfo := range nodes {
		if nodeInfo.Name != bestNodeName {
			scoreMp[nodeInfo.Name] = 0.0
			continue
		}

		// write the topology to task's pod
		err := setNpuTopologyToPod(task, nodeInfo, top)
		if err != nil {
			return nil, err
		}

		scoreMp[nodeInfo.Name] = nodeWeight
		break
	}

	return scoreMp, nil
}

// open-session register for batch Node Order fun
// for volcano frame,can not return error
func batchNodeOrderFn(task *api.TaskInfo, nodes []*api.NodeInfo) (map[string]float64, error) {
	var interPodAffinityScore schedulerApi.HostPriorityList
	var bestNodeName string
	var top []int
	var errGet error

	klog.V(logInfoLev).Infof("%s Enter batchNodeOrderFn", PluginName)
	defer klog.V(logInfoLev).Infof("%s leaving batchNodeOrderFn", PluginName)

	// init score-map
	scoreMp := initScoreMp(nodes, interPodAffinityScore)

	// 1.Get the task's NPU request
	taskReqNPU, errGet := getTaskNpuNum(task)
	if errGet != nil {
		// cannot return error for task is no npu kind possible.
		klog.V(logErrorLev).Infof("%s batchNodeOrderFn task :%s,%v", PluginName, task.Name, errGet)
		// cannot return nil，will panic
		return scoreMp, nil
	}
	klog.V(logInfoLev).Infof("%s batchNodeOrderFn Get %s taskReqNPU:%v", PluginName, task.Name, taskReqNPU)

	// 2.get the best node and top by A,B,C,D rules and require numbers.
	bestNodeName, top, errGet = getNpuAffinityBestNodeAndTop(taskReqNPU, nodes)
	if errGet != nil {
		// get suitable node failed
		klog.V(logErrorLev).Infof("%s batchNodeOrderFn task[%s],require npu:%d, failed[%v]",
			PluginName, task.Name, taskReqNPU, errGet)
		return scoreMp, nil
	}
	klog.V(logInfoLev).Infof("%s batchNodeOrderFn Get task for NPU number:%d, %s[%v]",
		PluginName, taskReqNPU, bestNodeName, top)

	// 3.scored the nodes and set top
	scoreMp, errGet = scoreAndSetSelectNodes(task, nodes, scoreMp, bestNodeName, top)
	if errGet != nil {
		// get suitable node failed
		klog.V(logErrorLev).Infof("%s scoreAndSetSelectNodes get err:%v", PluginName, errGet)
		return scoreMp, errGet
	}

	klog.V(logInfoLev).Infof("%s Total Score for task %s/%s is: %v", PluginName,
		task.Namespace, task.Name, scoreMp)
	return scoreMp, nil
}
