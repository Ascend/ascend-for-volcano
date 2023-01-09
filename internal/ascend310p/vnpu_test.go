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
Package ascend310p is using for HuaWei Ascend pin affinity schedule.
*/
package ascend310p

import (
	"errors"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"volcano.sh/apis/pkg/apis/scheduling"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/vnpu"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

type TestCheckStVJobReqTest struct {
	Name    string
	vHandle *vnpu.VirtualNPU
	Tasks   map[api.TaskID]util.NPUTask
	WantErr error
}

func buildTestCheckStVJobReqTestCase01() []TestCheckStVJobReqTest {
	test := []TestCheckStVJobReqTest{
		{
			Name:    "01-TestCheckStVJobReq will return err when vHandle.DynamicByConf is true",
			vHandle: &vnpu.VirtualNPU{DynamicByConf: true},
			WantErr: errors.New("volcano configuration presetVirtualDevice false, only support dynamic vnpu"),
		},
		{
			Name:    "02-TestCheckStVJobReq will return err when vHandle.DynamicByConf is true",
			vHandle: &vnpu.VirtualNPU{DynamicByConf: false},
			Tasks: map[api.TaskID]util.NPUTask{
				"1234": {
					Name:       "task0",
					ReqNPUNum:  1,
					ReqNPUName: "error npu name",
				}},
			WantErr: errors.New("task0 req error npu name not in template"),
		},
		{
			Name:    "03-TestCheckStVJobReq will return err when ReqNPUNum is not 1",
			vHandle: &vnpu.VirtualNPU{DynamicByConf: false},
			Tasks: map[api.TaskID]util.NPUTask{"1234": {
				Name:       "task0",
				ReqNPUName: PluginName,
			}},
			WantErr: errors.New("task0 req 0 not 1"),
		},
	}
	return test
}

func TestCheckStVJobReq(t *testing.T) {
	n := New(PluginName)
	npu, ok := n.(*ascend310P)
	if !ok {
		return
	}
	npu.NPUJob = &util.NPUJob{}
	tests := buildTestCheckStVJobReqTestCase01()
	for _, tt := range tests {
		npu.vHandle = tt.vHandle
		npu.Tasks = tt.Tasks
		t.Run(tt.Name, func(t *testing.T) {
			if err := npu.checkStVJobReq(); !reflect.DeepEqual(err, tt.WantErr) {
				t.Errorf("CheckStVJobReq() error = %#v, wantErr %#v", err, tt.WantErr)
			}
		})
	}
}

type TestCheckDyVJobReqTest struct {
	Name    string
	vHandle *vnpu.VirtualNPU
	NPUJob  *util.NPUJob
	Tasks   map[api.TaskID]util.NPUTask
	WantErr error
}

func buildTestCheckDyVJobReqTestCase01() []TestCheckDyVJobReqTest {
	test := []TestCheckDyVJobReqTest{
		{
			Name:    "01-TestCheckDyVJobReq will return err when job is not VJob",
			NPUJob:  &util.NPUJob{ReqNPUName: util.NPU310PCardName},
			vHandle: &vnpu.VirtualNPU{DynamicByConf: true},
			Tasks:   map[api.TaskID]util.NPUTask{"1234": {Name: "task0"}},
			WantErr: errors.New(" not VirtualNPU job"),
		},
		{
			Name:    "02-TestCheckStVJobReq will return err when vHandle.DynamicByConf is false",
			vHandle: &vnpu.VirtualNPU{DynamicByConf: false},
			NPUJob:  &util.NPUJob{ReqNPUName: util.AscendNPUCore},
			Tasks:   map[api.TaskID]util.NPUTask{"1234": {Name: "task1"}},
			WantErr: errors.New("volcano configuration presetVirtualDevice true, only support static vnpu"),
		},
		{
			Name:    "03-TestCheckStVJobReq will return err when ReqNPUNum is not 1",
			vHandle: &vnpu.VirtualNPU{DynamicByConf: true},
			NPUJob:  &util.NPUJob{ReqNPUName: util.AscendNPUCore},
			Tasks: map[api.TaskID]util.NPUTask{"1234": {
				Name:      "task2",
				ReqNPUNum: 3,
			}},
			WantErr: errors.New("task2 req err 3"),
		},
	}
	return test
}

func TestCheckDyVJobReq(t *testing.T) {
	n := New(PluginName)
	npu, ok := n.(*ascend310P)
	if !ok {
		return
	}
	tests := buildTestCheckDyVJobReqTestCase01()
	for _, tt := range tests {
		npu.vHandle = tt.vHandle
		npu.NPUJob = tt.NPUJob
		npu.Tasks = tt.Tasks
		t.Run(tt.Name, func(t *testing.T) {
			if err := npu.checkDyVJobReq(); !reflect.DeepEqual(err, tt.WantErr) {
				t.Errorf("CheckDyVJobReq() error = %#v, wantErr %#v", err, tt.WantErr)
			}
		})
	}
}

type TestValidDyVNPUTaskDVPPLabelTest struct {
	Name    string
	vHandle *vnpu.VirtualNPU
	NPUJob  *util.NPUJob
	Task    util.NPUTask
	WantErr error
}

func buildTestValidDyVNPUTaskDVPPLabelTestCase01() TestValidDyVNPUTaskDVPPLabelTest {
	test := TestValidDyVNPUTaskDVPPLabelTest{
		Name: "01-test will return err when task is not vnpu task",
		Task: util.NPUTask{
			Name:       "task01",
			ReqNPUName: PluginName,
			ReqNPUNum:  1,
		},
		WantErr: errors.New("not vNPU task"),
	}
	return test
}

func buildTestValidDyVNPUTaskDVPPLabelTestCase02() TestValidDyVNPUTaskDVPPLabelTest {
	test := TestValidDyVNPUTaskDVPPLabelTest{
		Name: "02-test will return nil when task is a multiple of eight",
		Task: util.NPUTask{
			Name:       "task02",
			ReqNPUName: util.AscendNPUCore,
			ReqNPUNum:  8,
		},
		WantErr: nil,
	}
	return test
}

func buildTestValidDyVNPUTaskDVPPLabelTestCase03() TestValidDyVNPUTaskDVPPLabelTest {
	test := TestValidDyVNPUTaskDVPPLabelTest{
		Name: "03-test will return err when task is 1 or util.NPUIndex2",
		Task: util.NPUTask{
			Name:       "task03",
			Label:      map[string]string{plugin.AscendVNPUDVPP: plugin.AscendDVPPValue},
			ReqNPUName: util.AscendNPUCore,
			ReqNPUNum:  1,
		},
		WantErr: errors.New("task03 dvpp label err:dvpp"),
	}
	return test
}

func buildTestValidDyVNPUTaskDVPPLabelTestCase04() TestValidDyVNPUTaskDVPPLabelTest {
	test := TestValidDyVNPUTaskDVPPLabelTest{
		Name: "03-test will return nil when task is util.NPUIndex4",
		Task: util.NPUTask{
			Name:       "task04",
			ReqNPUName: util.AscendNPUCore,
			ReqNPUNum:  4,
		},
		WantErr: nil,
	}
	return test
}

func buildTestValidDyVNPUTaskDVPPLabelTestCase05() TestValidDyVNPUTaskDVPPLabelTest {
	test := TestValidDyVNPUTaskDVPPLabelTest{
		Name: "03-test will return nil when task ReqNPUNum is other",
		Task: util.NPUTask{
			Name:       "task04",
			ReqNPUName: util.AscendNPUCore,
			ReqNPUNum:  3,
		},
		WantErr: errors.New("err require number:3"),
	}
	return test
}

func buildTestValidDyVNPUTaskDVPPLabelTestCase() []TestValidDyVNPUTaskDVPPLabelTest {
	tests := []TestValidDyVNPUTaskDVPPLabelTest{
		buildTestValidDyVNPUTaskDVPPLabelTestCase01(),
		buildTestValidDyVNPUTaskDVPPLabelTestCase02(),
		buildTestValidDyVNPUTaskDVPPLabelTestCase03(),
		buildTestValidDyVNPUTaskDVPPLabelTestCase04(),
		buildTestValidDyVNPUTaskDVPPLabelTestCase05(),
	}
	return tests
}

func TestValidDyVNPUTaskDVPPLabel(t *testing.T) {
	n := New(PluginName)
	npu, ok := n.(*ascend310P)
	if !ok {
		return
	}
	tests := buildTestValidDyVNPUTaskDVPPLabelTestCase()
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			if err := npu.validDyVNPUTaskDVPPLabel(tt.Task); !reflect.DeepEqual(err, tt.WantErr) {
				t.Errorf("ValidDyVNPUTaskDVPPLabel() error = %#v, wantErr %#v", err, tt.WantErr)
			}
		})
	}
}

