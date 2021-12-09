/*
Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package plugin is using for HuaWei Ascend pin affinity schedule.

*/
package plugin

const (
	// PluginName the HuaWei NPU 's plugin name.
	PluginName             = "huaweiNPU"
	a310NPUCardName        = "huawei.com/Ascend310"
	logErrorLev            = 1
	logWarningLev          = 2
	logInfoLev             = 3
	logDebugLev            = 4
	nodeNoFitSelectorError = "no matching label on this node"
	nodesNoMeetNPUReqError = "insufficient npus on the schedulable nodes in cluster"
	noneNPUPlugin          = "get nil NPUPlugin(not npu task)"
)
