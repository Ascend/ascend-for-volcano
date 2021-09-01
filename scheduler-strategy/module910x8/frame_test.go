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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ms "strconv"
	"testing"
	vapi "volcano.sh/volcano/pkg/scheduler/api"
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
	Convey("Test module91 0x8 Name", t, func() {
		npu := &module910x8{}

		Convey("Name() should return PluginName defined in const", func() {
			n := npu.Name()
			So(n, ShouldEqual, PluginName)
		})
	})
}

// TestMNPUIsMyTask
func TestMNPUIsMyTask(t *testing.T) {
	Convey("Test module910x8 IsMyTask", t, func() {
		npu := &module910x8{}

		Convey("IsMyTask() should return error when task doesn't request npu", func() {
			pod := buildNPUPod(MPodInfo{namespace: "default", groupName: "npu-group-106",
				podName: "npu-test-106", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
				reqNPUType: npu800And9000CardName, reqNpuNum: "0"})
			task := vapi.NewTaskInfo(pod)
			result := npu.IsMyTask(task)
			So(result, ShouldBeError)
		})
		Convey("IsMyTask() should return nil when task is of card type", func() {
			pod := buildNPUPod(MPodInfo{namespace: "default", groupName: "npu-group-107",
				podName: "npu-test-107", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
				reqNPUType: npu800And9000CardName, reqNpuNum: "1"})
			setPodSelector(pod, acceleratorType, cardAcceleratorType)
			task := vapi.NewTaskInfo(pod)
			result := npu.IsMyTask(task)
			So(result, ShouldBeError)
		})
		Convey("IsMyTask() should return nil when task is of module type", func() {
			pod := buildNPUPod(MPodInfo{namespace: "default", groupName: "npu-group-108",
				podName: "npu-test-108", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
				reqNPUType: npu800And9000CardName, reqNpuNum: "1"})
			setPodSelector(pod, acceleratorType, moduleAcceleratorType)
			task := vapi.NewTaskInfo(pod)
			result := npu.IsMyTask(task)
			So(result, ShouldBeNil)
		})
	})
}

// TestMNPUIsMyNode
func TestMNPUIsMyNode(t *testing.T) {
	Convey("Test module910x8 IsMyNode", t, func() {
		npu := &module910x8{}

		Convey("IsMyNode() should return error when node has no 	npu annotation", func() {
			node := buildNPUNode(MNodeInfo{nodeName, huaweiArchArm, "192", "755Gi",
				"1", "Ascend910-11"})
			setNodeAnnotation(node.Node, npu800And9000CardName, "")
			result := npu.IsMyNode(node)
			So(result, ShouldBeError)
		})
		Convey("IsMyNode() should return error when value of needed node selector is wrong", func() {
			node := buildNPUNode(MNodeInfo{nodeName, huaweiArchArm, "192", "755Gi",
				"1", "Ascend910-12"})
			setNodeLabel(node.Node, acceleratorType, cardAcceleratorType)
			result := npu.IsMyNode(node)
			So(result, ShouldBeError)
		})
		Convey("IsMyNode() should return nil when node is of module mode", func() {
			node := buildNPUNode(MNodeInfo{nodeName, huaweiArchArm, "192", "755Gi",
				"1", "Ascend910-16c-152-0"})
			setNodeLabel(node.Node, acceleratorType, moduleAcceleratorType)
			result := npu.IsMyNode(node)
			So(result, ShouldBeNil)
		})
		Convey("IsMyNode() should return nil when node has no selector 'accelerator-type'", func() {
			node := buildNPUNode(MNodeInfo{nodeName, huaweiArchArm, "192", "755Gi",
				"1", "Ascend910-16c-153-0"})
			result := npu.IsMyNode(node)
			So(result, ShouldBeNil)
		})
	})
}

