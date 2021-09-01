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

Package chip310x4 is using for HuaWei A300T Ascend pin affinity schedule.

*/
package chip310x4

import (
	. "github.com/smartystreets/goconvey/convey"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cs "strconv"
	"testing"
	vapi "volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/conf"
	"volcano.sh/volcano/pkg/scheduler/util"
)

type CNodeInfo struct {
	nodeName       string
	nodeArch       string
	cpu, mem       string
	npuAllocateNum string
	npuTop         string
}

type CPodInfo struct {
	namespace  string
	groupName  string
	podName    string
	nodeName   string
	reqCPUNum  string
	reqMem     string
	reqNPUType string
	reqNpuNum  string
}

const (
	nodeName     = "centos"
	constInt2000 = 2000
	constInt5000 = 5000
	NPUID0       = 0
	NPUID1       = 1
	NPUID2       = 2
	NPUID3       = 3
	NPUID4       = 4
	NPUID5       = 5
	NPUID7       = 7
	NPUID8       = 8
	NPUID9       = 9
	NPUID10      = 10
	NPUID11      = 11
	NPUID12      = 12
	NPUID13      = 13
	NPUID15      = 15
	NPUID16      = 16
	NPUID18      = 18
	NPUID20      = 20
	NPUID21      = 21
	NPUID22      = 22
	NPUID25      = 25
	NPUID29      = 29
	NPUID32      = 32
	NPUID33      = 33
	NPUID34      = 34
	NPUID35      = 35
	NPUID36      = 36
	NPUID37      = 37
	NPUID38      = 38
	NPUID39      = 39
)

// TestCNPUName
func TestCNPUName(t *testing.T) {
	Convey("Test chip310x4 Name", t, func() {
		npu := &chip310x4{}

		Convey("Name() should return PluginName defined in const", func() {
			n := npu.Name()
			So(n, ShouldEqual, PluginName)
		})
	})
}

// TestCNPUIsMyTask
func TestCNPUIsMyTask(t *testing.T) {
	Convey("Test chip310x4 IsMyTask", t, func() {
		npu := &chip310x4{}

		Convey("IsMyTask() should return error when task doesn't request NPU", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-100",
				podName: "npu-test-100", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
				reqNPUType: a310NPUChipName, reqNpuNum: "0"})
			task := vapi.NewTaskInfo(pod)
			result := npu.IsMyTask(task)
			So(result, ShouldBeError)
		})
		Convey("IsMyTask() should return error when task is of card mode", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-101",
				podName: "npu-test-101", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
				reqNPUType: a310NPUChipName, reqNpuNum: "1"})
			setPodSelector(pod, acceleratorType, cardAcceleratorType)
			task := vapi.NewTaskInfo(pod)
			result := npu.IsMyTask(task)
			So(result, ShouldBeError)
		})
		Convey("IsMyTask() should return nil when task is of chip type", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-102",
				podName: "npu-test-102", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
				reqNPUType: a310NPUChipName, reqNpuNum: "1"})
			task := vapi.NewTaskInfo(pod)
			result := npu.IsMyTask(task)
			So(result, ShouldBeNil)
		})
	})
}

// TestCNPUIsMyNode
func TestCNPUIsMyNode(t *testing.T) {
	Convey("Test chip310x4 IsMyNode", t, func() {
		npu := &chip310x4{}

		Convey("IsMyNode() should return error when node has no npu annotation", func() {
			node := buildNPUNode(CNodeInfo{nodeName, huaweiArchArm, "192", "755Gi",
				"0", ""})
			result := npu.IsMyNode(node)
			So(result, ShouldBeError)
		})
		Convey("IsMyNode() should return nil when node has npu annotation", func() {
			node := buildNPUNode(CNodeInfo{nodeName, huaweiArchArm, "192", "755Gi",
				"1", "Ascend310-11"})
			result := npu.IsMyNode(node)
			So(result, ShouldBeNil)
		})
	})
}

