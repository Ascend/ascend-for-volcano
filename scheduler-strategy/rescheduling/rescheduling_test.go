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
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"testing"
	"time"
	"volcano.sh/volcano/pkg/scheduler/api"
	ascendtest "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
)

type addScoreByFaultNPUTaskTestArgs struct {
	task     *api.TaskInfo
	scoreMap map[string]float64
	cacheFun func()
}

type addScoreByFaultNPUTaskTestTests []struct {
	name    string
	args    addScoreByFaultNPUTaskTestArgs
	want    map[string]float64
	wantErr error
}

func buildAddScoreByFaultNPUTaskTestCases() addScoreByFaultNPUTaskTestTests {
	task1 := ascendtest.FakeNormalTestTasks(1)[0]
	testScoreMap := map[string]float64{task1.NodeName: node910X8NPUNum * node910X8NPUNum}
	testCases := addScoreByFaultNPUTaskTestTests{
		{
			name:    "01-scoreMap-nil-test",
			args:    addScoreByFaultNPUTaskTestArgs{task: task1, scoreMap: nil, cacheFun: func() {}},
			want:    nil,
			wantErr: errors.New("AddScoreByFaultNPUTask scoreMap is nil"),
		},
		{
			name: "02-no-cache-err-test",
			args: addScoreByFaultNPUTaskTestArgs{
				task:     task1,
				scoreMap: testScoreMap,
				cacheFun: func() {
					ReSchedulerCache = make(map[string]interface{}, constIntNum2)
					reTask := map[api.JobID]ReSchedulerTasks{
						"haha": {nil, nil, nil, nil, task1.Namespace}}
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
					ReSchedulerCache = make(map[string]interface{}, constIntNum2)
					reTask := map[api.JobID]ReSchedulerTasks{
						task1.Job: {nil, nil, nil, nil, task1.Namespace}}
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

type checkFaultJobNodeTests []struct {
	name    string
	args    checkFaultJobNodeArgs
	wantErr error
}

func addTestTaskIntoReSchedulerCache(task *api.TaskInfo) {
	var reTask map[api.JobID]ReSchedulerTasks
	switch {
	case task == nil:
		reTask = map[api.JobID]ReSchedulerTasks{
			"haha": {nil, nil, nil, nil, ""}}
	default:
		reTask = map[api.JobID]ReSchedulerTasks{
			task.Job: {
				map[string]string{task.NodeName: task.NodeName},
				nil,
				nil,
				nil,
				""},
		}
	}
	ReSchedulerCache[CmJobKind] = reTask
}

func addTestNodeIntoReSchedulerCache(node *api.NodeInfo) {
	var faultNode map[string]FaultNodeState
	const testHeartBeat1 = 12345
	switch {
	case node == nil:
		faultNode = map[string]FaultNodeState{
			"hahaNode": {
				NodeName:   "nil",
				HealthCode: 0,
				UpdateTime: time.Now().Unix(),
				Heartbeat:  testHeartBeat1,
			},
		}
	default:
		faultNode = map[string]FaultNodeState{
			node.Name: {
				NodeName:   node.Name,
				HealthCode: 0,
				UpdateTime: time.Now().Unix(),
				Heartbeat:  testHeartBeat1,
			},
		}
	}
	ReSchedulerCache[CmNodeKind] = faultNode
}

func initTestReSchedulerCache() {
	ReSchedulerCache = make(map[string]interface{}, constIntNum2)
}

func buildCheckFaultJobNodeTestCases() checkFaultJobNodeTests {
	tasks := ascendtest.FakeNormalTestTasks(constIntNum2)
	nodes := ascendtest.FakeNormalTestNodes(constIntNum2)
	testCases := checkFaultJobNodeTests{
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
					addTestTaskIntoReSchedulerCache(nil)
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
					addTestTaskIntoReSchedulerCache(nil)
					addTestNodeIntoReSchedulerCache(nil)
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

type getDistributeUsableNPUTopArgs struct {
	nodeNPUTopology   []int
	netUnhealthyCards []int
}

type getDistributeUsableNPUTopTests []struct {
	name string
	args getDistributeUsableNPUTopArgs
	want []int
}

func buildGetDistributeUsableNPUTopTestCases() getDistributeUsableNPUTopTests {
	const constIntNum7 = 7
	testCases := getDistributeUsableNPUTopTests{
		{
			name: "01-nil node top-test",
			args: getDistributeUsableNPUTopArgs{
				nodeNPUTopology:   nil,
				netUnhealthyCards: []int{0, 1},
			},
			want: nil,
		},
		{
			name: "02-nil unhealthy top-test",
			args: getDistributeUsableNPUTopArgs{
				nodeNPUTopology:   []int{0, 1},
				netUnhealthyCards: nil,
			},
			want: []int{0, 1},
		},
		{
			name: "03-normal top-test",
			args: getDistributeUsableNPUTopArgs{
				nodeNPUTopology:   []int{0, 1, constIntNum2, constIntNum3, constIntNum7},
				netUnhealthyCards: []int{0, 1},
			},
			want: []int{constIntNum2, constIntNum3, constIntNum7},
		},
	}
	return testCases
}

// TestGetDistributeUsableNPUTop test GetDistributeUsableNPUTop function
func TestGetDistributeUsableNPUTop(t *testing.T) {
	tests := buildGetDistributeUsableNPUTopTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetDistributeUsableNPUTop(tt.args.nodeNPUTopology, tt.args.netUnhealthyCards)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetDistributeUsableNPUTop() = %v, want %v", got, tt.want)
			}
		})
	}
}

type getFaultNPUJobsArgs struct {
	jobs     map[string]*api.JobInfo
	cacheFun func()
}

type getFaultNPUJobsTests []struct {
	name    string
	args    getFaultNPUJobsArgs
	want    []FaultNPUJob
	wantErr error
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

func buildGetFaultNPUJobsTestCases() getFaultNPUJobsTests {
	jobInf := ascendtest.FakeNormalTestJob("pg1", 2)
	nodes := ascendtest.FakeNormalTestNodes(1)
	testCases := getFaultNPUJobsTests{
		{
			name: "01-job not set reschedule label-test",
			args: getFaultNPUJobsArgs{
				jobs: map[string]*api.JobInfo{jobInf.Name: jobInf}, cacheFun: func() {
					ascendtest.AddTestJobLabel(jobInf, "haha", "who")
				},
			},
			wantErr: errors.New("get none faultNPUJobs"),
		},
		{
			name: "02-job not has fault resources-test",
			args: getFaultNPUJobsArgs{
				jobs: map[string]*api.JobInfo{jobInf.Name: jobInf}, cacheFun: func() {
					ascendtest.AddTestJobLabel(jobInf, jobRescheduleLabelKey, jobRescheduleLabelValue)
					initTestReSchedulerCache()
					addTestNodeIntoReSchedulerCache(nil)
				},
			},
			wantErr: errors.New("get none faultNPUJobs"),
		},
		{
			name: "03-job not has rankIndex-test",
			args: getFaultNPUJobsArgs{
				jobs: map[string]*api.JobInfo{jobInf.Name: jobInf}, cacheFun: func() {
					ascendtest.AddTestJobLabel(jobInf, jobRescheduleLabelKey, jobRescheduleLabelValue)
					initTestReSchedulerCache()
					addTestNodeIntoReSchedulerCache(nodes[0])
				},
			},
			wantErr: errors.New("get none faultNPUJobs"),
		},
		{
			name: "04-job not has rankIndex-test",
			args: getFaultNPUJobsArgs{
				jobs: map[string]*api.JobInfo{jobInf.Name: jobInf}, cacheFun: func() {
					ascendtest.AddTestJobLabel(jobInf, jobRescheduleLabelKey, jobRescheduleLabelValue)
					initTestReSchedulerCache()
					addTestNodeIntoReSchedulerCache(nodes[0])
					addTestJobRankIndex(jobInf)
				},
			},
			wantErr: errors.New("get none faultNPUJobs"),
		},
	}
	return testCases
}

// TestGetFaultNPUJobs Test GetFaultNPUJobs function
func TestGetFaultNPUJobs(t *testing.T) {
	tests := buildGetFaultNPUJobsTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.cacheFun()
			got, err := GetFaultNPUJobs(tt.args.jobs)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("GetFaultNPUJobs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetFaultNPUJobs() got = %v, want %v", got, tt.want)
			}
		})
	}
}
