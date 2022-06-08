/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package module910x8 is using for HuaWei A800/9000 Ascend910 pin affinity schedule.

*/
package module910x8

import (
	"fmt"
	"reflect"
	"testing"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/util"

	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
)

type getCheckNPUResourceStableArgs struct {
	vNode *api.NodeInfo
}

type getCheckNPUResourceStableMapTest struct {
	name    string
	args    getCheckNPUResourceStableArgs
	wantErr error
}

func buildCheckNPUResourceStableTestCases() []getCheckNPUResourceStableMapTest {
	const npuNum = 2
	nodeInfo0 := test.BuildUnstableNode("node0", npu800And9000CardName, "", 0)
	nodeInfo1 := test.BuildUnstableNode("node1", npu800And9000CardName, "Ascend910-2", 0)
	nodeInfo2 := test.BuildUnstableNode("node2", npu800And9000CardName, "Ascend910-2", npuNum)
	nodeInfo3 := test.BuildUnstableNode("node3", npu800And9000CardName, "Ascend910-2,Ascend910-3", npuNum)
	testCases := []getCheckNPUResourceStableMapTest{
		{
			name:    "test0-CheckNPUResourceStable()\ncase0: insufficient npu",
			args:    getCheckNPUResourceStableArgs{vNode: nodeInfo0},
			wantErr: fmt.Errorf("getNodeNPUNumFromOthers %s : nil node(%s) top", nodesNoMeetNPUReqError, nodeInfo0.Name),
		},
		{
			name:    "test0-CheckNPUResourceStable()\ncase1: got no idle npu",
			args:    getCheckNPUResourceStableArgs{vNode: nodeInfo1},
			wantErr: fmt.Errorf("getNodeNPUNumFromIdle %s : get node idle npu failed", nodesNoMeetNPUReqError),
		},
		{
			name:    "test0-CheckNPUResourceStable()\ncase2: resource mismatch between allocatable and idle npu",
			args:    getCheckNPUResourceStableArgs{vNode: nodeInfo2},
			wantErr: fmt.Errorf("%s : node not stable for annotations(1) : idle(2)", nodeNotStableWarning),
		},
		{
			name:    "test0-CheckNPUResourceStable()\ncase3: resource mismatch between allocatable and idle npu",
			args:    getCheckNPUResourceStableArgs{vNode: nodeInfo3},
			wantErr: nil,
		},
	}
	return testCases
}

