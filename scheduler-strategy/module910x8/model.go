/*
Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.

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

Package module910x8 is using for HuaWei A800/9000 Ascend910 pin affinity schedule.

*/
package module910x8

import (
	"errors"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog"
	"strconv"
	"volcano.sh/volcano/pkg/apis/scheduling"
	"volcano.sh/volcano/pkg/scheduler/api"
	hwutil "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/util"
)

func getHccsFromNodeByPriority(nodeTop []int, priorityArray [npuNumPerHccs]int) ([]int, error) {
	leftHccsArray, rightHccsArray := getNodeHccsArray(nodeTop)
	leftHccsNPUNum := len(leftHccsArray)
	rightHccsNPUNum := len(rightHccsArray)

	klog.V(logDebugLev).Infof("%s getHccsFromNodeByPriority: %v-%v.", PluginName, leftHccsArray, rightHccsArray)
	for _, npuNumber := range priorityArray {
		if npuNumber == 0 {
			continue
		}

		if npuNumber == nodeNPUNumber {
			if len(nodeTop) == nodeNPUNumber {
				return nodeTop, nil
			}
			break
		}

		if leftHccsNPUNum == npuNumber {
			klog.V(logDebugLev).Infof("%s get %v.", PluginName, leftHccsArray)
			return leftHccsArray, nil
		}

		if rightHccsNPUNum == npuNumber {
			klog.V(logDebugLev).Infof("%s get %v.", PluginName, rightHccsArray)
			return rightHccsArray, nil
		}
	}

	err := errors.New("nodeTop not meet")
	klog.V(logErrorLev).Infof("%s getHccsFromNodeByPriority: %v-%v %s.",
		PluginName, leftHccsArray, rightHccsArray, err.Error())
	return nil, err
}

func getNPUAllocPriorityArray(taskNPUNumber int) ([npuNumPerHccs]int, error) {
	var priorityArray [npuNumPerHccs]int
	var err = error(nil)

	switch taskNPUNumber {
	case 0:
		klog.V(logInfoLev).Infof("%s task req npu is 0.", PluginName)
	case 1:
		// priority:1>3>2>4
		priorityArray = [npuNumPerHccs]int{1, constIntNum3, constIntNum2, npuNumPerHccs}
	case constIntNum2:
		// priority：2>npuNumPerHccs>3
		priorityArray = [npuNumPerHccs]int{constIntNum2, npuNumPerHccs, constIntNum3}
	case npuNumPerHccs:
		// priority：4
		priorityArray = [npuNumPerHccs]int{npuNumPerHccs}
	case nodeNPUNumber:
		priorityArray = [npuNumPerHccs]int{nodeNPUNumber}
	default:
		// For normal,can not be here. The pre function validate job has done this.
		err = fmt.Errorf("illegal request npu number: %d", taskNPUNumber)
	}

	if err != nil {
		klog.V(logErrorLev).Infof("%s %s.", PluginName, err.Error())
		return priorityArray, err
	}

	return priorityArray, nil
}

// Initializes the node priority series group according to the priority scheduling policy of 1 card.
func initOneCardPriNodeGroups(sNodeInf selectNodeInf,
	priNodeGroups []map[string]*npuPriNodeInf, addPriNodeGroupFn initPriNodeGroupFn) error {

	if len(priNodeGroups) < npuNumPerHccs {
		err := fmt.Errorf("priNodeGroups's length(%d) not meet(%d)", len(priNodeGroups), npuNumPerHccs)
		klog.V(logErrorLev).Infof("%s initOneCardPriNodeGroups %v.", PluginName, err.Error())
		return err
	}
	// priority:1>3>2>4
	if sNodeInf.leftNPUNum == 1 || sNodeInf.rightNPUNum == 1 {
		// A group
		addPriNodeGroupFn(priNodeGroups[0], "A")
		return nil
	}

	if sNodeInf.leftNPUNum == constIntNum3 || sNodeInf.rightNPUNum == constIntNum3 {
		// B group
		addPriNodeGroupFn(priNodeGroups[1], "B")
		return nil
	}

	if sNodeInf.leftNPUNum == constIntNum2 || sNodeInf.rightNPUNum == constIntNum2 {
		// C group
		addPriNodeGroupFn(priNodeGroups[constIntNum2], "C")
		return nil
	}

	if sNodeInf.leftNPUNum == npuNumPerHccs || sNodeInf.rightNPUNum == npuNumPerHccs {
		// D group
		addPriNodeGroupFn(priNodeGroups[constIntNum3], "D")
		return nil
	}

	// no satisfy HCCS,8*N need whole nodes
	klog.V(logErrorLev).Infof("%s initOneCardPriNodeGroups node(%s) (left :%d,right :%d) cannot fit.",
		PluginName, sNodeInf.nodeName, sNodeInf.leftNPUNum, sNodeInf.rightNPUNum)
	return errors.New(nodeNoFitNPUWarning)
}