// TestCNPUIsMyJob
func TestCNPUIsMyJob(t *testing.T) {
	Convey("Test chip310x4 IsMyJob", t, func() {
		npu := &chip310x4{}
		tasks := []*vapi.TaskInfo{}
		uid := vapi.JobID("xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxx10")

		Convey("IsMyJob() should return error when job request no npu", func() {
			tasks = append(tasks, vapi.NewTaskInfo(buildNPUPod(
				CPodInfo{namespace: "default", groupName: "npu-group-103",
					podName: "npu-test-103", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
					reqNPUType: a310NPUChipName, reqNpuNum: "0"})))
			job := vapi.NewJobInfo(uid, tasks...)
			result := npu.IsMyJob(job)
			So(result, ShouldBeError)
		})
		Convey("IsMyJob() should return nil when job is of card mode", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-105",
				podName: "npu-test-105", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
				reqNPUType: a310NPUChipName, reqNpuNum: "1"})
			tasks = append(tasks, vapi.NewTaskInfo(pod))
			job := vapi.NewJobInfo(uid, tasks...)
			result := npu.IsMyJob(job)
			So(result, ShouldBeNil)
		})
	})
}

// TestCNPUValidNPUJobFnInvalidSelector
func TestCNPUValidNPUJobFnInvalidSelector(t *testing.T) {
	Convey("Test chip310x4 ValidNPUJobFnInvalidSelector", t, func() {
		const (
			invalidSelectorKey   = "invalid-key"
			invalidSelectorValue = "no-selector"
			validNum2            = 2000
		)
		npu := &chip310x4{}
		tasks := []*vapi.TaskInfo{}
		uid := vapi.JobID("xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx5")

		Convey("ValidNPUJobFn() should return error for job without certain selector key", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-12",
				podName: "npu-test-35", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a310NPUChipName, reqNpuNum: "1"})
			setPodSelector(pod, acceleratorType, "")
			setPodSelector(pod, invalidSelectorKey, invalidSelectorValue)
			tasks = append(tasks, vapi.NewTaskInfo(pod))
			job := vapi.NewJobInfo(uid, tasks...)
			setJobResourceReq(job, a310NPUChipName, float64(validNum2))
			result := npu.ValidNPUJobFn(job)
			So(result, ShouldNotBeNil)
		})
		Convey("ValidNPUJobFn() should return error for job with invalid selector value", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-13",
				podName: "npu-test-36", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a310NPUChipName, reqNpuNum: "1"})
			setPodSelector(pod, acceleratorType, invalidSelectorValue)
			tasks = append(tasks, vapi.NewTaskInfo(pod))
			job := vapi.NewJobInfo(uid, tasks...)
			result := npu.ValidNPUJobFn(job)
			So(result, ShouldNotBeNil)
		})
	})
}

// TestCNPUValidNPUJobFnInvalidNum
func TestCNPUValidNPUJobFnInvalidNum(t *testing.T) {
	Convey("Test chip310x4 ValidNPUJobFnInvalidNum", t, func() {
		const (
			invalidNum0  = "0"
			invalidNum65 = "65"
			validNum1    = "1"
			validNum64   = "64"
		)
		npu := &chip310x4{}
		tasks := []*vapi.TaskInfo{}
		uid := vapi.JobID("xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx8")

		Convey("ValidNPUJobFn() should return error for job with invalid request number 0", func() {
			tasks = append(tasks, vapi.NewTaskInfo(buildNPUPod(
				CPodInfo{namespace: "default", groupName: "npu-group-17",
					podName: "npu-test-40", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
					reqNPUType: a310NPUChipName, reqNpuNum: invalidNum0})))
			job := vapi.NewJobInfo(uid, tasks...)
			result := npu.ValidNPUJobFn(job)
			So(result, ShouldNotBeNil)
		})
		Convey("ValidNPUJobFn() should return error for job with invalid request number 65", func() {
			tasks = append(tasks, vapi.NewTaskInfo(buildNPUPod(
				CPodInfo{namespace: "default", groupName: "npu-group-18",
					podName: "npu-test-41", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
					reqNPUType: a310NPUChipName, reqNpuNum: invalidNum65})))
			job := vapi.NewJobInfo(uid, tasks...)
			result := npu.ValidNPUJobFn(job)
			So(result, ShouldNotBeNil)
		})
		Convey("ValidNPUJobFn() should return error for job with valid request number 1", func() {
			tasks = append(tasks, vapi.NewTaskInfo(buildNPUPod(
				CPodInfo{namespace: "default", groupName: "npu-group-19",
					podName: "npu-test-42", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
					reqNPUType: a310NPUChipName, reqNpuNum: validNum1})))
			job := vapi.NewJobInfo(uid, tasks...)
			result := npu.ValidNPUJobFn(job)
			So(result, ShouldBeNil)
		})
		Convey("ValidNPUJobFn() should return error for job with valid request number 64", func() {
			tasks = append(tasks, vapi.NewTaskInfo(buildNPUPod(
				CPodInfo{namespace: "default", groupName: "npu-group-19",
					podName: "npu-test-42", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
					reqNPUType: a310NPUChipName, reqNpuNum: validNum64})))
			job := vapi.NewJobInfo(uid, tasks...)
			result := npu.ValidNPUJobFn(job)
			So(result, ShouldBeNil)
		})
	})
}

