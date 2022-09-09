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

	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/ascend310/chip310x4"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/base"
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
			Name: "01-CheckNodeNPUByTask return err when task is nil",
			Task: nil,
			Node: plugin.NPUNode{
				Name:       "node1",
				Annotation: map[string]string{util.NPU310CardName: "Ascend310-0,Ascend310-1"},
			},
			WantErr: errors.New(util.ArgumentError),
		},
		{
			Name: "02-CheckNodeNPUByTask return err when node annotation is nil",
			Task: test.FakeTaskWithResReq("pod1", util.NPU310CardName, util.NPUIndex2),
			Node: plugin.NPUNode{
				Name:       "node1",
				Annotation: nil,
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
				t.Errorf("CheckNodeNPUByTask() error = %v, wantErr %v", err, tt.WantErr)
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
				t.Errorf("CheckNodeNPUByTask() error = %v, wantErr %v", err, tt.WantErr)
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
				Annotation: map[string]string{util.NPU310CardName: "Ascend310-0,Ascend310-1"},
			},
			WantNode: nil,
		},
		{
			Name:     "02-ScoreBestNPUNodes return nil when node annotation is nil",
			Task:     test.FakeNormalTestTask("pod1", "node1", "vcjob"),
			Node:     plugin.NPUNode{Annotation: nil},
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
				t.Errorf("CheckNodeNPUByTask() error = %v, wantErr %v", err, tt.WantNode)
			}
		})
	}
}