type TestValidDyVNPUJobLabelTest struct {
	Name    string
	NPUJob  *util.NPUJob
	Tasks   map[api.TaskID]util.NPUTask
	WantErr error
}

func buildTestValidDyVNPUJobLabelTestCase() []TestValidDyVNPUJobLabelTest {
	tests := []TestValidDyVNPUJobLabelTest{
		{
			Name:    "01-test will return err when Job is not VJob",
			NPUJob:  &util.NPUJob{ReqNPUName: util.NPU310CardName},
			Tasks:   nil,
			WantErr: errors.New(" not VirtualNPU job"),
		},
		{
			Name:   "02-test will return nil when Job is ok",
			NPUJob: &util.NPUJob{ReqNPUName: util.AscendNPUCore},
			Tasks: map[api.TaskID]util.NPUTask{
				"task01": {Name: "task01", ReqNPUName: util.AscendNPUCore, ReqNPUNum: 4}},
			WantErr: nil,
		},
		{
			Name:   "03-test will return err when validDyVNPUTaskDVPPLabel is not passed",
			NPUJob: &util.NPUJob{ReqNPUName: util.AscendNPUCore},
			Tasks: map[api.TaskID]util.NPUTask{
				"task01": {Name: "task01", ReqNPUName: util.AscendNPUCore, ReqNPUNum: 3}},
			WantErr: errors.New("err require number:3"),
		},
	}
	return tests
}

