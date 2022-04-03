/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package ascendtest is using for HuaWei Ascend pin scheduling test.

*/
package ascendtest

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
	"volcano.sh/apis/pkg/apis/scheduling"
	"volcano.sh/volcano/pkg/scheduler/api"
)

// SetTestJobPodGroupStatus set test job's PodGroupStatus
func SetTestJobPodGroupStatus(job *api.JobInfo, status scheduling.PodGroupPhase) {
	if job.PodGroup == nil {
		AddTestJobPodGroup(job)
	}
}

// AddTestJobPodGroup set test job pg.
func AddTestJobPodGroup(job *api.JobInfo) {
	var minRes = make(v1.ResourceList, constIntNum3)
	if job == nil || job.PodGroup != nil {
		return
	}
	for _, task := range job.Tasks {
		for k, v := range task.Resreq.ScalarResources {
			minRes[k] = resource.MustParse(fmt.Sprintf("%f", v))
		}
	}

	pg := &api.PodGroup{
		PodGroup: scheduling.PodGroup{
			ObjectMeta: metav1.ObjectMeta{
				Name:      job.Name,
				Namespace: "vcjob",
			},
			Spec: scheduling.PodGroupSpec{
				Queue:        "c1",
				MinResources: &minRes,
			},
		},
		Version: api.PodGroupVersionV1Beta1,
	}
	job.SetPodGroup(pg)
}

// AddTestJobLabel add test job's label.
func AddTestJobLabel(job *api.JobInfo, key, value string) {
	if job.PodGroup == nil {
		AddTestJobPodGroup(job)
	}
	if job.PodGroup.Labels == nil {
		job.PodGroup.Labels = make(map[string]string, constIntNum3)
	}
	job.PodGroup.Labels = map[string]string{key: value}
}

// FakeNormalTestJob make normal test job.
func FakeNormalTestJob(jobName string, taskNum int) *api.JobInfo {
	return FakeNormalTestJobByCreatTime(jobName, taskNum, 0)
}

// FakeNormalTestJobByCreatTime make normal test job by create time.
func FakeNormalTestJobByCreatTime(jobName string, taskNum int, creatTime int64) *api.JobInfo {
	tasks := FakeNormalTestTasks(taskNum)
	job := api.NewJobInfo(api.JobID("vcjob/"+jobName), tasks...)
	job.Name = jobName
	for _, task := range tasks {
		task.Job = job.UID
	}
	job.PodGroup = new(api.PodGroup)
	job.PodGroup.Status.Phase = scheduling.PodGroupRunning
	job.CreationTimestamp = metav1.Time{Time: time.Unix(time.Now().Unix()+creatTime, 0)}
	return job
}

// SetFakeJobRequestSource add job require on total,task.
func SetFakeJobRequestSource(fJob *api.JobInfo, name string, value int) {
	AddTestJobPodGroup(fJob)
	SetTestJobPodGroupStatus(fJob, scheduling.PodGroupPending)

	var minRes = make(v1.ResourceList, constIntNum3)
	minRes[v1.ResourceName(name)] = resource.MustParse(fmt.Sprintf("%f", float64(value)))
	fJob.PodGroup.Spec.MinResources = &minRes
	if fJob.TotalRequest == nil {
		reqResource := api.NewResource(*fJob.PodGroup.Spec.MinResources)
		fJob.TotalRequest = reqResource
	}
	return
}
