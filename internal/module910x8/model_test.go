/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package module910x8 is using for HuaWei A800/9000 Ascend910 pin affinity schedule.

*/
package module910x8

import (
	"reflect"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
)

func testInsertNodeInPriGroup1PTasks(
	task *api.TaskInfo,
	priNodeGroups []map[string]*npuPriNodeInf,
	addPriNodeGroupFn initPriNodeGroupFn) {
	if len(priNodeGroups) < util.NPUIndex4 {
		return
	}
	convey.Convey("insertNodeInPriGroup() 1P should add node into the 1st group", func() {
		sNodeInf := selectNodeInf{nodeName: nodeName, allNPUNum: util.NPUIndex5, leftNPUNum: 1,
			rightNPUNum: util.NPUIndex4}
		err := insertNodeInPriGroup(task, sNodeInf, priNodeGroups, addPriNodeGroupFn)
		convey.So(err, convey.ShouldBeNil)
		convey.So(priNodeGroups[0][nodeName], convey.ShouldNotBeNil)
	})

	convey.Convey("insertNodeInPriGroup() 1P should add node into the 2nd group", func() {
		sNodeInf := selectNodeInf{nodeName: nodeName, allNPUNum: util.NPUIndex7, leftNPUNum: util.NPUIndex4,
			rightNPUNum: util.NPUIndex3}
		err := insertNodeInPriGroup(task, sNodeInf, priNodeGroups, addPriNodeGroupFn)
		convey.So(err, convey.ShouldBeNil)
		convey.So(priNodeGroups[1][nodeName], convey.ShouldNotBeNil)
	})

	convey.Convey("insertNodeInPriGroup() 1P should add node into the 3rd group", func() {
		sNodeInf := selectNodeInf{nodeName: nodeName, allNPUNum: util.NPUIndex6, leftNPUNum: util.NPUIndex2,
			rightNPUNum: util.NPUIndex4}
		err := insertNodeInPriGroup(task, sNodeInf, priNodeGroups, addPriNodeGroupFn)
		convey.So(err, convey.ShouldBeNil)
		convey.So(priNodeGroups[util.NPUIndex2][nodeName], convey.ShouldNotBeNil)
	})

	convey.Convey("insertNodeInPriGroup() 1P should add node into the 4th group", func() {
		sNodeInf := selectNodeInf{nodeName: nodeName, allNPUNum: util.NPUIndex4, leftNPUNum: 0,
			rightNPUNum: util.NPUIndex4}
		err := insertNodeInPriGroup(task, sNodeInf, priNodeGroups, addPriNodeGroupFn)
		convey.So(err, convey.ShouldBeNil)
		convey.So(priNodeGroups[util.NPUIndex3][nodeName], convey.ShouldNotBeNil)
	})

	convey.Convey("insertNodeInPriGroup() 1P should return error when neither hccl-ring has enough NPUs", func() {
		sNodeInf := selectNodeInf{nodeName: nodeName, allNPUNum: 0, leftNPUNum: 0, rightNPUNum: 0}
		err := insertNodeInPriGroup(task, sNodeInf, priNodeGroups, addPriNodeGroupFn)
		convey.So(err, convey.ShouldBeError)
	})
}

