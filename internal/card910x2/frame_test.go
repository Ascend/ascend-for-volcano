/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package card910x2 is using for HuaWei A300T Ascend pin affinity schedule.

*/
package card910x2

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/conf"
	"volcano.sh/volcano/pkg/scheduler/util"

	util2 "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/util"
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
	npuNum2InK8s = 2000
	npuNum3InK8s = 3000
	node32Core   = 32.00
	node29Core   = 29.00
)

// TestCNPUName
func TestCNPUName(t *testing.T) {
	convey.Convey("Test card910x2 Name", t, func() {
		npu := &card910x2{}

		convey.Convey("Name() should return PluginName defined in const", func() {
			n := npu.Name()
			convey.So(n, convey.ShouldEqual, PluginName)
		})
	})
}

// TestCNPUValidNPUJobFnInvalidSelector
func TestCNPUValidNPUJobFnInvalidSelector(t *testing.T) {
	convey.Convey("Test card910x2 ValidNPUJobFnInvalidSelector", t, func() {
		const (
			invalidSelectorKey   = "invalid-key"
			invalidSelectorValue = "no-selector"
			validNum2            = 2000
		)
		npu := &card910x2{}
		var tasks []*api.TaskInfo
		uid := api.JobID("xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx5")

		convey.Convey("ValidNPUJobFn() should return error for job without certain selector key", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-1",
				podName: "npu-test-1", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a300TNPUCardName, reqNpuNum: "1"})
			setPodSelector(pod, acceleratorType, "")
			setPodSelector(pod, invalidSelectorKey, invalidSelectorValue)
			tasks = append(tasks, api.NewTaskInfo(pod))
			job := api.NewJobInfo(uid, tasks...)
			setJobResourceReq(job, a300TNPUCardName, float64(validNum2))
			result := npu.ValidNPUJobFn(job)
			convey.So(result, convey.ShouldNotBeNil)
		})
		convey.Convey("ValidNPUJobFn() should return error for job with invalid selector value", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-2",
				podName: "npu-test-2", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a300TNPUCardName, reqNpuNum: "1"})
			setPodSelector(pod, acceleratorType, invalidSelectorValue)
			tasks = append(tasks, api.NewTaskInfo(pod))
			job := api.NewJobInfo(uid, tasks...)
			result := npu.ValidNPUJobFn(job)
			convey.So(result, convey.ShouldNotBeNil)
		})
	})
}

// TestCNPUValidNPUJobFnInvalidNum
func TestCNPUValidNPUJobFnInvalidNum(t *testing.T) {
	convey.Convey("Test card910x2 ValidNPUJobFnInvalidNum", t, func() {
		const (
			invalidNum0 = "0"
			invalidNum5 = "5"
			validNum1   = "1"
		)
		npu := &card910x2{}
		var tasks []*api.TaskInfo
		uid := api.JobID("xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx8")

		convey.Convey("ValidNPUJobFn() should return error for job with invalid request number 0", func() {
			tasks = append(tasks, api.NewTaskInfo(buildNPUPod(
				CPodInfo{namespace: "default", groupName: "npu-group-3",
					podName: "npu-test-3", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
					reqNPUType: a300TNPUCardName, reqNpuNum: invalidNum0})))
			job := api.NewJobInfo(uid, tasks...)
			result := npu.ValidNPUJobFn(job)
			convey.So(result, convey.ShouldNotBeNil)
		})
		convey.Convey("ValidNPUJobFn() should return error for job with invalid request number 5", func() {
			tasks = append(tasks, api.NewTaskInfo(buildNPUPod(
				CPodInfo{namespace: "default", groupName: "npu-group-4",
					podName: "npu-test-4", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
					reqNPUType: a300TNPUCardName, reqNpuNum: invalidNum5})))
			job := api.NewJobInfo(uid, tasks...)
			result := npu.ValidNPUJobFn(job)
			convey.So(result, convey.ShouldNotBeNil)
		})
		convey.Convey("ValidNPUJobFn() should return error for job with valid request number 1", func() {
			tasks = append(tasks, api.NewTaskInfo(buildNPUPod(
				CPodInfo{namespace: "default", groupName: "npu-group-5",
					podName: "npu-test-5", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
					reqNPUType: a300TNPUCardName, reqNpuNum: validNum1})))
			job := api.NewJobInfo(uid, tasks...)
			result := npu.ValidNPUJobFn(job)
			convey.So(result, convey.ShouldBeNil)
		})
	})
}

