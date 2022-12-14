/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.

# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
*/

/*

Package ascend910 is using for HuaWei Ascend pin affinity schedule.

*/
package ascend910

import (
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/base"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/rescheduling"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

type ascend910 struct {
	// base event handler
	base.NPUHandler
	// 910 support scheduler kinds.
	Kind map[string]base.AscendHandler
	// specific job use.
	handle   base.AscendHandler
	reHandle *rescheduling.ReScheduler
}

const (
	// PluginName name of plugin
	PluginName = util.NPU910CardName
	// Accelerator910Key 910 accelerator key
	Accelerator910Key = "accelerator-type"
	// Module910AcceleratorValue module value
	Module910AcceleratorValue = "module"
	// Card910AcceleratorValue card value
	Card910AcceleratorValue = "card"
)
