/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package module910x8 is using for HuaWei A800/9000 Ascend910 pin affinity schedule.

*/
package module910x8

import "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"

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
	npuNumPerHccs         = 4
	constIntNum1          = 1
	constIntNum2          = 2
	constIntNum3          = 3
	constIntNum4          = 4
	constIntNum5          = 5
	constIntNum6          = 6
	constIntNum7          = 7
	nodeNPUNumber         = 8

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
	plugin.HwEntity
}
