/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package test is using for HuaWei Ascend pin fault rescheduling.

*/
package test

import (
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/tools/record"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/cache"
	"volcano.sh/volcano/pkg/scheduler/framework"
	"volcano.sh/volcano/pkg/scheduler/util"
)

// FakeSSNReSchedule fake normal test ssn 2jobs.
func FakeSSNReSchedule() *framework.Session {
	binder := &util.FakeBinder{
		Binds:   map[string]string{},
		Channel: make(chan string),
	}
	schedulerCache := &cache.SchedulerCache{
		Nodes:         make(map[string]*api.NodeInfo),
		Jobs:          make(map[api.JobID]*api.JobInfo),
		Queues:        make(map[api.QueueID]*api.QueueInfo),
		Binder:        binder,
		StatusUpdater: &util.FakeStatusUpdater{},
		VolumeBinder:  &util.FakeVolumeBinder{},

		Recorder: record.NewFakeRecorder(npuIndex4),
	}
	nodes := FakeNormalTestNodes(npuIndex4)
	for _, node := range nodes {
		schedulerCache.AddNode(node.Node)
	}
	jobInf0 := FakeNormalTestJob("job0", npuIndex2)
	jobInf1 := FakeNormalTestJob("job1", npuIndex2)
	for _, task := range jobInf0.Tasks {
		schedulerCache.AddPod(task.Pod)
	}
	for _, task := range jobInf1.Tasks {
		schedulerCache.AddPod(task.Pod)
	}
	AddTestJobPodGroup(jobInf0)
	AddTestJobPodGroup(jobInf1)

	snapshot := schedulerCache.Snapshot()
	ssn := &framework.Session{
		UID:            uuid.NewUUID(),
		Jobs:           snapshot.Jobs,
		Nodes:          snapshot.Nodes,
		RevocableNodes: snapshot.RevocableNodes,
		Queues:         snapshot.Queues,
		NamespaceInfo:  snapshot.NamespaceInfo,
	}
	modifyJobInf(jobInf0, jobInf1)
	AddJobIntoFakeSSN(ssn, jobInf0)
	AddJobIntoFakeSSN(ssn, jobInf1)
	return ssn
}

func modifyJobInf(jobInf0 *api.JobInfo, jobInf1 *api.JobInfo) {
	taskName0 := `"vcjob"-"pod0"`
	taskName1 := `"vcjob"-"pod1"`
	annoCards := "Ascend910-0,Ascend910-1,Ascend910-2,Ascend910-3,Ascend910-4,Ascend910-5,Ascend910-6,Ascend910-7"
	jobInf0.PodGroup.Labels = make(map[string]string, npuIndex3)
	jobInf1.PodGroup.Labels = make(map[string]string, npuIndex3)
	jobInf0.PodGroup.Labels["fault-scheduling"] = "grace"
	jobInf1.PodGroup.Labels["fault-scheduling"] = "grace"
	jobInf0.Tasks[api.TaskID(taskName0)].Pod.Annotations = map[string]string{podRankIndex: "0",
		NPU910CardName: annoCards}
	jobInf0.Tasks[api.TaskID(taskName1)].Pod.Annotations = map[string]string{podRankIndex: "1",
		NPU910CardName: annoCards}
	jobInf1.Tasks[api.TaskID(taskName0)].Pod.Annotations =
		map[string]string{podRankIndex: "2", NPU910CardName: annoCards}
	jobInf1.Tasks[api.TaskID(taskName1)].Pod.Annotations =
		map[string]string{podRankIndex: "3", NPU910CardName: annoCards}
	jobInf0.Tasks[api.TaskID(taskName0)].NodeName = "node0"
	jobInf0.Tasks[api.TaskID(taskName1)].NodeName = "node1"
	jobInf1.Tasks[api.TaskID(taskName0)].NodeName = "node2"
	jobInf1.Tasks[api.TaskID(taskName1)].NodeName = "node3"
}
