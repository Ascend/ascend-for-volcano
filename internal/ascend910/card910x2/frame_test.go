/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package card910x2 is using for HuaWei Ascend pin affinity schedule.

*/
package card910x2

import (
	"errors"
	"reflect"
	"testing"

	"k8s.io/api/core/v1"
	"volcano.sh/volcano/pkg/scheduler/api"

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
				Reason:  "job req npu is invalid",
				Message: "job<vcjob/job01> req npu num<0> is invalid",
			},
		},
		{
			name: "02-ValidNPUJob should return error when task request npu more than 3",
			attr: attr2,
			wantErr: &api.ValidateResult{
				Pass:   false,
				Reason: "job req npu is invalid",
				Message: "huawei.com/Ascend910card checkCardDistributeTrainMode distributed card train job<pod0> " +
					"req npu<3> not equal<2>",
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
				Reason:  "job req npu is invalid",
				Message: "job<vcjob/job04> req npu num<0> is invalid",
			},
		},
		{
			name: "05-ValidNPUJob should return error when task request npu more than 4",
			attr: attr5,
			wantErr: &api.ValidateResult{
				Pass:   false,
				Reason: "job req npu is invalid",
				Message: "huawei.com/Ascend910card checkSingleTrainMode single trainning job<vcjob/job05> " +
					"has too many task: <2>",
			},
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

func buildCheckNodeNPUByTaskTestCases() []itest.CheckNodeNPUByTaskTestCase {
	return []itest.CheckNodeNPUByTaskTestCase{
		{
			Name: "01-CheckNodeNPUByTask when return nil node npu meet task req",
			Task: test.FakeTaskWithResReq("pod0", util.NPU910CardName, util.NPUIndex2),
			Node: plugin.NPUNode{
				Name:       "node1",
				Annotation: map[string]string{util.NPU910CardName: "Ascend910-0,Ascend910-1"},
			},
			WantErr: nil,
		},
		{
			Name: "02-CheckNodeNPUByTask return err when task is not npu task",
			Task: test.FakeTaskWithResReq("pod1", util.NPU910CardName, util.NPUIndex2),
			Node: plugin.NPUNode{
				Name:       "node1",
				Annotation: map[string]string{util.NPU910CardName: "Ascend910-0,Ascend910-1"},
			},
			WantErr: errors.New("task<pod1> is not npu task"),
		},
		{
			Name: "03-CheckNodeNPUByTask return err when node has no req npu",
			Task: test.FakeTaskWithResReq("pod0", util.NPU910CardName, util.NPUIndex2),
			Node: plugin.NPUNode{
				Name:       "node1",
				Annotation: map[string]string{util.NPU310PCardName: "Ascend910-0,Ascend910-1"},
			},
			WantErr: errors.New("getUsableTopFromNode node<node1> don't have npu<huawei.com/Ascend910>"),
		},
		{
			Name: "04-CheckNodeNPUByTask return err when node has no req npu",
			Task: test.FakeTaskWithResReq("pod0", util.NPU910CardName, util.NPUIndex2),
			Node: plugin.NPUNode{
				Name:       "node1",
				Annotation: map[string]string{util.NPU910CardName: "Ascend910-0, Ascend910-1"},
			},
			WantErr: errors.New("getUsableTopFromNode err: top string<Ascend910-0, Ascend910-1> convert faild"),
		},
		{
			Name: "05-CheckNodeNPUByTask return err when node has no req npu",
			Task: test.FakeTaskWithResReq("pod0", util.NPU910CardName, util.NPUIndex2),
			Node: plugin.NPUNode{
				Name:       "node1",
				Annotation: map[string]string{util.NPU910CardName: "Ascend910-0"},
			},
			WantErr: errors.New("node <node1> don't have enough resource <huawei.com/Ascend910>, req<2>, idle<1>"),
		},
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
			"node1": {Annotation: map[string]string{util.NPU910CardName: "Ascend910-0"},
				Allocate: map[v1.ResourceName]float64{util.NPU910CardName: npuNum2 * util.NPUHexKilo}},
			"node2": {Annotation: map[string]string{util.NPU910CardName: "Ascend910-0,Ascend910-1"},
				Allocate: map[v1.ResourceName]float64{util.NPU910CardName: npuNum2 * util.NPUHexKilo}},
			"node3": {Annotation: map[string]string{util.NPU910CardName: ""},
				Allocate: map[v1.ResourceName]float64{util.NPU910CardName: npuNum2 * util.NPUHexKilo}},
		},
	}
	npu.SetSchedulerEnv(env)
	testCases := buildScoreBestNPUNodesTestCases01()
	testCases = append(testCases, buildScoreBestNPUNodesTestCases02()...)
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			npu.SetSchedulerAttr(tt.attr)
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
			Annotation: map[string]string{util.NPU910CardName: "Ascend910-0"},
			Allocate:   map[v1.ResourceName]float64{util.NPU910CardName: 1 * util.NPUHexKilo},
			Idle:       map[v1.ResourceName]float64{util.NPU910CardName: 1 * util.NPUHexKilo},
		},
		podAnno: "Ascend910-0",
		wantNode: &plugin.NPUNode{
			Allocate:   map[v1.ResourceName]float64{util.NPU910CardName: 0},
			Idle:       map[v1.ResourceName]float64{util.NPU910CardName: 0},
			Annotation: map[string]string{util.NPU910CardName: ""},
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
				Annotation: map[string]string{util.NPU910CardName: "Ascend910-0,Ascend910-1"},
				Allocate:   map[v1.ResourceName]float64{util.NPU910CardName: util.NPUIndex2 * util.NPUHexKilo},
				Idle:       map[v1.ResourceName]float64{util.NPU910CardName: util.NPUIndex2 * util.NPUHexKilo},
			},
			podAnno: "Ascend910-0",
			wantNode: &plugin.NPUNode{
				Allocate:   map[v1.ResourceName]float64{util.NPU910CardName: 1 * util.NPUHexKilo},
				Idle:       map[v1.ResourceName]float64{util.NPU910CardName: 1 * util.NPUHexKilo},
				Annotation: map[string]string{util.NPU910CardName: "Ascend310-1"},
			},
		},
		{
			name: "03-UseAnnotation task will select all the npu when task req 2 npu",
			attr: attr2,
			task: test.FakeTaskWithResReq("pod0", util.NPU910CardName, util.NPUIndex2),
			node: plugin.NPUNode{
				Annotation: map[string]string{util.NPU910CardName: "Ascend910-0,Ascend910-1"},
				Allocate:   map[v1.ResourceName]float64{util.NPU910CardName: util.NPUIndex2 * util.NPUHexKilo},
				Idle:       map[v1.ResourceName]float64{util.NPU910CardName: util.NPUIndex2 * util.NPUHexKilo},
			},
			podAnno: "Ascend910-0,Ascend910-1",
			wantNode: &plugin.NPUNode{
				Allocate:   map[v1.ResourceName]float64{util.NPU910CardName: 0},
				Idle:       map[v1.ResourceName]float64{util.NPU910CardName: 0},
				Annotation: map[string]string{util.NPU910CardName: ""},
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
