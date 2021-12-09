/*
Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package card910x2 is using for HuaWei A300T Ascend pin affinity schedule.

*/
package card910x2

import (
	. "github.com/smartystreets/goconvey/convey"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
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
	constInt3000 = 3000
)

// TestCNPUName
func TestCNPUName(t *testing.T) {
	Convey("Test card910x2 Name", t, func() {
		npu := &card910x2{}

		Convey("Name() should return PluginName defined in const", func() {
			n := npu.Name()
			So(n, ShouldEqual, PluginName)
		})
	})
}

// TestCNPUValidNPUJobFnInvalidSelector
func TestCNPUValidNPUJobFnInvalidSelector(t *testing.T) {
	Convey("Test card910x2 ValidNPUJobFnInvalidSelector", t, func() {
		const (
			invalidSelectorKey   = "invalid-key"
			invalidSelectorValue = "no-selector"
			validNum2            = 2000
		)
		npu := &card910x2{}
		tasks := []*vapi.TaskInfo{}
		uid := vapi.JobID("xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx5")

		Convey("ValidNPUJobFn() should return error for job without certain selector key", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-1",
				podName: "npu-test-1", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a300TNPUCardName, reqNpuNum: "1"})
			setPodSelector(pod, acceleratorType, "")
			setPodSelector(pod, invalidSelectorKey, invalidSelectorValue)
			tasks = append(tasks, vapi.NewTaskInfo(pod))
			job := vapi.NewJobInfo(uid, tasks...)
			setJobResourceReq(job, a300TNPUCardName, float64(validNum2))
			result := npu.ValidNPUJobFn(job)
			So(result, ShouldNotBeNil)
		})
		Convey("ValidNPUJobFn() should return error for job with invalid selector value", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-2",
				podName: "npu-test-2", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a300TNPUCardName, reqNpuNum: "1"})
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
	Convey("Test card910x2 ValidNPUJobFnInvalidNum", t, func() {
		const (
			invalidNum0 = "0"
			invalidNum5 = "5"
			validNum1   = "1"
		)
		npu := &card910x2{}
		tasks := []*vapi.TaskInfo{}
		uid := vapi.JobID("xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx8")

		Convey("ValidNPUJobFn() should return error for job with invalid request number 0", func() {
			tasks = append(tasks, vapi.NewTaskInfo(buildNPUPod(
				CPodInfo{namespace: "default", groupName: "npu-group-3",
					podName: "npu-test-3", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
					reqNPUType: a300TNPUCardName, reqNpuNum: invalidNum0})))
			job := vapi.NewJobInfo(uid, tasks...)
			result := npu.ValidNPUJobFn(job)
			So(result, ShouldNotBeNil)
		})
		Convey("ValidNPUJobFn() should return error for job with invalid request number 5", func() {
			tasks = append(tasks, vapi.NewTaskInfo(buildNPUPod(
				CPodInfo{namespace: "default", groupName: "npu-group-4",
					podName: "npu-test-4", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
					reqNPUType: a300TNPUCardName, reqNpuNum: invalidNum5})))
			job := vapi.NewJobInfo(uid, tasks...)
			result := npu.ValidNPUJobFn(job)
			So(result, ShouldNotBeNil)
		})
		Convey("ValidNPUJobFn() should return error for job with valid request number 1", func() {
			tasks = append(tasks, vapi.NewTaskInfo(buildNPUPod(
				CPodInfo{namespace: "default", groupName: "npu-group-5",
					podName: "npu-test-5", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
					reqNPUType: a300TNPUCardName, reqNpuNum: validNum1})))
			job := vapi.NewJobInfo(uid, tasks...)
			result := npu.ValidNPUJobFn(job)
			So(result, ShouldBeNil)
		})
	})
}

// TestCNPUValidNPUJobFnInvalidModel
func TestCNPUValidNPUJobFnInvalidModel(t *testing.T) {
	Convey("Test card910x2 ValidNPUJobFnInvalidModel", t, func() {
		const (
			num2 = "2"
			num8 = "8"
		)
		npu := &card910x2{}
		tasks := []*vapi.TaskInfo{}
		uid := vapi.JobID("xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx9")

		Convey("ValidNPUJobFn() should return error for job with invalid single model", func() {
			tasks = append(tasks, vapi.NewTaskInfo(buildNPUPod(
				CPodInfo{namespace: "default", groupName: "npu-group-6",
					podName: "npu-test-6", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
					reqNPUType: a300TNPUCardName, reqNpuNum: num8})))
			job := vapi.NewJobInfo(uid, tasks...)
			result := npu.ValidNPUJobFn(job)
			So(result, ShouldNotBeNil)
		})
		Convey("ValidNPUJobFn() should return nil for job with valid model", func() {
			tasks = append(tasks, vapi.NewTaskInfo(buildNPUPod(
				CPodInfo{namespace: "default", groupName: "npu-group-7",
					podName: "npu-test-7", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
					reqNPUType: a300TNPUCardName, reqNpuNum: num2})))
			tasks = append(tasks, vapi.NewTaskInfo(buildNPUPod(
				CPodInfo{namespace: "default", groupName: "npu-group-8",
					podName: "npu-test-8", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
					reqNPUType: a300TNPUCardName, reqNpuNum: num2})))
			job := vapi.NewJobInfo(uid, tasks...)
			result := npu.ValidNPUJobFn(job)
			So(result, ShouldBeNil)
		})
	})
}

