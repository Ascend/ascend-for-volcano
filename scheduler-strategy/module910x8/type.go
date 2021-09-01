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

Package module910x8 is using for HuaWei A800/9000 Ascend910 pin affinity schedule.

*/
package module910x8

const (
	// PluginName the module910x8's plugin name.
	PluginName            = "A800-9000"
	npu800And9000CardName = "huawei.com/Ascend910"
	podPredicateTime      = "predicate-time"
	npu910CardPreName     = "Ascend910-"
	archSelector          = "host-arch"
	huaweiArchArm         = "huawei-arm"
	huaweiArchX86         = "huawei-x86"
	moduleAcceleratorType = "module"
	maxNPUNum             = 8

	npuNumPerHccs = 4
	constIntNum2  = 2
	constIntNum3  = 3
	nodeNPUNumber = 8

	logErrorLev = 1
	logInfoLev  = 3
	logDebugLev = 4

	nodeNoFitNPUWarning = "node no fit npu number"
	jobNoNPUCard        = "job no use npu"
	argumentError       = "invalid argument"

	nodesNoMeetNPUReqError     = "insufficient npus on the schedulable nodes in cluster"
	nodeNotStableWarning       = "the npus on this node are unstable"
	nodeNotMeetTopologyWarning = "the npus on this node don't satisfy the schedulable topology"
	nodeNotEnoughNPUWarning    = "insufficient number of available npus on this node"
	jobRestartReason           = "restart for NPU malfunction"
)

type npuPriNodeInf struct {
	// the priority for NPU HCCS top
	Name     string
	nodeName string
}

type initPriNodeGroupFn func(priNodeGroup map[string]*npuPriNodeInf, groupName string)

type module910x8 struct {
	name string
}
