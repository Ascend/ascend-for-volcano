/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package card310x4 is using for HuaWei A300T Ascend pin affinity schedule.
*/
package card310x4

import (
	cs "strconv"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	vapi "volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/conf"
	util2 "volcano.sh/volcano/pkg/scheduler/util"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/util"
	ascendtest "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
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
	NPUID6       = 6
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
	Convey("Test card310x4 Name", t, func() {
		npu := New(PluginName)

		Convey("Name() should return PluginName defined in const", func() {
			n := npu.Name()
			So(n, ShouldEqual, PluginName)
		})
	})
}

// TestCNPUIsMyTask
func TestCNPUIsMyTask(t *testing.T) {
	Convey("Test card310x4 IsMyTask", t, func() {
		npu := New(PluginName)
		Convey("IsMyTask() should return error when task doesn't request NPU", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-100",
				podName: "npu-test-100", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
				reqNPUType: a310NPUCardName, reqNpuNum: "0"})
			task := vapi.NewTaskInfo(pod)
			result := npu.IsMyTask(task)
			So(result, ShouldBeError)
		})
		Convey("IsMyTask() should return error when task is of chip mode", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-101",
				podName: "npu-test-101", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
				reqNPUType: a310NPUCardName, reqNpuNum: "1"})
			setPodSelector(pod, acceleratorType, chipAcceleratorType)
			task := vapi.NewTaskInfo(pod)
			result := npu.IsMyTask(task)
			So(result, ShouldBeError)
		})
		Convey("IsMyTask() should return nil when task is of card type", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-102",
				podName: "npu-test-102", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
				reqNPUType: a310NPUCardName, reqNpuNum: "1"})
			task := vapi.NewTaskInfo(pod)
			result := npu.IsMyTask(task)
			So(result, ShouldBeNil)
		})
	})
}

// TestCNPUIsMyNode
func TestCNPUIsMyNode(t *testing.T) {
	Convey("Test card310x4 IsMyNode", t, func() {
		npu := New(PluginName)

		Convey("IsMyNode() should return error when node has no npu annotation", func() {
			node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchArm, cpu: "192", mem: "755Gi",
				npuAllocateNum: "0", npuTop: ""})
			result := npu.IsMyNode(node)
			So(result, ShouldBeError)
		})
		Convey("IsMyNode() should return nil when node has npu annotation", func() {
			node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchArm, cpu: "192", mem: "755Gi",
				npuAllocateNum: "1", npuTop: "Ascend310-11"})
			result := npu.IsMyNode(node)
			So(result, ShouldBeNil)
		})
	})
}

// TestCNPUIsMyJob
func TestCNPUIsMyJob(t *testing.T) {
	Convey("Test card310x4 IsMyJob", t, func() {
		npu := New(PluginName)
		var tasks []*vapi.TaskInfo
		uid := vapi.JobID("xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxx10")

		Convey("IsMyJob() should return error when job request no npu", func() {
			tasks = append(tasks, vapi.NewTaskInfo(buildNPUPod(
				CPodInfo{namespace: "default", groupName: "npu-group-103",
					podName: "npu-test-103", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
					reqNPUType: a310NPUCardName, reqNpuNum: "0"})))
			job := vapi.NewJobInfo(uid, tasks...)
			result := npu.IsMyJob(job)
			So(result, ShouldBeError)
		})
		Convey("IsMyJob() should return error when job has no selector", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-104",
				podName: "npu-test-104", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
				reqNPUType: a310NPUCardName, reqNpuNum: "1"})
			setPodSelector(pod, archSelector, "")
			setPodSelector(pod, acceleratorType, "")
			tasks = append(tasks, vapi.NewTaskInfo(pod))
			job := vapi.NewJobInfo(uid, tasks...)
			result := npu.IsMyJob(job)
			So(result, ShouldBeError)
		})
		Convey("IsMyJob() should return nil when job is of card mode", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-105",
				podName: "npu-test-105", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
				reqNPUType: a310NPUCardName, reqNpuNum: "1"})
			tasks = append(tasks, vapi.NewTaskInfo(pod))
			job := vapi.NewJobInfo(uid, tasks...)
			result := npu.IsMyJob(job)
			So(result, ShouldBeNil)
		})
	})
}

