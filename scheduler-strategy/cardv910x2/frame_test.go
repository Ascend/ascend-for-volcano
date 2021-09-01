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

Package cardv910x2 is using for virtual HuaWei A300T schedule.

*/
package cardv910x2

import (
	. "github.com/smartystreets/goconvey/convey"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	cmetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
	"testing"
	vapi "volcano.sh/volcano/pkg/scheduler/api"
	v910 "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/commonv910"
	"volcano.sh/volcano/pkg/scheduler/util"
)

const (
	nodeName           = "ubuntu"
	constIntNum2       = 2
	labelSize          = 8
	annSize            = 8
	npuV910CardName16c = "huawei.com/Ascend910-16c"
)

type VCNodeInfo struct {
	nodeName       string
	nodeArch       string
	cpu, mem       string
	npuAllocateNum string
	npuTop         string
}

type VCPodInfo struct {
	namespace  string
	groupName  string
	podName    string
	nodeName   string
	reqCPUNum  string
	reqMem     string
	reqNPUType string
	reqNpuNum  string
}

// TestCVnpuName
func TestCVnpuName(t *testing.T) {
	Convey("Test cardv910x2 Name", t, func() {
		const (
			pluginName = "A300T-Vnpu"
		)
		vnpu := &cardv910x2{}

		Convey("Name() should return PluginName defined in const", func() {
			n := vnpu.Name()
			So(n, ShouldEqual, pluginName)
		})
	})
}

// TestCVnpuNew
func TestCVnpuNew(t *testing.T) {
	Convey("Test cardv910x2 New", t, func() {
		const (
			name = "cardv910x2"
		)
		expectResult := &cardv910x2{
			name: name,
			Vnpu: v910.Vnpu{
				MaxNPUNum: maxNPUNum,
			},
		}

		Convey("Name() should return a cardv910x2 instance", func() {
			result := New(name)
			So(result, ShouldResemble, expectResult)
		})
	})
}

// TestCVnpuGetNpuJobDefaultSelectorConfig
func TestCVnpuGetNpuJobDefaultSelectorConfig(t *testing.T) {
	Convey("Test cardv910x2 GetNpuJobDefaultSelectorConfig", t, func() {
		vnpu := &cardv910x2{}
		expectResult := map[string]string{
			archSelector:    huaweiArchArm + "|" + huaweiArchX86,
			acceleratorType: cardAcceleratorType + "|" + moduleAcceleratorType,
		}

		Convey("GetNpuJobDefaultSelectorConfig() should return default configs", func() {
			result := vnpu.GetNpuJobDefaultSelectorConfig()
			So(result, ShouldResemble, expectResult)
		})
	})
}

// TestCVnpuIsMyTask
func TestCVnpuIsMyTask(t *testing.T) {
	Convey("Test IsMyTask", t, func() {
		vnpu := &cardv910x2{}

		Convey("IsMyTask() should return error when task doesn't request vNPU", func() {
			pod := buildNPUPod(VCPodInfo{namespace: "default", groupName: "npu-group-0",
				podName: "npu-test-21", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
				reqNPUType: npuV910CardName16c, reqNpuNum: "0"})
			task := vapi.NewTaskInfo(pod)
			result := vnpu.IsMyTask(task)
			So(result, ShouldBeError)
		})
		Convey("IsMyTask() should return error when task is not of card mode", func() {
			pod := buildNPUPod(VCPodInfo{namespace: "default", groupName: "npu-group-1",
				podName: "npu-test-22", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
				reqNPUType: npuV910CardName16c, reqNpuNum: "1"})
			setPodSelector(pod, acceleratorType, moduleAcceleratorType)
			task := vapi.NewTaskInfo(pod)
			result := vnpu.IsMyTask(task)
			So(result, ShouldBeError)
		})
		Convey("IsMyTask() should return nil when task is of card type", func() {
			pod := buildNPUPod(VCPodInfo{namespace: "default", groupName: "npu-group-2",
				podName: "npu-test-23", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
				reqNPUType: npuV910CardName16c, reqNpuNum: "1"})
			setPodSelector(pod, acceleratorType, cardAcceleratorType)
			task := vapi.NewTaskInfo(pod)
			result := vnpu.IsMyTask(task)
			So(result, ShouldBeNil)
		})
	})
}

// TestCVnpuIsMyNode
func TestCVnpuIsMyNode(t *testing.T) {
	Convey("Test IsMyNode", t, func() {
		vnpu := &cardv910x2{}

		Convey("IsMyNode() should return error when node has no vnpu annotation", func() {
			node := buildNPUNode(VCNodeInfo{nodeName, huaweiArchArm, "192", "755Gi",
				"1", "Ascend910-16c-118-0"})
			setNodeAnnotation(node.Node, npuV910CardName16c, "")
			result := vnpu.IsMyNode(node)
			So(result, ShouldBeError)
		})
		Convey("IsMyNode() should return error when value of node selector 'host-arch' is empty", func() {
			node := buildNPUNode(VCNodeInfo{nodeName, huaweiArchArm, "192", "755Gi",
				"1", "Ascend910-16c-119-0"})
			setNodeLabel(node.Node, archSelector, "")
			result := vnpu.IsMyNode(node)
			So(result, ShouldBeError)
		})
		Convey("IsMyNode() should return nil when node is of card mode", func() {
			node := buildNPUNode(VCNodeInfo{nodeName, huaweiArchArm, "192", "755Gi",
				"1", "Ascend910-16c-120-0"})
			setNodeLabel(node.Node, archSelector, huaweiArchArm)
			setNodeLabel(node.Node, acceleratorType, cardAcceleratorType)
			result := vnpu.IsMyNode(node)
			So(result, ShouldBeNil)
		})
	})
}