// TestCNPUValidNPUJobFnInvalidModel
func TestCNPUValidNPUJobFnInvalidModel(t *testing.T) {
	convey.Convey("Test card910x2 ValidNPUJobFnInvalidModel", t, func() {
		const (
			num2 = "2"
			num8 = "8"
		)
		npu := &card910x2{}
		var tasks []*api.TaskInfo
		uid := api.JobID("xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx9")

		convey.Convey("ValidNPUJobFn() should return error for job with invalid single model", func() {
			tasks = append(tasks, api.NewTaskInfo(buildNPUPod(
				CPodInfo{namespace: "default", groupName: "npu-group-6",
					podName: "npu-test-6", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
					reqNPUType: a300TNPUCardName, reqNpuNum: num8})))
			job := api.NewJobInfo(uid, tasks...)
			result := npu.ValidNPUJobFn(job)
			convey.So(result, convey.ShouldNotBeNil)
		})
		convey.Convey("ValidNPUJobFn() should return nil for job with valid model", func() {
			tasks = append(tasks, api.NewTaskInfo(buildNPUPod(
				CPodInfo{namespace: "default", groupName: "npu-group-7",
					podName: "npu-test-7", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
					reqNPUType: a300TNPUCardName, reqNpuNum: num2})))
			tasks = append(tasks, api.NewTaskInfo(buildNPUPod(
				CPodInfo{namespace: "default", groupName: "npu-group-8",
					podName: "npu-test-8", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
					reqNPUType: a300TNPUCardName, reqNpuNum: num2})))
			job := api.NewJobInfo(uid, tasks...)
			result := npu.ValidNPUJobFn(job)
			convey.So(result, convey.ShouldBeNil)
		})
	})
}

// TestCNPUPreCheckNodeFnTaskError
func TestCNPUPreCheckNodeFnTaskError(t *testing.T) {
	convey.Convey("Test card910x2 PreCheckNodeFnTaskError", t, func() {
		npu := &card910x2{}
		var confs []conf.Configuration
		node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
			npuAllocateNum: "1", npuTop: "Ascend910-1"})

		convey.Convey("PreCheckNodeFn() should return error when task don't have selector", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-9",
				podName: "npu-test-9", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a300TNPUCardName, reqNpuNum: "1"})
			setPodSelector(pod, archSelector, "")
			setPodSelector(pod, acceleratorType, "")
			task := api.NewTaskInfo(pod)
			result := npu.PreCheckNodeFn(task, node, confs)
			convey.So(result, convey.ShouldBeError)
		})
		convey.Convey("PreCheckNodeFn() should return nil when task isn't npu task", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-10",
				podName: "npu-test-10", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: "", reqNpuNum: ""})
			setPodSelector(pod, archSelector, "")
			setPodSelector(pod, acceleratorType, "")
			task := api.NewTaskInfo(pod)
			result := npu.PreCheckNodeFn(task, node, confs)
			convey.So(result, convey.ShouldBeNil)
		})
	})
}

// TestCNPUPreCheckNodeFnNodeError
func TestCNPUPreCheckNodeFnNodeError(t *testing.T) {
	convey.Convey("Test card910x2 PreCheckNodeFnNodeError", t, func() {
		npu := &card910x2{}
		var confs []conf.Configuration
		node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
			npuAllocateNum: "1", npuTop: "Ascend910-3"})

		convey.Convey("PreCheckNodeFn() should return error when node don't have label", func() {
			task := api.NewTaskInfo(buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-11",
				podName: "npu-test-11", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a300TNPUCardName, reqNpuNum: "1"}))
			setNodeLabel(node.Node, archSelector, "")
			setNodeLabel(node.Node, acceleratorType, "")
			result := npu.PreCheckNodeFn(task, node, confs)
			convey.So(result, convey.ShouldBeError)
		})
		convey.Convey("PreCheckNodeFn() should return error when selectors arch mismatch with labels", func() {
			task := api.NewTaskInfo(buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-12",
				podName: "npu-test-12", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a300TNPUCardName, reqNpuNum: "1"}))
			// build a node with mismatch selector
			nodeArm := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchArm, cpu: "192", mem: "755Gi",
				npuAllocateNum: "1", npuTop: "Ascend910-5"})
			result := npu.PreCheckNodeFn(task, nodeArm, confs)
			convey.So(result, convey.ShouldBeError)
		})
		convey.Convey("PreCheckNodeFn() should return error when selectors model mismatch with labels", func() {
			task := api.NewTaskInfo(buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-13",
				podName: "npu-test-13", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a300TNPUCardName, reqNpuNum: "1"}))
			// build a node with mismatch selector
			setNodeLabel(node.Node, acceleratorType, moduleAcceleratorType)
			result := npu.PreCheckNodeFn(task, node, confs)
			convey.So(result, convey.ShouldBeError)
		})
	})
}

// TestCnpuPreCheckNodeFnSuccess
func TestCnpuPreCheckNodeFnSuccess(t *testing.T) {
	convey.Convey("Test card910x2 PreCheckNodeFnSuccess", t, func() {
		npu := &card910x2{}
		var confs []conf.Configuration
		node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
			npuAllocateNum: "1", npuTop: "Ascend910-6"})

		convey.Convey("PreCheckNodeFn() should return nil when selectors match with labels", func() {
			// build a task with no selector
			task := api.NewTaskInfo(buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-14",
				podName: "npu-test-14", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a300TNPUCardName, reqNpuNum: "1"}))
			result := npu.PreCheckNodeFn(task, node, confs)
			convey.So(result, convey.ShouldBeNil)
		})
	})
}

