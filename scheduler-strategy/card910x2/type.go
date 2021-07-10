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

Package card910x2 is using for HuaWei A300T Ascend pin affinity schedule.

*/
package card910x2

const (
	// PluginName the card910x2's plugin name.
	PluginName          = "A300T"
	a300TNPUCardName    = "huawei.com/Ascend910"
	podPredicateTime    = "predicate-time"
	a300tNPUCardPreName = "Ascend910-"

	archSelector          = "host-arch"
	huaweiArchArm         = "huawei-arm"
	huaweiArchX86         = "huawei-x86"
	acceleratorType       = "accelerator-type"
	cardAcceleratorType   = "card"
	moduleAcceleratorType = "module"

	npuNumPerHccs = 4
	constIntNum2  = 2
	constIntNum3  = 3
	logErrorLev   = 1
	logInfoLev    = 3
	logDebugLev   = 4

	nodesNoMeetNPUReqError     = "insufficient npus on the schedulable nodes in cluster"
	nodeNotStableWarning       = "the npus on this node are unstable"
	nodeNotMeetTopologyWarning = "the npus on this node don't satisfy the schedulable topology"
	nodeNotEnoughNPUWarning    = "insufficient number of available npus on this node"

	nodeNoFitNPUWarning = "node no fit npu number"
	jobNoNPUCard        = "job no use npu"
	modeNotCard         = "no card mode npu"
	argumentError       = "invalid argument"
)

type card910x2 struct {
	name string
}

type npuPriNodeInf struct {
	// the priority for NPU HCCS top
	Name     string
	nodeName string
}

type initPriNodeGroupFn func(priNodeGroup map[string]*npuPriNodeInf, groupName string)
