/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package rescheduling is using for HuaWei Ascend pin fault rescheduling.

*/
package rescheduling

import (
	"strconv"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
)

func addTestTaskIntoReSchedulerCache(tasks ...*api.TaskInfo) {
	var reTask = make(map[api.JobID]ReSchedulerTasks, util.NPUIndex2)
	const tmpNumber = 123456
	if len(tasks) == 0 {
		reTask = map[api.JobID]ReSchedulerTasks{
			"hahaTask": {nil, nil, nil, nil, nil, "", false, tmpNumber}}
		ReSchedulerCache[CmJobKind] = reTask
		return
	}
	for _, task := range tasks {
		reTask[task.Job] = ReSchedulerTasks{
			TaskName:    []string{task.Name},
			NodeNames:   []string{task.NodeName},
			RankIndexes: []string{"1"},
			Time:        nil,
			TaskUseNPUs: nil,
			NameSpace:   ""}
	}

	ReSchedulerCache[CmJobKind] = reTask
}

func addTestNodeIntoReSchedulerCache(nodes ...*api.NodeInfo) {
	var faultNode = make(map[string]FaultNodeState, util.NPUIndex2)
	const testHeartBeat1 = 12345

	if len(nodes) == 0 {
		faultNode["hahaNode"] = FaultNodeState{
			NodeName:          "nil",
			HealthCode:        0,
			UpdateTime:        time.Now().Unix(),
			Heartbeat:         testHeartBeat1,
			HeartbeatInterval: nodeUpdateTime,
		}
		ReSchedulerCache[CmNodeKind] = faultNode
		return
	}

	for _, node := range nodes {
		faultNode[node.Name] = FaultNodeState{
			NodeName:          node.Name,
			HealthCode:        0,
			UpdateTime:        time.Now().Unix(),
			Heartbeat:         testHeartBeat1,
			HeartbeatInterval: nodeUpdateTime,
		}
	}

	ReSchedulerCache[CmNodeKind] = faultNode
}

func addTestJobRankIndexIntoReschedulerCache(job *api.JobInfo) {
	var reRankIDs = make(map[api.JobID]FaultRankIDRecordJobCMData, util.NPUIndex2)
	const FaultRankIDs = "12345"
	if job == nil {
		reRankIDs = map[api.JobID]FaultRankIDRecordJobCMData{
			"hahaTask": {"", "", nil, nil, nil, 0}}
		ReSchedulerCache[CmJobRankIds] = reRankIDs
		return
	}
	podNames := make([]string, 0, util.NPUIndex3)
	podUIDs := make([]types.UID, 0, util.NPUIndex3)
	podTime := make([]int64, 0, util.NPUIndex3)
	for _, task := range job.Tasks {
		podNames = append(podNames, task.Pod.Name)
		podUIDs = append(podUIDs, task.Pod.UID)
		podTime = append(podTime, 0)
	}
	reRankIDs[job.UID] = FaultRankIDRecordJobCMData{
		NameSpace:     job.Namespace,
		FaultRankIds:  FaultRankIDs,
		PodsName:      podNames,
		PodsUID:       podUIDs,
		PodsCreatTime: podTime,
		CreatTime:     0,
	}
	ReSchedulerCache[CmJobRankIds] = reRankIDs
}

func addTmpAllocRankIndexIntoReschedulerCache(reRankIDs map[api.JobID]TaskUsedRankIndex) {
	ReSchedulerCache[TmpAllocRankIndexKind] = reRankIDs
}

func initTestReSchedulerCache() {
	ReSchedulerCache = make(map[string]interface{}, util.NPUIndex2)
}

func addTestJobRankIndex(job *api.JobInfo) {
	if job == nil {
		return
	}
	var i = -1
	for _, task := range job.Tasks {
		i++
		if task.Pod.Annotations == nil {
			task.Pod.Annotations = map[string]string{
				podRankIndex: strconv.Itoa(i),
			}
			continue
		}
		task.Pod.Annotations[podRankIndex] = strconv.Itoa(i)
	}
}

func addJobIntoFaultNPUJobStruct(job *api.JobInfo) FaultNPUJob {
	faultJob := FaultNPUJob{
		faultNPUJobBase: faultNPUJobBase{
			jobName:          job.Name,
			namespace:        job.Namespace,
			taskUseRankIndex: make(map[string]string, util.NPUIndex2),
			taskUseNode:      make(map[string]string, util.NPUIndex2),
		},
		taskUseNPUs: make(map[string]string, util.NPUIndex2),
	}

	i := 0
	for _, task := range job.Tasks {
		faultJob.taskUseRankIndex[task.Name] = strconv.Itoa(i)
		i++
		faultJob.taskUseNode[task.Name] = task.NodeName
		faultJob.taskUseNPUs[task.Name] = task.Pod.Annotations[npu800And9000CardName]
	}

	return faultJob
}