// TestCNPUValidNPUJobFnInvalidSelector
func TestCNPUValidNPUJobFnInvalidSelector(t *testing.T) {
	Convey("Test card310x4 ValidNPUJobFnInvalidSelector", t, func() {
		const (
			invalidSelectorKey   = "invalid-key"
			invalidSelectorValue = "no-selector"
			validNum2            = 2000
		)
		npu := New(PluginName)
		var tasks []*vapi.TaskInfo
		uid := vapi.JobID("xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx5")

		Convey("ValidNPUJobFn() should return error for job without certain selector key", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-12",
				podName: "npu-test-35", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a310NPUCardName, reqNpuNum: "1"})
			setPodSelector(pod, acceleratorType, "")
			setPodSelector(pod, invalidSelectorKey, invalidSelectorValue)
			tasks = append(tasks, vapi.NewTaskInfo(pod))
			job := vapi.NewJobInfo(uid, tasks...)
			setJobResourceReq(job, a310NPUCardName, float64(validNum2))
			result := npu.ValidNPUJobFn(job)
			So(result, ShouldNotBeNil)
		})
		Convey("ValidNPUJobFn() should return error for job with invalid selector value", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-13",
				podName: "npu-test-36", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a310NPUCardName, reqNpuNum: "1"})
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
	Convey("Test card310x4 ValidNPUJobFnInvalidNum", t, func() {
		const (
			invalidNum0 = "0"
			invalidNum5 = "5"
			validNum1   = "1"
			validNum4   = "4"
		)
		npu := New(PluginName)
		var tasks []*vapi.TaskInfo
		uid := vapi.JobID("xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx8")

		Convey("ValidNPUJobFn() should return error for job with invalid request number 0", func() {
			tasks = append(tasks, vapi.NewTaskInfo(buildNPUPod(
				CPodInfo{namespace: "default", groupName: "npu-group-17",
					podName: "npu-test-40", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
					reqNPUType: a310NPUCardName, reqNpuNum: invalidNum0})))
			job := vapi.NewJobInfo(uid, tasks...)
			result := npu.ValidNPUJobFn(job)
			So(result, ShouldNotBeNil)
		})
		Convey("ValidNPUJobFn() should return error for job with invalid request number 5", func() {
			tasks = append(tasks, vapi.NewTaskInfo(buildNPUPod(
				CPodInfo{namespace: "default", groupName: "npu-group-18",
					podName: "npu-test-41", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
					reqNPUType: a310NPUCardName, reqNpuNum: invalidNum5})))
			job := vapi.NewJobInfo(uid, tasks...)
			result := npu.ValidNPUJobFn(job)
			So(result, ShouldNotBeNil)
		})
		Convey("ValidNPUJobFn() should return error for job with valid request number 1", func() {
			tasks = append(tasks, vapi.NewTaskInfo(buildNPUPod(
				CPodInfo{namespace: "default", groupName: "npu-group-19",
					podName: "npu-test-42", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
					reqNPUType: a310NPUCardName, reqNpuNum: validNum1})))
			job := vapi.NewJobInfo(uid, tasks...)
			result := npu.ValidNPUJobFn(job)
			So(result, ShouldBeNil)
		})
		Convey("ValidNPUJobFn() should return error for job with valid request number 4", func() {
			tasks = append(tasks, vapi.NewTaskInfo(buildNPUPod(
				CPodInfo{namespace: "default", groupName: "npu-group-19",
					podName: "npu-test-42", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
					reqNPUType: a310NPUCardName, reqNpuNum: validNum4})))
			job := vapi.NewJobInfo(uid, tasks...)
			result := npu.ValidNPUJobFn(job)
			So(result, ShouldBeNil)
		})
	})
}

// TestCNPUValidNPUJobFnInvalidModel
func TestCNPUValidNPUJobFnInvalidModel(t *testing.T) {
	Convey("Test card310x4 ValidNPUJobFnInvalidModel", t, func() {
		const (
			num4 = "4"
			num5 = "5"
			num1 = "1"
		)
		npu := New(PluginName)
		var tasks []*vapi.TaskInfo
		uid := vapi.JobID("xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx9")

		Convey("ValidNPUJobFn() should return error for job with distributed", func() {
			tasks = append(tasks, vapi.NewTaskInfo(buildNPUPod(
				CPodInfo{namespace: "default", groupName: "npu-group-23",
					podName: "npu-test-46", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
					reqNPUType: a310NPUCardName, reqNpuNum: num1})))
			tasks = append(tasks, vapi.NewTaskInfo(buildNPUPod(
				CPodInfo{namespace: "default", groupName: "npu-group-123",
					podName: "npu-test-146", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
					reqNPUType: a310NPUCardName, reqNpuNum: num1})))
			job := vapi.NewJobInfo(uid, tasks...)
			result := npu.ValidNPUJobFn(job)
			So(result, ShouldBeNil)
		})
		Convey("ValidNPUJobFn() should return nil", func() {
			tasks = append(tasks, vapi.NewTaskInfo(buildNPUPod(
				CPodInfo{namespace: "default", groupName: "npu-group-23",
					podName: "npu-test-46", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
					reqNPUType: a310NPUCardName, reqNpuNum: num4})))
			tasks = append(tasks, vapi.NewTaskInfo(buildNPUPod(
				CPodInfo{namespace: "default", groupName: "npu-group-123",
					podName: "npu-test-146", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
					reqNPUType: a310NPUCardName, reqNpuNum: num4})))
			job := vapi.NewJobInfo(uid, tasks...)
			result := npu.ValidNPUJobFn(job)
			So(result, ShouldBeNil)
		})
		Convey("ValidNPUJobFn() should return nil for job with valid model", func() {
			tasks = append(tasks, vapi.NewTaskInfo(buildNPUPod(
				CPodInfo{namespace: "default", groupName: "npu-group-24",
					podName: "npu-test-47", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
					reqNPUType: a310NPUCardName, reqNpuNum: num5})))
			job := vapi.NewJobInfo(uid, tasks...)
			result := npu.ValidNPUJobFn(job)
			So(result, ShouldNotBeNil)
		})
	})
}

// TestCNPUPreCheckNodeFnTaskError
func TestCNPUPreCheckNodeFnTaskError(t *testing.T) {
	Convey("Test card310x4 PreCheckNodeFnTaskError", t, func() {
		npu := New(PluginName)
		var confs []conf.Configuration
		node := buildNPUNode(CNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
			"1", "Ascend310-1"})
		Convey("PreCheckNodeFn() should return nil when task isn't npu task", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-10",
				podName: "npu-test-10", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: "", reqNpuNum: ""})
			task := vapi.NewTaskInfo(pod)
			result := npu.PreCheckNodeFn(task, node, confs)
			So(result, ShouldBeNil)
		})
	})
}

// TestCnpuPreCheckNodeFnSuccess
func TestCnpuPreCheckNodeFnSuccess(t *testing.T) {
	Convey("Test card310x4 PreCheckNodeFnSuccess", t, func() {
		npu := New(PluginName)
		var confs []conf.Configuration
		node := buildNPUNode(CNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
			"1", "Ascend310-6"})

		Convey("PreCheckNodeFn() should return nil when selectors match with labels", func() {
			// build a task with no selector
			task := vapi.NewTaskInfo(buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-14",
				podName: "npu-test-14", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a310NPUCardName, reqNpuNum: "1"}))
			result := npu.PreCheckNodeFn(task, node, confs)
			So(result, ShouldBeNil)
		})
	})
}

