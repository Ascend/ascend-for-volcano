/*
Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package util is using for HuaWei Ascend9 pin affinity schedule utilities.

*/
package util

const (
	constIntNum2           = 2
	constIntNum3           = 3
	nodeNPUNumber          = 8
	logErrorLev            = 1
	logInfoLev             = 3
	logDebugLev            = 4
	npuNumPerHccs          = 4
	npuHex                 = 1000
	maxTaskNPUNum          = 10000
	archSelector           = "host-arch"
	huaweiArchArm          = "huawei-arm"
	huaweiArchX86          = "huawei-x86"
	accelerator            = "accelerator"
	accelerator910Value    = "huawei-Ascend910"
	accelerator310Value    = "huawei-Ascend310"
	acceleratorType        = "accelerator-type"
	cardAcceleratorType    = "card"
	moduleAcceleratorType  = "module"
	chipAcceleratorType    = "chip"
	nodeNoFitSelectorError = "no matching label on this node"
)
