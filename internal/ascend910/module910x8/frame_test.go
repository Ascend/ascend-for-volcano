/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package module910x8 is using for HuaWei Ascend pin affinity schedule.

*/
package module910x8

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

// TestNew
func TestNew(t *testing.T) {
	t.Run("test New", func(t *testing.T) {
		npu := New(SchedulerName)
		if npu.GetPluginName() != SchedulerName {
			t.Errorf("New() npu Name: %s, wantName: %s.", npu.GetPluginName(), SchedulerName)
		}
		if npu.GetAnnoName() != util.NPU910CardName {
			t.Errorf("New() npu annoName: %s, wantAnnoName: %s.", npu.GetPluginName(), util.NPU910CardName)
		}
		if npu.GetAnnoPreVal() != util.NPU910CardNamePre {
			t.Errorf("New() npu annoNamePre: %s, wantAnnoNamePre: %s.",
				npu.GetPluginName(), util.NPU910CardNamePre)
		}
	})
}

func buildValidNPUJobTestCase01() []itest.ValidNPUJobTestCase {
	job01 := test.FakeNormalTestJob("job01", 1)
	test.SetFakeJobResRequest(job01, util.NPU910CardName, "0")
	attr1 := itest.FakeSchedulerJobAttrByJob(job01)
	job02 := test.FakeNormalTestJob("job02", 1)
	test.SetFakeJobResRequest(job02, util.NPU910CardName, "3")
	attr2 := itest.FakeSchedulerJobAttrByJob(job02)
	job03 := test.FakeNormalTestJob("job03", 1)
	test.SetFakeJobResRequest(job03, util.NPU910CardName, "2")
	attr3 := itest.FakeSchedulerJobAttrByJob(job03)
	return []itest.ValidNPUJobTestCase{
		{
			Name: "01-ValidNPUJob should return error when job request no npu",
			Attr: attr1,
			WantErr: &api.ValidateResult{
				Pass:   false,
				Reason: "job req npu is invalid",
				Message: "huawei.com/Ascend910module checkSingleTrainMode job<vcjob/job01> req npu num is not " +
					"[1 or 2 or 4 or 8]",
			},
		},
		{
			Name: "02-ValidNPUJob should return error when tasks request is not 1-2-4-8",
			Attr: attr2,
			WantErr: &api.ValidateResult{
				Pass:   false,
				Reason: "job req npu is invalid",
				Message: "huawei.com/Ascend910module checkSingleTrainMode job<vcjob/job02> req npu num is not " +
					"[1 or 2 or 4 or 8]",
			},
		},
		{
			Name:    "03-ValidNPUJob should return nil when tasks request is valid",
			Attr:    attr3,
			WantErr: nil,
		},
	}
}

func buildValidNPUJobTestCase02() []itest.ValidNPUJobTestCase {
	job04 := test.FakeNormalTestJob("job04", util.NPUIndex2)
	test.SetFakeJobResRequest(job04, util.NPU910CardName, "1")
	attr4 := itest.FakeSchedulerJobAttrByJob(job04)
	job05 := test.FakeNormalTestJob("job05", util.NPUIndex2)
	test.SetFakeJobResRequest(job05, util.NPU910CardName, "8")
	attr5 := itest.FakeSchedulerJobAttrByJob(job05)
	task2 := util.NPUTask{ReqNPUNum: 1}
	attr5.Tasks["vcjob-pod1"] = task2
	job06 := test.FakeNormalTestJob("job06", util.NPUIndex2)
	test.SetFakeJobResRequest(job06, util.NPU910CardName, "8")
	attr6 := itest.FakeSchedulerJobAttrByJob(job06)
	return []itest.ValidNPUJobTestCase{
		{
			Name: "04-ValidNPUJob should return error when task request no npu",
			Attr: attr4,
			WantErr: &api.ValidateResult{
				Pass:   false,
				Reason: "job req npu is invalid",
				Message: "huawei.com/Ascend910module checkSingleTrainMode job<vcjob/job04> single trainning " +
					"has too many task:2",
			},
		},
		{
			Name: "05-ValidNPUJob should return error when task request npu more than 4",
			Attr: attr5,
			WantErr: &api.ValidateResult{
				Pass:    false,
				Reason:  "job req npu is invalid",
				Message: "checkModuleDistributeTrainMode distributeTrain task<> req npu[1] not equal [8]",
			},
		},
		{
			Name:    "06-ValidNPUJob should return nil when tasks request is valid",
			Attr:    attr6,
			WantErr: nil,
		},
	}
}

