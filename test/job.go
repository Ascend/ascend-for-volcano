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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"volcano.sh/apis/pkg/apis/scheduling"
	"volcano.sh/volcano/pkg/scheduler/api"
)

// SetTestJobPodGroupStatus set test job's PodGroupStatus
func SetTestJobPodGroupStatus(job *api.JobInfo, status scheduling.PodGroupPhase) {
	if job.PodGroup == nil {
		addTestJobPodGroup(job)
	}
}

func addTestJobPodGroup(job *api.JobInfo) {
	pg := &api.PodGroup{
		PodGroup: scheduling.PodGroup{
			ObjectMeta: metav1.ObjectMeta{
				Name:      job.Name,
				Namespace: "vcjob",
			},
			Spec: scheduling.PodGroupSpec{
				Queue: "c1",
			},
		},
		Version: api.PodGroupVersionV1Beta1,
	}
	job.SetPodGroup(pg)
}

// AddTestJobLabel add test job's label.
func AddTestJobLabel(job *api.JobInfo, key, value string) {
	if job.PodGroup == nil {
		addTestJobPodGroup(job)
	}
	if job.PodGroup.Labels == nil {
		job.PodGroup.Labels = make(map[string]string, constIntNum3)
	}
	job.PodGroup.Labels = map[string]string{key: value}
}

// FakeNormalTestJob make normal test job.
func FakeNormalTestJob(jobName string, taskNum int) *api.JobInfo {
	tasks := FakeNormalTestTasks(taskNum)
	job := api.NewJobInfo(api.JobID("vcjob/"+jobName), tasks...)
	job.Name = jobName
	for _, task := range tasks {
		task.Job = job.UID
	}
	job.PodGroup = new(api.PodGroup)
	job.PodGroup.Status.Phase = scheduling.PodGroupPending
	return job
}
