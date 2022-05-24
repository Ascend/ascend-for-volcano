/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package card310x4 is using for HuaWei A300T Ascend pin affinity schedule.

*/
package card310x4

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/util"
)

func judgeNodeAndTaskNPU(taskNPU int, nodeNPUTopology []int) error {
	cardNumGroups := getCardNumGroupsFromTop(nodeNPUTopology)

	for _, cardNumGroup := range cardNumGroups {
		if len(cardNumGroup) >= taskNPU {
			return nil
		}
	}

	var meetErr = fmt.Errorf("req npu(%d) illegal", taskNPU)

	klog.V(util.LogDebugLev).Infof("%s %v.", PluginName, meetErr)
	return meetErr
}

// Initializes the node priority series group according to the priority scheduling policy of 1 card.
func initOneCardPriNodeGroups(
	cardNumGroups [][]int,
	priNodeGroups []map[string]*npuPriNodeInf,
	addPriNodeGroupFn initPriNodeGroupFn) error {
	if len(priNodeGroups) < cardNPUNumber {
		err := fmt.Errorf("priNodeGroups's length(%d) not meet(%d)", len(priNodeGroups), cardNPUNumber)
		klog.V(util.LogDebugLev).Infof("%s initOneCardPriNodeGroups %v.", PluginName, err.Error())
		return err
	}
	bestGrade := cardNPUNumber
	// priority:1>3>2>4
	for _, cardNumGroup := range cardNumGroups {
		switch len(cardNumGroup) {
		case util.NPUIndex1:
			bestGrade = util.NPUIndex0
		case util.NPUIndex3:
			bestGrade = min(bestGrade, util.NPUIndex1)
		case util.NPUIndex2:
			bestGrade = min(bestGrade, util.NPUIndex2)
		case cardNPUNumber:
			bestGrade = min(bestGrade, util.NPUIndex3)
		default:
			klog.V(util.LogErrorLev).Infof("%s initOneCardPriNodeGroups err:%v.", PluginName, cardNumGroup)
		}
		if bestGrade == util.NPUIndex0 {
			break
		}
	}
	if bestGrade == cardNPUNumber {
		// no satisfy
		klog.V(util.LogDebugLev).Infof("%s initOneCardPriNodeGroups node(%v) cannot fit.", PluginName, cardNumGroups)
		return errors.New(nodeNoFitNPUWarning)
	}
	addPriNodeGroupFn(priNodeGroups[bestGrade], strconv.Itoa(bestGrade))
	return nil
}

// Initializes the node priority series group according to the priority scheduling policy of 2 cards.
func initTwoCardPriNodeGroups(
	cardNumGroups [][]int,
	priNodeGroups []map[string]*npuPriNodeInf,
	addPriNodeGroupFn initPriNodeGroupFn) error {

	if len(priNodeGroups) < cardNPUNumber {
		err := fmt.Errorf("priNodeGroups's length(%d) not meet(%d)", len(priNodeGroups), cardNPUNumber)
		klog.V(util.LogDebugLev).Infof("%s initOneCardPriNodeGroups %v.", PluginName, err.Error())
		return err
	}

	bestGrade := cardNPUNumber

	// priority:2>3>4
	for _, cardNumGroup := range cardNumGroups {
		switch len(cardNumGroup) {
		case util.NPUIndex2:
			bestGrade = util.NPUIndex0
		case util.NPUIndex3:
			bestGrade = min(bestGrade, util.NPUIndex1)
		case cardNPUNumber:
			bestGrade = min(bestGrade, util.NPUIndex2)
		default:
			klog.V(util.LogErrorLev).Infof("%s initTwoCardPriNodeGroups err:%v.", PluginName, cardNumGroup)
		}
		if bestGrade == util.NPUIndex0 {
			break
		}
	}

	if bestGrade == cardNPUNumber {
		// no satisfy
		klog.V(util.LogDebugLev).Infof("%s initOneCardPriNodeGroups node(%v) cannot fit.", PluginName, cardNumGroups)
		return errors.New(nodeNoFitNPUWarning)
	}

	addPriNodeGroupFn(priNodeGroups[bestGrade], strconv.Itoa(bestGrade))
	return nil
}

