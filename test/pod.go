/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package test is using for HuaWei Ascend pin scheduling test.

*/
package test

import (
	"fmt"
	"strconv"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	schedulingv1 "volcano.sh/apis/pkg/apis/scheduling/v1beta1"
	"volcano.sh/volcano/pkg/scheduler/api"
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
			UID:       types.UID(fmt.Sprintf("%v-%v", pod.Namespace, pod.Name)),
			Name:      pod.Name,
			Namespace: pod.Namespace,
			Labels:    pod.Labels,
			Annotations: map[string]string{
				schedulingv1.KubeGroupNameAnnotationKey: pod.GroupName,
			},
		},
		Status: v1.PodStatus{
			Phase: pod.Phase,
		},
		Spec: makePodSpec(pod),
	}
}

// SetTestNPUPodAnnotation set NPU pod annotation for add pod use npu resource.
func SetTestNPUPodAnnotation(pod *v1.Pod, key string, value string) {
	if pod.Annotations == nil {
		pod.Annotations = make(map[string]string, npuIndex3)
	}

	pod.Annotations[key] = value
}

// SetTestNPUPodSelector set NPU pod selector.
func SetTestNPUPodSelector(pod *v1.Pod, key string, value string) {
	if pod.Spec.NodeSelector == nil {
		pod.Spec.NodeSelector = make(map[string]string, npuIndex3)
	}

	pod.Spec.NodeSelector[key] = value
}

// FakeNormalTestTasks fake normal test tasks.
func FakeNormalTestTasks(num int) []*api.TaskInfo {
	var tasks []*api.TaskInfo

	for i := 0; i < num; i++ {
		strNum := strconv.Itoa(i)
		pod := NPUPod{
			Namespace: "vcjob", Name: "pod" + strNum, NodeName: "node" + strNum, GroupName: "pg" + strNum, Phase: v1.PodRunning,
		}
		task := api.NewTaskInfo(BuildNPUPod(pod))
		tasks = append(tasks, task)
	}

	return tasks
}