// TestMNPUIsMyJob
func TestMNPUIsMyJob(t *testing.T) {
	Convey("Test module910x8 IsMyJob", t, func() {
		vnpu := &module910x8{}
		tasks := []*vapi.TaskInfo{}
		uid := vapi.JobID("xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxx11")

		Convey("IsMyJob() should return error when job request no NPU", func() {
			tasks = append(tasks, vapi.NewTaskInfo(buildNPUPod(
				MPodInfo{namespace: "default", groupName: "npu-group-109",
					podName: "npu-test-109", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
					reqNPUType: npu800And9000CardName, reqNpuNum: "0"})))
			job := vapi.NewJobInfo(uid, tasks...)
			result := vnpu.IsMyJob(job)
			So(result, ShouldBeError)
		})
		Convey("IsMyJob() should return nil when job is of card type", func() {
			pod := buildNPUPod(MPodInfo{namespace: "default", groupName: "npu-group-110",
				podName: "npu-test-110", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
				reqNPUType: npu800And9000CardName, reqNpuNum: "1"})
			setPodSelector(pod, acceleratorType, cardAcceleratorType)
			tasks = append(tasks, vapi.NewTaskInfo(pod))
			job := vapi.NewJobInfo(uid, tasks...)
			result := vnpu.IsMyJob(job)
			So(result, ShouldBeError)
		})
		Convey("IsMyJob() should return nil when job is of module type", func() {
			tasks = append(tasks, vapi.NewTaskInfo(buildNPUPod(
				MPodInfo{namespace: "default", groupName: "npu-group-110",
					podName: "npu-test-110", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
					reqNPUType: npu800And9000CardName, reqNpuNum: "1"})))
			job := vapi.NewJobInfo(uid, tasks...)
			result := vnpu.IsMyJob(job)
			So(result, ShouldBeNil)
		})
	})
}

// TestMNPUValidNPUJobFnInvalidSelector
func TestMNPUValidNPUJobFnInvalidSelector(t *testing.T) {
	Convey("Test module910x8 ValidNPUJobFn", t, func() {
		const (
			invalidSelectorKey   = "invalid-key"
			invalidSelectorValue = "no-selector"
			validNum             = 8000
		)
		npu := &module910x8{}
		tasks := []*vapi.TaskInfo{}
		uid := vapi.JobID("xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx4")

		Convey("ValidNPUJobFn() should return error for job without certain selector key", func() {
			pod := buildNPUPod(MPodInfo{namespace: "default", groupName: "npu-group-10",
				podName: "npu-test-33", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: npu800And9000CardName, reqNpuNum: "1"})
			delete(pod.Spec.NodeSelector, archSelector)
			setPodSelector(pod, invalidSelectorKey, invalidSelectorValue)
			tasks = append(tasks, vapi.NewTaskInfo(pod))
			job := vapi.NewJobInfo(uid, tasks...)
			setJobResourceReq(job, npu800And9000CardName, float64(validNum))
			result := npu.ValidNPUJobFn(job)
			So(result, ShouldNotBeNil)
		})
		Convey("ValidNPUJobFn() should return error for job with invalid selector value", func() {
			pod := buildNPUPod(MPodInfo{namespace: "default", groupName: "npu-group-11",
				podName: "npu-test-34", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: npu800And9000CardName, reqNpuNum: "1"})
			setPodSelector(pod, archSelector, invalidSelectorValue)
			tasks = append(tasks, vapi.NewTaskInfo(pod))
			job := vapi.NewJobInfo(uid, tasks...)
			setJobResourceReq(job, npu800And9000CardName, float64(validNum))
			result := npu.ValidNPUJobFn(job)
			So(result, ShouldNotBeNil)
		})
	})
}

// TestMnpuValidNPUJobFnInvalidNum
func TestMNPUValidNPUJobFnInvalidNum(t *testing.T) {
	Convey("Test module910x8 ValidNPUJobFnInvalidNum", t, func() {
		const (
			invalidNum0 = "0"
			invalidNum3 = "3"
			validNum8   = "8"
		)
		npu := &module910x8{}
		tasks := []*vapi.TaskInfo{}
		uid := vapi.JobID("xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx6")

		Convey("ValidNPUJobFn() should return error for job with invalid request number 0", func() {
			tasks = append(tasks, vapi.NewTaskInfo(buildNPUPod(
				MPodInfo{namespace: "default", groupName: "npu-group-14",
					podName: "npu-test-37", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
					reqNPUType: npu800And9000CardName, reqNpuNum: invalidNum0})))
			job := vapi.NewJobInfo(uid, tasks...)
			result := npu.ValidNPUJobFn(job)
			So(result, ShouldNotBeNil)
		})
		Convey("ValidNPUJobFn() should return error for job with invalid request number 3", func() {
			tasks = append(tasks, vapi.NewTaskInfo(buildNPUPod(
				MPodInfo{namespace: "default", groupName: "npu-group-15",
					podName: "npu-test-38", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
					reqNPUType: npu800And9000CardName, reqNpuNum: invalidNum3})))
			job := vapi.NewJobInfo(uid, tasks...)
			result := npu.ValidNPUJobFn(job)
			So(result, ShouldNotBeNil)
		})
		Convey("ValidNPUJobFn() should return error for job with valid request number 8", func() {
			tasks = append(tasks, vapi.NewTaskInfo(buildNPUPod(
				MPodInfo{namespace: "default", groupName: "npu-group-16",
					podName: "npu-test-39", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
					reqNPUType: npu800And9000CardName, reqNpuNum: validNum8})))
			job := vapi.NewJobInfo(uid, tasks...)
			result := npu.ValidNPUJobFn(job)
			So(result, ShouldBeNil)
		})
	})
}

