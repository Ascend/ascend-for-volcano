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
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/base"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/rescheduling"
	itest "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/test"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// TestNew
func TestNew(t *testing.T) {
	t.Run("test New", func(t *testing.T) {
		npu := New(PluginName)
		if npu.GetPluginName() != PluginName {
			t.Errorf("New() npu Name: %s, wantName: %s.", npu.GetPluginName(), PluginName)
		}
		if npu.GetAnnoName() != util.NPU310PCardName {
			t.Errorf("New() npu annoName: %s, wantAnnoName: %s.", npu.GetPluginName(), util.NPU310PCardName)
		}
		if npu.GetAnnoPreVal() != util.NPU310PCardNamePre {
			t.Errorf("New() npu annoNamePre: %s, wantAnnoNamePre: %s.",
				npu.GetPluginName(), util.NPU310PCardNamePre)
		}
	})
}

type ascend310pFields struct {
	reHandle   *rescheduling.ReScheduler
	NPUHandler base.NPUHandler
}

type ascend310pPreStartActionArgs struct {
	ssn              *framework.Session
	addCache         bool
	cacheFuncBefore1 func()
	cacheFuncBefore2 func()
	cacheFuncBefore3 func()
	cacheFuncBefore4 func()
	cacheFuncBefore5 func()
	cacheFuncBefore6 func()
	cacheFuncBefore7 func()
	cacheFuncBefore8 func()
	cacheFuncBefore9 func()
	cacheFuncAfter1  func()
	cacheFuncAfter2  func()
	cacheFuncAfter3  func()
	cacheFuncAfter4  func()
	cacheFuncAfter5  func()
	cacheFuncAfter6  func()
	cacheFuncAfter7  func()
	cacheFuncAfter8  func()
	cacheFuncAfter9  func()
}

type ascend310pPreStartActionTests struct {
	name    string
	fields  ascend310pFields
	args    ascend310pPreStartActionArgs
	wantErr bool
}

func buildAscend310pPreStartActionTest1() ascend310pPreStartActionTests {
	test1 := ascend310pPreStartActionTests{
		name: "01-Ascend310pPreStartAction()-nil tp",
		fields: ascend310pFields{
			NPUHandler: base.NPUHandler{},
			reHandle:   nil,
		},
		args: ascend310pPreStartActionArgs{
			ssn:      nil,
			addCache: false,
		},
		wantErr: true,
	}
	return test1
}

func buildAscend310pPreStartActionTest2() ascend310pPreStartActionTests {
	test2 := ascend310pPreStartActionTests{
		name: "02-Ascend310pPreStartAction()-off rescheduling",
		fields: ascend310pFields{
			NPUHandler: base.NPUHandler{},
			reHandle:   nil,
		},
		args: ascend310pPreStartActionArgs{
			ssn:      test.FakeSSNReSchedule(),
			addCache: false,
		},
		wantErr: true,
	}
	test2.fields.NPUHandler.SchedulerJobAttr.Label = map[string]string{rescheduling.
		JobRescheduleLabelKey: rescheduling.JobOffRescheduleLabelValue}
	return test2
}

func buildAscend310pPreStartActionTest3() ascend310pPreStartActionTests {
	var tmpPatch1 *gomonkey.Patches
	var tmpPatch2 *gomonkey.Patches
	var tmpPatch3 *gomonkey.Patches
	var tmpPatch4 *gomonkey.Patches
	var tmpPatch5 *gomonkey.Patches
	var tmpPatch6 *gomonkey.Patches
	var tmpPatch7 *gomonkey.Patches
	var tmpPatch8 *gomonkey.Patches
	var tmpPatch9 *gomonkey.Patches
	test3 := ascend310pPreStartActionTests{
		name: "03-Ascend310pPreStartAction()-success",
		fields: ascend310pFields{
			NPUHandler: base.NPUHandler{},
			reHandle:   nil,
		},
		args: ascend310pPreStartActionArgs{
			ssn:              test.FakeSSNReSchedule(),
			addCache:         true,
			cacheFuncBefore1: func() { tmpPatch1 = itest.PatchNew() },
			cacheFuncBefore2: func() { tmpPatch2 = itest.PatchNewComRes() },
			cacheFuncBefore3: func() { tmpPatch3 = itest.PatchSynNode() },
			cacheFuncBefore4: func() { tmpPatch4 = itest.PatchAddNode() },
			cacheFuncBefore5: func() { tmpPatch5 = itest.PatchSynJob() },
			cacheFuncBefore6: func() { tmpPatch6 = itest.PatchForce() },
			cacheFuncBefore7: func() { tmpPatch7 = itest.PatchGetRun() },
			cacheFuncBefore8: func() { tmpPatch8 = itest.PatchAddJob() },
			cacheFuncBefore9: func() { tmpPatch9 = itest.PatchRestart() },
			cacheFuncAfter1:  func() { itest.PatchReset(tmpPatch1) },
			cacheFuncAfter2:  func() { itest.PatchReset(tmpPatch2) },
			cacheFuncAfter3:  func() { itest.PatchReset(tmpPatch3) },
			cacheFuncAfter4:  func() { itest.PatchReset(tmpPatch4) },
			cacheFuncAfter5:  func() { itest.PatchReset(tmpPatch5) },
			cacheFuncAfter6:  func() { itest.PatchReset(tmpPatch6) },
			cacheFuncAfter7:  func() { itest.PatchReset(tmpPatch7) },
			cacheFuncAfter8:  func() { itest.PatchReset(tmpPatch8) },
			cacheFuncAfter9:  func() { itest.PatchReset(tmpPatch9) },
		},
		wantErr: true,
	}
	test3.fields.NPUHandler.SchedulerJobAttr.Label = map[string]string{rescheduling.
		JobRescheduleLabelKey: rescheduling.JobGraceRescheduleLabelValue}
	return test3
}

func buildAscend310pPreStartActionTests() []ascend310pPreStartActionTests {

	tests := []ascend310pPreStartActionTests{
		buildAscend310pPreStartActionTest1(),
		buildAscend310pPreStartActionTest2(),
		buildAscend310pPreStartActionTest3(),
	}
	return tests
}

// TestAscend310pPreStartAction test for preStartAction
func TestAscend310pPreStartAction(t *testing.T) {
	tests := buildAscend310pPreStartActionTests()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.args.addCache {
				tt.args.cacheFuncBefore1()
				tt.args.cacheFuncBefore2()
				tt.args.cacheFuncBefore3()
				tt.args.cacheFuncBefore4()
				tt.args.cacheFuncBefore5()
				tt.args.cacheFuncBefore6()
				tt.args.cacheFuncBefore7()
				tt.args.cacheFuncBefore8()
				tt.args.cacheFuncBefore9()
			}
			tp := &ascend310P{
				NPUHandler: tt.fields.NPUHandler,
				reHandle:   tt.fields.reHandle,
			}
			if err := tp.PreStartAction(tt.args.ssn); (err != nil) != tt.wantErr {
				t.Errorf("PreStartAction() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.args.addCache {
				tt.args.cacheFuncAfter1()
				tt.args.cacheFuncAfter2()
				tt.args.cacheFuncAfter3()
				tt.args.cacheFuncAfter4()
				tt.args.cacheFuncAfter5()
				tt.args.cacheFuncAfter6()
				tt.args.cacheFuncAfter7()
				tt.args.cacheFuncAfter8()
				tt.args.cacheFuncAfter9()
			}
		})
	}
}
