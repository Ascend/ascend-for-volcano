/*
Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package rescheduling is using for HuaWei Ascend pin fault rescheduling.

*/
package rescheduling

import (
	"errors"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog"
	"reflect"
	"strconv"
	"strings"
	"time"
	"volcano.sh/volcano/pkg/scheduler/api"
)

func isTaskHasFaultNPU(taskNPUs []string, nodeFaultNPUs []string) bool {
	for _, taskNPU := range taskNPUs {
		for _, nodeFaultNPU := range nodeFaultNPUs {
			if strings.EqualFold(nodeFaultNPU, taskNPU) {
				return true
			}
		}
	}
	return false
}

// IsNPUFaultTask Check the NPU task whether is fault task.
func IsNPUFaultTask(task *api.TaskInfo) bool {
	value, valueOk := ReSchedulerCache[CmJobKind]
	if !valueOk {
		klog.V(logDebugLev).Infof("isNPUFaultTask: no fault job in ReSchedulerData .")
		return false
	}

	jobs, assertOK := value.(map[api.JobID]ReSchedulerTasks)
	if !assertOK {
		return false
	}

	if _, ok := jobs[task.Job]; ok {
		return true
	}
	return false
}

// CheckFaultJobNode During the pre-selection phase, check whether the node meets the fault task.
func CheckFaultJobNode(task *api.TaskInfo, node *api.NodeInfo) error {
	if IsNPUFaultTask(task) {
		klog.V(logErrorLev).Infof("%s is npu fault job.", task.Job)
		return nil
	}

	if IsNodeInFaultNodeList(node) {
		msg := fmt.Errorf("%s is in fault node cache", node.Name)
		klog.V(logErrorLev).Infof("IsNodeInFaultNodeList %v.", msg)
		return msg
	}

	if isNodeInFaultJobUseList(node) {
		msg := fmt.Errorf("%s is used by npu fault job:%s", node.Name, task.Job)
		klog.V(logErrorLev).Infof("%v.", msg)
		return msg
	}

	klog.V(logDebugLev).Infof("%s not in fault job use node list.", node.Name)

	return nil
}

func getPodRankIndex(pod *v1.Pod) (string, error) {
	rankIndex, ok := pod.Annotations[podRankIndex]
	if !ok {
		return "", errors.New("nil rankIndex")
	}

	index, err := strconv.Atoi(rankIndex)
	if err != nil {
		return "", fmt.Errorf("convert %v:%v", rankIndex, err)
	}

	if index > maxRankIndex {
		return "", fmt.Errorf("rankIndex:%v out of limit", index)
	}

	return rankIndex, nil
}

// AddScoreByFaultNPUTask returns nil to indicate that it is selected, and anything else to indicate that it is not.
func AddScoreByFaultNPUTask(task *api.TaskInfo, scoreMap map[string]float64) (map[string]float64, error) {
	if len(scoreMap) == 0 {
		mgs := fmt.Errorf("AddScoreByFaultNPUTask scoreMap is nil")
		klog.V(logErrorLev).Infof("%v.", mgs)
		return scoreMap, mgs
	}

	value, err := getReSchedulerTasksFromCache(task)
	if err != nil {
		return scoreMap, err
	}

	klog.V(logDebugLev).Infof("getReSchedulerTasksFromCache :%v.", value)
	for key, taskName := range value.TaskName {
		if taskName != task.Name {
			continue
		}
		nodeName := value.NodeNames[key]
		if _, ok := scoreMap[nodeName]; !ok {
			return scoreMap, nil
		}
		scoreMap[nodeName] += node910X8NPUNum * node910X8NPUNum
	}

	return scoreMap, nil
}

func getOldTaskNodeAndIndexList(rTask ReSchedulerTasks) (map[string]string, error) {
	var nodeAndIndexList = make(map[string]string, 1)
	for key := range rTask.TaskName {
		nodeName := rTask.NodeNames[key]
		rankIndex := rTask.RankIndexes[key]
		nodeAndIndexList[nodeName] = rankIndex
	}
	return nodeAndIndexList, nil
}

