/*
Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.
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

// ClusterNodePredicateFn pre-select cluster processing.
type ClusterNodePredicateFn func(*api.TaskInfo, *framework.Session) error
