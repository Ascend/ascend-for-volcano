/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package chip310x4 is using for HuaWei A300T Ascend pin affinity schedule.

*/
package chip310x4

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

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
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
	NPUID4       = 4
	NPUID61      = 61
	NPUID62      = 62
)

// TestCNPUName
func TestCNPUName(t *testing.T) {
	convey.Convey("Test chip310x4 Name", t, func() {
		npu := New(PluginName)

		convey.Convey("Name() should return PluginName defined in const", func() {
			n := npu.Name()
			convey.So(n, convey.ShouldEqual, PluginName)
		})
	})
}

// TestCNPUIsMyTask
func TestCNPUIsMyTask(t *testing.T) {
	convey.Convey("Test chip310x4 IsMyTask", t, func() {
		npu := New(PluginName)

		convey.Convey("IsMyTask() should return error when task doesn't request NPU", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-100",
				podName: "npu-test-100", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
				reqNPUType: a310NPUChipName, reqNpuNum: "0"})
			task := api.NewTaskInfo(pod)
			result := npu.IsMyTask(task)
			convey.So(result, convey.ShouldBeError)
		})
		convey.Convey("IsMyTask() should return error when task is of card mode", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-101",
				podName: "npu-test-101", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
				reqNPUType: a310NPUChipName, reqNpuNum: "1"})
			setPodLabel(pod, acceleratorType, cardAcceleratorType)
			task := api.NewTaskInfo(pod)
			result := npu.IsMyTask(task)
			convey.So(result, convey.ShouldBeError)
		})
		convey.Convey("IsMyTask() should return nil when task is of chip type", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-102",
				podName: "npu-test-102", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
				reqNPUType: a310NPUChipName, reqNpuNum: "1"})
			task := api.NewTaskInfo(pod)
			result := npu.IsMyTask(task)
			convey.So(result, convey.ShouldBeNil)
		})
	})
}

// TestCNPUIsMyNode
func TestCNPUIsMyNode(t *testing.T) {
	convey.Convey("Test chip310x4 IsMyNode", t, func() {
		npu := New(PluginName)

		convey.Convey("IsMyNode() should return error when node has no npu annotation", func() {
			node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchArm, cpu: "192", mem: "755Gi",
				npuAllocateNum: "0", npuTop: ""})
			result := npu.IsMyNode(node)
			convey.So(result, convey.ShouldBeError)
		})
		convey.Convey("IsMyNode() should return nil when node has npu annotation", func() {
			node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchArm, cpu: "192", mem: "755Gi",
				npuAllocateNum: "1", npuTop: "Ascend310-11"})
			result := npu.IsMyNode(node)
			convey.So(result, convey.ShouldBeNil)
		})
	})
}

// TestCNPUIsMyJob
func TestCNPUIsMyJob(t *testing.T) {
	convey.Convey("Test chip310x4 IsMyJob", t, func() {
		npu := New(PluginName)
		var tasks []*api.TaskInfo
		uid := api.JobID("xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxx10")

		convey.Convey("IsMyJob() should return error when job request no npu", func() {
			tasks = append(tasks, api.NewTaskInfo(buildNPUPod(
				CPodInfo{namespace: "default", groupName: "npu-group-103",
					podName: "npu-test-103", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
					reqNPUType: a310NPUChipName, reqNpuNum: "0"})))
			job := api.NewJobInfo(uid, tasks...)
			result := npu.IsMyJob(job)
			convey.So(result, convey.ShouldBeError)
		})
		convey.Convey("IsMyJob() should return nil when job is of card mode", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-105",
				podName: "npu-test-105", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
				reqNPUType: a310NPUChipName, reqNpuNum: "1"})
			tasks = append(tasks, api.NewTaskInfo(pod))
			job := api.NewJobInfo(uid, tasks...)
			result := npu.IsMyJob(job)
			convey.So(result, convey.ShouldBeNil)
		})
	})
}

