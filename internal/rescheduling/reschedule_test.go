/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package rescheduling is using for HuaWei Ascend pin fault rescheduling.

*/
package rescheduling

import (
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/conf"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

const (
	heartbeatInterval = 5
	fakeTime          = 123455
	fakeTime2         = 11111
	createTime        = 10000
	heartbeatTime     = 11110
	graceDeleteTime   = 900
	zero              = 0
	one               = 1
	two               = 2
	three             = 3
)

func fakeTestFaultCardUnhealthy(name string, nodeName string, faultType string) *FaultCard {
	return &FaultCard{
		IsFaultCard: true,
		NPUName:     name,
		NodeName:    nodeName,
		FaultType:   faultType,
	}
}

func fakeTestFaultCardHealthy(name string, nodeName string) *FaultCard {
	return &FaultCard{
		IsFaultCard: false,
		NPUName:     name,
		NodeName:    nodeName,
		FaultType:   CardHealthy,
	}
}

func fakeTestFaultCardsUnhealthy(nodeName string, isUnHealth bool) []FaultCard {
	var card0 *FaultCard
	if !isUnHealth {
		card0 = fakeTestFaultCardUnhealthy("Ascend910-0", nodeName, CardUnhealthy)
	} else {
		card0 = fakeTestFaultCardHealthy("Ascend910-0", nodeName)
	}
	cards := []FaultCard{
		*card0,
		*fakeTestFaultCardHealthy("Ascend910-1", nodeName),
		*fakeTestFaultCardHealthy("Ascend910-2", nodeName),
		*fakeTestFaultCardHealthy("Ascend910-3", nodeName),
		*fakeTestFaultCardHealthy("Ascend910-4", nodeName),
		*fakeTestFaultCardHealthy("Ascend910-5", nodeName),
		*fakeTestFaultCardHealthy("Ascend910-6", nodeName),
		*fakeTestFaultCardHealthy("Ascend910-7", nodeName),
	}
	return cards
}

func fakeTestFaultNodeCardUnhealthy(nodeName string, allCard []string) *FaultNode {
	updateTime := int64(fakeTime2)
	faultCards := fakeTestFaultCardsUnhealthy(nodeName, true)
	return &FaultNode{
		NodeName:            nodeName,
		UpdateTime:          updateTime,
		UnhealthyNPU:        []string{"Ascend910-0"},
		NetworkUnhealthyNPU: nil,
		IsFaultNode:         true,
		NodeDEnable:         true,
		NodeHealthState:     NodeCardUnhealthy,
		AllCards:            allCard,
		FaultCards:          faultCards,
		HeartbeatInterval:   heartbeatInterval,
		OldHeartbeatTime:    int64(heartbeatTime),
		UpdateHeartbeatTime: int64(heartbeatTime),
	}
}

func fakeTestFaultNodeNodeUnhealthy(nodeName string) *FaultNode {
	updateTime := int64(fakeTime2)
	faultCards := fakeTestFaultCardsUnhealthy(nodeName, false)
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
		HeartbeatInterval:   heartbeatInterval,
		OldHeartbeatTime:    int64(createTime),
		UpdateHeartbeatTime: int64(createTime),
	}
}

func fakeTestFaultNodeNodeHealthy(nodeName string) *FaultNode {
	updateTime := int64(fakeTime2)
	faultCards := fakeTestFaultCardsUnhealthy(nodeName, false)
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
		HeartbeatInterval:   heartbeatInterval,
		OldHeartbeatTime:    int64(heartbeatTime),
		UpdateHeartbeatTime: int64(heartbeatTime),
	}
}

func fakeTestFaultNodeNodeHealthyOneCard(nodeName string) *FaultNode {
	updateTime := int64(fakeTime2)
	faultCards := []FaultCard{*fakeTestFaultCardHealthy("Ascend910-0", nodeName)}
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
		HeartbeatInterval:   heartbeatInterval,
		OldHeartbeatTime:    int64(heartbeatTime),
		UpdateHeartbeatTime: int64(heartbeatTime),
	}
}

func fakeTestFaultTaskFault(name string, namespace string,
	nodeName string, nodeRankIndex string, podUID types.UID) *FaultTask {
	return &FaultTask{
		IsFaultTask:   true,
		TaskName:      name,
		TaskNamespace: namespace,
		NodeName:      nodeName,
		NodeRankIndex: nodeRankIndex,
		UseCardName:   []string{"Ascend910-0", "Ascend910-1", "Ascend910-2", "Ascend910-3"},
		PodCreateTime: int64(createTime),
		PodUID:        podUID,
	}
}

func fakeTestFaultTaskHealth(name string, namespace string, nodeName string,
	nodeRankIndex string, podUID types.UID) *FaultTask {
	return &FaultTask{
		IsFaultTask:   false,
		TaskName:      name,
		TaskNamespace: namespace,
		NodeName:      nodeName,
		NodeRankIndex: nodeRankIndex,
		UseCardName:   []string{"Ascend910-0", "Ascend910-1", "Ascend910-2", "Ascend910-3"},
		PodCreateTime: int64(createTime),
		PodUID:        podUID,
	}
}