// TestCnpuCheckNPUResourceStableFn
func TestCnpuCheckNPUResourceStableFn(t *testing.T) {
	convey.Convey("Test card910x2 CheckNPUResourceStableFn", t, func() {
		vnpu := &card910x2{}

		convey.Convey("CheckNPUResourceStableFn() missing resource type in idle", func() {
			node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "0", npuTop: ""})
			node.Node.Annotations[a300TNPUCardName] = "Ascend910-2"
			result := vnpu.CheckNPUResourceStableFn(node)
			convey.So(result, convey.ShouldBeError)
		})
		convey.Convey("CheckNPUResourceStableFn() should return error when node resources are unstable", func() {
			node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "1", npuTop: ""})
			result := vnpu.CheckNPUResourceStableFn(node)
			convey.So(result, convey.ShouldBeError)
		})
		convey.Convey("CheckNPUResourceStableFn() should return nil when node resources are stable", func() {
			node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "2", npuTop: "Ascend910-0,Ascend910-4"})
			result := vnpu.CheckNPUResourceStableFn(node)
			convey.So(result, convey.ShouldBeNil)
		})
	})
}

// TestCnpuCheckNodeNPUByTaskFn
func TestCnpuCheckNodeNPUByTaskFn(t *testing.T) {
	convey.Convey("Test job CheckNodeNPUByTaskFn", t, func() {
		npu := &card910x2{}
		pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-15",
			podName: "npu-test-15", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: a300TNPUCardName, reqNpuNum: "2"})
		task := api.NewTaskInfo(pod)

		convey.Convey("CheckNodeNPUByTaskFn() should return error when node doesn't meet task requests", func() {
			node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "1", npuTop: "Ascend910-4"})
			result := npu.CheckNodeNPUByTaskFn(task, node, true)
			convey.So(result, convey.ShouldBeError)
		})
		convey.Convey("CheckNodeNPUByTaskFn() should return error when node meets task requests", func() {
			node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "2", npuTop: "Ascend910-1,Ascend910-4"})
			result := npu.CheckNodeNPUByTaskFn(task, node, true)
			convey.So(result, convey.ShouldBeNil)
		})
	})
}

// TestCnpuCheckNodeNPUByTaskFnFail
func TestCnpuCheckNodeNPUByTaskFnError(t *testing.T) {
	convey.Convey("Test job CheckNodeNPUByTaskFnError", t, func() {
		npu := &card910x2{}

		convey.Convey("CheckNodeNPUByTaskFn() should return error when the requirement of the task is 0", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-115",
				podName: "npu-test-115", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a300TNPUCardName, reqNpuNum: "0"})
			task := api.NewTaskInfo(pod)
			node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "2", npuTop: "Ascend910-1,Ascend910-4"})
			result := npu.CheckNodeNPUByTaskFn(task, node, true)
			convey.So(result, convey.ShouldBeError)
		})

		convey.Convey("CheckNodeNPUByTaskFn() should return error when the number of NUP on a node is 0", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-116",
				podName: "npu-test-116", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a300TNPUCardName, reqNpuNum: "2"})
			task := api.NewTaskInfo(pod)
			node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "", npuTop: ""})
			result := npu.CheckNodeNPUByTaskFn(task, node, true)
			convey.So(result, convey.ShouldBeError)
		})
	})
}