// TestCNPUValidNPUJobFnInvalidSelector
func TestCNPUValidNPUJobFnInvalidSelector(t *testing.T) {
	convey.Convey("Test chip310x4 ValidNPUJobFnInvalidSelector", t, func() {
		const (
			invalidSelectorKey   = "invalid-key"
			invalidSelectorValue = "no-selector"
			validNum2            = 2000
		)
		npu := New(PluginName)
		var tasks []*api.TaskInfo
		uid := api.JobID("xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx5")

		convey.Convey("ValidNPUJobFn() should return error for job without certain selector key", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-12",
				podName: "npu-test-35", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a310NPUChipName, reqNpuNum: "1"})
			setPodLabel(pod, acceleratorType, "")
			setPodLabel(pod, invalidSelectorKey, invalidSelectorValue)
			tasks = append(tasks, api.NewTaskInfo(pod))
			job := api.NewJobInfo(uid, tasks...)
			setJobResourceReq(job, a310NPUChipName, float64(validNum2))
			result := npu.ValidNPUJobFn(job)
			convey.So(result, convey.ShouldNotBeNil)
		})
		convey.Convey("ValidNPUJobFn() should return error for job with invalid selector value", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-13",
				podName: "npu-test-36", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a310NPUChipName, reqNpuNum: "1"})
			setPodLabel(pod, acceleratorType, invalidSelectorValue)
			tasks = append(tasks, api.NewTaskInfo(pod))
			job := api.NewJobInfo(uid, tasks...)
			result := npu.ValidNPUJobFn(job)
			convey.So(result, convey.ShouldNotBeNil)
		})
	})
}

// TestCNPUValidNPUJobFnInvalidNum
func TestCNPUValidNPUJobFnInvalidNum(t *testing.T) {
	convey.Convey("Test chip310x4 ValidNPUJobFnInvalidNum", t, func() {
		const (
			invalidNum0  = "0"
			invalidNum65 = "65"
			validNum1    = "1"
			validNum64   = "64"
		)
		npu := New(PluginName)
		var tasks []*api.TaskInfo
		uid := api.JobID("xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx8")

		convey.Convey("ValidNPUJobFn() should return error for job with invalid request number 0", func() {
			tasks = append(tasks, api.NewTaskInfo(buildNPUPod(
				CPodInfo{namespace: "default", groupName: "npu-group-17",
					podName: "npu-test-40", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
					reqNPUType: a310NPUChipName, reqNpuNum: invalidNum0})))
			job := api.NewJobInfo(uid, tasks...)
			result := npu.ValidNPUJobFn(job)
			convey.So(result, convey.ShouldNotBeNil)
		})
		convey.Convey("ValidNPUJobFn() should return error for job with invalid request number 65", func() {
			tasks = append(tasks, api.NewTaskInfo(buildNPUPod(
				CPodInfo{namespace: "default", groupName: "npu-group-18",
					podName: "npu-test-41", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
					reqNPUType: a310NPUChipName, reqNpuNum: invalidNum65})))
			job := api.NewJobInfo(uid, tasks...)
			result := npu.ValidNPUJobFn(job)
			convey.So(result, convey.ShouldNotBeNil)
		})
		convey.Convey("ValidNPUJobFn() should return error for job with valid request number 1", func() {
			tasks = append(tasks, api.NewTaskInfo(buildNPUPod(
				CPodInfo{namespace: "default", groupName: "npu-group-19",
					podName: "npu-test-42", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
					reqNPUType: a310NPUChipName, reqNpuNum: validNum1})))
			job := api.NewJobInfo(uid, tasks...)
			result := npu.ValidNPUJobFn(job)
			convey.So(result, convey.ShouldBeNil)
		})
		convey.Convey("ValidNPUJobFn() should return error for job with valid request number 64", func() {
			tasks = append(tasks, api.NewTaskInfo(buildNPUPod(
				CPodInfo{namespace: "default", groupName: "npu-group-19",
					podName: "npu-test-42", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
					reqNPUType: a310NPUChipName, reqNpuNum: validNum64})))
			job := api.NewJobInfo(uid, tasks...)
			result := npu.ValidNPUJobFn(job)
			convey.So(result, convey.ShouldBeNil)
		})
	})
}

