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

Package ascendtest is using for HuaWei Ascend pin scheduling test.

*/
package ascendtest

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"strconv"
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
