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

Package modulev910x8 is using for virtual HuaWei Ascend910 schedule.

*/
package modulev910x8

import (
	. "github.com/smartystreets/goconvey/convey"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	mmetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
	"testing"
	vapi "volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/util"
)

const (
	archSelector          = "host-arch"
	huaweiArchArm         = "huawei-arm"
	huaweiArchX86         = "huawei-x86"
	acceleratorType       = "accelerator-type"
	cardAcceleratorType   = "card"
	moduleAcceleratorType = "module"

	nodeName           = "ubuntu"
	constIntNum2       = 2
	labelSize          = 8
	annSize            = 8
	npuV910CardName16c = "huawei.com/Ascend910-16c"
)

type VMNodeInfo struct {
	nodeName       string
	nodeArch       string
	cpu, mem       string
	npuAllocateNum string
	npuTop         string
}

type VMPodInfo struct {
	namespace  string
	groupName  string
	podName    string
	nodeName   string
	reqCPUNum  string
	reqMem     string
	reqNPUType string
	reqNpuNum  string
}

// TestMVnpuName
func TestMVnpuName(t *testing.T) {
	Convey("Test modulev910x8 Name", t, func() {
		const (
			pluginName = "A800-9000-VNpu"
		)
		vnpu := &modulev910x8{}

		Convey("Name() should return PluginName defined in const", func() {
			n := vnpu.Name()
			So(n, ShouldEqual, pluginName)
		})
	})
}

// TestMVnpuNew
func TestMVnpuNew(t *testing.T) {
	Convey("Test modulev910x8 New", t, func() {
		const (
			name = "modulev910x8"
		)
		expectResult := &modulev910x8{name: name}

		Convey("Name() should return a modulev910x8 instance", func() {
			result := New(name)
			So(result, ShouldResemble, expectResult)
		})
	})
}

// TestMVnpuIsMyTask
func TestMVnpuIsMyTask(t *testing.T) {
	Convey("Test IsMyTask", t, func() {
		vnpu := &modulev910x8{}

		Convey("IsMyTask() should return error when task doesn't request vNPU", func() {
			pod := buildNPUPod(VMPodInfo{namespace: "default", groupName: "npu-group-0",
				podName: "npu-test-28", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
				reqNPUType: npuV910CardName16c, reqNpuNum: "0"})
			task := vapi.NewTaskInfo(pod)
			result := vnpu.IsMyTask(task)
			So(result, ShouldBeError)
		})
		Convey("IsMyTask() should return nil when needed selector is empty", func() {
			pod := buildNPUPod(VMPodInfo{namespace: "default", groupName: "npu-group-1",
				podName: "npu-test-29", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
				reqNPUType: npuV910CardName16c, reqNpuNum: "1"})
			pod.Spec.NodeSelector = nil
			setPodSelector(pod, acceleratorType, "")
			task := vapi.NewTaskInfo(pod)
			result := vnpu.IsMyTask(task)
			So(result, ShouldBeNil)
		})
		Convey("IsMyTask() should return nil when task is of module type", func() {
			pod := buildNPUPod(VMPodInfo{namespace: "default", groupName: "npu-group-2",
				podName: "npu-test-30", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
				reqNPUType: npuV910CardName16c, reqNpuNum: "1"})
			setPodSelector(pod, acceleratorType, moduleAcceleratorType)
			task := vapi.NewTaskInfo(pod)
			result := vnpu.IsMyTask(task)
			So(result, ShouldBeNil)
		})
	})
}

// TestMVnpuIsMyNode
func TestMVnpuIsMyNode(t *testing.T) {
	Convey("Test IsMyNode", t, func() {
		vnpu := &modulev910x8{}

		Convey("IsMyNode() should return error when node has no vnpu annotation", func() {
			node := buildNPUNode(VMNodeInfo{nodeName, huaweiArchArm, "192", "755Gi",
				"1", "Ascend910-16c-150-0"})
			setNodeAnnotation(node.Node, npuV910CardName16c, "")
			result := vnpu.IsMyNode(node)
			So(result, ShouldBeError)
		})
		Convey("IsMyNode() should return error when value of needed node selector is wrong", func() {
			node := buildNPUNode(VMNodeInfo{nodeName, huaweiArchArm, "192", "755Gi",
				"1", "Ascend910-16c-151-0"})
			setNodeLabel(node.Node, acceleratorType, cardAcceleratorType)
			result := vnpu.IsMyNode(node)
			So(result, ShouldBeError)
		})
		Convey("IsMyNode() should return nil when value of needed node selector is correct", func() {
			node := buildNPUNode(VMNodeInfo{nodeName, huaweiArchArm, "192", "755Gi",
				"1", "Ascend910-16c-152-0"})
			setNodeLabel(node.Node, archSelector, huaweiArchArm)
			setNodeLabel(node.Node, acceleratorType, moduleAcceleratorType)
			result := vnpu.IsMyNode(node)
			So(result, ShouldBeNil)
		})
		Convey("IsMyNode() should return nil when node is of module mode", func() {
			node := buildNPUNode(VMNodeInfo{nodeName, huaweiArchArm, "192", "755Gi",
				"1", "Ascend910-16c-153-0"})
			result := vnpu.IsMyNode(node)
			So(result, ShouldBeNil)
		})
	})
}

