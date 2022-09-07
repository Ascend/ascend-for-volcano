/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package rescheduling is using for HuaWei Ascend pin fault rescheduling.

*/
package rescheduling

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"k8s.io/apimachinery/pkg/types"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/conf"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

func FakeTestFaultCardUnhealthy(name string, nodeName string, faultType string) *FaultCard {
	return &FaultCard{
		IsFaultCard: true,
		NPUName:     name,
		NodeName:    nodeName,
		FaultType:   faultType,
	}
}

func FakeTestFaultCardHealthy(name string, nodeName string) *FaultCard {
	return &FaultCard{
		IsFaultCard: false,
		NPUName:     name,
		NodeName:    nodeName,
		FaultType:   CardHealthy,
	}
}

func FakeTestFaultCardsUnhealthy(nodeName string) []FaultCard {
	cards := []FaultCard{
		*FakeTestFaultCardUnhealthy("Ascend910-0", nodeName, CardUnhealthy),
		*FakeTestFaultCardHealthy("Ascend910-1", nodeName),
		*FakeTestFaultCardHealthy("Ascend910-2", nodeName),
		*FakeTestFaultCardHealthy("Ascend910-3", nodeName),
		*FakeTestFaultCardHealthy("Ascend910-4", nodeName),
		*FakeTestFaultCardHealthy("Ascend910-5", nodeName),
		*FakeTestFaultCardHealthy("Ascend910-6", nodeName),
		*FakeTestFaultCardHealthy("Ascend910-7", nodeName),
	}
	return cards
}

func FakeTestFaultCardsHealthy(nodeName string) []FaultCard {
	cards := []FaultCard{
		*FakeTestFaultCardHealthy("Ascend910-0", nodeName),
		*FakeTestFaultCardHealthy("Ascend910-1", nodeName),
		*FakeTestFaultCardHealthy("Ascend910-2", nodeName),
		*FakeTestFaultCardHealthy("Ascend910-3", nodeName),
		*FakeTestFaultCardHealthy("Ascend910-4", nodeName),
		*FakeTestFaultCardHealthy("Ascend910-5", nodeName),
		*FakeTestFaultCardHealthy("Ascend910-6", nodeName),
		*FakeTestFaultCardHealthy("Ascend910-7", nodeName),
	}
	return cards
}

func FakeTestFaultNodeCardUnhealthy(nodeName string) *FaultNode {
	updateTime := int64(11111)
	faultCards := FakeTestFaultCardsUnhealthy(nodeName)
	return &FaultNode{
		NodeName:            nodeName,
		UpdateTime:          updateTime,
		UnhealthyNPU:        []string{"Ascend910-0"},
		NetworkUnhealthyNPU: nil,
		IsFaultNode:         true,
		NodeDEnable:         true,
		NodeHealthState:     NodeCardUnhealthy,
		AllCards: []string{"Ascend910-0", "Ascend910-1", "Ascend910-2", "Ascend910-3", "Ascend910-4",
			"Ascend910-5", "Ascend910-6", "Ascend910-7"},
		FaultCards:          faultCards,
		HeartbeatInterval:   5,
		OldHeartbeatTime:    int64(11110),
		UpdateHeartbeatTime: int64(11110),
	}
}

func FakeTestFaultNodeNodeUnhealthy(nodeName string) *FaultNode {
	updateTime := int64(11111)
	faultCards := FakeTestFaultCardsHealthy(nodeName)
	return &FaultNode{
		NodeName:            nodeName,
		UpdateTime:          updateTime,
		UnhealthyNPU:        nil,
		NetworkUnhealthyNPU: nil,
		IsFaultNode:         true,
		NodeDEnable:         true,
		NodeHealthState:     NodeUnhealthy,
		AllCards: []string{"Ascend910-0", "Ascend910-1", "Ascend910-2", "Ascend910-3", "Ascend910-4",
			"Ascend910-5", "Ascend910-6", "Ascend910-7"},
		FaultCards:          faultCards,
		HeartbeatInterval:   5,
		OldHeartbeatTime:    int64(10000),
		UpdateHeartbeatTime: int64(10000),
	}
}

