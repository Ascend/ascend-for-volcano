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
	// ErrorInt return -1 when get error for int
	ErrorInt = -1
	// NPUIndex2 the 2 index.
	NPUIndex2 = 2
	// NPUIndex3 the 3 index.
	NPUIndex3 = 3
	// NPUIndex8 the 8 index.
	NPUIndex8 = 8
	// NPUIndex7 the 7 index.
	NPUIndex7 = 7
	// NPUIndex4 the 4 index.
	NPUIndex4 = 4
	// NPUIndex5 the 5 index.
	NPUIndex5 = 5
	// NPUIndex1 the 1 index.
	NPUIndex1 = 1
	// CoreNum32 32 core 910
	CoreNum32 = 32
	// CoreNum30 30 core 910
	CoreNum30 = 30
	// CpuNum14 14 cpu 910
	CpuNum14 = 14
	// JobTemplateNum 11 vnpu job template
	JobTemplateNum = 11
	// MapInitNum for map init length.
	MapInitNum = 3
	// Base10 for const 10.
	Base10 = 10
	// BitSize64 for const 64
	BitSize64 = 64
	// NPUHexKilo for const 1000,volcano frame used.
	NPUHexKilo = 1000
	// AICPU cpu resource name
	AICPU = "cpu"
	// HwPreName pre name
	HwPreName = "huawei.com/"
	// NPUCardPreName for NPU card pre-Name.
	NPUCardPreName = "huawei.com/Ascend"
	// ArchSelector MindX-dl arch selector.
	ArchSelector = "host-arch"
	// HuaweiArchArm for arm.
	HuaweiArchArm = "huawei-arm"
	// HuaweiArchX86 for x86.
	HuaweiArchX86 = "huawei-x86"
	// RingController ascend-310P/ascend-310/ascend-910
	RingController = "ring-controller.atlas"

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
	// ServerType server type value takes Ascend310P-10-dual/Ascend910-32...
	ServerType = "servertype"
	// ServerTypeDual dual card
	ServerTypeDual = "dual"

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
	// AscendNPUCore for NPU core num
	AscendNPUCore = "huawei.com/npu-core"

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

// VTemplate for vNode resource
type VTemplate struct {
	// ChipKind Ascend910/Ascend310P
	ChipKind   string
	AICore     int
	AICPU      int
	DVPPEnable bool
}

// SchedulerJobAttr vcJob's attribute.
type SchedulerJobAttr struct {
	ComJob
	*NPUJob
}

// VResource resource dimensions
type VResource struct {
	Aicore int
	Aicpu  int
	DVPP   string
}
