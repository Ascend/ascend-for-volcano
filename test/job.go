/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package test is using for HuaWei Ascend pin scheduling test.

*/
package test

import (
	"fmt"
	"reflect"
	"strconv"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	var minRes = make(v1.ResourceList, npuIndex3)
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
func AddTestJobLabel(job *api.JobInfo, labelKey, labelValue string) {
	if job.PodGroup == nil {
		AddTestJobPodGroup(job)
	}
	if job.PodGroup.Labels == nil {
		job.PodGroup.Labels = make(map[string]string, npuIndex3)
	}
	job.PodGroup.Labels = map[string]string{labelKey: labelValue}
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

	var minRes = make(v1.ResourceList, npuIndex3)
	minRes[v1.ResourceName(name)] = resource.MustParse(fmt.Sprintf("%f", float64(value)))
	fJob.PodGroup.Spec.MinResources = &minRes
	if fJob.TotalRequest == nil {
		reqResource := api.NewResource(*fJob.PodGroup.Spec.MinResources)
		fJob.TotalRequest = reqResource
	}
	return
}

// UpdateFakeJobRequestSource Update job require on total,task.
func UpdateFakeJobRequestSource(fJob *api.JobInfo, name string, value int) {
	if fJob == nil {
		return
	}
	var minRes = make(v1.ResourceList, npuIndex3)
	minRes[v1.ResourceName(name)] = resource.MustParse(fmt.Sprintf("%f", float64(value)))
	if fJob.TotalRequest == nil || reflect.DeepEqual(fJob.TotalRequest, api.EmptyResource()) {
		reqResource := api.NewResource(minRes)
		fJob.TotalRequest = reqResource
	}
}

func setFakeJobSelector(fJob *api.JobInfo, selectorKey, selectorValue string) {
	if fJob.Tasks == nil || len(fJob.Tasks) == 0 {
		return
	}
	for _, task := range fJob.Tasks {
		setFakePodLabel(task.Pod, selectorKey, selectorValue)
	}
}

func setFakeJobResRequest(fJob *api.JobInfo, name v1.ResourceName, need string) {
	resources := v1.ResourceList{}
	AddResource(resources, name, need)
	if fJob.Tasks == nil || len(fJob.Tasks) == 0 {
		return
	}
	for _, task := range fJob.Tasks {
		task.Resreq = api.NewResource(resources)
	}
}

// BuildFakeJobWithSelectorAndSource build fake job with selector and request source
func BuildFakeJobWithSelectorAndSource(jobName, selectorKey, selectorValue, npuName string, npuNum int) *api.JobInfo {
	job := FakeNormalTestJob(jobName, 1)
	setFakeJobSelector(job, selectorKey, selectorValue)
	UpdateFakeJobRequestSource(job, npuName, npuNum)
	valueString := strconv.Itoa(npuNum)
	setFakeJobResRequest(job, v1.ResourceName(npuName), valueString)
	return job
}