// TestCNPUPreCheckNodeFnTaskError
func TestCNPUPreCheckNodeFnTaskError(t *testing.T) {
	Convey("Test card910x2 PreCheckNodeFnTaskError", t, func() {
		npu := &card910x2{}
		confs := []conf.Configuration{}
		node := buildNPUNode(CNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
			"1", "Ascend910-1"})

		Convey("PreCheckNodeFn() should return error when task don't have selector", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-9",
				podName: "npu-test-9", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a300TNPUCardName, reqNpuNum: "1"})
			setPodSelector(pod, archSelector, "")
			setPodSelector(pod, acceleratorType, "")
			task := vapi.NewTaskInfo(pod)
			result := npu.PreCheckNodeFn(task, node, confs)
			So(result, ShouldBeError)
		})
		Convey("PreCheckNodeFn() should return nil when task isn't npu task", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-10",
				podName: "npu-test-10", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: "", reqNpuNum: ""})
			setPodSelector(pod, archSelector, "")
			setPodSelector(pod, acceleratorType, "")
			task := vapi.NewTaskInfo(pod)
			result := npu.PreCheckNodeFn(task, node, confs)
			So(result, ShouldBeNil)
		})
	})
}

// TestCNPUPreCheckNodeFnNodeError
func TestCNPUPreCheckNodeFnNodeError(t *testing.T) {
	Convey("Test card910x2 PreCheckNodeFnNodeError", t, func() {
		npu := &card910x2{}
		confs := []conf.Configuration{}
		node := buildNPUNode(CNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
			"1", "Ascend910-3"})

		Convey("PreCheckNodeFn() should return error when node don't have label", func() {
			task := vapi.NewTaskInfo(buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-11",
				podName: "npu-test-11", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a300TNPUCardName, reqNpuNum: "1"}))
			setNodeLabel(node.Node, archSelector, "")
			setNodeLabel(node.Node, acceleratorType, "")
			result := npu.PreCheckNodeFn(task, node, confs)
			So(result, ShouldBeError)
		})
		Convey("PreCheckNodeFn() should return error when selectors arch mismatch with labels", func() {
			task := vapi.NewTaskInfo(buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-12",
				podName: "npu-test-12", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a300TNPUCardName, reqNpuNum: "1"}))
			// build a node with mismatch selector
			nodeArm := buildNPUNode(CNodeInfo{nodeName, huaweiArchArm, "192", "755Gi",
				"1", "Ascend910-5"})
			result := npu.PreCheckNodeFn(task, nodeArm, confs)
			So(result, ShouldBeError)
		})
		Convey("PreCheckNodeFn() should return error when selectors model mismatch with labels", func() {
			task := vapi.NewTaskInfo(buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-13",
				podName: "npu-test-13", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a300TNPUCardName, reqNpuNum: "1"}))
			// build a node with mismatch selector
			setNodeLabel(node.Node, acceleratorType, moduleAcceleratorType)
			result := npu.PreCheckNodeFn(task, node, confs)
			So(result, ShouldBeError)
		})
	})
}

// TestCnpuPreCheckNodeFnSuccess
func TestCnpuPreCheckNodeFnSuccess(t *testing.T) {
	Convey("Test card910x2 PreCheckNodeFnSuccess", t, func() {
		npu := &card910x2{}
		confs := []conf.Configuration{}
		node := buildNPUNode(CNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
			"1", "Ascend910-6"})

		Convey("PreCheckNodeFn() should return nil when selectors match with labels", func() {
			// build a task with no selector
			task := vapi.NewTaskInfo(buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-14",
				podName: "npu-test-14", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a300TNPUCardName, reqNpuNum: "1"}))
			result := npu.PreCheckNodeFn(task, node, confs)
			So(result, ShouldBeNil)
		})
	})
}