// Place nodes in the priority group.
func insertNodeInPriGroup(
	task *api.TaskInfo,
	sNodeInf selectNodeInf,
	priNodeGroups []map[string]*npuPriNodeInf,
	addPriNodeGroupFn initPriNodeGroupFn) error {

	var err error

	// 1.Get the task's NPU request
	taskReqNPU, errGet := hwutil.GetTaskNPUNum(task, npu800And9000CardName)
	if errGet != nil {
		// cannot return error for task is no npu kind possible.
		klog.V(logErrorLev).Infof("%s batchNodeOrderFn task :%s,%v.", PluginName, task.Name, errGet)
		// cannot return nil，will panic
		return nil
	}

	switch taskReqNPU {
	case 0:
		klog.V(logInfoLev).Infof("%s initPriNodeGroups task npu is 0.", PluginName)
	case 1:
		err = initOneCardPriNodeGroups(sNodeInf, priNodeGroups, addPriNodeGroupFn)
	case constIntNum2:
		err = initTwoCardPriNodeGroups(sNodeInf, priNodeGroups, addPriNodeGroupFn)
	case npuNumPerHccs:
		err = initFourCardPriNodeGroups(sNodeInf, priNodeGroups, addPriNodeGroupFn)
	case nodeNPUNumber:
		err = initEightCardPriNodeGroups(sNodeInf, priNodeGroups, addPriNodeGroupFn)
	default:
		// For normal,can not be here. The pre function validate job has done this.
		klog.V(logErrorLev).Infof("%s node(%s) (left :%d,right :%d) cannot fit %d,illegal task npu number.",
			PluginName, sNodeInf.nodeName, sNodeInf.leftNPUNum, sNodeInf.rightNPUNum, taskReqNPU)
		err = errors.New("illegal request npu number " + strconv.Itoa(taskReqNPU))
	}

	return err
}

// Initializes the node priority series group according to the priority scheduling policy of 2 cards.
func initTwoCardPriNodeGroups(
	sNodeInf selectNodeInf,
	priNodeGroups []map[string]*npuPriNodeInf,
	addPriNodeGroupFn initPriNodeGroupFn) error {

	if len(priNodeGroups) < npuNumPerHccs {
		err := fmt.Errorf("priNodeGroups's length(%d) not enough(%d)", len(priNodeGroups), npuNumPerHccs)
		klog.V(logErrorLev).Infof("%s initTwoCardPriNodeGroups %v.", PluginName, err.Error())
		return err
	}
	// priority：2>npuNumPerHccs>3
	if sNodeInf.leftNPUNum == constIntNum2 || sNodeInf.rightNPUNum == constIntNum2 {
		// A group
		addPriNodeGroupFn(priNodeGroups[0], "A")
		return nil
	}

	if sNodeInf.leftNPUNum == npuNumPerHccs || sNodeInf.rightNPUNum == npuNumPerHccs {
		// B group
		addPriNodeGroupFn(priNodeGroups[1], "B")
		return nil
	}

	if sNodeInf.leftNPUNum == constIntNum3 || sNodeInf.rightNPUNum == constIntNum3 {
		// C group
		addPriNodeGroupFn(priNodeGroups[constIntNum2], "C")
		return nil
	}

	// no satisfy HCCS,8*N need whole nodes
	klog.V(logErrorLev).Infof("%s initTwoCardPriNodeGroups node(%s) (left :%d,right :%d) cannot fit 2.",
		PluginName, sNodeInf.nodeName, sNodeInf.leftNPUNum, sNodeInf.rightNPUNum)

	return errors.New(nodeNoFitNPUWarning)
}