func setOnOldNodeTaskRankIndex(rTask ReSchedulerTasks, task *api.TaskInfo, node *api.NodeInfo) error {
	nodeAndIndexList, getOldErr := getOldTaskNodeAndIndexList(rTask)
	if getOldErr != nil {
		klog.V(logErrorLev).Infof("getOldTaskNodeAndIndexList: %v.", getOldErr)
	}

	now := time.Now().Unix()
	if rankIndex, ok := nodeAndIndexList[node.Name]; ok {
		klog.V(logInfoLev).Infof("%d: %s set old %s rankIndex %v.", now, task.Pod.Name, node.Name, rankIndex)
		task.Pod.Annotations[podRankIndex] = rankIndex
		return nil
	}
	return fmt.Errorf("%s not the previously used by the %s", node.Name, task.UID)
}

func setOnNewNodeTaskRankIndex(task *api.TaskInfo, node *api.NodeInfo) error {
	rankIndexData, ok := ReSchedulerCache[TmpAllocRankIndexKind]
	if !ok {
		return fmt.Errorf("no rankIndex cache")
	}
	rankIndexJobMap, assertOk := rankIndexData.(map[api.JobID]TaskUsedRankIndex)
	if !assertOk {
		return fmt.Errorf("%v assert to map[api.JobID]TaskUsedRankIndex failed", rankIndexData)
	}
	rankIndexMap, mapOk := rankIndexJobMap[task.Job]
	if !mapOk {
		return fmt.Errorf("no rankIndex in rankIndexMap")
	}

	now := time.Now().Unix()
	for rankIndex, data := range rankIndexMap.FaultNodeRankIndex {
		if rankIndexMap.UpdateTime == now {
			if data.UpdateTime == now {
				klog.V(logInfoLev).Infof("%d %s cannot use %v, has been used.", now, task.Pod.Name, rankIndex)
				continue
			}
		}
		klog.V(logInfoLev).Infof("%d: %s set new %s rankIndex %v.", now, task.Pod.Name, node.Name, rankIndex)
		task.Pod.Annotations[podRankIndex] = rankIndex
		rankIndexMap.UpdateTime = now
		data.UpdateTime = now
		rankIndexMap.FaultNodeRankIndex[rankIndex] = data
		rankIndexJobMap[task.Job] = rankIndexMap
		ReSchedulerCache[TmpAllocRankIndexKind] = rankIndexJobMap
		return nil
	}

	return fmt.Errorf("%s set rankIndex %s failed", task.Name, node.Name)
}

func setReSchedulerTaskRankIndex(rTask ReSchedulerTasks, task *api.TaskInfo, node *api.NodeInfo) error {
	klog.V(logDebugLev).Infof("%s SetFaultJobPodIndex from: %+v on %v.", task.Name, rTask, node.Name)

	setOldNodeErr := setOnOldNodeTaskRankIndex(rTask, task, node)
	if setOldNodeErr == nil {
		return nil
	}
	klog.V(logInfoLev).Infof("setReSchedulerTaskRankIndex %v.", setOldNodeErr)

	setNewNodeErr := setOnNewNodeTaskRankIndex(task, node)
	if setNewNodeErr != nil {
		klog.V(logInfoLev).Infof("setReSchedulerTaskRankIndex %v.", setNewNodeErr)
		return setNewNodeErr
	}

	klog.V(logInfoLev).Infof("setReSchedulerTaskRankIndex %s on %s success.", task.Name, node.Name)
	return nil
}

// SetFaultJobPodIndex Set the rankIndex of all pods of the failed task
// For there are gaps in the status of the volcano update podgroup, so set fault job rankIndex by job not task.
func SetFaultJobPodIndex(task *api.TaskInfo, node *api.NodeInfo) error {
	klog.V(logErrorLev).Infof("这里-2，%+v.", task)
	tmpValue, err := getReSchedulerTasksFromCache(task)
	if err != nil || reflect.DeepEqual(tmpValue, ReSchedulerTasks{}) {
		klog.V(logInfoLev).Infof("SetFaultJobPodIndex %s %v.", task.Name, err)
		return err
	}

	if setErr := setReSchedulerTaskRankIndex(tmpValue, task, node); setErr != nil {
		klog.V(logInfoLev).Infof("setReSchedulerTaskRankIndex %s %v.", task.Name, setErr)
		return setErr
	}
	return nil
}
