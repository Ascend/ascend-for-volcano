/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/*

Package vnpu is using for HuaWei Ascend pin fault rescheduling.

*/
package vnpu

import (
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/base"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

const (
	// VJobStatusUnhandled not handled
	VJobStatusUnhandled = "Unhandled"
	// VJobStatusNotPreSegmented not pre-allocated
	VJobStatusNotPreSegmented = "NotPreSegmented"
	// VJobStatusPreSegmented pre-allocated
	VJobStatusPreSegmented = "PreSegmented"
	// VJobStatusSegmented segmented
	VJobStatusSegmented = "Segmented"
	// VJobStatusAllocated allocated
	VJobStatusAllocated = "Allocated"
	// VJobStatusDestroying destroying
	VJobStatusDestroying = "Destroying"
	// VJobStatusDestroyed destroyed
	VJobStatusDestroyed = "Destroyed"

	// VNPUCMNameSpace for uninstall volcano, also delete cm
	VNPUCMNameSpace = "volcano-system"
	// VNPUCMName the cm intercommunicate to device-plugin.
	VNPUCMName = "mindx-dl-vnpu-manager"
	// VNPUCMDataKey cm date key
	VNPUCMDataKey = "VNPUCfg"
	// VNPUCacheCMName solidified the vnpu pre-alloc cache.
	VNPUCacheCMName = "mindx-dl-vnpu-cache"
	// VNPUNodeLabelKey for select vnpu node label key.
	VNPUNodeLabelKey = "npu-spec"
	// VNPUNodeLabelValue for select vnpu node label value.
	VNPUNodeLabelValue = "vnpu"
	// DeleteOverTime over time for job finish deal.
	DeleteOverTime = 5
	// JobPendingWaitTime The time wait for device-plugin create vnpu.
	JobPendingWaitTime = 300
	// VNPUScoreWeight for volcano select vnpu node core.
	VNPUScoreWeight = 64
	// PreAllocateFailureWaitTime wait time to judge pre-allocation failure
	PreAllocateFailureWaitTime = 10
)

// ComVNPU vnpu struct
type ComVNPU struct {
	*ComVNPUHandler
	// The vNPU chip divide name. Like huawei.com/Ascend910-16c,huawei.com/Ascend910-8c and so on.
	DivideKinds []string
	// divide vNPU coefficient for each chip.
	Coefficients map[string]int
	// the source of NPU cores in node annotation. like huawei.com/Ascend910-spec.
	NPUCardCoreKey string
}

// ComVNPUHandler vnpu handler
type ComVNPUHandler struct {
	*Action
	vNodes map[string]VNode
	base.NPUHandler
}

type staticVNPUHandler struct {
	*ComVNPUHandler
}

type dynamicVNPUHandler struct {
	*ComVNPUHandler
	*VCache
	dpVConfigMap *util.ComConfigMap
}

// VResource resource dimensions
type VResource struct {
	Aicore int
	Aicpu  int
	Vpc    int
	Vdec   int
	Jpegd  int
	Pngd   int
	Venc   int
	Jpege  int
}

// VChip vnpu chip
type VChip struct {
	cardName          string
	cardType          string
	segmentFlag       bool
	wholeChipUsedFlag bool
	allocatedRes      VResource
	segmentedRes      VResource
	unsegmentedRes    VResource
	mountedCoreNum    int
	segmentingCoreNum int
	vGroupFragNum     int
}

// VNode vnpu node
type VNode struct {
	nodeName           string
	nodeCardType       string
	nodeCardNum        int
	nodeChips          map[string]VChip
	nodeAllocatableRes VResource
	nodeSegmentedRes   VResource
	nodeUnsegmentedRes VResource
}

// Action vnpu actions
type Action struct {
	template map[string]VResource
}

// VCache vnpu cache
type VCache struct {
	vConfigMap *util.ComConfigMap
	vJobs      map[api.JobID]VJob
	checkCode  string
}

// VJob vnpu job
type VJob struct {
	jobUID        api.JobID
	jobStatus     string
	reqVNPUType   string
	reqNodeName   string
	reqCardName   string
	taskNum       int
	allocCardName string
	allocFlag     bool
	resourceReq   VResource
	createTime    int64
	allocTime     int64
	updateTime    int64
}

// VJobList struct for sorting
type VJobList []VJob

type dpvConfigMap struct {
	util.ComConfigMap
}