// TestCNPUValidNPUJobFnInvalidModel
func TestCNPUValidNPUJobFnInvalidModel(t *testing.T) {
	Convey("Test chip310x4 ValidNPUJobFnInvalidModel", t, func() {
		const (
			num4 = "4"
			num2 = "2"
		)
		npu := &chip310x4{}
		tasks := []*vapi.TaskInfo{}
		uid := vapi.JobID("xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx9")

		Convey("ValidNPUJobFn() should return error for job with distributed", func() {
			tasks = append(tasks, vapi.NewTaskInfo(buildNPUPod(
				CPodInfo{namespace: "default", groupName: "npu-group-23",
					podName: "npu-test-46", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
					reqNPUType: a310NPUChipName, reqNpuNum: num2})))
			tasks = append(tasks, vapi.NewTaskInfo(buildNPUPod(
				CPodInfo{namespace: "default", groupName: "npu-group-123",
					podName: "npu-test-146", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
					reqNPUType: a310NPUChipName, reqNpuNum: num2})))
			job := vapi.NewJobInfo(uid, tasks...)
			result := npu.ValidNPUJobFn(job)
			So(result, ShouldBeNil)
		})
		Convey("ValidNPUJobFn() should return nil for job with valid model", func() {
			tasks = append(tasks, vapi.NewTaskInfo(buildNPUPod(
				CPodInfo{namespace: "default", groupName: "npu-group-24",
					podName: "npu-test-47", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
					reqNPUType: a310NPUChipName, reqNpuNum: num4})))
			job := vapi.NewJobInfo(uid, tasks...)
			result := npu.ValidNPUJobFn(job)
			So(result, ShouldBeNil)
		})
	})
}

// TestCnpuPreCheckNodeFnSuccess
func TestCnpuPreCheckNodeFnSuccess(t *testing.T) {
	Convey("Test chip310x4 PreCheckNodeFnSuccess", t, func() {
		npu := &chip310x4{}
		confs := []conf.Configuration{}
		node := buildNPUNode(CNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
			"1", "Ascend310-6"})

		Convey("PreCheckNodeFn() should return nil", func() {
			// build a task with no selector
			task := vapi.NewTaskInfo(buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-14",
				podName: "npu-test-14", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a310NPUChipName, reqNpuNum: "1"}))
			result := npu.PreCheckNodeFn(task, node, confs)
			So(result, ShouldBeNil)
		})
	})
}

// TestCnpuCheckNPUResourceStableFn
func TestCnpuCheckNPUResourceStableFn(t *testing.T) {
	Convey("Test chip310x4 CheckNPUResourceStableFn", t, func() {
		vnpu := &chip310x4{}

		Convey("CheckNPUResourceStableFn() should return error when there's missing resource type in idle", func() {
			node := buildNPUNode(CNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
				"0", ""})
			node.Node.Annotations[a310NPUChipName] = "Ascend310-2"
			result := vnpu.CheckNPUResourceStableFn(node)
			So(result, ShouldBeError)
		})
		Convey("CheckNPUResourceStableFn() should return error when node resources are unstable", func() {
			node := buildNPUNode(CNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
				"1", ""})
			node.Node.Annotations[a310NPUChipName] = "Ascend310-2,Ascend310-3"
			result := vnpu.CheckNPUResourceStableFn(node)
			So(result, ShouldBeError)
		})
		Convey("CheckNPUResourceStableFn() should return nil when node resources are stable", func() {
			node := buildNPUNode(CNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
				"2", "Ascend310-2,Ascend310-3"})
			result := vnpu.CheckNPUResourceStableFn(node)
			So(result, ShouldBeNil)
		})
	})
}

