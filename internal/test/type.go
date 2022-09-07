/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package test is using for HuaWei Ascend pin fault rescheduling.

*/
package test

import (
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// InitMyJobPluginTestCase test case
type InitMyJobPluginTestCase struct {
	Name    string
	Attr    util.SchedulerJobAttr
	Env     plugin.ScheduleEnv
	WantErr error
}

// ValidNPUJobTestCase validNPUJob test case
type ValidNPUJobTestCase struct {
	WantErr *api.ValidateResult
	Name    string
	Attr    util.SchedulerJobAttr
}

// CheckNodeNPUByTaskTestCase CheckNodeNPUByTask test case
type CheckNodeNPUByTaskTestCase struct {
	Task    *api.TaskInfo
	Name    string
	Attr    util.SchedulerJobAttr
	Node    plugin.NPUNode
	WantErr error
}

// ScoreBestNPUNodesTestCase scoreBestNPUNodes test case
type ScoreBestNPUNodesTestCase struct {
	Task     *api.TaskInfo
	Nodes    []*api.NodeInfo
	ScoreMap map[string]float64
	WantSMap map[string]float64
	Name     string
	WantErr  error
}

// UseAnnotationTestCase useAnnotation test case
type UseAnnotationTestCase struct {
	Task     *api.TaskInfo
	WantNode *plugin.NPUNode
	Name     string
	Node     plugin.NPUNode
	PodAnno  string
}

// JudgeNodeAndTaskNPUTestCase JudgeNodeAndTaskNPU test case
type JudgeNodeAndTaskNPUTestCase struct {
	NodeTop []int
	Name    string
	TaskNPU int
	WantErr error
}

// SetMaxNodeNPUNumTestCase  SetMaxNodeNPUNum test case
type SetMaxNodeNPUNumTestCase struct {
	Name    string
	Num     int
	WantNum int
}
