/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package module910x8 is using for HuaWei A800/9000 Ascend910 pin affinity schedule.

*/
package module910x8

import (
	"strconv"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/conf"
	"volcano.sh/volcano/pkg/scheduler/util"
)

type MPodInfo struct {
	namespace  string
	groupName  string
	podName    string
	nodeName   string
	reqCPUNum  string
	reqMem     string
	reqNPUType string
	reqNpuNum  string
}

type MNodeInfo struct {
	nodeName       string
	nodeArch       string
	cpu, mem       string
	npuAllocateNum string
	npuTop         string
}

const (
	acceleratorType     = "accelerator-type"
	cardAcceleratorType = "card"
	nodeName            = "centos"
)

// TestMNPUName
func TestMNPUName(t *testing.T) {
	convey.Convey("Test module91 0x8 Name", t, func() {
		npu := &module910x8{}

		convey.Convey("Name() should return PluginName defined in const", func() {
			n := npu.Name()
			convey.So(n, convey.ShouldEqual, PluginName)
		})
	})
}

// TestMNPUIsMyTask
func TestMNPUIsMyTask(t *testing.T) {
	convey.Convey("Test module910x8 IsMyTask", t, func() {
		npu := &module910x8{}

		convey.Convey("IsMyTask() should return error when task doesn't request npu", func() {
			pod := buildNPUPod(MPodInfo{namespace: "default", groupName: "npu-group-106",
				podName: "npu-test-106", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
				reqNPUType: npu800And9000CardName, reqNpuNum: "0"})
			task := api.NewTaskInfo(pod)
			result := npu.IsMyTask(task)
			convey.So(result, convey.ShouldBeError)
		})
		convey.Convey("IsMyTask() should return nil when task is of card type", func() {
			pod := buildNPUPod(MPodInfo{namespace: "default", groupName: "npu-group-107",
				podName: "npu-test-107", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
				reqNPUType: npu800And9000CardName, reqNpuNum: "1"})
			setPodSelector(pod, acceleratorType, cardAcceleratorType)
			task := api.NewTaskInfo(pod)
			result := npu.IsMyTask(task)
			convey.So(result, convey.ShouldBeError)
		})
		convey.Convey("IsMyTask() should return nil when task is of module type", func() {
			pod := buildNPUPod(MPodInfo{namespace: "default", groupName: "npu-group-108",
				podName: "npu-test-108", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
				reqNPUType: npu800And9000CardName, reqNpuNum: "1"})
			setPodSelector(pod, acceleratorType, moduleAcceleratorType)
			task := api.NewTaskInfo(pod)
			result := npu.IsMyTask(task)
			convey.So(result, convey.ShouldBeNil)
		})
	})
}

// TestMNPUIsMyNode
func TestMNPUIsMyNode(t *testing.T) {
	convey.Convey("Test module910x8 IsMyNode", t, func() {
		npu := &module910x8{}

		convey.Convey("IsMyNode() should return error when node has no 	npu annotation", func() {
			node := buildNPUNode(MNodeInfo{nodeName: nodeName, nodeArch: huaweiArchArm, cpu: "192", mem: "755Gi",
				npuAllocateNum: "0", npuTop: "0"})
			result := npu.IsMyNode(node)
			convey.So(result, convey.ShouldBeError)
		})
		convey.Convey("IsMyNode() should return error when value of needed node selector is wrong", func() {
			node := buildNPUNode(MNodeInfo{nodeName: nodeName, nodeArch: huaweiArchArm, cpu: "192", mem: "755Gi",
				npuAllocateNum: "1", npuTop: "Ascend910-12"})
			setNodeLabel(node.Node, acceleratorType, cardAcceleratorType)
			result := npu.IsMyNode(node)
			convey.So(result, convey.ShouldBeError)
		})
		convey.Convey("IsMyNode() should return nil when node is of module mode", func() {
			node := buildNPUNode(MNodeInfo{nodeName: nodeName, nodeArch: huaweiArchArm, cpu: "192", mem: "755Gi",
				npuAllocateNum: "1", npuTop: "Ascend910-16c-152-0"})
			setNodeLabel(node.Node, acceleratorType, moduleAcceleratorType)
			result := npu.IsMyNode(node)
			convey.So(result, convey.ShouldBeNil)
		})
		convey.Convey("IsMyNode() should return nil when node has no selector 'accelerator-type'", func() {
			node := buildNPUNode(MNodeInfo{nodeName: nodeName, nodeArch: huaweiArchArm, cpu: "192", mem: "755Gi",
				npuAllocateNum: "1", npuTop: "Ascend910-16c-153-0"})
			result := npu.IsMyNode(node)
			convey.So(result, convey.ShouldBeNil)
		})
	})
}

