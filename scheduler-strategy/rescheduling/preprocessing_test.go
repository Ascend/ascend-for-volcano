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

Package rescheduling is using for HuaWei Ascend pin fault rescheduling.

*/
package rescheduling

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ms "strconv"
	"testing"
	vapi "volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/util"
)

var (
	index = 0
)

type MNodeInfo struct {
	nodeName       string
	nodeArch       string
	cpu, mem       string
	npuAllocateNum string
	npuTop         string
}

func buildNPUResourceList(cCPU string, cMemory string, npuResourceType v1.ResourceName, npu string) v1.ResourceList {
	npuNum, err := ms.Atoi(npu)
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

func setNodeLabel(CNode *v1.Node, labelKey string, labelValue string) {
	if labelValue == "" {
		delete(CNode.Labels, labelKey)
		return
	}
	CNode.Labels[labelKey] = labelValue
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

func setPodSelector(MPod *v1.Pod, selectorKey string, selectorValue string) {
	MPod.Spec.NodeSelector[selectorKey] = selectorValue
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

func buildJobWithTwoTasks(firstNodeName, secondNodeName, firstAnnotation, secondAnnotation, firstIndex, secondIndex string) *vapi.JobInfo {
	index++
	uid := vapi.JobID(fmt.Sprintf("xxxxxxxx-xxxx-x1xx-xxxx-xxxxxxxxxxx%d", index))
	firstTask := vapi.NewTaskInfo(buildNPUPod(VCPodInfo{namespace: "default",
		groupName: fmt.Sprintf("group-M-job-%d", index),
		podName:   fmt.Sprintf("npu-test-M-job-%d", index),
		nodeName:  firstNodeName, reqCPUNum: "10", reqMem: "10Gi",
		reqNPUType: npu800And9000CardName, reqNpuNum: "8"}))
	firstTask.Pod.Annotations = map[string]string{
		npu800And9000CardName: firstAnnotation,
		podRankIndex:          firstIndex,
	}
	index++
	secondTask := vapi.NewTaskInfo(buildNPUPod(VCPodInfo{namespace: "default",
		groupName: fmt.Sprintf("group-M-job-%d", index),
		podName:   fmt.Sprintf("npu-test-M-job-%d", index),
		nodeName:  secondNodeName, reqCPUNum: "15", reqMem: "15Gi",
		reqNPUType: npu800And9000CardName, reqNpuNum: "8"}))
	secondTask.Pod.Annotations = map[string]string{
		npu800And9000CardName: secondAnnotation,
		podRankIndex:          secondIndex,
	}
	tasks := []*vapi.TaskInfo{firstTask, secondTask}

	job := vapi.NewJobInfo(uid, tasks...)
	job.PodGroup = &vapi.PodGroup{}
	job.PodGroup.Status.Phase = "Running"

	return job
}

// TestMNPUGetFaultNPUJobs
func TestMNPUGetFaultNPUJobs(t *testing.T) {
	Convey("Test module910x8 getFaultNPUJobs", t, func() {
		const (
			annotation = "Ascend910-0,Ascend910-1,Ascend910-2,Ascend910-3," +
				"Ascend910-4,Ascend910-5,Ascend910-6,Ascend910-7"
			firstNodeName  = "testNode1"
			secondNodeName = "testNode2"
			firstIndex     = "0"
			secondIndex    = "1"
		)

		job := buildJobWithTwoTasks(firstNodeName, secondNodeName, annotation, annotation, firstIndex, secondIndex)
		jobs := map[string]*vapi.JobInfo{
			job.Name: job,
		}
		Convey("getFaultNPUJobs() should return", func() {
			var expectedResult []FaultNPUJob
			res, _ := GetFaultNPUJobs(jobs)
			So(res, ShouldBeNil)
			So(res, ShouldResemble, expectedResult)
		})
	})
}

// TestMNPUCheckFaultJobNode
func TestMNPUCheckFaultJobNode(t *testing.T) {
	Convey("Test module910x8 checkFaultJobNode", t, func() {
		rst := &ReSchedulerTasks{NodeNames: map[string]string{"fake_task": "NodeName"}}
		task := vapi.NewTaskInfo(buildNPUPod(VCPodInfo{namespace: "default", groupName: "group-M-model-0",
			podName: "npu-test-M-model-0", nodeName: "NodeName", reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: npu800And9000CardName, reqNpuNum: "4"}))
		node := buildNPUNode(MNodeInfo{"NodeName", huaweiArchX86, "192", "755Gi",
			"4", "Ascend910-4,Ascend910-6,Ascend910-7,Ascend910-5"})
		Convey("checkFaultJobNode() should return nil if task is in reschedule map", func() {
			tmp := make(map[vapi.JobID]ReSchedulerTasks)
			tmp[task.Job] = *rst
			ReSchedulerCache[CmJobKind] = tmp
			err := CheckFaultJobNode(task, node)
			So(err, ShouldBeNil)
		})
		Convey("checkFaultJobNode() should return error if node is in reschedule map", func() {
			// todo
			data := ReSchedulerCache[CmJobKind]
			delete(data.(map[vapi.JobID]ReSchedulerTasks), task.Job)
			tmp := make(map[vapi.JobID]ReSchedulerTasks)
			tmp[task.Job] = *rst
			ReSchedulerCache[CmJobKind] = tmp
			err := CheckFaultJobNode(task, node)
			So(err, ShouldBeNil)
		})
		Convey("checkFaultJobNode() should return nil if node isn't in reschedule map", func() {
			newNode := buildNPUNode(MNodeInfo{"newNode", huaweiArchArm, "192", "755Gi",
				"6", "Ascend910-4,Ascend910-6,Ascend910-7,Ascend910-5,Ascend910-1,Ascend910-2"})
			err := CheckFaultJobNode(task, newNode)
			So(err, ShouldBeNil)
		})
	})
}

// TestMNPUGetJobUsedNodes
func TestMNPUGetJobUsedNodes(t *testing.T) {
	Convey("Test module910x8 getJobUsedNodes", t, func() {
		uid := vapi.JobID("xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxx102")
		tasks := []*vapi.TaskInfo{}
		tasks = append(tasks, vapi.NewTaskInfo(buildNPUPod(
			VCPodInfo{namespace: "default", groupName: "group-M-model-2",
				podName: "npu-test-M-model-2", nodeName: "NodeName", reqCPUNum: "10", reqMem: "10Gi",
				reqNPUType: npu800And9000CardName, reqNpuNum: "4"})))
		job := vapi.NewJobInfo(uid, tasks...)
		job.PodGroup = &vapi.PodGroup{}
		Convey("getJobUsedNodes() should return error when job is not in running state", func() {
			job.PodGroup.Status.Phase = "Unknown"
			_, err := getJobUsedNodes(job)
			So(err, ShouldBeError)
		})
		job.PodGroup.Status.Phase = "Running"
		Convey("getJobUsedNodes() should return a correct NodeName to Pod map", func() {
			expectedResult := map[string]*v1.Pod{
				"NodeName": tasks[0].Pod,
			}
			result, err := getJobUsedNodes(job)
			So(err, ShouldBeNil)
			So(result, ShouldResemble, expectedResult)
		})
	})
}

func setNodeAnnotation(mNode *v1.Node, annotationKey string, annotationValue string) {
	mNode.Annotations[annotationKey] = annotationValue
}

// TestMNPUGetNodeIdleNPUIntCardsIncludeFaultTask
func TestMNPUGetNodeIdleNPUIntCardsIncludeFaultTask(t *testing.T) {
	Convey("Test module910x8 checkFaultJobNode", t, func() {
		const (
			constInt2 = 2
			constInt4 = 4
			constInt5 = 5
			constInt6 = 6
			constInt7 = 7
		)
		node := buildNPUNode(MNodeInfo{"NodeName", huaweiArchX86, "192", "755Gi",
			"4", "Ascend910-4,Ascend910-6,Ascend910-7,Ascend910-5"})
		task := vapi.NewTaskInfo(buildNPUPod(VCPodInfo{namespace: "default", groupName: "group-M-model-0",
			podName: "npu-test-M-model-0", nodeName: "NodeName", reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: npu800And9000CardName, reqNpuNum: "4"}))
		tmp := make(map[vapi.JobID]ReSchedulerTasks)
		tmp[task.Job] = ReSchedulerTasks{
			TaskUseNPUs: map[string]string{
				task.Name: "Ascend910-1,Ascend910-2",
			},
		}
		ReSchedulerCache[CmJobKind] = tmp
		Convey("", func() {
			setNodeAnnotation(node.Node, faultNPU, "Ascend910-1")
			result := GetNodeIdleNPUIntCardsIncludeFaultTask(task, node)
			So(result, ShouldResemble, []int{constInt4, constInt6, constInt7, constInt5, constInt2})
		})
	})
}