// TestCNPUValidNPUJobFnInvalidModel
func TestCNPUValidNPUJobFnInvalidModel(t *testing.T) {
	convey.Convey("Test chip310x4 ValidNPUJobFnInvalidModel", t, func() {
		const (
			num4 = "4"
			num2 = "2"
		)
		npu := New(PluginName)
		var tasks []*api.TaskInfo
		uid := api.JobID("xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx9")

		convey.Convey("ValidNPUJobFn() should return error for job with distributed", func() {
			tasks = append(tasks, api.NewTaskInfo(buildNPUPod(
				CPodInfo{namespace: "default", groupName: "npu-group-23",
					podName: "npu-test-46", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
					reqNPUType: a310NPUChipName, reqNpuNum: num2})))
			tasks = append(tasks, api.NewTaskInfo(buildNPUPod(
				CPodInfo{namespace: "default", groupName: "npu-group-123",
					podName: "npu-test-146", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
					reqNPUType: a310NPUChipName, reqNpuNum: num2})))
			job := api.NewJobInfo(uid, tasks...)
			result := npu.ValidNPUJobFn(job)
			convey.So(result, convey.ShouldBeNil)
		})
		convey.Convey("ValidNPUJobFn() should return nil for job with valid model", func() {
			tasks = append(tasks, api.NewTaskInfo(buildNPUPod(
				CPodInfo{namespace: "default", groupName: "npu-group-24",
					podName: "npu-test-47", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
					reqNPUType: a310NPUChipName, reqNpuNum: num4})))
			job := api.NewJobInfo(uid, tasks...)
			result := npu.ValidNPUJobFn(job)
			convey.So(result, convey.ShouldBeNil)
		})
	})
}

// TestCnpuPreCheckNodeFnSuccess
func TestCnpuPreCheckNodeFnSuccess(t *testing.T) {
	convey.Convey("Test chip310x4 PreCheckNodeFnSuccess", t, func() {
		npu := New(PluginName)
		var confs []conf.Configuration
		node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
			npuAllocateNum: "1", npuTop: "Ascend310-6"})

		convey.Convey("PreCheckNodeFn() should return nil", func() {
			// build a task with no selector
			task := api.NewTaskInfo(buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-14",
				podName: "npu-test-14", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a310NPUChipName, reqNpuNum: "1"}))
			result := npu.PreCheckNodeFn(task, node, confs)
			convey.So(result, convey.ShouldBeNil)
		})
	})
}

// TestCnpuCheckNPUResourceStableFn
func TestCnpuCheckNPUResourceStableFn(t *testing.T) {
	convey.Convey("Test chip310x4 CheckNPUResourceStableFn", t, func() {
		npu := New(PluginName)

		convey.Convey("CheckNPUResourceStableFn() should return error when there's missing resource type in idle", func() {
			node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "0", npuTop: ""})
			node.Node.Annotations[a310NPUChipName] = "Ascend310-2"
			result := npu.CheckNPUResourceStableFn(node)
			convey.So(result, convey.ShouldBeError)
		})
		convey.Convey("CheckNPUResourceStableFn() should return error when node resources are unstable", func() {
			node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "1", npuTop: ""})
			node.Node.Annotations[a310NPUChipName] = "Ascend310-2,Ascend310-3"
			result := npu.CheckNPUResourceStableFn(node)
			convey.So(result, convey.ShouldBeError)
		})
		convey.Convey("CheckNPUResourceStableFn() should return nil when node resources are stable", func() {
			node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "2", npuTop: "Ascend310-2,Ascend310-3"})
			result := npu.CheckNPUResourceStableFn(node)
			convey.So(result, convey.ShouldBeNil)
		})
	})
}

// TestCnpuCheckNodeNPUByTaskFn
func TestCnpuCheckNodeNPUByTaskFn(t *testing.T) {
	convey.Convey("Test job CheckNodeNPUByTaskFn", t, func() {
		npu := New(PluginName)
		pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-15",
			podName: "npu-test-15", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: a310NPUChipName, reqNpuNum: "5"})
		task := api.NewTaskInfo(pod)

		convey.Convey("CheckNodeNPUByTaskFn() should return error when node doesn't meet task requests", func() {
			node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "3", npuTop: "Ascend310-3,Ascend310-7,Ascend310-8"})
			result := npu.CheckNodeNPUByTaskFn(task, node, true)
			convey.So(result, convey.ShouldBeError)
		})
		convey.Convey("CheckNodeNPUByTaskFn() should return error when node meets task requests", func() {
			node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "2", npuTop: "Ascend310-6,Ascend310-7,Ascend310-8,Ascend310-9,Ascend310-10"})
			result := npu.CheckNodeNPUByTaskFn(task, node, true)
			convey.So(result, convey.ShouldBeNil)
		})
	})
}