// TestMNPUIsMyJob
func TestMNPUIsMyJob(t *testing.T) {
	convey.Convey("Test module910x8 IsMyJob", t, func() {
		vnpu := &module910x8{}
		var tasks []*api.TaskInfo
		uid := api.JobID("xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxx11")

		convey.Convey("IsMyJob() should return error when job request no NPU", func() {
			tasks = append(tasks, api.NewTaskInfo(buildNPUPod(
				MPodInfo{namespace: "default", groupName: "npu-group-109",
					podName: "npu-test-109", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
					reqNPUType: npu800And9000CardName, reqNpuNum: "0"})))
			job := api.NewJobInfo(uid, tasks...)
			result := vnpu.IsMyJob(job)
			convey.So(result, convey.ShouldBeError)
		})
		convey.Convey("IsMyJob() should return nil when job is of card type", func() {
			pod := buildNPUPod(MPodInfo{namespace: "default", groupName: "npu-group-110",
				podName: "npu-test-110", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
				reqNPUType: npu800And9000CardName, reqNpuNum: "1"})
			setPodSelector(pod, acceleratorType, cardAcceleratorType)
			tasks = append(tasks, api.NewTaskInfo(pod))
			job := api.NewJobInfo(uid, tasks...)
			result := vnpu.IsMyJob(job)
			convey.So(result, convey.ShouldBeError)
		})
		convey.Convey("IsMyJob() should return nil when job is of module type", func() {
			tasks = append(tasks, api.NewTaskInfo(buildNPUPod(
				MPodInfo{namespace: "default", groupName: "npu-group-110",
					podName: "npu-test-110", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
					reqNPUType: npu800And9000CardName, reqNpuNum: "1"})))
			job := api.NewJobInfo(uid, tasks...)
			result := vnpu.IsMyJob(job)
			convey.So(result, convey.ShouldBeNil)
		})
	})
}

// TestMNPUValidNPUJobFnInvalidSelector
func TestMNPUValidNPUJobFnInvalidSelector(t *testing.T) {
	convey.Convey("Test module910x8 ValidNPUJobFn", t, func() {
		const (
			invalidSelectorKey   = "invalid-key"
			invalidSelectorValue = "no-selector"
			validNum             = 8000
		)
		npu := &module910x8{}
		var tasks []*api.TaskInfo
		uid := api.JobID("xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx4")

		convey.Convey("ValidNPUJobFn() should return error for job without certain selector key", func() {
			pod := buildNPUPod(MPodInfo{namespace: "default", groupName: "npu-group-10",
				podName: "npu-test-33", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: npu800And9000CardName, reqNpuNum: "1"})
			delete(pod.Spec.NodeSelector, archSelector)
			setPodSelector(pod, invalidSelectorKey, invalidSelectorValue)
			tasks = append(tasks, api.NewTaskInfo(pod))
			job := api.NewJobInfo(uid, tasks...)
			setJobResourceReq(job, npu800And9000CardName, float64(validNum))
			result := npu.ValidNPUJobFn(job)
			convey.So(result, convey.ShouldNotBeNil)
		})
		convey.Convey("ValidNPUJobFn() should return error for job with invalid selector value", func() {
			pod := buildNPUPod(MPodInfo{namespace: "default", groupName: "npu-group-11",
				podName: "npu-test-34", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: npu800And9000CardName, reqNpuNum: "1"})
			setPodSelector(pod, archSelector, invalidSelectorValue)
			tasks = append(tasks, api.NewTaskInfo(pod))
			job := api.NewJobInfo(uid, tasks...)
			setJobResourceReq(job, npu800And9000CardName, float64(validNum))
			result := npu.ValidNPUJobFn(job)
			convey.So(result, convey.ShouldNotBeNil)
		})
	})
}

