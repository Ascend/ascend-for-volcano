/*
Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package ascendtest is using for HuaWei Ascend pin scheduling test.

*/
package ascendtest

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/tools/record"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/cache"
	"volcano.sh/volcano/pkg/scheduler/framework"
	"volcano.sh/volcano/pkg/scheduler/util"
)

// AddResource add resource into resourceList
func AddResource(resourceList v1.ResourceList, name v1.ResourceName, need string) {
	resourceList[name] = resource.MustParse(need)
}

// PrintError print Error for test
func PrintError(format string, args ...interface{}) {
	fmt.Printf("ERROR:"+format+"\n", args...)
}

// FakeNormalSSN fake normal test ssn.
func FakeNormalSSN() *framework.Session {
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

		Recorder: record.NewFakeRecorder(constIntNum3),
	}

	nodes := FakeNormalTestNodes(constIntNum3)
	for _, node := range nodes {
		schedulerCache.AddNode(node.Node)
	}
	jobInf := FakeNormalTestJob("pg1", constIntNum3)
	for _, task := range jobInf.Tasks {
		schedulerCache.AddPod(task.Pod)
	}
	addTestJobPodGroup(jobInf)

	snapshot := schedulerCache.Snapshot()
	ssn := &framework.Session{
		UID:            uuid.NewUUID(),
		Jobs:           snapshot.Jobs,
		Nodes:          snapshot.Nodes,
		RevocableNodes: snapshot.RevocableNodes,
		Queues:         snapshot.Queues,
		NamespaceInfo:  snapshot.NamespaceInfo,
	}

	return ssn
}