func TestCheckNPUResourceStable(t *testing.T) {
	tests := buildCheckNPUResourceStableTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkNPUResourceStable(tt.args.vNode)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("checkNPUResourceStable() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type clusterNodePredicateFnArgs struct {
	task     *api.TaskInfo
	ssn      *framework.Session
	cacheFun func()
}

type clusterNodePredicateFnTests struct {
	name    string
	args    clusterNodePredicateFnArgs
	wantErr error
}

func buildClusterNodePredicateFnTestCases() []clusterNodePredicateFnTests {
	task0 := test.FakeNormalTestTask("task0", "node0", "pg0")
	test.AddFakeTaskResReq(task0, "any other name", constIntNum1*util.NPUHex)
	ssn0 := test.FakeNormalSSN()

	task1 := test.FakeNormalTestTask("task1", "node1", "pg1")
	test.AddFakeTaskResReq(task1, npu800And9000CardName, constIntNum1*util.NPUHex)
	ssn1 := test.FakeNormalSSN()

	testCases := []clusterNodePredicateFnTests{
		{
			name:    "test3-ClusterNodePredicateFn()\ncase0: not a 910 task(return in branch 1)",
			args:    clusterNodePredicateFnArgs{task: task0, ssn: ssn0, cacheFun: func() {}},
			wantErr: nil,
		},
		{
			name:    "test3-ClusterNodePredicateFn()\ncase1: is a 910 task not an NPU fault task(return in branch 2)",
			args:    clusterNodePredicateFnArgs{task: task1, ssn: ssn1, cacheFun: func() {}},
			wantErr: nil,
		},
	}
	return testCases
}

func TestClusterNodePredicateFn(t *testing.T) {
	tests := buildClusterNodePredicateFnTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.cacheFun()
			err := clusterNodePredicateFn(tt.args.task, tt.args.ssn)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("clusterNodePredicateFn() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type getNodeHccsArrayArgs struct {
	nodeTop []int
}

type getNodeHccsArrayTests struct {
	name  string
	args  getNodeHccsArrayArgs
	want  []int
	want1 []int
}

func buildGetNodeHccsArrayTestCases() []getNodeHccsArrayTests {
	testCases := []getNodeHccsArrayTests{
		{
			name:  "test1-getNodeHccsArray()\ncase0: split topology to both sides",
			args:  getNodeHccsArrayArgs{nodeTop: []int{constIntNum1, constIntNum3, constIntNum5}},
			want:  []int{constIntNum1, constIntNum3},
			want1: []int{constIntNum5},
		},
		{
			name:  "test1-getNodeHccsArrayTest\ncase1: split topology to one side",
			args:  getNodeHccsArrayArgs{nodeTop: []int{constIntNum3, constIntNum2}},
			want:  []int{constIntNum3, constIntNum2},
			want1: nil,
		},
		{
			name:  "test1-getNodeHccsArrayTest\ncase2: split topology to neither side",
			args:  getNodeHccsArrayArgs{nodeTop: []int{}},
			want:  nil,
			want1: nil,
		},
	}
	return testCases
}

func TestGetNodeHccsArray(t *testing.T) {
	tests := buildGetNodeHccsArrayTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := getNodeHccsArray(tt.args.nodeTop)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getNodeHccsArray() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("getNodeHccsArray() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

type getNodeNPUNumFromOthersArgs struct {
	nodeInfo *api.NodeInfo
}

type getNodeNPUNumFromOthersTests struct {
	name    string
	args    getNodeNPUNumFromOthersArgs
	want    int
	wantErr error
}

func buildGetNodeNPUNumFromOthersTestCases() []getNodeNPUNumFromOthersTests {
	nodeInfo0 := test.FakeNormalTestNode("node0")
	nodeInfo1 := test.FakeNormalTestNode("node1")
	nodeInfo2 := test.FakeNormalTestNode("node2")
	test.SetTestNPUNodeAnnotation(nodeInfo0, npu800And9000CardName, "Ascend910-5,Ascend910-6,Ascend910-7")
	test.SetTestNPUNodeAnnotation(nodeInfo1, npu800And9000CardName, "")
	test.SetTestNPUNodeAnnotation(nodeInfo2, npu800And9000CardName, "Ascend910-5,Ascend910-6,Ascend910-7,"+
		"Ascend910-0,Ascend910-1,Ascend910-2,Ascend910-3,Ascend910-4,Ascend910-8")
	testCases := []getNodeNPUNumFromOthersTests{
		{
			name:    "test2-getNodeNPUNumFromOthers()\ncase0: return number of devices",
			args:    getNodeNPUNumFromOthersArgs{nodeInfo: nodeInfo0},
			want:    constIntNum3,
			wantErr: nil,
		},
		{
			name:    "test2-getNodeNPUNumFromOthers()\ncase1: return error owing to no device",
			args:    getNodeNPUNumFromOthersArgs{nodeInfo: nodeInfo1},
			want:    constIntNum0,
			wantErr: nil,
		},
		{
			name: "test2-getNodeNPUNumFromOthers()\ncase2: return error owing to device number exceed maxNum(8)",
			args: getNodeNPUNumFromOthersArgs{nodeInfo: nodeInfo2},
			want: constIntNum0,
			wantErr: fmt.Errorf("amount of npus exceeded the limitation, maximum(%d), actual(9)",
				maxNPUNum),
		},
	}
	return testCases
}

func TestGetNodeNPUNumFromOthers(t *testing.T) {
	tests := buildGetNodeNPUNumFromOthersTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getNodeNPUNumFromOthers(tt.args.nodeInfo)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("getNodeNPUNumFromOthers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getNodeNPUNumFromOthers() got = %v, want %v", got, tt.want)
			}
		})
	}
}

type initNodesNPUTopologyFnArgs struct {
	nodeInfo map[string]*api.NodeInfo
}

type initNodesNPUTopologyFnTests struct {
	name    string
	args    initNodesNPUTopologyFnArgs
	wantErr error
}

func buildInitNodesNPUTopologyFnTestCases() []initNodesNPUTopologyFnTests {
	node0 := test.FakeNormalTestNode("node0")
	test.SetNPUNodeLabel(node0.Node, archSelector, huaweiArchArm)
	test.SetNPUNodeLabel(node0.Node, acceleratorType, cardAcceleratorType)

	node1 := test.FakeNormalTestNode("node1")
	test.SetNPUNodeLabel(node1.Node, archSelector, huaweiArchArm)

	node2 := test.FakeNormalTestNode("node2")
	test.SetNPUNodeLabel(node2.Node, archSelector, huaweiArchArm)
	test.SetTestNPUNodeAnnotation(node2, npu800And9000CardName, "Ascend910-0")

	testCases := []initNodesNPUTopologyFnTests{
		{
			name:    "test4-initNodesNPUTopologyFnArgs()\ncase0: not module type, continue and return",
			args:    initNodesNPUTopologyFnArgs{nodeInfo: map[string]*api.NodeInfo{"n0": node0}},
			wantErr: nil,
		},
		{
			name:    "test4-initNodesNPUTopologyFnArgs()\ncase0: break when got no 910npu",
			args:    initNodesNPUTopologyFnArgs{nodeInfo: map[string]*api.NodeInfo{"n0": node0, "n1": node1, "n2": node2}},
			wantErr: nil,
		},
		{
			name:    "test4-initNodesNPUTopologyFnArgs()\ncase1: test success",
			args:    initNodesNPUTopologyFnArgs{nodeInfo: map[string]*api.NodeInfo{"n0": node0, "n2": node2}},
			wantErr: nil,
		},
	}
	return testCases
}

func TestInitNodesNPUTopologyFn(t *testing.T) {
	tests := buildInitNodesNPUTopologyFnTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := initNodesNPUTopologyFn(tt.args.nodeInfo)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("initNodesNPUTopologyFn() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
