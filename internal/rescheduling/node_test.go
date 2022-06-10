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

	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
)

type getDistributeUsableNPUTopArgs struct {
	nodeNPUTopology   []int
	netUnhealthyCards []int
}

type getDistributeUsableNPUTopTest struct {
	name string
	args getDistributeUsableNPUTopArgs
	want []int
}

func buildGetDistributeUsableNPUTopTestCases() []getDistributeUsableNPUTopTest {
	const constIntNum7 = 7
	testCases := []getDistributeUsableNPUTopTest{
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
				nodeNPUTopology:   []int{0, 1, util.NPUIndex2, util.NPUIndex3, constIntNum7},
				netUnhealthyCards: []int{0, 1},
			},
			want: []int{util.NPUIndex2, util.NPUIndex3, constIntNum7},
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

type getFaultTaskUseNodeInfoArgs struct {
	task     *api.TaskInfo
	ssn      *framework.Session
	cacheFun func()
}

type getFaultTaskUseNodeInfoTest struct {
	name    string
	args    getFaultTaskUseNodeInfoArgs
	wantErr error
}

func buildGetFaultTaskUseNodeInfoTestCases() []getFaultTaskUseNodeInfoTest {
	const constNum4, constNum3 = 4, 3
	tasks := test.FakeNormalTestTasks(constNum4)
	nodes := test.FakeNormalTestNodes(constNum4)
	ssnTest := test.FakeNormalSSN()
	testCases := []getFaultTaskUseNodeInfoTest{
		{
			name: "01-task not in fault task list-test",
			args: getFaultTaskUseNodeInfoArgs{
				task: tasks[0], ssn: nil, cacheFun: func() {
					initTestReSchedulerCache()
					addTestTaskIntoReSchedulerCache()
				},
			}, wantErr: fmt.Errorf("no %v in jobMap", tasks[0].Job),
		},
		{
			name: "02-task use node not ssn node list-test",
			args: getFaultTaskUseNodeInfoArgs{
				task: tasks[constNum3], ssn: ssnTest, cacheFun: func() {
					initTestReSchedulerCache()
					addTestTaskIntoReSchedulerCache(tasks[constNum3])
				},
			}, wantErr: fmt.Errorf("get node name %s failed", tasks[constNum3].NodeName),
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
			wantErr: fmt.Errorf("get fault task used node info %s in fault node list",
				tasks[0].NodeName),
		},
		{
			name: "04-task use node which not fault resource-test",
			args: getFaultTaskUseNodeInfoArgs{
				task: tasks[0], ssn: ssnTest, cacheFun: func() {
					initTestReSchedulerCache()
					addTestTaskIntoReSchedulerCache(tasks[0])
					addTestNodeIntoReSchedulerCache(nodes[constNum3])
				},
			}, wantErr: nil,
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

type getNetworkUnhealthyCardsTest struct {
	name string
	args getNetworkUnhealthyCardsArgs
	want []int
}

func buildTestGetNetworkUnhealthyCardsTestCases() []getNetworkUnhealthyCardsTest {
	const constNum4 = 4
	nodes := test.FakeNormalTestNodes(constNum4)
	testCases := []getNetworkUnhealthyCardsTest{
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

type isNodeInFaultNodeListArgs struct {
	node     *api.NodeInfo
	cacheFun func()
}

type isNodeInFaultNodeListTest struct {
	name string
	args isNodeInFaultNodeListArgs
	want bool
}

func buildIsNodeInFaultNodeListTestCases() []isNodeInFaultNodeListTest {
	nodes := test.FakeNormalTestNodes(util.NPUIndex2)
	testCases := []isNodeInFaultNodeListTest{
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
