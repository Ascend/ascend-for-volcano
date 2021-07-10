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

Package commonv910 is using for virtual HuaWei Ascend910 schedule.

*/
package commonv910

import (
	. "github.com/smartystreets/goconvey/convey"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
	"testing"
	vapi "volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/conf"
	"volcano.sh/volcano/pkg/scheduler/util"
)

const (
	nodeName     = "ubuntu"
	labelSize    = 8
	annSize      = 8
	constIntNum2 = 2
)

type VNodeInfo struct {
	nodeName       string
	nodeArch       string
	cpu, mem       string
	npuAllocateNum string
	npuTop         string
}

type VPodInfo struct {
	namespace  string
	groupName  string
	podName    string
	nodeName   string
	reqCPUNum  string
	reqMem     string
	reqNPUType string
	reqNpuNum  string
}

// TestGetVnpuType
func TestGetVnpuType(t *testing.T) {
	Convey("Test Get Vnpu Type", t, func() {
		Convey("should return a list of virtual npu types", func() {
			vTypes := []string{"huawei.com/Ascend910-2c", "huawei.com/Ascend910-4c",
				"huawei.com/Ascend910-8c", "huawei.com/Ascend910-16c"}
			n := GetVnpuType()
			So(n, ShouldResemble, vTypes)
		})
	})
}

// TestVnpuIsMyTask
func TestVnpuIsMyTask(t *testing.T) {
	Convey("Test IsMyTask in commonv910", t, func() {
		vnpu := &Vnpu{}
		Convey("IsMyTask() should return error since it's not implemented", func() {
			pod := buildNPUPod(VPodInfo{namespace: "default", groupName: "npu-4p", podName: "npu-4p",
				nodeName: nodeName, reqCPUNum: "10", reqMem: "20Gi", reqNPUType: npuV910CardName16c, reqNpuNum: "4"})
			task := vapi.NewTaskInfo(pod)
			result := vnpu.IsMyTask(task)
			So(result, ShouldBeError)
		})
	})
}

// TestVnpuIsMyNode
func TestVnpuIsMyNode(t *testing.T) {
	Convey("Test IsMyNode in commonv910", t, func() {
		vnpu := &Vnpu{}
		Convey("IsMyNode() should return error since it's not implemented", func() {
			node := buildNPUNode(VNodeInfo{nodeName, huaweiArchArm, "192", "755Gi",
				"1", "Ascend910-16c-118-0"})
			result := vnpu.IsMyNode(node)
			So(result, ShouldBeError)
		})
	})
}

// TestVnpuIsMyJob
func TestVnpuIsMyJob(t *testing.T) {
	Convey("Test IsMyJob in commonv910", t, func() {
		vnpu := &Vnpu{}
		tasks := []*vapi.TaskInfo{}
		uid := vapi.JobID("xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx1")

		Convey("IsMyJob() should return error since it's not implemented", func() {
			tasks = append(tasks, vapi.NewTaskInfo(
				buildNPUPod(VPodInfo{namespace: "default", groupName: "npu-group-0",
					podName: "npu-test-1", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
					reqNPUType: npuV910CardName16c, reqNpuNum: "1"})))
			tasks = append(tasks, vapi.NewTaskInfo(
				buildNPUPod(VPodInfo{namespace: "default", groupName: "npu-group-1",
					podName: "npu-test-2", nodeName: nodeName, reqCPUNum: "10", reqMem: "5Gi",
					reqNPUType: npuV910CardName16c, reqNpuNum: "2"})))
			job := vapi.NewJobInfo(uid, tasks...)
			result := vnpu.IsMyJob(job)
			So(result, ShouldBeError)
		})
	})
}

// TestVnpuName
func TestVnpuName(t *testing.T) {
	Convey("Test Vnpu Name", t, func() {
		const (
			pluginName = "Vnpu"
		)
		vnpu := &Vnpu{}
		Convey("Name() should return PluginName defined in const", func() {
			n := vnpu.Name()
			So(n, ShouldEqual, pluginName)
		})
	})
}

