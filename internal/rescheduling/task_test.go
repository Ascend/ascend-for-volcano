/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package rescheduling is using for HuaWei Ascend pin fault rescheduling.

*/
package rescheduling

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
)

type addScoreByFaultNPUTaskTestArgs struct {
	task     *api.TaskInfo
	scoreMap map[string]float64
	cacheFun func()
}

type addScoreByFaultNPUTaskTestTest struct {
	name    string
	args    addScoreByFaultNPUTaskTestArgs
	want    map[string]float64
	wantErr error
}

func buildAddScoreByFaultNPUTaskTestCases() []addScoreByFaultNPUTaskTestTest {
	task1 := test.FakeNormalTestTasks(1)[0]
	testScoreMap := map[string]float64{task1.NodeName: node910X8NPUNum * node910X8NPUNum}
	const tmpNumber = 123456
	testCases := []addScoreByFaultNPUTaskTestTest{
		{
			name:    "01-scoreMap-nil-test",
			args:    addScoreByFaultNPUTaskTestArgs{task: task1, scoreMap: nil, cacheFun: func() {}},
			want:    nil,
			wantErr: errors.New("add score by fault NPU task scoreMap is nil"),
		},
		{
			name: "02-no-cache-err-test",
			args: addScoreByFaultNPUTaskTestArgs{
				task:     task1,
				scoreMap: testScoreMap,
				cacheFun: func() {
					ReSchedulerCache = make(map[string]interface{}, util.NPUIndex2)
					reTask := map[api.JobID]ReSchedulerTasks{
						"haha": {nil, nil, nil, nil, nil, task1.Namespace, false, tmpNumber}}
					ReSchedulerCache[CmJobKind] = reTask
				},
			},
			want:    testScoreMap,
			wantErr: fmt.Errorf("no %v in jobMap", task1.Job),
		},
		{
			name: "03-ok-test",
			args: addScoreByFaultNPUTaskTestArgs{
				task:     task1,
				scoreMap: testScoreMap,
				cacheFun: func() {
					ReSchedulerCache = make(map[string]interface{}, util.NPUIndex2)
					reTask := map[api.JobID]ReSchedulerTasks{
						task1.Job: {nil, nil, nil, nil, nil, task1.Namespace, false, tmpNumber}}
					ReSchedulerCache[CmJobKind] = reTask
				},
			},
			want:    testScoreMap,
			wantErr: nil,
		},
	}
	return testCases
}