// TestCnpuCheckNodeNPUByTaskFnFail
func TestCnpuCheckNodeNPUByTaskFnError(t *testing.T) {
	convey.Convey("Test job CheckNodeNPUByTaskFnError", t, func() {
		npu := New(PluginName)

		convey.Convey("CheckNodeNPUByTaskFn() should return error when the requirement of the task is 0", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-115",
				podName: "npu-test-115", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a310NPUChipName, reqNpuNum: "0"})
			task := api.NewTaskInfo(pod)
			node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "2", npuTop: "Ascend310-1,Ascend310-4"})
			result := npu.CheckNodeNPUByTaskFn(task, node, true)
			convey.So(result, convey.ShouldBeError)
		})

		convey.Convey("CheckNodeNPUByTaskFn() should return error when the number of NUP on a node is 0", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-116",
				podName: "npu-test-116", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a310NPUChipName, reqNpuNum: "2"})
			task := api.NewTaskInfo(pod)
			node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "", npuTop: ""})
			result := npu.CheckNodeNPUByTaskFn(task, node, true)
			convey.So(result, convey.ShouldBeError)
		})
	})
}

// TestCnpuGetNPUAffinityBestNodesFnReq1
func TestCnpuGetNPUAffinityBestNodesFnReq1(t *testing.T) {
	convey.Convey("Test chip310x4 GetNPUAffinityBestNodesFn When Request is 1", t, func() {
		const (
			nodeName1 = "centos1"
			nodeName2 = "centos2"
		)
		npu := New(PluginName)
		convey.Convey("GetNPUAffinityBestNodesFn() should return correct result when request NPU num is 1", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-17",
				podName: "npu-test-17", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a310NPUChipName, reqNpuNum: "1"})
			task := api.NewTaskInfo(pod)
			var nodes []*api.NodeInfo
			node1 := buildNPUNode(CNodeInfo{nodeName: nodeName1, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "3", npuTop: "Ascend310-0,Ascend310-4,Ascend310-6,Ascend310-9,Ascend310-10," +
					"Ascend310-11,Ascend310-20,Ascend310-21,Ascend310-22,Ascend310-23"})
			nodes = append(nodes, node1)
			node2 := buildNPUNode(CNodeInfo{nodeName: nodeName2, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "4", npuTop: "Ascend310-4,Ascend310-6,Ascend310-9,Ascend310-10," +
					"Ascend310-11,Ascend310-20,Ascend310-21,Ascend310-22,Ascend310-23"})
			nodes = append(nodes, node2)
			result, err := npu.GetNPUAffinityBestNodesFn(task, nodes, false)
			convey.So(result, convey.ShouldBeNil)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestCnpuScoreBestNPUNodesFn
func TestCnpuScoreBestNPUNodesFn(t *testing.T) {
	convey.Convey("Test chip310x4 ScoreBestNPUNodesFn", t, func() {
		const (
			nodeName1 = "euler1"
			nodeName2 = "euler2"
		)
		npu := New(PluginName)
		bestNodes := map[string]int{
			nodeName1: 0,
			nodeName2: constIntNum2,
		}
		pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-18",
			podName: "npu-test-18", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: a310NPUChipName, reqNpuNum: "2"})
		task := api.NewTaskInfo(pod)

		var nodes []*api.NodeInfo
		node1 := buildNPUNode(CNodeInfo{nodeName: nodeName1, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
			npuAllocateNum: "4", npuTop: "Ascend310-0,Ascend310-2"})
		nodes = append(nodes, node1)
		node2 := buildNPUNode(CNodeInfo{nodeName: nodeName2, nodeArch: huaweiArchArm, cpu: "192", mem: "755Gi",
			npuAllocateNum: "4", npuTop: "Ascend310-0,Ascend310-2,Ascend310-3,Ascend310-1"})
		nodes = append(nodes, node2)

		convey.Convey("ScoreBestNPUNodesFn() should return correct result", func() {
			scoreMap := make(map[string]float64)
			_, err := npu.ScoreBestNPUNodesFn(scoreMap, bestNodes, task, nodes)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestCnpuUpdateNPUNodeUsedCardFn1
func TestCnpuUpdateNPUNodeUsedCardFn1(t *testing.T) {
	convey.Convey("Test chip310x4 UpdateNPUNodeUsedCardFn", t, func() {
		npu := New(PluginName)
		node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
			npuAllocateNum: "7", npuTop: "Ascend310-0,Ascend310-1,Ascend310-2,Ascend310-3,Ascend310-61," +
				"Ascend310-62,Ascend310-63"})
		node.Others = map[string]interface{}{
			a310NPUChipName: "Ascend310-0,Ascend310-1,Ascend310-2,Ascend310-3,Ascend310-61,Ascend310-62,Ascend310-63",
		}

		convey.Convey("UpdateNPUNodeUsedCardFn() should successfully update node.others", func() {
			top := []int{0, constIntNum3, NPUID61, NPUID62}
			expectedResult := map[string]interface{}{
				a310NPUChipName: "Ascend310-1,Ascend310-2,Ascend310-63",
			}
			err := npu.UpdateNPUNodeUsedCardFn(node, top)
			convey.So(err, convey.ShouldBeNil)
			convey.So(node.Others, convey.ShouldResemble, expectedResult)
		})
	})
}

// TestCnpuUpdateNPUNodeUsedCardFnError2
func TestCnpuUpdateNPUNodeUsedCardFn2(t *testing.T) {
	convey.Convey("Test chip310x4 UpdateNPUNodeUsedCardFn", t, func() {
		npu := New(PluginName)

		convey.Convey("UpdateNPUNodeUsedCardFn() should should successfully update node.others", func() {
			node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "4", npuTop: "Ascend310-0,Ascend310-2,Ascend310-3,Ascend310-1"})
			node.Others = map[string]interface{}{
				a310NPUChipName: "Ascend310-0,Ascend310-2,Ascend310-3,Ascend310-1",
			}
			top := []int{0, 1, constIntNum2, constIntNum3}
			expectedResult := map[string]interface{}{
				a310NPUChipName: "",
			}
			err := npu.UpdateNPUNodeUsedCardFn(node, top)
			convey.So(err, convey.ShouldBeNil)
			convey.So(node.Others, convey.ShouldResemble, expectedResult)
		})
	})
}

// TestCnpuUpdateNPUNodeUsedCardFnError
func TestCnpuUpdateNPUNodeUsedCardFnError1(t *testing.T) {
	convey.Convey("Test chip310x4 UpdateNPUNodeUsedCardFnError", t, func() {
		npu := New(PluginName)

		convey.Convey("UpdateNPUNodeUsedCardFn() should return error when top's type mismatch", func() {
			node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "4", npuTop: "Ascend310-0,Ascend310-2,Ascend310-3,Ascend310-1"})
			top := []string{"0", "4"}
			err := npu.UpdateNPUNodeUsedCardFn(node, top)
			convey.So(err, convey.ShouldBeError)
		})

		convey.Convey("UpdateNPUNodeUsedCardFn() should return error when node's npuTop is empty", func() {
			node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "8", npuTop: ""})
			top := []int{0, NPUID4}
			err := npu.UpdateNPUNodeUsedCardFn(node, top)
			convey.So(err, convey.ShouldBeError)
		})

	})
}