// TestCnpuCheckNPUResourceStableFn
func TestCnpuCheckNPUResourceStableFn(t *testing.T) {
	Convey("Test card910x2 CheckNPUResourceStableFn", t, func() {
		vnpu := &card910x2{}

		Convey("CheckNPUResourceStableFn() should return error when there's missing resource type in idle", func() {
			node := buildNPUNode(CNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
				"0", ""})
			node.Node.Annotations[a300TNPUCardName] = "Ascend910-2"
			result := vnpu.CheckNPUResourceStableFn(node)
			So(result, ShouldBeError)
		})
		Convey("CheckNPUResourceStableFn() should return error when node resources are unstable", func() {
			node := buildNPUNode(CNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
				"1", ""})
			result := vnpu.CheckNPUResourceStableFn(node)
			So(result, ShouldBeError)
		})
		Convey("CheckNPUResourceStableFn() should return nil when node resources are stable", func() {
			node := buildNPUNode(CNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
				"2", "Ascend910-0,Ascend910-4"})
			result := vnpu.CheckNPUResourceStableFn(node)
			So(result, ShouldBeNil)
		})
	})
}

// TestCnpuCheckNodeNPUByTaskFn
func TestCnpuCheckNodeNPUByTaskFn(t *testing.T) {
	Convey("Test job CheckNodeNPUByTaskFn", t, func() {
		npu := &card910x2{}
		pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-15",
			podName: "npu-test-15", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: a300TNPUCardName, reqNpuNum: "2"})
		task := vapi.NewTaskInfo(pod)

		Convey("CheckNodeNPUByTaskFn() should return error when node doesn't meet task requests", func() {
			node := buildNPUNode(CNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
				"1", "Ascend910-4"})
			result := npu.CheckNodeNPUByTaskFn(task, node, true)
			So(result, ShouldBeError)
		})
		Convey("CheckNodeNPUByTaskFn() should return error when node meets task requests", func() {
			node := buildNPUNode(CNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
				"2", "Ascend910-1,Ascend910-4"})
			result := npu.CheckNodeNPUByTaskFn(task, node, true)
			So(result, ShouldBeNil)
		})
	})
}

// TestCnpuCheckNodeNPUByTaskFnFail
func TestCnpuCheckNodeNPUByTaskFnError(t *testing.T) {
	Convey("Test job CheckNodeNPUByTaskFnError", t, func() {
		npu := &card910x2{}

		Convey("CheckNodeNPUByTaskFn() should return error when the requirement of the task is 0", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-115",
				podName: "npu-test-115", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a300TNPUCardName, reqNpuNum: "0"})
			task := vapi.NewTaskInfo(pod)
			node := buildNPUNode(CNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
				"2", "Ascend910-1,Ascend910-4"})
			result := npu.CheckNodeNPUByTaskFn(task, node, true)
			So(result, ShouldBeError)
		})

		Convey("CheckNodeNPUByTaskFn() should return error when the number of NUP on a node is 0", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-116",
				podName: "npu-test-116", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a300TNPUCardName, reqNpuNum: "2"})
			task := vapi.NewTaskInfo(pod)
			node := buildNPUNode(CNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
				"", ""})
			result := npu.CheckNodeNPUByTaskFn(task, node, true)
			So(result, ShouldBeError)
		})
	})
}

// TestCnpuGetNPUAffinityBestNodesFn
func TestCnpuGetNPUAffinityBestNodesFn(t *testing.T) {
	Convey("Test card910x2 GetNPUAffinityBestNodesFn", t, func() {
		const (
			nodeName1 = "centos1"
			nodeName2 = "centos2"
			constNum  = 0
		)
		npu := &card910x2{}

		Convey("GetNPUAffinityBestNodesFn() should return correct result when request NPU num is 2", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-16",
				podName: "npu-test-16", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a300TNPUCardName, reqNpuNum: "2"})
			task := vapi.NewTaskInfo(pod)
			nodes := []*vapi.NodeInfo{}
			node1 := buildNPUNode(CNodeInfo{nodeName1, huaweiArchX86, "192", "755Gi",
				"1", "Ascend910-0"})
			nodes = append(nodes, node1)
			node2 := buildNPUNode(CNodeInfo{nodeName2, huaweiArchX86, "192", "755Gi",
				"2", "Ascend910-4,Ascend910-1"})
			nodes = append(nodes, node2)

			result, err := npu.GetNPUAffinityBestNodesFn(task, nodes, false)

			So(result[nodeName2], ShouldEqual, constNum)
			So(err, ShouldBeNil)
		})

		Convey("GetNPUAffinityBestNodesFn() should return correct result when request NPU num is 1", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-17",
				podName: "npu-test-17", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a300TNPUCardName, reqNpuNum: "1"})
			task := vapi.NewTaskInfo(pod)
			nodes := []*vapi.NodeInfo{}
			node1 := buildNPUNode(CNodeInfo{nodeName1, huaweiArchX86, "192", "755Gi",
				"1", "Ascend910-0"})
			nodes = append(nodes, node1)
			node2 := buildNPUNode(CNodeInfo{nodeName2, huaweiArchX86, "192", "755Gi",
				"2", "Ascend910-4,Ascend910-1"})
			nodes = append(nodes, node2)

			result, err := npu.GetNPUAffinityBestNodesFn(task, nodes, false)

			So(result[nodeName1], ShouldEqual, constNum)
			So(err, ShouldBeNil)
		})
	})
}