// TestMnpuValidNPUJobFnInvalidNum
func TestMNPUValidNPUJobFnInvalidNum(t *testing.T) {
	convey.Convey("Test module910x8 ValidNPUJobFnInvalidNum", t, func() {
		const (
			invalidNum0 = "0"
			invalidNum3 = "3"
			validNum8   = "8"
		)
		npu := &module910x8{}
		var tasks []*api.TaskInfo
		uid := api.JobID("xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx6")

		convey.Convey("ValidNPUJobFn() should return error for job with invalid request number 0", func() {
			tasks = append(tasks, api.NewTaskInfo(buildNPUPod(
				MPodInfo{namespace: "default", groupName: "npu-group-14",
					podName: "npu-test-37", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
					reqNPUType: npu800And9000CardName, reqNpuNum: invalidNum0})))
			job := api.NewJobInfo(uid, tasks...)
			result := npu.ValidNPUJobFn(job)
			convey.So(result, convey.ShouldNotBeNil)
		})
		convey.Convey("ValidNPUJobFn() should return error for job with invalid request number 3", func() {
			tasks = append(tasks, api.NewTaskInfo(buildNPUPod(
				MPodInfo{namespace: "default", groupName: "npu-group-15",
					podName: "npu-test-38", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
					reqNPUType: npu800And9000CardName, reqNpuNum: invalidNum3})))
			job := api.NewJobInfo(uid, tasks...)
			result := npu.ValidNPUJobFn(job)
			convey.So(result, convey.ShouldNotBeNil)
		})
		convey.Convey("ValidNPUJobFn() should return error for job with valid request number 8", func() {
			tasks = append(tasks, api.NewTaskInfo(buildNPUPod(
				MPodInfo{namespace: "default", groupName: "npu-group-16",
					podName: "npu-test-39", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
					reqNPUType: npu800And9000CardName, reqNpuNum: validNum8})))
			job := api.NewJobInfo(uid, tasks...)
			result := npu.ValidNPUJobFn(job)
			convey.So(result, convey.ShouldBeNil)
		})
	})
}

// TestMNPUValidNPUJobFnInvalidModel
func TestMNPUValidNPUJobFnInvalidModel(t *testing.T) {
	convey.Convey("Test module910x8 ValidNPUJobFnInvalidModel", t, func() {
		const (
			num8  = "8"
			num16 = "16"
		)
		npu := &module910x8{}
		var tasks []*api.TaskInfo
		uid := api.JobID("xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx7")

		convey.Convey("ValidNPUJobFn() should return error for job with invalid single model", func() {
			tasks = append(tasks, api.NewTaskInfo(buildNPUPod(
				MPodInfo{namespace: "default", groupName: "npu-group-20",
					podName: "npu-test-43", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
					reqNPUType: npu800And9000CardName, reqNpuNum: num16})))
			job := api.NewJobInfo(uid, tasks...)
			result := npu.ValidNPUJobFn(job)
			convey.So(result, convey.ShouldNotBeNil)
		})
		convey.Convey("ValidNPUJobFn() should return nil for job with valid model", func() {
			tasks = append(tasks, api.NewTaskInfo(buildNPUPod(
				MPodInfo{namespace: "default", groupName: "npu-group-21",
					podName: "npu-test-44", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
					reqNPUType: npu800And9000CardName, reqNpuNum: num8})))
			tasks = append(tasks, api.NewTaskInfo(buildNPUPod(
				MPodInfo{namespace: "default", groupName: "npu-group-22",
					podName: "npu-test-45", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
					reqNPUType: npu800And9000CardName, reqNpuNum: num8})))
			job := api.NewJobInfo(uid, tasks...)
			result := npu.ValidNPUJobFn(job)
			convey.So(result, convey.ShouldBeNil)
		})
	})
}