func FakeTestFaultNodeNodeHealthy(nodeName string) *FaultNode {
	updateTime := int64(11111)
	faultCards := FakeTestFaultCardsHealthy(nodeName)
	return &FaultNode{
		NodeName:            nodeName,
		UpdateTime:          updateTime,
		UnhealthyNPU:        nil,
		NetworkUnhealthyNPU: nil,
		IsFaultNode:         false,
		NodeDEnable:         true,
		NodeHealthState:     NodeHealthy,
		AllCards: []string{"Ascend910-0", "Ascend910-1", "Ascend910-2", "Ascend910-3", "Ascend910-4",
			"Ascend910-5", "Ascend910-6", "Ascend910-7"},
		FaultCards:          faultCards,
		HeartbeatInterval:   5,
		OldHeartbeatTime:    int64(11110),
		UpdateHeartbeatTime: int64(11110),
	}
}

func FakeTestFaultNodeNodeHealthyOneCard(nodeName string) *FaultNode {
	updateTime := int64(11111)
	faultCards := []FaultCard{*FakeTestFaultCardHealthy("Ascend910-0", nodeName)}
	return &FaultNode{
		NodeName:            nodeName,
		UpdateTime:          updateTime,
		UnhealthyNPU:        nil,
		NetworkUnhealthyNPU: nil,
		IsFaultNode:         false,
		NodeDEnable:         true,
		NodeHealthState:     NodeHealthy,
		AllCards:            []string{"Ascend910-0"},
		FaultCards:          faultCards,
		HeartbeatInterval:   5,
		OldHeartbeatTime:    int64(11110),
		UpdateHeartbeatTime: int64(11110),
	}
}

func FakeTestFaultTaskFault(name string, namespace string,
	nodeName string, nodeRankIndex string, podUID types.UID) *FaultTask {
	return &FaultTask{
		IsFaultTask:   true,
		TaskName:      name,
		TaskNamespace: namespace,
		NodeName:      nodeName,
		NodeRankIndex: nodeRankIndex,
		UseCardName:   []string{"Ascend910-0", "Ascend910-1", "Ascend910-2", "Ascend910-3"},
		PodCreateTime: int64(10000),
		PodUID:        podUID,
	}
}

func FakeTestFaultTaskHealth(name string, namespace string, nodeName string,
	nodeRankIndex string, podUID types.UID) *FaultTask {
	return &FaultTask{
		IsFaultTask:   false,
		TaskName:      name,
		TaskNamespace: namespace,
		NodeName:      nodeName,
		NodeRankIndex: nodeRankIndex,
		UseCardName:   []string{"Ascend910-0", "Ascend910-1", "Ascend910-2", "Ascend910-3"},
		PodCreateTime: int64(10000),
		PodUID:        podUID,
	}
}

func FakeTestFaultJob(
	nodeNames []string, jobRankIds []string, faultTasks []FaultTask, jobName string, nameSpace string) *FaultJob {
	updateTime := int64(11111)
	return &FaultJob{
		ReScheduleKey:       JobGraceRescheduleLabelValue,
		IsFaultJob:          true,
		JobName:             jobName,
		JobNamespace:        nameSpace,
		JobRankIds:          jobRankIds, // []string{"0", "10", "28"},
		NodeNames:           nodeNames,  //[]string{"ubuntu", "ubuntu1", "ubuntu3"},
		FaultTasks:          faultTasks,
		UpdateTime:          updateTime,
		JobRankIdCreateTime: int64(10000),
	}
}

func FakeReSchedulerCache() *DealReSchedulerCache {
	nodeNames := []string{"ubuntu1", "ubuntu2", "ubuntu3", "ubuntu4"}
	nodeRankIds := []string{"0", "1", "2", "3"}
	taskNames := []string{"task1", "task2", "task3", "task4"}
	jobRankIds := []string{"0", "10", "18", "27"}
	nameSpace := "vcjob"
	jobName := "job1"
	faultTasks := []FaultTask{
		*FakeTestFaultTaskHealth(taskNames[0], nameSpace, nodeNames[0], nodeRankIds[0], "pod1"),
		*FakeTestFaultTaskFault(taskNames[1], nameSpace, nodeNames[1], nodeRankIds[1], "pod2"),
		*FakeTestFaultTaskHealth(taskNames[2], nameSpace, nodeNames[2], nodeRankIds[2], "pod3"),
		*FakeTestFaultTaskFault(taskNames[3], nameSpace, nodeNames[3], nodeRankIds[3], "pod4"),
	}
	return &DealReSchedulerCache{
		FaultNodes: []FaultNode{
			*FakeTestFaultNodeNodeHealthy(nodeNames[0]),
			*FakeTestFaultNodeCardUnhealthy(nodeNames[1]),
			*FakeTestFaultNodeNodeHealthy(nodeNames[2]),
			*FakeTestFaultNodeNodeUnhealthy(nodeNames[3]),
		},
		FaultJobs: []FaultJob{
			*FakeTestFaultJob(nodeNames, jobRankIds, faultTasks, jobName, nameSpace),
		},
		DealReSchedulerConfigmap: nil,
	}
}

