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
	"context"
	"encoding/json"
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	"strconv"
	"strings"
	time2 "time"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/util"
)

func init() {
	ReSchedulerCache = make(map[string]interface{}, constIntNum3)
	reSchedulerJobController = map[string]struct{}{
		JobGraceRescheduleLabelValue: {},
		JobForceRescheduleLabelValue: {},
		JobOffRescheduleLabelValue:   {}}
}

// Delete expired data.
func updateReSchedulerDataFromSession(ssn *framework.Session) error {
	for dataID, tmpValue := range ReSchedulerCache {
		switch dataID {
		case CmJobKind:
			if err := synReSchedulerJobCache(ssn, tmpValue); err != nil {
				klog.V(logDebugLev).Infof("synReSchedulerJobCache %v.", err)
			}
		case CmNodeKind:
			if err := synReSchedulerNodeCache(ssn, tmpValue); err != nil {
				klog.V(logDebugLev).Infof("synReSchedulerNodeCache %v.", err)
			}
		case CmCardKind:
			if err := synReSchedulerCardCache(ssn, tmpValue); err != nil {
				klog.V(logDebugLev).Infof("synReSchedulerCardCache %v.", err)
			}
		case CmNodeHeartbeatKind:
			if err := synNodeHeartbeatCache(ssn, tmpValue); err != nil {
				klog.V(logDebugLev).Infof("synNodeHeartbeatCache %v.", err)
			}
		case CmJobRankIds:
			if err := synJobRankIdsCache(ssn, tmpValue); err != nil {
				klog.V(logDebugLev).Infof("synJobRankIdsCache %v.", err)
			}
		default:
			klog.V(logErrorLev).Infof("not support %v:%v.", dataID, tmpValue)
		}
	}
	// 2.add new node
	if err := updateNodeIntoNodesHeartbeatTmp(ssn); err != nil {
		klog.V(logErrorLev).Infof("updateNodeIntoNodesHeartbeatTmp %v.", err)
	}
	return nil
}

func getJobUsedNodeRankIds(job *api.JobInfo, nodeAndPods map[string]*v1.Pod) (
	map[api.JobID]FaultRankIDRecordJobCMData, error) {
	nodeRankIds := make(map[api.JobID]FaultRankIDRecordJobCMData, constIntNum3)
	var PodsName []string
	var PodsUID []types.UID
	var PodsCreatTime []int64
	var rankIds []string
	for _, task := range job.Tasks {
		tmpPod, ok := nodeAndPods[task.NodeName]
		if !ok {
			continue
		}
		rankIndexStr, err := getPodRankIndex(tmpPod)
		if err != nil {
			klog.V(logErrorLev).Infof("getPodRankIndex %s %v.", tmpPod.Name, err)
			continue
		}
		rankIndex, covertErr := strconv.Atoi(rankIndexStr)
		if covertErr != nil {
			klog.V(logErrorLev).Infof("%s getJobUsedNodeRankIds covert %v.", covertErr, rankIndexStr)
			continue
		}
		taskUseNPUsStr, err := getPodUsedNPUS(tmpPod, npu800And9000CardName, node910X8NPUNum)
		if err != nil {
			klog.V(logInfoLev).Infof("getPodUsedNPUS %v.", err)
			continue
		}
		taskUseNPUs := util.ChangeTopToIntArray(strings.Join(taskUseNPUsStr, ","), npu800And9000CardPreName)
		for _, tmp := range taskUseNPUs {
			rankIds = append(rankIds, strconv.Itoa(tmp+rankIndex*8))
		}
		PodsName = append(PodsName, tmpPod.Name)
		PodsUID = append(PodsUID, tmpPod.UID)
		PodsCreatTime = append(PodsCreatTime, tmpPod.CreationTimestamp.Unix())
	}
	dataBuffer, err := json.Marshal(rankIds)
	if err != nil {
		klog.V(logErrorLev).Infof("marshalCacheDataToString err: %v.", err)
		return nil, err
	}
	tmpData := FaultRankIDRecordJobCMData{
		NameSpace:    job.Namespace,
		FaultRankIds: string(dataBuffer),
		// key is podName,value is pod create time
		PodsName:      PodsName,
		PodsUID:       PodsUID,
		PodsCreatTime: PodsCreatTime,
		// this the record create time.
		CreatTime: time2.Now().Unix(),
	}
	nodeRankIds[job.UID] = tmpData
	return nil, nil
}