func TestValidDyVNPUJobLabel(t *testing.T) {
	n := New(PluginName)
	npu, ok := n.(*ascend310P)
	if !ok {
		return
	}
	npu.NPUJob = &util.NPUJob{}
	tests := buildTestValidDyVNPUJobLabelTestCase()
	for _, tt := range tests {
		npu.NPUJob = tt.NPUJob
		npu.Tasks = tt.Tasks
		t.Run(tt.Name, func(t *testing.T) {
			if err := npu.validDyVNPUJobLabel(); !reflect.DeepEqual(err, tt.WantErr) {
				t.Errorf("ValidDyVNPUJobLabel() error = %#v, wantErr %#v", err, tt.WantErr)
			}
		})
	}

}

type TestValidDyVNPUJobTest struct {
	Name string
	VJob *util.VJob
	Want *api.ValidateResult
}

func buildTestValidDyVNPUJobTest() []TestValidDyVNPUJobTest {
	tests := []TestValidDyVNPUJobTest{
		{
			Name: "01-test ValidDyVNPUJob will return nil when VJob status is Running",
			VJob: &util.VJob{Status: scheduling.PodGroupRunning},
			Want: nil,
		},
		{
			Name: "02-test ValidDyVNPUJob will return when check VJob Request is invalid",
			VJob: &util.VJob{Status: scheduling.PodGroupUnknown},
			Want: &api.ValidateResult{Pass: false, Reason: " not VirtualNPU job", Message: " not VirtualNPU job"},
		},
		{
			Name: "03-test ValidDyVNPUJob will return when validDyVNPUJobLabel is invalid",
			VJob: &util.VJob{Status: scheduling.PodGroupUnknown},
			Want: &api.ValidateResult{Pass: false, Reason: " not VirtualNPU job", Message: " not VirtualNPU job"},
		},
	}
	return tests
}

func TestValidDyVNPUJob(t *testing.T) {
	n := New(PluginName)
	npu, ok := n.(*ascend310P)
	if !ok {
		return
	}
	npu.NPUJob = &util.NPUJob{}
	tests := buildTestValidDyVNPUJobTest()
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			npu.VJob = tt.VJob
			if got := npu.validDyVNPUJob(); !reflect.DeepEqual(got, tt.Want) {
				t.Errorf("ValidDyVNPUJob() got = %#v, want %#v", got, tt.Want)
			}
		})
	}
}

type TestGetAllDyJobsTest struct {
	Name string
	Jobs map[api.JobID]plugin.SchedulerJob
	Want map[api.JobID]plugin.SchedulerJob
}

