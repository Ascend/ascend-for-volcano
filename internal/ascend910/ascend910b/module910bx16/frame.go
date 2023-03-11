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
Package module910bx16 is using for HuaWei A800/9000 Ascend910B A+X pin affinity schedule.
*/
package module910bx16

import (
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/base"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// New return npu plugin
func New(name string) base.AscendHandler {
	m := &module910bx16{}
	m.SetPluginName(name)
	m.SetAnnoName(util.NPU910CardName)
	m.SetAnnoPreVal(util.NPU910CardNamePre)
	m.SetDefaultJobSchedulerConfig(nil)
	m.SetMaxNodeNPUNum(nodeNPUNumber)
	m.SetAcceleratorValue(util.JobKind910BValue)
	m.SetArch(util.HuaweiArchX86)
	m.SetSingleAllowNumsMap(map[int]struct{}{1: {}, util.NPUIndex2: {}, util.NPUIndex4: {}, util.NPUIndex8: {},
		util.NPUIndex16: {}})
	m.AffScoreList = [][]int{
		{util.AffScore0, util.AffScore1, util.AffScore2, util.AffScore3, util.AffScore4, util.AffScore5,
			util.AffScore6, util.AffScore7},
		{util.AffScore8, util.AffScore0, util.AffScore1, util.AffScore2, util.AffScore3, util.AffScore4,
			util.AffScore5, util.AffScore6},
		{util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8,
			util.AffScore8, util.AffScore8},
		{util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore0, util.AffScore1, util.AffScore2,
			util.AffScore3, util.AffScore4},
		{util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8,
			util.AffScore8, util.AffScore8},
		{util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8,
			util.AffScore8, util.AffScore8},
		{util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8,
			util.AffScore8, util.AffScore8},
		{util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8,
			util.AffScore8, util.AffScore0},
	}
	return m
}

// ValidNPUJob check job req npu num and mode
func (tp *module910bx16) ValidNPUJob() *api.ValidateResult {
	return tp.Valid910bNPUJob()
}

// PreStartAction pre-processing actions for rescheduling
func (tp *module910bx16) PreStartAction(ssn *framework.Session) error {
	return tp.PreStartActionCheck(ssn)
}

// PreStopAction post-processing actions for re-scheduling
func (tp *module910bx16) PreStopAction(env *plugin.ScheduleEnv) error {
	return tp.PreStopActionCheck(env)
}

// CheckNodeNPUByTask check nod npu meet task req
func (tp *module910bx16) CheckNodeNPUByTask(task *api.TaskInfo, node plugin.NPUNode) error {
	return tp.Check910bNodeNPUByTask(task, node)
}

func (tp *module910bx16) ScoreBestNPUNodes(task *api.TaskInfo, nodes []*api.NodeInfo, sMap map[string]float64) error {
	return tp.ScoreAscendNPUNodes(task, nodes, sMap)
}

// UseAnnotation select npu for task from node
func (tp *module910bx16) UseAnnotation(task *api.TaskInfo, node plugin.NPUNode) *plugin.NPUNode {
	return tp.Use910bAnnotation(task, node)
}

// ReleaseAnnotation Release used resource.
func (tp *module910bx16) ReleaseAnnotation(_ *api.TaskInfo, node plugin.NPUNode) *plugin.NPUNode {
	return &node
}
