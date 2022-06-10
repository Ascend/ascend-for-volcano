/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package rescheduling is using for HuaWei Ascend pin fault rescheduling.

*/
package rescheduling

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"

	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
)

type setFaultInNodeAndJobsArgs struct {
	fNPUJobs []FaultNPUJob
	jobs     map[string]*api.JobInfo
	cacheFun func()
}

type setFaultInNodeAndJobsTest struct {
	ssn     *framework.Session
	name    string
	args    setFaultInNodeAndJobsArgs
	wantErr error
}

func buildSetFaultInNodeAndJobsTestCases() []setFaultInNodeAndJobsTest {
	testSsn := test.FakeNormalSSN()
	fakeJob1 := test.FakeNormalTestJob("pg", util.NPUIndex3)
	fakeJob2 := test.FakeNormalTestJob("pg1", util.NPUIndex3)
	mapJob1 := map[string]*api.JobInfo{fakeJob1.Name: fakeJob1}
	mapJob2 := map[string]*api.JobInfo{fakeJob2.Name: fakeJob2}
	faultJob1 := addJobIntoFaultNPUJobStruct(fakeJob1)
	testCases := []setFaultInNodeAndJobsTest{
		{
			ssn:  testSsn,
			name: "01-job not in record fault Cache-test",
			args: setFaultInNodeAndJobsArgs{
				jobs: mapJob2, fNPUJobs: []FaultNPUJob{faultJob1}, cacheFun: func() {
					initTestReSchedulerCache()
				},
			},
			wantErr: nil,
		},
		{
			ssn:  testSsn,
			name: "02-write in fault Cache success-test",
			args: setFaultInNodeAndJobsArgs{
				jobs: mapJob1, fNPUJobs: []FaultNPUJob{faultJob1}, cacheFun: func() {
					initTestReSchedulerCache()
				},
			},
			wantErr: nil,
		},
	}
	return testCases
}

// TestSetFaultInNodeAndJobs test SetFaultInNodeAndJobs function.
func TestSetFaultInNodeAndJobs(t *testing.T) {
	tests := buildSetFaultInNodeAndJobsTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.cacheFun()
			err := SetFaultInNodeAndJobs(tt.ssn, tt.args.fNPUJobs, tt.args.jobs)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("SetFaultInNodeAndJobs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type readFaultNPUJobsFromCMArgs struct {
	ssn            *framework.Session
	cacheFunBefore func()
	cacheFunAfter  func()
}

type readFaultNPUJobsFromCMTest struct {
	name    string
	args    readFaultNPUJobsFromCMArgs
	wantErr error
}

func buildReadFaultNPUJobsFromCMTestCases() []readFaultNPUJobsFromCMTest {
	testSsn := test.FakeNormalSSN()
	job1 := test.FakeNormalTestJob("pg1", util.NPUIndex2)
	nodes := test.FakeNormalTestNodes(1)
	addTestNodeIntoReSchedulerCache(nodes[0])
	test.SetNPUNodeLabel(testSsn.Nodes[nodes[0].Name].Node, nodeDEnableKey, nodeDEnableOnValue)

	var tmpPatche *gomonkey.Patches
	testCases := []readFaultNPUJobsFromCMTest{
		{
			name: "01-read from config map success Cache-test",
			args: readFaultNPUJobsFromCMArgs{
				ssn: testSsn, cacheFunBefore: func() {
					initTestReSchedulerCache()
					addTestJobIntoReSchedulerCache(job1)
					tmpPatche = gomonkey.ApplyFunc(util.GetConfigMapWithRetry, fakeErrorHeartbeatCmData)
				}, cacheFunAfter: func() {
					if tmpPatche != nil {
						tmpPatche.Reset()
					}
				},
			},
			wantErr: nil,
		},
	}
	return testCases
}

// TestReadFaultNPUJobsFromCM test ReadFaultNPUJobsFromCM function
func TestReadFaultNPUJobsFromCM(t *testing.T) {
	tests := buildReadFaultNPUJobsFromCMTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.cacheFunBefore()
			err := ReadFaultNPUJobsFromCM(tt.args.ssn)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("ReadFaultNPUJobsFromCM() error = %v, wantErr %v", err, tt.wantErr)
			}
			tt.args.cacheFunAfter()
		})
	}
}

type writeReSchedulerDataToCMArgs struct {
	ssn             *framework.Session
	reSchedulerData map[string]interface{}
	cacheFunBefore  func()
	cacheFunAfter   func()
}