func addTestJobIntoReSchedulerCache(job *api.JobInfo) {
	var reJobs = make(map[api.JobID]ReSchedulerTasks, util.NPUIndex2)
	const tmpNumber = 123456
	if job == nil {
		reJobs = map[api.JobID]ReSchedulerTasks{
			"hahaTask": {nil, nil, nil, nil, nil, "", false, tmpNumber}}
		ReSchedulerCache[CmJobKind] = reJobs
		return
	}
	for _, task := range job.Tasks {
		reJobs[task.Job] = ReSchedulerTasks{
			TaskName:    []string{task.Name},
			NodeNames:   []string{task.NodeName},
			RankIndexes: []string{task.Pod.Annotations[podRankIndex]},
			Time:        nil,
			TaskUseNPUs: []string{task.Pod.Annotations[npu800And9000CardName]},
			NameSpace:   "vcjob"}
	}

	ReSchedulerCache[CmJobKind] = reJobs
}

func addTestCardIntoReSchedulerCache(nodeName string, faultCards, netUnhealthyCards []string) {
	var reCard = make(map[string]FaultNPUsOnNode, util.NPUIndex2)

	reCard[nodeName] = FaultNPUsOnNode{
		NodeName:             nodeName,
		FaultNPUs:            faultCards,
		NetworkUnhealthyNPUs: netUnhealthyCards,
		UpdateTime:           time.Now().Unix()}

	ReSchedulerCache[CmCardKind] = reCard
}

func setTestNodeNPUFaultInSSN(ssn *framework.Session, fNPU FaultNPUsOnNode) {
	for nodeName := range ssn.Nodes {
		if fNPU.NodeName == nodeName {
			ssn.Nodes[nodeName].Node.Annotations[faultNPU] = strings.Join(fNPU.FaultNPUs, ",")
			ssn.Nodes[nodeName].Node.Annotations[networkUnhealthyNPU] = strings.Join(fNPU.NetworkUnhealthyNPUs, ",")
			return
		}
	}
}

func setTestNodeSateFaultInSSN(ssn *framework.Session, fNode FaultNodeState) {
	const constNumber10 = 10
	for nodeName := range ssn.Nodes {
		if fNode.NodeName == nodeName {
			ssn.Nodes[nodeName].Node.Annotations[faultNPU] = strconv.FormatInt(fNode.Heartbeat, constNumber10)
			ssn.Nodes[nodeName].Node.Annotations[nodeHeartbeatInterval] = strconv.Itoa(fNode.HeartbeatInterval)
			return
		}
	}
}

func setTestSsnNode(ssn *framework.Session, setData interface{}) {
	switch value := setData.(type) {
	case FaultNPUsOnNode:
		setTestNodeNPUFaultInSSN(ssn, value)
	case FaultNodeState:
		setTestNodeSateFaultInSSN(ssn, value)
	default:
		test.PrintError("not support type:%+v", setData)
	}
}

func fakeCmDataWithoutCmNodeHeartbeat(_ kubernetes.Interface, _, _ string) *v1.ConfigMap {
	var data = make(map[string]string, 1)
	data[CmJobKind] = `{"vcjob/pg1":"{\"NodeNames\":{\"pg1\":\"node0\"},\"RankIndexes\":{\"pg1\":\"0\"},\"Time\":
						{\"pg1\":1634782292,\"TaskUseNPUs\":{\"pg1\":\"Ascend910-1\"},\"NameSpace\":\"vcjob\"}"}`
	data[CmNodeKind] = `{"node0":{"NodeName":"node0","HealthCode":0,"UpdateTime":1634782870,"Heartbeat":1634782117,
						"HeartbeatInterval":5}}`
	data[CmCardKind] = `{"node0":{"NodeName":"node0","FaultNPUs":["Ascend910-2"],"NetworkUnhealthyNPUs":
						["Ascend910-1"],"UpdateTime":1634731710}}`
	var faultNPUConfigMap = &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cmName,
			Namespace: cmNameSpace,
		},
		Data: data,
	}
	return faultNPUConfigMap
}

func fakeErrorHeartbeatCmData(_ kubernetes.Interface, _, _ string) (*v1.ConfigMap, error) {
	faultNPUConfigMap := fakeCmDataWithoutCmNodeHeartbeat(nil, "", "")
	faultNPUConfigMap.Data[CmNodeHeartbeatKind] = `"haha"`
	return faultNPUConfigMap, nil
}

func addTestHeartbeatIntoReSchedulerCache(nodeName string) {
	var reCard = make(map[string]NormalNodeHeartbeat, util.NPUIndex2)
	const tmpNumber = 123456
	reCard[nodeName] = NormalNodeHeartbeat{
		NodeDHeartbeat:      tmpNumber,
		UpdateHeartbeatTime: tmpNumber,
		HeartbeatInterval:   tmpNumber,
		UpdateTime:          time.Now().Unix()}

	ReSchedulerCache[CmNodeHeartbeatKind] = reCard
}
