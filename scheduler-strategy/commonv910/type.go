/*
Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.

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

Package commonv910 is using for virtual HuaWei Ascend910 schedule.

*/
package commonv910

const (
	// PluginName the vNPU's plugin name.
	PluginName           = "Vnpu"
	npu910CardName       = "huawei.com/Ascend910"
	npu910CardNamePrefix = "huawei.com/"
	podPredicateTime     = "predicate-time"
	npuV910CardName2c    = "huawei.com/Ascend910-2c"
	npuV910CardName4c    = "huawei.com/Ascend910-4c"
	npuV910CardName8c    = "huawei.com/Ascend910-8c"
	npuV910CardName16c   = "huawei.com/Ascend910-16c"
	npuV910CardCoef2c    = 2
	npuV910CardCoef4c    = 4
	npuV910CardCoef8c    = 8
	npuV910CardCoef16c   = 16
	npu910CardCoef       = 32
	const2               = 2
	const3               = 3

	logErrorLev   = 1
	logInfoLev    = 3
	logDebugLev   = 4
	constIntNum8  = 8
	npuHex        = 1000
	archSelector  = "host-arch"
	huaweiArchArm = "huawei-arm"
	huaweiArchX86 = "huawei-x86"

	nodeNoFitSelectorError   = "no matching label on this node"
	nodesNoMeetNPUReqError   = "insufficient Vnpus on the schedulable nodes in cluster"
	nodeNotStableWarning     = "the Vnpus on this node are unstable"
	nodeNotEnoughVnpuWarning = "insufficient number of available Vnpus on this node"
)

var (
	vnpuCoefficients = map[string]int{
		npuV910CardName2c:  npuV910CardCoef2c,
		npuV910CardName4c:  npuV910CardCoef4c,
		npuV910CardName8c:  npuV910CardCoef8c,
		npuV910CardName16c: npuV910CardCoef16c,
	}
)

// Vnpu type
type Vnpu struct {
	MaxNPUNum int
}