// TestMNPUValidNPUJobFnInvalidModel
func TestMNPUValidNPUJobFnInvalidModel(t *testing.T) {
	Convey("Test module910x8 ValidNPUJobFnInvalidModel", t, func() {
		const (
			num8  = "8"
			num16 = "16"
		)
		npu := &module910x8{}
		tasks := []*vapi.TaskInfo{}
		uid := vapi.JobID("xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx7")

		Convey("ValidNPUJobFn() should return error for job with invalid single model", func() {
			tasks = append(tasks, vapi.NewTaskInfo(buildNPUPod(
				MPodInfo{namespace: "default", groupName: "npu-group-20",
					podName: "npu-test-43", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
					reqNPUType: npu800And9000CardName, reqNpuNum: num16})))
			job := vapi.NewJobInfo(uid, tasks...)
			result := npu.ValidNPUJobFn(job)
			So(result, ShouldNotBeNil)
		})
		Convey("ValidNPUJobFn() should return nil for job with valid model", func() {
			tasks = append(tasks, vapi.NewTaskInfo(buildNPUPod(
				MPodInfo{namespace: "default", groupName: "npu-group-21",
					podName: "npu-test-44", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
					reqNPUType: npu800And9000CardName, reqNpuNum: num8})))
			tasks = append(tasks, vapi.NewTaskInfo(buildNPUPod(
				MPodInfo{namespace: "default", groupName: "npu-group-22",
					podName: "npu-test-45", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
					reqNPUType: npu800And9000CardName, reqNpuNum: num8})))
			job := vapi.NewJobInfo(uid, tasks...)
			result := npu.ValidNPUJobFn(job)
			So(result, ShouldBeNil)
		})
	})
}

// TestMNPUPreCheckNodeFnTaskError
func TestMNPUPreCheckNodeFnTaskError(t *testing.T) {
	Convey("Test module910x8 PreCheckNodeFn", t, func() {
		npu := &module910x8{}
		confs := []conf.Configuration{}
		node := buildNPUNode(MNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
			"1", "Ascend910-1"})

		Convey("PreCheckNodeFn() should return error when task don't have selector", func() {
			pod := buildNPUPod(MPodInfo{namespace: "default", groupName: "npu-group-8",
				podName: "npu-test-49", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: npu800And9000CardName, reqNpuNum: "1"})
			delete(pod.Spec.NodeSelector, archSelector)
			task := vapi.NewTaskInfo(pod)
			result := npu.PreCheckNodeFn(task, node, confs)
			So(result, ShouldBeError)
		})
		Convey("PreCheckNodeFn() should return nil when task isn't npu task", func() {
			pod := buildNPUPod(MPodInfo{namespace: "default", groupName: "npu-group-9",
				podName: "npu-test-50", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: "", reqNpuNum: ""})
			delete(pod.Spec.NodeSelector, archSelector)
			task := vapi.NewTaskInfo(pod)
			result := npu.PreCheckNodeFn(task, node, confs)
			So(result, ShouldBeNil)
		})
	})
}

// TestMNPUPreCheckNodeFnNodeError
func TestMNPUPreCheckNodeFnNodeError(t *testing.T) {
	Convey("Test module910x8 PreCheckNodeFnNodeError", t, func() {
		npu := &module910x8{}
		confs := []conf.Configuration{}
		node := buildNPUNode(MNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
			"1", "Ascend910-8"})

		Convey("PreCheckNodeFn() should return error when node don't have label", func() {
			task := vapi.NewTaskInfo(buildNPUPod(MPodInfo{namespace: "default", groupName: "npu-group-10",
				podName: "npu-test-51", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: npu800And9000CardName, reqNpuNum: "1"}))
			setNodeLabel(node.Node, archSelector, "")
			result := npu.PreCheckNodeFn(task, node, confs)
			So(result, ShouldBeError)
		})
		Convey("PreCheckNodeFn() should return error when selectors mismatch with labels", func() {
			task := vapi.NewTaskInfo(buildNPUPod(MPodInfo{namespace: "default", groupName: "npu-group-11",
				podName: "npu-test-52", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: npu800And9000CardName, reqNpuNum: "1"}))
			// build a node with mismatch selector
			nodeArm := buildNPUNode(MNodeInfo{nodeName, huaweiArchArm, "192", "755Gi",
				"1", "Ascend910-4"})
			result := npu.PreCheckNodeFn(task, nodeArm, confs)
			So(result, ShouldBeError)
		})
	})
}