// TestCnpuGetNPUAffinityBestNodesFn
func TestCnpuGetNPUAffinityBestNodesFn(t *testing.T) {
	convey.Convey("Test card910x2 GetNPUAffinityBestNodesFn", t, func() {
		const (
			nodeName1 = "centos1"
			nodeName2 = "centos2"
			constNum  = 0
		)
		npu := &card910x2{}

		convey.Convey("GetNPUAffinityBestNodesFn() should return correct result when request NPU num is 2", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-16",
				podName: "npu-test-16", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a300TNPUCardName, reqNpuNum: "2"})
			task := api.NewTaskInfo(pod)
			var nodes []*api.NodeInfo
			node1 := buildNPUNode(CNodeInfo{nodeName: nodeName1, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "1", npuTop: "Ascend910-0"})
			nodes = append(nodes, node1)
			node2 := buildNPUNode(CNodeInfo{nodeName: nodeName2, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "2", npuTop: "Ascend910-4,Ascend910-1"})
			nodes = append(nodes, node2)

			result, err := npu.GetNPUAffinityBestNodesFn(task, nodes, false)

			convey.So(result[nodeName2], convey.ShouldEqual, constNum)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("GetNPUAffinityBestNodesFn() should return correct result when request NPU num is 1", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-17",
				podName: "npu-test-17", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a300TNPUCardName, reqNpuNum: "1"})
			task := api.NewTaskInfo(pod)
			var nodes []*api.NodeInfo
			node1 := buildNPUNode(CNodeInfo{nodeName: nodeName1, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "1", npuTop: "Ascend910-0"})
			nodes = append(nodes, node1)
			node2 := buildNPUNode(CNodeInfo{nodeName: nodeName2, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "2", npuTop: "Ascend910-4,Ascend910-1"})
			nodes = append(nodes, node2)

			result, err := npu.GetNPUAffinityBestNodesFn(task, nodes, false)

			convey.So(result[nodeName1], convey.ShouldEqual, constNum)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestCnpuGetNPUAffinityBestNodesFn
func TestCnpuGetNPUAffinityBestNodesFnError(t *testing.T) {
	convey.Convey("Test card910x2 GetNPUAffinityBestNodesFnError", t, func() {
		const (
			nodeName1 = "centos1"
			nodeName2 = "centos2"
		)
		npu := &card910x2{}

		pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-117",
			podName: "npu-test-117", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: a300TNPUCardName, reqNpuNum: "2"})
		task := api.NewTaskInfo(pod)
		var nodes []*api.NodeInfo

		convey.Convey("GetNPUAffinityBestNodesFn() should return error when node is empty", func() {
			_, err := npu.GetNPUAffinityBestNodesFn(task, nodes, false)
			convey.So(err, convey.ShouldBeError)
		})

		node := buildNPUNode(CNodeInfo{nodeName: nodeName1, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
			npuAllocateNum: "1", npuTop: "Ascend910-0"})
		nodes = append(nodes, node)

		convey.Convey("GetNPUAffinityBestNodesFn() should return error when none bestNodes", func() {
			result, err := npu.GetNPUAffinityBestNodesFn(task, nodes, false)
			convey.So(result[nodeName2], convey.ShouldEqual, 0)
			convey.So(err, convey.ShouldBeError)
		})

		convey.Convey("GetNPUAffinityBestNodesFn() should return error when request NPU number is 0", func() {
			task.Resreq.ScalarResources[a300TNPUCardName] = 0
			_, err := npu.GetNPUAffinityBestNodesFn(task, nodes, false)
			convey.So(err, convey.ShouldBeError)
		})

		convey.Convey("GetNPUAffinityBestNodesFn() should return error when request number is greater than 3", func() {
			task.Resreq.ScalarResources[a300TNPUCardName] = npuNum3InK8s
			_, err := npu.GetNPUAffinityBestNodesFn(task, nodes, false)
			convey.So(err, convey.ShouldBeError)
		})
	})
}

// TestCnpuScoreBestNPUNodesFn
func TestCnpuScoreBestNPUNodesFn(t *testing.T) {
	convey.Convey("Test card910x2 ScoreBestNPUNodesFn", t, func() {
		const (
			nodeName1 = "euler1"
			nodeName2 = "euler2"
		)
		npu := &card910x2{}
		bestNodes := map[string]int{
			nodeName1: 0,
			nodeName2: util2.NPUIndex3,
		}
		pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-18",
			podName: "npu-test-18", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: a300TNPUCardName, reqNpuNum: "4"})
		task := api.NewTaskInfo(pod)

		var nodes []*api.NodeInfo
		node1 := buildNPUNode(CNodeInfo{nodeName: nodeName1, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
			npuAllocateNum: "8", npuTop: "Ascend910-5,Ascend910-2,Ascend910-0,Ascend910-1,Ascend910-3," +
				"Ascend910-6,Ascend910-7"})
		nodes = append(nodes, node1)
		node2 := buildNPUNode(CNodeInfo{nodeName: nodeName2, nodeArch: huaweiArchArm, cpu: "192", mem: "755Gi",
			npuAllocateNum: "8", npuTop: "Ascend910-5,Ascend910-2,Ascend910-0,Ascend910-1,Ascend910-3," +
				"Ascend910-6,Ascend910-7"})
		nodes = append(nodes, node2)

		convey.Convey("ScoreBestNPUNodesFn() should return correct result", func() {
			scoreMap := make(map[string]float64)
			expectedResult := map[string]float64{
				nodeName1: node32Core,
				nodeName2: node29Core,
			}
			result, err := npu.ScoreBestNPUNodesFn(scoreMap, bestNodes, task, nodes)
			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldResemble, expectedResult)
		})
	})
}

