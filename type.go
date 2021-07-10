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

Package huaweinpu is using for HuaWei Ascend pin affinity schedule.

*/
package main

import (
	"volcano.sh/volcano/pkg/scheduler/framework"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

const (
	// PluginName use in frame
	PluginName  = "huaweinpu-v2.0.2"
	logErrorLev = 1
	logInfoLev  = 3
	logDebugLev = 4
)

type huaweiNPUPlugin struct {
	*plugin.ScheduleHandler
	// Arguments given for the plugin
	pluginArguments framework.Arguments
}