// TestValidNPUJob
func TestValidNPUJob(t *testing.T) {
	npu := New(SchedulerName)
	testCases := buildValidNPUJobTestCase01()
	testCases = append(testCases, buildValidNPUJobTestCase02()...)
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			npu.SetSchedulerAttr(tt.Attr)
			if err := npu.ValidNPUJob(); !reflect.DeepEqual(err, tt.WantErr) {
				t.Errorf("ValidNPUJob() error = %v, wantErr %v", err, tt.WantErr)
			}
		})
	}
}

func buildCheckNodeNPUByTaskTestCases01() []itest.CheckNodeNPUByTaskTestCase {
	return []itest.CheckNodeNPUByTaskTestCase{
		{
			Name: "01-CheckNodeNPUByTask return nil when node npu meet task req",
			Task: test.FakeTaskWithResReq("pod0", util.NPU910CardName, util.NPUIndex4),
			Node: plugin.NPUNode{
				Name: "node1",
				Annotation: map[string]string{
					util.NPU910CardName: "Ascend910-1,Ascend910-2,Ascend910-3",
					networkUnhealthyNPU: "Ascend910-0"},
			},
			WantErr: nil,
		},
		{
			Name: "02-CheckNodeNPUByTask return err when task is not npu task",
			Task: test.FakeTaskWithResReq("pod1", util.NPU910CardName, util.NPUIndex4),
			Node: plugin.NPUNode{
				Name:       "node1",
				Annotation: map[string]string{util.NPU910CardName: "Ascend910-0,Ascend910-1,Ascend910-2,Ascend910-3"},
			},
			WantErr: errors.New("task<pod1> is not npu task"),
		},
		{
			Name: "03-CheckNodeNPUByTask return err when node has no req npu",
			Task: test.FakeTaskWithResReq("pod0", util.NPU310PCardName, util.NPUIndex4),
			Node: plugin.NPUNode{
				Name:       "node1",
				Annotation: map[string]string{util.NPU310CardName: "Ascend310-0,Ascend310-1,Ascend310-2"},
			},
			WantErr: errors.New("getUsableTopFromNode node<node1> don't have npu<huawei.com/Ascend910>"),
		},
	}

}

func buildCheckNodeNPUByTaskTestCases02() []itest.CheckNodeNPUByTaskTestCase {
	return []itest.CheckNodeNPUByTaskTestCase{
		{
			Name: "04-CheckNodeNPUByTask return err when node has no req npu",
			Task: test.FakeTaskWithResReq("pod0", util.NPU910CardName, util.NPUIndex4),
			Node: plugin.NPUNode{
				Name:       "node1",
				Annotation: map[string]string{util.NPU910CardName: "Ascend910-0, Ascend910-1"},
			},
			WantErr: errors.New("getUsableTopFromNode node<node1> don't have npu<huawei.com/Ascend910>"),
		},
		{
			Name: "05-CheckNodeNPUByTask return err when node has no req npu",
			Task: test.FakeTaskWithResReq("pod0", util.NPU910CardName, util.NPUIndex4),
			Node: plugin.NPUNode{
				Name: "node1",
				Annotation: map[string]string{
					util.NPU910CardName: "Ascend910-0,Ascend910-1,Ascend910-4",
				},
			},
			WantErr: errors.New("CheckNodeNPUByTask the npus on this node don't satisfy the schedulable " +
				"topology err: [0 1 4] not meet req npu(4)"),
		},
		{
			Name: "06-CheckNodeNPUByTask return err when node has no req npu",
			Task: test.FakeTaskWithResReq("pod0", util.NPU910CardName, util.NPUIndex4),
			Node: plugin.NPUNode{
				Name: "node1",
				Annotation: map[string]string{
					util.NPU910CardName: "Ascend910-0,Ascend910-1,Ascend910-4",
					networkUnhealthyNPU: "Ascend910-5",
				},
			},
			WantErr: errors.New("CheckNodeNPUByTask the npus on this node don't satisfy the schedulable " +
				"topology err: [0 1 4 5] not meet req npu(4)"),
		},
	}
}

