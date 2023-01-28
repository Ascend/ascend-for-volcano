/*
Copyright(C)2020-2023. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package util is using for the total variable.
*/
package util

import (
	"errors"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"k8s.io/api/core/v1"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"
)

type GetRealPodByTaskTest struct {
	name    string
	npuTask *NPUTask
	ssn     *framework.Session
	want    *v1.Pod
	wantErr bool
}

func buildGetRealPodByTaskTestCase01() []GetRealPodByTaskTest {
	tests := []GetRealPodByTaskTest{
		{
			name:    "01-GetRealPodByTask will return err when asTask is nil",
			npuTask: nil,
			ssn:     nil,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "02-GetRealPodByTask will return err when ssn is nil",
			npuTask: &NPUTask{ReqNPUNum: 1},
			ssn:     nil,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "03-GetRealPodByTask will return err when ReqNPUName is nil",
			npuTask: &NPUTask{ReqNPUNum: 1},
			ssn:     &framework.Session{},
			want:    nil,
			wantErr: true,
		},
	}
	return tests
}

func TestGetRealPodByTask(t *testing.T) {
	tests := buildGetRealPodByTaskTestCase01()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.npuTask.GetRealPodByTask(tt.ssn)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRealPodByTask() error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetRealPodByTask() got = %#v, want %#v", got, tt.want)
			}
		})
	}
}

type DeleteRealPodByTaskTest struct {
	name     string
	npuTask  *NPUTask
	ssn      *framework.Session
	waitTime int64
	wantErr  bool
}

func buildDeleteRealPodByTaskTestCase01() DeleteRealPodByTaskTest {
	test01 := DeleteRealPodByTaskTest{
		name:    "01-DeleteRealPodByTaskTest will return err when task is nil",
		wantErr: true,
	}
	return test01
}

func buildDeleteRealPodByTaskTestCase02() DeleteRealPodByTaskTest {
	test01 := DeleteRealPodByTaskTest{
		name:    "02-DeleteRealPodByTaskTest will return err when POD is nil",
		npuTask: &NPUTask{ReqNPUNum: 1, Name: "task01"},
		ssn: &framework.Session{Jobs: map[api.JobID]*api.JobInfo{"job01": {Tasks: map[api.TaskID]*api.TaskInfo{
			"task01": {Name: "task01", Namespace: "default"}}}}},
		wantErr: true,
	}
	return test01
}

func buildDeleteRealPodByTaskTestCase03() DeleteRealPodByTaskTest {
	test01 := DeleteRealPodByTaskTest{
		name:    "03-DeleteRealPodByTaskTest will return err when ssn is nil",
		npuTask: &NPUTask{ReqNPUNum: 1, Name: "task01"},
		ssn:     nil,
		wantErr: true,
	}
	return test01
}

func buildDeleteRealPodByTaskTestCase() []DeleteRealPodByTaskTest {
	tests := []DeleteRealPodByTaskTest{
		buildDeleteRealPodByTaskTestCase01(),
		buildDeleteRealPodByTaskTestCase02(),
		buildDeleteRealPodByTaskTestCase03(),
	}
	return tests
}

func TestDeleteRealPodByTask(t *testing.T) {
	tests := buildDeleteRealPodByTaskTestCase()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.npuTask.DeleteRealPodByTask(tt.ssn, tt.waitTime); (err != nil) != tt.wantErr {
				t.Errorf("DeleteRealPodByTask() error = %#v, wantErr %#v", err, tt.wantErr)
			}
		})
	}
}

type EvictJobByTaskTest struct {
	name     string
	ssn      *framework.Session
	taskName string
	reason   string
	asTask   *NPUTask
	wantErr  bool
}

func buildEvictJobByTaskTestCase01() EvictJobByTaskTest {
	test01 := EvictJobByTaskTest{
		name:    "01-EvictJobByTaskTest will return err when Task is nil",
		asTask:  nil,
		wantErr: true,
	}
	return test01
}