// TestCVnpuIsMyJob
func TestCVnpuIsMyJob(t *testing.T) {
	Convey("Test IsMyJob", t, func() {
		vnpu := &cardv910x2{}
		tasks := []*vapi.TaskInfo{}
		uid := vapi.JobID("xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx")

		Convey("IsMyJob() should return error when job request no vNPU", func() {
			tasks = append(tasks, vapi.NewTaskInfo(buildNPUPod(
				VCPodInfo{namespace: "default", groupName: "npu-group-3",
					podName: "npu-test-24", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
					reqNPUType: npuV910CardName16c, reqNpuNum: "0"})))
			job := vapi.NewJobInfo(uid, tasks...)
			result := vnpu.IsMyJob(job)
			So(result, ShouldBeError)
		})
		Convey("IsMyJob() should return error when job has no selector", func() {
			pod := buildNPUPod(VCPodInfo{namespace: "default", groupName: "npu-group-4",
				podName: "npu-test-25", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
				reqNPUType: npuV910CardName16c, reqNpuNum: "1"})
			setPodSelector(pod, archSelector, "")
			tasks = append(tasks, vapi.NewTaskInfo(pod))
			job := vapi.NewJobInfo(uid, tasks...)
			result := vnpu.IsMyJob(job)
			So(result, ShouldBeError)
		})
		Convey("IsMyJob() should return error when job has no selector acceleratorType", func() {
			tasks = append(tasks, vapi.NewTaskInfo(buildNPUPod(
				VCPodInfo{namespace: "default", groupName: "npu-group-5",
					podName: "npu-test-26", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
					reqNPUType: npuV910CardName16c, reqNpuNum: "1"})))
			job := vapi.NewJobInfo(uid, tasks...)
			result := vnpu.IsMyJob(job)
			So(result, ShouldBeError)
		})
		Convey("IsMyJob() should return nil when job is of card mode", func() {
			pod := buildNPUPod(VCPodInfo{namespace: "default", groupName: "npu-group-6",
				podName: "npu-test-27", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
				reqNPUType: npuV910CardName16c, reqNpuNum: "1"})
			setPodSelector(pod, acceleratorType, cardAcceleratorType)
			tasks = append(tasks, vapi.NewTaskInfo(pod))
			job := vapi.NewJobInfo(uid, tasks...)
			result := vnpu.IsMyJob(job)
			So(result, ShouldBeNil)
		})
	})
}

func setPodSelector(cPod *v1.Pod, selectorKey string, selectorValue string) {
	if selectorValue == "" {
		delete(cPod.Spec.NodeSelector, selectorKey)
		return
	}
	cPod.Spec.NodeSelector[selectorKey] = selectorValue
}

func setNodeAnnotation(cNode *v1.Node, annotationKey string, annotationValue string) {
	cNode.Annotations[annotationKey] = annotationValue
}

func setNodeLabel(cNode *v1.Node, labelKey string, labelValue string) {
	if labelValue == "" {
		delete(cNode.Labels, labelKey)
		return
	}
	cNode.Labels[labelKey] = labelValue
}

func buildNPUPod(podInfo VCPodInfo) *v1.Pod {

	pod := util.BuildPod(podInfo.namespace, podInfo.podName, podInfo.nodeName, v1.PodPending,
		buildNPUResourceList(podInfo.reqCPUNum, podInfo.reqMem,
			v1.ResourceName(podInfo.reqNPUType), podInfo.reqNpuNum),
		podInfo.groupName, make(map[string]string, constIntNum2),
		make(map[string]string, constIntNum2))

	setPodSelector(pod, archSelector, huaweiArchX86)

	return pod
}

func buildNPUResourceList(cCPU string, cMemory string, npuResourceType v1.ResourceName, npu string) v1.ResourceList {
	npuNum, err := strconv.Atoi(npu)
	if err != nil {
		return nil
	}

	if npuNum == 0 {
		return v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse(cCPU),
			v1.ResourceMemory: resource.MustParse(cMemory),
		}
	}

	return v1.ResourceList{
		v1.ResourceCPU:    resource.MustParse(cCPU),
		v1.ResourceMemory: resource.MustParse(cMemory),
		npuResourceType:   resource.MustParse(npu),
	}
}

func buildNPUNode(cNode VCNodeInfo) *vapi.NodeInfo {
	nodeCapacity := buildNPUResourceList(cNode.cpu, cNode.mem, npuV910CardName16c, strconv.Itoa(constIntNum2))
	nodeAlloc := buildNPUResourceList(cNode.cpu, cNode.mem, npuV910CardName16c, cNode.npuAllocateNum)
	labels := make(map[string]string, labelSize)
	ann := make(map[string]string, annSize)

	v1node := &v1.Node{
		ObjectMeta: cmetav1.ObjectMeta{
			Name:        cNode.nodeName,
			Labels:      labels,
			Annotations: ann,
		},
		Status: v1.NodeStatus{
			Capacity:    nodeCapacity,
			Allocatable: nodeAlloc,
		},
	}

	if cNode.npuAllocateNum != "0" {
		v1node.Annotations[npuV910CardName16c] = cNode.npuTop
	}

	setNodeLabel(v1node, archSelector, cNode.nodeArch)

	node := vapi.NewNodeInfo(v1node)
	return node
}
