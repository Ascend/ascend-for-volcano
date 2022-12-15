/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package ascend310 is using for HuaWei A800/9000 Ascend910 pin affinity schedule.
*/
package ascend310

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/ascend310/chip310x4"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/base"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/rescheduling"
	itest "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/test"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
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
		if npu.GetAnnoName() != util.NPU310CardName {
			t.Errorf("New() npu annoName: %s, wantAnnoName: %s.", npu.GetPluginName(), util.NPU310CardName)
		}
		if npu.GetAnnoPreVal() != util.NPU310CardNamePre {
			t.Errorf("New() npu annoNamePre: %s, wantAnnoNamePre: %s.",
				npu.GetPluginName(), util.NPU310CardNamePre)
		}
	})
}

type initMyJobPluginTestCase struct {
	name    string
	attr    util.SchedulerJobAttr
	env     plugin.ScheduleEnv
	handler base.AscendHandler
	wantErr error
}

func buildInitMyJobPluginTestCases() []initMyJobPluginTestCase {
	return []initMyJobPluginTestCase{
		{
			name: "01-InitMyJobPlugin return nil when define accelerator the handler will be define as card",
			attr: util.SchedulerJobAttr{
				ComJob: util.ComJob{Label: map[string]string{Accelerator310Key: Card310AcceleratorValue}},
				NPUJob: &util.NPUJob{ReqNPUName: util.NPU310CardName},
			},
			env:     plugin.ScheduleEnv{},
			wantErr: nil,
		},
		{
			name: "02-InitMyJobPlugin return nil when not define accelerator the handler will be define as chip",
			attr: util.SchedulerJobAttr{
				ComJob: util.ComJob{Label: map[string]string{}},
				NPUJob: &util.NPUJob{ReqNPUName: util.NPU310CardName},
			},
			env:     plugin.ScheduleEnv{},
			handler: chip310x4.New(chip310x4.SchedulerName),
			wantErr: nil,
		},
		{
			name: "03-InitMyJobPlugin return error when not define accelerator the handler will be define as chip",
			attr: util.SchedulerJobAttr{
				ComJob: util.ComJob{Label: map[string]string{}},
				NPUJob: &util.NPUJob{ReqNPUName: util.NPU310PCardName},
			},
			env:     plugin.ScheduleEnv{},
			handler: chip310x4.New(chip310x4.SchedulerName),
			wantErr: fmt.Errorf("not support %s", util.NPU310PCardName+Chip310AcceleratorValue),
		},
	}
}

func TestInitMyJobPlugin(t *testing.T) {
	npu := New(PluginName)
	testCases := buildInitMyJobPluginTestCases()
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			err := npu.InitMyJobPlugin(tt.attr, tt.env)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("ValidNPUJob() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func buildCheckNodeNPUByTaskTestCases() []itest.CheckNodeNPUByTaskTestCase {
	return []itest.CheckNodeNPUByTaskTestCase{
		{
			Name: "01-CheckNodeNPUByDyTask return err when task is nil",
			Task: nil,
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Name:       "node1",
					Annotation: map[string]string{util.NPU310CardName: "Ascend310-0,Ascend310-1"},
				},
			},
			WantErr: errors.New(util.ArgumentError),
		},
		{
			Name: "02-CheckNodeNPUByDyTask return err when node annotation is nil",
			Task: test.FakeTaskWithResReq("pod1", util.NPU310CardName, util.NPUIndex2),
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Name:       "node1",
					Annotation: nil,
				},
			},
			WantErr: errors.New(util.ArgumentError),
		},
	}
}

// TestCheckNodeNPUByTask
func TestCheckNodeNPUByTask(t *testing.T) {
	npu := New(PluginName)
	testCases := buildCheckNodeNPUByTaskTestCases()
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			if err := npu.CheckNodeNPUByTask(tt.Task, tt.Node); !reflect.DeepEqual(err, tt.WantErr) {
				t.Errorf("CheckNodeNPUByDyTask() error = %v, wantErr %v", err, tt.WantErr)
			}
		})
	}
}

func buildScoreBestNPUNodesTestCases() []itest.ScoreBestNPUNodesTestCase {
	return []itest.ScoreBestNPUNodesTestCase{
		{
			Name:     "01-ScoreBestNPUNodes return err when task is nil",
			Task:     nil,
			Nodes:    []*api.NodeInfo{test.FakeNormalTestNode("node1")},
			ScoreMap: map[string]float64{"node1": 0},
			WantSMap: map[string]float64{"node1": 0},
			WantErr:  errors.New(util.ArgumentError),
		},
		{
			Name:     "02-ScoreBestNPUNodes return err when nodes is empty",
			Task:     test.FakeNormalTestTask("pod1", "node1", "vcjob"),
			Nodes:    []*api.NodeInfo{},
			ScoreMap: map[string]float64{"node1": 0},
			WantSMap: map[string]float64{"node1": 0},
			WantErr:  errors.New(util.ArgumentError),
		},
		{
			Name:     "03-ScoreBestNPUNodes return err when scoreMap is empty",
			Task:     test.FakeNormalTestTask("pod1", "node1", "vcjob"),
			Nodes:    []*api.NodeInfo{test.FakeNormalTestNode("node1")},
			ScoreMap: map[string]float64{},
			WantSMap: map[string]float64{},
			WantErr:  errors.New(util.ArgumentError),
		},
	}
}