func buildCheckNodeNPUByTaskTestCases03() []itest.CheckNodeNPUByTaskTestCase {
	return []itest.CheckNodeNPUByTaskTestCase{
		{
			Name: "01-CheckNodeNPUByTask return err when task is nil",
			Task: nil,
			Node: plugin.NPUNode{
				Name:       "node1",
				Annotation: map[string]string{util.NPU310PCardName: "Ascend310P-0,Ascend310P-1"},
			},
			WantErr: errors.New(util.ArgumentError),
		},
		{
			Name: "02-CheckNodeNPUByTask return err when node annotation is nil",
			Task: test.FakeTaskWithResReq("pod1", util.NPU310PCardName, util.NPUIndex2),
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
	npu := New(SchedulerName)
	job := test.FakeNormalTestJob("job", 1)
	test.SetFakeJobResRequest(job, util.NPU910CardName, "4")
	attr := itest.FakeSchedulerJobAttrByJob(job)
	sJob := plugin.SchedulerJob{}
	sJob.SchedulerJobAttr = attr
	env := plugin.ScheduleEnv{
		Jobs: map[api.JobID]plugin.SchedulerJob{job.UID: sJob},
	}
	npu.SetSchedulerAttr(attr)
	npu.SetSchedulerEnv(env)
	testCases := buildCheckNodeNPUByTaskTestCases01()
	testCases = append(testCases, buildCheckNodeNPUByTaskTestCases02()...)
	testCases = append(testCases, buildCheckNodeNPUByTaskTestCases03()...)
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			if err := npu.CheckNodeNPUByTask(tt.Task, tt.Node); !reflect.DeepEqual(err, tt.WantErr) {
				t.Errorf("ValidNPUJob() error = %v, wantErr %v", err, tt.WantErr)
			}
		})
	}
}

func buildScoreBestNPUNodesTestCases01() []itest.ScoreBestNPUNodesTestCase {
	return []itest.ScoreBestNPUNodesTestCase{
		{
			Name:     "01-ScoreBestNPUNodes return err when task is not this job npu task ",
			Task:     test.FakeTaskWithResReq("pod1", util.NPU910CardName, 1),
			Nodes:    []*api.NodeInfo{{Name: "node1"}, {Name: "node2"}},
			ScoreMap: map[string]float64{"node1": 0, "node2": 0},
			WantSMap: map[string]float64{"node1": 0, "node2": 0},
			WantErr:  errors.New("task<pod1> is not npu task"),
		},
		{
			Name:     "02-ScoreBestNPUNodes scoreMap no refresh when node is not this job npu node",
			Task:     test.FakeTaskWithResReq("pod0", util.NPU910CardName, 1),
			Nodes:    []*api.NodeInfo{{Name: "node13"}},
			ScoreMap: map[string]float64{"node13": 0},
			WantSMap: map[string]float64{"node13": 0},
			WantErr:  nil,
		},
		{
			Name:     "03-ScoreBestNPUNodes scoreMap no refresh when node netUnhealthyNPU not define",
			Task:     test.FakeTaskWithResReq("pod0", util.NPU910CardName, 1),
			Nodes:    []*api.NodeInfo{{Name: "node14"}},
			ScoreMap: map[string]float64{"node14": 0},
			WantSMap: map[string]float64{"node14": 0},
			WantErr:  nil,
		},
	}
}

