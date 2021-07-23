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

Package module910x8 is using for HuaWei A800/9000 Ascend910 pin affinity schedule.

*/
package module910x8

import (
	. "github.com/smartystreets/goconvey/convey"
	v1 "k8s.io/api/core/v1"
	"testing"
	vapi "volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/util"
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
			err := judgeNodeAndTaskNPU(0, []int{0, 1, 2, 3})
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
			err := judgeNodeAndTaskNPU(4, []int{4, 5, 6})
			So(err, ShouldBeError)
		})
		Convey("judgeNodeAndTaskNPU() should return nil when node satisfies req of 8", func() {
			err := judgeNodeAndTaskNPU(8, []int{0, 1, 2, 3, 4, 5, 6, 7})
			So(err, ShouldBeNil)
		})
		Convey("judgeNodeAndTaskNPU() should return er when req num is invalid", func() {
			err := judgeNodeAndTaskNPU(7, []int{0, 1, 2, 3, 4, 5, 6, 7})
			So(err, ShouldBeError)
		})
	})
}

// TestMNPUCheckFaultJobNode
func TestMNPUCheckFaultJobNode(t *testing.T) {
	Convey("Test module910x8 checkFaultJobNode", t, func() {
		rst := &util.ReSchedulerTasks{NodeNames: map[string]string{"fake_task": nodeName}}
		task := vapi.NewTaskInfo(buildNPUPod(MPodInfo{namespace: "default", groupName: "group-M-model-0",
			podName: "npu-test-M-model-0", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: npu800And9000CardName, reqNpuNum: "4"}))
		anotherTask := vapi.NewTaskInfo(buildNPUPod(MPodInfo{namespace: "default", groupName: "group-M-model-1",
			podName: "npu-test-M-model-1", nodeName: nodeName, reqCPUNum: "40", reqMem: "10Gi",
			reqNPUType: npu800And9000CardName, reqNpuNum: "1"}))
		node := buildNPUNode(MNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
			"4", "Ascend910-4,Ascend910-6,Ascend910-7,Ascend910-5"})
		Convey("checkFaultJobNode() should return nil if task is in reschedule map", func() {
			util.ReSchedulerJobs[task.Job] = *rst
			err := checkFaultJobNode(task, node)
			So(err, ShouldBeNil)
		})
		Convey("checkFaultJobNode() should return error if node is in reschedule map", func() {
			delete(util.ReSchedulerJobs, task.Job)
			util.ReSchedulerJobs[anotherTask.Job] = *rst
			err := checkFaultJobNode(task, node)
			So(err, ShouldBeError)
		})
		Convey("checkFaultJobNode() should return nil if node isn't in reschedule map", func() {
			newNode := buildNPUNode(MNodeInfo{"newNode", huaweiArchArm, "192", "755Gi",
				"6", "Ascend910-4,Ascend910-6,Ascend910-7,Ascend910-5,Ascend910-1,Ascend910-2"})
			err := checkFaultJobNode(task, newNode)
			So(err, ShouldBeNil)
		})
	})
}

// TestMNPUGetJobUsedNodes
func TestMNPUGetJobUsedNodes(t *testing.T) {
	Convey("Test module910x8 getJobUsedNodes", t, func() {
		uid := vapi.JobID("xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxx102")
		tasks := []*vapi.TaskInfo{}
		tasks = append(tasks, vapi.NewTaskInfo(buildNPUPod(
			MPodInfo{namespace: "default", groupName: "group-M-model-2",
				podName: "npu-test-M-model-2", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
				reqNPUType: npu800And9000CardName, reqNpuNum: "4"})))
		job := vapi.NewJobInfo(uid, tasks...)
		job.PodGroup = &vapi.PodGroup{}
		Convey("getJobUsedNodes() should return error when job is not in running state", func() {
			job.PodGroup.Status.Phase = "Unknown"
			_, err := getJobUsedNodes(job)
			So(err, ShouldBeError)
		})
		job.PodGroup.Status.Phase = "Running"
		Convey("getJobUsedNodes() should return a correct nodeName to Pod map", func() {
			expectedResult := map[string]*v1.Pod{
				nodeName: tasks[0].Pod,
			}
			result, err := getJobUsedNodes(job)
			So(err, ShouldBeNil)
			So(result, ShouldResemble, expectedResult)
		})
	})
}