// TestAddScoreByFaultNPUTask Test AddScoreByFaultNPUTask function
func TestAddScoreByFaultNPUTask(t *testing.T) {
	testCases := buildAddScoreByFaultNPUTaskTestCases()
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.cacheFun()
			got, err := AddScoreByFaultNPUTask(tt.args.task, tt.args.scoreMap)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("%v error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%v got = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

type checkFaultJobNodeArgs struct {
	task     *api.TaskInfo
	node     *api.NodeInfo
	cacheFun func()
}

type checkFaultJobNodeTest struct {
	name    string
	args    checkFaultJobNodeArgs
	wantErr error
}

func buildCheckFaultJobNodeTestCases() []checkFaultJobNodeTest {
	tasks := test.FakeNormalTestTasks(util.NPUIndex2)
	nodes := test.FakeNormalTestNodes(util.NPUIndex2)
	testCases := []checkFaultJobNodeTest{
		{
			name: "01-task in npu fault task-test",
			args: checkFaultJobNodeArgs{
				task: tasks[0], node: nil, cacheFun: func() {
					initTestReSchedulerCache()
					addTestTaskIntoReSchedulerCache(tasks[0])
				},
			},
			wantErr: nil,
		},
		{
			name: "02-node in fault node list-test",
			args: checkFaultJobNodeArgs{
				task: tasks[0], node: nodes[0], cacheFun: func() {
					initTestReSchedulerCache()
					addTestTaskIntoReSchedulerCache()
					addTestNodeIntoReSchedulerCache(nodes[0])
				},
			},
			wantErr: fmt.Errorf("%s is in fault node cache", nodes[0].Name),
		},
		{
			name: "03-node not in fault node list-test",
			args: checkFaultJobNodeArgs{
				task: tasks[1], node: nodes[0], cacheFun: func() {
					initTestReSchedulerCache()
					addTestTaskIntoReSchedulerCache(tasks[0])
					addTestNodeIntoReSchedulerCache(nodes[1])
				},
			},
			wantErr: fmt.Errorf("%s is used by npu fault job:%s", nodes[0].Name, tasks[1].Job),
		},
		{
			name: "04-job uses fault node-test",
			args: checkFaultJobNodeArgs{
				task: tasks[0], node: nodes[0], cacheFun: func() {
					initTestReSchedulerCache()
					addTestTaskIntoReSchedulerCache()
					addTestNodeIntoReSchedulerCache()
				},
			},
			wantErr: nil,
		},
	}
	return testCases
}

// TestCheckFaultJobNode test CheckFaultJobNode function
func TestCheckFaultJobNode(t *testing.T) {
	tests := buildCheckFaultJobNodeTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.cacheFun()
			if err := CheckFaultJobNode(tt.args.task, tt.args.node); !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type isNPUFaultTaskArgs struct {
	task     *api.TaskInfo
	cacheFun func()
}

type isNPUFaultTaskTest struct {
	name string
	args isNPUFaultTaskArgs
	want bool
}

func buildsIsNPUFaultTaskTestCases() []isNPUFaultTaskTest {
	tasks := test.FakeNormalTestTasks(util.NPUIndex2)
	testCases := []isNPUFaultTaskTest{
		{
			name: "01-no ReSchedulerCache-test",
			args: isNPUFaultTaskArgs{
				task: tasks[0], cacheFun: func() {
					initTestReSchedulerCache()
				},
			},
			want: false,
		},
		{
			name: "02-Not NPU fault task-test",
			args: isNPUFaultTaskArgs{
				task: tasks[0], cacheFun: func() {
					initTestReSchedulerCache()
					addTestTaskIntoReSchedulerCache(tasks[1])
				},
			},
			want: false,
		},
		{
			name: "03-Has NPU fault task-test",
			args: isNPUFaultTaskArgs{
				task: tasks[0], cacheFun: func() {
					initTestReSchedulerCache()
					addTestTaskIntoReSchedulerCache(tasks[0])
				},
			},
			want: true,
		},
	}
	return testCases
}

// TestIsNPUFaultTask Test IsNPUFaultTask function
func TestIsNPUFaultTask(t *testing.T) {
	tests := buildsIsNPUFaultTaskTestCases()
	for _, tt := range tests {
		tt.args.cacheFun()
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNPUFaultTask(tt.args.task); got != tt.want {
				t.Errorf("IsNPUFaultTask() = %v, want %v", got, tt.want)
			}
		})
	}
}

type setFaultJobPodIndexArgs struct {
	task     *api.TaskInfo
	node     *api.NodeInfo
	cacheFun func()
}

type setFaultJobPodIndexTest struct {
	name    string
	args    setFaultJobPodIndexArgs
	wantErr error
}

func buildSetFaultJobPodIndexTestCases() []setFaultJobPodIndexTest {
	tasks := test.FakeNormalTestTasks(util.NPUIndex2)
	nodes := test.FakeNormalTestNodes(util.NPUIndex2)
	testCases := []setFaultJobPodIndexTest{
		{
			name: "01-no record fault Cache-test",
			args: setFaultJobPodIndexArgs{
				task: tasks[0], node: nodes[0], cacheFun: func() {
					initTestReSchedulerCache()
				},
			},
			wantErr: fmt.Errorf("no %v in jobMap", tasks[0].Job),
		},
		{
			name: "02-not in record fault Cache-test",
			args: setFaultJobPodIndexArgs{
				task: tasks[0], node: nodes[1], cacheFun: func() {
					initTestReSchedulerCache()
					addTestTaskIntoReSchedulerCache(tasks[1])
				},
			},
			wantErr: fmt.Errorf("no %v in jobMap", tasks[0].Job),
		},
		{
			name: "03-in record fault Cache-test",
			args: setFaultJobPodIndexArgs{
				task: tasks[0], node: nodes[0], cacheFun: func() {
					initTestReSchedulerCache()
					addTestTaskIntoReSchedulerCache(tasks[0])
				},
			},
			wantErr: nil,
		},
	}
	return testCases
}

// TestSetFaultJobPodIndex test SetFaultJobPodIndex function.
func TestSetFaultJobPodIndex(t *testing.T) {
	tests := buildSetFaultJobPodIndexTestCases()
	for _, tt := range tests {
		tt.args.cacheFun()
		t.Run(tt.name, func(t *testing.T) {
			err := SetFaultJobPodIndex(tt.args.task, tt.args.node)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("SetFaultJobPodIndex() error = %v, wantErr %v", err, tt.wantErr)
			}
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