// TestVnpuValidNPUJobFnInvalidSelector
func TestVnpuValidNPUJobFnInvalidSelector(t *testing.T) {
	Convey("Test job ValidNPUJobFn InvalidSelector", t, func() {
		const (
			invalidSelectorKey   = "invalid-key"
			invalidSelectorValue = "no-selector"
			validNum             = 1000
		)
		vnpu := &Vnpu{}
		tasks := []*vapi.TaskInfo{}
		uid := vapi.JobID("xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx2")

		Convey("ValidNPUJobFn() should return error for job without certain selector key", func() {
			pod := buildNPUPod(VPodInfo{namespace: "default", groupName: "npu-group-2",
				podName: "npu-test-5", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: npuV910CardName16c, reqNpuNum: "1"})
			delete(pod.Spec.NodeSelector, archSelector)
			setPodSelector(pod, invalidSelectorKey, invalidSelectorValue)
			tasks = append(tasks, vapi.NewTaskInfo(pod))
			job := vapi.NewJobInfo(uid, tasks...)
			setJobResourceReq(job, npuV910CardName16c, float64(validNum))
			result := vnpu.ValidNPUJobFn(job)
			So(result, ShouldNotBeNil)
		})
		Convey("ValidNPUJobFn() should return error for job with invalid selector value", func() {
			pod := buildNPUPod(VPodInfo{namespace: "default", groupName: "npu-group-3",
				podName: "npu-test-6", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: npuV910CardName16c, reqNpuNum: "1"})
			setPodSelector(pod, archSelector, invalidSelectorValue)
			tasks = append(tasks, vapi.NewTaskInfo(pod))
			job := vapi.NewJobInfo(uid, tasks...)
			setJobResourceReq(job, npuV910CardName16c, float64(validNum))
			result := vnpu.ValidNPUJobFn(job)
			So(result, ShouldNotBeNil)
		})
	})
}

// TestVnpuValidNPUJobFnInvalidReq
func TestVnpuValidNPUJobFnInvalidReq(t *testing.T) {
	Convey("Test job ValidNPUJobFn InvalidReq", t, func() {
		const (
			invalidResourceType  = "huawei.com/Ascend910-5c"
			noPrefixResourceType = "no-prefix"
			validNum             = 1000
			invalidNum           = 2000
		)
		vnpu := &Vnpu{}
		tasks := []*vapi.TaskInfo{}
		uid := vapi.JobID("xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx3")

		Convey("ValidNPUJobFn() should return error for job with invalid resource prefix", func() {
			pod := buildNPUPod(VPodInfo{namespace: "default", groupName: "npu-group-4",
				podName: "npu-test-7", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: npuV910CardName16c, reqNpuNum: "1"})
			setPodSelector(pod, archSelector, huaweiArchArm)
			tasks = append(tasks, vapi.NewTaskInfo(pod))
			job := vapi.NewJobInfo(uid, tasks...)
			setJobResourceReq(job, noPrefixResourceType, float64(validNum))
			result := vnpu.ValidNPUJobFn(job)
			So(result, ShouldBeNil)
		})
		Convey("ValidNPUJobFn() should return error for job with invalid npu request type", func() {
			pod := buildNPUPod(VPodInfo{namespace: "default", groupName: "npu-group-5",
				podName: "npu-test-8", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: npuV910CardName16c, reqNpuNum: "1"})
			setPodSelector(pod, archSelector, huaweiArchX86)
			tasks = append(tasks, vapi.NewTaskInfo(pod))
			job := vapi.NewJobInfo(uid, tasks...)
			setJobResourceReq(job, invalidResourceType, float64(validNum))
			result := vnpu.ValidNPUJobFn(job)
			So(result, ShouldNotBeNil)
		})
		Convey("ValidNPUJobFn() should return error for job with invalid npu request num", func() {
			pod := buildNPUPod(VPodInfo{namespace: "default", groupName: "npu-group-6",
				podName: "npu-test-9", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: npuV910CardName16c, reqNpuNum: "1"})
			setPodSelector(pod, archSelector, huaweiArchX86)
			tasks = append(tasks, vapi.NewTaskInfo(pod))
			job := vapi.NewJobInfo(uid, tasks...)
			setJobResourceReq(job, npuV910CardName16c, float64(invalidNum))
			result := vnpu.ValidNPUJobFn(job)
			So(result, ShouldNotBeNil)
		})
	})
}

