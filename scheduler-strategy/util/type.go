/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package util is using for HuaWei Ascend9 pin affinity schedule utilities.

*/
package util

const (
	LogErrorLev   = 1
	LogInfoLev    = 3
	LogDebugLev   = 4
	ConstIntNum0  = 0
	ConstIntNum1  = 1
	ConstIntNum2  = 2
	ConstIntNum3  = 3
	ConstIntNum8  = 8
	ConstIntNum10 = 10
	NPUHex        = 1000

	nodeNPUNumber       = 8
	npuNumPerHccs       = 4
	maxTaskNPUNum       = 10000
	ArchSelector        = "host-arch"
	HuaweiArchArm       = "huawei-arm"
	HuaweiArchX86       = "huawei-x86"
	accelerator         = "accelerator"
	CommCardPreName     = "huawei.com/Ascend"
	accelerator910Value = "huawei-Ascend910"
	accelerator310Value = "huawei-Ascend310"
	// Fault910NPU  get 910 Fault npus.
	Fault910NPU = "huawei.com/Ascend910-Unhealthy"
	// Fault710NPU get 710 Fault npus.
	Fault710NPU            = "huawei.com/Ascend710-Unhealthy"
	AcceleratorType        = "accelerator-type"
	CardAcceleratorType    = "card"
	ModuleAcceleratorType  = "module"
	ChipAcceleratorType    = "chip"
	NodeNoFitSelectorError = "no matching label on this node"
)
