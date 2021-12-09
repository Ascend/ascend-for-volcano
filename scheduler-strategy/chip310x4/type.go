/*
Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package chip310x4 is using for HuaWei A300T Ascend pin affinity schedule.

*/
package chip310x4

const (
	// PluginName the chip310x4's plugin name.
	PluginName         = "A310-chip"
	a310NPUChipName    = "huawei.com/Ascend310"
	podPredicateTime   = "predicate-time"
	a310NPUCardPreName = "Ascend310-"

	archSelector  = "host-arch"
	huaweiArchArm = "huawei-arm"
	huaweiArchX86 = "huawei-x86"

	acceleratorType     = "npu-310-strategy"
	cardAcceleratorType = "card"
	chipAcceleratorType = "chip"

	nodeNPUNumber = 64
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

	jobNoNPUCard  = "job no use npu"
	modeNotChip   = "no chip mode npu"
	argumentError = "invalid argument"
)

type chip310x4 struct {
	name string
}