type writeReSchedulerDataToCMTest struct {
	name    string
	args    writeReSchedulerDataToCMArgs
	wantErr error
}

func buildWriteReSchedulerDataToCMTestCases() []writeReSchedulerDataToCMTest {
	testSsn := test.FakeNormalSSN()
	job1 := test.FakeNormalTestJob("pg1", util.NPUIndex2)
	nodes := test.FakeNormalTestNodes(1)
	test.SetNPUNodeLabel(testSsn.Nodes[nodes[0].Name].Node, nodeDEnableKey, nodeDEnableOnValue)

	var tmpPatche *gomonkey.Patches
	testCases := []writeReSchedulerDataToCMTest{
		{
			name: "01-write data to config map success Cache-test Cache-test",
			args: writeReSchedulerDataToCMArgs{
				ssn: testSsn, reSchedulerData: ReSchedulerCache, cacheFunBefore: func() {
					initTestReSchedulerCache()
					addTestJobIntoReSchedulerCache(job1)
					addTestNodeIntoReSchedulerCache(nodes[0])
					addTestCardIntoReSchedulerCache(nodes[0].Name, nil, []string{"Ascend910-0", "Ascend910-1"})
					addTestHeartbeatIntoReSchedulerCache(nodes[0].Name)
					tmpPatche = gomonkey.ApplyFunc(util.CreateOrUpdateConfigMap,
						func(k8s kubernetes.Interface, cm *v1.ConfigMap, cmName, cmNameSpac string) error {
							return nil
						})
				}, cacheFunAfter: func() {
					if tmpPatche != nil {
						tmpPatche.Reset()
					}
				},
			},
			wantErr: nil,
		},
	}
	return testCases
}

// TestWriteReSchedulerDataToCM test WriteReSchedulerDataToCM function
func TestWriteReSchedulerDataToCM(t *testing.T) {
	tests := buildWriteReSchedulerDataToCMTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.cacheFunBefore()
			err := WriteReSchedulerDataToCM(tt.args.ssn, tt.args.reSchedulerData)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("WriteReSchedulerDataToCM() error = %v, wantErr %v", err, tt.wantErr)
			}
			tt.args.cacheFunAfter()
		})
	}
}

type getRankIndexMapByTaskArgs struct {
	task      *api.TaskInfo
	cacheFunc func()
}

type getRankIndexMapByTaskTests struct {
	name    string
	args    getRankIndexMapByTaskArgs
	want    TaskUsedRankIndex
	wantErr error
}

func buildGetRankIndexMapByTaskTestCases() []getRankIndexMapByTaskTests {
	const tmpNumber = 123456
	task0 := test.FakeNormalTestTask("task0", "node0", "pg0")
	job1 := test.FakeNormalTestJob("pg1", util.NPUIndex2)
	task1 := test.FakeNormalTestTask("task1", "node1", "pg1")

	var reRankIDs = make(map[api.JobID]TaskUsedRankIndex, util.NPUIndex2)
	reRankIDs[job1.UID] = TaskUsedRankIndex{
		FaultNodeRankIndex: map[string]struct{ UpdateTime int64 }{
			"node1": {tmpNumber}},
		UpdateTime: tmpNumber}

	testCases := []getRankIndexMapByTaskTests{
		{
			name:    "01-getRankIndexMapByTask()- no rankidx-test",
			args:    getRankIndexMapByTaskArgs{task: task0, cacheFunc: func() {}},
			want:    TaskUsedRankIndex{},
			wantErr: fmt.Errorf("no rankIndex cache"),
		},
		{
			name: "02-getRankIndexMapByTask()- success-test",
			args: getRankIndexMapByTaskArgs{task: task1, cacheFunc: func() {
				initTestReSchedulerCache()
				addTmpAllocRankIndexIntoReschedulerCache(reRankIDs)
			}},
			want: TaskUsedRankIndex{
				FaultNodeRankIndex: map[string]struct{ UpdateTime int64 }{
					"node1": {tmpNumber}},
				UpdateTime: tmpNumber},
			wantErr: nil,
		},
	}
	return testCases
}

func TestGetRankIndexMapByTask(t *testing.T) {
	tests := buildGetRankIndexMapByTaskTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.cacheFunc()
			got, err := getRankIndexMapByTask(tt.args.task)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("getRankIndexMapByTask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getRankIndexMapByTask() got = %v, want %v", got, tt.want)
			}
		})
	}
}