func buildScoreBestNPUNodesTestCases02() []itest.ScoreBestNPUNodesTestCase {
	const (
		score232 = 232
		score240 = 240
		score248 = 248
		score256 = 256
	)
	return []itest.ScoreBestNPUNodesTestCase{
		{
			Name:     "04-ScoreBestNPUNodes scoreMap no refresh when node has no npu",
			Task:     test.FakeTaskWithResReq("pod0", util.NPU910CardName, 1),
			Nodes:    []*api.NodeInfo{{Name: "node15"}},
			ScoreMap: map[string]float64{"node15": 0},
			WantSMap: map[string]float64{"node15": 0},
			WantErr:  nil,
		},
		{
			Name:     "05-ScoreBestNPUNodes return nil when node npu not meet task req",
			Task:     test.FakeTaskWithResReq("pod0", util.NPU910CardName, 1),
			Nodes:    []*api.NodeInfo{{Name: "node1"}, {Name: "node3"}, {Name: "node11"}, {Name: "node12"}},
			ScoreMap: map[string]float64{"node1": 0, "node3": 0, "node11": 0, "node12": 0},
			WantSMap: map[string]float64{"node1": score256, "node3": score248, "node11": score240, "node12": score232},
			WantErr:  nil,
		},
		{
			Name:     "06-ScoreBestNPUNodes return nil when node npu not meet task req",
			Task:     test.FakeTaskWithResReq("pod0", util.NPU910CardName, 1),
			Nodes:    []*api.NodeInfo{{Name: "node12"}, {Name: "node11"}},
			ScoreMap: map[string]float64{"node12": 0, "node11": 0},
			WantSMap: map[string]float64{"node12": score232, "node11": score240},
			WantErr:  nil,
		},
	}
}

func buildFakeScheduleEnv() plugin.ScheduleEnv {
	const allocateNPUNum8 = 8
	return plugin.ScheduleEnv{
		Nodes: map[string]plugin.NPUNode{
			"node1": {Annotation: map[string]string{util.NPU910CardName: "Ascend910-0", networkUnhealthyNPU: ""},
				Allocate: map[v1.ResourceName]float64{util.NPU910CardName: allocateNPUNum8 * util.NPUHexKilo}},
			"node2": {Annotation: map[string]string{util.NPU910CardName: "Ascend910-0,Ascend910-1"}},
			"node3": {Annotation: map[string]string{util.NPU910CardName: "Ascend910-0,Ascend910-1,Ascend910-2",
				networkUnhealthyNPU: ""},
				Allocate: map[v1.ResourceName]float64{util.NPU910CardName: allocateNPUNum8 * util.NPUHexKilo}},
			"node11": {Annotation: map[string]string{util.NPU910CardName: "Ascend910-4,Ascend910-5",
				networkUnhealthyNPU: ""},
				Allocate: map[v1.ResourceName]float64{util.NPU910CardName: allocateNPUNum8 * util.NPUHexKilo}},
			"node12": {Annotation: map[string]string{util.NPU910CardName: "Ascend910-4,Ascend910-5,Ascend910-6," +
				"Ascend910-7", networkUnhealthyNPU: ""},
				Allocate: map[v1.ResourceName]float64{util.NPU910CardName: allocateNPUNum8 * util.NPUHexKilo}},
			"node13": {Annotation: map[string]string{}},
			"node14": {Annotation: map[string]string{util.NPU910CardName: "Ascend910-0"}},
			"node15": {Annotation: map[string]string{util.NPU910CardName: "", networkUnhealthyNPU: ""}},
		},
	}
}

// TestCheckNodeNPUByTask
func TestScoreBestNPUNodes(t *testing.T) {
	npu := New(SchedulerName)
	env := buildFakeScheduleEnv()
	npu.SetSchedulerEnv(env)
	job := test.FakeNormalTestJob("job", 1)
	test.SetFakeJobResRequest(job, util.NPU910CardName, "1")
	attr := itest.FakeSchedulerJobAttrByJob(job)
	npu.SetSchedulerAttr(attr)
	testCases := buildScoreBestNPUNodesTestCases01()
	testCases = append(testCases, buildScoreBestNPUNodesTestCases02()...)
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			err := npu.ScoreBestNPUNodes(tt.Task, tt.Nodes, tt.ScoreMap)
			if !reflect.DeepEqual(err, tt.WantErr) || !reflect.DeepEqual(tt.ScoreMap, tt.WantSMap) {
				t.Errorf("ScoreBestNPUNodes() scoreMap: %v, wantSMap: %v, error = %v, wantErr %v",
					tt.ScoreMap, tt.WantSMap, err, tt.WantErr)
			}
		})
	}
}

