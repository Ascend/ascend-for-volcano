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

Package cardv910x2 is using for virtual HuaWei A300T schedule.

*/
package cardv910x2

import (
	"volcano.sh/volcano/pkg/scheduler/plugins/huaweinpu/scheduler-strategy/commonv910"
)

const (
	// PluginName the cardv910x2's plugin name.
	PluginName = "A300T-Vnpu"

	mapInitLen  = 3
	logDebugLev = 4

	archSelector          = "host-arch"
	huaweiArchArm         = "huawei-arm"
	huaweiArchX86         = "huawei-x86"
	acceleratorType       = "accelerator-type"
	cardAcceleratorType   = "card"
	moduleAcceleratorType = "module"
)

type cardv910x2 struct {
	name string
	commonv910.Vnpu
}