func buildTestGetAllDyJobsTestCase() []TestGetAllDyJobsTest {
	tests := []TestGetAllDyJobsTest{
		{
			Name: "01-getAllDyJobs will return jobMap when VJob is nil",
			Jobs: map[api.JobID]plugin.SchedulerJob{},
			Want: map[api.JobID]plugin.SchedulerJob{},
		},
		{
			Name: "01-getAllDyJobs will return jobMap when VJob is nil",
			Jobs: map[api.JobID]plugin.SchedulerJob{
				"Job01": {SchedulerJobAttr: util.SchedulerJobAttr{
					NPUJob: &util.NPUJob{VJob: &util.VJob{Type: util.JobTypeDyCut}}}}},
			Want: map[api.JobID]plugin.SchedulerJob{"Job01": {SchedulerJobAttr: util.SchedulerJobAttr{
				NPUJob: &util.NPUJob{VJob: &util.VJob{Type: util.JobTypeDyCut}}}}},
		},
	}
	return tests
}

func TestGetAllDyJobs(t *testing.T) {
	n := New(PluginName)
	npu, ok := n.(*ascend310P)
	if !ok {
		return
	}
	tests := buildTestGetAllDyJobsTestCase()
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			npu.Jobs = tt.Jobs
			if got := npu.getAllDyJobs(); !reflect.DeepEqual(got, tt.Want) {
				t.Errorf("ValidNPUJob() got = %#v, want %#v", got, tt.Want)
			}
		})
	}
}

func TestGetFailedDyTasksFromJobs(t *testing.T) {
	tests := []struct {
		Name  string
		vJobs map[api.JobID]plugin.SchedulerJob
		Want  map[api.TaskID]util.NPUTask
	}{
		{
			Name: "01-getFailedDyTasksFromJobs will return vTask when call this function",
			vJobs: map[api.JobID]plugin.SchedulerJob{
				"vjob01": {
					SchedulerJobAttr: util.SchedulerJobAttr{NPUJob: &util.NPUJob{
						Tasks: map[api.TaskID]util.NPUTask{"Task01": {VTask: &util.VTask{Status: util.TaskStatusFailed}}}}},
				},
			},
			Want: map[api.TaskID]util.NPUTask{"Task01": {VTask: &util.VTask{Status: util.TaskStatusFailed}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			if got := getFailedDyTasksFromJobs(tt.vJobs); !reflect.DeepEqual(got, tt.Want) {
				t.Errorf("GetFailedDyTasksFromJobs() got = %#v, want %#v", got, tt.Want)
			}
		})
	}
}

func TestGetDyFailedNamespaces(t *testing.T) {
	tests := []struct {
		Name string
		VT   map[api.TaskID]util.NPUTask
		Want map[string]struct{}
	}{
		{
			Name: "01-testGetDyFailedNamespaces will return nsMap when when call this function",
			VT: map[api.TaskID]util.NPUTask{
				"task01": {NameSpace: "default"},
				"task02": {NameSpace: "vcjob"},
				"task03": {NameSpace: "kube-system"},
			},
			Want: map[string]struct{}{
				"default":     {},
				"vcjob":       {},
				"kube-system": {},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			if got := getDyFailedNamespaces(tt.VT); !reflect.DeepEqual(got, tt.Want) {
				t.Errorf("GetDyFailedNamespaces() got = %#v, want %#v", got, tt.Want)
			}
		})
	}
}

func TestGetAllDyFailedTasks(t *testing.T) {
	tests := []struct {
		Name  string
		SSN   *framework.Session
		nsMap map[string]struct{}
		Want  []api.TaskID
	}{
		{
			Name: "01-testGetAllDyFailedTasks will return IDs when when call this function",
			SSN:  &framework.Session{},
			nsMap: map[string]struct{}{
				"default":     {},
				"vcjob":       {},
				"kube-system": {},
			},
			Want: []api.TaskID{"0001", "0001", "0001"},
		},
	}

	patch := gomonkey.ApplyFunc(vnpu.GetSegmentFailureTaskIDs,
		func(ssn *framework.Session, namespace string) []api.TaskID {
			return []api.TaskID{"0001"}
		})

	defer patch.Reset()

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			if got := getAllDyFailedTasks(tt.SSN, tt.nsMap); !reflect.DeepEqual(got, tt.Want) {
				t.Errorf("GetAllDyFailedTasks() got = %#v, want %#v", got, tt.Want)
			}
		})
	}
}

