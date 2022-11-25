// Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.

// Package util is using for the total variable.
package util

import (
	"k8s.io/api/core/v1"
	"volcano.sh/apis/pkg/apis/scheduling"
	"volcano.sh/volcano/pkg/scheduler/api"
)

const (
	// LogErrorLev for log error.
	LogErrorLev = 1
	// LogWarningLev for log warning.
	LogWarningLev = 2
	// LogInfoLev for log information.
	LogInfoLev = 3
	// LogDebugLev for log debug.
	LogDebugLev = 4
	// NPUIndex2 the 2 index.
	NPUIndex2 = 2
	// NPUIndex3 the 3 index.
	NPUIndex3 = 3
	// NPUIndex8 the 8 index.
	NPUIndex8 = 8
	// NPUIndex4 the 4 index.
	NPUIndex4 = 4
	// MapInitNum for map init length.
	MapInitNum = 3
	// Base10 for const 10.
	Base10 = 10
	// BitSize64 for const 64
	BitSize64 = 64
	// NPUHexKilo for const 1000,volcano frame used.
	NPUHexKilo = 1000
	// NPUCardPreName for NPU card pre-Name.
	NPUCardPreName = "huawei.com/Ascend"
	// ArchSelector MindX-dl arch selector.
	ArchSelector = "host-arch"
	// HuaweiArchArm for arm.
	HuaweiArchArm = "huawei-arm"
	// HuaweiArchX86 for x86.
	HuaweiArchX86 = "huawei-x86"

	// Accelerator for custom tag.
	Accelerator = "accelerator"
	// Accelerator910Value Ascend910 tag.
	Accelerator910Value = "huawei-Ascend910"
	// Accelerator310Value Ascend310 tag.
	Accelerator310Value = "huawei-Ascend310"

	// CMSelectorKey selector key in scheduler configmap.
	CMSelectorKey = "selector"
	// CMInitParamKey init param key in scheduler configmap
	CMInitParamKey = "init-params"
	// AcceleratorType for selector.
	AcceleratorType = "accelerator-type"
	// CardAcceleratorType for card mode.
	CardAcceleratorType = "card"
	// ModuleAcceleratorType for module mode.
	ModuleAcceleratorType = "module"
	// ChipAcceleratorType for chip mode.
	ChipAcceleratorType = "chip"
	// HalfAcceleratorType for half mode
	HalfAcceleratorType = "half"

	// NPU910CardName for judge 910 npu resource.
	NPU910CardName = "huawei.com/Ascend910"
	// NPU910CardNamePre for getting card number.
	NPU910CardNamePre = "Ascend910-"
	// NPU310PCardName for judge 310P npu resource.
	NPU310PCardName = "huawei.com/Ascend310P"
	// NPU310CardName for judge 310 npu resource.
	NPU310CardName = "huawei.com/Ascend310"
	// NPU310CardNamePre for getting card number.
	NPU310CardNamePre = "Ascend310-"
	// NPU310PCardNamePre for getting card number.
	NPU310PCardNamePre = "Ascend310P-"
	// AscendNPUPodRealUse for NPU pod real use cards.
	AscendNPUPodRealUse = "huawei.com/AscendReal"

	// SegmentEnable for VNPU segment enable flag. Default is "false".
	SegmentEnable = "presetVirtualDevice"
	// SegmentNoEnable SegmentEnable not enable.
	SegmentNoEnable = "SegmentEnable not enable"

	// DevInfoNameSpace device-plugin install Namespace
	DevInfoNameSpace = "kube-system"
	// DevInfoPreName like "mindx-dl-deviceinfo-ubuntu"
	DevInfoPreName = "mindx-dl-deviceinfo-"
	// DevInfoCMKey mindx-dl-deviceinfo configmap key
	DevInfoCMKey = "DeviceInfoCfg"
	// RePropertyCacheName rescheduling keyword in init env.cache
	RePropertyCacheName = "re-scheduling"
	// JobRecovery keywords for retain
	JobRecovery = "job-recovery"

	// PodPredicateTime set pod PodPredicateTime for using by device-plugin.
	PodPredicateTime = "predicate-time"
	// NodeNotMeetTopologyWarning node not satisfy the schedulable topology warning.
	NodeNotMeetTopologyWarning = "the npus on this node don't satisfy the schedulable topology"
	// ArgumentError argument nil error.
	ArgumentError = "invalid argument"
)

// NPUTask for npu task need.
type NPUTask struct {
	TaskName   string
	ReqNPUName string
	ReqNPUNum  int
	// Selector the same as job.
	Selector map[string]string
	Label    map[string]string
}

// ComJob all vcJob has.
type ComJob struct {
	JobName   api.JobID
	NameSpace string
	Selector  map[string]string
	Label     map[string]string
}

// NPUJob only npu vcJob have.
type NPUJob struct {
	ReqNPUName string
	ReqNPUNum  int
	PGStatus   scheduling.PodGroupPhase
	TaskStatus []v1.PodPhase
	CreateTime int64
	// the mapKey is taskID,not Name.
	Tasks map[string]NPUTask
}

// SchedulerJobAttr vcJob's attribute.
type SchedulerJobAttr struct {
	ComJob
	*NPUJob
}

// ComConfigMap common config map
type ComConfigMap struct {
	Name      string
	Namespace string
	Data      map[string]string
}
