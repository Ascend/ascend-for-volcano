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

Package card910x2 is using for HuaWei A300T Ascend pin affinity schedule.

*/
package card910x2

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
	nodeName = "centos"
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

// TestCNPUIsMyTask
func TestCNPUIsMyTask(t *testing.T) {
	Convey("Test card910x2 IsMyTask", t, func() {
		npu := &card910x2{}

		Convey("IsMyTask() should return error when task doesn't request NPU", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-100",
				podName: "npu-test-100", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
				reqNPUType: a300TNPUCardName, reqNpuNum: "0"})
			task := vapi.NewTaskInfo(pod)
			result := npu.IsMyTask(task)
			So(result, ShouldBeError)
		})
		Convey("IsMyTask() should return error when task is of module mode", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-101",
				podName: "npu-test-101", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
				reqNPUType: a300TNPUCardName, reqNpuNum: "1"})
			setPodSelector(pod, acceleratorType, moduleAcceleratorType)
			task := vapi.NewTaskInfo(pod)
			result := npu.IsMyTask(task)
			So(result, ShouldBeError)
		})
		Convey("IsMyTask() should return nil when task is of card type", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-102",
				podName: "npu-test-102", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
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
				CPodInfo{namespace: "default", groupName: "npu-group-103",
					podName: "npu-test-103", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
					reqNPUType: a300TNPUCardName, reqNpuNum: "0"})))
			job := vapi.NewJobInfo(uid, tasks...)
			result := npu.IsMyJob(job)
			So(result, ShouldBeError)
		})
		Convey("IsMyJob() should return error when job has no selector", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-104",
				podName: "npu-test-104", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
				reqNPUType: a300TNPUCardName, reqNpuNum: "1"})
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
				reqNPUType: a300TNPUCardName, reqNpuNum: "1"})
			tasks = append(tasks, vapi.NewTaskInfo(pod))
			job := vapi.NewJobInfo(uid, tasks...)
			result := npu.IsMyJob(job)
			So(result, ShouldBeNil)
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
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-12",
				podName: "npu-test-35", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
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
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-13",
				podName: "npu-test-36", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
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
				CPodInfo{namespace: "default", groupName: "npu-group-17",
					podName: "npu-test-40", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
					reqNPUType: a300TNPUCardName, reqNpuNum: invalidNum0})))
			job := vapi.NewJobInfo(uid, tasks...)
			result := npu.ValidNPUJobFn(job)
			So(result, ShouldNotBeNil)
		})
		Convey("ValidNPUJobFn() should return error for job with invalid request number 5", func() {
			tasks = append(tasks, vapi.NewTaskInfo(buildNPUPod(
				CPodInfo{namespace: "default", groupName: "npu-group-18",
					podName: "npu-test-41", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
					reqNPUType: a300TNPUCardName, reqNpuNum: invalidNum5})))
			job := vapi.NewJobInfo(uid, tasks...)
			result := npu.ValidNPUJobFn(job)
			So(result, ShouldNotBeNil)
		})
		Convey("ValidNPUJobFn() should return error for job with valid request number 1", func() {
			tasks = append(tasks, vapi.NewTaskInfo(buildNPUPod(
				CPodInfo{namespace: "default", groupName: "npu-group-19",
					podName: "npu-test-42", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
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
				CPodInfo{namespace: "default", groupName: "npu-group-23",
					podName: "npu-test-46", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
					reqNPUType: a300TNPUCardName, reqNpuNum: num8})))
			job := vapi.NewJobInfo(uid, tasks...)
			result := npu.ValidNPUJobFn(job)
			So(result, ShouldNotBeNil)
		})
		Convey("ValidNPUJobFn() should return nil for job with valid model", func() {
			tasks = append(tasks, vapi.NewTaskInfo(buildNPUPod(
				CPodInfo{namespace: "default", groupName: "npu-group-24",
					podName: "npu-test-47", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
					reqNPUType: a300TNPUCardName, reqNpuNum: num2})))
			tasks = append(tasks, vapi.NewTaskInfo(buildNPUPod(
				CPodInfo{namespace: "default", groupName: "npu-group-25",
					podName: "npu-test-48", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
					reqNPUType: a300TNPUCardName, reqNpuNum: num2})))
			job := vapi.NewJobInfo(uid, tasks...)
			result := npu.ValidNPUJobFn(job)
			So(result, ShouldBeNil)
		})
	})
}