// TestVnpuValidNPUJobFnSuccess
func TestVnpuValidNPUJobFnSuccess(t *testing.T) {
	Convey("Test job ValidNPUJobFn Success", t, func() {
		const (
			validNum = 1000
		)
		vnpu := &Vnpu{}
		tasks := []*vapi.TaskInfo{}
		uid := vapi.JobID("xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx4")

		Convey("ValidNPUJobFn() should return nil for job with valid selectors", func() {
			tasks = append(tasks, vapi.NewTaskInfo(buildNPUPod(
				VPodInfo{namespace: "default", groupName: "npu-group-7",
					podName: "npu-test-3", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
					reqNPUType: npuV910CardName16c, reqNpuNum: "1"})))
			job := vapi.NewJobInfo(uid, tasks...)
			setJobResourceReq(job, npuV910CardName16c, float64(validNum))
			result := vnpu.ValidNPUJobFn(job)
			So(result, ShouldBeNil)
		})
	})
}

// TestVnpuPreCheckNodeFnTaskError
func TestVnpuPreCheckNodeFnTaskError(t *testing.T) {
	Convey("Test job PreCheckNodeFn TaskError", t, func() {
		vnpu := &Vnpu{}
		confs := []conf.Configuration{}
		node := buildNPUNode(VNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
			"1", "Ascend910-16c-131-1"})

		Convey("PreCheckNodeFn() should return error when task don't have selector", func() {
			pod := buildNPUPod(VPodInfo{namespace: "default", groupName: "npu-group-8",
				podName: "npu-test-10", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: npuV910CardName16c, reqNpuNum: "1"})
			delete(pod.Spec.NodeSelector, archSelector)
			task := vapi.NewTaskInfo(pod)
			result := vnpu.PreCheckNodeFn(task, node, confs)
			So(result, ShouldBeError)
		})
		Convey("PreCheckNodeFn() should return nil when task don't have selector and no resource requested", func() {
			pod := buildNPUPod(VPodInfo{namespace: "default", groupName: "npu-group-9",
				podName: "npu-test-11'", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: "", reqNpuNum: ""})
			delete(pod.Spec.NodeSelector, archSelector)
			task := vapi.NewTaskInfo(pod)
			result := vnpu.PreCheckNodeFn(task, node, confs)
			So(result, ShouldBeNil)
		})
	})
}

// TestVnpuPreCheckNodeFnNodeError
func TestVnpuPreCheckNodeFnNodeError(t *testing.T) {
	Convey("Test job PreCheckNodeFn NodeError", t, func() {
		vnpu := &Vnpu{}
		confs := []conf.Configuration{}
		node := buildNPUNode(VNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
			"1", "Ascend910-16c-119-1"})

		Convey("PreCheckNodeFn() should return error when node don't have label", func() {
			task := vapi.NewTaskInfo(buildNPUPod(VPodInfo{namespace: "default", groupName: "npu-group-10",
				podName: "npu-test-12", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: npuV910CardName16c, reqNpuNum: "1"}))
			setNodeLabel(node.Node, archSelector, "")
			result := vnpu.PreCheckNodeFn(task, node, confs)
			So(result, ShouldBeError)
		})
		Convey("PreCheckNodeFn() should return error when selectors mismatch with labels", func() {
			task := vapi.NewTaskInfo(buildNPUPod(VPodInfo{namespace: "default", groupName: "npu-group-11",
				podName: "npu-test-13", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: npuV910CardName16c, reqNpuNum: "1"}))
			// build a node with mismatch selector
			nodeArm := buildNPUNode(VNodeInfo{nodeName, huaweiArchArm, "192", "755Gi",
				"2", "Ascend910-16c-140-1,Ascend910-16c-141-1"})
			result := vnpu.PreCheckNodeFn(task, nodeArm, confs)
			So(result, ShouldBeError)
		})
	})
}

