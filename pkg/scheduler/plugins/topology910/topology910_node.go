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
	"strings"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/conf"
	"volcano.sh/volcano/pkg/scheduler/framework"
)

func initNodesNpuAllocTopology(nodes map[string]*api.NodeInfo) {
	var nodeTop = []int{0, 1, 2, 3, 4, 5, 6, 7}

	for _, node := range nodes {
		node.Others = make(map[string]interface{}, magicNumInt1)
		saveTopologyInMap(node.Others, changeIntArrToStr(nodeTop))
	}
}

func isNpuNode(node *api.NodeInfo) error {
	_, ok := node.Node.Annotations[npu910CardName]
	if !ok {
		return errors.New("no npu node")
	}

	return nil
}

func getNodeSelector(node *api.NodeInfo) (map[string]string, error) {
	if err := isNpuNode(node); err != nil {
		// node no npu or k8s is not the match, cannot be selected
		klog.V(logInfoLev).Infof("%s node(%s) : %v", PluginName, node.Name, err)
		return node.Node.Labels, nil
	}

	_, ok := node.Node.Labels[archSelector]
	if !ok {
		return nil, errors.New("not Selector")
	}

	return node.Node.Labels, nil
}

func checkTaskAndNodeSelectorMeet(tSelectors map[string]string,
	nSelector map[string]string,
	conf map[string]string) error {

	for taskKey, taskValue := range tSelectors {
		confValue, confOk := conf[taskKey]
		if !confOk {
			return fmt.Errorf("conf has no task selector:%s", taskKey)
		}

		nodeValue, nodeOk := nSelector[taskKey]
		if !nodeOk {
			return fmt.Errorf("node has no task selector:%s", taskKey)
		}

		if !strings.Contains(confValue, taskValue) || !strings.EqualFold(taskValue, nodeValue) {
			return fmt.Errorf("selector(%s) not equal: task(%s) node(%s) conf(%s)",
				taskKey, taskValue, nodeValue, confValue)
		}
	}

	return nil
}

func isSelectorMeetNode(task *api.TaskInfo, node *api.NodeInfo, conf map[string]string) error {
	taskSelectors := getTaskSelectors(task)
	if taskSelectors == nil || len(taskSelectors) == 0 {
		klog.V(logDebugLev).Infof("task(%s) has no selectors", task.Name)
		return fmt.Errorf("task(%s) no selector", task.Name)
	}

	nodeSelector, errNode := getNodeSelector(node)
	if errNode != nil {
		klog.V(logErrorLev).Infof("%s node(%s) has no selector", PluginName, task.Name)
		return fmt.Errorf("node(%s) no selector", node.Name)
	}

	if err := checkTaskAndNodeSelectorMeet(taskSelectors, nodeSelector, conf); err != nil {
		klog.V(logErrorLev).Infof("%s isSelectorMeetNode,err:%v", PluginName, err)
		return err
	}

	return nil
}

// get top kind like（int[]） from node inf
func getTopFromNode(node *api.NodeInfo) []int {
	var topInt []int

	topStr, ok := node.Node.Annotations[npu910CardName]
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

func getNodeNpuNumFromAnnotation(nodeInfo *api.NodeInfo) (int, error) {
	top := getTopFromNode(nodeInfo)
	if top == nil {
		return 0, fmt.Errorf("nil node(%s) top", nodeInfo.Name)
	}
	nodeNpuIdleNumFromTop := len(top)

	return nodeNpuIdleNumFromTop, nil
}

func getNodeNpuNumFromIdle(nodeInfo *api.NodeInfo) (int, error) {
	nodeNpuIdleNumFromIdle, ok := nodeInfo.Idle.ScalarResources[npu910CardName]
	if !ok {
		klog.V(logErrorLev).Infof("%s getNodeNpuNumFromIdle failed", PluginName)
		return 0, errors.New("get node idle npu failed")
	}

	return int(nodeNpuIdleNumFromIdle / npuHex), nil
}

func checkNodeNpuStabilize(nodeNpuIdleNumFromTop int, nodeNpuIdleNumFromIdle int) error {
	if nodeNpuIdleNumFromTop != nodeNpuIdleNumFromIdle {
		return fmt.Errorf("node annotations(%d) not same node idle(%d)",
			nodeNpuIdleNumFromTop, nodeNpuIdleNumFromIdle)
	}

	return nil
}

func checkNpuResourceStable(task *api.TaskInfo, nodeInfo *api.NodeInfo) error {
	if err := isNpuTask(task); err != nil {
		// task no need npu
		klog.V(logInfoLev).Infof("%s isNpuTask %s : %v", PluginName, task.Name, err)
		return nil
	}

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
	klog.V(logInfoLev).Infof("%s Predicate get Top NPU req num:%v,left num:%d,right num:%d",
		PluginName, taskNpu, leftCardNum, rightCardNum)
	return leftCardNum, rightCardNum
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
	case nodeNpuNumber:
		reFlag = (leftCardNum + rightCardNum) == nodeNpuNumber
	default:
		// single pod(task) cannot require more than 8 npu
		klog.V(logErrorLev).Infof("%s Predicate jobs req more than 8 NPUs :%d", PluginName, taskNpu)
	}

	if reFlag {
		return nil
	}

	return errors.New("no meet")
}