// TestMNPUPreCheckNodeFnTaskError
func TestMNPUPreCheckNodeFnTaskError(t *testing.T) {
	convey.Convey("Test module910x8 PreCheckNodeFn", t, func() {
		npu := &module910x8{}
		var confs []conf.Configuration
		node := buildNPUNode(MNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
			npuAllocateNum: "1", npuTop: "Ascend910-1"})

		convey.Convey("PreCheckNodeFn() should return error when task don't have selector", func() {
			pod := buildNPUPod(MPodInfo{namespace: "default", groupName: "npu-group-8",
				podName: "npu-test-49", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: npu800And9000CardName, reqNpuNum: "1"})
			delete(pod.Spec.NodeSelector, archSelector)
			task := api.NewTaskInfo(pod)
			result := npu.PreCheckNodeFn(task, node, confs)
			convey.So(result, convey.ShouldBeError)
		})
		convey.Convey("PreCheckNodeFn() should return nil when task isn't npu task", func() {
			pod := buildNPUPod(MPodInfo{namespace: "default", groupName: "npu-group-9",
				podName: "npu-test-50", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: "", reqNpuNum: ""})
			delete(pod.Spec.NodeSelector, archSelector)
			task := api.NewTaskInfo(pod)
			result := npu.PreCheckNodeFn(task, node, confs)
			convey.So(result, convey.ShouldBeNil)
		})
	})
}

// TestMNPUPreCheckNodeFnNodeError
func TestMNPUPreCheckNodeFnNodeError(t *testing.T) {
	convey.Convey("Test module910x8 PreCheckNodeFnNodeError", t, func() {
		npu := &module910x8{}
		var confs []conf.Configuration
		node := buildNPUNode(MNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
			npuAllocateNum: "1", npuTop: "Ascend910-8"})

		convey.Convey("PreCheckNodeFn() should return error when node don't have label", func() {
			task := api.NewTaskInfo(buildNPUPod(MPodInfo{namespace: "default", groupName: "npu-group-10",
				podName: "npu-test-51", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: npu800And9000CardName, reqNpuNum: "1"}))
			setNodeLabel(node.Node, archSelector, "")
			result := npu.PreCheckNodeFn(task, node, confs)
			convey.So(result, convey.ShouldBeError)
		})
		convey.Convey("PreCheckNodeFn() should return error when selectors mismatch with labels", func() {
			task := api.NewTaskInfo(buildNPUPod(MPodInfo{namespace: "default", groupName: "npu-group-11",
				podName: "npu-test-52", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: npu800And9000CardName, reqNpuNum: "1"}))
			// build a node with mismatch selector
			nodeArm := buildNPUNode(MNodeInfo{nodeName: nodeName, nodeArch: huaweiArchArm, cpu: "192", mem: "755Gi",
				npuAllocateNum: "1", npuTop: "Ascend910-4"})
			result := npu.PreCheckNodeFn(task, nodeArm, confs)
			convey.So(result, convey.ShouldBeError)
		})
	})
}

// TestMnpuPreCheckNodeFnSuccess
func TestMnpuPreCheckNodeFnSuccess(t *testing.T) {
	convey.Convey("Test module910x8 PreCheckNodeFnSuccess", t, func() {
		npu := &module910x8{}
		var confs []conf.Configuration
		node := buildNPUNode(MNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
			npuAllocateNum: "1", npuTop: "Ascend910-5"})

		convey.Convey("PreCheckNodeFn() should return nil when selectors match with labels", func() {
			// build a task with no selector
			task := api.NewTaskInfo(buildNPUPod(MPodInfo{namespace: "default", groupName: "npu-group-12",
				podName: "npu-test-53", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: npu800And9000CardName, reqNpuNum: "1"}))
			result := npu.PreCheckNodeFn(task, node, confs)
			convey.So(result, convey.ShouldBeNil)
		})
	})
}

