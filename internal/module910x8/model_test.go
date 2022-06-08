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

func TestMNPUInsertNodeInPriGroup1PTasks(t *testing.T) {
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

		convey.Convey("insertNodeInPriGroup() 1P should add node into the 1st group", func() {
			sNodeInf := selectNodeInf{nodeName: nodeName, allNPUNum: constIntNum5, leftNPUNum: 1,
				rightNPUNum: constIntNum4}
			err := insertNodeInPriGroup(task1p, sNodeInf, priNodeGroups, addPriNodeGroupFn)
			convey.So(err, convey.ShouldBeNil)
			convey.So(priNodeGroups[0][nodeName], convey.ShouldNotBeNil)
		})

		convey.Convey("insertNodeInPriGroup() 1P should add node into the 2nd group", func() {
			sNodeInf := selectNodeInf{nodeName: nodeName, allNPUNum: constIntNum7, leftNPUNum: constIntNum4,
				rightNPUNum: constIntNum3}
			err := insertNodeInPriGroup(task1p, sNodeInf, priNodeGroups, addPriNodeGroupFn)
			convey.So(err, convey.ShouldBeNil)
			convey.So(priNodeGroups[1][nodeName], convey.ShouldNotBeNil)
		})

		convey.Convey("insertNodeInPriGroup() 1P should add node into the 3rd group", func() {
			sNodeInf := selectNodeInf{nodeName: nodeName, allNPUNum: constIntNum6, leftNPUNum: constIntNum2,
				rightNPUNum: constIntNum4}
			err := insertNodeInPriGroup(task1p, sNodeInf, priNodeGroups, addPriNodeGroupFn)
			convey.So(err, convey.ShouldBeNil)
			convey.So(priNodeGroups[constIntNum2][nodeName], convey.ShouldNotBeNil)
		})

		convey.Convey("insertNodeInPriGroup() 1P should add node into the 4th group", func() {
			sNodeInf := selectNodeInf{nodeName: nodeName, allNPUNum: constIntNum4, leftNPUNum: 0,
				rightNPUNum: constIntNum4}
			err := insertNodeInPriGroup(task1p, sNodeInf, priNodeGroups, addPriNodeGroupFn)
			convey.So(err, convey.ShouldBeNil)
			convey.So(priNodeGroups[constIntNum3][nodeName], convey.ShouldNotBeNil)
		})

		convey.Convey("insertNodeInPriGroup() 1P should return error when neither hccl-ring has enough NPUs", func() {
			sNodeInf := selectNodeInf{nodeName: nodeName, allNPUNum: 0, leftNPUNum: 0, rightNPUNum: 0}
			err := insertNodeInPriGroup(task1p, sNodeInf, priNodeGroups, addPriNodeGroupFn)
			convey.So(err, convey.ShouldBeError)
		})
	})
}

func TestMNPUInsertNodeInPriGroup2PTasks(t *testing.T) {
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
		task2p := api.NewTaskInfo(buildNPUPod(MPodInfo{namespace: "default", groupName: "group-M-model-2p",
			podName: "npu-test-M-model-2p", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: npu800And9000CardName, reqNpuNum: "2"}))

		convey.Convey("insertNodeInPriGroup() 2P should add node into the 1st group", func() {
			sNodeInf := selectNodeInf{nodeName: nodeName, allNPUNum: constIntNum6, leftNPUNum: constIntNum4,
				rightNPUNum: constIntNum2}
			err := insertNodeInPriGroup(task2p, sNodeInf, priNodeGroups, addPriNodeGroupFn)
			convey.So(err, convey.ShouldBeNil)
			convey.So(priNodeGroups[0][nodeName], convey.ShouldNotBeNil)
		})

		convey.Convey("insertNodeInPriGroup() 2P should add node into the 2nd group", func() {
			sNodeInf := selectNodeInf{nodeName: nodeName, allNPUNum: nodeNPUNumber, leftNPUNum: constIntNum4,
				rightNPUNum: constIntNum4}
			err := insertNodeInPriGroup(task2p, sNodeInf, priNodeGroups, addPriNodeGroupFn)
			convey.So(err, convey.ShouldBeNil)
			convey.So(priNodeGroups[1][nodeName], convey.ShouldNotBeNil)
		})

		convey.Convey("insertNodeInPriGroup() 2P should add node into the 3rd group", func() {
			sNodeInf := selectNodeInf{nodeName: nodeName, allNPUNum: constIntNum3, leftNPUNum: 0,
				rightNPUNum: constIntNum3}
			err := insertNodeInPriGroup(task2p, sNodeInf, priNodeGroups, addPriNodeGroupFn)
			convey.So(err, convey.ShouldBeNil)
			convey.So(priNodeGroups[constIntNum2][nodeName], convey.ShouldNotBeNil)
		})

		convey.Convey("insertNodeInPriGroup() 2P should return error when neither hccl-ring has enough NPUs", func() {
			sNodeInf := selectNodeInf{nodeName: nodeName, allNPUNum: constIntNum2, leftNPUNum: 1, rightNPUNum: 1}
			err := insertNodeInPriGroup(task2p, sNodeInf, priNodeGroups, addPriNodeGroupFn)
			convey.So(err, convey.ShouldBeError)
		})
	})
}