// TestCnpuCheckNodeNPUByTaskFn
func TestCnpuCheckNodeNPUByTaskFn(t *testing.T) {
	Convey("Test job CheckNodeNPUByTaskFn", t, func() {
		npu := &chip310x4{}
		pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-15",
			podName: "npu-test-15", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: a310NPUChipName, reqNpuNum: "5"})
		task := vapi.NewTaskInfo(pod)

		Convey("CheckNodeNPUByTaskFn() should return error when node doesn't meet task requests", func() {
			node := buildNPUNode(CNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
				"3", "Ascend310-3,Ascend310-7,Ascend310-8"})
			result := npu.CheckNodeNPUByTaskFn(task, node, true)
			So(result, ShouldBeError)
		})
		Convey("CheckNodeNPUByTaskFn() should return error when node meets task requests", func() {
			node := buildNPUNode(CNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
				"2", "Ascend310-6,Ascend310-7,Ascend310-8,Ascend310-9,Ascend310-10"})
			result := npu.CheckNodeNPUByTaskFn(task, node, true)
			So(result, ShouldBeNil)
		})
	})
}

// TestCnpuCheckNodeNPUByTaskFnFail
func TestCnpuCheckNodeNPUByTaskFnError(t *testing.T) {
	Convey("Test job CheckNodeNPUByTaskFnError", t, func() {
		npu := &chip310x4{}

		Convey("CheckNodeNPUByTaskFn() should return error when the requirement of the task is 0", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-115",
				podName: "npu-test-115", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a310NPUChipName, reqNpuNum: "0"})
			task := vapi.NewTaskInfo(pod)
			node := buildNPUNode(CNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
				"2", "Ascend310-1,Ascend310-4"})
			result := npu.CheckNodeNPUByTaskFn(task, node, true)
			So(result, ShouldBeError)
		})

		Convey("CheckNodeNPUByTaskFn() should return error when the number of NUP on a node is 0", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-116",
				podName: "npu-test-116", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a310NPUChipName, reqNpuNum: "2"})
			task := vapi.NewTaskInfo(pod)
			node := buildNPUNode(CNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
				"", ""})
			result := npu.CheckNodeNPUByTaskFn(task, node, true)
			So(result, ShouldBeError)
		})
	})
}

// TestCnpuGetNPUAffinityBestNodesFnReq1
func TestCnpuGetNPUAffinityBestNodesFnReq1(t *testing.T) {
	Convey("Test chip310x4 GetNPUAffinityBestNodesFn When Request is 1", t, func() {
		const (
			nodeName1 = "centos1"
			nodeName2 = "centos2"
		)
		npu := &chip310x4{}
		Convey("GetNPUAffinityBestNodesFn() should return correct result when request NPU num is 1", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-17",
				podName: "npu-test-17", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a310NPUChipName, reqNpuNum: "1"})
			task := vapi.NewTaskInfo(pod)
			nodes := []*vapi.NodeInfo{}
			node1 := buildNPUNode(CNodeInfo{nodeName1, huaweiArchX86, "192", "755Gi",
				"3", "Ascend310-0,Ascend310-4,Ascend310-6,Ascend310-9,Ascend310-10," +
					"Ascend310-11,Ascend310-20,Ascend310-21,Ascend310-22,Ascend310-23"})
			nodes = append(nodes, node1)
			node2 := buildNPUNode(CNodeInfo{nodeName2, huaweiArchX86, "192", "755Gi",
				"4", "Ascend310-4,Ascend310-6,Ascend310-9,Ascend310-10," +
					"Ascend310-11,Ascend310-20,Ascend310-21,Ascend310-22,Ascend310-23"})
			nodes = append(nodes, node2)
			result, err := npu.GetNPUAffinityBestNodesFn(task, nodes, false)
			So(result, ShouldBeNil)
			So(err, ShouldBeNil)
		})
	})
}

// TestCnpuScoreBestNPUNodesFn
func TestCnpuScoreBestNPUNodesFn(t *testing.T) {
	Convey("Test chip310x4 ScoreBestNPUNodesFn", t, func() {
		const (
			nodeName1 = "euler1"
			nodeName2 = "euler2"
		)
		npu := &chip310x4{}
		bestNodes := map[string]int{
			nodeName1: 0,
			nodeName2: constIntNum2,
		}
		pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-18",
			podName: "npu-test-18", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: a310NPUChipName, reqNpuNum: "2"})
		task := vapi.NewTaskInfo(pod)

		nodes := []*vapi.NodeInfo{}
		node1 := buildNPUNode(CNodeInfo{nodeName1, huaweiArchX86, "192", "755Gi",
			"4", "Ascend310-0,Ascend310-2"})
		nodes = append(nodes, node1)
		node2 := buildNPUNode(CNodeInfo{nodeName2, huaweiArchArm, "192", "755Gi",
			"4", "Ascend310-0,Ascend310-2,Ascend310-3,Ascend310-1"})
		nodes = append(nodes, node2)

		Convey("ScoreBestNPUNodesFn() should return correct result", func() {
			scoreMap := make(map[string]float64)
			_, err := npu.ScoreBestNPUNodesFn(scoreMap, bestNodes, task, nodes)
			So(err, ShouldBeNil)
		})
	})
}

