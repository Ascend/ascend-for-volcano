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
Package base is using for HuaWei Ascend pin affinity schedule.
*/
package base

import (
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// AscendHandler ascend npu event handler
type AscendHandler interface {
	plugin.ISchedulerPlugin
	SetSchedulerAttr(util.SchedulerJobAttr)
	SetSchedulerEnv(plugin.ScheduleEnv)
	SetMaxNodeNPUNum(int)
	SetMaxCardNPUNum(int)
}

// NPUHandler base npu handler
type NPUHandler struct {
	plugin.SchedulerPlugin
	util.SchedulerJobAttr
	plugin.ScheduleEnv
	MaxNodeNPUNum int
	MaxCardNPUNum int
}