func FakeNPUNodeNilDeviceInfo(name string) *plugin.NPUNode {
	nodeInfo := test.FakeNormalTestNode(name)
	return &plugin.NPUNode{
		Name:       name,
		Capability: nodeInfo.Capability.ScalarResources,
		Allocate:   nodeInfo.Allocatable.ScalarResources,
		Idle:       nodeInfo.Idle.ScalarResources,
		Annotation: nodeInfo.Node.Annotations,
		Label:      nodeInfo.Node.Labels,
	}
}

func FakeNPUNodeWithDeviceInfo(name string) *plugin.NPUNode {
	anno := map[string]string{
		nodeHeartbeatInterval:                            "10",
		nodeHeartbeat:                                    "8",
		util.NPU910CardName:                              "Ascend910-0,Ascend910-1,Ascend910-2",
		util.NPU910CardName + "-" + CardUnhealthy:        "Ascend910-1",
		util.NPU910CardName + "-" + CardNetworkUnhealthy: "Ascend910-2",
	}
	nodeInfo := test.FakeNormalTestNode(name)
	npuNode := &plugin.NPUNode{
		Name:       name,
		Capability: nodeInfo.Capability.ScalarResources,
		Allocate:   nodeInfo.Allocatable.ScalarResources,
		Idle:       nodeInfo.Idle.ScalarResources,
		Annotation: anno,
		Label:      nodeInfo.Node.Labels,
	}
	return npuNode
}

// New ReScheduler test
type newArgs struct {
	env             *plugin.ScheduleEnv
	jobType         string
	cacheFuncBefore func()
	cacheFuncAfter  func()
}

type newTests struct {
	name string
	args newArgs
	want *ReScheduler
}

type FaultReSchedulerGetGraceDeleteTimeFields struct {
	GraceDeleteTime      int64
	DealReSchedulerCache *DealReSchedulerCache
}

type FaultReSchedulerGetGraceDeleteTimeArgs struct {
	conf           []conf.Configuration
	cacheFunBefore func()
	cacheFunAfter  func()
}

type FaultReSchedulerGetGraceDeleteTimeTests struct {
	name    string
	fields  FaultReSchedulerGetGraceDeleteTimeFields
	args    FaultReSchedulerGetGraceDeleteTimeArgs
	want    int64
	wantErr bool
}

func fakeSchedulerConfiguration(_ string, _ []conf.Configuration) *conf.Configuration {
	schedulerConf := &conf.Configuration{
		Name:      "util.CMInitParamKey",
		Arguments: map[string]string{GraceOverTimeKey: "800"},
	}
	return schedulerConf
}

func FakeSchedulerConfGraceOverTime() []conf.Configuration {
	schedulerConf := []conf.Configuration{
		{
			Name:      "util.CMInitParamKey",
			Arguments: map[string]string{GraceOverTimeKey: "800"},
		},
	}
	return schedulerConf
}

func buildFaultReSchedulerGetGraceDeleteTestCases() []FaultReSchedulerGetGraceDeleteTimeTests {
	var tmpPatche *gomonkey.Patches
	conf2 := FakeSchedulerConfGraceOverTime()
	testCases := []FaultReSchedulerGetGraceDeleteTimeTests{
		{
			name: "01-test FaultReSchedulerGetGraceDelete-no config",
			fields: FaultReSchedulerGetGraceDeleteTimeFields{
				GraceDeleteTime:      900,
				DealReSchedulerCache: FakeReSchedulerCache(),
			},
			args: FaultReSchedulerGetGraceDeleteTimeArgs{
				conf: nil,
				cacheFunBefore: func() {
					tmpPatche = gomonkey.ApplyFunc(util.GetConfigFromSchedulerConfigMap, fakeSchedulerConfiguration)
				},
				cacheFunAfter: func() {
					if tmpPatche != nil {
						tmpPatche.Reset()
					}
				},
			},
			want:    900,
			wantErr: true,
		},
		{
			name: "02-test FaultReSchedulerGetGraceDelete-succeed",
			fields: FaultReSchedulerGetGraceDeleteTimeFields{
				GraceDeleteTime:      900,
				DealReSchedulerCache: FakeReSchedulerCache(),
			},
			args: FaultReSchedulerGetGraceDeleteTimeArgs{
				conf: conf2,
				cacheFunBefore: func() {
					tmpPatche = gomonkey.ApplyFunc(util.GetConfigFromSchedulerConfigMap, fakeSchedulerConfiguration)
				},
				cacheFunAfter: func() {
					if tmpPatche != nil {
						tmpPatche.Reset()
					}
				},
			},
			want:    900,
			wantErr: false,
		},
	}
	return testCases
}