// TestCNPUPreCheckNodeFnTaskError
func TestCNPUPreCheckNodeFnTaskError(t *testing.T) {
	Convey("Test module910x8 PreCheckNodeFnTaskError", t, func() {
		npu := &card910x2{}
		confs := []conf.Configuration{}
		node := buildNPUNode(CNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
			"1", "Ascend910-1"})

		Convey("PreCheckNodeFn() should return error when task don't have selector", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-8",
				podName: "npu-test-54", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a300TNPUCardName, reqNpuNum: "1"})
			setPodSelector(pod, archSelector, "")
			setPodSelector(pod, acceleratorType, "")
			task := vapi.NewTaskInfo(pod)
			result := npu.PreCheckNodeFn(task, node, confs)
			So(result, ShouldBeError)
		})
		Convey("PreCheckNodeFn() should return nil when task isn't npu task", func() {
			pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-9",
				podName: "npu-test-55", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
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
	Convey("Test module910x8 PreCheckNodeFnNodeError", t, func() {
		npu := &card910x2{}
		confs := []conf.Configuration{}
		node := buildNPUNode(CNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
			"1", "Ascend910-3"})

		Convey("PreCheckNodeFn() should return error when node don't have label", func() {
			task := vapi.NewTaskInfo(buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-10",
				podName: "npu-test-56", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a300TNPUCardName, reqNpuNum: "1"}))
			setNodeLabel(node.Node, archSelector, "")
			setNodeLabel(node.Node, acceleratorType, "")
			result := npu.PreCheckNodeFn(task, node, confs)
			So(result, ShouldBeError)
		})
		Convey("PreCheckNodeFn() should return error when selectors arch mismatch with labels", func() {
			task := vapi.NewTaskInfo(buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-11",
				podName: "npu-test-57", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: a300TNPUCardName, reqNpuNum: "1"}))
			// build a node with mismatch selector
			nodeArm := buildNPUNode(CNodeInfo{nodeName, huaweiArchArm, "192", "755Gi",
				"1", "Ascend910-5"})
			result := npu.PreCheckNodeFn(task, nodeArm, confs)
			So(result, ShouldBeError)
		})
		Convey("PreCheckNodeFn() should return error when selectors model mismatch with labels", func() {
			task := vapi.NewTaskInfo(buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-11",
				podName: "npu-test-58", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
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
	Convey("Test module910x8 PreCheckNodeFnSuccess", t, func() {
		npu := &card910x2{}
		confs := []conf.Configuration{}
		node := buildNPUNode(CNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
			"1", "Ascend910-6"})

		Convey("PreCheckNodeFn() should return nil when selectors match with labels", func() {
			// build a task with no selector
			task := vapi.NewTaskInfo(buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-12",
				podName: "npu-test-59", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
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
			node.Node.Annotations[a300TNPUCardName] = "Ascend910-2,Ascend910-8"
			result := vnpu.CheckNPUResourceStableFn(node)
			So(result, ShouldBeError)
		})
		Convey("CheckNPUResourceStableFn() should return nil when node resources are stable", func() {
			node := buildNPUNode(CNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
				"2", "Ascend910-4,Ascend910-7"})
			result := vnpu.CheckNPUResourceStableFn(node)
			So(result, ShouldBeNil)
		})
	})
}

// TestCnpuCheckNodeNPUByTaskFn
func TestCnpuCheckNodeNPUByTaskFn(t *testing.T) {
	Convey("Test job CheckNodeNPUByTaskFn", t, func() {
		npu := &card910x2{}
		pod := buildNPUPod(CPodInfo{namespace: "default", groupName: "npu-group-13",
			podName: "npu-test-61", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: a300TNPUCardName, reqNpuNum: "2"})
		task := vapi.NewTaskInfo(pod)

		Convey("CheckNodeNPUByTaskFn() should return error when node doesn't meet task requests", func() {
			node := buildNPUNode(CNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
				"1", "Ascend910-4"})
			result := npu.CheckNodeNPUByTaskFn(task, node)
			So(result, ShouldBeError)
		})
		Convey("CheckNodeNPUByTaskFn() should return error when node meets task requests", func() {
			node := buildNPUNode(CNodeInfo{nodeName, huaweiArchX86, "192", "755Gi",
				"2", "Ascend910-1,Ascend910-4"})
			result := npu.CheckNodeNPUByTaskFn(task, node)
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
	return node
}