// TestCnpuGetNPUAffinityBestNodesFn
func TestCnpuGetNPUAffinityBestNodesFnError(t *testing.T) {
	Convey("Test card910x2 GetNPUAffinityBestNodesFnError", t, func() {
		const (
			nodeName1 = "centos1"
			nodeName2 = "centos2"
		)
		npu := &card910x2{}

		pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-117",
			podName: "npu-test-117", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: a300TNPUCardName, reqNpuNum: "2"})
		task := vapi.NewTaskInfo(pod)
		nodes := []*vapi.NodeInfo{}

		Convey("GetNPUAffinityBestNodesFn() should return error when node is empty", func() {
			_, err := npu.GetNPUAffinityBestNodesFn(task, nodes, false)
			So(err, ShouldBeError)
		})

		node := buildNPUNode(CNodeInfo{nodeName1, huaweiArchX86, "192", "755Gi",
			"1", "Ascend910-0"})
		nodes = append(nodes, node)

		Convey("GetNPUAffinityBestNodesFn() should return error when none bestNodes", func() {
			result, err := npu.GetNPUAffinityBestNodesFn(task, nodes, false)
			So(result[nodeName2], ShouldEqual, 0)
			So(err, ShouldBeError)
		})

		Convey("GetNPUAffinityBestNodesFn() should return error when request NPU number is 0", func() {
			task.Resreq.ScalarResources[a300TNPUCardName] = 0
			_, err := npu.GetNPUAffinityBestNodesFn(task, nodes, false)
			So(err, ShouldBeError)
		})

		Convey("GetNPUAffinityBestNodesFn() should return error when request number is greater than 3", func() {
			task.Resreq.ScalarResources[a300TNPUCardName] = constInt3000
			_, err := npu.GetNPUAffinityBestNodesFn(task, nodes, false)
			So(err, ShouldBeError)
		})
	})
}

// TestCnpuScoreBestNPUNodesFn
func TestCnpuScoreBestNPUNodesFn(t *testing.T) {
	Convey("Test card910x2 ScoreBestNPUNodesFn", t, func() {
		const (
			nodeName1 = "euler1"
			nodeName2 = "euler2"
		)
		npu := &card910x2{}
		bestNodes := map[string]int{
			nodeName1: 0,
			nodeName2: constIntNum3,
		}
		pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-18",
			podName: "npu-test-18", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: a300TNPUCardName, reqNpuNum: "4"})
		task := vapi.NewTaskInfo(pod)

		nodes := []*vapi.NodeInfo{}
		node1 := buildNPUNode(CNodeInfo{nodeName1, huaweiArchX86, "192", "755Gi",
			"8", "Ascend910-5,Ascend910-2,Ascend910-0,Ascend910-1,Ascend910-3," +
				"Ascend910-6,Ascend910-7"})
		nodes = append(nodes, node1)
		node2 := buildNPUNode(CNodeInfo{nodeName2, huaweiArchArm, "192", "755Gi",
			"8", "Ascend910-5,Ascend910-2,Ascend910-0,Ascend910-1,Ascend910-3," +
				"Ascend910-6,Ascend910-7"})
		nodes = append(nodes, node2)

		Convey("ScoreBestNPUNodesFn() should return correct result", func() {
			scoreMap := make(map[string]float64)
			expectedResult := map[string]float64{
				nodeName1: 32.00,
				nodeName2: 29.00,
			}
			result, err := npu.ScoreBestNPUNodesFn(scoreMap, bestNodes, task, nodes)
			So(err, ShouldBeNil)
			So(result, ShouldResemble, expectedResult)
		})
	})
}