// TestCnpuCheckNPUResourceStableFn
func TestCnpuCheckNPUResourceStableFn(t *testing.T) {
	Convey("Test card310x4 CheckNPUResourceStableFn", t, func() {
		npu := New(PluginName)

		Convey("CheckNPUResourceStableFn() should return error when there's missing resource type in idle", func() {
			node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "0", npuTop: ""})
			node.Node.Annotations[a310NPUCardName] = "Ascend310-2"
			result := npu.CheckNPUResourceStableFn(node)
			So(result, ShouldBeError)
		})
		Convey("CheckNPUResourceStableFn() should return error when node resources are unstable", func() {
			node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "1", npuTop: ""})
			node.Node.Annotations[a310NPUCardName] = "Ascend310-2,Ascend310-3"
			result := npu.CheckNPUResourceStableFn(node)
			So(result, ShouldBeError)
		})
		Convey("CheckNPUResourceStableFn() should return nil when node resources are stable", func() {
			node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "2", npuTop: "Ascend310-2,Ascend310-3"})
			result := npu.CheckNPUResourceStableFn(node)
			So(result, ShouldBeNil)
		})
	})
}

// TestCnpuCheckNodeNPUByTaskFn
func TestCnpuCheckNodeNPUByTaskFn(t *testing.T) {
	Convey("Test job CheckNodeNPUByTaskFn", t, func() {
		npu := New(PluginName)
		pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-15",
			podName: "npu-test-15", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: a310NPUCardName, reqNpuNum: "2"})
		task := vapi.NewTaskInfo(pod)

		Convey("CheckNodeNPUByTaskFn() should return error when node doesn't meet task requests", func() {
			node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "3", npuTop: "Ascend310-3,Ascend310-7,Ascend310-8"})
			result := npu.CheckNodeNPUByTaskFn(task, node, true)
			So(result, ShouldBeError)
		})
		Convey("CheckNodeNPUByTaskFn() should return error when node meets task requests", func() {
			node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "2", npuTop: "Ascend310-8,Ascend310-9"})
			result := npu.CheckNodeNPUByTaskFn(task, node, true)
			So(result, ShouldBeNil)
		})
	})
}

// TestCnpuCheckNodeNPUByTaskFnFail
func TestCnpuCheckNodeNPUByTaskFnError(t *testing.T) {
	Convey("Test job CheckNodeNPUByTaskFnError", t, func() {
		npu := New(PluginName)

		Convey("CheckNodeNPUByTaskFn() should return error when the requirement of the task is 0", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-115",
				podName: "npu-test-115", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a310NPUCardName, reqNpuNum: "0"})
			task := vapi.NewTaskInfo(pod)
			node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "2", npuTop: "Ascend310-1,Ascend310-4"})
			result := npu.CheckNodeNPUByTaskFn(task, node, true)
			So(result, ShouldBeError)
		})

		Convey("CheckNodeNPUByTaskFn() should return error when the number of NUP on a node is 0", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-116",
				podName: "npu-test-116", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a310NPUCardName, reqNpuNum: "2"})
			task := vapi.NewTaskInfo(pod)
			node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "", npuTop: ""})
			result := npu.CheckNodeNPUByTaskFn(task, node, true)
			So(result, ShouldBeError)
		})
	})
}

// TestCnpuGetNPUAffinityBestNodesFnReq1
func TestCnpuGetNPUAffinityBestNodesFnReq1(t *testing.T) {
	Convey("Test card310x4 GetNPUAffinityBestNodesFn When Request is 1", t, func() {
		const (
			nodeName1 = "centos1"
			nodeName2 = "centos2"
			nodeName3 = "centos3"
			nodeName4 = "centos4"
			constNum0 = 0
			constNum1 = 1
			constNum3 = 3
			constNum4 = 4
		)
		npu := New(PluginName)
		Convey("GetNPUAffinityBestNodesFn() should return correct result when request NPU num is 1", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-17",
				podName: "npu-test-17", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a310NPUCardName, reqNpuNum: "1"})
			task := vapi.NewTaskInfo(pod)
			var nodes []*vapi.NodeInfo
			node1 := buildNPUNode(CNodeInfo{nodeName: nodeName1, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "10", npuTop: "Ascend310-0,Ascend310-4,Ascend310-6,Ascend310-9,Ascend310-10," +
					"Ascend310-11,Ascend310-20,Ascend310-21,Ascend310-22,Ascend310-23"})
			nodes = append(nodes, node1)
			node2 := buildNPUNode(CNodeInfo{nodeName: nodeName2, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "9", npuTop: "Ascend310-4,Ascend310-6,Ascend310-9,Ascend310-10," +
					"Ascend310-11,Ascend310-20,Ascend310-21,Ascend310-22,Ascend310-23"})
			nodes = append(nodes, node2)
			node3 := buildNPUNode(CNodeInfo{nodeName: nodeName3, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "7", npuTop: "Ascend310-9,Ascend310-10,Ascend310-11,Ascend310-20," +
					"Ascend310-21,Ascend310-22,Ascend310-23"})
			nodes = append(nodes, node3)
			node4 := buildNPUNode(CNodeInfo{nodeName: nodeName4, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "4", npuTop: "Ascend310-20,Ascend310-21,Ascend310-22,Ascend310-23"})
			nodes = append(nodes, node4)
			result, err := npu.GetNPUAffinityBestNodesFn(task, nodes, false)
			So(result[nodeName1], ShouldEqual, constNum0)
			So(result[nodeName2], ShouldEqual, constNum1)
			So(result[nodeName3], ShouldEqual, constNum1)
			So(result[nodeName4], ShouldEqual, constNum3)
			So(len(result), ShouldEqual, constNum4)
			So(err, ShouldBeNil)
		})
	})
}

