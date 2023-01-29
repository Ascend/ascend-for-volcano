/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package test is using for HuaWei Ascend pin scheduling test.
*/
package test

import (
	"fmt"
	"strconv"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

func makePodSpec(pod NPUPod) v1.PodSpec {
	return v1.PodSpec{
		NodeName:     pod.NodeName,
		NodeSelector: pod.Selector,
		Containers:   []v1.Container{{Resources: v1.ResourceRequirements{Requests: pod.ReqSource}}},
	}
}

// BuildNPUPod built Pod object
func BuildNPUPod(pod NPUPod) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			UID:       types.UID(fmt.Sprintf("%#v-%#v", pod.Namespace, pod.Name)),
			Name:      pod.Name,
			Namespace: pod.Namespace,
			Labels:    pod.Labels,
			Annotations: map[string]string{
				v1beta1.KubeGroupNameAnnotationKey: pod.GroupName,
			},
		},
		Status: v1.PodStatus{
			Phase: pod.Phase,
		},
		Spec: makePodSpec(pod),
	}
}

// SetTestNPUPodAnnotation set NPU pod annotation for add pod use npu resource.
func SetTestNPUPodAnnotation(pod *v1.Pod, annotationKey string, annotationValue string) {
	if pod.Annotations == nil {
		pod.Annotations = make(map[string]string, npuIndex3)
	}

	pod.Annotations[annotationKey] = annotationValue
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

// FakeNormalTestTask fake normal test task.
func FakeNormalTestTask(name string, nodename string, groupname string) *api.TaskInfo {
	pod := NPUPod{
		Namespace: "vcjob", Name: name, NodeName: nodename, GroupName: groupname, Phase: v1.PodRunning,
		Labels:    make(map[string]string, util.MapInitNum),
		ReqSource: buildNPUResourceList("1", strconv.Itoa(NPUHexKilo), NPU910CardName, strconv.Itoa(NPUIndex8)),
	}
	task := api.NewTaskInfo(BuildNPUPod(pod))
	return task
}

// FakeVNPUTestTask fake vnpu test task.
func FakeVNPUTestTask(name string, nodename string, groupname string, num int) *api.TaskInfo {
	pod := NPUPod{
		Namespace: "vcjob", Name: name, NodeName: nodename, GroupName: groupname, Phase: v1.PodRunning,
		Labels:    make(map[string]string, util.MapInitNum),
		ReqSource: buildNPUResourceList("1", strconv.Itoa(NPUHexKilo), util.AscendNPUCore, strconv.Itoa(num)),
	}
	task := api.NewTaskInfo(BuildNPUPod(pod))
	return task
}

// FakeNormalTestTasks fake normal test tasks.
func FakeNormalTestTasks(num int) []*api.TaskInfo {
	var tasks []*api.TaskInfo

	for i := 0; i < num; i++ {
		strNum := strconv.Itoa(i)
		task := FakeNormalTestTask("pod"+strNum, "node"+strNum, "pg"+strNum)
		tasks = append(tasks, task)
	}

	return tasks
}

// BuildPodWithReqResource build pod with request resource
func BuildPodWithReqResource(resourceName v1.ResourceName, resourceNum string) *v1.Pod {
	resourceList := v1.ResourceList{}
	AddResource(resourceList, resourceName, resourceNum)
	return BuildNPUPod(NPUPod{ReqSource: resourceList})
}

// BuildTestTaskWithAnnotation build test task with annotation
func BuildTestTaskWithAnnotation(npuName, npuNum, npuAllocate string) *api.TaskInfo {
	pod := BuildPodWithReqResource(v1.ResourceName(npuName), npuNum)
	SetTestNPUPodAnnotation(pod, npuName, npuAllocate)
	return api.NewTaskInfo(pod)
}

// AddFakeTaskResReq add require resource of fake task.
func AddFakeTaskResReq(vTask *api.TaskInfo, name string, value float64) {
	if vTask == nil {
		return
	}

	if len(vTask.Resreq.ScalarResources) == 0 {
		vTask.Resreq.ScalarResources = make(map[v1.ResourceName]float64, npuIndex3)
	}
	vTask.Resreq.ScalarResources[v1.ResourceName(name)] = value
}

// SetFakeNPUTaskStatus task set same status.
func SetFakeNPUTaskStatus(fTask *api.TaskInfo, status api.TaskStatus) {
	if fTask == nil {
		return
	}
	fTask.Status = status
	return
}

// SetFakeNPUPodStatus set fake pod status.
func SetFakeNPUPodStatus(fPod *v1.Pod, status v1.PodPhase) {
	if fPod == nil {
		return
	}
	fPod.Status.Phase = status
	return
}

// AddTestTaskLabel add test job's label.
func AddTestTaskLabel(task *api.TaskInfo, labelKey, labelValue string) {
	if len(task.Pod.Spec.NodeSelector) == 0 {
		task.Pod.Spec.NodeSelector = make(map[string]string, npuIndex3)
	}
	task.Pod.Spec.NodeSelector[labelKey] = labelValue

	if len(task.Pod.Labels) == 0 {
		task.Pod.Labels = make(map[string]string, npuIndex3)
	}
	task.Pod.Labels[labelKey] = labelValue
}
