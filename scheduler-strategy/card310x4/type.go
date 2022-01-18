/*
Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package card310x4 is using for HuaWei A300T Ascend pin affinity schedule.

*/
package card310x4

import "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/common"

const (
	// PluginName the card310x4's plugin name.
	PluginName         = "A310-card"
	a310NPUCardName    = "huawei.com/Ascend310"
	podPredicateTime   = "predicate-time"
	a310NPUCardPreName = "Ascend310-"
	a310FaultNPUName   = "huawei.com/Ascend310-Unhealthy"

	archSelector  = "host-arch"
	huaweiArchArm = "huawei-arm"
	huaweiArchX86 = "huawei-x86"

	acceleratorType     = "npu-310-strategy"
	cardAcceleratorType = "card"
	chipAcceleratorType = "chip"

	nodeNPUNumber  = 64
	constIntNum0   = 0
	constIntNum1   = 1
	constIntNum2   = 2
	constIntNum3   = 3
	cardNPUNumber  = 4
	constIntNum5   = 5
	constIntNum10  = 10
	constNPUWeight = 8.0

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
	com  common.Scheduler
	re   common.ReScheduler
}

type npuPriNodeInf struct {
	// the priority for NPU top
	Name     string
	nodeName string
}

type initPriNodeGroupFn func(priNodeGroup map[string]*npuPriNodeInf, groupName string)