func fakeTestFaultJob(
	nodeNames []string, jobRankIds []string, faultTasks []FaultTask, jobName string, nameSpace string) *FaultJob {
	updateTime := int64(fakeTime2)
	return &FaultJob{
		ReScheduleKey:       JobGraceRescheduleLabelValue,
		IsFaultJob:          true,
		JobName:             jobName,
		JobUID:              api.JobID(nameSpace + `/` + jobName),
		JobNamespace:        nameSpace,
		JobRankIds:          jobRankIds,
		NodeNames:           nodeNames,
		FaultTasks:          faultTasks,
		UpdateTime:          updateTime,
		JobRankIdCreateTime: int64(createTime),
	}
}

func fakeReSchedulerCache() *DealReSchedulerCache {
	nodeNames := []string{"ubuntu1", "ubuntu2", "ubuntu3", "ubuntu4"}
	nodeRankIds := []string{"0", "1", "2", "3"}
	taskNames := []string{"task1", "task2", "task3", "task4"}
	jobRankIds := []string{"0", "10", "18", "27"}
	allCard := []string{"Ascend910-0", "Ascend910-1", "Ascend910-2", "Ascend910-3", "Ascend910-4",
		"Ascend910-5", "Ascend910-6", "Ascend910-7"}
	nameSpace := "vcjob"
	jobName := "job1"
	faultTasks := []FaultTask{
		*fakeTestFaultTaskHealth(taskNames[zero], nameSpace, nodeNames[zero], nodeRankIds[zero], "pod1"),
		*fakeTestFaultTaskFault(taskNames[one], nameSpace, nodeNames[one], nodeRankIds[one], "pod2"),
		*fakeTestFaultTaskHealth(taskNames[two], nameSpace, nodeNames[two], nodeRankIds[two], "pod3"),
		*fakeTestFaultTaskFault(taskNames[three], nameSpace, nodeNames[three], nodeRankIds[three], "pod4"),
	}
	return &DealReSchedulerCache{
		FaultNodes: []FaultNode{
			*fakeTestFaultNodeNodeHealthy(nodeNames[zero]),
			*fakeTestFaultNodeCardUnhealthy(nodeNames[one], allCard),
			*fakeTestFaultNodeNodeHealthy(nodeNames[two]),
			*fakeTestFaultNodeNodeUnhealthy(nodeNames[three]),
		},
		FaultJobs: []FaultJob{
			*fakeTestFaultJob(nodeNames, jobRankIds, faultTasks, jobName, nameSpace),
		},
		DealReSchedulerConfigmap: nil,
	}
}

func fakeNPUNodeNilDeviceInfo(name string) *plugin.NPUNode {
	nodeInfo := test.FakeNormalTestNode(name)
	return &plugin.NPUNode{
		CommonNode: plugin.CommonNode{
			Name:       name,
			Capability: nodeInfo.Capability.ScalarResources,
			Allocate:   nodeInfo.Allocatable.ScalarResources,
			Idle:       nodeInfo.Idle.ScalarResources,
			Annotation: nodeInfo.Node.Annotations,
			Label:      nodeInfo.Node.Labels,
		},
	}
}

func fakeNPUNodeWithDeviceInfo(name string) *plugin.NPUNode {
	anno := map[string]string{
		nodeHeartbeatInterval:                            "10",
		nodeHeartbeat:                                    "8",
		util.NPU910CardName:                              "Ascend910-0,Ascend910-1,Ascend910-2",
		util.NPU910CardName + "-" + CardUnhealthy:        "Ascend910-1",
		util.NPU910CardName + "-" + CardNetworkUnhealthy: "Ascend910-2",
	}
	nodeInfo := test.FakeNormalTestNode(name)
	npuNode := &plugin.NPUNode{
		CommonNode: plugin.CommonNode{
			Name:       name,
			Capability: nodeInfo.Capability.ScalarResources,
			Allocate:   nodeInfo.Allocatable.ScalarResources,
			Idle:       nodeInfo.Idle.ScalarResources,
			Annotation: anno,
			Label:      nodeInfo.Node.Labels,
		},
	}
	return npuNode
}