func TestFaultReScheduler_GetGraceDeleteTime(t *testing.T) {
	tests := buildFaultReSchedulerGetGraceDeleteTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reScheduler := &ReScheduler{
				GraceDeleteTime:      tt.fields.GraceDeleteTime,
				DealReSchedulerCache: tt.fields.DealReSchedulerCache,
			}
			got, err := reScheduler.GetGraceDeleteTime(tt.args.conf)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetGraceDeleteTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetGraceDeleteTime() got = %v, want %v", got, tt.want)
			}
		})
	}
}

type ReSchedulerAddFaultJobWithSessionArgs struct {
	jobs        map[api.JobID]*api.JobInfo
	cardName    string
	cardPreName string
}

type ReSchedulerAddFaultJobWithSessionTests struct {
	name    string
	fields  *ReScheduler
	args    ReSchedulerAddFaultJobWithSessionArgs
	wantErr bool
}

func fakeCacheNoneFJobReSchedulerAddFaultJobWithSession() *DealReSchedulerCache {
	reCache := DealReSchedulerCache{
		FaultNodes: []FaultNode{
			{
				NodeName:            "node0",
				UpdateTime:          123456,
				UnhealthyNPU:        []string{"Ascend910-0"},
				NetworkUnhealthyNPU: nil,
				IsFaultNode:         true,
				NodeDEnable:         true,
				NodeHealthState:     NodeCardUnhealthy,
				AllCards: []string{"Ascend910-0", "Ascend910-1", "Ascend910-2", "Ascend910-3",
					"Ascend910-4", "Ascend910-5", "Ascend910-6", "Ascend910-7"},
				FaultCards: []FaultCard{
					*FakeTestFaultCardUnhealthy("Ascend910-0", "node0", NodeCardUnhealthy),
					*FakeTestFaultCardHealthy("Ascend910-1", "node0"),
					*FakeTestFaultCardHealthy("Ascend910-2", "node0"),
					*FakeTestFaultCardHealthy("Ascend910-3", "node0"),
					*FakeTestFaultCardHealthy("Ascend910-4", "node0"),
					*FakeTestFaultCardHealthy("Ascend910-5", "node0"),
					*FakeTestFaultCardHealthy("Ascend910-6", "node0"),
					*FakeTestFaultCardHealthy("Ascend910-7", "node0"),
				},
				HeartbeatInterval:   5,
				OldHeartbeatTime:    0,
				UpdateHeartbeatTime: 0,
			},
		},
		FaultJobs:                nil,
		DealReSchedulerConfigmap: nil,
	}
	return &reCache
}

func fakeFaultTask2P(ns string, name string, node string, job string, index string) FaultTask {
	fTask := FaultTask{
		IsFaultTask:   true,
		TaskUID:       api.TaskID("\"" + ns + "\"" + "\"" + name + "\""),
		TaskName:      name,
		TaskNamespace: ns,
		NodeName:      node,
		JobName:       job,
		NodeRankIndex: index,
		UseCardName:   []string{"Ascend910-0", "Ascend910-1"},
		PodCreateTime: 123455,
		PodUID:        types.UID("\"" + ns + "\"" + "\"" + name + "\""),
	}
	return fTask
}