// TestCnpuUpdateNPUNodeUsedCardFn1
func TestCnpuUpdateNPUNodeUsedCardFn1(t *testing.T) {
	Convey("Test chip310x4 UpdateNPUNodeUsedCardFn", t, func() {
		npu := &chip310x4{}
		node := buildNPUNode(CNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
			"7", "Ascend310-0,Ascend310-1,Ascend310-2,Ascend310-3,Ascend310-61," +
				"Ascend310-62,Ascend310-63"})
		node.Others = map[string]interface{}{
			a310NPUChipName: "Ascend310-0,Ascend310-1,Ascend310-2,Ascend310-3,Ascend310-61,Ascend310-62,Ascend310-63",
		}

		Convey("UpdateNPUNodeUsedCardFn() should successfully update node.others", func() {
			top := []int{0, 3, 61, 62}
			expectedResult := map[string]interface{}{
				a310NPUChipName: "Ascend310-1,Ascend310-2,Ascend310-63",
			}
			err := npu.UpdateNPUNodeUsedCardFn(node, top)
			So(err, ShouldBeNil)
			So(node.Others, ShouldResemble, expectedResult)
		})
	})
}

// TestCnpuUpdateNPUNodeUsedCardFnError2
func TestCnpuUpdateNPUNodeUsedCardFn2(t *testing.T) {
	Convey("Test chip310x4 UpdateNPUNodeUsedCardFn", t, func() {
		npu := &chip310x4{}

		Convey("UpdateNPUNodeUsedCardFn() should should successfully update node.others", func() {
			node := buildNPUNode(CNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
				"4", "Ascend310-0,Ascend310-2,Ascend310-3,Ascend310-1"})
			node.Others = map[string]interface{}{
				a310NPUChipName: "Ascend310-0,Ascend310-2,Ascend310-3,Ascend310-1",
			}
			top := []int{0, 1, 2, 3}
			expectedResult := map[string]interface{}{
				a310NPUChipName: "",
			}
			err := npu.UpdateNPUNodeUsedCardFn(node, top)
			So(err, ShouldBeNil)
			So(node.Others, ShouldResemble, expectedResult)
		})
	})
}

// TestCnpuUpdateNPUNodeUsedCardFnError
func TestCnpuUpdateNPUNodeUsedCardFnError1(t *testing.T) {
	Convey("Test chip310x4 UpdateNPUNodeUsedCardFnError", t, func() {
		npu := &chip310x4{}

		Convey("UpdateNPUNodeUsedCardFn() should return error when top's type mismatch", func() {
			node := buildNPUNode(CNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
				"4", "Ascend310-0,Ascend310-2,Ascend310-3,Ascend310-1"})
			top := []string{"0", "4"}
			err := npu.UpdateNPUNodeUsedCardFn(node, top)
			So(err, ShouldBeError)
		})

		Convey("UpdateNPUNodeUsedCardFn() should return error when node's npuTop is empty", func() {
			node := buildNPUNode(CNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
				"8", ""})
			top := []int{0, 4}
			err := npu.UpdateNPUNodeUsedCardFn(node, top)
			So(err, ShouldBeError)
		})

	})
}

// TestCnpuGetReleaseNPUTopologyFn
func TestCnpuGetReleaseNPUTopologyFn(t *testing.T) {
	Convey("Test chip310x4 GetReleaseNPUTopologyFn", t, func() {
		npu := &chip310x4{}
		task := vapi.NewTaskInfo(buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-19",
			podName: "npu-test-19", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: a310NPUChipName, reqNpuNum: "2"}))
		Convey("GetReleaseNPUTopologyFn() should return correct card id slice", func() {
			task.Pod.Annotations[a310NPUChipName] = "Ascend310-0,Ascend310-4"
			expectedResult := []int{0, 4}
			result, err := npu.GetReleaseNPUTopologyFn(task)
			So(err, ShouldBeNil)
			So(result, ShouldResemble, expectedResult)
		})
	})
}