// TestCnpuGetNPUAffinityBestNodesFnReq2
func TestCnpuGetNPUAffinityBestNodesFnReq2(t *testing.T) {
	Convey("Test card310x4 GetNPUAffinityBestNodesFn When Request is 2", t, func() {
		const (
			nodeName1 = "centos1"
			nodeName2 = "centos2"
			nodeName3 = "centos3"
			nodeName4 = "centos4"
			constNum0 = 0
			constNum1 = 1
			constNum2 = 2
		)
		npu := New(PluginName)

		Convey("GetNPUAffinityBestNodesFn() should return correct result when request NPU num is 2", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-17",
				podName: "npu-test-17", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a310NPUCardName, reqNpuNum: "2"})
			task := vapi.NewTaskInfo(pod)
			var nodes []*vapi.NodeInfo
			node1 := buildNPUNode(CNodeInfo{nodeName: nodeName1, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "10", npuTop: "Ascend310-0,Ascend310-4,Ascend310-6,Ascend310-9,Ascend310-10," +
					"Ascend310-11,Ascend310-20,Ascend310-21,Ascend310-22,Ascend310-23"})
			nodes = append(nodes, node1)
			node2 := buildNPUNode(CNodeInfo{nodeName: nodeName2, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "9", npuTop: "Ascend310-4,Ascend310-6,Ascend310-9,Ascend310-10," +
					"Ascend310-11,Ascend310-20,Ascend310-21,Ascend310-22,Ascend310-23"})
			nodes = append(nodes, node2)
			node3 := buildNPUNode(CNodeInfo{nodeName: nodeName3, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "7", npuTop: "Ascend310-9,Ascend310-10,Ascend310-11,Ascend310-20," +
					"Ascend310-21,Ascend310-22,Ascend310-23"})
			nodes = append(nodes, node3)
			node4 := buildNPUNode(CNodeInfo{nodeName: nodeName4, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "4", npuTop: "Ascend310-20,Ascend310-21,Ascend310-22,Ascend310-23"})
			nodes = append(nodes, node4)
			result, err := npu.GetNPUAffinityBestNodesFn(task, nodes, false)
			So(result[nodeName1], ShouldEqual, constNum0)
			So(result[nodeName2], ShouldEqual, constNum0)
			So(result[nodeName3], ShouldEqual, constNum1)
			So(result[nodeName4], ShouldEqual, constNum2)
			So(err, ShouldBeNil)
		})
	})
}

// TestCnpuGetNPUAffinityBestNodesFnReq3
func TestCnpuGetNPUAffinityBestNodesFnReq3(t *testing.T) {
	Convey("Test card310x4 GetNPUAffinityBestNodesFn When Request is 3", t, func() {
		const (
			nodeName1 = "centos1"
			nodeName2 = "centos2"
			nodeName3 = "centos3"
			nodeName4 = "centos4"
			constNum0 = 0
			constNum1 = 1
		)
		npu := New(PluginName)

		Convey("GetNPUAffinityBestNodesFn() should return correct result when request NPU num is 3", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-17",
				podName: "npu-test-17", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a310NPUCardName, reqNpuNum: "3"})
			task := vapi.NewTaskInfo(pod)
			var nodes []*vapi.NodeInfo
			node1 := buildNPUNode(CNodeInfo{nodeName: nodeName1, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "10", npuTop: "Ascend310-0,Ascend310-4,Ascend310-6,Ascend310-9,Ascend310-10," +
					"Ascend310-11,Ascend310-20,Ascend310-21,Ascend310-22,Ascend310-23"})
			nodes = append(nodes, node1)
			node2 := buildNPUNode(CNodeInfo{nodeName: nodeName2, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "9", npuTop: "Ascend310-4,Ascend310-6,Ascend310-9,Ascend310-10," +
					"Ascend310-11,Ascend310-20,Ascend310-21,Ascend310-22,Ascend310-23"})
			nodes = append(nodes, node2)
			node3 := buildNPUNode(CNodeInfo{nodeName: nodeName3, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "7", npuTop: "Ascend310-9,Ascend310-10,Ascend310-11,Ascend310-20," +
					"Ascend310-21,Ascend310-22,Ascend310-23"})
			nodes = append(nodes, node3)
			node4 := buildNPUNode(CNodeInfo{nodeName: nodeName4, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "4", npuTop: "Ascend310-20,Ascend310-21,Ascend310-22,Ascend310-23"})
			nodes = append(nodes, node4)
			result, err := npu.GetNPUAffinityBestNodesFn(task, nodes, false)
			So(result[nodeName1], ShouldEqual, constNum0)
			So(result[nodeName2], ShouldEqual, constNum0)
			So(result[nodeName3], ShouldEqual, constNum0)
			So(result[nodeName4], ShouldEqual, constNum1)
			So(err, ShouldBeNil)
		})
	})
}