func TestGetDyFailedTaskIDsInFaileds(t *testing.T) {
	tests := []struct {
		Name string
		VT   map[api.TaskID]util.NPUTask
		Ids  []api.TaskID
		Want []api.TaskID
	}{
		{
			Name: "01-testGetDyFailedTaskIDsInFaileds will return tIDs when call this function",
			VT: map[api.TaskID]util.NPUTask{
				"task01": {NameSpace: "default"},
				"task02": {NameSpace: "vcjob"},
				"task03": {NameSpace: "kube-system"},
			},
			Ids:  []api.TaskID{"task01", "task02", "task03"},
			Want: []api.TaskID{"task01", "task02", "task03"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			if got := getDyFailedTaskIDsInFaileds(tt.Ids, tt.VT); !reflect.DeepEqual(got, tt.Want) {
				t.Errorf("GetDyFailedTaskIDsInFaileds() got = %#v, want %#v", got, tt.Want)
			}
		})
	}
}

func TestGetDyFailedTasksFromFailed(t *testing.T) {
	tests := []struct {
		Name string
		ssn  *framework.Session
		VT   map[api.TaskID]util.NPUTask
		Want []api.TaskID
	}{
		{
			Name: "01-getDyFailedTasksFromFailed will return taskId when call this function",
			ssn:  &framework.Session{},
			VT: map[api.TaskID]util.NPUTask{
				"task01": {NameSpace: "default"},
			},
			Want: []api.TaskID{"task01"},
		},
	}

	patch := gomonkey.ApplyFunc(vnpu.GetSegmentFailureTaskIDs,
		func(ssn *framework.Session, namespace string) []api.TaskID {
			return []api.TaskID{"task01"}
		})

	defer patch.Reset()

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			if got := getDyFailedTasksFromFailed(tt.ssn, tt.VT); !reflect.DeepEqual(got, tt.Want) {
				t.Errorf("GetDyFailedTasksFromFailed() got = %#v, want %#v", got, tt.Want)
			}
		})
	}
}

type TestGetRestartDyTasksFromJobsTest struct {
	Name string
	VJob map[api.JobID]plugin.SchedulerJob
	ssn  *framework.Session
	Want []util.NPUTask
}

func buildTestGetRestartDyTasksFromJobsTestCase() []TestGetRestartDyTasksFromJobsTest {
	tests := []TestGetRestartDyTasksFromJobsTest{
		{
			Name: "01-GetRestartDyTasksFromJobs will return nil  when vjob is nil",
			VJob: nil,
			ssn:  nil,
			Want: nil,
		},
		{
			Name: "02-GetRestartDyTasksFromJobs will return nil  when fTIDs is 0",
			VJob: nil,
			ssn:  nil,
			Want: nil,
		},
		{
			Name: "03-GetRestartDyTasksFromJobs will return nSlice  when call this method",
			VJob: map[api.JobID]plugin.SchedulerJob{
				"vjob01": {
					SchedulerJobAttr: util.SchedulerJobAttr{NPUJob: &util.NPUJob{
						Tasks: map[api.TaskID]util.NPUTask{
							"task01": {VTask: &util.VTask{Status: util.TaskStatusFailed}}}}}}},
			Want: []util.NPUTask{{VTask: &util.VTask{Status: util.TaskStatusFailed}}},
		},
	}
	return tests

}

func TestGetRestartDyTasksFromJobs(t *testing.T) {
	n := New(PluginName)
	npu, ok := n.(*ascend310P)
	if !ok {
		return
	}
	tests := buildTestGetRestartDyTasksFromJobsTestCase()
	for i, tt := range tests {

		patch2 := gomonkey.ApplyFunc(getDyFailedTasksFromFailed, func(ssn *framework.Session, vT map[api.TaskID]util.NPUTask) []api.TaskID {
			if i == 0 {
				return nil
			}
			return []api.TaskID{"task01"}
		})

		defer patch2.Reset()

		t.Run(tt.Name, func(t *testing.T) {
			if got := npu.getRestartDyTasksFromJobs(tt.VJob, tt.ssn); !reflect.DeepEqual(got, tt.Want) {
				t.Errorf("GetAllDyFailedTasks() got = %#v, want %#v", got, tt.Want)
			}
		})
	}

}