// TestMnpuPreCheckNodeFnSuccess
func TestMnpuPreCheckNodeFnSuccess(t *testing.T) {
	Convey("Test module910x8 PreCheckNodeFnSuccess", t, func() {
		npu := &module910x8{}
		confs := []conf.Configuration{}
		node := buildNPUNode(MNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
			"1", "Ascend910-5"})

		Convey("PreCheckNodeFn() should return nil when selectors match with labels", func() {
			// build a task with no selector
			task := vapi.NewTaskInfo(buildNPUPod(MPodInfo{namespace: "default", groupName: "npu-group-12",
				podName: "npu-test-53", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: npu800And9000CardName, reqNpuNum: "1"}))
			result := npu.PreCheckNodeFn(task, node, confs)
			So(result, ShouldBeNil)
		})
	})
}

// TestMnpuCheckNPUResourceStableFn
func TestMnpuCheckNPUResourceStableFn(t *testing.T) {
	Convey("Test module910x8 CheckNPUResourceStableFn", t, func() {
		npu := &module910x8{}

		Convey("CheckNPUResourceStableFn() should return error when there's missing resource type in idle", func() {
			node := buildNPUNode(MNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
				"0", ""})
			node.Node.Annotations[npu800And9000CardName] = "Ascend910-1"
			result := npu.CheckNPUResourceStableFn(node)
			So(result, ShouldBeError)
		})
		Convey("CheckNPUResourceStableFn() should return error when node resources are unstable", func() {
			node := buildNPUNode(MNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
				"1", ""})
			node.Node.Annotations[npu800And9000CardName] = "Ascend910-1,Ascend910-2"
			result := npu.CheckNPUResourceStableFn(node)
			So(result, ShouldBeError)
		})
		Convey("CheckNPUResourceStableFn() should return nil when node resources are stable", func() {
			node := buildNPUNode(MNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
				"2", "Ascend910-2,Ascend910-3"})
			result := npu.CheckNPUResourceStableFn(node)
			So(result, ShouldBeNil)
		})
	})
}

// TestMnpuCheckNodeNPUByTaskFn
func TestMnpuCheckNodeNPUByTaskFn(t *testing.T) {
	Convey("Test module910x8 CheckNodeNPUByTaskFn", t, func() {
		npu := &module910x8{}
		pod := buildNPUPod(MPodInfo{namespace: "default", groupName: "npu-group-13",
			podName: "npu-test-60", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: npu800And9000CardName, reqNpuNum: "4"})
		task := vapi.NewTaskInfo(pod)

		Convey("CheckNodeNPUByTaskFn() should return error when node doesn't meet task requests", func() {
			node := buildNPUNode(MNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
				"2", "Ascend910-5,Ascend910-2"})
			result := npu.CheckNodeNPUByTaskFn(task, node, true)
			So(result, ShouldBeError)
		})
		Convey("CheckNodeNPUByTaskFn() should return error when node meets task requests", func() {
			node := buildNPUNode(MNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
				"4", "Ascend910-1,Ascend910-0,Ascend910-3,Ascend910-2"})
			result := npu.CheckNodeNPUByTaskFn(task, node, true)
			So(result, ShouldBeNil)
		})
	})
}