// TestCnpuGetReleaseNPUTopologyFnError
func TestCnpuGetReleaseNPUTopologyFnError(t *testing.T) {
	Convey("Test chip310x4 GetReleaseNPUTopologyFnError", t, func() {
		npu := &chip310x4{}
		task := vapi.NewTaskInfo(buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-122",
			podName: "npu-test-122", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: a310NPUChipName, reqNpuNum: "2"}))
		Convey("GetReleaseNPUTopologyFn() should return error when the annotations of pod is wrong", func() {
			task.Pod.Annotations[a310NPUChipName] = "Ascend3100,Ascend3104"
			_, err := npu.GetReleaseNPUTopologyFn(task)
			So(err, ShouldBeError)
		})
	})
}

// TestCnpuUpdateReleaseNPUNodeTopologyFn
func TestCnpuUpdateReleaseNPUNodeTopologyFn(t *testing.T) {
	Convey("Test chip310x4 UpdateReleaseNPUNodeTopologyFn", t, func() {
		npu := &chip310x4{}
		node := buildNPUNode(CNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
			"4", "Ascend310-0,Ascend310-2,Ascend310-61,Ascend310-1"})
		node.Others = map[string]interface{}{
			a310NPUChipName: "Ascend310-0,Ascend310-2",
		}
		Convey("UpdateNPUNodeUsedCardFn() should successfully update node.others", func() {
			top := []int{1, 61}
			err := npu.UpdateReleaseNPUNodeTopologyFn(node, top)
			So(err, ShouldBeNil)
			// 0, 1, 2, 61  disorderly
			So(node.Others, ShouldNotBeNil)
		})

	})
}

// TestCnpuUpdateReleaseNPUNodeTopologyFnError
func TestCnpuUpdateReleaseNPUNodeTopologyFnError(t *testing.T) {
	Convey("", t, func() {
		npu := &chip310x4{}

		Convey("UpdateNPUNodeUsedCardFn() should return error when top's type is mismatch", func() {
			node := buildNPUNode(CNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
				"4", "Ascend310-0,Ascend310-2,Ascend310-3,Ascend310-1"})
			node.Others = map[string]interface{}{
				a310NPUChipName: "Ascend310-0,Ascend310-4",
			}
			top := []string{"0", "3"}
			err := npu.UpdateReleaseNPUNodeTopologyFn(node, top)
			So(err, ShouldBeError)
		})

		Convey("UpdateNPUNodeUsedCardFn() should return error when node's others is wrong", func() {
			node := buildNPUNode(CNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
				"4", "Ascend310-0,Ascend310-2,Ascend310-3,Ascend310-1"})
			node.Others = map[string]interface{}{
				a310NPUChipName: "Ascend310-0Ascend310-2",
			}
			top := []int{0, 3}
			err := npu.UpdateReleaseNPUNodeTopologyFn(node, top)
			So(err, ShouldBeError)
		})
	})
}

// TestCnpuGetAllocatedNPUFromTopologyFn
func TestCnpuGetAllocatedNPUFromTopologyFn(t *testing.T) {
	Convey("Test chip310x4 GetAllocatedNPUFromTopologyFn", t, func() {
		npu := &chip310x4{}
		Convey("GetAllocatedNPUFromTopologyFn() should return correct result when reqNpuNum is 2", func() {
			task := vapi.NewTaskInfo(buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-20",
				podName: "npu-test-20", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a310NPUChipName, reqNpuNum: "2"}))
			node := buildNPUNode(CNodeInfo{nodeName, huaweiArchArm, "192", "755Gi",
				"2", "Ascend310-0,Ascend310-3"})
			expectedResult := []int{0, 3}
			result, err := npu.GetAllocatedNPUFromTopologyFn(task, node, false)
			So(err, ShouldBeNil)
			So(result, ShouldResemble, expectedResult)
		})
		Convey("GetAllocatedNPUFromTopologyFn() should return correct result when reqNpuNum is 1", func() {
			task := vapi.NewTaskInfo(buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-21",
				podName: "npu-test-21", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a310NPUChipName, reqNpuNum: "1"}))
			node := buildNPUNode(CNodeInfo{nodeName, huaweiArchArm, "192", "755Gi",
				"2", "Ascend310-0,Ascend310-3"})
			result, err := npu.GetAllocatedNPUFromTopologyFn(task, node, false)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
		})
	})
}

