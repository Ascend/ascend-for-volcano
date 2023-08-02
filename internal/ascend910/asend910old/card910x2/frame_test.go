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
Package card910x2 is using for HuaWei Ascend pin affinity schedule.
*/
package card910x2

import (
	"errors"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"k8s.io/api/core/v1"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/base"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/rescheduling"
	itest "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/test"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

type validNPUJobTestCase struct {
	wantErr *api.ValidateResult
	name    string
	attr    util.SchedulerJobAttr
}

const (
	score56 = 56
	score64 = 64
)

func buildValidNPUJobTestCase01() []validNPUJobTestCase {
	job01 := test.FakeNormalTestJob("job01", 1)
	test.SetFakeJobResRequest(job01, util.NPU910CardName, "0")
	attr1 := itest.FakeSchedulerJobAttrByJob(job01)
	job02 := test.FakeNormalTestJob("job02", 1)
	test.SetFakeJobResRequest(job02, util.NPU910CardName, "3")
	attr2 := itest.FakeSchedulerJobAttrByJob(job02)
	job03 := test.FakeNormalTestJob("job02", 1)
	test.SetFakeJobResRequest(job03, util.NPU910CardName, "2")
	attr3 := itest.FakeSchedulerJobAttrByJob(job03)
	return []validNPUJobTestCase{
		{
			name: "01-ValidNPUJob should return error when job request no npu",
			attr: attr1,
			wantErr: &api.ValidateResult{
				Pass:    false,
				Reason:  "huawei.com/Ascend910card checkSingleTrainMode vcjob/job01 req npu not in [1,2]",
				Message: "huawei.com/Ascend910card checkSingleTrainMode vcjob/job01 req npu not in [1,2]",
			},
		},
		{
			name: "02-ValidNPUJob should return error when task request npu more than 3",
			attr: attr2,
			wantErr: &api.ValidateResult{
				Pass:    false,
				Reason:  "huawei.com/Ascend910card checkSingleTrainMode vcjob/job02 req npu not in [1,2]",
				Message: "huawei.com/Ascend910card checkSingleTrainMode vcjob/job02 req npu not in [1,2]",
			},
		},
		{
			name:    "03-ValidNPUJob should return nil when tasks request is valid",
			attr:    attr3,
			wantErr: nil,
		},
	}
}

func buildValidNPUJobTestCase02() []validNPUJobTestCase {
	job04 := test.FakeNormalTestJob("job04", util.NPUIndex2)
	test.SetFakeJobResRequest(job04, util.NPU910CardName, "0")
	attr4 := itest.FakeSchedulerJobAttrByJob(job04)
	task := util.NPUTask{ReqNPUNum: 1}
	attr4.Tasks["vcjob-pod1"] = task
	job05 := test.FakeNormalTestJob("job05", util.NPUIndex2)
	test.SetFakeJobResRequest(job05, util.NPU910CardName, "1")
	attr5 := itest.FakeSchedulerJobAttrByJob(job05)
	job06 := test.FakeNormalTestJob("job06", util.NPUIndex2)
	test.SetFakeJobResRequest(job06, util.NPU910CardName, "2")
	attr6 := itest.FakeSchedulerJobAttrByJob(job06)
	return []validNPUJobTestCase{
		{
			name: "04-ValidNPUJob should return error when task request no npu",
			attr: attr4,
			wantErr: &api.ValidateResult{
				Pass:    false,
				Reason:  "checkModuleDistributeTrainMode pod0 req 0 not in [1,2,4,8]",
				Message: "checkModuleDistributeTrainMode pod0 req 0 not in [1,2,4,8]",
			},
		},
		{
			name:    "05-ValidNPUJob should return error when task request npu more than 4",
			attr:    attr5,
			wantErr: nil,
		},
		{
			name:    "06-ValidNPUJob should return nil when tasks request is valid",
			attr:    attr6,
			wantErr: nil,
		},
	}
}

func TestValidNPUJob(t *testing.T) {
	npu := New(SchedulerName)
	testCases := buildValidNPUJobTestCase01()
	testCases = append(testCases, buildValidNPUJobTestCase02()...)
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			npu.SetSchedulerAttr(tt.attr)
			if err := npu.ValidNPUJob(); !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("ValidNPUJob() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func buildCheckNodeNPUByTaskTestCase1() itest.CheckNodeNPUByTaskTestCase {
	return itest.CheckNodeNPUByTaskTestCase{
		Name: "01-CheckNodeNPUByTask when return nil node npu meet task req",
		Task: test.FakeTaskWithResReq("pod0", util.NPU910CardName, util.NPUIndex2),
		Node: plugin.NPUNode{
			CommonNode: plugin.CommonNode{
				Name:       "node1",
				Annotation: map[string]string{util.NPU910CardName: "Ascend910-0,Ascend910-1"},
			},
		},
		WantErr: nil,
	}
}

func buildCheckNodeNPUByTaskTestCase2() itest.CheckNodeNPUByTaskTestCase {
	return itest.CheckNodeNPUByTaskTestCase{
		Name: "02-CheckNodeNPUByTask return err when task is not npu task",
		Task: test.FakeTaskWithResReq("pod1", util.NPU910CardName, util.NPUIndex2),
		Node: plugin.NPUNode{
			CommonNode: plugin.CommonNode{
				Name:       "node1",
				Annotation: map[string]string{util.NPU910CardName: "Ascend910-0,Ascend910-1"},
			},
		},
		WantErr: errors.New("task<pod1> is not npu task"),
	}
}

func buildCheckNodeNPUByTaskTestCase3() itest.CheckNodeNPUByTaskTestCase {
	return itest.CheckNodeNPUByTaskTestCase{
		Name: "03-CheckNodeNPUByTask return err when node has no req npu",
		Task: test.FakeTaskWithResReq("pod0", util.NPU910CardName, util.NPUIndex2),
		Node: plugin.NPUNode{
			CommonNode: plugin.CommonNode{
				Name:       "node1",
				Annotation: map[string]string{util.NPU310PCardName: "Ascend910-0,Ascend910-1"},
			},
		},
		WantErr: errors.New("getUsableTopFromNode node1 don't have huawei.com/Ascend910"),
	}
}

func buildCheckNodeNPUByTaskTestCase4() itest.CheckNodeNPUByTaskTestCase {
	return itest.CheckNodeNPUByTaskTestCase{
		Name: "04-CheckNodeNPUByTask return err when node has no req npu",
		Task: test.FakeTaskWithResReq("pod0", util.NPU910CardName, util.NPUIndex2),
		Node: plugin.NPUNode{
			CommonNode: plugin.CommonNode{
				Name:       "node1",
				Annotation: map[string]string{util.NPU910CardName: "Ascend910-0, Ascend910-1"},
			},
		},
		WantErr: errors.New("node <node1> don't have enough resource <huawei.com/Ascend910>, req<2>, idle<0>"),
	}
}

func buildCheckNodeNPUByTaskTestCase5() itest.CheckNodeNPUByTaskTestCase {
	return itest.CheckNodeNPUByTaskTestCase{
		Name: "05-CheckNodeNPUByTask return err when node has no req npu",
		Task: test.FakeTaskWithResReq("pod0", util.NPU910CardName, util.NPUIndex2),
		Node: plugin.NPUNode{
			CommonNode: plugin.CommonNode{
				Name:       "node1",
				Annotation: map[string]string{util.NPU910CardName: "Ascend910-0"},
			},
		},
		WantErr: errors.New("node <node1> don't have enough resource <huawei.com/Ascend910>, req<2>, idle<1>"),
	}
}

func buildCheckNodeNPUByTaskTestCases() []itest.CheckNodeNPUByTaskTestCase {
	return []itest.CheckNodeNPUByTaskTestCase{
		buildCheckNodeNPUByTaskTestCase1(),
		buildCheckNodeNPUByTaskTestCase2(),
		buildCheckNodeNPUByTaskTestCase3(),
		buildCheckNodeNPUByTaskTestCase4(),
		buildCheckNodeNPUByTaskTestCase5(),
	}
}

// TestCheckNodeNPUByTask
func TestCheckNodeNPUByTask(t *testing.T) {
	npu := New(SchedulerName)
	job := test.FakeNormalTestJob("job", 1)
	test.SetFakeJobResRequest(job, util.NPU910CardName, "2")
	attr := itest.FakeSchedulerJobAttrByJob(job)
	sJob := plugin.SchedulerJob{}
	sJob.SchedulerJobAttr = attr
	env := plugin.ScheduleEnv{
		Jobs: map[api.JobID]plugin.SchedulerJob{job.UID: sJob},
	}
	npu.SetSchedulerAttr(attr)
	npu.SetSchedulerEnv(env)
	testCases := buildCheckNodeNPUByTaskTestCases()
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			if err := npu.CheckNodeNPUByTask(tt.Task, tt.Node); !reflect.DeepEqual(err, tt.WantErr) {
				t.Errorf("ValidNPUJob() error = %v, wantErr %v", err, tt.WantErr)
			}
		})
	}
}