// TestCnpuGetReleaseNPUTopologyFn
func TestCnpuGetReleaseNPUTopologyFn(t *testing.T) {
	convey.Convey("Test chip310x4 GetReleaseNPUTopologyFn", t, func() {
		npu := New(PluginName)
		task := api.NewTaskInfo(buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-19",
			podName: "npu-test-19", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: a310NPUChipName, reqNpuNum: "2"}))
		convey.Convey("GetReleaseNPUTopologyFn() should return correct card id slice", func() {
			task.Pod.Annotations[a310NPUChipName] = "Ascend310-0,Ascend310-4"
			expectedResult := []int{0, NPUID4}
			result, err := npu.GetReleaseNPUTopologyFn(task)
			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldResemble, expectedResult)
		})
	})
}

// TestCnpuGetReleaseNPUTopologyFnError
func TestCnpuGetReleaseNPUTopologyFnError(t *testing.T) {
	convey.Convey("Test chip310x4 GetReleaseNPUTopologyFnError", t, func() {
		npu := New(PluginName)
		task := api.NewTaskInfo(buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-122",
			podName: "npu-test-122", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: a310NPUChipName, reqNpuNum: "2"}))
		convey.Convey("GetReleaseNPUTopologyFn() should return error when the annotations of pod is wrong", func() {
			task.Pod.Annotations[a310NPUChipName] = "Ascend3100,Ascend3104"
			_, err := npu.GetReleaseNPUTopologyFn(task)
			convey.So(err, convey.ShouldBeError)
		})
	})
}