// TestCnpuGetNPUAffinityBestNodesFnReq4
func TestCnpuGetNPUAffinityBestNodesFnReq4(t *testing.T) {
	Convey("Test card310x4 GetNPUAffinityBestNodesFn When Request is 4", t, func() {
		const (
			nodeName1 = "centos1"
			nodeName2 = "centos2"
			nodeName3 = "centos3"
			nodeName4 = "centos4"
			constNum0 = 0
		)
		npu := New(PluginName)

		Convey("GetNPUAffinityBestNodesFn() should return correct result when request NPU num is 4", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-17",
				podName: "npu-test-17", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a310NPUCardName, reqNpuNum: "4"})
			task := vapi.NewTaskInfo(pod)
			var nodes []*vapi.NodeInfo
			node1 := buildNPUNode(CNodeInfo{nodeName: nodeName1, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "10", npuTop: "Ascend310-0,Ascend310-4,Ascend310-6,Ascend310-9,Ascend310-10," +
					"Ascend310-11,Ascend310-20,Ascend310-21,Ascend310-22,Ascend310-23"})
			nodes = append(nodes, node1)
			node2 := buildNPUNode(CNodeInfo{nodeName: nodeName2, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "9", npuTop: "Ascend310-4,Ascend310-6,Ascend310-9,Ascend310-10," +
					"Ascend310-11,Ascend310-20,Ascend310-21,Ascend310-22,Ascend310-23"})
			nodes = append(nodes, node2)
			node3 := buildNPUNode(CNodeInfo{nodeName: nodeName3, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "7", npuTop: "Ascend310-9,Ascend310-10,Ascend310-11,Ascend310-20," +
					"Ascend310-21,Ascend310-22,Ascend310-23"})
			nodes = append(nodes, node3)
			node4 := buildNPUNode(CNodeInfo{nodeName: nodeName4, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "4", npuTop: "Ascend310-20,Ascend310-21,Ascend310-22,Ascend310-23"})
			nodes = append(nodes, node4)
			result, err := npu.GetNPUAffinityBestNodesFn(task, nodes, false)
			So(result[nodeName1], ShouldEqual, constNum0)
			So(result[nodeName2], ShouldEqual, constNum0)
			So(result[nodeName3], ShouldEqual, constNum0)
			So(result[nodeName4], ShouldEqual, constNum0)
			So(err, ShouldBeNil)
		})
	})
}

// TestCnpuGetNPUAffinityBestNodesFn
func TestCnpuGetNPUAffinityBestNodesFnError(t *testing.T) {
	Convey("Test card310x4 GetNPUAffinityBestNodesFnError", t, func() {
		const (
			nodeName1 = "centos1"
			nodeName2 = "centos2"
		)
		npu := New(PluginName)

		pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-117",
			podName: "npu-test-117", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: a310NPUCardName, reqNpuNum: "2"})
		task := vapi.NewTaskInfo(pod)
		var nodes []*vapi.NodeInfo

		Convey("GetNPUAffinityBestNodesFn() should return error when node is empty", func() {
			_, err := npu.GetNPUAffinityBestNodesFn(task, nodes, false)
			So(err, ShouldBeError)
		})

		node := buildNPUNode(CNodeInfo{nodeName: nodeName1, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
			npuAllocateNum: "1", npuTop: "Ascend310-0"})
		nodes = append(nodes, node)

		Convey("GetNPUAffinityBestNodesFn() should return error when none bestNodes", func() {
			result, err := npu.GetNPUAffinityBestNodesFn(task, nodes, false)
			So(result[nodeName2], ShouldEqual, 0)
			So(err, ShouldBeError)
		})

		Convey("GetNPUAffinityBestNodesFn() should return error when request NPU number is 0", func() {
			task.Resreq.ScalarResources[a310NPUCardName] = 0
			_, err := npu.GetNPUAffinityBestNodesFn(task, nodes, false)
			So(err, ShouldBeError)
		})

		Convey("GetNPUAffinityBestNodesFn() should return error when request number is greater than 4", func() {
			task.Resreq.ScalarResources[a310NPUCardName] = constInt5000
			_, err := npu.GetNPUAffinityBestNodesFn(task, nodes, false)
			So(err, ShouldBeError)
		})
	})
}

// TestCnpuScoreBestNPUNodesFn
func TestCnpuScoreBestNPUNodesFn(t *testing.T) {
	Convey("Test card310x4 ScoreBestNPUNodesFn", t, func() {
		const (
			nodeName1 = "euler1"
			nodeName2 = "euler2"
		)
		npu := New(PluginName)
		bestNodes := map[string]int{
			nodeName1: 0,
			nodeName2: util.ConstIntNum2,
		}
		pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-18",
			podName: "npu-test-18", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: a310NPUCardName, reqNpuNum: "2"})
		task := vapi.NewTaskInfo(pod)

		var nodes []*vapi.NodeInfo
		node1 := buildNPUNode(CNodeInfo{nodeName: nodeName1, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
			npuAllocateNum: "4", npuTop: "Ascend310-0,Ascend310-2"})
		nodes = append(nodes, node1)
		node2 := buildNPUNode(CNodeInfo{nodeName: nodeName2, nodeArch: huaweiArchArm, cpu: "192", mem: "755Gi",
			npuAllocateNum: "4", npuTop: "Ascend310-0,Ascend310-2,Ascend310-3,Ascend310-1"})
		nodes = append(nodes, node2)

		Convey("ScoreBestNPUNodesFn() should return correct result", func() {
			scoreMap := make(map[string]float64)
			expectedResult := map[string]float64(nil)
			result, err := npu.ScoreBestNPUNodesFn(scoreMap, bestNodes, task, nodes)
			So(err, ShouldNotBeNil)
			So(result, ShouldResemble, expectedResult)
		})
	})
}