// TestMnpuCheckNPUResourceStableFn
func TestMnpuCheckNPUResourceStableFn(t *testing.T) {
	convey.Convey("Test module910x8 CheckNPUResourceStableFn", t, func() {
		npu := &module910x8{}

		convey.Convey("CheckNPUResourceStableFn() should return error when there's missing resource type in idle", func() {
			node := buildNPUNode(MNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "0", npuTop: ""})
			node.Node.Annotations[npu800And9000CardName] = "Ascend910-1"
			result := npu.CheckNPUResourceStableFn(node)
			convey.So(result, convey.ShouldBeError)
		})
		convey.Convey("CheckNPUResourceStableFn() should return error when node resources are unstable", func() {
			node := buildNPUNode(MNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "1", npuTop: ""})
			node.Node.Annotations[npu800And9000CardName] = "Ascend910-1,Ascend910-2"
			result := npu.CheckNPUResourceStableFn(node)
			convey.So(result, convey.ShouldBeError)
		})
		convey.Convey("CheckNPUResourceStableFn() should return nil when node resources are stable", func() {
			node := buildNPUNode(MNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "2", npuTop: "Ascend910-2,Ascend910-3"})
			result := npu.CheckNPUResourceStableFn(node)
			convey.So(result, convey.ShouldBeNil)
		})
	})
}

// TestMnpuCheckNodeNPUByTaskFn
func TestMnpuCheckNodeNPUByTaskFn(t *testing.T) {
	convey.Convey("Test module910x8 CheckNodeNPUByTaskFn", t, func() {
		npu := &module910x8{}
		pod := buildNPUPod(MPodInfo{namespace: "default", groupName: "npu-group-13",
			podName: "npu-test-60", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: npu800And9000CardName, reqNpuNum: "4"})
		task := api.NewTaskInfo(pod)

		convey.Convey("CheckNodeNPUByTaskFn() should return error when node doesn't meet task requests", func() {
			node := buildNPUNode(MNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "2", npuTop: "Ascend910-5,Ascend910-2"})
			result := npu.CheckNodeNPUByTaskFn(task, node, true)
			convey.So(result, convey.ShouldBeError)
		})
		convey.Convey("CheckNodeNPUByTaskFn() should return error when node meets task requests", func() {
			node := buildNPUNode(MNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "4", npuTop: "Ascend910-1,Ascend910-0,Ascend910-3,Ascend910-2"})
			result := npu.CheckNodeNPUByTaskFn(task, node, true)
			convey.So(result, convey.ShouldBeNil)
		})
	})
}