func addJobsRankIdsIntoCache(jobsRankIds map[api.JobID]FaultRankIDRecordJobCMData) error {
	jobsRankIdsFromCache, getErr := getRankIdJobsFromCache()
	if getErr != nil {
		klog.V(logDebugLev).Infof("addJobsRankIdsIntoCache %v.", getErr)
		return getErr
	}

	for jobID, rankIDData := range jobsRankIds {
		jobsRankIdsFromCache[jobID] = rankIDData
	}

	ReSchedulerCache[CmJobRankIds] = jobsRankIdsFromCache
	return nil
}

func writeFaultNodeRankIdsByJobInCache(ssn *framework.Session) error {
	if len(ssn.Jobs) == 0 {
		klog.V(logDebugLev).Infof("writeFaultNodeRankIdsByJobInCache none jobs in ssn.")
		return nil
	}
	for _, job := range ssn.Jobs {
		nodeAndPods, getErr := getRunningJobUsedNodes(job)
		if getErr != nil {
			klog.V(logDebugLev).Infof("%s getRunningJobUsedNodes %v.", job.Name, getErr)
			continue
		}
		if isJobHasFaultNodes(nodeAndPods) {
			klog.V(logDebugLev).Infof("%s isJobHasFaultNodes %+v.", job.Name, nodeAndPods)
			continue
		}
		jobsRankIds, getRankIdsErr := getJobUsedNodeRankIds(job, nodeAndPods)
		if getRankIdsErr != nil {
			klog.V(logDebugLev).Infof("%s getJobUsedNodeRankIds %s %+v.", job.Name, nodeAndPods, getRankIdsErr)
			continue
		}
		if addErr := addJobsRankIdsIntoCache(jobsRankIds); addErr != nil {
			klog.V(logDebugLev).Infof("%s addJobsRankIdsIntoCache %v %+v.", job.Name, jobsRankIds, addErr)
			continue
		}
	}
	return nil
}

// Write fault resource(NPUs,nodes) into cache.
func writeFaultResourceInfInCache(ssn *framework.Session, npus []FaultNPUsOnNode, nodes []FaultNodeState) error {
	// 1.Write fault NPU cards into cache.
	var cardMap = make(map[string]FaultNPUsOnNode, 1)
	for _, card := range npus {
		tmp := card
		cardMap[card.NodeName] = tmp
	}
	ReSchedulerCache[CmCardKind] = cardMap
	// 2.Write fault NPU nodes into cache.
	var nodeMap = make(map[string]FaultNodeState, 1)
	for _, nodeState := range nodes {
		tmp := nodeState
		nodeMap[nodeState.NodeName] = tmp
	}
	ReSchedulerCache[CmNodeKind] = nodeMap
	// 3.Writes the chip involved in the failed node to the cache.
	if writeRankIdsErr := writeFaultNodeRankIdsByJobInCache(ssn); writeRankIdsErr != nil {
		klog.V(logDebugLev).Infof("writeFaultNodeRankIdsByJobInCache %v.", writeRankIdsErr)
		return writeRankIdsErr
	}

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

	if writeErr := writeFaultResourceInfInCache(ssn, faultNPUs, faultNodes); writeErr != nil {
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

func getJobPodsInfoFromK8s(ssn *framework.Session, dJob *api.JobInfo) (map[string]*v1.Pod, error) {
	tmpPods := make(map[string]*v1.Pod, constIntNum3)
	for _, task := range dJob.Tasks {
		pod, err := ssn.KubeClient().CoreV1().Pods(dJob.Namespace).Get(context.TODO(),
			task.Pod.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		tmpPods[pod.Name] = pod
	}
	return tmpPods, nil
}