// TestVnpuPreCheckNodeFnSuccess
func TestVnpuPreCheckNodeFnSuccess(t *testing.T) {
	Convey("Test job PreCheckNodeFn Success", t, func() {
		vnpu := &Vnpu{}
		confs := []conf.Configuration{}
		node := buildNPUNode(VNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
			"1", "Ascend910-16c-132-1"})

		Convey("PreCheckNodeFn() should return nil when selectors match with labels", func() {
			// build a task with no selector
			task := vapi.NewTaskInfo(buildNPUPod(VPodInfo{namespace: "default", groupName: "npu-group-12",
				podName: "npu-test-14", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: npuV910CardName16c, reqNpuNum: "1"}))
			result := vnpu.PreCheckNodeFn(task, node, confs)
			So(result, ShouldBeNil)
		})
	})
}

// TestVnpuCheckNPUResourceStableFn
func TestVnpuCheckNPUResourceStableFn(t *testing.T) {
	Convey("Test job CheckNPUResourceStableFn", t, func() {
		vnpu := &Vnpu{}

		Convey("CheckNPUResourceStableFn() should return error when there's missing resource type in idle", func() {
			// build a node with 2 Ascend-910-16c in annotation and no idle Ascend-910-16c
			node := buildNPUNode(VNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
				"0", ""})
			node.Node.Annotations[npuV910CardName16c] = "Ascend910-16c-142-1,Ascend910-16c-143-1"
			result := vnpu.CheckNPUResourceStableFn(node)
			So(result, ShouldBeError)
		})
		Convey("CheckNPUResourceStableFn() should return error when node resources are unstable", func() {
			// build a node with 2 Ascend-910-16c in annotation and 1 idle Ascend-910-16c
			node := buildNPUNode(VNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
				"1", ""})
			node.Node.Annotations[npuV910CardName16c] = "Ascend910-16c-144-1,Ascend910-16c-145-1"
			result := vnpu.CheckNPUResourceStableFn(node)
			So(result, ShouldBeError)
		})
		Convey("CheckNPUResourceStableFn() should return nil when node resources are stable", func() {
			// build a node with 2 Ascend-910-16c in annotation and 1 idle Ascend-910-16c
			node := buildNPUNode(VNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
				"2", "Ascend910-16c-146-1,Ascend910-16c-147-1"})
			result := vnpu.CheckNPUResourceStableFn(node)
			So(result, ShouldBeNil)
		})
	})
}

// TestVnpuCheckNodeNPUByTaskFn
func TestVnpuCheckNodeNPUByTaskFn(t *testing.T) {
	Convey("Test job CheckNodeNPUByTaskFn", t, func() {
		vnpu := &Vnpu{}
		pod := buildNPUPod(VPodInfo{namespace: "default", groupName: "npu-group-13",
			podName: "npu-test-15", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: npuV910CardName16c, reqNpuNum: "1"})
		task := vapi.NewTaskInfo(pod)

		Convey("CheckNodeNPUByTaskFn() should return error when node doesn't meet task requests", func() {
			node := buildNPUNode(VNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
				"0", ""})
			result := vnpu.CheckNodeNPUByTaskFn(task, node)
			So(result, ShouldBeError)
		})
		Convey("CheckNodeNPUByTaskFn() should return error when node meets task requests", func() {
			node := buildNPUNode(VNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
				"1", "Ascend910-16c-199-1"})
			result := vnpu.CheckNodeNPUByTaskFn(task, node)
			So(result, ShouldBeNil)
		})
	})
}