type FaultReSchedulerGetGraceDeleteTimeFields struct {
	DealReSchedulerCache *DealReSchedulerCache
	GraceDeleteTime      int64
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

func fakeSchedulerConfGraceOverTime() []conf.Configuration {
	schedulerConf := []conf.Configuration{
		{
			Name:      "util.CMInitParamKey",
			Arguments: map[string]string{GraceOverTimeKey: "800"},
		},
	}
	return schedulerConf
}

func buildFaultReSchedulerGetGraceDeleteArgs(conf2 []conf.Configuration,
	tmpPatche *gomonkey.Patches) FaultReSchedulerGetGraceDeleteTimeArgs {
	args := FaultReSchedulerGetGraceDeleteTimeArgs{
		conf: conf2,
		cacheFunBefore: func() {
			tmpPatche = gomonkey.ApplyFunc(util.GetConfigFromSchedulerConfigMap, fakeSchedulerConfiguration)
		},
		cacheFunAfter: func() {
			if tmpPatche != nil {
				tmpPatche.Reset()
			}
		},
	}
	return args
}

func buildFaultReSchedulerGetGraceDeleteTestCases() []FaultReSchedulerGetGraceDeleteTimeTests {
	var tmpPatche *gomonkey.Patches
	conf2 := fakeSchedulerConfGraceOverTime()
	testCases := []FaultReSchedulerGetGraceDeleteTimeTests{
		{
			name: "01-test FaultReSchedulerGetGraceDelete-no config",
			fields: FaultReSchedulerGetGraceDeleteTimeFields{
				GraceDeleteTime:      graceDeleteTime,
				DealReSchedulerCache: fakeReSchedulerCache(),
			},
			args:    buildFaultReSchedulerGetGraceDeleteArgs(nil, tmpPatche),
			want:    graceDeleteTime,
			wantErr: true,
		},
		{
			name: "02-test FaultReSchedulerGetGraceDelete-succeed",
			fields: FaultReSchedulerGetGraceDeleteTimeFields{
				GraceDeleteTime:      graceDeleteTime,
				DealReSchedulerCache: fakeReSchedulerCache(),
			},
			args:    buildFaultReSchedulerGetGraceDeleteArgs(conf2, tmpPatche),
			want:    graceDeleteTime,
			wantErr: false,
		},
	}
	return testCases
}

// TestFaultReSchedulerGetGraceDeleteTime test for grace delete time
func TestFaultReSchedulerGetGraceDeleteTime(t *testing.T) {
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
	fields  *ReScheduler
	name    string
	args    ReSchedulerAddFaultJobWithSessionArgs
	wantErr bool
}

func fakeCacheNoneFJobReSchedulerAddFaultJobWithSession() *DealReSchedulerCache {
	reCache := DealReSchedulerCache{
		FaultNodes: []FaultNode{
			{
				NodeName:            "node0",
				UpdateTime:          fakeTime,
				UnhealthyNPU:        []string{"Ascend910-0"},
				NetworkUnhealthyNPU: nil,
				IsFaultNode:         true,
				NodeDEnable:         true,
				NodeHealthState:     NodeCardUnhealthy,
				AllCards: []string{"Ascend910-0", "Ascend910-1", "Ascend910-2", "Ascend910-3",
					"Ascend910-4", "Ascend910-5", "Ascend910-6", "Ascend910-7"},
				FaultCards: []FaultCard{
					*fakeTestFaultCardUnhealthy("Ascend910-0", "node0", NodeCardUnhealthy),
					*fakeTestFaultCardHealthy("Ascend910-1", "node0"),
					*fakeTestFaultCardHealthy("Ascend910-2", "node0"),
					*fakeTestFaultCardHealthy("Ascend910-3", "node0"),
					*fakeTestFaultCardHealthy("Ascend910-4", "node0"),
					*fakeTestFaultCardHealthy("Ascend910-5", "node0"),
					*fakeTestFaultCardHealthy("Ascend910-6", "node0"),
					*fakeTestFaultCardHealthy("Ascend910-7", "node0"),
				},
				HeartbeatInterval:   heartbeatInterval,
				OldHeartbeatTime:    zero,
				UpdateHeartbeatTime: zero,
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
		TaskUID:       api.TaskID(`"` + ns + `"` + `"` + name + `"`),
		TaskName:      name,
		TaskNamespace: ns,
		NodeName:      node,
		JobName:       job,
		NodeRankIndex: index,
		UseCardName:   []string{"Ascend910-0", "Ascend910-1"},
		PodCreateTime: fakeTime,
		PodUID:        types.UID(`"` + ns + `"` + `"` + name + `"`),
	}
	return fTask
}

func fakeCacheWithFJobReSchedulerAddFaultJobWithSession() *DealReSchedulerCache {
	reCache := fakeCacheNoneFJobReSchedulerAddFaultJobWithSession()
	reCache.FaultJobs = []FaultJob{
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
			UpdateTime:          test.FakeUpdateTime + 1,
			JobRankIdCreateTime: test.FakeUpdateTime,
		},
	}
	return reCache
}

func reAddFaultJobWithSessionModifyJobInfo(jobInfos map[api.JobID]*api.JobInfo) map[api.JobID]*api.JobInfo {
	jobInfos["vcjob/job0"].PodGroup.Labels = map[string]string{JobRescheduleLabelKey: "grace"}
	jobInfos["vcjob/job1"].PodGroup.Labels = map[string]string{JobRescheduleLabelKey: "grace"}
	jobInfos["vcjob/job0"].Tasks[`"vcjob"-"pod1"`].Pod.Annotations =
		map[string]string{podRankIndex: "0", util.NPU910CardName: "Ascend910-0,Ascend910-1"}
	jobInfos["vcjob/job0"].Tasks[`"vcjob"-"pod1"`].Pod.Annotations =
		map[string]string{podRankIndex: "1", util.NPU910CardName: "Ascend910-0,Ascend910-1"}
	jobInfos["vcjob/job1"].Tasks[`"vcjob"-"pod1"`].Pod.Annotations =
		map[string]string{podRankIndex: "2", util.NPU910CardName: "Ascend910-0,Ascend910-1"}
	jobInfos["vcjob/job1"].Tasks[`"vcjob"-"pod1"`].Pod.Annotations =
		map[string]string{podRankIndex: "3", util.NPU910CardName: "Ascend910-0,Ascend910-1"}
	jobInfos["vcjob/job1"].Tasks[`"vcjob"-"pod0"`].NodeName = "node3"
	jobInfos["vcjob/job1"].Tasks[`"vcjob"-"pod1"`].NodeName = "node4"
	return jobInfos
}

func reCreateNPUTask910(name, namespace string, reqResourceNum int) util.NPUTask {
	return util.NPUTask{
		TaskName:   namespace + "/" + name,
		ReqNPUName: "huawei.com/Ascend910",
		ReqNPUNum:  reqResourceNum,
		Selector:   nil,
	}
}

func addNPUTaskToNPUJob(npuJob plugin.SchedulerJob, taskName, taskNamespace string, reqNPUNum int) plugin.SchedulerJob {
	task := reCreateNPUTask910(taskName, taskNamespace, reqNPUNum)
	npuJob.Tasks[taskName] = task
	npuJob.ReqNPUNum += reqNPUNum
	return npuJob
}

func reCreateSchedulerJob910(namespace string, UID api.JobID) plugin.SchedulerJob {
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
				ReqNPUNum:  zero,
				Tasks:      make(map[string]util.NPUTask, util.MapInitNum),
			},
		},
	}
	return sJob
}