func TestMNPUInsertNodeInPriGroup4PTasks(t *testing.T) {
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
		task4p := api.NewTaskInfo(buildNPUPod(MPodInfo{namespace: "default", groupName: "group-M-model-4p",
			podName: "npu-test-M-model-4p", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: npu800And9000CardName, reqNpuNum: "4"}))

		convey.Convey("insertNodeInPriGroup() 4P should add node into 1st group", func() {
			sNodeInf := selectNodeInf{nodeName: nodeName, allNPUNum: constIntNum6, leftNPUNum: constIntNum2,
				rightNPUNum: constIntNum4}
			err := insertNodeInPriGroup(task4p, sNodeInf, priNodeGroups, addPriNodeGroupFn)
			convey.So(err, convey.ShouldBeNil)
			convey.So(priNodeGroups[0][nodeName], convey.ShouldNotBeNil)
		})

		convey.Convey("insertNodeInPriGroup() 4P should return error when neither hccl-ring has enough NPUs", func() {
			sNodeInf := selectNodeInf{nodeName: nodeName, allNPUNum: constIntNum4, leftNPUNum: constIntNum2,
				rightNPUNum: constIntNum2}
			err := insertNodeInPriGroup(task4p, sNodeInf, priNodeGroups, addPriNodeGroupFn)
			convey.So(err, convey.ShouldBeError)
		})
	})
}

func TestMNPUInsertNodeInPriGroup8PTasks(t *testing.T) {
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
		task8p := api.NewTaskInfo(buildNPUPod(MPodInfo{namespace: "default", groupName: "group-M-model-8p",
			podName: "npu-test-M-model-8p", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: npu800And9000CardName, reqNpuNum: "8"}))

		convey.Convey("insertNodeInPriGroup() 8P should add node into 1st group", func() {
			sNodeInf := selectNodeInf{nodeName: nodeName, allNPUNum: nodeNPUNumber, leftNPUNum: constIntNum4,
				rightNPUNum: constIntNum4}
			err := insertNodeInPriGroup(task8p, sNodeInf, priNodeGroups, addPriNodeGroupFn)
			convey.So(err, convey.ShouldBeNil)
			convey.So(priNodeGroups[0][nodeName], convey.ShouldNotBeNil)
		})

		convey.Convey("insertNodeInPriGroup() 8P should return error when node doesn't have enough NPUs", func() {
			sNodeInf := selectNodeInf{nodeName: nodeName, allNPUNum: constIntNum5, leftNPUNum: 1,
				rightNPUNum: constIntNum4}
			err := insertNodeInPriGroup(task8p, sNodeInf, priNodeGroups, addPriNodeGroupFn)
			convey.So(err, convey.ShouldBeError)
		})
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
		task7p := api.NewTaskInfo(buildNPUPod(MPodInfo{namespace: "default", groupName: "group-M-model-7p",
			podName: "npu-test-M-model-7p", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: npu800And9000CardName, reqNpuNum: "7"}))
		convey.Convey("insertNodeInPriGroup() 7P should return error", func() {
			sNodeInf := selectNodeInf{nodeName: nodeName, allNPUNum: nodeNPUNumber, leftNPUNum: constIntNum4,
				rightNPUNum: constIntNum4}
			err := insertNodeInPriGroup(task7p, sNodeInf, priNodeGroups, addPriNodeGroupFn)
			convey.So(err, convey.ShouldBeError)
		})
	})
}

// TestMNPUJudgeNodeAndTaskNPU
func TestMNPUJudgeNodeAndTaskNPU(t *testing.T) {
	convey.Convey("Test module910x8 judgeNodeAndTaskNPU", t, func() {
		convey.Convey("judgeNodeAndTaskNPU() should return nil when taskNPU is 0", func() {
			err := judgeNodeAndTaskNPU(0, []int{0, 1, constIntNum2, constIntNum3})
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("judgeNodeAndTaskNPU() should return err when node doesn't satisfy req of 1", func() {
			err := judgeNodeAndTaskNPU(1, []int{})
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("judgeNodeAndTaskNPU() should return nil when node satisfies req of 2", func() {
			err := judgeNodeAndTaskNPU(constIntNum2, []int{0, 1})
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("judgeNodeAndTaskNPU() should return nil when node doesn't satisfy req of 4", func() {
			err := judgeNodeAndTaskNPU(constIntNum4, []int{constIntNum4, constIntNum5, constIntNum6})
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("judgeNodeAndTaskNPU() should return nil when node satisfies req of 8", func() {
			err := judgeNodeAndTaskNPU(nodeNPUNumber, []int{0, constIntNum1, constIntNum2, constIntNum3,
				constIntNum4, constIntNum5, constIntNum6, constIntNum7})
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("judgeNodeAndTaskNPU() should return er when req num is invalid", func() {
			err := judgeNodeAndTaskNPU(constIntNum7, []int{0, constIntNum1, constIntNum2, constIntNum3,
				constIntNum4, constIntNum5, constIntNum6, constIntNum7})
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
	test.AddFakeTaskResReq(taskInfo0, npu310CardName, constIntNum2*util.NPUHex)

	nodeInfo1 := test.FakeNormalTestNode("node1")
	taskInfo1 := test.FakeNormalTestTask("task1", "node1", "pg1")
	test.AddFakeTaskResReq(taskInfo1, npu800And9000CardName, constIntNum2*util.NPUHex)

	nodeInfo2 := test.FakeNormalTestNode("node2")
	taskInfo2 := test.FakeNormalTestTask("task2", "node2", "pg2")
	test.AddFakeTaskResReq(taskInfo2, npu800And9000CardName, constIntNum2*util.NPUHex)
	test.SetFakeNodeIdleSource(nodeInfo2, npu800And9000CardName, constIntNum8)

	nodeInfo3 := test.FakeNormalTestNode("node3")
	taskInfo3 := test.FakeNormalTestTask("task3", "node3", "pg3")
	test.AddFakeTaskResReq(taskInfo3, npu800And9000CardName, constIntNum4*util.NPUHex)
	test.SetFakeNodeIdleSource(nodeInfo3, npu800And9000CardName, constIntNum2)

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