// TestVnpuGetNPUAffinityBestNodesFn
func TestVnpuGetNPUAffinityBestNodesFn(t *testing.T) {
	Convey("Test GetNPUAffinityBestNodesFn", t, func() {
		const (
			nodeNameCase1 = "node1"
			nodeNameCase2 = "node2"
			nodeNameCase3 = "node3"
		)
		bestNodesMap := map[string]int{
			nodeNameCase1: 0,
			nodeNameCase2: 0,
			nodeNameCase3: 0,
		}

		vnpu := &Vnpu{}
		pod := buildNPUPod(VPodInfo{namespace: "default", groupName: "npu-group-14",
			podName: "npu-test-16", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
			reqNPUType: npuV910CardName16c, reqNpuNum: "1"})
		task := vapi.NewTaskInfo(pod)
		node1 := buildNPUNode(VNodeInfo{nodeNameCase1, huaweiArchX86, "192", "755Gi",
			"3", "Ascend910-16c-110-1,Ascend910-16c-111-1,Ascend910-16c-112-1"})
		node2 := buildNPUNode(VNodeInfo{nodeNameCase2, huaweiArchArm, "192", "755Gi",
			"3", "Ascend910-16c-113-1,Ascend910-16c-114-1,Ascend910-16c-115-1"})
		node3 := buildNPUNode(VNodeInfo{nodeNameCase3, huaweiArchArm, "192", "755Gi",
			"3", "Ascend910-16c-116-1,Ascend910-16c-117-1,Ascend910-16c-118-1"})

		Convey("GetNPUAffinityBestNodesFn() should return correct bestNodesMap", func() {
			nodes := []*vapi.NodeInfo{node1, node2, node3, nil}
			result, err := vnpu.GetNPUAffinityBestNodesFn(task, nodes)
			So(result, ShouldResemble, bestNodesMap)
			So(err, ShouldBeNil)
		})
	})
}

// TestVnpuScoreBestNPUNodesFn
func TestVnpuScoreBestNPUNodesFn(t *testing.T) {
	Convey("Test ScoreBestNPUNodesFn", t, func() {
		vnpu := &Vnpu{}
		const (
			nodeNameCase1 = "node1"
			nodeNameCase2 = "node2"
		)
		bestNodesMap := map[string]int{
			nodeNameCase1: 0,
			nodeNameCase2: 0,
		}
		node1 := buildNPUNode(VNodeInfo{nodeNameCase1, huaweiArchX86, "192", "755Gi",
			"3", "Ascend910-16c-103-1,Ascend910-16c-104-1,Ascend910-16c-105-1"})
		node2 := buildNPUNode(VNodeInfo{nodeNameCase2, huaweiArchArm, "192", "755Gi",
			"3", "Ascend910-16c-106-1,Ascend910-16c-107-1,Ascend910-16c-108-1"})

		Convey("ScoreBestNPUNodesFn() should return correct scoreMap", func() {
			scoreMap := map[string]float64{
				nodeNameCase1: 3.0,
			}
			expectResult := map[string]float64{
				nodeNameCase1: 0.0,
				nodeNameCase2: 0.0,
			}
			task := vapi.TaskInfo{}
			nodes := []*vapi.NodeInfo{node1, node2}
			result, err := vnpu.ScoreBestNPUNodesFn(scoreMap, bestNodesMap, &task, nodes)
			So(result, ShouldResemble, expectResult)
			So(err, ShouldBeNil)
		})
	})
}