// TestMNPUGetNodeIdleNPUIntCardsIncludeFaultTask
func TestMNPUGetNodeIdleNPUIntCardsIncludeFaultTask(t *testing.T) {
	Convey("Test module910x8 checkFaultJobNode", t, func() {
		const (
			constInt2 = 2
			constInt4 = 4
			constInt5 = 5
			constInt6 = 6
			constInt7 = 7
		)
		node := buildNPUNode(MNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
			"4", "Ascend910-4,Ascend910-6,Ascend910-7,Ascend910-5"})
		task := vapi.NewTaskInfo(buildNPUPod(MPodInfo{namespace: "default", groupName: "group-M-model-0",
			podName: "npu-test-M-model-0", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: npu800And9000CardName, reqNpuNum: "4"}))
		util.ReSchedulerJobs[task.Job] = util.ReSchedulerTasks{
			TaskUseNPUs: map[string]string{
				task.Name: "Ascend910-1,Ascend910-2",
			},
		}
		Convey("", func() {
			setNodeAnnotation(node.Node, faultNPU, "Ascend910-1")
			result := getNodeIdleNPUIntCardsIncludeFaultTask(task, node)
			So(result, ShouldResemble, []int{constInt4, constInt6, constInt7, constInt5, constInt2})
		})
	})
}

// TestMNPUIsNodeMeetTaskReqNPUSource
func TestMNPUIsNodeMeetTaskReqNPUSource(t *testing.T) {
	Convey("Test module910x8 isNodeMeetTaskReqNPUSource", t, func() {
		Convey("isNodeMeetTaskReqNPUSource() should return true case 1", func() {
			node := buildNPUNode(MNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
				"4", "Ascend910-4,Ascend910-6,Ascend910-7,Ascend910-5"})
			task := vapi.NewTaskInfo(buildNPUPod(MPodInfo{namespace: "default", groupName: "group-M-model-0",
				podName: "npu-test-M-model-0", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: npu800And9000CardName, reqNpuNum: "4"}))
			util.ReSchedulerJobs[task.Job] = util.ReSchedulerTasks{
				TaskUseNPUs: map[string]string{
					task.Name: "Ascend910-1,Ascend910-2",
				},
			}
			setNodeAnnotation(node.Node, faultNPU, "Ascend910-1")
			result := isNodeMeetTaskReqNPUSource(task, node)
			So(result, ShouldEqual, true)
		})
		Convey("isNodeMeetTaskReqNPUSource() should return error case 1", func() {
			node := buildNPUNode(MNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
				"8", "Ascend910-4,Ascend910-6,Ascend910-7," +
					"Ascend910-1,Ascend910-2,Ascend910-3,Ascend910-0"})
			task := vapi.NewTaskInfo(buildNPUPod(MPodInfo{namespace: "default", groupName: "group-M-model-0",
				podName: "npu-test-M-model-0", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: npu800And9000CardName, reqNpuNum: "8"}))
			setNodeAnnotation(node.Node, faultNPU, "Ascend910-5")
			result := isNodeMeetTaskReqNPUSource(task, node)
			So(result, ShouldEqual, false)
		})
		Convey("isNodeMeetTaskReqNPUSource() should return error case 2", func() {
			node := buildNPUNode(MNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
				"2", "Ascend910-0,Ascend910-6"})
			task := vapi.NewTaskInfo(buildNPUPod(MPodInfo{namespace: "default", groupName: "group-M-model-0",
				podName: "npu-test-M-model-0", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: npu800And9000CardName, reqNpuNum: "2"}))
			util.ReSchedulerJobs[task.Job] = util.ReSchedulerTasks{
				TaskUseNPUs: map[string]string{
					task.Name: "Ascend910-3,Ascend910-2",
				},
			}
			setNodeAnnotation(node.Node, faultNPU, "Ascend910-3,Ascend910-0")
			result := isNodeMeetTaskReqNPUSource(task, node)
			So(result, ShouldEqual, false)
		})
	})
}
