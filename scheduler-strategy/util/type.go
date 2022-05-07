/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package util is using for HuaWei Ascend9 pin affinity schedule utilities.

*/
package util

const (
	// LogErrorLev for log error.
	LogErrorLev = 1
	// LogInfoLev for log information.
	LogInfoLev = 3
	// LogDebugLev for log debug.
	LogDebugLev = 4
	// ConstIntNum0 for const 0.
	ConstIntNum0 = 0
	// ConstIntNum1 for const 1.
	ConstIntNum1 = 1
	// ConstIntNum2 for const 2.
	ConstIntNum2 = 2
	// ConstIntNum3 for const 3.
	ConstIntNum3 = 3
	// ConstIntNum8 for const 8.
	ConstIntNum8 = 8
	// ConstIntNum10 for const 10.
	ConstIntNum10 = 10
	// NPUHex for const 1000,volcano frame used.
	NPUHex = 1000

	nodeNPUNumber = 8
	npuNumPerHccs = 4
	// ArchSelector MindX-dl arch selector.
	ArchSelector = "host-arch"
	// HuaweiArchArm for arm.
	HuaweiArchArm = "huawei-arm"
	// HuaweiArchX86 for x86.
	HuaweiArchX86 = "huawei-x86"
	accelerator   = "accelerator"
	// CommCardPreName for NPU card pre-name.
	CommCardPreName     = "huawei.com/Ascend"
	accelerator910Value = "huawei-Ascend910"
	accelerator310Value = "huawei-Ascend310"
	// Fault910NPU  get 910 Fault npus.
	Fault910NPU = "huawei.com/Ascend910-Unhealthy"
	// Fault710NPU get 710 Fault npus.
	Fault710NPU = "huawei.com/Ascend710-Unhealthy"
	// AcceleratorType for selector.
	AcceleratorType = "accelerator-type"
	// CardAcceleratorType for card mode.
	CardAcceleratorType = "card"
	// ModuleAcceleratorType for module mode.
	ModuleAcceleratorType = "module"
	// ChipAcceleratorType for chip mode.
	ChipAcceleratorType = "chip"
	// NodeNoFitSelectorError for node no fit selector error.
	NodeNoFitSelectorError = "no matching label on this node"
	// CMInitParamKey init param key in scheduler configmap
	CMInitParamKey = "init-params"
	// CMSelectorKey selector key in scheduler configmap
	CMSelectorKey = "selector"
	// SegmentEnable for VNPU segment enable flag. Default is "false".
	SegmentEnable = "presetVirtualDevice"
	// SegmentNoEnable SegmentEnable not enable.
	SegmentNoEnable = "SegmentEnable not enable"
)
