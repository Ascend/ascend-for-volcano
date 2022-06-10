/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package rescheduling is using for HuaWei Ascend pin fault rescheduling.

*/
package rescheduling

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/types"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
)

type getFaultNPUJobsArgs struct {
	jobs     map[string]*api.JobInfo
	cacheFun func()
}

type getFaultNPUJobsTest struct {
	name    string
	args    getFaultNPUJobsArgs
	want    []FaultNPUJob
	wantErr error
}

func buildGetFaultNPUJobsTestCases() []getFaultNPUJobsTest {
	jobInf := test.FakeNormalTestJob("pg1", util.NPUIndex2)
	nodes := test.FakeNormalTestNodes(1)
	testCases := []getFaultNPUJobsTest{
		{
			name: "01-job not set reschedule label-test",
			args: getFaultNPUJobsArgs{
				jobs: map[string]*api.JobInfo{jobInf.Name: jobInf}, cacheFun: func() {
					test.AddTestJobLabel(jobInf, "haha", "who")
				},
			},
			wantErr: errors.New("get none faultNPUJobs"),
		},
		{
			name: "02-job not has fault resources-test",
			args: getFaultNPUJobsArgs{
				jobs: map[string]*api.JobInfo{jobInf.Name: jobInf}, cacheFun: func() {
					test.AddTestJobLabel(jobInf, jobRescheduleLabelKey, JobGraceRescheduleLabelValue)
					initTestReSchedulerCache()
					addTestNodeIntoReSchedulerCache()
				},
			},
			wantErr: errors.New("get none faultNPUJobs"),
		},
		{
			name: "03-job not has rankIndex-test",
			args: getFaultNPUJobsArgs{
				jobs: map[string]*api.JobInfo{jobInf.Name: jobInf}, cacheFun: func() {
					test.AddTestJobLabel(jobInf, jobRescheduleLabelKey, JobForceRescheduleLabelValue)
					initTestReSchedulerCache()
					addTestNodeIntoReSchedulerCache(nodes[0])
				},
			},
			wantErr: errors.New("get none faultNPUJobs"),
		},
		{
			name: "04-job not has rankIndex-test",
			args: getFaultNPUJobsArgs{
				jobs: map[string]*api.JobInfo{jobInf.Name: jobInf}, cacheFun: func() {
					test.AddTestJobLabel(jobInf, jobRescheduleLabelKey, JobForceRescheduleLabelValue)
					initTestReSchedulerCache()
					addTestNodeIntoReSchedulerCache(nodes[0])
					addTestJobRankIndex(jobInf)
				},
			},
			wantErr: nil,
		},
	}
	return testCases
}

// TestGetFaultNPUJobs Test GetFaultNPUJobs function
func TestGetFaultNPUJobs(t *testing.T) {
	tests := buildGetFaultNPUJobsTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.cacheFun()
			_, err := GetFaultNPUJobs(tt.args.jobs)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("GetFaultNPUJobs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

type getRestartNPUFaultJobsArgs struct {
	faultNPUJobs []FaultNPUJob
	jobs         map[string]*api.JobInfo
}

type getRestartNPUFaultJobsTest struct {
	name    string
	args    getRestartNPUFaultJobsArgs
	want    []*api.JobInfo
	wantErr error
}

func buildTestGetRestartNPUFaultJobsTestCases() []getRestartNPUFaultJobsTest {
	job := test.FakeNormalTestJob("pg1", util.NPUIndex2)
	faultJob := addJobIntoFaultNPUJobStruct(job)
	testCases := []getRestartNPUFaultJobsTest{
		{
			name: "01-no restart jobs-test",
			args: getRestartNPUFaultJobsArgs{
				faultNPUJobs: nil, jobs: map[string]*api.JobInfo{job.Name: job},
			},
			want:    nil,
			wantErr: errors.New("none restart jobs get"),
		},
		{
			name: "02-has restart jobs-test",
			args: getRestartNPUFaultJobsArgs{
				faultNPUJobs: []FaultNPUJob{faultJob}, jobs: map[string]*api.JobInfo{job.Name: job},
			},
			want:    []*api.JobInfo{job},
			wantErr: nil,
		},
	}
	return testCases
}

// TestGetRestartNPUFaultJobs Test GetRestartNPUFaultJobs function.
func TestGetRestartNPUFaultJobs(t *testing.T) {

	tests := buildTestGetRestartNPUFaultJobsTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetRestartNPUFaultJobs(tt.args.faultNPUJobs, tt.args.jobs)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("GetRestartNPUFaultJobs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetRestartNPUFaultJobs() got = %v, want %v", got, tt.want)
			}
		})
	}
}

