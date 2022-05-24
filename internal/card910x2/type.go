/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package card910x2 is using for HuaWei A300T Ascend pin affinity schedule.

*/
package card910x2

import "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"

const (
	// PluginName the card910x2's plugin name.
	PluginName          = "A300T"
	a300TNPUCardName    = "huawei.com/Ascend910"
	podPredicateTime    = "predicate-time"
	a300tNPUCardPreName = "Ascend910-"
	maxNPUNum           = 2

	archSelector          = "host-arch"
	huaweiArchArm         = "huawei-arm"
	huaweiArchX86         = "huawei-x86"
	acceleratorType       = "accelerator-type"
	cardAcceleratorType   = "card"
	moduleAcceleratorType = "module"

	npuNumPerHccs = 4
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
	plugin.HwEntity
}

type npuPriNodeInf struct {
	// the priority for NPU HCCS top
	Name     string
	nodeName string
}

type initPriNodeGroupFn func(priNodeGroup map[string]*npuPriNodeInf, groupName string)