func checkNodeNpuByTask(task *api.TaskInfo, node *api.NodeInfo) error {
	if err := isNpuTask(task); err != nil {
		// task no need npu
		klog.V(logInfoLev).Infof("%s isNpuTask %s : %v", PluginName, task.Name, err)
		return nil
	}

	taskNpu, taskError := getTaskNpuNum(task)
	if taskError != nil {
		return taskError
	}

	nodeNpuTopology := getTopFromNode(node)
	if nodeNpuTopology == nil {
		// node has none npu
		klog.V(logInfoLev).Infof("%s checkNodeNpuByTask nil,node name:%s,task req npu:%d",
			PluginName, node.Name, taskNpu)
		return errors.New("node no available npu")
	}
	klog.V(logInfoLev).Infof("%s checkNodeNpuByTask node(%s)top:%v", PluginName, node.Name, nodeNpuTopology)

	err := judgeNodeAndTaskNpu(taskNpu, nodeNpuTopology)
	if err != nil {
		return err
	}

	return nil
}

func nodePredicate(task *api.TaskInfo, node *api.NodeInfo, conf []conf.Configuration) error {
	schedulerConf := getSchedulerSelectorConfig(conf)
	if schedulerConf == nil || len(schedulerConf) == 0 {
		// get scheduler selector configure failed, but need continue
		klog.V(logErrorLev).Infof("%s JobName: %s get selector nil", PluginName, task.Name)
		return errors.New("get scheduler selector nil")
	}

	// select node by architect
	if err := isSelectorMeetNode(task, node, schedulerConf); err != nil {
		// get scheduler selector configure failed, but need continue
		klog.V(logErrorLev).Infof("%s taskName: %s ,nodeName %s : %v", PluginName, task.Name, node.Name, err)
		return err
	}

	// check resource stabilize
	if err := checkNpuResourceStable(task, node); err != nil {
		// npu not be Stable by k8s,cannot select.
		klog.V(logInfoLev).Infof("%s checkNpuResourceStable %s : %v ,cannot be selected.", PluginName,
			node.Name, err)
		return err
	}

	if err := checkNodeNpuByTask(task, node); err != nil {
		// npu not be Stable by k8s,cannot select.
		klog.V(logInfoLev).Infof("%s checkNodeNpuByTask :%v ,cannot be selected.", PluginName, err)
		return err
	}

	return nil
}

func getTopIntFromAnnotations(Annotations map[string]string) []int {
	tmpTopStr, ok := Annotations[npu910CardName]
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

func getNodeNpuStrFromOther(mapInter map[string]interface{}) string {
	valueTmp, ok := mapInter[npu910CardName]
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

	node.Others[npu910CardName] = newNodeTopStr

	return nil
}

func useAnnotation(node *api.NodeInfo, task *api.TaskInfo) {
	// get task use top
	taskTopInt := getTopIntFromAnnotations(task.Pod.Annotations)
	if taskTopInt == nil {
		klog.V(logErrorLev).Infof("%s useAnnotation failed task:%s", PluginName, task.Name)
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

	klog.V(logInfoLev).Infof("%s useAnnotation top(%s) to node[%s] successes",
		PluginName, newNodeTopStr, node.Name)
	err := reloadNewTopToNodeOther(node, newNodeTopStr)
	if err != nil {
		klog.V(logErrorLev).Infof("%s reloadNewTopToNode failed", PluginName)
		return
	}
	return
}

// update node annotation of npu
func npuAllocateFunc(event *framework.Event, nodeMap map[string]*api.NodeInfo) {
	nodeName := event.Task.NodeName
	node, found := nodeMap[nodeName]
	if !found {
		klog.V(logWarningLev).Infof("%s npuAllocateFunc NOT EXIST node [%s]", PluginName, nodeName)
	} else {
		useAnnotation(node, event.Task)
		klog.V(logDebugLev).Infof("%s useAnnotation node [%s]'s top", PluginName, nodeName)
	}
}

func getRealTopAfterRelease(nodeTopInt []int, taskTopInt []int) string {
	var tmpTopInt []int
	tmpTopMap := make(map[int]int, nodeNpuNumber)
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
		klog.V(logErrorLev).Infof("%s releaseAnnotation failed task:%s", PluginName, task.Name)
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

	klog.V(logInfoLev).Infof("%s useAnnotation node(%s) top(%s) successes", PluginName, node.Name, newNodeTopStr)
	return
}

// release node annotation of npu
func npuDeallocateFunc(event *framework.Event, nodeMap map[string]*api.NodeInfo) {
	nodeName := event.Task.NodeName
	node, found := nodeMap[nodeName]
	if !found {
		klog.V(logWarningLev).Infof("%s npuDeallocateFunc from NOT EXIST node [%s]", PluginName, nodeName)
	} else {
		releaseAnnotation(node, event.Task)
		klog.V(logDebugLev).Infof("%s releaseAnnotation node [%s]'s top", PluginName, nodeName)
	}
}
