/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package rescheduling is using for HuaWei Ascend pin fault rescheduling.

*/
package rescheduling

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/util"
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
		klog.V(util.LogErrorLev).Infof("isNPUFaultTask: no fault job in ReSchedulerData .")
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
		klog.V(util.LogErrorLev).Infof("%s is npu fault job.", task.Job)
		return nil
	}

	if IsNodeInFaultNodeList(node) {
		msg := fmt.Errorf("%s is in fault node cache", node.Name)
		klog.V(util.LogErrorLev).Infof("IsNodeInFaultNodeList %v.", msg)
		return msg
	}

	if isNodeInFaultJobUseList(node) {
		msg := fmt.Errorf("%s is used by npu fault job:%s", node.Name, task.Job)
		klog.V(util.LogErrorLev).Infof("%v.", msg)
		return msg
	}

	klog.V(util.LogErrorLev).Infof("%s not in fault job use node list.", node.Name)

	return nil
}

func getPodRankIndex(pod *v1.Pod) (string, error) {
	rankIndex, ok := pod.Annotations[podRankIndex]
	if !ok {
		return "", errors.New("nil rankIndex")
	}

	index, err := strconv.Atoi(rankIndex)
	if err != nil {
		return "", fmt.Errorf("convert %#v:%#v", rankIndex, err)
	}

	if index > maxRankIndex || index < 0 {
		return "", fmt.Errorf("rankIndex:%v out of limit", index)
	}

	return rankIndex, nil
}