func initThreeCardPriNodeGroups(
	cardNumGroups [][]int,
	priNodeGroups []map[string]*npuPriNodeInf,
	addPriNodeGroupFn initPriNodeGroupFn) error {

	if len(priNodeGroups) < cardNPUNumber {
		err := fmt.Errorf("priNodeGroups's length(%d) not meet(%d)", len(priNodeGroups), cardNPUNumber)
		klog.V(util.LogDebugLev).Infof("%s initOneCardPriNodeGroups %v.", PluginName, err.Error())
		return err
	}

	bestGrade := cardNPUNumber

	// priority:3>4
	for _, cardNumGroup := range cardNumGroups {
		switch len(cardNumGroup) {
		case util.NPUIndex3:
			bestGrade = util.NPUIndex0
		case cardNPUNumber:
			bestGrade = min(bestGrade, util.NPUIndex1)
		default:
			klog.V(util.LogErrorLev).Infof("%s initThreeCardPriNodeGroups err:%v.", PluginName, cardNumGroup)
		}
		if bestGrade == util.NPUIndex0 {
			break
		}
	}

	if bestGrade == cardNPUNumber {
		// no satisfy
		klog.V(util.LogDebugLev).Infof("%s initOneCardPriNodeGroups node(%v) cannot fit.", PluginName, cardNumGroups)
		return errors.New(nodeNoFitNPUWarning)
	}

	addPriNodeGroupFn(priNodeGroups[bestGrade], strconv.Itoa(bestGrade))
	return nil
}

func initFourCardPriNodeGroups(
	cardNumGroups [][]int,
	priNodeGroups []map[string]*npuPriNodeInf,
	addPriNodeGroupFn initPriNodeGroupFn) error {

	if len(priNodeGroups) < cardNPUNumber {
		err := fmt.Errorf("priNodeGroups's length(%d) not meet(%d)", len(priNodeGroups), cardNPUNumber)
		klog.V(util.LogDebugLev).Infof("%s initOneCardPriNodeGroups %v.", PluginName, err.Error())
		return err
	}

	bestGrade := cardNPUNumber

	// priority:3>4
	for _, cardNumGroup := range cardNumGroups {
		if len(cardNumGroup) == cardNPUNumber {
			// A group
			bestGrade = util.NPUIndex0
			break
		}
	}

	if bestGrade == cardNPUNumber {
		// no satisfy
		klog.V(util.LogDebugLev).Infof("%s initOneCardPriNodeGroups node(%v) cannot fit.", PluginName, cardNumGroups)
		return errors.New(nodeNoFitNPUWarning)
	}

	addPriNodeGroupFn(priNodeGroups[bestGrade], strconv.Itoa(bestGrade))
	return nil
}

// Place nodes in the priority group.
func insertNodeInPriGroup(
	task *api.TaskInfo,
	cardNumGroups [][]int,
	priNodeGroups []map[string]*npuPriNodeInf,
	addPriNodeGroupFn initPriNodeGroupFn) error {

	var err error

	// 1.Get the task's NPU request
	taskReqNPU, errGet := util.GetTaskNPUNum(task, a310NPUCardName)
	if errGet != nil {
		// cannot return error for task is no npu kind possible.
		klog.V(util.LogDebugLev).Infof("%s batchNodeOrderFn task :%s,%v.", PluginName, task.Name, errGet)
		// cannot return nil，will panic
		return nil
	}

	switch taskReqNPU {
	case util.NPUIndex0:
		klog.V(util.LogInfoLev).Infof("%s initPriNodeGroups task npu is 0.", PluginName)
	case util.NPUIndex1:
		err = initOneCardPriNodeGroups(cardNumGroups, priNodeGroups, addPriNodeGroupFn)
	case util.NPUIndex2:
		err = initTwoCardPriNodeGroups(cardNumGroups, priNodeGroups, addPriNodeGroupFn)
	case util.NPUIndex3:
		err = initThreeCardPriNodeGroups(cardNumGroups, priNodeGroups, addPriNodeGroupFn)
	case cardNPUNumber:
		err = initFourCardPriNodeGroups(cardNumGroups, priNodeGroups, addPriNodeGroupFn)
	default:
		// For normal,can not be here. The pre function validate job has done this.
		klog.V(util.LogDebugLev).Infof("%s node(%v) not fit task request %d.", PluginName, cardNumGroups, taskReqNPU)
		err = errors.New("illegal request npu number " + strconv.Itoa(taskReqNPU))
	}

	return err
}

