/*
Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package rescheduling is using for HuaWei Ascend pin fault rescheduling.

*/
package rescheduling

import (
	"errors"
	"fmt"
	"github.com/agiledragon/gomonkey/v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"reflect"
	"strconv"
	"strings"
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
	const tmpNumber = 123456
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
					ReSchedulerCache = make(map[string]interface{}, constIntNum2)
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

type checkFaultJobNodeTests []struct {
	name    string
	args    checkFaultJobNodeArgs
	wantErr error
}

func addTestTaskIntoReSchedulerCache(tasks ...*api.TaskInfo) {
	var reTask = make(map[api.JobID]ReSchedulerTasks, constIntNum2)
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
	var faultNode = make(map[string]FaultNodeState, constIntNum2)
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
	jobInf := ascendtest.FakeNormalTestJob("pg1", constIntNum2)
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
			wantErr: nil,
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
			_, err := GetFaultNPUJobs(tt.args.jobs)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("GetFaultNPUJobs() error = %v, wantErr %v", err, tt.wantErr)
				return
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

func addTestCardIntoReSchedulerCache(nodeName string, faultCards, netUnhealthyCards []string) {
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
			if got := GetNetworkUnhealthyCards(tt.args.node.Name); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetNetworkUnhealthyCards() = %v, want %v", got, tt.want)
			}
		})
	}
}

type getRestartNPUFaultJobsArgs struct {
	faultNPUJobs []FaultNPUJob
	jobs         map[string]*api.JobInfo
}

type getRestartNPUFaultJobsTests []struct {
	name    string
	args    getRestartNPUFaultJobsArgs
	want    []*api.JobInfo
	wantErr error
}

func addJobIntoFaultNPUJobStruct(job *api.JobInfo) FaultNPUJob {
	faultJob := FaultNPUJob{
		faultNPUJobBase: faultNPUJobBase{
			jobName:          job.Name,
			namespace:        job.Namespace,
			taskUseRankIndex: make(map[string]string, constIntNum2),
			taskUseNode:      make(map[string]string, constIntNum2),
		},
		taskUseNPUs: make(map[string]string, constIntNum2),
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

func buildTestGetRestartNPUFaultJobsTestCases() getRestartNPUFaultJobsTests {
	job := ascendtest.FakeNormalTestJob("pg1", constIntNum2)
	faultJob := addJobIntoFaultNPUJobStruct(job)
	testCases := getRestartNPUFaultJobsTests{
		{
			name: "01-no restart jobs-test",
			args: getRestartNPUFaultJobsArgs{
				faultNPUJobs: nil, jobs: map[string]*api.JobInfo{job.Name: job},
			},
			want:    nil,
			wantErr: errors.New("none restart jobs get"),
		},
		{
			name: "02-has restart jobs-test",
			args: getRestartNPUFaultJobsArgs{
				faultNPUJobs: []FaultNPUJob{faultJob}, jobs: map[string]*api.JobInfo{job.Name: job},
			},
			want:    []*api.JobInfo{job},
			wantErr: nil,
		},
	}
	return testCases
}

// TestGetRestartNPUFaultJobs Test GetRestartNPUFaultJobs function.
func TestGetRestartNPUFaultJobs(t *testing.T) {

	tests := buildTestGetRestartNPUFaultJobsTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetRestartNPUFaultJobs(tt.args.faultNPUJobs, tt.args.jobs)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("GetRestartNPUFaultJobs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetRestartNPUFaultJobs() got = %v, want %v", got, tt.want)
			}
		})
	}
}

type isDistributedJobArgs struct {
	job *api.JobInfo
}

type isDistributedJobTests []struct {
	name string
	args isDistributedJobArgs
	want bool
}