type scoreBestNPUNodesTestCase struct {
	name     string
	attr     util.SchedulerJobAttr
	task     *api.TaskInfo
	nodes    []*api.NodeInfo
	scoreMap map[string]float64
	wantSMap map[string]float64
	wantErr  error
}

func buildScoreBestNPUNodesTestCases01() []scoreBestNPUNodesTestCase {
	job1 := test.FakeNormalTestJob("job01", 1)
	test.SetFakeJobResRequest(job1, util.NPU910CardName, "1")
	attr1 := itest.FakeSchedulerJobAttrByJob(job1)
	task1 := test.FakeTaskWithResReq("pod0", util.NPU910CardName, 1)
	nodes := []*api.NodeInfo{{Name: "node1"}, {Name: "node2"}}

	return []scoreBestNPUNodesTestCase{
		{
			name:     "01-ScoreBestNPUNodes return err when task is not my task",
			attr:     attr1,
			task:     test.FakeTaskWithResReq("pod1", util.NPU910CardName, 1),
			nodes:    nodes,
			scoreMap: map[string]float64{"node1": 0, "node2": 0},
			wantSMap: map[string]float64{"node1": 0, "node2": 0},
			wantErr:  errors.New("task<pod1> is not npu task"),
		},
		{
			name:     "02-ScoreBestNPUNodes return error when node npu is invalid",
			attr:     attr1,
			task:     task1,
			nodes:    []*api.NodeInfo{{Name: "node3"}},
			scoreMap: map[string]float64{"node3": 0},
			wantSMap: map[string]float64{"node3": 0},
			wantErr:  nil,
		},
		{
			name:     "03-ScoreBestNPUNodes return nil when node npu meet task req",
			attr:     attr1,
			task:     task1,
			nodes:    nodes,
			scoreMap: map[string]float64{"node1": 0, "node2": 0},
			wantSMap: map[string]float64{"node1": score64, "node2": score56},
			wantErr:  nil,
		},
	}
}

