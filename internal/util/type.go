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
	// NPUIndex0 for const 0.
	NPUIndex0 = 0
	// NPUIndex1 for const 1.
	NPUIndex1 = 1
	// NPUIndex2 for const 2.
	NPUIndex2 = 2
	// NPUIndex3 for const 3.
	NPUIndex3 = 3
	// NPUIndex4 for const 4.
	NPUIndex4 = 4
	// NPUIndex5 for const 5.
	NPUIndex5 = 5
	// NPUIndex6 for const 6.
	NPUIndex6 = 6
	// NPUIndex7 for const 7.
	NPUIndex7 = 7
	// NPUIndex8 for const 8.
	NPUIndex8 = 8
	// Base10 for const 10.
	Base10 = 10
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
	// Fault310PNPU get 310P Fault npus.
	Fault310PNPU = "huawei.com/Ascend310P-Unhealthy"
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
	// SegmentSetFalse presetVirtualDevice has been set as false, it must be true
	SegmentSetFalse = "presetVirtualDevice has been set as false"
	// DevInfoNameSpace device-plugin install namespace
	DevInfoNameSpace = "kube-system"
	// DevInfoPreName like "mindx-dl-deviceinfo-ubuntu"
	DevInfoPreName = "mindx-dl-deviceinfo-"
	// DevInfoCMKey mindx-dl-deviceinfo configmap key
	DevInfoCMKey = "DeviceInfoCfg"
)

// NodeDeviceInfo like node annotation.
type NodeDeviceInfo struct {
	DeviceList map[string]string
	UpdateTime int64
}

// NodeDeviceInfoWithDevPlugin a node has one by cm.
type NodeDeviceInfoWithDevPlugin struct {
	DeviceInfo NodeDeviceInfo
	CheckCode  string
}

// nodeDeviceInfoCache for record nodeDeviceInfo from device-plugin. It will be optimized later in the refactoring.
var nodeDeviceInfoCache map[string]*NodeDeviceInfo
