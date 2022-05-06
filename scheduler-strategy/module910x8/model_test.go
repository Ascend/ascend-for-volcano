/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package module910x8 is using for HuaWei A800/9000 Ascend910 pin affinity schedule.

*/
package module910x8

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	vapi "volcano.sh/volcano/pkg/scheduler/api"
)

func testInsertNodeInPriGroup1PTasks(
	task *vapi.TaskInfo,
	priNodeGroups []map[string]*npuPriNodeInf,
	addPriNodeGroupFn initPriNodeGroupFn) {
	Convey("insertNodeInPriGroup() 1P should add node into the 1st group", func() {
		sNodeInf := selectNodeInf{nodeName: nodeName, allNPUNum: 5, leftNPUNum: 1, rightNPUNum: 4}
		err := insertNodeInPriGroup(task, sNodeInf, priNodeGroups, addPriNodeGroupFn)
		So(err, ShouldBeNil)
		So(priNodeGroups[0][nodeName], ShouldNotBeNil)
	})

	Convey("insertNodeInPriGroup() 1P should add node into the 2nd group", func() {
		sNodeInf := selectNodeInf{nodeName: nodeName, allNPUNum: 7, leftNPUNum: 4, rightNPUNum: 3}
		err := insertNodeInPriGroup(task, sNodeInf, priNodeGroups, addPriNodeGroupFn)
		So(err, ShouldBeNil)
		So(priNodeGroups[1][nodeName], ShouldNotBeNil)
	})

	Convey("insertNodeInPriGroup() 1P should add node into the 3rd group", func() {
		sNodeInf := selectNodeInf{nodeName: nodeName, allNPUNum: 6, leftNPUNum: 2, rightNPUNum: 4}
		err := insertNodeInPriGroup(task, sNodeInf, priNodeGroups, addPriNodeGroupFn)
		So(err, ShouldBeNil)
		So(priNodeGroups[2][nodeName], ShouldNotBeNil)
	})

	Convey("insertNodeInPriGroup() 1P should add node into the 4th group", func() {
		sNodeInf := selectNodeInf{nodeName: nodeName, allNPUNum: 4, leftNPUNum: 0, rightNPUNum: 4}
		err := insertNodeInPriGroup(task, sNodeInf, priNodeGroups, addPriNodeGroupFn)
		So(err, ShouldBeNil)
		So(priNodeGroups[3][nodeName], ShouldNotBeNil)
	})

	Convey("insertNodeInPriGroup() 1P should return error when neither hccl-ring has enough NPUs", func() {
		sNodeInf := selectNodeInf{nodeName: nodeName, allNPUNum: 0, leftNPUNum: 0, rightNPUNum: 0}
		err := insertNodeInPriGroup(task, sNodeInf, priNodeGroups, addPriNodeGroupFn)
		So(err, ShouldBeError)
	})
}

func testInsertNodeInPriGroup2PTasks(
	task *vapi.TaskInfo,
	priNodeGroups []map[string]*npuPriNodeInf,
	addPriNodeGroupFn initPriNodeGroupFn) {
	Convey("insertNodeInPriGroup() 2P should add node into the 1st group", func() {
		sNodeInf := selectNodeInf{nodeName: nodeName, allNPUNum: 6, leftNPUNum: 4, rightNPUNum: 2}
		err := insertNodeInPriGroup(task, sNodeInf, priNodeGroups, addPriNodeGroupFn)
		So(err, ShouldBeNil)
		So(priNodeGroups[0][nodeName], ShouldNotBeNil)
	})

	Convey("insertNodeInPriGroup() 2P should add node into the 2nd group", func() {
		sNodeInf := selectNodeInf{nodeName: nodeName, allNPUNum: 8, leftNPUNum: 4, rightNPUNum: 4}
		err := insertNodeInPriGroup(task, sNodeInf, priNodeGroups, addPriNodeGroupFn)
		So(err, ShouldBeNil)
		So(priNodeGroups[1][nodeName], ShouldNotBeNil)
	})

	Convey("insertNodeInPriGroup() 2P should add node into the 3rd group", func() {
		sNodeInf := selectNodeInf{nodeName: nodeName, allNPUNum: 3, leftNPUNum: 0, rightNPUNum: 3}
		err := insertNodeInPriGroup(task, sNodeInf, priNodeGroups, addPriNodeGroupFn)
		So(err, ShouldBeNil)
		So(priNodeGroups[2][nodeName], ShouldNotBeNil)
	})

	Convey("insertNodeInPriGroup() 2P should return error when neither hccl-ring has enough NPUs", func() {
		sNodeInf := selectNodeInf{nodeName: nodeName, allNPUNum: 2, leftNPUNum: 1, rightNPUNum: 1}
		err := insertNodeInPriGroup(task, sNodeInf, priNodeGroups, addPriNodeGroupFn)
		So(err, ShouldBeError)
	})
}

func testInsertNodeInPriGroup4PTasks(
	task *vapi.TaskInfo,
	priNodeGroups []map[string]*npuPriNodeInf,
	addPriNodeGroupFn initPriNodeGroupFn) {

	Convey("insertNodeInPriGroup() 4P should add node into 1st group", func() {
		sNodeInf := selectNodeInf{nodeName: nodeName, allNPUNum: 6, leftNPUNum: 2, rightNPUNum: 4}
		err := insertNodeInPriGroup(task, sNodeInf, priNodeGroups, addPriNodeGroupFn)
		So(err, ShouldBeNil)
		So(priNodeGroups[0][nodeName], ShouldNotBeNil)
	})

	Convey("insertNodeInPriGroup() 4P should return error when neither hccl-ring has enough NPUs", func() {
		sNodeInf := selectNodeInf{nodeName: nodeName, allNPUNum: 4, leftNPUNum: 2, rightNPUNum: 2}
		err := insertNodeInPriGroup(task, sNodeInf, priNodeGroups, addPriNodeGroupFn)
		So(err, ShouldBeError)
	})
}