func buildScoreBestNPUNodesTestCases02() []scoreBestNPUNodesTestCase {
	job2 := test.FakeNormalTestJob("job01", util.NPUIndex2)
	test.SetFakeJobResRequest(job2, util.NPU910CardName, "2")
	attr2 := itest.FakeSchedulerJobAttrByJob(job2)
	task2 := test.FakeTaskWithResReq("pod1", util.NPU910CardName, util.NPUIndex2)
	nodes := []*api.NodeInfo{{Name: "node1"}, {Name: "node2"}}
	return []scoreBestNPUNodesTestCase{
		{
			name:     "04-ScoreBestNPUNodes return nil when node npu meet task req",
			attr:     attr2,
			task:     task2,
			nodes:    nodes,
			scoreMap: map[string]float64{"node1": 0, "node2": 0},
			wantSMap: map[string]float64{"node1": 0, "node2": score64},
			wantErr:  nil,
		},
	}

}

// TestCheckNodeNPUByTask
func TestScoreBestNPUNodes(t *testing.T) {
	const npuNum2 = 2
	npu := New(SchedulerName)
	env := plugin.ScheduleEnv{
		Nodes: map[string]plugin.NPUNode{
			"node1": {CommonNode: plugin.CommonNode{Annotation: map[string]string{util.NPU910CardName: "Ascend910-0"},
				Allocate: map[v1.ResourceName]float64{util.NPU910CardName: npuNum2 * util.NPUHexKilo}}},
			"node2": {CommonNode: plugin.CommonNode{Annotation: map[string]string{util.NPU910CardName: "Ascend910-0," +
				"Ascend910-1"},
				Allocate: map[v1.ResourceName]float64{util.NPU910CardName: npuNum2 * util.NPUHexKilo}}},
			"node3": {CommonNode: plugin.CommonNode{Annotation: map[string]string{util.NPU910CardName: ""},
				Allocate: map[v1.ResourceName]float64{util.NPU910CardName: npuNum2 * util.NPUHexKilo}}},
		},
	}
	//npu.SetSchedulerEnv(env)
	testCases := buildScoreBestNPUNodesTestCases01()
	testCases = append(testCases, buildScoreBestNPUNodesTestCases02()...)
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			npu.SetSchedulerAttr(tt.attr)
			env.Jobs = map[api.JobID]plugin.SchedulerJob{test.FakeJobName: {SchedulerJobAttr: tt.attr}}
			npu.SetSchedulerEnv(env)
			err := npu.ScoreBestNPUNodes(tt.task, tt.nodes, tt.scoreMap)
			if !reflect.DeepEqual(err, tt.wantErr) || !reflect.DeepEqual(tt.scoreMap, tt.wantSMap) {
				t.Errorf("ScoreBestNPUNodes() scoreMap: %v, wantSMap: %v, error = %v, wantErr %v",
					tt.scoreMap, tt.wantSMap, err, tt.wantErr)
			}
		})
	}
}

