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
	"strconv"
	"strings"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/util"
)

// According to need card number,get best node from 4 pri-node-list-group.
func getBestNodesMap(priNodeGroups []map[string]*npuPriNodeInf) (map[string]int, error) {
	var bestNodesMap = make(map[string]int, constIntNum2)

	for i := 0; i < npuNumPerHccs; i++ {
		for nodeName := range priNodeGroups[i] {
			tmpName := nodeName
			bestNodesMap[tmpName] = i
		}
	}

	if len(bestNodesMap) == 0 {
		return nil, fmt.Errorf("%s none bestNodes", PluginName)
	}

	return bestNodesMap, nil
}

func isNPUFaultTask(task *api.TaskInfo) bool {
	if _, ok := util.ReSchedulerJobs[task.Job]; ok {
		return true
	}
	return false
}

func getPodRankIndex(pod *v1.Pod) (string, error) {
	rankIndex, ok := pod.Annotations[podRankIndex]
	if !ok {
		return "", errors.New("nil rankIndex")
	}

	index, err := strconv.Atoi(rankIndex)
	if err != nil {
		return "", err
	}

	if index > maxRankIndex {
		return "", errors.New("rankIndex out Out of limit")
	}

	return rankIndex, nil
}

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

func getTaskUseNPUs(nodesTask map[string]*v1.Pod, nodeName string) ([]string, error) {
	tmpPod, ok := nodesTask[nodeName]
	if !ok {
		return nil, fmt.Errorf("not use %s", nodeName)
	}

	strNpu, npuOK := tmpPod.Annotations[npu800And9000CardName]
	if !npuOK {
		return nil, fmt.Errorf("%s has no NPU", tmpPod.Name)
	}

	taskNPUs := strings.Split(strNpu, ",")
	if len(taskNPUs) > nodeNPUNumber {
		err := fmt.Errorf("get err %s npus %v", tmpPod.Name, taskNPUs)
		return nil, err
	}

	return taskNPUs, nil
}

func getTaskUseNodeInfo(task *api.TaskInfo, ssn *framework.Session) (*api.NodeInfo, error) {
	faultTasks, ok := util.ReSchedulerJobs[task.Job]
	if !ok {
		return nil, fmt.Errorf("get jobId%s failed", task.Job)
	}

	nodeName, ok := faultTasks.NodeNames[task.Name]
	if !ok {
		return nil, fmt.Errorf("get taskName %s failed", task.Name)
	}

	node, ok := ssn.Nodes[nodeName]
	if !ok {
		return nil, fmt.Errorf("get node name %s failed", nodeName)
	}

	return node, nil
}
