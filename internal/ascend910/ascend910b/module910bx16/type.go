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
Package module910bx16 is using for HuaWei Ascend 910B(A+X) pin affinity schedule.
*/
package module910bx16

import (
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/base"
)

type module910bx16 struct {
	base.NPUHandler
	affScoreList [][]int
}

const (
	// SchedulerName name of scheduler
	SchedulerName = "huawei.com/Ascend910module-910B-16"
	// module910BAcceleratorValue partner value 16 card only for A+X A2
	module910BAcceleratorValue = "module-910B-16"
	nodeNPUNumber              = 16
	nodeWeight                 = 8.0
)
