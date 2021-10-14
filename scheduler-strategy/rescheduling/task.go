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

Package rescheduling is using for HuaWei Ascend pin fault rescheduling.

*/
package rescheduling

import (
	"errors"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog"
	"strconv"
	"strings"
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
	for taskName, nodeName := range value.NodeNames {
		if taskName != task.Name {
			continue
		}
		if _, ok := scoreMap[nodeName]; !ok {
			return scoreMap, nil
		}
		scoreMap[nodeName] += node910X8NPUNum * node910X8NPUNum
	}

	return scoreMap, nil
}

// SetFaultJobPodIndex Set the rankIndex of all pods of the failed task
// For there are gaps in the status of the volcano update podgroup, so set fault job rankIndex by job not task.
func SetFaultJobPodIndex(job *api.JobInfo) error {
	for _, task := range job.Tasks {
		tmpValue, err := getReSchedulerTasksFromCache(task)
		if err != nil {
			klog.V(logInfoLev).Infof("SetFaultJobPodIndex %s %v.", task.Name, err)
			return err
		}
		klog.V(logDebugLev).Infof("%s SetFaultJobPodIndex from buffer: %v.", task.Job, tmpValue)
		for taskName, rankIndex := range tmpValue.RankIndexes {
			if taskName == task.Name {
				klog.V(logInfoLev).Infof("set %s rankIndex %v.", task.Pod.Name, rankIndex)
				task.Pod.Annotations[podRankIndex] = rankIndex
				break
			}
		}
	}
	return nil
}