func reNewReScheduler(graceTime int64) *ReScheduler {
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

func addReCacheToReScheduler(reScheduler *ReScheduler, reCache *DealReSchedulerCache) {
	reScheduler.DealReSchedulerCache = reCache
}

func addSchedulerJobToReScheduler(reScheduler *ReScheduler, sJob map[api.JobID]plugin.SchedulerJob) {
	reScheduler.Jobs = sJob
}

func buildReSchedulerAddFaultJobWithSession() []ReSchedulerAddFaultJobWithSessionTests {
	jobInfos1 := map[api.JobID]*api.JobInfo{
		"vcjob/job0": test.FakeNormalTestJob("job0", util.NPUIndex2),
		"vcjob/job1": test.FakeNormalTestJob("job1", util.NPUIndex2),
	}
	jobInfos1 = reAddFaultJobWithSessionModifyJobInfo(jobInfos1)
	jobs11 := reCreateSchedulerJob910("vcjob", "job0")
	jobs11 = addNPUTaskToNPUJob(jobs11, "task0", "vcjob", util.NPUIndex4)
	jobs11 = addNPUTaskToNPUJob(jobs11, "task1", "vcjob", util.NPUIndex4)
	jobs12 := reCreateSchedulerJob910("vcjob", "job1")
	jobs12 = addNPUTaskToNPUJob(jobs12, "task3", "vcjob", util.NPUIndex4)
	jobs12 = addNPUTaskToNPUJob(jobs12, "task4", "vcjob", util.NPUIndex4)
	jobs1 := map[api.JobID]plugin.SchedulerJob{
		"vcjob/job0": jobs11,
		"vcjob/job1": jobs12,
	}

	reCache1 := fakeCacheNoneFJobReSchedulerAddFaultJobWithSession()
	reScheduler1 := reNewReScheduler(0)
	addSchedulerJobToReScheduler(reScheduler1, jobs1)
	addReCacheToReScheduler(reScheduler1, reCache1)

	reCache2 := fakeCacheWithFJobReSchedulerAddFaultJobWithSession()
	reScheduler2 := reNewReScheduler(DefaultGraceOverTime)
	addSchedulerJobToReScheduler(reScheduler2, jobs1)
	addReCacheToReScheduler(reScheduler2, reCache2)
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

// TestReSchedulerAddFaultJobWithSession test for add fault job
func TestReSchedulerAddFaultJobWithSession(t *testing.T) {
	tests := buildReSchedulerAddFaultJobWithSession()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reScheduler := tt.fields
			if err := reScheduler.AddFaultJobWithSession(
				tt.args.jobs, tt.args.cardName, tt.args.cardPreName); (err != nil) != tt.wantErr {
				t.Errorf("AddFaultJobWithSession() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type FaultNodeGetUnhealthyCardsFromDeviceInfoArgs struct {
	node     *plugin.NPUNode
	cardName string
}

type FaultNodeGetUnhealthyCardsFromDeviceInfoTests struct {
	fields  *FaultNode
	name    string
	args    FaultNodeGetUnhealthyCardsFromDeviceInfoArgs
	want    []string
	wantErr bool
}

func buildFaultNodeGetUnhealthyCardsFromDeviceInfoTests() []FaultNodeGetUnhealthyCardsFromDeviceInfoTests {
	node2 := fakeNPUNodeWithDeviceInfo("node0")
	test1 := FaultNodeGetUnhealthyCardsFromDeviceInfoTests{
		name:   "01-FaultNodeGetNodeHeartbeatIntervalFromDeviceInfoTests() nil device info",
		fields: fakeTestFaultNodeNodeHealthy("node0"),
		args: FaultNodeGetUnhealthyCardsFromDeviceInfoArgs{
			node:     fakeNPUNodeNilDeviceInfo("node0"),
			cardName: util.NPU910CardName,
		},
		want:    nil,
		wantErr: true,
	}
	test2 := FaultNodeGetUnhealthyCardsFromDeviceInfoTests{
		name:   "02-FaultNodeGetNodeHeartbeatIntervalFromDeviceInfoTests() succeed",
		fields: fakeTestFaultNodeNodeHealthy("node0"),
		args: FaultNodeGetUnhealthyCardsFromDeviceInfoArgs{
			node:     node2,
			cardName: util.NPU910CardName,
		},
		want:    []string{"Ascend910-1"},
		wantErr: false,
	}
	tests := []FaultNodeGetUnhealthyCardsFromDeviceInfoTests{
		test1,
		test2,
	}
	return tests
}

// TestFaultNodeGetUnhealthyCardsFromDeviceInfo test for get unhealthy card
func TestFaultNodeGetUnhealthyCardsFromDeviceInfo(t *testing.T) {
	tests := buildFaultNodeGetUnhealthyCardsFromDeviceInfoTests()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fNode := tt.fields
			got, err := fNode.getUnhealthyCardsFromDeviceInfo(tt.args.node, tt.args.cardName)
			if (err != nil) != tt.wantErr {
				t.Errorf("getUnhealthyCardsFromDeviceInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getUnhealthyCardsFromDeviceInfo() got = %v, want %v", got, tt.want)
			}
		})
	}
}

type FaultNodeGetNetworkUnhealthyCardsFromDeviceInfoArgs struct {
	node     *plugin.NPUNode
	cardName string
}

type GetNetworkUnhealthyCardsFromDeviceInfoTests struct {
	fields  *FaultNode
	name    string
	args    FaultNodeGetNetworkUnhealthyCardsFromDeviceInfoArgs
	want    []string
	wantErr bool
}

func buildFaultNodeGetNetworkUnhealthyCardsFromDeviceInfoTests() []GetNetworkUnhealthyCardsFromDeviceInfoTests {
	node2 := fakeNPUNodeWithDeviceInfo("node0")
	test1 := GetNetworkUnhealthyCardsFromDeviceInfoTests{
		name:   "01-FaultNodeGetNodeHeartbeatIntervalFromDeviceInfoTests() nil device info",
		fields: fakeTestFaultNodeNodeHealthy("node0"),
		args: FaultNodeGetNetworkUnhealthyCardsFromDeviceInfoArgs{
			node:     fakeNPUNodeNilDeviceInfo("node0"),
			cardName: util.NPU910CardName,
		},
		want:    nil,
		wantErr: true,
	}
	test2 := GetNetworkUnhealthyCardsFromDeviceInfoTests{
		name:   "02-FaultNodeGetNodeHeartbeatIntervalFromDeviceInfoTests() succeed",
		fields: fakeTestFaultNodeNodeHealthy("node0"),
		args: FaultNodeGetNetworkUnhealthyCardsFromDeviceInfoArgs{
			node:     node2,
			cardName: util.NPU910CardName,
		},
		want:    []string{"Ascend910-2"},
		wantErr: false,
	}
	tests := []GetNetworkUnhealthyCardsFromDeviceInfoTests{
		test1,
		test2,
	}
	return tests
}

// TestFaultNodeGetNetworkUnhealthyCardsFromDeviceInfo test for get network unhealthy card
func TestFaultNodeGetNetworkUnhealthyCardsFromDeviceInfo(t *testing.T) {
	tests := buildFaultNodeGetNetworkUnhealthyCardsFromDeviceInfoTests()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fNode := tt.fields
			got, err := fNode.getNetworkUnhealthyCardsFromDeviceInfo(tt.args.node, tt.args.cardName)
			if (err != nil) != tt.wantErr {
				t.Errorf("getNetworkUnhealthyCardsFromDeviceInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getNetworkUnhealthyCardsFromDeviceInfo() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func fakeTestFaultNodeCardNetworkUnhealthyOneCard(nodeName string) *FaultNode {
	updateTime := test.FakeUpdateTime + 1
	faultCards := fakeTestFaultCardsUnhealthy(nodeName, true)
	return &FaultNode{
		NodeName:            nodeName,
		UpdateTime:          updateTime,
		UnhealthyNPU:        nil,
		NetworkUnhealthyNPU: []string{"Ascend910-0"},
		IsFaultNode:         true,
		NodeDEnable:         true,
		NodeHealthState:     NodeCardNetworkUnhealthy,
		AllCards:            []string{"Ascend910-0"},
		FaultCards:          faultCards,
		HeartbeatInterval:   test.NPUIndex5,
		OldHeartbeatTime:    test.FakeUpdateTime,
		UpdateHeartbeatTime: test.FakeUpdateTime,
	}
}

type FaultNodeNewFaultCardHandlersArgs struct {
	node     *plugin.NPUNode
	cardName string
}

type FaultNodeNewFaultCardHandlersTests struct {
	fields  *FaultNode
	name    string
	args    FaultNodeNewFaultCardHandlersArgs
	want    []FaultCard
	wantErr bool
}

func faultNodeNewFaultCardHandlersFaultCard(isFault bool, name, nodeName, faultType string) FaultCard {
	return FaultCard{
		IsFaultCard: isFault,
		NPUName:     name,
		NodeName:    nodeName,
		FaultType:   faultType,
	}
}

func buildFaultNodeNewFaultCardHandlers() []FaultNodeNewFaultCardHandlersTests {
	test1 := FaultNodeNewFaultCardHandlersTests{
		name:   "01-newFaultCardHandlers()-NodeUnhealthy",
		fields: fakeTestFaultNodeNodeHealthyOneCard("node0"),
		args: FaultNodeNewFaultCardHandlersArgs{
			node: fakeNPUNodeWithDeviceInfo("node0"),
		},
		want: []FaultCard{
			faultNodeNewFaultCardHandlersFaultCard(false, "Ascend910-0", "node0", "Healthy"),
		},
		wantErr: false,
	}
	test2 := FaultNodeNewFaultCardHandlersTests{
		name:   "02-newFaultCardHandlers()-CardUnhealthy",
		fields: fakeTestFaultNodeCardUnhealthy("node0", []string{"Ascend910-0"}),
		args: FaultNodeNewFaultCardHandlersArgs{
			node: fakeNPUNodeWithDeviceInfo("node0"),
		},
		want: []FaultCard{
			faultNodeNewFaultCardHandlersFaultCard(
				true, "Ascend910-0", "node0", "Unhealthy"),
		},
		wantErr: false,
	}
	test3 := FaultNodeNewFaultCardHandlersTests{
		name:   "03-newFaultCardHandlers()-CardNetworkUnhealthy",
		fields: fakeTestFaultNodeCardNetworkUnhealthyOneCard("node0"),
		args: FaultNodeNewFaultCardHandlersArgs{
			node: fakeNPUNodeWithDeviceInfo("node0"),
		},
		want: []FaultCard{
			faultNodeNewFaultCardHandlersFaultCard(true, "Ascend910-0", "node0", "NetworkUnhealthy"),
		},
		wantErr: false,
	}
	tests := []FaultNodeNewFaultCardHandlersTests{
		test1,
		test2,
		test3,
	}
	return tests
}

// TestFaultNodeNewFaultCardHandlers test for new fault card handlers
func TestFaultNodeNewFaultCardHandlers(t *testing.T) {
	tests := buildFaultNodeNewFaultCardHandlers()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fNode := tt.fields
			got, err := fNode.createFaultCardHandlers(tt.args.node)
			if (err != nil) != tt.wantErr {
				t.Errorf("newFaultCardHandlers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newFaultCardHandlers() got = %v, want %v", got, tt.want)
			}
		})
	}
}