// TestMnpuGetNPUAffinityBestNodesFn
func TestMnpuGetNPUAffinityBestNodesFn(t *testing.T) {
	Convey("Test module910x8 GetNPUAffinityBestNodesFn", t, func() {
		const (
			nodeName1 = "centos1"
			nodeName2 = "centos2"
			constNum  = 0
		)
		npu := &module910x8{}
		pod := buildNPUPod(MPodInfo{namespace: "default", groupName: "npu-group-61",
			podName: "npu-test-61", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: npu800And9000CardName, reqNpuNum: "4"})
		task := vapi.NewTaskInfo(pod)
		nodes := []*vapi.NodeInfo{}

		Convey("GetNPUAffinityBestNodesFn() should return correct result", func() {
			node1 := buildNPUNode(MNodeInfo{nodeName1, huaweiArchX86, "192", "755Gi",
				"5", "Ascend910-5,Ascend910-2,Ascend910-0,Ascend910-1,Ascend910-3"})
			nodes = append(nodes, node1)
			node2 := buildNPUNode(MNodeInfo{nodeName2, huaweiArchX86, "192", "755Gi",
				"6", "Ascend910-5,Ascend910-4,Ascend910-6,Ascend910-7,Ascend910-0,Ascend910-1"})
			nodes = append(nodes, node2)

			result, err := npu.GetNPUAffinityBestNodesFn(task, nodes, false)

			So(result[nodeName1], ShouldEqual, constNum)
			So(err, ShouldBeNil)
		})
	})
}

// TestMnpuScoreBestNPUNodesFn
func TestMnpuScoreBestNPUNodesFn(t *testing.T) {
	Convey("Test module910x8 ScoreBestNPUNodesFn", t, func() {
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
		task := vapi.NewTaskInfo(pod)

		nodes := []*vapi.NodeInfo{}
		node1 := buildNPUNode(MNodeInfo{nodeName1, huaweiArchX86, "192", "755Gi",
			"8", "Ascend910-5,Ascend910-2,Ascend910-0,Ascend910-1,Ascend910-3," +
				"Ascend910-6,Ascend910-7"})
		nodes = append(nodes, node1)
		node2 := buildNPUNode(MNodeInfo{nodeName2, huaweiArchArm, "192", "755Gi",
			"8", "Ascend910-5,Ascend910-2,Ascend910-0,Ascend910-1,Ascend910-3," +
				"Ascend910-6,Ascend910-7"})
		nodes = append(nodes, node2)

		Convey("ScoreBestNPUNodesFn() should return err when scoreMap is nil", func() {
			var scoreMap map[string]float64
			_, err := npu.ScoreBestNPUNodesFn(scoreMap, bestNodes, task, nodes)
			So(err, ShouldNotBeNil)
		})

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

// TestMnpuGetAllocatedNPUFromTopologyFn
func TestMnpuGetAllocatedNPUFromTopologyFn(t *testing.T) {
	Convey("Test module910x8 GetAllocatedNPUFromTopologyFn", t, func() {
		npu := &module910x8{}

		Convey("GetAllocatedNPUFromTopologyFn()", func() {
			pod := buildNPUPod(MPodInfo{namespace: "default", groupName: "npu-group-64",
				podName: "npu-test-64", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: npu800And9000CardName, reqNpuNum: "8"})
			task := vapi.NewTaskInfo(pod)
			node := buildNPUNode(MNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
				"8", ""})
			setNodeAnnotation(node.Node, npu800And9000CardName, "")
			var expectedResult []int
			result, err := npu.GetAllocatedNPUFromTopologyFn(task, node, false)
			So(err, ShouldBeError)
			So(result, ShouldResemble, expectedResult)
		})
		Convey("GetAllocatedNPUFromTopologyFn() should return correct result case 1", func() {
			pod := buildNPUPod(MPodInfo{namespace: "default", groupName: "npu-group-63",
				podName: "npu-test-63", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: npu800And9000CardName, reqNpuNum: "4"})
			task := vapi.NewTaskInfo(pod)
			node := buildNPUNode(MNodeInfo{nodeName, huaweiArchArm, "192", "755Gi",
				"5", "Ascend910-5,Ascend910-2,Ascend910-0,Ascend910-1,Ascend910-3"})
			expectedResult := []int{2, 0, 1, 3}
			result, err := npu.GetAllocatedNPUFromTopologyFn(task, node, false)
			So(err, ShouldBeNil)
			So(result, ShouldResemble, expectedResult)
		})
		Convey("GetAllocatedNPUFromTopologyFn() should return correct result case 2", func() {
			pod := buildNPUPod(MPodInfo{namespace: "default", groupName: "npu-group-65",
				podName: "npu-test-65", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: npu800And9000CardName, reqNpuNum: "2"})
			task := vapi.NewTaskInfo(pod)
			node := buildNPUNode(MNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
				"2", "Ascend910-5,Ascend910-7"})
			expectedResult := []int{5, 7}
			result, err := npu.GetAllocatedNPUFromTopologyFn(task, node, false)
			So(err, ShouldBeNil)
			So(result, ShouldResemble, expectedResult)
		})
	})
}

// TestMnpuSetNPUTopologyToPodFn
func TestMnpuSetNPUTopologyToPodFn(t *testing.T) {
	Convey("Test module910x8 SetNPUTopologyToPodFn", t, func() {
		npu := &module910x8{}
		task := vapi.NewTaskInfo(buildNPUPod(MPodInfo{namespace: "default", groupName: "npu-group-64",
			podName: "npu-test-64", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: npu800And9000CardName, reqNpuNum: "4"}))
		Convey("SetNPUTopologyToPodFn() should return error when top is of wrong type", func() {
			top := []string{}
			err := npu.SetNPUTopologyToPodFn(task, top)
			So(err, ShouldBeError)
			So(task.Pod.Annotations[npu800And9000CardName], ShouldEqual, "")
		})
		Convey("SetNPUTopologyToPodFn() should write correct info in pod annotation", func() {
			top := []int{2, 0, 1, 3}
			expectedResult := "Ascend910-2,Ascend910-0,Ascend910-1,Ascend910-3"
			err := npu.SetNPUTopologyToPodFn(task, top)
			So(err, ShouldBeNil)
			So(task.Pod.Annotations[npu800And9000CardName], ShouldEqual, expectedResult)
		})
	})
}

// TestMnpuUpdateNPUNodeUsedCardFn
func TestMnpuUpdateNPUNodeUsedCardFn(t *testing.T) {
	Convey("Test module910x8 UpdateNPUNodeUsedCardFn", t, func() {
		npu := &module910x8{}
		node := buildNPUNode(MNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
			"4", "Ascend910-2,Ascend910-0,Ascend910-1,Ascend910-3"})
		node.Others = map[string]interface{}{
			npu800And9000CardName: "Ascend910-2,Ascend910-0,Ascend910-1,Ascend910-3",
		}
		Convey("UpdateNPUNodeUsedCardFn() should successfully update node.others", func() {
			top := []int{0, 1}
			expectedResult := map[string]interface{}{
				npu800And9000CardName: "Ascend910-2,Ascend910-3",
			}
			err := npu.UpdateNPUNodeUsedCardFn(node, top)
			So(err, ShouldBeNil)
			So(node.Others, ShouldResemble, expectedResult)
		})
	})
}