func buildEvictJobByTaskTestCase02() EvictJobByTaskTest {
	test02 := EvictJobByTaskTest{
		name:    "02-EvictJobByTaskTest will return err when ssn is nil",
		asTask:  &NPUTask{ReqNPUNum: 1, Name: "task01"},
		ssn:     nil,
		wantErr: true,
	}
	return test02
}

func buildEvictJobByTaskTestCase03() EvictJobByTaskTest {
	test03 := EvictJobByTaskTest{
		name:   "03-EvictJobByTaskTest will return err when ssn is nil",
		asTask: &NPUTask{ReqNPUNum: 1, Name: "task01"},
		ssn: &framework.Session{Jobs: map[api.JobID]*api.JobInfo{"job01": {Tasks: map[api.TaskID]*api.TaskInfo{
			"task01": {Name: "task01",
				Namespace: "default",
				Pod: &v1.Pod{Status: v1.PodStatus{
					Conditions: []v1.PodCondition{{Message: "PodCondition-message"}}}}}}}}},
		taskName: "task01",
		wantErr:  true,
	}
	return test03
}

func buildEvictJobByTaskTestCase() []EvictJobByTaskTest {
	tests := []EvictJobByTaskTest{
		buildEvictJobByTaskTestCase01(),
		buildEvictJobByTaskTestCase02(),
		buildEvictJobByTaskTestCase03(),
	}
	return tests
}

func TestEvictJobByTask(t *testing.T) {
	tests := buildEvictJobByTaskTestCase()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patch := gomonkey.ApplyMethod(reflect.TypeOf(tt.ssn),
				"Evict", func(*framework.Session, *api.TaskInfo, string) error {
					return errors.New("mock error Evict")
				})

			defer patch.Reset()

			if err := tt.asTask.EvictJobByTask(tt.ssn, tt.reason, tt.taskName); (err != nil) != tt.wantErr {
				t.Errorf("EvictJobByTask() error = %#v, wantErr %#v", err, tt.wantErr)
			}
		})
	}
}

func TestForceDeletePodByTaskInf(t *testing.T) {
	type args struct {
		ssn      *framework.Session
		reason   string
		nodeName string
	}
	tests := []struct {
		name    string
		asTask  *NPUTask
		args    args
		wantErr bool
	}{
		{
			name: "01-ForceDeletePodByTaskInf will return err when ssn is nil",
			asTask: &NPUTask{
				Name: "task01",
				VTask: &VTask{
					Allocated: TaskAllocated{NodeName: "huawei"},
				},
			},
			args:    args{ssn: nil},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := tt.asTask.ForceDeletePodByTaskInf(tt.args.ssn,
				tt.args.reason, tt.args.nodeName); (err != nil) != tt.wantErr {
				t.Errorf("ForceDeletePodByTaskInf() error = %#v, wantErr %#v", err, tt.wantErr)
			}
		})
	}
}

type IsTaskInItsNodeTest struct {
	name     string
	asTask   *NPUTask
	ssn      *framework.Session
	nodeName string
	want     bool
}

func buildIsTaskInItsNodeTestCase() []IsTaskInItsNodeTest {
	tests := []IsTaskInItsNodeTest{
		{
			name: "01-IsTaskInItsNode will return false when Nodes is nil",
			asTask: &NPUTask{
				Name: "task01",
				VTask: &VTask{
					Allocated: TaskAllocated{NodeName: "huawei"},
				},
			},
			ssn:  &framework.Session{Nodes: map[string]*api.NodeInfo{}},
			want: false,
		},
		{
			name: "02-IsTaskInItsNode will return false when NodeInfo is nil",
			asTask: &NPUTask{
				Name: "task01",
				VTask: &VTask{
					Allocated: TaskAllocated{NodeName: "huawei"},
				},
			},
			ssn:      &framework.Session{Nodes: map[string]*api.NodeInfo{"huawei": {}}},
			nodeName: "huawei",
			want:     false,
		},
	}
	return tests
}

func TestIsTaskInItsNode(t *testing.T) {
	tests := buildIsTaskInItsNodeTestCase()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.asTask.IsTaskInItsNode(tt.ssn, tt.nodeName); got != tt.want {
				t.Errorf("IsTaskInItsNode() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
