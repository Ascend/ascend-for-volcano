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

Package plugin is using for HuaWei Ascend pin affinity schedule.

*/
package plugin

const (
	// PluginName the HuaWei NPU 's plugin name.
	PluginName             = "huaweiNPU"
	logErrorLev            = 1
	logWarningLev          = 2
	logInfoLev             = 3
	logDebugLev            = 4
	nodeNoFitSelectorError = "no matching label on this node"
	nodesNoMeetNPUReqError = "insufficient npus on the schedulable nodes in cluster"
	noneNPUPlugin          = "get nil NPUPlugin(not npu task)"
)