type useAnnotationTestCase struct {
	name     string
	attr     util.SchedulerJobAttr
	task     *api.TaskInfo
	node     plugin.NPUNode
	wantNode *plugin.NPUNode
	podAnno  string
}

func buildUseAnnotationTestCase1(attr1 util.SchedulerJobAttr) useAnnotationTestCase {
	test1 := useAnnotationTestCase{
		name: "01-UseAnnotation task will select the npu which is the only one on the card",
		attr: attr1,
		task: test.FakeTaskWithResReq("pod0", util.NPU910CardName, 1),
		node: plugin.NPUNode{
			CommonNode: plugin.CommonNode{
				Annotation: map[string]string{util.NPU910CardName: "Ascend910-0"},
				Allocate:   map[v1.ResourceName]float64{util.NPU910CardName: 1 * util.NPUHexKilo},
				Idle:       map[v1.ResourceName]float64{util.NPU910CardName: 1 * util.NPUHexKilo},
			},
		},
		podAnno: "Ascend910-0",
		wantNode: &plugin.NPUNode{
			CommonNode: plugin.CommonNode{
				Allocate:   map[v1.ResourceName]float64{util.NPU910CardName: 0},
				Idle:       map[v1.ResourceName]float64{util.NPU910CardName: 0},
				Annotation: map[string]string{util.NPU910CardName: ""},
			},
		},
	}
	return test1
}

func buildUseAnnotationTestCases() []useAnnotationTestCase {
	job1 := test.FakeNormalTestJob("job", 1)
	test.SetFakeJobResRequest(job1, util.NPU910CardName, "1")
	attr1 := itest.FakeSchedulerJobAttrByJob(job1)

	job2 := test.FakeNormalTestJob("job", 1)
	test.SetFakeJobResRequest(job2, util.NPU910CardName, "2")
	attr2 := itest.FakeSchedulerJobAttrByJob(job2)

	return []useAnnotationTestCase{
		buildUseAnnotationTestCase1(attr1),
		{
			name: "02-UseAnnotation task will select the one between tuo npu",
			attr: attr1,
			task: test.FakeTaskWithResReq("pod0", util.NPU910CardName, 1),
			node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Annotation: map[string]string{util.NPU910CardName: "Ascend910-0,Ascend910-1"},
					Allocate:   map[v1.ResourceName]float64{util.NPU910CardName: util.NPUIndex2 * util.NPUHexKilo},
					Idle:       map[v1.ResourceName]float64{util.NPU910CardName: util.NPUIndex2 * util.NPUHexKilo},
				},
			},
			podAnno: "Ascend910-0",
			wantNode: &plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Allocate:   map[v1.ResourceName]float64{util.NPU910CardName: 1 * util.NPUHexKilo},
					Idle:       map[v1.ResourceName]float64{util.NPU910CardName: 1 * util.NPUHexKilo},
					Annotation: map[string]string{util.NPU910CardName: "Ascend310-1"},
				},
			},
		},
		{
			name: "03-UseAnnotation task will select all the npu when task req 2 npu",
			attr: attr2,
			task: test.FakeTaskWithResReq("pod0", util.NPU910CardName, util.NPUIndex2),
			node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Annotation: map[string]string{util.NPU910CardName: "Ascend910-0,Ascend910-1"},
					Allocate:   map[v1.ResourceName]float64{util.NPU910CardName: util.NPUIndex2 * util.NPUHexKilo},
					Idle:       map[v1.ResourceName]float64{util.NPU910CardName: util.NPUIndex2 * util.NPUHexKilo},
				},
			},
			podAnno: "Ascend910-0,Ascend910-1",
			wantNode: &plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Allocate:   map[v1.ResourceName]float64{util.NPU910CardName: 0},
					Idle:       map[v1.ResourceName]float64{util.NPU910CardName: 0},
					Annotation: map[string]string{util.NPU910CardName: ""},
				},
			},
		},
	}
}

func TestUseAnnotation(t *testing.T) {
	npu := New(SchedulerName)
	testCases := buildUseAnnotationTestCases()
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			npu.SetSchedulerAttr(tt.attr)
			env := plugin.ScheduleEnv{
				Jobs: map[api.JobID]plugin.SchedulerJob{test.FakeJobName: {SchedulerJobAttr: tt.attr}},
			}
			npu.SetSchedulerEnv(env)
			node := npu.UseAnnotation(tt.task, tt.node)
			if !reflect.DeepEqual(node.Annotation, tt.node.Annotation) ||
				!reflect.DeepEqual(node.Idle, tt.node.Idle) ||
				!reflect.DeepEqual(node.Allocate, tt.node.Allocate) ||
				!reflect.DeepEqual(tt.task.Pod.Annotations[util.NPU910CardName], tt.podAnno) {
				t.Errorf("UseAnnotation() node: %v, wantNode: %v, anno %v, wantAnno %v",
					node, tt.wantNode, tt.task.Pod.Annotations, tt.podAnno)
			}
		})
	}
}