// TestVnpuGetAllocatedNPUFromTopologyFn
func TestVnpuGetAllocatedNPUFromTopologyFn(t *testing.T) {
	Convey("Test GetAllocatedNPUFromTopologyFn", t, func() {
		vnpu := &Vnpu{}
		var expectResult []string
		pod := buildNPUPod(VPodInfo{namespace: "default", groupName: "npu-group-15",
			podName: "npu-test-17", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
			reqNPUType: npuV910CardName16c, reqNpuNum: "1"})
		task := vapi.NewTaskInfo(pod)

		Convey("GetAllocatedNPUFromTopologyFn() should return name of the allocated devices", func() {
			node := buildNPUNode(VNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
				"1", "Ascend910-16c-133-1"})
			expectResult = []string{"Ascend910-16c-133-1"}
			result, err := vnpu.GetAllocatedNPUFromTopologyFn(task, node)
			So(result, ShouldResemble, expectResult)
			So(err, ShouldBeNil)
		})
		Convey("GetAllocatedNPUFromTopologyFn() should return error when node annotation is empty", func() {
			node := buildNPUNode(VNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
				"1", ""})
			result, err := vnpu.GetAllocatedNPUFromTopologyFn(task, node)
			So(result, ShouldResemble, expectResult)
			So(err, ShouldBeError)
		})
		Convey("GetAllocatedNPUFromTopologyFn() should return error when task request no vNPU", func() {
			podNoReq := buildNPUPod(VPodInfo{namespace: "default", groupName: "npu-group-16",
				podName: "npu-test-18", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
				reqNPUType: npuV910CardName16c, reqNpuNum: "0"})
			taskNoReq := vapi.NewTaskInfo(podNoReq)
			node := buildNPUNode(VNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
				"1", ""})
			result, err := vnpu.GetAllocatedNPUFromTopologyFn(taskNoReq, node)
			So(result, ShouldResemble, expectResult)
			So(err, ShouldBeError)
		})
	})
}

// TestVnpuSetNPUTopologyToPodFn
func TestVnpuSetNPUTopologyToPodFn(t *testing.T) {
	Convey("Test SetNPUTopologyToPodFn", t, func() {
		vnpu := &Vnpu{}
		pod := buildNPUPod(VPodInfo{namespace: "default", groupName: "npu-group-17",
			podName: "npu-test-19", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
			reqNPUType: npuV910CardName16c, reqNpuNum: "1"})
		task := vapi.NewTaskInfo(pod)

		Convey("SetNPUTopologyToPodFn() should return error when top isn't of slice of string", func() {
			top := []int{}
			result := vnpu.SetNPUTopologyToPodFn(task, top)
			So(result, ShouldBeError)
		})
		Convey("SetNPUTopologyToPodFn() should write top into pod annotation", func() {
			top := []string{"Ascend910-2c-190-1"}
			result := vnpu.SetNPUTopologyToPodFn(task, top)
			So(result, ShouldBeNil)
			topFromPod := task.Pod.Annotations[npuV910CardName2c]
			So(topFromPod, ShouldEqual, top[0])
		})
	})
}