// TestCnpuGetAllocatedNPUFromTopologyFnError
func TestCnpuGetAllocatedNPUFromTopologyFnError(t *testing.T) {
	Convey("Test chip310x4 GetAllocatedNPUFromTopologyFnError", t, func() {
		npu := &chip310x4{}

		pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-123",
			podName: "npu-test-123", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: a310NPUChipName, reqNpuNum: "2"})
		task := vapi.NewTaskInfo(pod)
		node := buildNPUNode(CNodeInfo{nodeName, huaweiArchArm, "192", "755Gi",
			"2", "Ascend310-0,Ascend310-4"})

		Convey("GetAllocatedNPUFromTopologyFn() should return error when reqNpuNum of pod is 0", func() {
			task.Resreq.ScalarResources[a310NPUChipName] = 0
			_, err := npu.GetAllocatedNPUFromTopologyFn(task, node, false)
			So(err, ShouldBeError)
		})

		Convey("GetAllocatedNPUFromTopologyFn() should return error when reqNpuNum Is greater than 4", func() {
			task.Resreq.ScalarResources[a310NPUChipName] = constInt5000
			_, err := npu.GetAllocatedNPUFromTopologyFn(task, node, false)
			So(err, ShouldBeError)
		})

		Convey("GetAllocatedNPUFromTopologyFn() should return nil when npuTop of node is wrong", func() {
			task.Resreq.ScalarResources[a310NPUChipName] = constInt2000
			node.Node.Annotations[a310NPUChipName] = "Ascend310-0Ascend310-4"
			result, err := npu.GetAllocatedNPUFromTopologyFn(task, node, false)
			So(err, ShouldBeNil)
			So(result, ShouldBeNil)
		})

		Convey("GetAllocatedNPUFromTopologyFn() should return correct result when reqNpuNum is 0", func() {
			task.Resreq.ScalarResources[a310NPUChipName] = 0
			node.Node.Annotations[a310NPUChipName] = "Ascend310-0,Ascend310-3"
			result, err := npu.GetAllocatedNPUFromTopologyFn(task, node, false)
			So(err, ShouldBeError)
			So(result, ShouldBeNil)
		})

		Convey("GetAllocatedNPUFromTopologyFn() should return error when none node meet request", func() {
			task.Resreq.ScalarResources[a310NPUChipName] = constInt2000
			node.Node.Annotations[a310NPUChipName] = "Ascend310-0"
			_, err := npu.GetAllocatedNPUFromTopologyFn(task, node, false)
			So(err, ShouldBeError)
		})
	})
}

// nodeTop: [0, 2, 4, 5, 7, 8, 9, 10, 11, 12, 13, 15, 16, 18, 20, 21, 22, 25, 29, 32, 33, 34, 35, 36, 37, 38, 39]
// cardNumGroups: [2, 3, 4, 3, 2, 3, 1, 1, 4, 4] Each group of four NPU
func TestCnpuGetFitCardFromNodeByPriorityFn(t *testing.T) {
	Convey("Test chip310x4 GetFitCardFromNodeByPriorityFn", t, func() {
		nodeTop := []int{NPUID0, NPUID2, NPUID4, NPUID5, NPUID7, NPUID8, NPUID9, NPUID10, NPUID11, NPUID12,
			NPUID13, NPUID15, NPUID16, NPUID18, NPUID20, NPUID21, NPUID22, NPUID25, NPUID29, NPUID32,
			NPUID33, NPUID34, NPUID35, NPUID36, NPUID37, NPUID38, NPUID39}

		priorityArray := [cardNPUNumber]int{1, 2, 3, 4}
		Convey("getFitCardFromNodeByPriorityFn() should return correct result "+
			"when taskNPUNumber is 6", func() {
			taskNPUNumber := 6
			result, err := getFitCardFromNodeByPriority(taskNPUNumber, nodeTop, priorityArray)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
		})

		Convey("getFitCardFromNodeByPriorityFn() should return correct result "+
			"when taskNPUNumber is 7", func() {
			taskNPUNumber := 7
			result, err := getFitCardFromNodeByPriority(taskNPUNumber, nodeTop, priorityArray)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
		})

		Convey("getFitCardFromNodeByPriorityFn() should return correct result "+
			"when taskNPUNumber is 20", func() {
			taskNPUNumber := 21
			result, err := getFitCardFromNodeByPriority(taskNPUNumber, nodeTop, priorityArray)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
		})

		Convey("getFitCardFromNodeByPriorityFn() should return correct result "+
			"when taskNPUNumber is 27", func() {
			taskNPUNumber := 27
			result, err := getFitCardFromNodeByPriority(taskNPUNumber, nodeTop, priorityArray)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
		})
	})
}