// TestCnpuScoreBestNPUNodesFnError
func TestCnpuScoreBestNPUNodesFnError(t *testing.T) {
	Convey("Test card910x2 ScoreBestNPUNodesFnError", t, func() {
		const (
			nodeName1 = "euler1"
			nodeName2 = "euler2"
			nodeName3 = "euler3"
			nodeName4 = "euler4"
		)
		npu := &card910x2{}
		bestNodes := map[string]int{
			nodeName1: 0,
			nodeName2: constIntNum3,
		}
		pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-121",
			podName: "npu-test-121", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: a300TNPUCardName, reqNpuNum: "4"})
		task := vapi.NewTaskInfo(pod)

		nodes := []*vapi.NodeInfo{}
		node1 := buildNPUNode(CNodeInfo{nodeName3, huaweiArchX86, "192", "755Gi",
			"8", "Ascend910-5,Ascend910-2,Ascend910-0,Ascend910-1,Ascend910-3," +
				"Ascend910-6,Ascend910-7"})
		nodes = append(nodes, node1)
		node2 := buildNPUNode(CNodeInfo{nodeName4, huaweiArchArm, "192", "755Gi",
			"8", "Ascend910-5,Ascend910-2,Ascend910-0,Ascend910-1,Ascend910-3," +
				"Ascend910-6,Ascend910-7"})
		nodes = append(nodes, node2)

		Convey("ScoreBestNPUNodesFn() should return err when scoreMap is nil", func() {
			var scoreMap map[string]float64
			_, err := npu.ScoreBestNPUNodesFn(scoreMap, bestNodes, task, nodes)
			So(err, ShouldBeError)
		})

		Convey("ScoreBestNPUNodesFn() should return 0 of scores when node name mismatch", func() {
			scoreMap := make(map[string]float64)
			expectedResult := map[string]float64{
				nodeName1: 0.0,
				nodeName2: 0.0,
			}
			result, err := npu.ScoreBestNPUNodesFn(scoreMap, bestNodes, task, nodes)
			So(err, ShouldBeNil)
			So(result, ShouldResemble, expectedResult)
		})
	})
}

// TestCnpuUpdateNPUNodeUsedCardFn1
func TestCnpuUpdateNPUNodeUsedCardFn1(t *testing.T) {
	Convey("Test card910x2 UpdateNPUNodeUsedCardFn", t, func() {
		npu := &card910x2{}
		node := buildNPUNode(CNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
			"8", "Ascend910-2,Ascend910-0,Ascend910-1,Ascend910-3,Ascend910-6,Ascend910-4," +
				"Ascend910-5,Ascend910-7"})

		Convey("UpdateNPUNodeUsedCardFn() should successfully update node.others", func() {
			top := []int{0, 4}
			expectedResult := map[string]interface{}{
				a300TNPUCardName: "Ascend910-2,Ascend910-1,Ascend910-3,Ascend910-6,Ascend910-5,Ascend910-7",
			}
			err := npu.UpdateNPUNodeUsedCardFn(node, top)
			So(err, ShouldBeNil)
			So(node.Others, ShouldResemble, expectedResult)
		})
	})
}

// TestCnpuUpdateNPUNodeUsedCardFnError2
func TestCnpuUpdateNPUNodeUsedCardFn2(t *testing.T) {
	Convey("Test card910x2 UpdateNPUNodeUsedCardFn", t, func() {
		npu := &card910x2{}

		Convey("UpdateNPUNodeUsedCardFn() should should successfully update node.others", func() {
			node := buildNPUNode(CNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
				"8", "Ascend910-2,Ascend910-0,Ascend910-1,Ascend910-3,Ascend910-6,Ascend910-4," +
					"Ascend910-5,Ascend910-7"})
			node.Others = map[string]interface{}{
				a300TNPUCardName: "Ascend910-2,Ascend910-0,Ascend910-1,Ascend910-3,Ascend910-6,Ascend910-4," +
					"Ascend910-5,Ascend910-7",
			}
			top := []int{0, 1, 2, 3, 4, 5, 6, 7}
			expectedResult := map[string]interface{}{
				a300TNPUCardName: "",
			}
			err := npu.UpdateNPUNodeUsedCardFn(node, top)
			So(err, ShouldBeNil)
			So(node.Others, ShouldResemble, expectedResult)
		})
	})
}