type isDistributedJobArgs struct {
	job *api.JobInfo
}

type isDistributedJobTest struct {
	name string
	args isDistributedJobArgs
	want bool
}

func buildsDistributedJobTestCases() []isDistributedJobTest {
	job1 := test.FakeNormalTestJob("pg1", 1)
	job2 := test.FakeNormalTestJob("pg1", util.NPUIndex2)
	testCases := []isDistributedJobTest{
		{
			name: "01-not distributed job-test",
			args: isDistributedJobArgs{
				job: job1,
			},
			want: false,
		},
		{
			name: "02-distributed job-test",
			args: isDistributedJobArgs{
				job: job2,
			},
			want: true,
		},
	}
	return testCases
}

// TestIsDistributedJob Test IsDistributedJob function.
func TestIsDistributedJob(t *testing.T) {
	tests := buildsDistributedJobTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsDistributedJob(tt.args.job); got != tt.want {
				t.Errorf("IsDistributedJob() = %v, want %v", got, tt.want)
			}
		})
	}
}

type releaseFaultJobTakeNodesArgs struct {
	job      *api.JobInfo
	cacheFun func()
}

type releaseFaultJobTakeNodesTest struct {
	name    string
	args    releaseFaultJobTakeNodesArgs
	wantErr error
}

func buildReleaseFaultJobTakeNodesTestCases() []releaseFaultJobTakeNodesTest {
	fakeJob := test.FakeNormalTestJob("pg", util.NPUIndex3)
	fakeJob1 := test.FakeNormalTestJob("pg", util.NPUIndex3)
	testCases := []releaseFaultJobTakeNodesTest{
		{
			name: "01-no record fault Cache-test",
			args: releaseFaultJobTakeNodesArgs{
				job: fakeJob, cacheFun: func() {
					initTestReSchedulerCache()
				},
			},
			wantErr: nil,
		},
		{
			name: "02-job not in fault Cache-test",
			args: releaseFaultJobTakeNodesArgs{
				job: fakeJob, cacheFun: func() {
					initTestReSchedulerCache()
					addTestJobIntoReSchedulerCache(fakeJob1)
				},
			},
			wantErr: nil,
		},
		{
			name: "03-release job from fault Cache-test",
			args: releaseFaultJobTakeNodesArgs{
				job: fakeJob, cacheFun: func() {
					initTestReSchedulerCache()
					addTestJobIntoReSchedulerCache(fakeJob)
				},
			},
			wantErr: nil,
		},
	}
	return testCases
}