func testInsertNodeInPriGroup8PTasks(
	task *vapi.TaskInfo,
	priNodeGroups []map[string]*npuPriNodeInf,
	addPriNodeGroupFn initPriNodeGroupFn) {

	Convey("insertNodeInPriGroup() 8P should add node into 1st group", func() {
		sNodeInf := selectNodeInf{nodeName: nodeName, allNPUNum: 8, leftNPUNum: 4, rightNPUNum: 4}
		err := insertNodeInPriGroup(task, sNodeInf, priNodeGroups, addPriNodeGroupFn)
		So(err, ShouldBeNil)
		So(priNodeGroups[0][nodeName], ShouldNotBeNil)
	})

	Convey("insertNodeInPriGroup() 8P should return error when node doesn't have enough NPUs", func() {
		sNodeInf := selectNodeInf{nodeName: nodeName, allNPUNum: 5, leftNPUNum: 1, rightNPUNum: 4}
		err := insertNodeInPriGroup(task, sNodeInf, priNodeGroups, addPriNodeGroupFn)
		So(err, ShouldBeError)
	})
}

// TestMNPUInsertNodeInPriGroup
func TestMNPUInsertNodeInPriGroup(t *testing.T) {
	Convey("Test module910x8 insertNodeInPriGroup", t, func() {
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
		task1p := vapi.NewTaskInfo(buildNPUPod(MPodInfo{namespace: "default", groupName: "group-M-model-1p",
			podName: "npu-test-M-model-1p", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: npu800And9000CardName, reqNpuNum: "1"}))
		task2p := vapi.NewTaskInfo(buildNPUPod(MPodInfo{namespace: "default", groupName: "group-M-model-2p",
			podName: "npu-test-M-model-2p", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: npu800And9000CardName, reqNpuNum: "2"}))
		task4p := vapi.NewTaskInfo(buildNPUPod(MPodInfo{namespace: "default", groupName: "group-M-model-4p",
			podName: "npu-test-M-model-4p", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: npu800And9000CardName, reqNpuNum: "4"}))
		task8p := vapi.NewTaskInfo(buildNPUPod(MPodInfo{namespace: "default", groupName: "group-M-model-8p",
			podName: "npu-test-M-model-8p", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: npu800And9000CardName, reqNpuNum: "8"}))

		testInsertNodeInPriGroup1PTasks(task1p, priNodeGroups, addPriNodeGroupFn)
		testInsertNodeInPriGroup2PTasks(task2p, priNodeGroups, addPriNodeGroupFn)
		testInsertNodeInPriGroup4PTasks(task4p, priNodeGroups, addPriNodeGroupFn)
		testInsertNodeInPriGroup8PTasks(task8p, priNodeGroups, addPriNodeGroupFn)

		task7p := vapi.NewTaskInfo(buildNPUPod(MPodInfo{namespace: "default", groupName: "group-M-model-7p",
			podName: "npu-test-M-model-7p", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: npu800And9000CardName, reqNpuNum: "7"}))
		Convey("insertNodeInPriGroup() 7P should return error", func() {
			sNodeInf := selectNodeInf{nodeName: nodeName, allNPUNum: 8, leftNPUNum: 4, rightNPUNum: 4}
			err := insertNodeInPriGroup(task7p, sNodeInf, priNodeGroups, addPriNodeGroupFn)
			So(err, ShouldBeError)
		})
	})
}

// TestMNPUJudgeNodeAndTaskNPU
func TestMNPUJudgeNodeAndTaskNPU(t *testing.T) {
	Convey("Test module910x8 judgeNodeAndTaskNPU", t, func() {
		Convey("judgeNodeAndTaskNPU() should return nil when taskNPU is 0", func() {
			err := judgeNodeAndTaskNPU(0, []int{0, 1, constIntNum2, constIntNum3})
			So(err, ShouldBeNil)
		})
		Convey("judgeNodeAndTaskNPU() should return err when node doesn't satisfy req of 1", func() {
			err := judgeNodeAndTaskNPU(1, []int{})
			So(err, ShouldBeError)
		})
		Convey("judgeNodeAndTaskNPU() should return nil when node satisfies req of 2", func() {
			err := judgeNodeAndTaskNPU(2, []int{0, 1})
			So(err, ShouldBeNil)
		})
		Convey("judgeNodeAndTaskNPU() should return nil when node doesn't satisfy req of 4", func() {
			err := judgeNodeAndTaskNPU(4, []int{constIntNum4, constIntNum5, constIntNum6})
			So(err, ShouldBeError)
		})
		Convey("judgeNodeAndTaskNPU() should return nil when node satisfies req of 8", func() {
			err := judgeNodeAndTaskNPU(8, []int{0, constIntNum1, constIntNum2, constIntNum3,
				constIntNum4, constIntNum5, constIntNum6, constIntNum7})
			So(err, ShouldBeNil)
		})
		Convey("judgeNodeAndTaskNPU() should return er when req num is invalid", func() {
			err := judgeNodeAndTaskNPU(7, []int{0, constIntNum1, constIntNum2, constIntNum3,
				constIntNum4, constIntNum5, constIntNum6, constIntNum7})
			So(err, ShouldBeError)
		})
	})
}