// TestCnpuUpdateNPUNodeUsedCardFnError
func TestCnpuUpdateNPUNodeUsedCardFnError1(t *testing.T) {
	Convey("Test card910x2 UpdateNPUNodeUsedCardFnError", t, func() {
		npu := &card910x2{}

		Convey("UpdateNPUNodeUsedCardFn() should return error when top's type mismatch", func() {
			node := buildNPUNode(CNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
				"8", "Ascend910-2,Ascend910-0,Ascend910-1,Ascend910-3,Ascend910-6,Ascend910-4," +
					"Ascend910-5,Ascend910-7"})
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
	Convey("Test card910x2 GetReleaseNPUTopologyFn", t, func() {
		npu := &card910x2{}
		task := vapi.NewTaskInfo(buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-19",
			podName: "npu-test-19", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: a300TNPUCardName, reqNpuNum: "2"}))
		Convey("GetReleaseNPUTopologyFn() should return correct card id slice", func() {
			task.Pod.Annotations[a300TNPUCardName] = "Ascend910-0,Ascend910-4"
			expectedResult := []int{0, 4}
			result, err := npu.GetReleaseNPUTopologyFn(task)
			So(err, ShouldBeNil)
			So(result, ShouldResemble, expectedResult)
		})
	})
}

// TestCnpuGetReleaseNPUTopologyFnError
func TestCnpuGetReleaseNPUTopologyFnError(t *testing.T) {
	Convey("Test card910x2 GetReleaseNPUTopologyFnError", t, func() {
		npu := &card910x2{}
		task := vapi.NewTaskInfo(buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-122",
			podName: "npu-test-122", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: a300TNPUCardName, reqNpuNum: "2"}))
		Convey("GetReleaseNPUTopologyFn() should return error when the annotations of pod is wrong", func() {
			task.Pod.Annotations[a300TNPUCardName] = "Ascend9100,Ascend9104"
			_, err := npu.GetReleaseNPUTopologyFn(task)
			So(err, ShouldBeError)
		})
	})
}

// TestCnpuUpdateReleaseNPUNodeTopologyFn
func TestCnpuUpdateReleaseNPUNodeTopologyFn(t *testing.T) {
	Convey("", t, func() {
		npu := &card910x2{}
		node := buildNPUNode(CNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
			"8", "Ascend910-2,Ascend910-1,Ascend910-3,Ascend910-6,Ascend910-5,Ascend910-7"})
		Convey("UpdateNPUNodeUsedCardFn() should successfully update node.others", func() {
			top := []int{0, 4}
			err := npu.UpdateReleaseNPUNodeTopologyFn(node, top)
			So(err, ShouldBeNil)
		})

	})
}

// TestCnpuUpdateReleaseNPUNodeTopologyFnError
func TestCnpuUpdateReleaseNPUNodeTopologyFnError(t *testing.T) {
	Convey("", t, func() {
		npu := &card910x2{}

		Convey("UpdateNPUNodeUsedCardFn() should return error when top's type is mismatch", func() {
			node := buildNPUNode(CNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
				"6", "Ascend910-2,Ascend910-1,Ascend910-3,Ascend910-6,Ascend910-5,Ascend910-7"})
			node.Others = map[string]interface{}{
				a300TNPUCardName: "Ascend910-0,Ascend910-4",
			}
			top := []string{"0", "4"}
			err := npu.UpdateReleaseNPUNodeTopologyFn(node, top)
			So(err, ShouldBeError)
		})

		Convey("UpdateNPUNodeUsedCardFn() should return error when node's others is wrong", func() {
			node := buildNPUNode(CNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
				"6", "Ascend910-2,Ascend910-1,Ascend910-3,Ascend910-6,Ascend910-5,Ascend910-7"})
			node.Others = map[string]interface{}{
				a300TNPUCardName: "Ascend910-0Ascend910-4",
			}
			top := []int{0, 4}
			err := npu.UpdateReleaseNPUNodeTopologyFn(node, top)
			So(err, ShouldBeError)
		})
	})
}

// TestCnpuGetAllocatedNPUFromTopologyFn
func TestCnpuGetAllocatedNPUFromTopologyFn(t *testing.T) {
	Convey("Test card910x2 GetAllocatedNPUFromTopologyFn", t, func() {
		npu := &card910x2{}

		Convey("GetAllocatedNPUFromTopologyFn() should return correct result when reqNpuNum is 2", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-20",
				podName: "npu-test-20", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a300TNPUCardName, reqNpuNum: "2"})
			task := vapi.NewTaskInfo(pod)
			node := buildNPUNode(CNodeInfo{nodeName, huaweiArchArm, "192", "755Gi",
				"2", "Ascend910-0,Ascend910-4"})
			expectedResult := []int{0, 4}
			result, err := npu.GetAllocatedNPUFromTopologyFn(task, node, false)
			So(err, ShouldBeNil)
			So(result, ShouldResemble, expectedResult)
		})

		Convey("GetAllocatedNPUFromTopologyFn() should return correct result when reqNpuNum is 1", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-21",
				podName: "npu-test-21", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a300TNPUCardName, reqNpuNum: "1"})
			task := vapi.NewTaskInfo(pod)
			node := buildNPUNode(CNodeInfo{nodeName, huaweiArchArm, "192", "755Gi",
				"2", "Ascend910-0,Ascend910-4"})
			expectedResult := []int{0}
			result, err := npu.GetAllocatedNPUFromTopologyFn(task, node, false)
			So(err, ShouldBeNil)
			So(result, ShouldResemble, expectedResult)
		})
	})
}