// TestCnpuScoreBestNPUNodesFnError
func TestCnpuScoreBestNPUNodesFnError(t *testing.T) {
	Convey("Test card310x4 ScoreBestNPUNodesFnError", t, func() {
		const (
			nodeName1 = "euler1"
			nodeName2 = "euler2"
		)
		npu := New(PluginName)
		bestNodes := map[string]int{
			nodeName1: 0,
			nodeName2: util.ConstIntNum3,
		}
		pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-121",
			podName: "npu-test-121", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: a310NPUCardName, reqNpuNum: "4"})
		task := vapi.NewTaskInfo(pod)

		var nodes []*vapi.NodeInfo
		node1 := buildNPUNode(CNodeInfo{nodeName: nodeName1, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
			npuAllocateNum: "4", npuTop: "Ascend310-0,Ascend310-2,Ascend310-3,Ascend310-1"})
		nodes = append(nodes, node1)
		node2 := buildNPUNode(CNodeInfo{nodeName: nodeName2, nodeArch: huaweiArchArm, cpu: "192", mem: "755Gi",
			npuAllocateNum: "4", npuTop: "Ascend310-0,Ascend310-2,Ascend310-3,Ascend310-1"})
		nodes = append(nodes, node2)

		Convey("ScoreBestNPUNodesFn() should return err when scoreMap is nil", func() {
			var scoreMap map[string]float64
			_, err := npu.ScoreBestNPUNodesFn(scoreMap, bestNodes, task, nodes)
			So(err, ShouldBeError)
		})
	})
}

// TestCnpuUpdateNPUNodeUsedCardFn1
func TestCnpuUpdateNPUNodeUsedCardFn1(t *testing.T) {
	Convey("Test card310x4 UpdateNPUNodeUsedCardFn", t, func() {
		npu := New(PluginName)
		node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
			npuAllocateNum: "7", npuTop: "Ascend310-0,Ascend310-1,Ascend310-2,Ascend310-3,Ascend310-61," +
				"Ascend310-62,Ascend310-63"})
		node.Others = map[string]interface{}{
			a310NPUCardName: "Ascend310-0,Ascend310-1,Ascend310-2,Ascend310-3,Ascend310-61,Ascend310-62,Ascend310-63",
		}

		Convey("UpdateNPUNodeUsedCardFn() should successfully update node.others", func() {
			top := []int{0, 3, 61, 62}
			expectedResult := map[string]interface{}{
				a310NPUCardName: "Ascend310-1,Ascend310-2,Ascend310-63",
			}
			err := npu.UpdateNPUNodeUsedCardFn(node, top)
			So(err, ShouldBeNil)
			So(node.Others, ShouldResemble, expectedResult)
		})
	})
}

// TestCnpuUpdateNPUNodeUsedCardFnError2
func TestCnpuUpdateNPUNodeUsedCardFn2(t *testing.T) {
	Convey("Test card310x4 UpdateNPUNodeUsedCardFn", t, func() {
		npu := New(PluginName)

		Convey("UpdateNPUNodeUsedCardFn() should should successfully update node.others", func() {
			node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "4", npuTop: "Ascend310-0,Ascend310-2,Ascend310-3,Ascend310-1"})
			node.Others = map[string]interface{}{
				a310NPUCardName: "Ascend310-0,Ascend310-2,Ascend310-3,Ascend310-1",
			}
			top := []int{0, 1, 2, 3}
			expectedResult := map[string]interface{}{
				a310NPUCardName: "",
			}
			err := npu.UpdateNPUNodeUsedCardFn(node, top)
			So(err, ShouldBeNil)
			So(node.Others, ShouldResemble, expectedResult)
		})
	})
}

// TestCnpuUpdateNPUNodeUsedCardFnError
func TestCnpuUpdateNPUNodeUsedCardFnError1(t *testing.T) {
	Convey("Test card310x4 UpdateNPUNodeUsedCardFnError", t, func() {
		npu := New(PluginName)

		Convey("UpdateNPUNodeUsedCardFn() should return error when top's type mismatch", func() {
			node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "4", npuTop: "Ascend310-0,Ascend310-2,Ascend310-3,Ascend310-1"})
			top := []string{"0", "4"}
			err := npu.UpdateNPUNodeUsedCardFn(node, top)
			So(err, ShouldBeError)
		})

		Convey("UpdateNPUNodeUsedCardFn() should return error when node's npuTop is empty", func() {
			node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "8", npuTop: ""})
			top := []int{0, 4}
			err := npu.UpdateNPUNodeUsedCardFn(node, top)
			So(err, ShouldBeError)
		})

	})
}

// TestCnpuGetReleaseNPUTopologyFn
func TestCnpuGetReleaseNPUTopologyFn(t *testing.T) {
	Convey("Test card310x4 GetReleaseNPUTopologyFn", t, func() {
		npu := New(PluginName)
		task := vapi.NewTaskInfo(buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-19",
			podName: "npu-test-19", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: a310NPUCardName, reqNpuNum: "2"}))
		Convey("GetReleaseNPUTopologyFn() should return correct card id slice", func() {
			task.Pod.Annotations[a310NPUCardName] = "Ascend310-0,Ascend310-4"
			expectedResult := []int{0, 4}
			result, err := npu.GetReleaseNPUTopologyFn(task)
			So(err, ShouldBeNil)
			So(result, ShouldResemble, expectedResult)
		})
	})
}

// TestCnpuGetReleaseNPUTopologyFnError
func TestCnpuGetReleaseNPUTopologyFnError(t *testing.T) {
	Convey("Test card310x4 GetReleaseNPUTopologyFnError", t, func() {
		npu := New(PluginName)
		task := vapi.NewTaskInfo(buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-122",
			podName: "npu-test-122", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: a310NPUCardName, reqNpuNum: "2"}))
		Convey("GetReleaseNPUTopologyFn() should return error when the annotations of pod is wrong", func() {
			task.Pod.Annotations[a310NPUCardName] = "Ascend3100,Ascend3104"
			_, err := npu.GetReleaseNPUTopologyFn(task)
			So(err, ShouldBeError)
		})
	})
}

