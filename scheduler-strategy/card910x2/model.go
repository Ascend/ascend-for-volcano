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

Package card910x2 is using for HuaWei A300T Ascend pin affinity schedule.

*/
package card910x2

import (
	"errors"
	"fmt"
	"k8s.io/klog"
	"strconv"
	"volcano.sh/volcano/pkg/scheduler/api"
	hwutil "volcano.sh/volcano/pkg/scheduler/plugins/huaweinpu/scheduler-strategy/util"
)

func judgeNodeAndTaskNPU(taskNPU int, nodeNPUTopology []int) error {
	var meetErr = fmt.Errorf("req npu(%d) illegal", taskNPU)
	var reFlag = false

	// record the npu card number of HCCS rings
	leftCardNum, rightCardNum := hwutil.GetNodeHccsCardNum(nodeNPUTopology)

	switch taskNPU {
	case 0:
		return nil
	case 1:
		reFlag = (leftCardNum > 0) || (rightCardNum > 0)
	case constIntNum2:
		reFlag = len(nodeNPUTopology) == constIntNum2
	default:
		// single pod(task) cannot require npu not belong to mode
		// this kind job has been deal with job logical
		klog.V(logErrorLev).Infof("judgeNodeAndTaskNPU %s : %v.", PluginName, meetErr)
	}

	if reFlag {
		return nil
	}

	klog.V(logErrorLev).Infof("%s %v.", PluginName, meetErr)
	return meetErr
}

// Initializes the node priority series group according to the priority scheduling policy of 1 card.
func initOneCardPriNodeGroups(
	cardIds []int,
	priNodeGroups []map[string]*npuPriNodeInf,
	addPriNodeGroupFn initPriNodeGroupFn) error {

	if len(cardIds) == 1 {
		// A group
		addPriNodeGroupFn(priNodeGroups[0], "A")
		return nil
	}

	if len(cardIds) == constIntNum2 {
		// C group
		if len(priNodeGroups) > constIntNum2 {
			addPriNodeGroupFn(priNodeGroups[1], "B")
			return nil
		}

		err := fmt.Errorf("priNodeGroups's length(%d) not enough(%d)", len(priNodeGroups), constIntNum2)
		klog.V(logErrorLev).Infof("%s initOneCardPriNodeGroups %v.", PluginName, err.Error())
		return err
	}

	// no satisfy HCCS,8*N need whole nodes
	klog.V(logErrorLev).Infof("%s initOneCardPriNodeGroups node(%v) cannot fit.", PluginName, cardIds)
	return errors.New(nodeNoFitNPUWarning)
}

// Initializes the node priority series group according to the priority scheduling policy of 2 cards.
func initTwoCardPriNodeGroups(
	cardIds []int,
	priNodeGroups []map[string]*npuPriNodeInf,
	addPriNodeGroupFn initPriNodeGroupFn) error {

	if len(cardIds) == constIntNum2 {
		// C group
		addPriNodeGroupFn(priNodeGroups[0], "A")
		return nil
	}

	// no satisfy HCCS,8*N need whole nodes
	klog.V(logErrorLev).Infof("%s initOneCardPriNodeGroups node(%v) cannot fit.", PluginName, cardIds)
	return errors.New(nodeNoFitNPUWarning)
}

// Place nodes in the priority group.
func insertNodeInPriGroup(
	task *api.TaskInfo,
	cardIds []int,
	priNodeGroups []map[string]*npuPriNodeInf,
	addPriNodeGroupFn initPriNodeGroupFn) error {

	var err error

	// 1.Get the task's NPU request
	taskReqNPU, errGet := hwutil.GetTaskNPUNum(task, a300TNPUCardName)
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
		err = initOneCardPriNodeGroups(cardIds, priNodeGroups, addPriNodeGroupFn)
	case constIntNum2:
		err = initTwoCardPriNodeGroups(cardIds, priNodeGroups, addPriNodeGroupFn)
	default:
		// For normal,can not be here. The pre function validate job has done this.
		klog.V(logErrorLev).Infof("%s node(%v) not fit task request %d.", PluginName, cardIds, taskReqNPU)
		err = errors.New("illegal request npu number " + strconv.Itoa(taskReqNPU))
	}

	return err
}

func getNPUAllocPriorityArray(taskNPUNumber int) ([npuNumPerHccs]int, error) {
	var priorityArray [npuNumPerHccs]int
	var err = error(nil)

	switch taskNPUNumber {
	case 0:
		klog.V(logInfoLev).Infof("%s task req npu is 0.", PluginName)
	case 1:
		// priority:1>2
		priorityArray = [npuNumPerHccs]int{1, constIntNum2}
	case constIntNum2:
		priorityArray = [npuNumPerHccs]int{constIntNum2}
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

func getHccsFromNodeByPriority(nodeTop []int, priorityArray [npuNumPerHccs]int) ([]int, error) {
	npuNum := len(nodeTop)

	klog.V(logDebugLev).Infof("%s getHccsFromNodeByPriority nodeTop: %v.", PluginName, nodeTop)
	for _, npuNumber := range priorityArray {
		if npuNumber == 0 {
			continue
		}

		if npuNumber == npuNum {
			return nodeTop, nil
		}
	}

	err := errors.New("nodeTop not meet")
	klog.V(logErrorLev).Infof("%s getHccsFromNodeByPriority: %v %s.", PluginName, nodeTop, err.Error())
	return nil, err
}