// TestCnpuScoreBestNPUNodesFnError
func TestCnpuScoreBestNPUNodesFnError(t *testing.T) {
	convey.Convey("Test card910x2 ScoreBestNPUNodesFnError", t, func() {
		const (
			nodeName1 = "euler1"
			nodeName2 = "euler2"
			nodeName3 = "euler3"
			nodeName4 = "euler4"
		)
		npu := &card910x2{}
		bestNodes := map[string]int{
			nodeName1: 0,
			nodeName2: util2.NPUIndex3,
		}
		pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-121",
			podName: "npu-test-121", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: a300TNPUCardName, reqNpuNum: "4"})
		task := api.NewTaskInfo(pod)

		var nodes []*api.NodeInfo
		node1 := buildNPUNode(CNodeInfo{nodeName: nodeName3, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
			npuAllocateNum: "8", npuTop: "Ascend910-5,Ascend910-2,Ascend910-0,Ascend910-1,Ascend910-3," +
				"Ascend910-6,Ascend910-7"})
		nodes = append(nodes, node1)
		node2 := buildNPUNode(CNodeInfo{nodeName: nodeName4, nodeArch: huaweiArchArm, cpu: "192", mem: "755Gi",
			npuAllocateNum: "8", npuTop: "Ascend910-5,Ascend910-2,Ascend910-0,Ascend910-1,Ascend910-3," +
				"Ascend910-6,Ascend910-7"})
		nodes = append(nodes, node2)

		convey.Convey("ScoreBestNPUNodesFn() should return err when scoreMap is nil", func() {
			var scoreMap map[string]float64
			_, err := npu.ScoreBestNPUNodesFn(scoreMap, bestNodes, task, nodes)
			convey.So(err, convey.ShouldBeError)
		})

		convey.Convey("ScoreBestNPUNodesFn() should return 0 of scores when node name mismatch", func() {
			scoreMap := make(map[string]float64)
			expectedResult := map[string]float64{
				nodeName1: 0.0,
				nodeName2: 0.0,
			}
			result, err := npu.ScoreBestNPUNodesFn(scoreMap, bestNodes, task, nodes)
			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldResemble, expectedResult)
		})
	})
}

// TestCnpuUpdateNPUNodeUsedCardFn1
func TestCnpuUpdateNPUNodeUsedCardFn1(t *testing.T) {
	convey.Convey("Test card910x2 UpdateNPUNodeUsedCardFn", t, func() {
		npu := &card910x2{}
		node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
			npuAllocateNum: "8", npuTop: "Ascend910-2,Ascend910-0,Ascend910-1,Ascend910-3,Ascend910-6,Ascend910-4," +
				"Ascend910-5,Ascend910-7"})

		convey.Convey("UpdateNPUNodeUsedCardFn() should successfully update node.others", func() {
			top := []int{0, util2.NPUIndex4}
			expectedResult := map[string]interface{}{
				a300TNPUCardName: "Ascend910-2,Ascend910-1,Ascend910-3,Ascend910-6,Ascend910-5,Ascend910-7",
			}
			err := npu.UpdateNPUNodeUsedCardFn(node, top)
			convey.So(err, convey.ShouldBeNil)
			convey.So(node.Others, convey.ShouldResemble, expectedResult)
		})
	})
}

// TestCnpuUpdateNPUNodeUsedCardFnError2
func TestCnpuUpdateNPUNodeUsedCardFn2(t *testing.T) {
	convey.Convey("Test card910x2 UpdateNPUNodeUsedCardFn", t, func() {
		npu := &card910x2{}

		convey.Convey("UpdateNPUNodeUsedCardFn() should should successfully update node.others", func() {
			node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "8", npuTop: "Ascend910-2,Ascend910-0,Ascend910-1,Ascend910-3,Ascend910-6,Ascend910-4," +
					"Ascend910-5,Ascend910-7"})
			node.Others = map[string]interface{}{
				a300TNPUCardName: "Ascend910-2,Ascend910-0,Ascend910-1,Ascend910-3,Ascend910-6,Ascend910-4," +
					"Ascend910-5,Ascend910-7",
			}
			top := []int{0, 1, util2.NPUIndex2, util2.NPUIndex3, npuNumPerHccs, util2.NPUIndex5, util2.NPUIndex6,
				util2.NPUIndex7}
			expectedResult := map[string]interface{}{
				a300TNPUCardName: "",
			}
			err := npu.UpdateNPUNodeUsedCardFn(node, top)
			convey.So(err, convey.ShouldBeNil)
			convey.So(node.Others, convey.ShouldResemble, expectedResult)
		})
	})
}

// TestCnpuUpdateNPUNodeUsedCardFnError
func TestCnpuUpdateNPUNodeUsedCardFnError1(t *testing.T) {
	convey.Convey("Test card910x2 UpdateNPUNodeUsedCardFnError", t, func() {
		npu := &card910x2{}

		convey.Convey("UpdateNPUNodeUsedCardFn() should return error when top's type mismatch", func() {
			node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "8", npuTop: "Ascend910-2,Ascend910-0,Ascend910-1,Ascend910-3,Ascend910-6,Ascend910-4," +
					"Ascend910-5,Ascend910-7"})
			top := []string{"0", "4"}
			err := npu.UpdateNPUNodeUsedCardFn(node, top)
			convey.So(err, convey.ShouldBeError)
		})

		convey.Convey("UpdateNPUNodeUsedCardFn() should return error when node's npuTop is empty", func() {
			node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "8", npuTop: ""})
			top := []int{0, util2.NPUIndex4}
			err := npu.UpdateNPUNodeUsedCardFn(node, top)
			convey.So(err, convey.ShouldBeError)
		})

	})
}