// TestCnpuUpdateReleaseNPUNodeTopologyFn
func TestCnpuUpdateReleaseNPUNodeTopologyFn(t *testing.T) {
	Convey("Test card310x4 UpdateReleaseNPUNodeTopologyFn", t, func() {
		npu := New(PluginName)
		node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
			npuAllocateNum: "4", npuTop: "Ascend310-0,Ascend310-2,Ascend310-3,Ascend310-1"})
		node.Others = map[string]interface{}{
			a310NPUCardName: "Ascend310-0,Ascend310-2",
		}
		Convey("UpdateNPUNodeUsedCardFn() should successfully update node.others", func() {
			top := []int{1, 3}
			err := npu.UpdateReleaseNPUNodeTopologyFn(node, top)
			So(err, ShouldBeNil)
			// 0, 1, 2, 3  disorderly
			So(node.Others, ShouldNotBeNil)
		})

	})
}

// TestCnpuUpdateReleaseNPUNodeTopologyFnError
func TestCnpuUpdateReleaseNPUNodeTopologyFnError(t *testing.T) {
	Convey("", t, func() {
		npu := New(PluginName)

		Convey("UpdateNPUNodeUsedCardFn() should return error when top's type is mismatch", func() {
			node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "4", npuTop: "Ascend310-0,Ascend310-2,Ascend310-3,Ascend310-1"})
			node.Others = map[string]interface{}{
				a310NPUCardName: "Ascend310-0,Ascend310-4",
			}
			top := []string{"0", "3"}
			err := npu.UpdateReleaseNPUNodeTopologyFn(node, top)
			So(err, ShouldBeError)
		})

		Convey("UpdateNPUNodeUsedCardFn() should return error when node's others is wrong", func() {
			node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "4", npuTop: "Ascend310-0,Ascend310-2,Ascend310-3,Ascend310-1"})
			node.Others = map[string]interface{}{
				a310NPUCardName: "Ascend310-0Ascend310-2",
			}
			top := []int{0, 3}
			err := npu.UpdateReleaseNPUNodeTopologyFn(node, top)
			So(err, ShouldBeError)
		})
	})
}

// TestCnpuGetAllocatedNPUFromTopologyFn
func TestCnpuGetAllocatedNPUFromTopologyFn(t *testing.T) {
	Convey("Test card310x4 GetAllocatedNPUFromTopologyFn", t, func() {
		npu := New(PluginName)

		Convey("GetAllocatedNPUFromTopologyFn() should return correct result when reqNpuNum is 2", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-20",
				podName: "npu-test-20", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a310NPUCardName, reqNpuNum: "2"})
			task := vapi.NewTaskInfo(pod)
			node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchArm, cpu: "192", mem: "755Gi",
				npuAllocateNum: "2", npuTop: "Ascend310-0,Ascend310-3"})
			expectedResult := []int{0, 3}
			result, err := npu.GetAllocatedNPUFromTopologyFn(task, node, false)
			So(err, ShouldBeNil)
			So(result, ShouldResemble, expectedResult)
		})

		Convey("GetAllocatedNPUFromTopologyFn() should return correct result when reqNpuNum is 1", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-21",
				podName: "npu-test-21", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a310NPUCardName, reqNpuNum: "1"})
			task := vapi.NewTaskInfo(pod)
			node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchArm, cpu: "192", mem: "755Gi",
				npuAllocateNum: "2", npuTop: "Ascend310-0,Ascend310-3"})
			expectedResult := []int{0}
			result, err := npu.GetAllocatedNPUFromTopologyFn(task, node, false)
			So(err, ShouldBeNil)
			So(result, ShouldResemble, expectedResult)
		})
	})
}

// TestCnpuGetAllocatedNPUFromTopologyFnError
func TestCnpuGetAllocatedNPUFromTopologyFnError(t *testing.T) {
	Convey("Test card310x4 GetAllocatedNPUFromTopologyFnError", t, func() {
		npu := New(PluginName)

		pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-123",
			podName: "npu-test-123", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: a310NPUCardName, reqNpuNum: "2"})
		task := vapi.NewTaskInfo(pod)
		node := buildNPUNode(CNodeInfo{nodeName: nodeName, nodeArch: huaweiArchArm, cpu: "192", mem: "755Gi",
			npuAllocateNum: "2", npuTop: "Ascend310-0,Ascend310-4"})

		Convey("GetAllocatedNPUFromTopologyFn() should return error when reqNpuNum of pod is 0", func() {
			task.Resreq.ScalarResources[a310NPUCardName] = 0
			_, err := npu.GetAllocatedNPUFromTopologyFn(task, node, false)
			So(err, ShouldBeError)
		})

		Convey("GetAllocatedNPUFromTopologyFn() should return error when reqNpuNum Is greater than 4", func() {
			task.Resreq.ScalarResources[a310NPUCardName] = constInt5000
			_, err := npu.GetAllocatedNPUFromTopologyFn(task, node, false)
			So(err, ShouldBeError)
		})

		Convey("GetAllocatedNPUFromTopologyFn() should return nil when npuTop of node is wrong", func() {
			task.Resreq.ScalarResources[a310NPUCardName] = constInt2000
			node.Others[a310NPUCardName] = "Ascend310-0Ascend310-4"
			result, err := npu.GetAllocatedNPUFromTopologyFn(task, node, false)
			So(err, ShouldBeNil)
			So(result, ShouldBeNil)
		})

		Convey("GetAllocatedNPUFromTopologyFn() should return correct result when reqNpuNum is 0", func() {
			task.Resreq.ScalarResources[a310NPUCardName] = 0
			node.Node.Annotations[a310NPUCardName] = "Ascend310-0,Ascend310-3"
			result, err := npu.GetAllocatedNPUFromTopologyFn(task, node, false)
			So(err, ShouldBeError)
			So(result, ShouldBeNil)
		})

		Convey("GetAllocatedNPUFromTopologyFn() should return error when none node meet request", func() {
			task.Resreq.ScalarResources[a310NPUCardName] = constInt2000
			node.Node.Annotations[a310NPUCardName] = "Ascend310-0"
			_, err := npu.GetAllocatedNPUFromTopologyFn(task, node, false)
			So(err, ShouldBeError)
		})
	})
}