// TestCnpuUpdateReleaseNPUNodeTopologyFn
func TestCnpuUpdateReleaseNPUNodeTopologyFn(t *testing.T) {
	convey.Convey("Test chip310x4 UpdateReleaseNPUNodeTopologyFn", t, func() {
		npu := New(PluginName)
		node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
			npuAllocateNum: "4", npuTop: "Ascend310-0,Ascend310-2,Ascend310-61,Ascend310-1"})
		convey.Convey("UpdateNPUNodeUsedCardFn() should successfully update node.others", func() {
			top := []int{1, NPUID61}
			err := npu.UpdateReleaseNPUNodeTopologyFn(node, top)
			convey.So(err, convey.ShouldBeNil)
			// 0, 1, 2, 61  disorderly
			convey.So(node.Others, convey.ShouldNotBeNil)
		})

	})
}

// TestCnpuUpdateReleaseNPUNodeTopologyFnError
func TestCnpuUpdateReleaseNPUNodeTopologyFnError(t *testing.T) {
	convey.Convey("", t, func() {
		npu := New(PluginName)

		convey.Convey("UpdateNPUNodeUsedCardFn() should return error when top's type is mismatch", func() {
			node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "4", npuTop: "Ascend310-0,Ascend310-2,Ascend310-3,Ascend310-1"})
			node.Others = map[string]interface{}{
				a310NPUChipName: "Ascend310-0,Ascend310-4",
			}
			top := []string{"0", "3"}
			err := npu.UpdateReleaseNPUNodeTopologyFn(node, top)
			convey.So(err, convey.ShouldBeError)
		})

		convey.Convey("UpdateNPUNodeUsedCardFn() should return error when node's others is wrong", func() {
			node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "4", npuTop: "Ascend310-0,Ascend310-2,Ascend310-3,Ascend310-1"})
			node.Others = map[string]interface{}{
				a310NPUChipName: "Ascend310-0Ascend310-2",
			}
			top := []int{0, constIntNum3}
			err := npu.UpdateReleaseNPUNodeTopologyFn(node, top)
			convey.So(err, convey.ShouldBeError)
		})
	})
}

// TestCnpuGetAllocatedNPUFromTopologyFn
func TestCnpuGetAllocatedNPUFromTopologyFn(t *testing.T) {
	convey.Convey("Test chip310x4 GetAllocatedNPUFromTopologyFn", t, func() {
		npu := New(PluginName)
		convey.Convey("GetAllocatedNPUFromTopologyFn() should return correct result when reqNpuNum is 2", func() {
			task := api.NewTaskInfo(buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-20",
				podName: "npu-test-20", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a310NPUChipName, reqNpuNum: "2"}))
			node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchArm, cpu: "192", mem: "755Gi",
				npuAllocateNum: "2", npuTop: "Ascend310-0,Ascend310-3"})
			expectedResult := []int{0, constIntNum3}
			result, err := npu.GetAllocatedNPUFromTopologyFn(task, node, false)
			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldResemble, expectedResult)
		})
		convey.Convey("GetAllocatedNPUFromTopologyFn() should return correct result when reqNpuNum is 1", func() {
			task := api.NewTaskInfo(buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-21",
				podName: "npu-test-21", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a310NPUChipName, reqNpuNum: "1"}))
			node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchArm, cpu: "192", mem: "755Gi",
				npuAllocateNum: "2", npuTop: "Ascend310-0,Ascend310-3"})
			result, err := npu.GetAllocatedNPUFromTopologyFn(task, node, false)
			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldNotBeNil)
		})
	})
}