// TestCnpuGetReleaseNPUTopologyFn
func TestCnpuGetReleaseNPUTopologyFn(t *testing.T) {
	convey.Convey("Test card910x2 GetReleaseNPUTopologyFn", t, func() {
		npu := &card910x2{}
		task := api.NewTaskInfo(buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-19",
			podName: "npu-test-19", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: a300TNPUCardName, reqNpuNum: "2"}))
		convey.Convey("GetReleaseNPUTopologyFn() should return correct card id slice", func() {
			task.Pod.Annotations[a300TNPUCardName] = "Ascend910-0,Ascend910-4"
			expectedResult := []int{0, util2.NPUIndex4}
			result, err := npu.GetReleaseNPUTopologyFn(task)
			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldResemble, expectedResult)
		})
	})
}

// TestCnpuGetReleaseNPUTopologyFnError
func TestCnpuGetReleaseNPUTopologyFnError(t *testing.T) {
	convey.Convey("Test card910x2 GetReleaseNPUTopologyFnError", t, func() {
		npu := &card910x2{}
		task := api.NewTaskInfo(buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-122",
			podName: "npu-test-122", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: a300TNPUCardName, reqNpuNum: "2"}))
		convey.Convey("GetReleaseNPUTopologyFn() should return error when the annotations of pod is wrong", func() {
			task.Pod.Annotations[a300TNPUCardName] = "Ascend9100,Ascend9104"
			_, err := npu.GetReleaseNPUTopologyFn(task)
			convey.So(err, convey.ShouldBeError)
		})
	})
}

// TestCnpuUpdateReleaseNPUNodeTopologyFn
func TestCnpuUpdateReleaseNPUNodeTopologyFn(t *testing.T) {
	convey.Convey("", t, func() {
		npu := &card910x2{}
		node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
			npuAllocateNum: "8", npuTop: "Ascend910-2,Ascend910-1,Ascend910-3,Ascend910-6,Ascend910-5,Ascend910-7"})
		convey.Convey("UpdateNPUNodeUsedCardFn() should successfully update node.others", func() {
			top := []int{0, util2.NPUIndex4}
			err := npu.UpdateReleaseNPUNodeTopologyFn(node, top)
			convey.So(err, convey.ShouldBeNil)
		})

	})
}

// TestCnpuUpdateReleaseNPUNodeTopologyFnError
func TestCnpuUpdateReleaseNPUNodeTopologyFnError(t *testing.T) {
	convey.Convey("", t, func() {
		npu := &card910x2{}

		convey.Convey("UpdateNPUNodeUsedCardFn() should return error when top's type is mismatch", func() {
			node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "6", npuTop: "Ascend910-2,Ascend910-1,Ascend910-3,Ascend910-6,Ascend910-5,Ascend910-7"})
			node.Others = map[string]interface{}{
				a300TNPUCardName: "Ascend910-0,Ascend910-4",
			}
			top := []string{"0", "4"}
			err := npu.UpdateReleaseNPUNodeTopologyFn(node, top)
			convey.So(err, convey.ShouldBeError)
		})

		convey.Convey("UpdateNPUNodeUsedCardFn() should return error when node's others is wrong", func() {
			node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "6", npuTop: "Ascend910-2,Ascend910-1,Ascend910-3,Ascend910-6,Ascend910-5,Ascend910-7"})
			node.Others = map[string]interface{}{
				a300TNPUCardName: "Ascend910-0Ascend910-4",
			}
			top := []int{0, util2.NPUIndex4}
			err := npu.UpdateReleaseNPUNodeTopologyFn(node, top)
			convey.So(err, convey.ShouldBeError)
		})
	})
}

// TestCnpuGetAllocatedNPUFromTopologyFn
func TestCnpuGetAllocatedNPUFromTopologyFn(t *testing.T) {
	convey.Convey("Test card910x2 GetAllocatedNPUFromTopologyFn", t, func() {
		npu := &card910x2{}

		convey.Convey("GetAllocatedNPUFromTopologyFn() should return correct result when reqNpuNum is 2", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-20",
				podName: "npu-test-20", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a300TNPUCardName, reqNpuNum: "2"})
			task := api.NewTaskInfo(pod)
			node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchArm, cpu: "192", mem: "755Gi",
				npuAllocateNum: "2", npuTop: "Ascend910-0,Ascend910-4"})
			expectedResult := []int{0, npuNumPerHccs}
			result, err := npu.GetAllocatedNPUFromTopologyFn(task, node, false)
			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldResemble, expectedResult)
		})

		convey.Convey("GetAllocatedNPUFromTopologyFn() should return correct result when reqNpuNum is 1", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-21",
				podName: "npu-test-21", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a300TNPUCardName, reqNpuNum: "1"})
			task := api.NewTaskInfo(pod)
			node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchArm, cpu: "192", mem: "755Gi",
				npuAllocateNum: "2", npuTop: "Ascend910-0,Ascend910-4"})
			expectedResult := []int{0}
			result, err := npu.GetAllocatedNPUFromTopologyFn(task, node, false)
			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldResemble, expectedResult)
		})
	})
}