func buildsDistributedJobTestCases() isDistributedJobTests {
	job1 := ascendtest.FakeNormalTestJob("pg1", 1)
	job2 := ascendtest.FakeNormalTestJob("pg1", constIntNum2)
	testCases := isDistributedJobTests{
		{
			name: "01-not distributed job-test",
			args: isDistributedJobArgs{
				job: job1,
			},
			want: false,
		},
		{
			name: "02-distributed job-test",
			args: isDistributedJobArgs{
				job: job2,
			},
			want: true,
		},
	}
	return testCases
}

// TestIsDistributedJob Test IsDistributedJob function.
func TestIsDistributedJob(t *testing.T) {
	tests := buildsDistributedJobTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsDistributedJob(tt.args.job); got != tt.want {
				t.Errorf("IsDistributedJob() = %v, want %v", got, tt.want)
			}
		})
	}
}

type isNPUFaultTaskArgs struct {
	task     *api.TaskInfo
	cacheFun func()
}

type isNPUFaultTaskTests []struct {
	name string
	args isNPUFaultTaskArgs
	want bool
}

func buildsIsNPUFaultTaskTestCases() isNPUFaultTaskTests {
	tasks := ascendtest.FakeNormalTestTasks(constIntNum2)
	testCases := isNPUFaultTaskTests{
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

type isNodeInFaultNodeListArgs struct {
	node     *api.NodeInfo
	cacheFun func()
}

type isNodeInFaultNodeListTests []struct {
	name string
	args isNodeInFaultNodeListArgs
	want bool
}

func buildIsNodeInFaultNodeListTestCases() isNodeInFaultNodeListTests {
	nodes := ascendtest.FakeNormalTestNodes(constIntNum2)
	testCases := isNodeInFaultNodeListTests{
		{
			name: "01-no ReSchedulerCache-test",
			args: isNodeInFaultNodeListArgs{
				node: nodes[0], cacheFun: func() {
					initTestReSchedulerCache()
				},
			},
			want: false,
		},
		{
			name: "02-not NPU fault node-test",
			args: isNodeInFaultNodeListArgs{
				node: nodes[0], cacheFun: func() {
					initTestReSchedulerCache()
					addTestNodeIntoReSchedulerCache(nodes[1])
				},
			},
			want: false,
		},
		{
			name: "03-has ReSchedulerCache-test",
			args: isNodeInFaultNodeListArgs{
				node: nodes[0], cacheFun: func() {
					initTestReSchedulerCache()
					addTestNodeIntoReSchedulerCache(nodes[0])
				},
			},
			want: true,
		},
	}
	return testCases
}

// TestIsNodeInFaultNodeList Test IsNodeInFaultNodeList function.
func TestIsNodeInFaultNodeList(t *testing.T) {
	tests := buildIsNodeInFaultNodeListTestCases()
	for _, tt := range tests {
		tt.args.cacheFun()
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNodeInFaultNodeList(tt.args.node); got != tt.want {
				t.Errorf("IsNodeInFaultNodeList() = %v, want %v", got, tt.want)
			}
		})
	}
}

type recordFaultInfInCacheArgs struct {
	ssn       *framework.Session
	npuNumber int
	cacheFun  func()
}

type recordFaultInfInCacheTests []struct {
	name    string
	args    recordFaultInfInCacheArgs
	wantErr error
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
		ascendtest.PrintError("not support type:%+v", setData)
	}
}

func buildRecordFaultInfInCacheTestCases() recordFaultInfInCacheTests {
	const constNumber64 = 12345677
	ssnTest := ascendtest.FakeNormalSSN()
	fNPUs := FaultNPUsOnNode{NodeName: "node1", FaultNPUs: []string{"Ascend910-0", "Ascend910-1", "Ascend910-1"},
		NetworkUnhealthyNPUs: []string{}, UpdateTime: constNumber64}
	fNode := FaultNodeState{NodeName: "node1", HealthCode: 0, UpdateTime: constNumber64, Heartbeat: constNumber64,
		HeartbeatInterval: nodeUpdateTime}
	testCases := recordFaultInfInCacheTests{
		{
			name: "01-record fault success-test",
			args: recordFaultInfInCacheArgs{
				ssn: ssnTest, npuNumber: node910X8NPUNum, cacheFun: func() {
					initTestReSchedulerCache()
					setTestSsnNode(ssnTest, fNPUs)
					setTestSsnNode(ssnTest, fNode)
				},
			},
			wantErr: nil,
		},
	}
	return testCases
}