func buildUseAnnotationTestCases01() []itest.UseAnnotationTestCase {
	return []itest.UseAnnotationTestCase{
		{
			Name: "01-UseAnnotation task will select the npu which is the only one on the card",
			Task: test.FakeTaskWithResReq("pod0", util.NPU910CardName, 1),
			Node: plugin.NPUNode{
				Annotation: map[string]string{util.NPU910CardName: "Ascend910-0,Ascend910-4,Ascend910-5",
					networkUnhealthyNPU: ""},
			},
			PodAnno: "Ascend910-0",
			WantNode: &plugin.NPUNode{
				Annotation: map[string]string{util.NPU910CardName: ""},
			},
		},
		{
			Name: "02-UseAnnotation task will select the npu which is on the card that has 3 npu other than 2",
			Task: test.FakeTaskWithResReq("pod0", util.NPU910CardName, 1),
			Node: plugin.NPUNode{
				Annotation: map[string]string{util.NPU910CardName: "Ascend910-0,Ascend910-1,Ascend910-2,Ascend910-4," +
					"Ascend910-5", networkUnhealthyNPU: ""},
			},
			PodAnno: "Ascend910-0",
			WantNode: &plugin.NPUNode{
				Annotation: map[string]string{util.NPU910CardName: "Ascend910-1,Ascend910-2,Ascend910-4,Ascend910-5"},
			},
		},
	}
}

func buildUseAnnotationTestCases02() []itest.UseAnnotationTestCase {
	return []itest.UseAnnotationTestCase{
		{
			Name: "03-UseAnnotation task will select the npu which is on the card that has 3 npu other than 2",
			Task: test.FakeTaskWithResReq("pod0", util.NPU910CardName, 1),
			Node: plugin.NPUNode{
				Annotation: map[string]string{util.NPU910CardName: "Ascend910-0,Ascend910-1,Ascend910-4,Ascend910-5," +
					"Ascend910-6", networkUnhealthyNPU: ""},
			},
			PodAnno: "Ascend910-4",
			WantNode: &plugin.NPUNode{
				Annotation: map[string]string{util.NPU910CardName: "Ascend910-0,Ascend910-1,Ascend910-5,Ascend910-6"},
			},
		},
		{
			Name: "04-UseAnnotation task will select the npu which is on the card that has 2 npu other than 4",
			Task: test.FakeTaskWithResReq("pod0", util.NPU910CardName, 1),
			Node: plugin.NPUNode{
				Annotation: map[string]string{util.NPU910CardName: "Ascend910-0,Ascend910-1,Ascend910-4,Ascend910-5," +
					"Ascend910-6,Ascend910-7", networkUnhealthyNPU: ""}},
			PodAnno: "Ascend910-0",
			WantNode: &plugin.NPUNode{
				Annotation: map[string]string{util.NPU910CardName: "Ascend910-1,Ascend910-4,Ascend910-5,Ascend910-6," +
					"Ascend910-0,Ascend910-7"},
			},
		},
	}
}

func TestUseAnnotation(t *testing.T) {
	npu := New(SchedulerName)
	job := test.FakeNormalTestJob("job", 1)
	test.SetFakeJobResRequest(job, util.NPU910CardName, "1")
	attr := itest.FakeSchedulerJobAttrByJob(job)
	npu.SetSchedulerAttr(attr)
	testCases := buildUseAnnotationTestCases01()
	testCases = append(testCases, buildUseAnnotationTestCases02()...)
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			node := npu.UseAnnotation(tt.Task, tt.Node)
			if tt.Task != nil && tt.Node.Annotation != nil && (!reflect.DeepEqual(node.Annotation,
				tt.Node.Annotation)) || !reflect.DeepEqual(tt.Task.Pod.Annotations[util.NPU910CardName], tt.PodAnno) {
				t.Errorf("UseAnnotation() node: %v, wantNode: %v, anno %v, wantAnno %v",
					node, tt.WantNode, tt.Task.Pod.Annotations, tt.PodAnno)
			}
			if (tt.Task == nil || tt.Node.Annotation == nil) || !reflect.DeepEqual(node, tt.WantNode) {
				t.Errorf("UseAnnotation() node: %v, wantNode: %v", node, tt.WantNode)
			}
		})
	}
}