// TestReleaseFaultJobTakeNodes test ReleaseFaultJobTakeNodes function.
func TestReleaseFaultJobTakeNodes(t *testing.T) {
	tests := buildReleaseFaultJobTakeNodesTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.cacheFun()
			err := ReleaseFaultJobTakeNodes(tt.args.job)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("ReleaseFaultJobTakeNodes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type getRecordJobPodsArgs struct {
	dJob     *api.JobInfo
	cacheFun func()
}

type getRecordJobPodsTests struct {
	name    string
	args    getRecordJobPodsArgs
	want    map[string]int64
	want1   map[string]types.UID
	wantErr error
}

func buildGetRecordJobPodsTestCases() []getRecordJobPodsTests {
	fakeJob0 := test.FakeNormalTestJob("pg0", util.NPUIndex3)
	fakeJob0.Namespace = "vcjob"
	fakeJob1 := test.FakeNormalTestJob("pg1", util.NPUIndex3)
	fakeJob1.Namespace = "vcjob"
	testCases := []getRecordJobPodsTests{
		{
			name: "01-GetRecordJobPods()- no rescheduler cache configured-test",
			args: getRecordJobPodsArgs{
				dJob: fakeJob0,
				cacheFun: func() {
					initTestReSchedulerCache()
				},
			},
			want:    nil,
			want1:   nil,
			wantErr: fmt.Errorf("none %v in cache", CmJobRankIds),
		},
		{
			name: "02- GetRecordJobPods()- the job to be scheduled is not the job recorded in cache-test",
			args: getRecordJobPodsArgs{
				dJob: fakeJob0,
				cacheFun: func() {
					initTestReSchedulerCache()
					addTestJobIntoReSchedulerCache(fakeJob1)
					addTestJobRankIndexIntoReschedulerCache(fakeJob1)
				},
			},
			want:    nil,
			want1:   nil,
			wantErr: fmt.Errorf("none job %v in cache", api.JobID(fakeJob0.Namespace+"/"+fakeJob0.Name)),
		},
		{
			name: "03- GetRecordJobPods()- -test",
			args: getRecordJobPodsArgs{
				dJob: fakeJob0,
				cacheFun: func() {
					initTestReSchedulerCache()
					addTestJobIntoReSchedulerCache(fakeJob0)
					addTestJobRankIndexIntoReschedulerCache(fakeJob0)
				},
			},
			want:    map[string]int64{"pod0": 0, "pod1": 0, "pod2": 0},
			want1:   map[string]types.UID{"pod0": "vcjob-pod0", "pod1": "vcjob-pod1", "pod2": "vcjob-pod2"},
			wantErr: nil,
		},
	}
	return testCases
}

//TestGetRecordJobPods test GetRecordJobPods function
func TestGetRecordJobPods(t *testing.T) {
	tests := buildGetRecordJobPodsTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.cacheFun()
			got, got1, err := GetRecordJobPods(tt.args.dJob)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("GetRecordJobPods() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetRecordJobPods() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("GetRecordJobPods() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

type getNeedForceDeleteDelayingJobsArgs struct {
	ssn       *framework.Session
	dJobs     map[api.JobID]ReSchedulerTasks
	cacheFunc func()
}

type getNeedForceDeleteDelayingJobsTests struct {
	name    string
	args    getNeedForceDeleteDelayingJobsArgs
	want    []*api.JobInfo
	wantErr error
}

func buildGetNeedForceDeleteDelayingJobs() []getNeedForceDeleteDelayingJobsTests {
	const tmpNumber = 123456
	ssn0 := test.FakeNormalSSN()
	job0 := test.FakeNormalTestJob("pg0", util.NPUIndex3)
	reJobs0 := map[api.JobID]ReSchedulerTasks{
		job0.UID: {nil, nil, nil,
			nil, nil, "", false, tmpNumber}}

	ssn1 := test.FakeNormalSSN()
	job1 := test.FakeNormalTestJob("pg0", util.NPUIndex3)
	test.AddJobIntoFakeSSN(ssn1, job1)
	var reJobs1 = make(map[api.JobID]ReSchedulerTasks, util.NPUIndex2)
	for _, task := range job1.Tasks {
		reJobs1[task.Job] = ReSchedulerTasks{
			TaskName:    []string{task.Name},
			NodeNames:   []string{task.NodeName},
			RankIndexes: []string{task.Pod.Annotations[podRankIndex]},
			Time:        nil,
			TaskUseNPUs: []string{task.Pod.Annotations[npu800And9000CardName]},
			NameSpace:   "vcjob"}
	}

	testCases := []getNeedForceDeleteDelayingJobsTests{
		{
			name: "01-GetNeedForceDeleteDelayingJobs()- return because cannot get job-test",
			args: getNeedForceDeleteDelayingJobsArgs{
				ssn:       ssn0,
				dJobs:     reJobs0,
				cacheFunc: func() {},
			},
			want:    nil,
			wantErr: errors.New("get none jobs"),
		},

		{
			name: "03-GetNeedForceDeleteDelayingJobs()- success-test",
			args: getNeedForceDeleteDelayingJobsArgs{
				ssn:   ssn1,
				dJobs: reJobs1,
				cacheFunc: func() {
					initTestReSchedulerCache()
					addTestJobIntoReSchedulerCache(job1)
					addTestJobRankIndexIntoReschedulerCache(job1)
				},
			},
			want:    []*api.JobInfo{job1},
			wantErr: nil,
		},
	}
	return testCases
}

//TestGetNeedForceDeleteDelayingJobs test GetNeedForceDeleteDelayingJobs function
func TestGetNeedForceDeleteDelayingJobs(t *testing.T) {
	tests := buildGetNeedForceDeleteDelayingJobs()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.cacheFunc()
			got, err := GetNeedForceDeleteDelayingJobs(tt.args.ssn, tt.args.dJobs)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("GetNeedForceDeleteDelayingJobs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetNeedForceDeleteDelayingJobs() got = %v, want %v", got, tt.want)
			}
		})
	}
}

type getFaultJobPODRankIndexMapFromCacheArgs struct {
	restartJob *api.JobInfo
	cacheFunc  func()
}

type getFaultJobPODRankIndexMapFromCacheTests struct {
	name    string
	args    getFaultJobPODRankIndexMapFromCacheArgs
	want    map[string]string
	wantErr error
}

