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

Package card310x4 is using for HuaWei A300T Ascend pin affinity schedule.

*/
package card310x4

const (
	// PluginName the card310x4's plugin name.
	PluginName         = "A310-card"
	a310NPUCardName    = "huawei.com/Ascend310"
	podPredicateTime   = "predicate-time"
	a310NPUCardPreName = "Ascend310-"

	archSelector  = "host-arch"
	huaweiArchArm = "huawei-arm"
	huaweiArchX86 = "huawei-x86"

	acceleratorType     = "npu-310-strategy"
	cardAcceleratorType = "card"
	chipAcceleratorType = "chip"

	nodeNPUNumber = 64
	constIntNum0  = 0
	constIntNum1  = 1
	constIntNum2  = 2
	constIntNum3  = 3
	cardNPUNumber = 4
	constIntNum5  = 5

	logErrorLev = 1
	logInfoLev  = 3
	logDebugLev = 4

	nodesNoMeetNPUReqError     = "insufficient npus on the schedulable nodes in cluster"
	nodeNotStableWarning       = "the npus on this node are unstable"
	nodeNotMeetTopologyWarning = "the npus on this node don't satisfy the schedulable topology"
	nodeNotEnoughNPUWarning    = "insufficient number of available npus on this node"

	nodeNoFitNPUWarning = "node no fit npu number"
	jobNoNPUCard        = "job no use npu"
	modeNotCard         = "no card mode npu"
	argumentError       = "invalid argument"
)

type card310x4 struct {
	name string
}

type npuPriNodeInf struct {
	// the priority for NPU top
	Name     string
	nodeName string
}

type initPriNodeGroupFn func(priNodeGroup map[string]*npuPriNodeInf, groupName string)