// TestMnpuGetReleaseNPUTopologyFn
func TestMnpuGetReleaseNPUTopologyFn(t *testing.T) {
	Convey("Test module910x8 GetReleaseNPUTopologyFn", t, func() {
		npu := &module910x8{}
		task := vapi.NewTaskInfo(buildNPUPod(MPodInfo{namespace: "default", groupName: "npu-group-64",
			podName: "npu-test-64", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: npu800And9000CardName, reqNpuNum: "4"}))
		Convey("GetReleaseNPUTopologyFn() should return correct card id slice", func() {
			task.Pod.Annotations[npu800And9000CardName] = "Ascend910-0,Ascend910-1,Ascend910-2,Ascend910-3,Ascend910-6"
			expectedResult := []int{0, 1, 2, 3, 6}
			result, err := npu.GetReleaseNPUTopologyFn(task)
			So(err, ShouldBeNil)
			So(result, ShouldResemble, expectedResult)
		})
	})
}

// TestMnpuUpdateReleaseNPUNodeTopologyFn
func TestMnpuUpdateReleaseNPUNodeTopologyFn(t *testing.T) {
	Convey("Test module910x8 UpdateReleaseNPUNodeTopologyFn", t, func() {
		npu := &module910x8{}
		node := buildNPUNode(MNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
			"4", "Ascend910-4,Ascend910-6,Ascend910-7,Ascend910-5"})
		node.Others = map[string]interface{}{
			npu800And9000CardName: "Ascend910-4,Ascend910-6,Ascend910-7,Ascend910-5",
		}
		Convey("UpdateNPUNodeUsedCardFn() should successfully update node.others", func() {
			top := []int{0, 1}
			err := npu.UpdateReleaseNPUNodeTopologyFn(node, top)
			So(err, ShouldBeNil)
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

func setJobResourceReq(MJob *vapi.JobInfo, resource string, num float64) {
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
	npuNum, err := ms.Atoi(npu)
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

func buildNPUNode(MNode MNodeInfo) *vapi.NodeInfo {
	nodeCapacity := buildNPUResourceList(MNode.cpu, MNode.mem, npu800And9000CardName, ms.Itoa(constIntNum2))
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

	node := vapi.NewNodeInfo(v1node)
	return node
}