// TestVnpuUpdateNPUNodeUsedCardFn
func TestVnpuUpdateNPUNodeUsedCardFn(t *testing.T) {
	Convey("Test UpdateNPUNodeUsedCardFn", t, func() {
		vnpu := &Vnpu{}
		node := buildNPUNode(VNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
			"3", "Ascend910-16c-109-1,Ascend910-16c-110-1,Ascend910-16c-111-1"})

		Convey("UpdateNPUNodeUsedCardFn() should return error when top isn't of slice of string", func() {
			top := []int{}
			result := vnpu.UpdateNPUNodeUsedCardFn(node, top)
			So(result, ShouldBeError)
		})
		Convey("UpdateNPUNodeUsedCardFn() should return error when no certain vnpu type in node.Others", func() {
			top := []string{"Ascend910-16c-113-0"}
			result := vnpu.UpdateNPUNodeUsedCardFn(node, top)
			So(result, ShouldBeError)
		})
		Convey("UpdateNPUNodeUsedCardFn() should return error when node.Others has invalid content", func() {
			node.Others = map[string]interface{}{npuV910CardName16c: 0}
			top := []string{"Ascend910-16c-112-0"}
			result := vnpu.UpdateNPUNodeUsedCardFn(node, top)
			So(result, ShouldBeError)
		})
		Convey("UpdateNPUNodeUsedCardFn() should successfully update node.Others", func() {
			node.Others = map[string]interface{}{
				npuV910CardName16c: "Ascend910-16c-111-0,Ascend910-16c-112-0",
			}
			top := []string{"Ascend910-16c-111-0"}
			result := vnpu.UpdateNPUNodeUsedCardFn(node, top)
			So(result, ShouldBeNil)
			expectResult := "Ascend910-16c-112-0"
			So(node.Others[npuV910CardName16c], ShouldResemble, expectResult)
		})
	})
}

// TestVnpuGetReleaseNPUTopologyFn
func TestVnpuGetReleaseNPUTopologyFn(t *testing.T) {
	Convey("Test GetReleaseNPUTopologyFn", t, func() {
		vnpu := &Vnpu{}
		var expectResult []string

		pod := buildNPUPod(VPodInfo{namespace: "default", groupName: "npu-group-18",
			podName: "npu-test-20", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
			reqNPUType: npuV910CardName16c, reqNpuNum: "1"})
		task := vapi.NewTaskInfo(pod)
		Convey("GetReleaseNPUTopologyFn() should return error when pod has no vnpu annotation", func() {
			taskTopArr, err := vnpu.GetReleaseNPUTopologyFn(task)
			So(err, ShouldBeError)
			So(taskTopArr, ShouldResemble, expectResult)
		})
		Convey("GetReleaseNPUTopologyFn() should return error when pod has empty annotation", func() {
			task.Pod.Annotations[npuV910CardName16c] = ""
			taskTopArr, err := vnpu.GetReleaseNPUTopologyFn(task)
			So(err, ShouldBeError)
			So(taskTopArr, ShouldResemble, expectResult)
		})
		Convey("GetReleaseNPUTopologyFn() should return nil when pod has correct annotation", func() {
			topStr := "Ascend910-2c-190-1"
			expectResult = append(expectResult, topStr)
			task.Pod.Annotations[npuV910CardName16c] = topStr
			taskTopArr, err := vnpu.GetReleaseNPUTopologyFn(task)
			So(err, ShouldBeNil)
			So(taskTopArr, ShouldResemble, expectResult)
		})
	})
}

// TestVnpuUpdateReleaseNPUNodeTopologyFn
func TestVnpuUpdateReleaseNPUNodeTopologyFn(t *testing.T) {
	Convey("Test UpdateReleaseNPUNodeTopologyFn", t, func() {
		vnpu := &Vnpu{}
		node := buildNPUNode(VNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
			"3", "Ascend910-16c-112-1,Ascend910-16c-113-1,Ascend910-16c-114-1"})
		Convey("UpdateReleaseNPUNodeTopologyFn() should return error when content of top is invalid", func() {
			top := []string{"Ascend910-5c-111-0"}
			result := vnpu.UpdateReleaseNPUNodeTopologyFn(node, top)
			So(result, ShouldBeError)
		})
		Convey("UpdateReleaseNPUNodeTopologyFn() should successfully update node.Others", func() {
			node.Others = map[string]interface{}{
				npuV910CardName16c: "Ascend910-16c-111-0,Ascend910-16c-112-0,Ascend910-16c-113-0",
			}
			top := []string{"Ascend910-16c-114-0"}
			result := vnpu.UpdateReleaseNPUNodeTopologyFn(node, top)
			So(result, ShouldBeNil)
		})
	})
}