func fakeCacheWithFJobReSchedulerAddFaultJobWithSession() *DealReSchedulerCache {
	reCache := DealReSchedulerCache{
		FaultNodes: []FaultNode{
			{
				NodeName:            "node0",
				UpdateTime:          123456,
				UnhealthyNPU:        []string{"Ascend910-0"},
				NetworkUnhealthyNPU: nil,
				IsFaultNode:         true,
				NodeDEnable:         true,
				NodeHealthState:     NodeCardUnhealthy,
				AllCards: []string{"Ascend910-0", "Ascend910-1", "Ascend910-2", "Ascend910-3",
					"Ascend910-4", "Ascend910-5", "Ascend910-6", "Ascend910-7"},
				FaultCards: []FaultCard{
					*FakeTestFaultCardUnhealthy("Ascend910-0", "node0", NodeCardUnhealthy),
					*FakeTestFaultCardHealthy("Ascend910-1", "node0"),
					*FakeTestFaultCardHealthy("Ascend910-2", "node0"),
					*FakeTestFaultCardHealthy("Ascend910-3", "node0"),
					*FakeTestFaultCardHealthy("Ascend910-4", "node0"),
					*FakeTestFaultCardHealthy("Ascend910-5", "node0"),
					*FakeTestFaultCardHealthy("Ascend910-6", "node0"),
					*FakeTestFaultCardHealthy("Ascend910-7", "node0"),
				},
				HeartbeatInterval:   5,
				OldHeartbeatTime:    0,
				UpdateHeartbeatTime: 0,
			},
		},
		FaultJobs: []FaultJob{
			{
				ReScheduleKey: "grace",
				IsFaultJob:    true,
				IsInSession:   true,
				JobName:       "job0",
				JobUID:        "vcjob/job0",
				JobNamespace:  "test",
				JobRankIds:    []string{"0", "1", "8", "9"},
				NodeNames:     []string{"node0", "node1"},
				FaultTasks: []FaultTask{
					fakeFaultTask2P("vcjob", "pod0", "node0", "job0", "0"),
					fakeFaultTask2P("vcjob", "pod1", "node1", "job0", "1"),
				},
				UpdateTime:          123455,
				JobRankIdCreateTime: 123454,
			},
		},
		DealReSchedulerConfigmap: nil,
	}
	return &reCache
}

func ReAddFaultJobWithSessionModifyJobInfo(jobInfos map[api.JobID]*api.JobInfo) map[api.JobID]*api.JobInfo {
	jobInfos["vcjob/job0"].PodGroup.Labels = map[string]string{JobRescheduleLabelKey: "grace"}
	jobInfos["vcjob/job1"].PodGroup.Labels = map[string]string{JobRescheduleLabelKey: "grace"}
	jobInfos["vcjob/job0"].Tasks["\"vcjob\"-\"pod0\""].Pod.Annotations =
		map[string]string{podRankIndex: "0", util.NPU910CardName: "Ascend910-0,Ascend910-1"}
	jobInfos["vcjob/job0"].Tasks["\"vcjob\"-\"pod1\""].Pod.Annotations =
		map[string]string{podRankIndex: "1", util.NPU910CardName: "Ascend910-0,Ascend910-1"}
	jobInfos["vcjob/job1"].Tasks["\"vcjob\"-\"pod0\""].Pod.Annotations =
		map[string]string{podRankIndex: "2", util.NPU910CardName: "Ascend910-0,Ascend910-1"}
	jobInfos["vcjob/job1"].Tasks["\"vcjob\"-\"pod1\""].Pod.Annotations =
		map[string]string{podRankIndex: "3", util.NPU910CardName: "Ascend910-0,Ascend910-1"}
	jobInfos["vcjob/job1"].Tasks["\"vcjob\"-\"pod0\""].NodeName = "node3"
	jobInfos["vcjob/job1"].Tasks["\"vcjob\"-\"pod1\""].NodeName = "node4"
	return jobInfos
}

func ReCreateNPUTask910(name, namespace string, reqResourceNum int) util.NPUTask {
	return util.NPUTask{
		TaskName:   namespace + "/" + name,
		ReqNPUName: "huawei.com/Ascend910",
		ReqNPUNum:  reqResourceNum,
		Selector:   nil,
	}
}

func AddNPUTaskToNPUJob(npuJob plugin.SchedulerJob, taskName, taskNamespace string, reqNPUNum int) plugin.SchedulerJob {
	task := ReCreateNPUTask910(taskName, taskNamespace, reqNPUNum)
	npuJob.Tasks[taskName] = task
	npuJob.ReqNPUNum += reqNPUNum
	return npuJob
}