type card910x2Fields struct {
	reHandle     *rescheduling.ReScheduler
	NPUHandler   base.NPUHandler
	affScoreList [][]int
}

type card910x2PreStartActionArgs struct {
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

type card910x2PreStartActionTests struct {
	name    string
	fields  card910x2Fields
	args    card910x2PreStartActionArgs
	wantErr bool
}

func buildCard910x2PreStartActionTest1() card910x2PreStartActionTests {
	test1 := card910x2PreStartActionTests{
		name: "01-Card910x2PreStartAction()-nil tp",
		fields: card910x2Fields{
			NPUHandler:   base.NPUHandler{},
			affScoreList: nil,
			reHandle:     nil,
		},
		args: card910x2PreStartActionArgs{
			ssn:      nil,
			addCache: false,
		},
		wantErr: true,
	}
	return test1
}

func buildCard910x2PreStartActionTest2() card910x2PreStartActionTests {
	test2 := card910x2PreStartActionTests{
		name: "02-Card910x2PreStartAction()-off rescheduling",
		fields: card910x2Fields{
			NPUHandler:   base.NPUHandler{},
			affScoreList: nil,
			reHandle:     nil,
		},
		args: card910x2PreStartActionArgs{
			ssn:      test.FakeSSNReSchedule(),
			addCache: false,
		},
		wantErr: true,
	}
	test2.fields.NPUHandler.SchedulerJobAttr.Label = map[string]string{rescheduling.
		JobRescheduleLabelKey: rescheduling.JobOffRescheduleLabelValue}
	return test2
}

func buildCard910x2PreStartActionTest3() card910x2PreStartActionTests {
	var tmpPatch1 *gomonkey.Patches
	var tmpPatch2 *gomonkey.Patches
	var tmpPatch3 *gomonkey.Patches
	var tmpPatch4 *gomonkey.Patches
	var tmpPatch5 *gomonkey.Patches
	var tmpPatch6 *gomonkey.Patches
	var tmpPatch7 *gomonkey.Patches
	var tmpPatch8 *gomonkey.Patches
	var tmpPatch9 *gomonkey.Patches
	test3 := card910x2PreStartActionTests{
		name: "03-Card910x2PreStartAction()-success",
		fields: card910x2Fields{
			NPUHandler:   base.NPUHandler{},
			affScoreList: nil,
			reHandle:     nil,
		},
		args: card910x2PreStartActionArgs{
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
			cacheFuncAfter1:  func() { test.PatchReset(tmpPatch1) },
			cacheFuncAfter2:  func() { test.PatchReset(tmpPatch2) },
			cacheFuncAfter3:  func() { test.PatchReset(tmpPatch3) },
			cacheFuncAfter4:  func() { test.PatchReset(tmpPatch4) },
			cacheFuncAfter5:  func() { test.PatchReset(tmpPatch5) },
			cacheFuncAfter6:  func() { test.PatchReset(tmpPatch6) },
			cacheFuncAfter7:  func() { test.PatchReset(tmpPatch7) },
			cacheFuncAfter8:  func() { test.PatchReset(tmpPatch8) },
			cacheFuncAfter9:  func() { test.PatchReset(tmpPatch9) },
		},
		wantErr: true,
	}
	test3.fields.NPUHandler.SchedulerJobAttr.Label = map[string]string{rescheduling.
		JobRescheduleLabelKey: rescheduling.JobGraceRescheduleLabelValue}
	return test3
}

func buildCard910x2PreStartActionTests() []card910x2PreStartActionTests {

	tests := []card910x2PreStartActionTests{
		buildCard910x2PreStartActionTest1(),
		buildCard910x2PreStartActionTest2(),
		buildCard910x2PreStartActionTest3(),
	}
	return tests
}

// TestCard910x2PreStartAction test for preStartAction
func TestCard910x2PreStartAction(t *testing.T) {
	tests := buildCard910x2PreStartActionTests()
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
			tp := &card910x2{
				NPUHandler:   tt.fields.NPUHandler,
				affScoreList: tt.fields.affScoreList,
				reHandle:     tt.fields.reHandle,
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