// TestVnpuinitVNodesFn
func TestVnpuinitVNodesFn(t *testing.T) {
	Convey("Test initVNodesFn", t, func() {
		Convey("", func() {
			node1 := buildNPUNode(VNodeInfo{"ubuntu-1", huaweiArchX86, "192", "755Gi",
				"1", "Ascend910-16c-117-0"})
			node2 := buildNPUNode(VNodeInfo{"ubuntu-2", huaweiArchArm, "192", "755Gi",
				"1", "Ascend910-16c-118-0"})
			node3 := buildNPUNode(VNodeInfo{"ubuntu-3", huaweiArchArm, "192", "755Gi",
				"1", "Ascend910-16c-119-0"})
			setNodeAnnotation(node3.Node, npuV910CardName16c, "")
			nodes := map[string]*vapi.NodeInfo{
				"1": node1,
				"2": node2,
				"3": node3,
			}
			result := initVNodesFn(nodes)
			So(result, ShouldBeNil)
		})
	})
}

func setPodSelector(pod *v1.Pod, selectorKey string, selectorValue string) {
	pod.Spec.NodeSelector[selectorKey] = selectorValue
}

func setNodeLabel(node *v1.Node, labelKey string, labelValue string) {
	if labelValue == "" {
		delete(node.Labels, labelKey)
		return
	}
	node.Labels[labelKey] = labelValue
}

func setNodeAnnotation(node *v1.Node, annKey string, annValue string) {
	if annValue == "" {
		delete(node.Annotations, annKey)
		return
	}
	node.Annotations[annKey] = annValue
}

func setJobResourceReq(job *vapi.JobInfo, resource string, num float64) {
	job.TotalRequest.ScalarResources[v1.ResourceName(resource)] = num
}

func buildNPUPod(podInfo VPodInfo) *v1.Pod {

	pod := util.BuildPod(podInfo.namespace, podInfo.podName, podInfo.nodeName, v1.PodPending,
		buildNPUResourceList(podInfo.reqCPUNum, podInfo.reqMem,
			v1.ResourceName(podInfo.reqNPUType), podInfo.reqNpuNum),
		podInfo.groupName, make(map[string]string, constIntNum2),
		make(map[string]string, constIntNum2))

	setPodSelector(pod, archSelector, huaweiArchX86)

	return pod
}

func buildNPUResourceList(cpu string, memory string, npuResourceType v1.ResourceName, npu string) v1.ResourceList {
	npuNum, err := strconv.Atoi(npu)
	if err != nil {
		return nil
	}

	if npuNum == 0 {
		return v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse(cpu),
			v1.ResourceMemory: resource.MustParse(memory),
		}
	}

	return v1.ResourceList{
		v1.ResourceCPU:    resource.MustParse(cpu),
		v1.ResourceMemory: resource.MustParse(memory),
		npuResourceType:   resource.MustParse(npu),
	}
}

func buildNPUNode(nodeInfo VNodeInfo) *vapi.NodeInfo {
	nodeCapacity := buildNPUResourceList(nodeInfo.cpu, nodeInfo.mem, npuV910CardName16c, strconv.Itoa(constIntNum2))
	nodeAlloc := buildNPUResourceList(nodeInfo.cpu, nodeInfo.mem, npuV910CardName16c, nodeInfo.npuAllocateNum)
	labels := make(map[string]string, labelSize)
	ann := make(map[string]string, annSize)

	v1node := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:        nodeInfo.nodeName,
			Labels:      labels,
			Annotations: ann,
		},
		Status: v1.NodeStatus{
			Capacity:    nodeCapacity,
			Allocatable: nodeAlloc,
		},
	}

	if nodeInfo.npuAllocateNum != "0" {
		v1node.Annotations[npuV910CardName16c] = nodeInfo.npuTop
	}

	setNodeLabel(v1node, archSelector, nodeInfo.nodeArch)

	node := vapi.NewNodeInfo(v1node)
	return node
}
