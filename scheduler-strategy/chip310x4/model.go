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

Package chip310x4 is using for HuaWei A300T Ascend pin affinity schedule.

*/
package chip310x4

import (
	"errors"
	"fmt"
	"k8s.io/klog"
	"math/rand"
	"time"
)

func judgeNodeAndTaskNPU(taskNPU int, nodeNPUTopology []int) error {
	if len(nodeNPUTopology) >= taskNPU {
		return nil
	}

	var meetErr = fmt.Errorf("req npu(%d) illegal", taskNPU)
	klog.V(logErrorLev).Infof("%s %v.", PluginName, meetErr)
	return meetErr
}

func getNPUAllocPriorityArray(taskNPUNumber int) ([cardNPUNumber]int, error) {
	var priorityArray [cardNPUNumber]int
	var err = error(nil)
	// priority:1>2>3>4
	priorityArray = [cardNPUNumber]int{1, constIntNum2, constIntNum3, cardNPUNumber}
	if err != nil {
		klog.V(logErrorLev).Infof("%s %s.", PluginName, err.Error())
		return priorityArray, err
	}
	return priorityArray, nil
}

// getFitCardFromNodeByPriority get the appropriate card randomly from the Node via Priority
// for example:
// nodeTop: [0, 2, 4, 5, 7, 8, 9, 10, 11, 12, 13, 15, 16, 18, 20, 21, 22, 25, 29, 32, 33, 34, 35, 36, 37, 38, 39]
// cardNumGroups: [2, 3, 4, 3, 2, 3, 1, 1, 4, 4] Each group of four NPU
// taskNPUNumber: 11    priorityArray: [1,2,3,4]
// occupy first policy: 0, 2, 16, 18, 25, 29     remains: 5
// select the chip policy randomly: [5, 7, 8], [13, 15, 16], [20, 21, 22] choose one group,
// Assume that choice [13, 15, 16] remains: 2
// select the NPU policy randomly: [5, 7, 8], [20, 21, 22] choose two at random from one of the groups
func getFitCardFromNodeByPriority(taskNPUNumber int, nodeTop []int, priorityArray [cardNPUNumber]int) ([]int, error) {
	rand.Seed(time.Now().UnixNano())
	cardNumGroups := getCardNumGroupsFromTop(nodeTop)
	// index: indicates the number of remaining chips in the card，value: index array of the card
	npuNumberIndex := getNPUIndex(cardNumGroups)
	selectedCardTop, err := getSelectedCardTop(taskNPUNumber, priorityArray, cardNumGroups, npuNumberIndex)
	if err != nil {
		klog.V(logErrorLev).Infof("%s getHccsFromNodeByPriority: %v %s.", PluginName, nodeTop, err.Error())
		return nil, err
	}
	return selectedCardTop, nil
}

// getSelectedCardTop get selected card top
// priority to fill->Randomly chosen card->Select the chips on the cards at random
func getSelectedCardTop(taskNPUNumber int, priorityArray [4]int, cardNumGroups [][]int,
	npuNumberIndex [][]int) ([]int, error) {

	err := errors.New("nodeTop not meet")
	selectedNodeTop := make([]int, 0, taskNPUNumber)
	for _, arrLen := range priorityArray {
		curGroup := npuNumberIndex[arrLen]
		curGroupNum := arrLen * len(curGroup)
		if curGroupNum == 0 {
			continue
		}
		// curGroupNum <= taskNPUNumber: use all the chips in the group
		// curGroupNum >  taskNPUNumber: random use of part of the card chip(Use the idea of perfect shuffling)
		if curGroupNum > taskNPUNumber {
			cardNum := taskNPUNumber / arrLen
			cardIndex := getKRandNumFromN(len(curGroup), cardNum+1)
			for i := 1; i <= cardNum; i++ {
				selectedNodeTop = append(selectedNodeTop, cardNumGroups[curGroup[cardIndex[i]]]...)
			}
			remain := taskNPUNumber - cardNum*arrLen
			npuIndex := getKRandNumFromN(arrLen, remain)
			for i := 0; i < len(npuIndex); i++ {
				selectedNodeTop = append(selectedNodeTop, cardNumGroups[curGroup[cardIndex[0]]][npuIndex[i]])
			}
		} else {
			for _, GroupIndex := range curGroup {
				selectedNodeTop = append(selectedNodeTop, cardNumGroups[GroupIndex]...)
			}
		}
		taskNPUNumber -= curGroupNum
		if taskNPUNumber <= 0 {
			return selectedNodeTop, nil
		}
	}
	return nil, err
}

// getNPUIndex get NPU index by cardNumGroups
func getNPUIndex(cardNumGroups [][]int) [][]int {
	npuNumberIndex := make([][]int, constIntNum5)
	for index, cardNumGroup := range cardNumGroups {
		npuNumberIndex[len(cardNumGroup)] = append(npuNumberIndex[len(cardNumGroup)], index)
	}
	return npuNumberIndex
}

func getKRandNumFromN(n, k int) []int {
	arr := make([]int, n)
	for i := 0; i < n; i++ {
		arr[i] = i
	}
	for i := 0; i < k; i++ {
		n--
		randNum := rand.Intn(n + 1)
		arr[randNum], arr[n] = arr[n], arr[randNum]
	}
	return arr[n:]
}