// TestCnpuGetAllocatedNPUFromTopologyFnError
func TestCnpuGetAllocatedNPUFromTopologyFnError(t *testing.T) {
	convey.Convey("Test chip310x4 GetAllocatedNPUFromTopologyFnError", t, func() {
		npu := New(PluginName)

		pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-123",
			podName: "npu-test-123", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: a310NPUChipName, reqNpuNum: "2"})
		task := api.NewTaskInfo(pod)
		node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchArm, cpu: "192", mem: "755Gi",
			npuAllocateNum: "2", npuTop: "Ascend310-0,Ascend310-4"})

		convey.Convey("GetAllocatedNPUFromTopologyFn() should return error when reqNpuNum of pod is 0", func() {
			task.Resreq.ScalarResources[a310NPUChipName] = 0
			_, err := npu.GetAllocatedNPUFromTopologyFn(task, node, false)
			convey.So(err, convey.ShouldBeError)
		})

		convey.Convey("GetAllocatedNPUFromTopologyFn() should return error when reqNpuNum Is greater than 4", func() {
			task.Resreq.ScalarResources[a310NPUChipName] = constInt5000
			_, err := npu.GetAllocatedNPUFromTopologyFn(task, node, false)
			convey.So(err, convey.ShouldBeError)
		})

		convey.Convey("GetAllocatedNPUFromTopologyFn() should return nil when npuTop of node is wrong", func() {
			task.Resreq.ScalarResources[a310NPUChipName] = constInt2000
			result, err := npu.GetAllocatedNPUFromTopologyFn(task, node, false)
			convey.So(err, convey.ShouldBeNil)
			convey.So(reflect.DeepEqual(result, []int{0, cardNPUNumber}), convey.ShouldBeTrue)
		})

		convey.Convey("GetAllocatedNPUFromTopologyFn() should return correct result when reqNpuNum is 0", func() {
			task.Resreq.ScalarResources[a310NPUChipName] = 0
			node.Node.Annotations[a310NPUChipName] = "Ascend310-0,Ascend310-3"
			result, err := npu.GetAllocatedNPUFromTopologyFn(task, node, false)
			convey.So(err, convey.ShouldBeError)
			convey.So(result, convey.ShouldBeNil)
		})

		convey.Convey("GetAllocatedNPUFromTopologyFn() should return error when none node meet request", func() {
			task.Resreq.ScalarResources[a310NPUChipName] = constInt2000
			node.Others[a310NPUChipName] = "Ascend310-0"
			_, err := npu.GetAllocatedNPUFromTopologyFn(task, node, false)
			convey.So(err, convey.ShouldBeError)
		})
	})
}

// TestCnpuSetNPUTopologyToPodFn
func TestCnpuSetNPUTopologyToPodFn(t *testing.T) {
	convey.Convey("Test chip310x4 SetNPUTopologyToPodFn", t, func() {
		npu := New(PluginName)
		task := api.NewTaskInfo(buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-22",
			podName: "npu-test-22", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: a310NPUChipName, reqNpuNum: "4"}))
		convey.Convey("SetNPUTopologyToPodFn() should return error when top is of wrong type", func() {
			var top []string
			err := npu.SetNPUTopologyToPodFn(task, top)
			convey.So(err, convey.ShouldBeError)
			convey.So(task.Pod.Annotations[a310NPUChipName], convey.ShouldEqual, "")
		})
		convey.Convey("SetNPUTopologyToPodFn() should write correct info in pod annotation", func() {
			top := []int{0, NPUID4}
			expectedResult := "Ascend310-0,Ascend310-4"
			err := npu.SetNPUTopologyToPodFn(task, top)
			convey.So(err, convey.ShouldBeNil)
			convey.So(task.Pod.Annotations[a310NPUChipName], convey.ShouldEqual, expectedResult)
		})
	})
}

func setPodLabel(CPod *v1.Pod, selectorKey string, selectorValue string) {
	if selectorValue == "" {
		delete(CPod.Labels, selectorKey)
		return
	}
	CPod.Labels[selectorKey] = selectorValue
}

func setJobResourceReq(CJob *api.JobInfo, resource string, num float64) {
	CJob.TotalRequest.ScalarResources[v1.ResourceName(resource)] = num
}

func buildNPUPod(podInfo CPodInfo) *v1.Pod {

	pod := util.BuildPod(podInfo.namespace, podInfo.podName, podInfo.nodeName, v1.PodPending,
		buildNPUResourceList(podInfo.reqCPUNum, podInfo.reqMem,
			v1.ResourceName(podInfo.reqNPUType), podInfo.reqNpuNum),
		podInfo.groupName, make(map[string]string, constIntNum2),
		make(map[string]string, constIntNum2))

	setPodLabel(pod, archSelector, huaweiArchX86)
	test.SetTestNPUPodSelector(pod, archSelector, huaweiArchX86)
	setPodLabel(pod, acceleratorType, chipAcceleratorType)

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
	nodeCapacity := buildNPUResourceList(CNode.cpu, CNode.mem, a310NPUChipName, strconv.Itoa(constIntNum2))
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

	test.SetNPUNodeLabel(v1node, archSelector, CNode.nodeArch)

	node := api.NewNodeInfo(v1node)
	if CNode.npuAllocateNum != "0" {
		node.Others = map[string]interface{}{
			a310NPUChipName: CNode.npuTop,
		}
	}
	return node
}