// TestCnpuGetAllocatedNPUFromTopologyFnError
func TestCnpuGetAllocatedNPUFromTopologyFnError(t *testing.T) {
	convey.Convey("Test card910x2 GetAllocatedNPUFromTopologyFnError", t, func() {
		npu := &card910x2{}

		pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-123",
			podName: "npu-test-123", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: a300TNPUCardName, reqNpuNum: "2"})
		task := api.NewTaskInfo(pod)
		node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchArm, cpu: "192", mem: "755Gi",
			npuAllocateNum: "2", npuTop: "Ascend910-0,Ascend910-4"})

		convey.Convey("GetAllocatedNPUFromTopologyFn() should return error when reqNpuNum of pod is 0", func() {
			task.Resreq.ScalarResources[a300TNPUCardName] = 0
			_, err := npu.GetAllocatedNPUFromTopologyFn(task, node, false)
			convey.So(err, convey.ShouldBeError)
		})

		convey.Convey("GetAllocatedNPUFromTopologyFn() should return error when reqNpuNum Is greater than 2", func() {
			task.Resreq.ScalarResources[a300TNPUCardName] = npuNum3InK8s
			_, err := npu.GetAllocatedNPUFromTopologyFn(task, node, false)
			convey.So(err, convey.ShouldBeError)
		})

		convey.Convey("GetAllocatedNPUFromTopologyFn() should return nil when npuTop of node is wrong", func() {
			task.Resreq.ScalarResources[a300TNPUCardName] = npuNum2InK8s
			result, err := npu.GetAllocatedNPUFromTopologyFn(task, node, false)
			convey.So(err, convey.ShouldBeNil)
			convey.So(reflect.DeepEqual(result, []int{0, npuNumPerHccs}), convey.ShouldBeTrue)
		})

		convey.Convey("GetAllocatedNPUFromTopologyFn() should return correct result when reqNpuNum is 0", func() {
			task.Resreq.ScalarResources[a300TNPUCardName] = 0
			result, err := npu.GetAllocatedNPUFromTopologyFn(task, node, false)
			convey.So(err, convey.ShouldBeError)
			convey.So(result, convey.ShouldBeNil)
		})

		convey.Convey("GetAllocatedNPUFromTopologyFn() should return error when none node meet request", func() {
			task.Resreq.ScalarResources[a300TNPUCardName] = npuNum2InK8s
			_, err := npu.GetAllocatedNPUFromTopologyFn(task, node, false)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestCnpuSetNPUTopologyToPodFn
func TestCnpuSetNPUTopologyToPodFn(t *testing.T) {
	convey.Convey("Test card910x2 SetNPUTopologyToPodFn", t, func() {
		npu := &card910x2{}
		task := api.NewTaskInfo(buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-22",
			podName: "npu-test-22", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: a300TNPUCardName, reqNpuNum: "4"}))
		convey.Convey("SetNPUTopologyToPodFn() should return error when top is of wrong type", func() {
			var top []string
			err := npu.SetNPUTopologyToPodFn(task, top)
			convey.So(err, convey.ShouldBeError)
			convey.So(task.Pod.Annotations[a300TNPUCardName], convey.ShouldEqual, "")
		})
		convey.Convey("SetNPUTopologyToPodFn() should write correct info in pod annotation", func() {
			top := []int{0, util2.NPUIndex4}
			expectedResult := "Ascend910-0,Ascend910-4"
			err := npu.SetNPUTopologyToPodFn(task, top)
			convey.So(err, convey.ShouldBeNil)
			convey.So(task.Pod.Annotations[a300TNPUCardName], convey.ShouldEqual, expectedResult)
		})
	})
}

// TestCNPUIsMyTask
func TestCNPUIsMyTask(t *testing.T) {
	convey.Convey("Test card910x2 IsMyTask", t, func() {
		npu := &card910x2{}

		convey.Convey("IsMyTask() should return error when task doesn't request NPU", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-23",
				podName: "npu-test-23", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
				reqNPUType: a300TNPUCardName, reqNpuNum: "0"})
			task := api.NewTaskInfo(pod)
			result := npu.IsMyTask(task)
			convey.So(result, convey.ShouldBeError)
		})
		convey.Convey("IsMyTask() should return error when task is of module mode", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-24",
				podName: "npu-test-24", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
				reqNPUType: a300TNPUCardName, reqNpuNum: "1"})
			setPodSelector(pod, acceleratorType, moduleAcceleratorType)
			task := api.NewTaskInfo(pod)
			result := npu.IsMyTask(task)
			convey.So(result, convey.ShouldBeError)
		})
		convey.Convey("IsMyTask() should return nil when task is of card type", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-25",
				podName: "npu-test-25", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
				reqNPUType: a300TNPUCardName, reqNpuNum: "1"})
			task := api.NewTaskInfo(pod)
			result := npu.IsMyTask(task)
			convey.So(result, convey.ShouldBeNil)
		})
	})
}