// Initializes the node priority series group according to the priority scheduling policy of 4 cards.
func initFourCardPriNodeGroups(
	sNodeInf selectNodeInf,
	priNodeGroups []map[string]*npuPriNodeInf,
	addPriNodeGroupFn initPriNodeGroupFn) error {

	if len(priNodeGroups) < constIntNum2 {
		err := fmt.Errorf("priNodeGroups's length(%d) not meet(%d)", len(priNodeGroups), constIntNum2)
		klog.V(logErrorLev).Infof("%s initFourCardPriNodeGroups %v.", PluginName, err.Error())
		return err
	}
	// only can allocate 4
	if sNodeInf.leftNPUNum == npuNumPerHccs || sNodeInf.rightNPUNum == npuNumPerHccs {
		// A group
		addPriNodeGroupFn(priNodeGroups[0], "A")
		return nil
	}
	// no satisfy HCCS,8*N need whole nodes
	klog.V(logErrorLev).Infof("%s initFoureCardPriNodeGroups node(%s) (left :%d,right :%d) cannot fit 4.",
		PluginName, sNodeInf.nodeName, sNodeInf.leftNPUNum, sNodeInf.rightNPUNum)
	return errors.New(nodeNoFitNPUWarning)
}

// Initializes the node priority series group according to the priority scheduling policy of 8 cards.
func initEightCardPriNodeGroups(
	sNodeInf selectNodeInf,
	priNodeGroups []map[string]*npuPriNodeInf,
	addPriNodeGroupFn initPriNodeGroupFn) error {

	if len(priNodeGroups) < constIntNum2 {
		err := fmt.Errorf("priNodeGroups's length(%d) not enough(%d)", len(priNodeGroups), constIntNum2)
		klog.V(logErrorLev).Infof("%s initEightCardPriNodeGroups %v.", PluginName, err.Error())
		return err
	}

	if sNodeInf.allNPUNum == nodeNPUNumber {
		addPriNodeGroupFn(priNodeGroups[0], "A")
		return nil
	}

	klog.V(logErrorLev).Infof("%s initEightCardPriNodeGroups node(%s) (all:%d) cannot fit 8.",
		PluginName, sNodeInf.nodeName, sNodeInf.allNPUNum)
	return errors.New(nodeNoFitNPUWarning)
}

func judgeNodeAndTaskNPU(taskNPU int, nodeNPUTopology []int) error {
	var meetErr = fmt.Errorf("%v not meet req npu(%d)", nodeNPUTopology, taskNPU)
	var reFlag = false

	// record the npu card number of HCCS rings
	leftCardNum, rightCardNum := hwutil.GetNodeHccsCardNum(nodeNPUTopology)

	switch taskNPU {
	case 0:
		return nil
	case 1:
		reFlag = (leftCardNum > 0) || (rightCardNum > 0)
	case constIntNum2:
		reFlag = (leftCardNum > 1) || (rightCardNum > 1)
	case npuNumPerHccs:
		reFlag = (leftCardNum == npuNumPerHccs) || (rightCardNum == npuNumPerHccs)
	case nodeNPUNumber:
		reFlag = (leftCardNum + rightCardNum) == nodeNPUNumber
	default:
		// single pod(task) cannot require npu not belong to mode
		// this kind job has been deal with job logical
		klog.V(logErrorLev).Infof("judgeNodeAndTaskNPU %s : %v.", PluginName, meetErr)
	}

	if reFlag {
		return nil
	}

	klog.V(logErrorLev).Infof("%s %v not meet req %d.", PluginName, nodeNPUTopology, taskNPU)
	return meetErr
}

func isNodeInFaultJobUseList(task *api.TaskInfo, node *api.NodeInfo) bool {
	value, ok := hwutil.ReSchedulerJobs[task.Job]
	if !ok {
		return false
	}

	for _, nodeName := range value.NodeNames {
		if nodeName == node.Name {
			return true
		}
	}
	return false
}

// task
func checkFaultJobNode(task *api.TaskInfo, node *api.NodeInfo) error {
	if isNPUFaultTask(task) {
		klog.V(logErrorLev).Infof("%s %s is npu fault job.", PluginName, task.Job)
		return nil
	}

	if isNodeInFaultJobUseList(task, node) {
		msg := fmt.Errorf("%s is used by npu fault job:%s", node.Name, task.Job)
		klog.V(logErrorLev).Infof("%s %s.", PluginName, msg)
		return msg
	}

	return nil
}

func getJobUsedNodes(job *api.JobInfo) (map[string]*v1.Pod, error) {
	var nodeNames = make(map[string]*v1.Pod, constIntNum3)

	if job.PodGroup.Status.Phase != scheduling.PodGroupRunning {
		return nil, errors.New("not running job")
	}

	for _, task := range job.Tasks {
		nodeNames[task.NodeName] = task.Pod
	}
	klog.V(logDebugLev).Infof("%s getJobUsedNodes %s use %v.", PluginName, job.Name, nodeNames)
	return nodeNames, nil
}