// TestRecordFaultInfInCache test RecordFaultInfInCache function.
func TestRecordFaultInfInCache(t *testing.T) {
	tests := buildRecordFaultInfInCacheTestCases()
	for _, tt := range tests {
		tt.args.cacheFun()
		t.Run(tt.name, func(t *testing.T) {
			err := RecordFaultInfInCache(tt.args.ssn, tt.args.npuNumber)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type releaseFaultJobTakeNodesArgs struct {
	job      *api.JobInfo
	cacheFun func()
}

type releaseFaultJobTakeNodesTests []struct {
	name    string
	args    releaseFaultJobTakeNodesArgs
	wantErr error
}

func addTestJobIntoReSchedulerCache(job *api.JobInfo) {
	var reJobs = make(map[api.JobID]ReSchedulerTasks, constIntNum2)
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

func buildReleaseFaultJobTakeNodesTestCases() releaseFaultJobTakeNodesTests {
	fakeJob := ascendtest.FakeNormalTestJob("pg", constIntNum3)
	fakeJob1 := ascendtest.FakeNormalTestJob("pg", constIntNum3)
	testCases := releaseFaultJobTakeNodesTests{
		{
			name: "01-no record fault Cache-test",
			args: releaseFaultJobTakeNodesArgs{
				job: fakeJob, cacheFun: func() {
					initTestReSchedulerCache()
				},
			},
			wantErr: nil,
		},
		{
			name: "02-job not in fault Cache-test",
			args: releaseFaultJobTakeNodesArgs{
				job: fakeJob, cacheFun: func() {
					initTestReSchedulerCache()
					addTestJobIntoReSchedulerCache(fakeJob1)
				},
			},
			wantErr: nil,
		},
		{
			name: "03-release job from fault Cache-test",
			args: releaseFaultJobTakeNodesArgs{
				job: fakeJob, cacheFun: func() {
					initTestReSchedulerCache()
					addTestJobIntoReSchedulerCache(fakeJob)
				},
			},
			wantErr: nil,
		},
	}
	return testCases
}

// TestReleaseFaultJobTakeNodes test ReleaseFaultJobTakeNodes function.
func TestReleaseFaultJobTakeNodes(t *testing.T) {
	tests := buildReleaseFaultJobTakeNodesTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.cacheFun()
			err := ReleaseFaultJobTakeNodes(tt.args.job)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("ReleaseFaultJobTakeNodes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type setFaultJobPodIndexArgs struct {
	task     *api.TaskInfo
	node     *api.NodeInfo
	cacheFun func()
}

type setFaultJobPodIndexTests []struct {
	name    string
	args    setFaultJobPodIndexArgs
	wantErr error
}

func buildSetFaultJobPodIndexTestCases() setFaultJobPodIndexTests {
	tasks := ascendtest.FakeNormalTestTasks(constIntNum2)
	nodes := ascendtest.FakeNormalTestNodes(constIntNum2)
	testCases := setFaultJobPodIndexTests{
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

type setFaultInNodeAndJobsArgs struct {
	fNPUJobs []FaultNPUJob
	jobs     map[string]*api.JobInfo
	cacheFun func()
}

type setFaultInNodeAndJobsTests []struct {
	ssn     *framework.Session
	name    string
	args    setFaultInNodeAndJobsArgs
	wantErr error
}

func buildSetFaultInNodeAndJobsTestCases() setFaultInNodeAndJobsTests {
	testSsn := ascendtest.FakeNormalSSN()
	fakeJob1 := ascendtest.FakeNormalTestJob("pg", constIntNum3)
	fakeJob2 := ascendtest.FakeNormalTestJob("pg1", constIntNum3)
	mapJob1 := map[string]*api.JobInfo{fakeJob1.Name: fakeJob1}
	mapJob2 := map[string]*api.JobInfo{fakeJob2.Name: fakeJob2}
	faultJob1 := addJobIntoFaultNPUJobStruct(fakeJob1)
	testCases := setFaultInNodeAndJobsTests{
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

type readFaultNPUJobsFromCMTests []struct {
	name    string
	args    readFaultNPUJobsFromCMArgs
	wantErr error
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

func buildReadFaultNPUJobsFromCMTestCases() readFaultNPUJobsFromCMTests {
	testSsn := ascendtest.FakeNormalSSN()
	job1 := ascendtest.FakeNormalTestJob("pg1", constIntNum2)
	nodes := ascendtest.FakeNormalTestNodes(1)
	addTestNodeIntoReSchedulerCache(nodes[0])
	ascendtest.SetNPUNodeLabel(testSsn.Nodes[nodes[0].Name].Node, nodeDEnableKey, nodeDEnableOnValue)

	var tmpPatche *gomonkey.Patches
	testCases := readFaultNPUJobsFromCMTests{
		{
			name: "01-read from config map success Cache-test",
			args: readFaultNPUJobsFromCMArgs{
				ssn: testSsn, cacheFunBefore: func() {
					initTestReSchedulerCache()
					addTestJobIntoReSchedulerCache(job1)
					tmpPatche = gomonkey.ApplyFunc(getConfigMapWithRetry, fakeErrorHeartbeatCmData)
				}, cacheFunAfter: func() {
					tmpPatche.Reset()
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

type writeReSchedulerDataToCMTests []struct {
	name    string
	args    writeReSchedulerDataToCMArgs
	wantErr error
}

func addTestHeartbeatIntoReSchedulerCache(nodeName string) {
	var reCard = make(map[string]NormalNodeHeartbeat, constIntNum2)
	const tmpNumber = 123456
	reCard[nodeName] = NormalNodeHeartbeat{
		NodeDHeartbeat:      tmpNumber,
		UpdateHeartbeatTime: tmpNumber,
		HeartbeatInterval:   tmpNumber,
		UpdateTime:          time.Now().Unix()}

	ReSchedulerCache[CmNodeHeartbeatKind] = reCard
}

func buildWriteReSchedulerDataToCMTestCases() writeReSchedulerDataToCMTests {
	testSsn := ascendtest.FakeNormalSSN()
	job1 := ascendtest.FakeNormalTestJob("pg1", constIntNum2)
	nodes := ascendtest.FakeNormalTestNodes(1)
	ascendtest.SetNPUNodeLabel(testSsn.Nodes[nodes[0].Name].Node, nodeDEnableKey, nodeDEnableOnValue)

	var tmpPatche *gomonkey.Patches
	testCases := writeReSchedulerDataToCMTests{
		{
			name: "01-write data to config map success Cache-test Cache-test",
			args: writeReSchedulerDataToCMArgs{
				ssn: testSsn, reSchedulerData: ReSchedulerCache, cacheFunBefore: func() {
					initTestReSchedulerCache()
					addTestJobIntoReSchedulerCache(job1)
					addTestNodeIntoReSchedulerCache(nodes[0])
					addTestCardIntoReSchedulerCache(nodes[0].Name, nil, []string{"Ascend910-0", "Ascend910-1"})
					addTestHeartbeatIntoReSchedulerCache(nodes[0].Name)
					tmpPatche = gomonkey.ApplyFunc(createOrUpdateConfigMap,
						func(k8s kubernetes.Interface, cm *v1.ConfigMap) error {
							return nil
						})
				}, cacheFunAfter: func() {
					tmpPatche.Reset()
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