func buildGetFaultJobPODRankIndexMapFromCacheTestCases() []getFaultJobPODRankIndexMapFromCacheTests {
	job0 := test.FakeNormalTestJob("pg0", util.NPUIndex2)
	testCases := []getFaultJobPODRankIndexMapFromCacheTests{
		{
			name: "01-getFaultJobPODRankIndexMapFromCache()- read cache failed-test",
			args: getFaultJobPODRankIndexMapFromCacheArgs{
				restartJob: job0,
				cacheFunc:  func() {},
			},
			want:    nil,
			wantErr: fmt.Errorf("%s none rankIndex in cache", job0.UID),
		},
	}
	return testCases
}

func TestGetFaultJobPODRankIndexMapFromCache(t *testing.T) {
	tests := buildGetFaultJobPODRankIndexMapFromCacheTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.cacheFunc()
			got, err := getFaultJobPODRankIndexMapFromCache(tt.args.restartJob)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("getFaultJobPODRankIndexMapFromCache() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getFaultJobPODRankIndexMapFromCache() got = %v, want %v", got, tt.want)
			}
		})
	}
}

type CheckJobPodStatusOKArgs struct {
	ssn *framework.Session
	job *api.JobInfo
}

type CheckJobPodStatusOKTests struct {
	name string
	args CheckJobPodStatusOKArgs
	want bool
}

func buildCheckJobPodStatusOK() []CheckJobPodStatusOKTests {
	ssn0 := test.FakeNormalSSN()
	job0 := test.FakeNormalTestJob("pg0", 0)
	test.AddTestJobPodGroup(job0)
	test.AddJobIntoFakeSSN(ssn0, job0)
	testCases := []CheckJobPodStatusOKTests{
		{
			name: "01-CheckJobPodStatusOK()- success but with no pod-test",
			args: CheckJobPodStatusOKArgs{
				ssn: ssn0,
				job: job0,
			},
			want: true,
		},
	}
	return testCases
}

func TestCheckJobPodStatusOK(t *testing.T) {
	tests := buildCheckJobPodStatusOK()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkJobPodStatusOK(tt.args.ssn, tt.args.job); got != tt.want {
				t.Errorf("checkJobPodStatusOK() = %v, want %v", got, tt.want)
			}
		})
	}
}

type SynReSchedulerJobCacheArgs struct {
	ssn      *framework.Session
	tmpValue interface{}
}

type SynReSchedulerJobCacheTests struct {
	name    string
	args    SynReSchedulerJobCacheArgs
	wantErr error
}

func buildSynReSchedulerJobCacheTestCases() []SynReSchedulerJobCacheTests {
	const tmpNumber = 123456
	ssn0 := test.FakeNormalSSN()
	tmpValue0 := 0

	ssn1 := test.FakeNormalSSN()
	tmpValue1 := map[api.JobID]ReSchedulerTasks{
		"haha": {nil, nil, nil,
			nil, nil, "vcjob", false, tmpNumber}}

	ssn2 := test.FakeNormalSSN()
	job2 := test.FakeNormalTestJob("pg0", util.NPUIndex2)
	job2.MinAvailable = util.NPUIndex3
	test.AddJobIntoFakeSSN(ssn2, job2)
	tmpValue2 := getRejobs(job2)

	testCases := []SynReSchedulerJobCacheTests{
		{
			name: "01-SynReSchedulerJobCache()- tmpvalue not a rescheduletask-test",
			args: SynReSchedulerJobCacheArgs{
				ssn:      ssn0,
				tmpValue: tmpValue0,
			},
			wantErr: fmt.Errorf("convert %v to map[api.JobID]ReSchedulerTasks failed", tmpValue0),
		},
		{
			name: "02-SynReSchedulerJobCache()- success with empty retask-test",
			args: SynReSchedulerJobCacheArgs{
				ssn:      ssn1,
				tmpValue: tmpValue1,
			},
			wantErr: nil,
		},

		{
			name: "03-SynReSchedulerJobCache()- success with empty retask-test",
			args: SynReSchedulerJobCacheArgs{
				ssn:      ssn2,
				tmpValue: tmpValue2,
			},
			wantErr: nil,
		},
	}
	return testCases
}

func TestSynReSchedulerJobCache(t *testing.T) {
	tests := buildSynReSchedulerJobCacheTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := synReSchedulerJobCache(tt.args.ssn, tt.args.tmpValue)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("synReSchedulerJobCache() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