// AddScoreByFaultNPUTask returns nil to indicate that it is selected, and anything else to indicate that it is not.
func AddScoreByFaultNPUTask(task *api.TaskInfo, scoreMap map[string]float64) (map[string]float64, error) {
	if len(scoreMap) == 0 {
		mgs := fmt.Errorf("add score by fault NPU task scoreMap is nil")
		klog.V(util.LogErrorLev).Infof("%v.", mgs)
		return scoreMap, mgs
	}

	value, err := getReSchedulerTasksFromCache(task)
	if err != nil {
		return scoreMap, err
	}

	klog.V(util.LogErrorLev).Infof("getReSchedulerTasksFromCache :%v.", value)
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

func setOnOldNodeTaskRankIndex(rTask ReSchedulerTasks, task *api.TaskInfo, node *api.NodeInfo, now int64) error {
	nodeAndIndexList, getOldErr := getOldTaskNodeAndIndexList(rTask)
	if getOldErr != nil {
		klog.V(util.LogErrorLev).Infof("getOldTaskNodeAndIndexList: %v.", getOldErr)
		return getOldErr
	}

	if rankIndex, ok := nodeAndIndexList[node.Name]; ok {
		task.Pod.Annotations[podRankIndex] = rankIndex
		klog.V(util.LogInfoLev).Infof("%d: %s set old %s rankIndex %v.", now, task.Name, node.Name, rankIndex)
		return nil
	}
	return fmt.Errorf("%s not the previously used by the %s", node.Name, task.Name)
}

func getRankIndexMapByTask(task *api.TaskInfo) (TaskUsedRankIndex, error) {
	rankIndexData, ok := ReSchedulerCache[TmpAllocRankIndexKind]
	if !ok {
		return TaskUsedRankIndex{}, fmt.Errorf("no rankIndex cache")
	}
	rankIndexJobMap, assertOk := rankIndexData.(map[api.JobID]TaskUsedRankIndex)
	if !assertOk {
		return TaskUsedRankIndex{}, fmt.Errorf("%v assert map[api.JobID]TaskUsedRankIndex error", rankIndexData)
	}
	rankIndexMap, mapOk := rankIndexJobMap[task.Job]
	if !mapOk {
		return TaskUsedRankIndex{}, fmt.Errorf("no rankIndex in rankIndexMap")
	}
	return rankIndexMap, nil
}

func updateRankIndexMapByTask(task *api.TaskInfo, rankIndex string, now int64) error {
	rankIndexMap, getErr := getRankIndexMapByTask(task)
	if getErr != nil {
		return getErr
	}
	data, ok := rankIndexMap.FaultNodeRankIndex[rankIndex]
	if !ok {
		return fmt.Errorf("%s not in rankIndexMap", rankIndex)
	}
	data.UpdateTime = now
	rankIndexMap.FaultNodeRankIndex[rankIndex] = data

	rankIndexData, ok := ReSchedulerCache[TmpAllocRankIndexKind]
	if !ok {
		return fmt.Errorf("no rankIndex cache")
	}
	rankIndexJobMap, assertOk := rankIndexData.(map[api.JobID]TaskUsedRankIndex)
	if !assertOk {
		return fmt.Errorf("%v assert to map[api.JobID]TaskUsedRankIndex failed", rankIndexData)
	}
	rankIndexMap.UpdateTime = now
	rankIndexJobMap[task.Job] = rankIndexMap
	ReSchedulerCache[TmpAllocRankIndexKind] = rankIndexJobMap
	return nil
}

func setOnNewNodeTaskRankIndex(task *api.TaskInfo, node *api.NodeInfo, now int64) error {
	rankIndexMap, getErr := getRankIndexMapByTask(task)
	if getErr != nil {
		return getErr
	}

	for rankIndex, data := range rankIndexMap.FaultNodeRankIndex {
		if rankIndexMap.UpdateTime == now {
			if data.UpdateTime == now {
				klog.V(util.LogInfoLev).Infof("%d %s has been used %v before.", now, task.Pod.Name, rankIndex)
				continue
			}
		}
		if updateErr := updateRankIndexMapByTask(task, rankIndex, now); updateErr != nil {
			return updateErr
		}
		task.Pod.Annotations[podRankIndex] = rankIndex
		klog.V(util.LogInfoLev).Infof("%d: %s set new %s rankIndex %v.", now, task.Name, node.Name, rankIndex)
		return nil
	}

	return fmt.Errorf("%s set rankIndex %s failed", task.Name, node.Name)
}

func setReSchedulerTaskRankIndex(rTask ReSchedulerTasks, task *api.TaskInfo, node *api.NodeInfo) error {
	klog.V(util.LogErrorLev).Infof("%s SetFaultJobPodIndex from: %+v on %v.", task.Name, rTask, node.Name)
	now := time.Now().Unix()
	setOldNodeErr := setOnOldNodeTaskRankIndex(rTask, task, node, now)
	if setOldNodeErr == nil {
		return nil
	}
	klog.V(util.LogInfoLev).Infof("setReSchedulerTaskRankIndex on old node %v.", setOldNodeErr)

	setNewNodeErr := setOnNewNodeTaskRankIndex(task, node, now)
	if setNewNodeErr != nil {
		klog.V(util.LogInfoLev).Infof("setReSchedulerTaskRankIndex on new node %v.", setNewNodeErr)
		return setNewNodeErr
	}

	klog.V(util.LogInfoLev).Infof("setReSchedulerTaskRankIndex %s on new %s success.", task.Name, node.Name)
	return nil
}

// SetFaultJobPodIndex Set the rankIndex of all pods of the failed task
// For there are gaps in the status of the volcano update podgroup, so set fault job rankIndex by job not task.
func SetFaultJobPodIndex(task *api.TaskInfo, node *api.NodeInfo) error {
	tmpValue, err := getReSchedulerTasksFromCache(task)
	if err != nil || reflect.DeepEqual(tmpValue, ReSchedulerTasks{}) {
		klog.V(util.LogInfoLev).Infof("SetFaultJobPodIndex %s %v.", task.Name, err)
		return err
	}

	if setErr := setReSchedulerTaskRankIndex(tmpValue, task, node); setErr != nil {
		klog.V(util.LogInfoLev).Infof("SetFaultJobPodIndex %s %v.", task.Name, setErr)
		return setErr
	}
	return nil
}