func testInsertNodeInPriGroup2PTasks(
	task *api.TaskInfo,
	priNodeGroups []map[string]*npuPriNodeInf,
	addPriNodeGroupFn initPriNodeGroupFn) {
	if len(priNodeGroups) < util.NPUIndex4 {
		return
	}
	convey.Convey("insertNodeInPriGroup() 2P should add node into the 1st group", func() {
		sNodeInf := selectNodeInf{nodeName: nodeName, allNPUNum: util.NPUIndex6, leftNPUNum: util.NPUIndex4,
			rightNPUNum: util.NPUIndex2}
		err := insertNodeInPriGroup(task, sNodeInf, priNodeGroups, addPriNodeGroupFn)
		convey.So(err, convey.ShouldBeNil)
		convey.So(priNodeGroups[0][nodeName], convey.ShouldNotBeNil)
	})

	convey.Convey("insertNodeInPriGroup() 2P should add node into the 2nd group", func() {
		sNodeInf := selectNodeInf{nodeName: nodeName, allNPUNum: nodeNPUNumber, leftNPUNum: util.NPUIndex4,
			rightNPUNum: util.NPUIndex4}
		err := insertNodeInPriGroup(task, sNodeInf, priNodeGroups, addPriNodeGroupFn)
		convey.So(err, convey.ShouldBeNil)
		convey.So(priNodeGroups[1][nodeName], convey.ShouldNotBeNil)
	})

	convey.Convey("insertNodeInPriGroup() 2P should add node into the 3rd group", func() {
		sNodeInf := selectNodeInf{nodeName: nodeName, allNPUNum: util.NPUIndex3, leftNPUNum: 0,
			rightNPUNum: util.NPUIndex3}
		err := insertNodeInPriGroup(task, sNodeInf, priNodeGroups, addPriNodeGroupFn)
		convey.So(err, convey.ShouldBeNil)
		convey.So(priNodeGroups[util.NPUIndex2][nodeName], convey.ShouldNotBeNil)
	})

	convey.Convey("insertNodeInPriGroup() 2P should return error when neither hccl-ring has enough NPUs", func() {
		sNodeInf := selectNodeInf{nodeName: nodeName, allNPUNum: util.NPUIndex2, leftNPUNum: 1, rightNPUNum: 1}
		err := insertNodeInPriGroup(task, sNodeInf, priNodeGroups, addPriNodeGroupFn)
		convey.So(err, convey.ShouldBeError)
	})
}

func testInsertNodeInPriGroup4PTasks(
	task *api.TaskInfo,
	priNodeGroups []map[string]*npuPriNodeInf,
	addPriNodeGroupFn initPriNodeGroupFn) {
	if len(priNodeGroups) < util.NPUIndex4 {
		return
	}
	convey.Convey("insertNodeInPriGroup() 4P should add node into 1st group", func() {
		sNodeInf := selectNodeInf{nodeName: nodeName, allNPUNum: util.NPUIndex6, leftNPUNum: util.NPUIndex2,
			rightNPUNum: util.NPUIndex4}
		err := insertNodeInPriGroup(task, sNodeInf, priNodeGroups, addPriNodeGroupFn)
		convey.So(err, convey.ShouldBeNil)
		convey.So(priNodeGroups[0][nodeName], convey.ShouldNotBeNil)
	})

	convey.Convey("insertNodeInPriGroup() 4P should return error when neither hccl-ring has enough NPUs", func() {
		sNodeInf := selectNodeInf{nodeName: nodeName, allNPUNum: util.NPUIndex4, leftNPUNum: util.NPUIndex2,
			rightNPUNum: util.NPUIndex2}
		err := insertNodeInPriGroup(task, sNodeInf, priNodeGroups, addPriNodeGroupFn)
		convey.So(err, convey.ShouldBeError)
	})
}

func testInsertNodeInPriGroup8PTasks(
	task *api.TaskInfo,
	priNodeGroups []map[string]*npuPriNodeInf,
	addPriNodeGroupFn initPriNodeGroupFn) {
	if len(priNodeGroups) < util.NPUIndex4 {
		return
	}
	convey.Convey("insertNodeInPriGroup() 8P should add node into 1st group", func() {
		sNodeInf := selectNodeInf{nodeName: nodeName, allNPUNum: nodeNPUNumber, leftNPUNum: util.NPUIndex4,
			rightNPUNum: util.NPUIndex4}
		err := insertNodeInPriGroup(task, sNodeInf, priNodeGroups, addPriNodeGroupFn)
		convey.So(err, convey.ShouldBeNil)
		convey.So(priNodeGroups[0][nodeName], convey.ShouldNotBeNil)
	})

	convey.Convey("insertNodeInPriGroup() 8P should return error when node doesn't have enough NPUs", func() {
		sNodeInf := selectNodeInf{nodeName: nodeName, allNPUNum: util.NPUIndex5, leftNPUNum: 1,
			rightNPUNum: util.NPUIndex4}
		err := insertNodeInPriGroup(task, sNodeInf, priNodeGroups, addPriNodeGroupFn)
		convey.So(err, convey.ShouldBeError)
	})
}

