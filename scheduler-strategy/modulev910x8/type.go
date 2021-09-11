/*
Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.

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

Package modulev910x8 is using for virtual HuaWei Ascend910 schedule.

*/
package modulev910x8

import "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/commonv910"

const (
	// PluginName the modulev910x8's plugin name.
	PluginName = "A800-9000-VNpu"

	maxNPUNum   = 8
	logDebugLev = 4
)

type modulev910x8 struct {
	name string
	commonv910.Vnpu
}