func ReCreateSchedulerJob910(namespace string, UID api.JobID) plugin.SchedulerJob {
	sJob := plugin.SchedulerJob{
		SchedulerJobAttr: util.SchedulerJobAttr{
			ComJob: util.ComJob{
				JobName:   UID,
				NameSpace: namespace,
				Selector:  nil,
				Label:     nil,
			},
			NPUJob: &util.NPUJob{
				ReqNPUName: "huawei.com/Ascend910",
				ReqNPUNum:  0,
				Tasks:      make(map[string]util.NPUTask, util.MapInitNum),
			},
		},
	}
	return sJob
}

func ReNewRescheduler(graceTime int64) *ReScheduler {
	fakeReScheduler := ReScheduler{
		GraceDeleteTime:      graceTime,
		Level:                "",
		Jobs:                 nil,
		Nodes:                nil,
		kubeClient:           nil,
		DealReSchedulerCache: nil,
	}
	return &fakeReScheduler
}

func AddReCacheToReScheduler(reScheduler *ReScheduler, reCache *DealReSchedulerCache) {
	reScheduler.DealReSchedulerCache = reCache
}

func AddSchedulerJobToReScheduler(reScheduler *ReScheduler, sJob map[api.JobID]plugin.SchedulerJob) {
	reScheduler.Jobs = sJob
}

func buildReSchedulerAddFaultJobWithSession() []ReSchedulerAddFaultJobWithSessionTests {
	jobInfos1 := map[api.JobID]*api.JobInfo{
		"vcjob/job0": test.FakeNormalTestJob("job0", 2),
		"vcjob/job1": test.FakeNormalTestJob("job1", 2),
	}
	jobInfos1 = ReAddFaultJobWithSessionModifyJobInfo(jobInfos1)
	jobs11 := ReCreateSchedulerJob910("vcjob", "job0")
	jobs11 = AddNPUTaskToNPUJob(jobs11, "task0", "vcjob", 4)
	jobs11 = AddNPUTaskToNPUJob(jobs11, "task1", "vcjob", 4)
	jobs12 := ReCreateSchedulerJob910("vcjob", "job1")
	jobs12 = AddNPUTaskToNPUJob(jobs12, "task3", "vcjob", 4)
	jobs12 = AddNPUTaskToNPUJob(jobs12, "task4", "vcjob", 4)
	jobs1 := map[api.JobID]plugin.SchedulerJob{
		"vcjob/job0": jobs11,
		"vcjob/job1": jobs12,
	}

	reCache1 := fakeCacheNoneFJobReSchedulerAddFaultJobWithSession()
	reScheduler1 := ReNewRescheduler(0)
	AddSchedulerJobToReScheduler(reScheduler1, jobs1)
	AddReCacheToReScheduler(reScheduler1, reCache1)

	reCache2 := fakeCacheWithFJobReSchedulerAddFaultJobWithSession()
	reScheduler2 := ReNewRescheduler(900)
	AddSchedulerJobToReScheduler(reScheduler2, jobs1)
	AddReCacheToReScheduler(reScheduler2, reCache2)
	test1 := ReSchedulerAddFaultJobWithSessionTests{
		name:   "01-AddFaultJobWithSession()-GraceDeleteTime 0",
		fields: reScheduler1,
		args: ReSchedulerAddFaultJobWithSessionArgs{
			jobs:        jobInfos1,
			cardName:    "huawei.com/Ascend910",
			cardPreName: "Ascend910-",
		},
		wantErr: false,
	}
	test2 := ReSchedulerAddFaultJobWithSessionTests{
		name:   "02-AddFaultJobWithSession()-GraceDeleteTime 900",
		fields: reScheduler2,
		args: ReSchedulerAddFaultJobWithSessionArgs{
			jobs:        jobInfos1,
			cardName:    "",
			cardPreName: "",
		},
		wantErr: false,
	}
	tests := []ReSchedulerAddFaultJobWithSessionTests{
		test1,
		test2,
	}
	return tests
}

func TestReScheduler_AddFaultJobWithSession(t *testing.T) {
	tests := buildReSchedulerAddFaultJobWithSession()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reScheduler := tt.fields
			if err := reScheduler.AddFaultJobWithSession(tt.args.jobs, tt.args.cardName, tt.args.cardPreName);
			(err != nil) != tt.wantErr {
				t.Errorf("AddFaultJobWithSession() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
