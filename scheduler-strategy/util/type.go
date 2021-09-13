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
	acceleratorValue       = "huawei-Ascend910"
	acceleratorType        = "accelerator-type"
	cardAcceleratorType    = "card"
	moduleAcceleratorType  = "module"
	nodeNoFitSelectorError = "no matching label on this node"
)
