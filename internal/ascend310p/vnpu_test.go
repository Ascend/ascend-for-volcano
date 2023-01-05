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

	"volcano.sh/volcano/pkg/scheduler/api"

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
				t.Errorf("ValidNPUJob() error = %#v, wantErr %#v", err, tt.WantErr)
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
				t.Errorf("ValidNPUJob() error = %#v, wantErr %#v", err, tt.WantErr)
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
				t.Errorf("ValidNPUJob() error = %#v, wantErr %#v", err, tt.WantErr)
			}
		})
	}
}