// TestCnpuGetAllocatedNPUFromTopologyFnError
func TestCnpuGetAllocatedNPUFromTopologyFnError(t *testing.T) {
	Convey("Test card910x2 GetAllocatedNPUFromTopologyFnError", t, func() {
		npu := &card910x2{}

		pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-123",
			podName: "npu-test-123", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: a300TNPUCardName, reqNpuNum: "2"})
		task := vapi.NewTaskInfo(pod)
		node := buildNPUNode(CNodeInfo{nodeName, huaweiArchArm, "192", "755Gi",
			"2", "Ascend910-0,Ascend910-4"})

		Convey("GetAllocatedNPUFromTopologyFn() should return error when reqNpuNum of pod is 0", func() {
			task.Resreq.ScalarResources[a300TNPUCardName] = 0
			_, err := npu.GetAllocatedNPUFromTopologyFn(task, node, false)
			So(err, ShouldBeError)
		})

		Convey("GetAllocatedNPUFromTopologyFn() should return error when reqNpuNum Is greater than 2", func() {
			task.Resreq.ScalarResources[a300TNPUCardName] = constInt3000
			_, err := npu.GetAllocatedNPUFromTopologyFn(task, node, false)
			So(err, ShouldBeError)
		})

		Convey("GetAllocatedNPUFromTopologyFn() should return nil when npuTop of node is wrong", func() {
			task.Resreq.ScalarResources[a300TNPUCardName] = constInt2000
			result, err := npu.GetAllocatedNPUFromTopologyFn(task, node, false)
			So(err, ShouldBeNil)
			So(reflect.DeepEqual(result, []int{0, npuNumPerHccs}), ShouldBeTrue)
		})

		Convey("GetAllocatedNPUFromTopologyFn() should return correct result when reqNpuNum is 0", func() {
			task.Resreq.ScalarResources[a300TNPUCardName] = 0
			result, err := npu.GetAllocatedNPUFromTopologyFn(task, node, false)
			So(err, ShouldBeError)
			So(result, ShouldBeNil)
		})

		Convey("GetAllocatedNPUFromTopologyFn() should return error when none node meet request", func() {
			task.Resreq.ScalarResources[a300TNPUCardName] = constInt2000
			_, err := npu.GetAllocatedNPUFromTopologyFn(task, node, false)
			So(err, ShouldBeNil)
		})
	})
}

// TestCnpuSetNPUTopologyToPodFn
func TestCnpuSetNPUTopologyToPodFn(t *testing.T) {
	Convey("Test card910x2 SetNPUTopologyToPodFn", t, func() {
		npu := &card910x2{}
		task := vapi.NewTaskInfo(buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-22",
			podName: "npu-test-22", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: a300TNPUCardName, reqNpuNum: "4"}))
		Convey("SetNPUTopologyToPodFn() should return error when top is of wrong type", func() {
			top := []string{}
			err := npu.SetNPUTopologyToPodFn(task, top)
			So(err, ShouldBeError)
			So(task.Pod.Annotations[a300TNPUCardName], ShouldEqual, "")
		})
		Convey("SetNPUTopologyToPodFn() should write correct info in pod annotation", func() {
			top := []int{0, 4}
			expectedResult := "Ascend910-0,Ascend910-4"
			err := npu.SetNPUTopologyToPodFn(task, top)
			So(err, ShouldBeNil)
			So(task.Pod.Annotations[a300TNPUCardName], ShouldEqual, expectedResult)
		})
	})
}