// nodeTop: [0, 2, 4, 5, 7, 8, 9, 10, 11, 12, 13, 15, 16, 18, 20, 21, 22, 25, 29, 32, 33, 34, 35, 36, 37, 38, 39]
// cardNumGroups: [2, 3, 4, 3, 2, 3, 1, 1, 4, 4] Each group of four NPU
func TestCnpuGetFitCardFromNodeByPriorityFn(t *testing.T) {
	Convey("Test card310x4 GetFitCardFromNodeByPriorityFn", t, func() {
		nodeTop := []int{NPUID0, NPUID2, NPUID4, NPUID5, NPUID7, NPUID8, NPUID9, NPUID10, NPUID11, NPUID12,
			NPUID13, NPUID15, NPUID16, NPUID18, NPUID20, NPUID21, NPUID22, NPUID25, NPUID29, NPUID32, NPUID33,
			NPUID34, NPUID35, NPUID36, NPUID37, NPUID38, NPUID39}

		Convey("getFitCardFromNodeByPriorityFn() should return correct result "+
			"when priorityArray is [1, 3, 2, 4]", func() {
			priorityArray := [cardNPUNumber]int{1, 3, 2, 4}
			result, err := getFitCardFromNodeByPriority(nodeTop, priorityArray)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
		})

		Convey("getFitCardFromNodeByPriorityFn() should return correct result "+
			"when priorityArray is [2, 3, 4]", func() {
			priorityArray := [cardNPUNumber]int{2, 3, 4}
			result, err := getFitCardFromNodeByPriority(nodeTop, priorityArray)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
		})

		Convey("getFitCardFromNodeByPriorityFn() should return correct result "+
			"when priorityArray is [3, 4]", func() {
			priorityArray := [cardNPUNumber]int{3, 4}
			result, err := getFitCardFromNodeByPriority(nodeTop, priorityArray)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
		})

		Convey("getFitCardFromNodeByPriorityFn() should return correct result "+
			"when priorityArray is [4]", func() {
			priorityArray := [cardNPUNumber]int{4}
			result, err := getFitCardFromNodeByPriority(nodeTop, priorityArray)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
		})
	})
}

func TestCnpuGetFitCardFromNodeByPriorityFnNotMatch(t *testing.T) {
	Convey("Test card310x4 GetFitCardFromNodeByPriorityFn", t, func() {
		nodeTop := []int{NPUID0, NPUID1, NPUID2, NPUID3, NPUID6, NPUID7, NPUID8, NPUID9, NPUID11}

		Convey("getFitCardFromNodeByPriorityFn() should return correct result "+
			"when priorityArray is [1, 3, 2, 4]", func() {
			priorityArray := [cardNPUNumber]int{1, 3, 2, 4}
			result, err := getFitCardFromNodeByPriority(nodeTop, priorityArray)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
		})
	})
}

// TestCnpuSetNPUTopologyToPodFn
func TestCnpuSetNPUTopologyToPodFn(t *testing.T) {
	Convey("Test card310x4 SetNPUTopologyToPodFn", t, func() {
		npu := &card310x4{}
		task := vapi.NewTaskInfo(buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-22",
			podName: "npu-test-22", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: a310NPUCardName, reqNpuNum: "4"}))
		Convey("SetNPUTopologyToPodFn() should return error when top is of wrong type", func() {
			var top []string
			err := npu.SetNPUTopologyToPodFn(task, top)
			So(err, ShouldBeError)
			So(task.Pod.Annotations[a310NPUCardName], ShouldEqual, "")
		})
		Convey("SetNPUTopologyToPodFn() should write correct info in pod annotation", func() {
			top := []int{0, 4}
			expectedResult := "Ascend310-0,Ascend310-4"
			err := npu.SetNPUTopologyToPodFn(task, top)
			So(err, ShouldBeNil)
			So(task.Pod.Annotations[a310NPUCardName], ShouldEqual, expectedResult)
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

	pod := util2.BuildPod(podInfo.namespace, podInfo.podName, podInfo.nodeName, v1.PodPending,
		buildNPUResourceList(podInfo.reqCPUNum, podInfo.reqMem,
			v1.ResourceName(podInfo.reqNPUType), podInfo.reqNpuNum),
		podInfo.groupName, make(map[string]string, util.ConstIntNum2),
		make(map[string]string, util.ConstIntNum2))

	setPodSelector(pod, archSelector, huaweiArchX86)
	ascendtest.SetTestNPUPodSelector(pod, archSelector, huaweiArchX86)
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
	nodeCapacity := buildNPUResourceList(CNode.cpu, CNode.mem, a310NPUCardName, cs.Itoa(util.ConstIntNum2))
	nodeAlloc := buildNPUResourceList(CNode.cpu, CNode.mem, a310NPUCardName, CNode.npuAllocateNum)
	labels := make(map[string]string, util.ConstIntNum2)
	ann := make(map[string]string, util.ConstIntNum2)

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
		v1node.Annotations[a310NPUCardName] = CNode.npuTop
	}

	setNodeLabel(v1node, archSelector, CNode.nodeArch)

	node := vapi.NewNodeInfo(v1node)
	setNodeOthers(node)
	return node
}

func setNodeOthers(node *vapi.NodeInfo) {
	node.Others = make(map[string]interface{}, 1)
	for k, v := range node.Node.Annotations {
		node.Others[k] = v
	}
}
