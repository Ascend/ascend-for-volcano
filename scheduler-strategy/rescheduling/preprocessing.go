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
	"fmt"
	"k8s.io/klog"
	time2 "time"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"
	hwutil "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/util"
)

func init() {
	ReSchedulerCache = make(map[string]interface{}, constIntNum3)
}

// Delete expired data.
func updateReSchedulerDataFromSession(ssn *framework.Session) error {
	klog.V(logDebugLev).Infof("updateReSchedulerDataFromSession get buffer %v.", ReSchedulerCache)
	for dataID, tmpValue := range ReSchedulerCache {
		switch dataID {
		case CmJobKind:
			if err := synReSchedulerJobCache(ssn, tmpValue); err != nil {
				klog.V(logDebugLev).Infof("synReSchedulerDataAndCache %v.", err)
			}
			continue
		case CmNodeKind:
			if err := synReSchedulerNodeCache(ssn, tmpValue); err != nil {
				klog.V(logDebugLev).Infof("synReSchedulerDataAndCache %v.", err)
			}
			continue
		case CmCardKind:
			if err := synReSchedulerCardCache(ssn, tmpValue); err != nil {
				klog.V(logDebugLev).Infof("synReSchedulerDataAndCache %v.", err)
			}
			continue
		default:
			klog.V(logErrorLev).Infof("not support %v.", dataID)
		}
	}

	return nil
}

// Write fault resource(NPUs,nodes) into cache.
func writeFaultResourceInfInCache(npus []FaultNPUsOnNode, nodes []FaultNodeState) error {
	var cardMap = make(map[string]FaultNPUsOnNode, 1)
	for _, card := range npus {
		tmp := card
		cardMap[card.NodeName] = tmp
	}
	ReSchedulerCache[CmCardKind] = cardMap

	var nodeMap = make(map[string]FaultNodeState, 1)
	for _, nodeState := range nodes {
		tmp := nodeState
		nodeMap[nodeState.NodeName] = tmp
	}
	ReSchedulerCache[CmNodeKind] = nodeMap

	return nil
}

// RecordFaultInfInCache Record the fault information(card/node) in the cache
func RecordFaultInfInCache(ssn *framework.Session, npuNumber int) error {
	// 1.Get fault NPUs and its nodes from running vcjob.
	faultNPUs, npuErr := getInoperableNPUCards(ssn.Nodes, npuNumber)
	if npuErr != nil {
		klog.V(logDebugLev).Infof("getInoperableNPUCards %v.", npuErr)
	}
	// 2.Obtaining the Faulty Node from nodeD.
	faultNodes, nodeErr := getInoperableNodes(ssn.Nodes)
	if nodeErr != nil {
		klog.V(logDebugLev).Infof("getInoperableNodes %v.", nodeErr)
	}

	if npuErr != nil && nodeErr != nil {
		return fmt.Errorf("%v %v", npuErr, nodeErr)
	}

	if writeErr := writeFaultResourceInfInCache(faultNPUs, faultNodes); writeErr != nil {
		klog.V(logErrorLev).Infof("writeFaultResourceInfInCache %v.", writeErr)
		return writeErr
	}
	return nil
}

// SetFaultInNodeAndJobs Recorded the information about the faulty task in the cache.
func SetFaultInNodeAndJobs(fNPUJobs []FaultNPUJob, jobs map[string]*api.JobInfo) error {
	for _, tmpFaultNPUJob := range fNPUJobs {
		tmpTask := ReSchedulerTasks{
			NodeNames:   make(map[string]string, constIntNum3),
			RankIndexes: make(map[string]string, constIntNum3),
			Time:        make(map[string]int64, constIntNum3),
			TaskUseNPUs: make(map[string]string, constIntNum3),
			NameSpace:   tmpFaultNPUJob.namespace}

		for taskName, nodeName := range tmpFaultNPUJob.taskUseNode {
			rankIndex, indexOK := tmpFaultNPUJob.taskUseRankIndex[taskName]
			if !indexOK {
				klog.V(logErrorLev).Infof("%s get rankIndex failed.", taskName)
				continue
			}

			useNPUs, npuOK := tmpFaultNPUJob.taskUseNPUs[taskName]
			if !npuOK {
				klog.V(logErrorLev).Infof("%s get use NPUs failed.", taskName)
				continue
			}
			tmpTask.NodeNames[taskName] = nodeName
			tmpTask.TaskUseNPUs[taskName] = useNPUs
			tmpTask.RankIndexes[taskName] = rankIndex
			tmpTask.Time[taskName] = time2.Now().Unix()
		}

		if err := writeFaultJobInfInCache(jobs, tmpFaultNPUJob, tmpTask); err != nil {
			klog.V(logErrorLev).Infof("recordNPUFaultJobToBuffer :%v.", err)
			return err
		}
	}

	return nil
}

// GetNodeIdleNPUIntCardsIncludeFaultTask Getting the number of NPU chips idle on the node includes the failure task
func GetNodeIdleNPUIntCardsIncludeFaultTask(task *api.TaskInfo, node *api.NodeInfo) []int {
	var returnNPUs []int
	var temp []int

	nodeNPUTopology := hwutil.GetTopFromNodeOthers(node, npu800And9000CardName, npu800And9000CardPreName)
	taskNPUTopology := getTaskUseNPUIntCards(task, npu800And9000CardName, npu800And9000CardPreName)
	temp = append(temp, nodeNPUTopology...)
	allIntNPUs := append(temp, taskNPUTopology...)

	allNodeFaultNPUs, err := getNodeFaultNPUsByInt(node)
	if err != nil {
		return nil
	}

	// Gets the difference set of slices.
	for _, card := range allIntNPUs {
		i := 0
		for _, fCard := range allNodeFaultNPUs {
			if card == fCard {
				i++
				break
			}
		}
		if i == 0 {
			returnNPUs = append(returnNPUs, card)
		}
	}

	return returnNPUs
}