// TestMnpuGetNPUAffinityBestNodesFn
func TestMnpuGetNPUAffinityBestNodesFn(t *testing.T) {
	convey.Convey("Test module910x8 GetNPUAffinityBestNodesFn", t, func() {
		const (
			nodeName1 = "centos1"
			nodeName2 = "centos2"
			constNum  = 0
		)
		npu := &module910x8{}
		pod := buildNPUPod(MPodInfo{namespace: "default", groupName: "npu-group-61",
			podName: "npu-test-61", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: npu800And9000CardName, reqNpuNum: "4"})
		task := api.NewTaskInfo(pod)
		var nodes []*api.NodeInfo

		convey.Convey("GetNPUAffinityBestNodesFn() should return correct result", func() {
			node1 := buildNPUNode(MNodeInfo{nodeName: nodeName1, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "5", npuTop: "Ascend910-5,Ascend910-2,Ascend910-0,Ascend910-1,Ascend910-3"})
			nodes = append(nodes, node1)
			node2 := buildNPUNode(MNodeInfo{nodeName: nodeName2, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "6", npuTop: "Ascend910-5,Ascend910-4,Ascend910-6,Ascend910-7,Ascend910-0,Ascend910-1"})
			nodes = append(nodes, node2)

			result, err := npu.GetNPUAffinityBestNodesFn(task, nodes, false)

			convey.So(result[nodeName1], convey.ShouldEqual, constNum)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestMnpuScoreBestNPUNodesFn
func TestMnpuScoreBestNPUNodesFn(t *testing.T) {
	convey.Convey("Test module910x8 ScoreBestNPUNodesFn", t, func() {
		const (
			nodeName1 = "euler1"
			nodeName2 = "euler2"
		)
		npu := &module910x8{}
		bestNodes := map[string]int{
			nodeName1: 0,
			nodeName2: constIntNum3,
		}
		pod := buildNPUPod(MPodInfo{namespace: "default", groupName: "npu-group-62",
			podName: "npu-test-62", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: npu800And9000CardName, reqNpuNum: "4"})
		task := api.NewTaskInfo(pod)

		var nodes []*api.NodeInfo
		node1 := buildNPUNode(MNodeInfo{nodeName: nodeName1, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
			npuAllocateNum: "8", npuTop: "Ascend910-5,Ascend910-2,Ascend910-0,Ascend910-1,Ascend910-3," +
				"Ascend910-6,Ascend910-7"})
		nodes = append(nodes, node1)
		node2 := buildNPUNode(MNodeInfo{nodeName: nodeName2, nodeArch: huaweiArchArm, cpu: "192", mem: "755Gi",
			npuAllocateNum: "8", npuTop: "Ascend910-5,Ascend910-2,Ascend910-0,Ascend910-1,Ascend910-3," +
				"Ascend910-6,Ascend910-7"})
		nodes = append(nodes, node2)

		convey.Convey("ScoreBestNPUNodesFn() should return err when scoreMap is nil", func() {
			var scoreMap map[string]float64
			_, err := npu.ScoreBestNPUNodesFn(scoreMap, bestNodes, task, nodes)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("ScoreBestNPUNodesFn() should return correct result", func() {
			scoreMap := make(map[string]float64)
			expectedResult := map[string]float64(nil)
			result, err := npu.ScoreBestNPUNodesFn(scoreMap, bestNodes, task, nodes)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(result, convey.ShouldResemble, expectedResult)
		})
	})
}

// TestMnpuGetAllocatedNPUFromTopologyFn
func TestMnpuGetAllocatedNPUFromTopologyFn(t *testing.T) {
	convey.Convey("Test module910x8 GetAllocatedNPUFromTopologyFn", t, func() {
		npu := &module910x8{}

		convey.Convey("GetAllocatedNPUFromTopologyFn()", func() {
			pod := buildNPUPod(MPodInfo{namespace: "default", groupName: "npu-group-64",
				podName: "npu-test-64", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: npu800And9000CardName, reqNpuNum: "8"})
			task := api.NewTaskInfo(pod)
			node := buildNPUNode(MNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "8", npuTop: ""})
			setNodeAnnotation(node.Node, npu800And9000CardName, "")
			var expectedResult []int
			result, err := npu.GetAllocatedNPUFromTopologyFn(task, node, false)
			convey.So(err, convey.ShouldBeError)
			convey.So(result, convey.ShouldResemble, expectedResult)
		})
		convey.Convey("GetAllocatedNPUFromTopologyFn() should return correct result case 1", func() {
			pod := buildNPUPod(MPodInfo{namespace: "default", groupName: "npu-group-63",
				podName: "npu-test-63", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: npu800And9000CardName, reqNpuNum: "4"})
			task := api.NewTaskInfo(pod)
			node := buildNPUNode(MNodeInfo{nodeName: nodeName, nodeArch: huaweiArchArm, cpu: "192", mem: "755Gi",
				npuAllocateNum: "5", npuTop: "Ascend910-5,Ascend910-2,Ascend910-0,Ascend910-1,Ascend910-3"})
			expectedResult := []int{constIntNum2, 0, 1, constIntNum3}
			result, err := npu.GetAllocatedNPUFromTopologyFn(task, node, false)
			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldResemble, expectedResult)
		})
		convey.Convey("GetAllocatedNPUFromTopologyFn() should return correct result case 2", func() {
			pod := buildNPUPod(MPodInfo{namespace: "default", groupName: "npu-group-65",
				podName: "npu-test-65", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: npu800And9000CardName, reqNpuNum: "2"})
			task := api.NewTaskInfo(pod)
			node := buildNPUNode(MNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "2", npuTop: "Ascend910-5,Ascend910-7"})
			expectedResult := []int{constIntNum5, constIntNum7}
			result, err := npu.GetAllocatedNPUFromTopologyFn(task, node, false)
			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldResemble, expectedResult)
		})
	})
}

// TestMnpuSetNPUTopologyToPodFn
func TestMnpuSetNPUTopologyToPodFn(t *testing.T) {
	convey.Convey("Test module910x8 SetNPUTopologyToPodFn", t, func() {
		npu := &module910x8{}
		task := api.NewTaskInfo(buildNPUPod(MPodInfo{namespace: "default", groupName: "npu-group-64",
			podName: "npu-test-64", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: npu800And9000CardName, reqNpuNum: "4"}))
		convey.Convey("SetNPUTopologyToPodFn() should return error when top is of wrong type", func() {
			var top []string
			err := npu.SetNPUTopologyToPodFn(task, top)
			convey.So(err, convey.ShouldBeError)
			convey.So(task.Pod.Annotations[npu800And9000CardName], convey.ShouldEqual, "")
		})
		convey.Convey("SetNPUTopologyToPodFn() should write correct info in pod annotation", func() {
			top := []int{constIntNum2, 0, 1, constIntNum3}
			expectedResult := "Ascend910-2,Ascend910-0,Ascend910-1,Ascend910-3"
			err := npu.SetNPUTopologyToPodFn(task, top)
			convey.So(err, convey.ShouldBeNil)
			convey.So(task.Pod.Annotations[npu800And9000CardName], convey.ShouldEqual, expectedResult)
		})
	})
}

// TestMnpuUpdateNPUNodeUsedCardFn
func TestMnpuUpdateNPUNodeUsedCardFn(t *testing.T) {
	convey.Convey("Test module910x8 UpdateNPUNodeUsedCardFn", t, func() {
		npu := &module910x8{}
		node := buildNPUNode(MNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
			npuAllocateNum: "4", npuTop: "Ascend910-2,Ascend910-0,Ascend910-1,Ascend910-3"})
		node.Others = map[string]interface{}{
			npu800And9000CardName: "Ascend910-2,Ascend910-0,Ascend910-1,Ascend910-3",
		}
		convey.Convey("UpdateNPUNodeUsedCardFn() should successfully update node.others", func() {
			top := []int{0, 1}
			expectedResult := map[string]interface{}{
				npu800And9000CardName: "Ascend910-2,Ascend910-3",
			}
			err := npu.UpdateNPUNodeUsedCardFn(node, top)
			convey.So(err, convey.ShouldBeNil)
			convey.So(node.Others, convey.ShouldResemble, expectedResult)
		})
	})
}

// TestMnpuGetReleaseNPUTopologyFn
func TestMnpuGetReleaseNPUTopologyFn(t *testing.T) {
	convey.Convey("Test module910x8 GetReleaseNPUTopologyFn", t, func() {
		npu := &module910x8{}
		task := api.NewTaskInfo(buildNPUPod(MPodInfo{namespace: "default", groupName: "npu-group-64",
			podName: "npu-test-64", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: npu800And9000CardName, reqNpuNum: "4"}))
		convey.Convey("GetReleaseNPUTopologyFn() should return correct card id slice", func() {
			task.Pod.Annotations[npu800And9000CardName] = "Ascend910-0,Ascend910-1,Ascend910-2,Ascend910-3,Ascend910-6"
			expectedResult := []int{0, 1, constIntNum2, constIntNum3, constIntNum6}
			result, err := npu.GetReleaseNPUTopologyFn(task)
			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldResemble, expectedResult)
		})
	})
}

// TestMnpuUpdateReleaseNPUNodeTopologyFn
func TestMnpuUpdateReleaseNPUNodeTopologyFn(t *testing.T) {
	convey.Convey("Test module910x8 UpdateReleaseNPUNodeTopologyFn", t, func() {
		npu := &module910x8{}
		node := buildNPUNode(MNodeInfo{nodeName: nodeName, nodeArch: huaweiArchX86, cpu: "192", mem: "755Gi",
			npuAllocateNum: "4", npuTop: "Ascend910-4,Ascend910-6,Ascend910-7,Ascend910-5"})
		node.Others = map[string]interface{}{
			npu800And9000CardName: "Ascend910-4,Ascend910-6,Ascend910-7,Ascend910-5",
		}
		convey.Convey("UpdateNPUNodeUsedCardFn() should successfully update node.others", func() {
			top := []int{0, 1}
			err := npu.UpdateReleaseNPUNodeTopologyFn(node, top)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func setPodSelector(MPod *v1.Pod, selectorKey string, selectorValue string) {
	MPod.Spec.NodeSelector[selectorKey] = selectorValue
}

func setNodeLabel(MNode *v1.Node, labelKey string, labelValue string) {
	if labelValue == "" {
		delete(MNode.Labels, labelKey)
		return
	}
	MNode.Labels[labelKey] = labelValue
}

func setNodeAnnotation(MNode *v1.Node, annKey string, annValue string) {
	if annValue == "" {
		delete(MNode.Annotations, annKey)
		return
	}
	MNode.Annotations[annKey] = annValue
}

func setJobResourceReq(MJob *api.JobInfo, resource string, num float64) {
	MJob.TotalRequest.ScalarResources[v1.ResourceName(resource)] = num
}

func buildNPUPod(podInfo MPodInfo) *v1.Pod {

	pod := util.BuildPod(podInfo.namespace, podInfo.podName, podInfo.nodeName, v1.PodPending,
		buildNPUResourceList(podInfo.reqCPUNum, podInfo.reqMem,
			v1.ResourceName(podInfo.reqNPUType), podInfo.reqNpuNum),
		podInfo.groupName, make(map[string]string, constIntNum2),
		make(map[string]string, constIntNum2))

	setPodSelector(pod, archSelector, huaweiArchX86)

	return pod
}

func buildNPUResourceList(MCpu string, MMemory string, npuResourceType v1.ResourceName, npu string) v1.ResourceList {
	npuNum, err := strconv.Atoi(npu)
	if err != nil {
		return nil
	}

	if npuNum == 0 {
		return v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse(MCpu),
			v1.ResourceMemory: resource.MustParse(MMemory),
		}
	}

	return v1.ResourceList{
		v1.ResourceCPU:    resource.MustParse(MCpu),
		v1.ResourceMemory: resource.MustParse(MMemory),
		npuResourceType:   resource.MustParse(npu),
	}
}

func buildNPUNode(MNode MNodeInfo) *api.NodeInfo {
	nodeCapacity := buildNPUResourceList(MNode.cpu, MNode.mem, npu800And9000CardName, strconv.Itoa(constIntNum2))
	nodeAlloc := buildNPUResourceList(MNode.cpu, MNode.mem, npu800And9000CardName, MNode.npuAllocateNum)
	labels := make(map[string]string, constIntNum2)
	ann := make(map[string]string, constIntNum2)

	v1node := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:        MNode.nodeName,
			Labels:      labels,
			Annotations: ann,
		},
		Status: v1.NodeStatus{
			Capacity:    nodeCapacity,
			Allocatable: nodeAlloc,
		},
	}

	if MNode.npuAllocateNum != "0" {
		v1node.Annotations[npu800And9000CardName] = MNode.npuTop
	}

	setNodeLabel(v1node, archSelector, MNode.nodeArch)

	node := api.NewNodeInfo(v1node)
	if MNode.npuAllocateNum != "0" {
		node.Others = map[string]interface{}{
			npu800And9000CardName: MNode.npuTop,
		}
	}
	return node
}