// TestMNPUInsertNodeInPriGroup
func TestMNPUInsertNodeInPriGroup(t *testing.T) {
	convey.Convey("Test module910x8 insertNodeInPriGroup", t, func() {
		var priNodeGroups []map[string]*npuPriNodeInf
		for i := 0; i < npuNumPerHccs; i++ {
			priNodeGroups = append(priNodeGroups, make(map[string]*npuPriNodeInf, 1))
		}
		addPriNodeGroupFn := func(priNodeGroup map[string]*npuPriNodeInf, groupName string) {
			priNodeGroup[nodeName] = &npuPriNodeInf{
				Name:     groupName,
				nodeName: nodeName,
			}
		}
		task1p := api.NewTaskInfo(buildNPUPod(MPodInfo{namespace: "default", groupName: "group-M-model-1p",
			podName: "npu-test-M-model-1p", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: npu800And9000CardName, reqNpuNum: "1"}))
		task2p := api.NewTaskInfo(buildNPUPod(MPodInfo{namespace: "default", groupName: "group-M-model-2p",
			podName: "npu-test-M-model-2p", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: npu800And9000CardName, reqNpuNum: "2"}))
		task4p := api.NewTaskInfo(buildNPUPod(MPodInfo{namespace: "default", groupName: "group-M-model-4p",
			podName: "npu-test-M-model-4p", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: npu800And9000CardName, reqNpuNum: "4"}))
		task8p := api.NewTaskInfo(buildNPUPod(MPodInfo{namespace: "default", groupName: "group-M-model-8p",
			podName: "npu-test-M-model-8p", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: npu800And9000CardName, reqNpuNum: "8"}))

		testInsertNodeInPriGroup1PTasks(task1p, priNodeGroups, addPriNodeGroupFn)
		testInsertNodeInPriGroup2PTasks(task2p, priNodeGroups, addPriNodeGroupFn)
		testInsertNodeInPriGroup4PTasks(task4p, priNodeGroups, addPriNodeGroupFn)
		testInsertNodeInPriGroup8PTasks(task8p, priNodeGroups, addPriNodeGroupFn)

		task7p := api.NewTaskInfo(buildNPUPod(MPodInfo{namespace: "default", groupName: "group-M-model-7p",
			podName: "npu-test-M-model-7p", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: npu800And9000CardName, reqNpuNum: "7"}))
		convey.Convey("insertNodeInPriGroup() 7P should return error", func() {
			sNodeInf := selectNodeInf{nodeName: nodeName, allNPUNum: nodeNPUNumber, leftNPUNum: util.NPUIndex4,
				rightNPUNum: util.NPUIndex4}
			err := insertNodeInPriGroup(task7p, sNodeInf, priNodeGroups, addPriNodeGroupFn)
			convey.So(err, convey.ShouldBeError)
		})
	})
}