type FaultNodeUpdateFaultNodesAttrArgs struct {
	node *plugin.NPUNode
}

type FaultNodeUpdateFaultNodesAttrTests struct {
	name    string
	fields  FaultNode
	args    FaultNodeUpdateFaultNodesAttrArgs
	wantErr bool
}

func buildFaultNodeUpdateFaultNodesAttrTestCases() []FaultNodeUpdateFaultNodesAttrTests {
	allCard := []string{"Ascend910-0", "Ascend910-1", "Ascend910-2", "Ascend910-3", "Ascend910-4",
		"Ascend910-5", "Ascend910-6", "Ascend910-7"}
	test1 := FaultNodeUpdateFaultNodesAttrTests{
		name:   "node0",
		fields: *fakeTestFaultNodeNodeHealthy("node0"),
		args: FaultNodeUpdateFaultNodesAttrArgs{
			fakeNPUNodeNilDeviceInfo("node0"),
		},
	}
	test2 := FaultNodeUpdateFaultNodesAttrTests{
		name:   "node1",
		fields: *fakeTestFaultNodeNodeUnhealthy("node1"),
		args: FaultNodeUpdateFaultNodesAttrArgs{
			fakeNPUNodeNilDeviceInfo("node1"),
		},
		wantErr: false,
	}
	test3 := FaultNodeUpdateFaultNodesAttrTests{
		name:   "node2",
		fields: *fakeTestFaultNodeCardUnhealthy("node2", allCard),
		args: FaultNodeUpdateFaultNodesAttrArgs{
			fakeNPUNodeNilDeviceInfo("node2"),
		},
		wantErr: false,
	}
	test4 := FaultNodeUpdateFaultNodesAttrTests{
		name:   "node3",
		fields: *fakeTestFaultNodeNodeUnhealthy("node3"),
		args: FaultNodeUpdateFaultNodesAttrArgs{
			fakeNPUNodeNilDeviceInfo("node3"),
		},
		wantErr: false,
	}
	tests := []FaultNodeUpdateFaultNodesAttrTests{
		test1,
		test2,
		test3,
		test4,
	}
	return tests
}