// TestCheckNodeNPUByTask
func TestScoreBestNPUNodes(t *testing.T) {
	npu := New(PluginName)
	testCases := buildScoreBestNPUNodesTestCases()
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			if err := npu.ScoreBestNPUNodes(tt.Task, tt.Nodes, tt.ScoreMap); !reflect.DeepEqual(err, tt.WantErr) {
				t.Errorf("CheckNodeNPUByDyTask() error = %v, wantErr %v", err, tt.WantErr)
			}
		})
	}
}

func buildUseAnnotationTestCases() []itest.UseAnnotationTestCase {
	return []itest.UseAnnotationTestCase{
		{
			Name: "01-ScoreBestNPUNodes return nil when task is nil",
			Task: nil,
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Annotation: map[string]string{util.NPU310CardName: "Ascend310-0,Ascend310-1"},
				},
			},
			WantNode: nil,
		},
		{
			Name: "02-ScoreBestNPUNodes return nil when node annotation is nil",
			Task: test.FakeNormalTestTask("pod1", "node1", "vcjob"),
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Annotation: nil,
				},
			},
			WantNode: nil,
		},
	}
}

// TestUseAnnotation
func TestUseAnnotation(t *testing.T) {
	npu := New(PluginName)
	testCases := buildUseAnnotationTestCases()
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			if err := npu.UseAnnotation(tt.Task, tt.Node); !reflect.DeepEqual(err, tt.WantNode) {
				t.Errorf("CheckNodeNPUByDyTask() error = %v, wantErr %v", err, tt.WantNode)
			}
		})
	}
}

type ascend310Fields struct {
	reHandle   *rescheduling.ReScheduler
	NPUHandler base.NPUHandler
	Kind       map[string]base.AscendHandler
	handle     base.AscendHandler
}

type ascend310PreStartActionArgs struct {
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

type ascend310PreStartActionTests struct {
	name    string
	fields  ascend310Fields
	args    ascend310PreStartActionArgs
	wantErr bool
}

func buildAscend310PreStartActionTest1() ascend310PreStartActionTests {
	test1 := ascend310PreStartActionTests{
		name: "01-Ascend310PreStartAction()-nil tp",
		fields: ascend310Fields{
			NPUHandler: base.NPUHandler{},
			reHandle:   nil,
		},
		args: ascend310PreStartActionArgs{
			ssn:      nil,
			addCache: false,
		},
		wantErr: true,
	}
	return test1
}

func buildAscend310PreStartActionTest2() ascend310PreStartActionTests {
	test2 := ascend310PreStartActionTests{
		name: "02-Ascend310PreStartAction()-off rescheduling",
		fields: ascend310Fields{
			NPUHandler: base.NPUHandler{},
			reHandle:   nil,
		},
		args: ascend310PreStartActionArgs{
			ssn:      test.FakeSSNReSchedule(),
			addCache: false,
		},
		wantErr: true,
	}
	test2.fields.NPUHandler.SchedulerJobAttr.Label = map[string]string{rescheduling.
		JobRescheduleLabelKey: rescheduling.JobOffRescheduleLabelValue}
	return test2
}

func buildAscend310PreStartActionTest3() ascend310PreStartActionTests {
	var tmpPatch1 *gomonkey.Patches
	var tmpPatch2 *gomonkey.Patches
	var tmpPatch3 *gomonkey.Patches
	var tmpPatch4 *gomonkey.Patches
	var tmpPatch5 *gomonkey.Patches
	var tmpPatch6 *gomonkey.Patches
	var tmpPatch7 *gomonkey.Patches
	var tmpPatch8 *gomonkey.Patches
	var tmpPatch9 *gomonkey.Patches
	test3 := ascend310PreStartActionTests{
		name: "03-Ascend310PreStartAction()-success",
		fields: ascend310Fields{
			NPUHandler: base.NPUHandler{},
			reHandle:   nil,
		},
		args: ascend310PreStartActionArgs{
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

func buildAscend310PreStartActionTests() []ascend310PreStartActionTests {
	tests := []ascend310PreStartActionTests{
		buildAscend310PreStartActionTest1(),
		buildAscend310PreStartActionTest2(),
		buildAscend310PreStartActionTest3(),
	}
	return tests
}

// TestAscend310PreStartAction test for preStartAction
func TestAscend310PreStartAction(t *testing.T) {
	tests := buildAscend310PreStartActionTests()
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
			tp := &asend310{
				NPUHandler: tt.fields.NPUHandler,
				Kind:       tt.fields.Kind,
				handle:     tt.fields.handle,
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