// TestMNPUJudgeNodeAndTaskNPU
func TestMNPUJudgeNodeAndTaskNPU(t *testing.T) {
	convey.Convey("Test module910x8 judgeNodeAndTaskNPU", t, func() {
		convey.Convey("judgeNodeAndTaskNPU() should return nil when taskNPU is 0", func() {
			err := judgeNodeAndTaskNPU(0, []int{0, 1, util.NPUIndex2, util.NPUIndex3})
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("judgeNodeAndTaskNPU() should return err when node doesn't satisfy req of 1", func() {
			err := judgeNodeAndTaskNPU(1, []int{})
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("judgeNodeAndTaskNPU() should return nil when node satisfies req of 2", func() {
			err := judgeNodeAndTaskNPU(util.NPUIndex2, []int{0, 1})
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("judgeNodeAndTaskNPU() should return nil when node doesn't satisfy req of 4", func() {
			err := judgeNodeAndTaskNPU(util.NPUIndex4, []int{util.NPUIndex4, util.NPUIndex5, util.NPUIndex6})
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("judgeNodeAndTaskNPU() should return nil when node satisfies req of 8", func() {
			err := judgeNodeAndTaskNPU(nodeNPUNumber, []int{0, 1, util.NPUIndex2, util.NPUIndex3,
				util.NPUIndex4, util.NPUIndex5, util.NPUIndex6, util.NPUIndex7})
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("judgeNodeAndTaskNPU() should return er when req num is invalid", func() {
			err := judgeNodeAndTaskNPU(util.NPUIndex7, []int{0, 1, util.NPUIndex2, util.NPUIndex3,
				util.NPUIndex4, util.NPUIndex5, util.NPUIndex6, util.NPUIndex7})
			convey.So(err, convey.ShouldBeError)
		})
	})
}

type IsUnstableNodeMeetTaskReqNPUSourceArgs struct {
	task *api.TaskInfo
	node *api.NodeInfo
}

type IsUnstableNodeMeetTaskReqNPUSourceTests struct {
	name string
	args IsUnstableNodeMeetTaskReqNPUSourceArgs
	want bool
}

func buildIsUnstableNodeMeetTaskReqNPUSourceTestCases() []IsUnstableNodeMeetTaskReqNPUSourceTests {
	nodeInfo0 := test.FakeNormalTestNode("node0")
	taskInfo0 := test.FakeNormalTestTask("task0", "node0", "pg0")
	test.AddFakeTaskResReq(taskInfo0, npu310CardName, util.NPUIndex2*util.NPUHex)

	nodeInfo1 := test.FakeNormalTestNode("node1")
	taskInfo1 := test.FakeNormalTestTask("task1", "node1", "pg1")
	test.AddFakeTaskResReq(taskInfo1, npu800And9000CardName, util.NPUIndex2*util.NPUHex)

	nodeInfo2 := test.FakeNormalTestNode("node2")
	taskInfo2 := test.FakeNormalTestTask("task2", "node2", "pg2")
	test.AddFakeTaskResReq(taskInfo2, npu800And9000CardName, util.NPUIndex2*util.NPUHex)
	test.SetFakeNodeIdleSource(nodeInfo2, npu800And9000CardName, util.NPUIndex8)

	nodeInfo3 := test.FakeNormalTestNode("node3")
	taskInfo3 := test.FakeNormalTestTask("task3", "node3", "pg3")
	test.AddFakeTaskResReq(taskInfo3, npu800And9000CardName, util.NPUIndex4*util.NPUHex)
	test.SetFakeNodeIdleSource(nodeInfo3, npu800And9000CardName, util.NPUIndex2)

	testCases := []IsUnstableNodeMeetTaskReqNPUSourceTests{
		{
			name: "test-IsUnstableNodeMeetTaskReqNPUSource()\ncase0: task require 310 npu thus return false",
			args: IsUnstableNodeMeetTaskReqNPUSourceArgs{
				task: taskInfo0,
				node: nodeInfo0,
			},
			want: false,
		},
		{
			name: "test-IsUnstableNodeMeetTaskReqNPUSource()\ncase1: no idle resource on node",
			args: IsUnstableNodeMeetTaskReqNPUSourceArgs{
				task: taskInfo1,
				node: nodeInfo1,
			},
			want: false,
		},
		{
			name: "test-IsUnstableNodeMeetTaskReqNPUSource()\ncase2: success",
			args: IsUnstableNodeMeetTaskReqNPUSourceArgs{
				task: taskInfo2,
				node: nodeInfo2,
			},
			want: true,
		},
		{
			name: "test-IsUnstableNodeMeetTaskReqNPUSource()\ncase3: node idle npu less than task require",
			args: IsUnstableNodeMeetTaskReqNPUSourceArgs{
				task: taskInfo3,
				node: nodeInfo3,
			},
			want: false,
		},
	}
	return testCases
}

func TestIsUnstableNodeMeetTaskReqNPUSource(t *testing.T) {
	tests := buildIsUnstableNodeMeetTaskReqNPUSourceTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isUnstableNodeMeetTaskReqNPUSource(tt.args.task, tt.args.node)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("isUnstableNodeMeetTaskReqNPUSource() = %v, want %v", got, tt.want)
			}
		})
	}
}
