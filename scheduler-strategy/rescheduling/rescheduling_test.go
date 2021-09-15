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
	"volcano.sh/volcano/pkg/scheduler/framework"
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

func addTestTaskIntoReSchedulerCache(tasks ...*api.TaskInfo) {
	var reTask = make(map[api.JobID]ReSchedulerTasks, constIntNum2)

	if len(tasks) == 0 {
		reTask = map[api.JobID]ReSchedulerTasks{
			"hahaTask": {nil, nil, nil, nil, ""}}
		ReSchedulerCache[CmJobKind] = reTask
		return
	}
	for _, task := range tasks {
		reTask[task.Job] = ReSchedulerTasks{
			NodeNames:   map[string]string{task.Name: task.NodeName},
			RankIndexes: nil,
			Time:        nil,
			TaskUseNPUs: nil,
			NameSpace:   ""}
	}

	ReSchedulerCache[CmJobKind] = reTask
}

func addTestNodeIntoReSchedulerCache(nodes ...*api.NodeInfo) {
	var faultNode = make(map[string]FaultNodeState, constIntNum2)
	const testHeartBeat1 = 12345

	if len(nodes) == 0 {
		faultNode["hahaNode"] = FaultNodeState{
			NodeName:   "nil",
			HealthCode: 0,
			UpdateTime: time.Now().Unix(),
			Heartbeat:  testHeartBeat1,
		}
		ReSchedulerCache[CmNodeKind] = faultNode
		return
	}

	for _, node := range nodes {
		faultNode[node.Name] = FaultNodeState{
			NodeName:   node.Name,
			HealthCode: 0,
			UpdateTime: time.Now().Unix(),
			Heartbeat:  testHeartBeat1,
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
					ascendtest.AddTestJobLabel(jobInf, jobRescheduleLabelKey, JobGraceRescheduleLabelValue)
					initTestReSchedulerCache()
					addTestNodeIntoReSchedulerCache()
				},
			},
			wantErr: errors.New("get none faultNPUJobs"),
		},
		{
			name: "03-job not has rankIndex-test",
			args: getFaultNPUJobsArgs{
				jobs: map[string]*api.JobInfo{jobInf.Name: jobInf}, cacheFun: func() {
					ascendtest.AddTestJobLabel(jobInf, jobRescheduleLabelKey, JobForceRescheduleLabelValue)
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
					ascendtest.AddTestJobLabel(jobInf, jobRescheduleLabelKey, JobForceRescheduleLabelValue)
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

type getFaultTaskUseNodeInfoArgs struct {
	task     *api.TaskInfo
	ssn      *framework.Session
	cacheFun func()
}

type getFaultTaskUseNodeInfoTests []struct {
	name    string
	args    getFaultTaskUseNodeInfoArgs
	wantErr error
}

func buildGetFaultTaskUseNodeInfoTestCases() getFaultTaskUseNodeInfoTests {
	const constNum4, constNum3 = 4, 3
	tasks := ascendtest.FakeNormalTestTasks(constNum4)
	nodes := ascendtest.FakeNormalTestNodes(constNum4)
	ssnTest := ascendtest.FakeNormalSSN()
	testCases := getFaultTaskUseNodeInfoTests{
		{
			name: "01-task not in fault task list-test",
			args: getFaultTaskUseNodeInfoArgs{
				task: tasks[0], ssn: nil, cacheFun: func() {
					initTestReSchedulerCache()
					addTestTaskIntoReSchedulerCache()
				},
			},
			wantErr: fmt.Errorf("no %v in jobMap", tasks[0].Job),
		},
		{
			name: "02-task use node not ssn node list-test",
			args: getFaultTaskUseNodeInfoArgs{
				task: tasks[constNum3], ssn: ssnTest, cacheFun: func() {
					initTestReSchedulerCache()
					addTestTaskIntoReSchedulerCache(tasks[constNum3])
				},
			},
			wantErr: fmt.Errorf("get node name %s failed", tasks[constNum3].NodeName),
		},
		{
			name: "03-task use node in fault node list-test",
			args: getFaultTaskUseNodeInfoArgs{
				task: tasks[0], ssn: ssnTest, cacheFun: func() {
					initTestReSchedulerCache()
					addTestTaskIntoReSchedulerCache(tasks[0])
					addTestNodeIntoReSchedulerCache(nodes[0])
				},
			},
			wantErr: fmt.Errorf("GetFaultTaskUseNodeInfo %s in fault node list", tasks[0].NodeName),
		},
		{
			name: "04-task use node which not fault resource-test",
			args: getFaultTaskUseNodeInfoArgs{
				task: tasks[0], ssn: ssnTest, cacheFun: func() {
					initTestReSchedulerCache()
					addTestTaskIntoReSchedulerCache(tasks[0])
					addTestNodeIntoReSchedulerCache(nodes[constNum3])
				},
			},
			wantErr: nil,
		},
	}
	return testCases
}

// TestGetFaultTaskUseNodeInfo test GetFaultTaskUseNodeInfo function.
func TestGetFaultTaskUseNodeInfo(t *testing.T) {
	tests := buildGetFaultTaskUseNodeInfoTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.cacheFun()
			_, err := GetFaultTaskUseNodeInfo(tt.args.task, tt.args.ssn)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("GetFaultTaskUseNodeInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

type getNetworkUnhealthyCardsArgs struct {
	node     *api.NodeInfo
	cacheFun func()
}

type getNetworkUnhealthyCardsTests []struct {
	name string
	args getNetworkUnhealthyCardsArgs
	want []int
}

func addTestCardIntoReSchedulerCache(nodeName string, faultCards []string, netUnhealthyCards []string) {
	var reCard = make(map[string]FaultNPUsOnNode, constIntNum2)

	reCard[nodeName] = FaultNPUsOnNode{
		NodeName:             nodeName,
		FaultNPUs:            faultCards,
		NetworkUnhealthyNPUs: netUnhealthyCards,
		UpdateTime:           time.Now().Unix()}

	ReSchedulerCache[CmCardKind] = reCard
}

func buildTestGetNetworkUnhealthyCardsTestCases() getNetworkUnhealthyCardsTests {
	const constNum4 = 4
	nodes := ascendtest.FakeNormalTestNodes(constNum4)
	testCases := getNetworkUnhealthyCardsTests{
		{
			name: "01-no ReSchedulerCache-test",
			args: getNetworkUnhealthyCardsArgs{
				node: nodes[0], cacheFun: func() {},
			},
			want: nil,
		},
		{
			name: "02-get networkUnhealthy cards ok-test",
			args: getNetworkUnhealthyCardsArgs{
				node: nodes[0], cacheFun: func() {
					initTestReSchedulerCache()
					addTestCardIntoReSchedulerCache(nodes[0].Name, nil, []string{"Ascend910-0", "Ascend910-1"})
				},
			},
			want: []int{0, 1},
		},
	}
	return testCases
}

// TestGetNetworkUnhealthyCards test GetNetworkUnhealthyCards function.
func TestGetNetworkUnhealthyCards(t *testing.T) {
	tests := buildTestGetNetworkUnhealthyCardsTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.cacheFun()
			if got := GetNetworkUnhealthyCards(tt.args.node); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetNetworkUnhealthyCards() = %v, want %v", got, tt.want)
			}
		})
	}
}

type getNodeIdleNPUIntCardsIncludeFaultTaskArgs struct {
	task     *api.TaskInfo
	node     *api.NodeInfo
	cacheFun func()
}

type getNodeIdleNPUIntCardsIncludeFaultTaskTests []struct {
	name string
	args getNodeIdleNPUIntCardsIncludeFaultTaskArgs
	want []int
}

func buildTestGetNodeIdleNPUIntCardsIncludeFaultTaskTestCases() getNodeIdleNPUIntCardsIncludeFaultTaskTests {
	nodes := ascendtest.FakeNormalTestNodes(constIntNum2)
	tasks := ascendtest.FakeNormalTestTasks(constIntNum2)
	var nodeNPU = "Ascend910-0,Ascend910-1,Ascend910-2,Ascend910-3"
	var nodeFaultNPU = "Ascend910-0,Ascend910-1"
	testCases := getNodeIdleNPUIntCardsIncludeFaultTaskTests{
		{
			name: "01-no ReSchedulerCache-test",
			args: getNodeIdleNPUIntCardsIncludeFaultTaskArgs{
				task: tasks[0], node: nodes[0], cacheFun: func() {
					initTestReSchedulerCache()
				},
			},
			want: nil,
		},
		{
			name: "02-task npu not include node fault npu-test",
			args: getNodeIdleNPUIntCardsIncludeFaultTaskArgs{
				task: tasks[0], node: nodes[0], cacheFun: func() {
					initTestReSchedulerCache()
					ascendtest.SetTestNPUNodeOther(nodes[0], npu800And9000CardName, nodeNPU)
					ascendtest.SetTestNPUNodeAnnotation(nodes[0], faultNPU, nodeFaultNPU)
				},
			},
			want: []int{2, 3},
		},
		{
			name: "03-task npu include node fault npu-test",
			args: getNodeIdleNPUIntCardsIncludeFaultTaskArgs{
				task: tasks[0], node: nodes[0], cacheFun: func() {
					initTestReSchedulerCache()
					ascendtest.SetTestNPUNodeOther(nodes[0], npu800And9000CardName, nodeNPU)
					ascendtest.SetTestNPUNodeAnnotation(nodes[0], faultNPU, nodeFaultNPU)
					ascendtest.SetTestNPUPodAnnotation(tasks[0].Pod, npu800And9000CardName, nodeFaultNPU)
					addTestTaskIntoReSchedulerCache(tasks[0])
				},
			},
			want: []int{2, 3},
		},
	}
	return testCases
}

func TestGetNodeIdleNPUIntCardsIncludeFaultTask(t *testing.T) {
	tests := buildTestGetNodeIdleNPUIntCardsIncludeFaultTaskTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.cacheFun()
			got := GetNodeIdleNPUIntCardsIncludeFaultTask(tt.args.task, tt.args.node)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetNodeIdleNPUIntCardsIncludeFaultTask() = %v, want %v", got, tt.want)
			}
		})
	}
}