func getNPUAllocPriorityArray(taskNPUNumber int) ([cardNPUNumber]int, error) {
	var priorityArray [cardNPUNumber]int
	var err = error(nil)
	switch taskNPUNumber {
	case util.NPUIndex0:
		klog.V(util.LogInfoLev).Infof("%s task req npu is 0.", PluginName)
	case util.NPUIndex1:
		// priority:1>3>2>4
		priorityArray = [cardNPUNumber]int{1, util.NPUIndex3, util.NPUIndex2, cardNPUNumber}
	case util.NPUIndex2:
		// priority：2>3>4
		priorityArray = [cardNPUNumber]int{util.NPUIndex2, util.NPUIndex3, cardNPUNumber}
	case util.NPUIndex3:
		// priority：3>4
		priorityArray = [cardNPUNumber]int{util.NPUIndex3, cardNPUNumber}
	case cardNPUNumber:
		priorityArray = [cardNPUNumber]int{cardNPUNumber}
	default:
		// For normal,can not be here. The pre function validate job has done this.
		err = fmt.Errorf("illegal request npu number: %d", taskNPUNumber)
	}
	if err != nil {
		klog.V(util.LogDebugLev).Infof("%s %s.", PluginName, err.Error())
		return priorityArray, err
	}
	return priorityArray, nil
}

// getFitCardFromNodeByPriority get the appropriate card randomly from the Node via Priority
// For example:
// nodeTop: [0, 2, 4, 5, 7, 8, 9, 10, 11, 12, 13, 15, 16, 18, 20, 21, 22, 25, 29, 32, 33, 34, 35, 36, 37, 38, 39]
// cardNumGroups: [2, 3, 4, 3, 2, 3, 1, 1, 4, 4] Each group of four NPU
// priorityArray： [1, 3, 2, 4] -> [25, 29] (Group 6, 7)
func getFitCardFromNodeByPriority(nodeTop []int, priorityArray [cardNPUNumber]int) ([]int, error) {
	rand.Seed(time.Now().UnixNano())
	existNPU, npuNumberIndex := getExistIDAndIndexGroupOfNPU(nodeTop)
	selectedCardTop, err := getSelectedCardTop(priorityArray, existNPU, npuNumberIndex)
	if err != nil {
		klog.V(util.LogDebugLev).Infof("%s getHccsFromNodeByPriority: %v %s.", PluginName, nodeTop, err.Error())
		return nil, err
	}
	return selectedCardTop, nil
}

// getSelectedCardTop get selected card top
func getSelectedCardTop(priorityArray [cardNPUNumber]int, existNPU map[int]bool, index [][]int) ([]int, error) {
	err := errors.New("nodeTop not meet")
	for _, arrLen := range priorityArray {
		selectedGroup := index[arrLen]
		if len(selectedGroup) == 0 {
			klog.V(util.LogDebugLev).Infof("%s getSelectedCardTop %d group %+v.", PluginName, arrLen, priorityArray)
			continue
		}
		randNum := rand.Intn(len(selectedGroup))
		selectedIndex := selectedGroup[randNum]
		var selectedNodeTop []int
		for i := selectedIndex * cardNPUNumber; i < (selectedIndex+1)*cardNPUNumber; i++ {
			if existNPU[i] {
				selectedNodeTop = append(selectedNodeTop, i)
			}
		}
		if len(selectedNodeTop) != arrLen {
			return nil, err
		}
		return selectedNodeTop, nil
	}
	return nil, err
}

// getExistIDAndIndexGroupOfNPU get existID and indexGroup Of NPU
func getExistIDAndIndexGroupOfNPU(nodeTop []int) (map[int]bool, [][]int) {
	cardNumGroups := getCardNumGroupsFromTop(nodeTop)
	npuNumberIndex := make([][]int, constIntNum5)
	existNPU := make(map[int]bool, len(nodeTop))
	for _, nodeID := range nodeTop {
		existNPU[nodeID] = true
	}
	for index, cardNumGroup := range cardNumGroups {
		npuNumberIndex[len(cardNumGroup)] = append(npuNumberIndex[len(cardNumGroup)], index)
	}
	return existNPU, npuNumberIndex
}