// TestFaultNodeUpdateFaultNodesAttr test fault node attribute
func TestFaultNodeUpdateFaultNodesAttr(t *testing.T) {
	tests := buildFaultNodeUpdateFaultNodesAttrTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fNode := *fakeTestFaultNodeNodeHealthy("node0")
			if err := fNode.updateFaultNodesAttr(tt.args.node); (err != nil) != tt.wantErr {
				t.Errorf("updateFaultNodesAttr() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type TestReScheduler struct {
	DealReSchedulerCache *DealReSchedulerCache
	GraceDeleteTime      int64
	Level                string
	Jobs                 map[api.JobID]plugin.SchedulerJob
	Nodes                map[string]plugin.NPUNode
	kubeClient           kubernetes.Interface
}

type ReSchedulerCheckNodeNPUByTaskArgs struct {
	task   *api.TaskInfo
	vcNode plugin.NPUNode
}

type ReSchedulerCheckNodeNPUByTaskTests struct {
	name    string
	fields  TestReScheduler
	args    ReSchedulerCheckNodeNPUByTaskArgs
	wantErr bool
}

func buildReSchedulerCheckNodeNPUByTaskTests() []ReSchedulerCheckNodeNPUByTaskTests {
	faultNode := fakeTestFaultNodeNodeUnhealthy("node0")
	faultTask00 := fakeTestFaultTaskFault("pod0", "vcjob", "node0", "0", "pppp")
	faultTask01 := fakeTestFaultTaskHealth("pod1", "vcjob", "node1", "1", "oooo")
	faultJob0 := fakeTestFaultJob([]string{"node0", "node1"}, []string{"0", "1"}, []FaultTask{*faultTask00,
		*faultTask01}, "job0", "vcjob")
	field1 := TestReScheduler{
		DealReSchedulerCache: &DealReSchedulerCache{
			DealReSchedulerConfigmap:   nil,
			FaultNodes:                 []FaultNode{*faultNode},
			FaultJobs:                  []FaultJob{*faultJob0},
			NodeHeartbeats:             nil,
			AllocNodeRankOccurrenceMap: nil,
		},
		GraceDeleteTime: 0,
		Level:           "",
		Jobs:            nil,
		Nodes:           nil,
		kubeClient:      nil,
	}
	arg1 := ReSchedulerCheckNodeNPUByTaskArgs{
		task: test.FakeNormalTestTask("pod1", "node1", "job0"),
		vcNode: plugin.NPUNode{
			CommonNode: plugin.CommonNode{
				Name: "node2",
			},
		},
	}
	test1 := ReSchedulerCheckNodeNPUByTaskTests{
		name:    "01-CheckNodeNPUByTaskTests()-old task bind to new pod should be abandoned",
		fields:  field1,
		args:    arg1,
		wantErr: true,
	}
	tests := []ReSchedulerCheckNodeNPUByTaskTests{
		test1,
	}
	return tests
}

// TestReSchedulerCheckNodeNPUByTask test for re-scheduler check node NPU
func TestReSchedulerCheckNodeNPUByTask(t *testing.T) {
	tests := buildReSchedulerCheckNodeNPUByTaskTests()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reScheduler := fakeTestTTReScheduler(tt.fields)
			if err := reScheduler.CheckNodeNPUByTask(tt.args.task, tt.args.vcNode); (err != nil) != tt.wantErr {
				t.Errorf("CheckNodeNPUByTask() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type ReSchedulerCheckNodeCurNodeIsFaultArgs struct {
	curFJob *FaultJob
	task    *api.TaskInfo
	vcNode  plugin.NPUNode
}

type ReSchedulerCheckNodeCurNodeIsFaultTests struct {
	name    string
	fields  TestReScheduler
	args    ReSchedulerCheckNodeCurNodeIsFaultArgs
	wantErr bool
}

func buildReSchedulerCheckNodeCurNodeIsFaultTests() []ReSchedulerCheckNodeCurNodeIsFaultTests {
	test1 := ReSchedulerCheckNodeCurNodeIsFaultTests{
		name: "01-checkNodeCurNodeIsFault()-succeed",
		fields: TestReScheduler{
			DealReSchedulerCache: &DealReSchedulerCache{
				FaultNodes: []FaultNode{*fakeTestFaultNodeNodeUnhealthy("node0")},
				FaultJobs:  nil,
			},
		},
		args: ReSchedulerCheckNodeCurNodeIsFaultArgs{
			curFJob: &FaultJob{
				FaultTasks: []FaultTask{
					{
						TaskName: "pod0",
					},
				},
			},
			vcNode: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Name: "node0",
				},
			},
			task: &api.TaskInfo{
				Name: "pod0",
			},
		},
		wantErr: true,
	}
	tests := []ReSchedulerCheckNodeCurNodeIsFaultTests{
		test1,
	}
	return tests
}

// TestReSchedulerCheckNodeCurNodeIsFault test for check current node is fault node
func TestReSchedulerCheckNodeCurNodeIsFault(t *testing.T) {
	tests := buildReSchedulerCheckNodeCurNodeIsFaultTests()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reScheduler := fakeTestTTReScheduler(tt.fields)
			if err := reScheduler.checkNodeCurNodeIsFault(tt.args.vcNode, tt.args.task); (err != nil) != tt.wantErr {
				t.Errorf("checkNodeCurNodeIsFault() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type ReSchedulerUseAnnotationArgs struct {
	task *api.TaskInfo
	node *plugin.NPUNode
}

type ReSchedulerUseAnnotationTests struct {
	name    string
	fields  TestReScheduler
	args    ReSchedulerUseAnnotationArgs
	wantErr bool
}

func buildReSchedulerUseAnnotationTestArgs(nodeName string) ReSchedulerUseAnnotationArgs {
	args := ReSchedulerUseAnnotationArgs{
		task: &api.TaskInfo{
			Job:       "vcjob/job0",
			Name:      "pod1",
			Namespace: "vcjob",
			Pod: &v1.Pod{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: make(map[string]string, util.MapInitNum),
				},
			},
		},
		node: &plugin.NPUNode{
			CommonNode: plugin.CommonNode{
				Name: nodeName,
			},
		},
	}
	return args
}

func buildReSchedulerUseAnnotationTestFields(faultNode *FaultNode, faultJob0 *FaultJob,
	allocNodeRankTimeMap map[api.JobID][]AllocNodeRankOccurrence) TestReScheduler {
	reScheduler := TestReScheduler{
		DealReSchedulerCache: &DealReSchedulerCache{
			DealReSchedulerConfigmap:   nil,
			FaultNodes:                 []FaultNode{*faultNode},
			FaultJobs:                  []FaultJob{*faultJob0},
			NodeHeartbeats:             nil,
			AllocNodeRankOccurrenceMap: allocNodeRankTimeMap,
		},
		Jobs:       nil,
		Nodes:      nil,
		kubeClient: nil,
	}
	return reScheduler
}

func buildReSchedulerUseAnnotationRankIndexMap(nodeName string, rankIndex string, occ int) AllocNodeRankOccurrence {
	mapData := AllocNodeRankOccurrence{
		NodeName:   nodeName,
		RankIndex:  rankIndex,
		Occurrence: occ,
	}
	return mapData
}

func buildReSchedulerUseAnnotationTests() []ReSchedulerUseAnnotationTests {
	faultTask00 := fakeTestFaultTaskFault("pod0", "vcjob", "node0", "0", "pppp")
	faultTask01 := fakeTestFaultTaskHealth("pod1", "vcjob", "node1", "1", "oooo")
	faultJob0 := fakeTestFaultJob([]string{"node0", "node1"}, []string{"0", "1"}, []FaultTask{*faultTask00,
		*faultTask01}, "job0", "vcjob")
	faultNode := fakeTestFaultNodeNodeUnhealthy("node0")
	allocNodeRankTimeMap := map[api.JobID][]AllocNodeRankOccurrence{
		api.JobID("vcjob/job0"): {
			buildReSchedulerUseAnnotationRankIndexMap("node0", "0", 0),
			buildReSchedulerUseAnnotationRankIndexMap("node1", "1", 0),
		},
	}
	test1 := ReSchedulerUseAnnotationTests{
		name:    "01-UseAnnotation()-old node use old rankIndex",
		fields:  buildReSchedulerUseAnnotationTestFields(faultNode, faultJob0, allocNodeRankTimeMap),
		args:    buildReSchedulerUseAnnotationTestArgs("node1"),
		wantErr: false,
	}
	test2 := ReSchedulerUseAnnotationTests{
		name:    "02-UseAnnotation()-new node use old fault node rankIndex",
		fields:  buildReSchedulerUseAnnotationTestFields(faultNode, faultJob0, allocNodeRankTimeMap),
		args:    buildReSchedulerUseAnnotationTestArgs("node2"),
		wantErr: false,
	}
	allocNodeRankTimeMap3 := map[api.JobID][]AllocNodeRankOccurrence{
		api.JobID("vcjob/job0"): {
			buildReSchedulerUseAnnotationRankIndexMap("node0", "0", 1),
			buildReSchedulerUseAnnotationRankIndexMap("node1", "1", 0),
		},
	}
	test3 := ReSchedulerUseAnnotationTests{
		name:    "03-UseAnnotation()-fault node rankIndex occupied",
		fields:  buildReSchedulerUseAnnotationTestFields(faultNode, faultJob0, allocNodeRankTimeMap3),
		args:    buildReSchedulerUseAnnotationTestArgs("node2"),
		wantErr: false,
	}
	tests := []ReSchedulerUseAnnotationTests{
		test1,
		test2,
		test3,
	}
	return tests
}

// TestReSchedulerUseAnnotation test for use annotation
func TestReSchedulerUseAnnotation(t *testing.T) {
	tests := buildReSchedulerUseAnnotationTests()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reScheduler := fakeTestTTReScheduler(tt.fields)
			if err := reScheduler.UseAnnotation(tt.args.task, tt.args.node); (err != nil) != tt.wantErr {
				t.Errorf("UseAnnotation() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func fakeTestTTReScheduler(fields TestReScheduler) *ReScheduler {
	return &ReScheduler{
		DealReSchedulerCache: fields.DealReSchedulerCache,
		GraceDeleteTime:      fields.GraceDeleteTime,
		Level:                fields.Level,
		Jobs:                 fields.Jobs,
		Nodes:                fields.Nodes,
		kubeClient:           fields.kubeClient,
	}
}
