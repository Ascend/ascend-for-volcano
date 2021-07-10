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

Package npuinterface is using for HuaWei Ascend pin affinity schedule frame interface.

*/
package npuinterface

import (
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"
)

// InitNodesNPUTopologyFn Init all npu nodes's topology.
type InitNodesNPUTopologyFn func(map[string]*api.NodeInfo) error

// PreHandleFaultNPUFn handle NPU fault chip.
type PreHandleFaultNPUFn func(*framework.Session) error