// TestCNPUIsMyTask
func TestCNPUIsMyTask(t *testing.T) {
	Convey("Test card910x2 IsMyTask", t, func() {
		npu := &card910x2{}

		Convey("IsMyTask() should return error when task doesn't request NPU", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-23",
				podName: "npu-test-23", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
				reqNPUType: a300TNPUCardName, reqNpuNum: "0"})
			task := vapi.NewTaskInfo(pod)
			result := npu.IsMyTask(task)
			So(result, ShouldBeError)
		})
		Convey("IsMyTask() should return error when task is of module mode", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-24",
				podName: "npu-test-24", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
				reqNPUType: a300TNPUCardName, reqNpuNum: "1"})
			setPodSelector(pod, acceleratorType, moduleAcceleratorType)
			task := vapi.NewTaskInfo(pod)
			result := npu.IsMyTask(task)
			So(result, ShouldBeError)
		})
		Convey("IsMyTask() should return nil when task is of card type", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-25",
				podName: "npu-test-25", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
				reqNPUType: a300TNPUCardName, reqNpuNum: "1"})
			task := vapi.NewTaskInfo(pod)
			result := npu.IsMyTask(task)
			So(result, ShouldBeNil)
		})
	})
}

// TestCNPUIsMyNode
func TestCNPUIsMyNode(t *testing.T) {
	Convey("Test card910x2 IsMyNode", t, func() {
		npu := &card910x2{}

		Convey("IsMyNode() should return error when node has no npu annotation", func() {
			node := buildNPUNode(CNodeInfo{nodeName, huaweiArchArm, "192", "755Gi",
				"0", ""})
			result := npu.IsMyNode(node)
			So(result, ShouldBeError)
		})
		Convey("IsMyNode() should return error when node is of module mode", func() {
			node := buildNPUNode(CNodeInfo{nodeName, huaweiArchArm, "192", "755Gi",
				"1", "Ascend910-10"})
			setNodeLabel(node.Node, acceleratorType, moduleAcceleratorType)
			result := npu.IsMyNode(node)
			So(result, ShouldBeError)
		})
		Convey("IsMyNode() should return nil when node is of card mode", func() {
			node := buildNPUNode(CNodeInfo{nodeName, huaweiArchArm, "192", "755Gi",
				"1", "Ascend910-11"})
			result := npu.IsMyNode(node)
			So(result, ShouldBeNil)
		})
	})
}

// TestCNPUIsMyJob
func TestCNPUIsMyJob(t *testing.T) {
	Convey("Test card910x2 IsMyJob", t, func() {
		npu := &card910x2{}
		tasks := []*vapi.TaskInfo{}
		uid := vapi.JobID("xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxx10")

		Convey("IsMyJob() should return error when job request no npu", func() {
			tasks = append(tasks, vapi.NewTaskInfo(buildNPUPod(
				CPodInfo{namespace: "default", groupName: "npu-group-26",
					podName: "npu-test-26", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
					reqNPUType: a300TNPUCardName, reqNpuNum: "0"})))
			job := vapi.NewJobInfo(uid, tasks...)
			result := npu.IsMyJob(job)
			So(result, ShouldBeError)
		})
		Convey("IsMyJob() should return error when job has no selector", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-27",
				podName: "npu-test-27", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
				reqNPUType: a300TNPUCardName, reqNpuNum: "1"})
			setPodSelector(pod, archSelector, "")
			setPodSelector(pod, acceleratorType, "")
			tasks = append(tasks, vapi.NewTaskInfo(pod))
			job := vapi.NewJobInfo(uid, tasks...)
			result := npu.IsMyJob(job)
			So(result, ShouldBeError)
		})
		Convey("IsMyJob() should return nil when job is of card mode", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-28",
				podName: "npu-test-28", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
				reqNPUType: a300TNPUCardName, reqNpuNum: "1"})
			tasks = append(tasks, vapi.NewTaskInfo(pod))
			job := vapi.NewJobInfo(uid, tasks...)
			result := npu.IsMyJob(job)
			So(result, ShouldBeNil)
		})
	})
}

func setPodSelector(CPod *v1.Pod, selectorKey string, selectorValue string) {
	if selectorValue == "" {
		delete(CPod.Spec.NodeSelector, selectorKey)
		return
	}
	CPod.Spec.NodeSelector[selectorKey] = selectorValue
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
	setPodSelector(pod, acceleratorType, cardAcceleratorType)

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
	nodeCapacity := buildNPUResourceList(CNode.cpu, CNode.mem, a300TNPUCardName, cs.Itoa(constIntNum2))
	nodeAlloc := buildNPUResourceList(CNode.cpu, CNode.mem, a300TNPUCardName, CNode.npuAllocateNum)
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
		v1node.Annotations[a300TNPUCardName] = CNode.npuTop
	}

	setNodeLabel(v1node, archSelector, CNode.nodeArch)
	setNodeLabel(v1node, acceleratorType, cardAcceleratorType)

	node := vapi.NewNodeInfo(v1node)
	if CNode.npuAllocateNum != "0" {
		node.Others = map[string]interface{}{
			a300TNPUCardName: CNode.npuTop,
		}
	}
	return node
}