func TestCnpuGetFitCardFromNodeByPriorityFnNotMatch(t *testing.T) {
	Convey("Test chip310x4 GetFitCardFromNodeByPriorityFn When nodeTop number is 64", t, func() {
		nodeTop := make([]int, 64)
		for i := 0; i < 64; i++ {
			nodeTop[i] = i
		}

		Convey("getFitCardFromNodeByPriorityFn() should return correct result "+
			"when nodeTop number is 64", func() {
			priorityArray := [cardNPUNumber]int{NPUID1, NPUID2, NPUID3, NPUID4}
			taskNPUNumber := 64
			result, err := getFitCardFromNodeByPriority(taskNPUNumber, nodeTop, priorityArray)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
		})
	})
}

// TestCnpuSetNPUTopologyToPodFn
func TestCnpuSetNPUTopologyToPodFn(t *testing.T) {
	Convey("Test chip310x4 SetNPUTopologyToPodFn", t, func() {
		npu := &chip310x4{}
		task := vapi.NewTaskInfo(buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-22",
			podName: "npu-test-22", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: a310NPUChipName, reqNpuNum: "4"}))
		Convey("SetNPUTopologyToPodFn() should return error when top is of wrong type", func() {
			top := []string{}
			err := npu.SetNPUTopologyToPodFn(task, top)
			So(err, ShouldBeError)
			So(task.Pod.Annotations[a310NPUChipName], ShouldEqual, "")
		})
		Convey("SetNPUTopologyToPodFn() should write correct info in pod annotation", func() {
			top := []int{0, 4}
			expectedResult := "Ascend310-0,Ascend310-4"
			err := npu.SetNPUTopologyToPodFn(task, top)
			So(err, ShouldBeNil)
			So(task.Pod.Annotations[a310NPUChipName], ShouldEqual, expectedResult)
		})
	})
}

func setPodSelector(CPod *v1.Pod, selectorKey string, selectorValue string) {
	if selectorValue == "" {
		delete(CPod.Labels, selectorKey)
		return
	}
	CPod.Labels[selectorKey] = selectorValue
}

func setNodeLabel(CNode *v1.Node, labelKey string, labelValue string) {
	if labelValue == "" {
		delete(CNode.Labels, labelKey)
		return
	}
	CNode.Labels[labelKey] = labelValue
}

func setJobResourceReq(CJob *vapi.JobInfo, resource string, num float64) {
	CJob.TotalRequest.ScalarResources[v1.ResourceName(resource)] = num
}

func buildNPUPod(podInfo CPodInfo) *v1.Pod {

	pod := util.BuildPod(podInfo.namespace, podInfo.podName, podInfo.nodeName, v1.PodPending,
		buildNPUResourceList(podInfo.reqCPUNum, podInfo.reqMem,
			v1.ResourceName(podInfo.reqNPUType), podInfo.reqNpuNum),
		podInfo.groupName, make(map[string]string, constIntNum2),
		make(map[string]string, constIntNum2))

	setPodSelector(pod, archSelector, huaweiArchX86)
	setPodSelector(pod, acceleratorType, chipAcceleratorType)

	return pod
}

func buildNPUResourceList(CCpu string, CMemory string, npuResourceType v1.ResourceName, npu string) v1.ResourceList {
	npuNum, err := cs.Atoi(npu)
	if err != nil {
		return nil
	}

	if npuNum == 0 {
		return v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse(CCpu),
			v1.ResourceMemory: resource.MustParse(CMemory),
		}
	}

	return v1.ResourceList{
		v1.ResourceCPU:    resource.MustParse(CCpu),
		v1.ResourceMemory: resource.MustParse(CMemory),
		npuResourceType:   resource.MustParse(npu),
	}
}

func buildNPUNode(CNode CNodeInfo) *vapi.NodeInfo {
	nodeCapacity := buildNPUResourceList(CNode.cpu, CNode.mem, a310NPUChipName, cs.Itoa(constIntNum2))
	nodeAlloc := buildNPUResourceList(CNode.cpu, CNode.mem, a310NPUChipName, CNode.npuAllocateNum)
	labels := make(map[string]string, constIntNum2)
	ann := make(map[string]string, constIntNum2)

	v1node := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:        CNode.nodeName,
			Labels:      labels,
			Annotations: ann,
		},
		Status: v1.NodeStatus{
			Capacity:    nodeCapacity,
			Allocatable: nodeAlloc,
		},
	}

	if CNode.npuAllocateNum != "0" {
		v1node.Annotations[a310NPUChipName] = CNode.npuTop
	}

	setNodeLabel(v1node, archSelector, CNode.nodeArch)

	node := vapi.NewNodeInfo(v1node)
	return node
}