// TestCNPUIsMyNode
func TestCNPUIsMyNode(t *testing.T) {
	convey.Convey("Test card910x2 IsMyNode", t, func() {
		npu := &card910x2{}

		convey.Convey("IsMyNode() should return error when node has no npu annotation", func() {
			node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchArm, cpu: "192", mem: "755Gi",
				npuAllocateNum: "0", npuTop: ""})
			result := npu.IsMyNode(node)
			convey.So(result, convey.ShouldBeError)
		})
		convey.Convey("IsMyNode() should return error when node is of module mode", func() {
			node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchArm, cpu: "192", mem: "755Gi",
				npuAllocateNum: "1", npuTop: "Ascend910-10"})
			setNodeLabel(node.Node, acceleratorType, moduleAcceleratorType)
			result := npu.IsMyNode(node)
			convey.So(result, convey.ShouldBeError)
		})
		convey.Convey("IsMyNode() should return nil when node is of card mode", func() {
			node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchArm, cpu: "192", mem: "755Gi",
				npuAllocateNum: "1", npuTop: "Ascend910-11"})
			result := npu.IsMyNode(node)
			convey.So(result, convey.ShouldBeNil)
		})
	})
}

// TestCNPUIsMyJob
func TestCNPUIsMyJob(t *testing.T) {
	convey.Convey("Test card910x2 IsMyJob", t, func() {
		npu := &card910x2{}
		var tasks []*api.TaskInfo
		uid := api.JobID("xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxx10")

		convey.Convey("IsMyJob() should return error when job request no npu", func() {
			tasks = append(tasks, api.NewTaskInfo(buildNPUPod(
				CPodInfo{namespace: "default", groupName: "npu-group-26",
					podName: "npu-test-26", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
					reqNPUType: a300TNPUCardName, reqNpuNum: "0"})))
			job := api.NewJobInfo(uid, tasks...)
			result := npu.IsMyJob(job)
			convey.So(result, convey.ShouldBeError)
		})
		convey.Convey("IsMyJob() should return error when job has no selector", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-27",
				podName: "npu-test-27", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
				reqNPUType: a300TNPUCardName, reqNpuNum: "1"})
			setPodSelector(pod, archSelector, "")
			setPodSelector(pod, acceleratorType, "")
			tasks = append(tasks, api.NewTaskInfo(pod))
			job := api.NewJobInfo(uid, tasks...)
			result := npu.IsMyJob(job)
			convey.So(result, convey.ShouldBeError)
		})
		convey.Convey("IsMyJob() should return nil when job is of card mode", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-28",
				podName: "npu-test-28", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
				reqNPUType: a300TNPUCardName, reqNpuNum: "1"})
			tasks = append(tasks, api.NewTaskInfo(pod))
			job := api.NewJobInfo(uid, tasks...)
			result := npu.IsMyJob(job)
			convey.So(result, convey.ShouldBeNil)
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

func setJobResourceReq(CJob *api.JobInfo, resource string, num float64) {
	CJob.TotalRequest.ScalarResources[v1.ResourceName(resource)] = num
}

func buildNPUPod(podInfo CPodInfo) *v1.Pod {

	pod := util.BuildPod(podInfo.namespace, podInfo.podName, podInfo.nodeName, v1.PodPending,
		buildNPUResourceList(podInfo.reqCPUNum, podInfo.reqMem,
			v1.ResourceName(podInfo.reqNPUType), podInfo.reqNpuNum),
		podInfo.groupName, make(map[string]string, util2.NPUIndex2),
		make(map[string]string, util2.NPUIndex2))

	setPodSelector(pod, archSelector, huaweiArchX86)
	setPodSelector(pod, acceleratorType, cardAcceleratorType)

	return pod
}

func buildNPUResourceList(CCpu string, CMemory string, npuResourceType v1.ResourceName, npu string) v1.ResourceList {
	npuNum, err := strconv.Atoi(npu)
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

func buildNPUNode(CNode CNodeInfo) *api.NodeInfo {
	nodeCapacity := buildNPUResourceList(CNode.cpu, CNode.mem, a300TNPUCardName, strconv.Itoa(util2.NPUIndex2))
	nodeAlloc := buildNPUResourceList(CNode.cpu, CNode.mem, a300TNPUCardName, CNode.npuAllocateNum)
	labels := make(map[string]string, util2.NPUIndex2)
	ann := make(map[string]string, util2.NPUIndex2)

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

	node := api.NewNodeInfo(v1node)
	if CNode.npuAllocateNum != "0" {
		node.Others = map[string]interface{}{
			a300TNPUCardName: CNode.npuTop,
		}
	}
	return node
}
