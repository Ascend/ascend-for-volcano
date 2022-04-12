/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
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

// HwEntity for all volcano-npu plugin.
type HwEntity struct {
	// the new func add name
	PluginName string
	// in k8s annotation huawei.com/Ascend310,huawei.com/Ascend910
	AnnoName string
	// huawei.com/
	AnnoPreVal string
	// fault chip name. like huawei.com/Ascend910-Unhealthy.
	AnnoUnhealthyName string
	// config like arm x86
	DefaultJobSchedulerConfig map[string]string

	HwNPUSchedulerPlugin
}