// TestMVnpuIsMyJob
func TestMVnpuIsMyJob(t *testing.T) {

	Convey("Test IsMyJob", t, func() {
		vnpu := &modulev910x8{}
		tasks := []*vapi.TaskInfo{}
		uid := vapi.JobID("xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx")

		Convey("IsMyJob() should return error when job request no vNPU", func() {
			tasks = append(tasks, vapi.NewTaskInfo(buildNPUPod(
				VMPodInfo{namespace: "default", groupName: "npu-group-3",
					podName: "npu-test-31", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
					reqNPUType: npuV910CardName16c, reqNpuNum: "0"})))
			job := vapi.NewJobInfo(uid, tasks...)
			result := vnpu.IsMyJob(job)
			So(result, ShouldBeError)
		})
		Convey("IsMyJob() should return nil when job is of module type", func() {
			tasks = append(tasks, vapi.NewTaskInfo(buildNPUPod(
				VMPodInfo{namespace: "default", groupName: "npu-group-4",
					podName: "npu-test-32", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
					reqNPUType: npuV910CardName16c, reqNpuNum: "1"})))
			job := vapi.NewJobInfo(uid, tasks...)
			result := vnpu.IsMyJob(job)
			So(result, ShouldBeNil)
		})
	})
}

func setPodSelector(mPod *v1.Pod, selectorKey string, selectorValue string) {
	if selectorValue == "" {
		delete(mPod.Spec.NodeSelector, selectorKey)
		return
	}
	mPod.Spec.NodeSelector[selectorKey] = selectorValue
}

func setNodeAnnotation(mNode *v1.Node, annotationKey string, annotationValue string) {
	mNode.Annotations[annotationKey] = annotationValue
}

func setNodeLabel(mNode *v1.Node, labelKey string, labelValue string) {
	if labelValue == "" {
		delete(mNode.Labels, labelKey)
		return
	}
	mNode.Labels[labelKey] = labelValue
}

func buildNPUPod(podInfo VMPodInfo) *v1.Pod {

	pod := util.BuildPod(podInfo.namespace, podInfo.podName, podInfo.nodeName, v1.PodPending,
		buildNPUResourceList(podInfo.reqCPUNum, podInfo.reqMem,
			v1.ResourceName(podInfo.reqNPUType), podInfo.reqNpuNum),
		podInfo.groupName, make(map[string]string, constIntNum2),
		make(map[string]string, constIntNum2))

	setPodSelector(pod, archSelector, huaweiArchX86)

	return pod
}

func buildNPUResourceList(mCPU string, mMemory string, npuResourceType v1.ResourceName, npu string) v1.ResourceList {
	npuNum, err := strconv.Atoi(npu)
	if err != nil {
		return nil
	}

	if npuNum == 0 {
		return v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse(mCPU),
			v1.ResourceMemory: resource.MustParse(mMemory),
		}
	}

	return v1.ResourceList{
		v1.ResourceCPU:    resource.MustParse(mCPU),
		v1.ResourceMemory: resource.MustParse(mMemory),
		npuResourceType:   resource.MustParse(npu),
	}
}

func buildNPUNode(mNode VMNodeInfo) *vapi.NodeInfo {
	nodeCapacity := buildNPUResourceList(mNode.cpu, mNode.mem, npuV910CardName16c, strconv.Itoa(constIntNum2))
	nodeAlloc := buildNPUResourceList(mNode.cpu, mNode.mem, npuV910CardName16c, mNode.npuAllocateNum)
	labels := make(map[string]string, labelSize)
	ann := make(map[string]string, annSize)

	v1node := &v1.Node{
		ObjectMeta: mmetav1.ObjectMeta{
			Name:        mNode.nodeName,
			Labels:      labels,
			Annotations: ann,
		},
		Status: v1.NodeStatus{
			Capacity:    nodeCapacity,
			Allocatable: nodeAlloc,
		},
	}

	if mNode.npuAllocateNum != "0" {
		v1node.Annotations[npuV910CardName16c] = mNode.npuTop
	}

	setNodeLabel(v1node, archSelector, mNode.nodeArch)

	node := vapi.NewNodeInfo(v1node)
	return node
}
