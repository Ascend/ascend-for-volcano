/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package test is using for HuaWei Ascend pin scheduling test.

*/
package test

import (
	"k8s.io/api/core/v1"
)

const (
	npuIndex2 = 2
	npuIndex3 = 3
	// NPUIndex4 for re-scheduler tests
	NPUIndex4 = 4
	// NPUIndex5 for re-scheduler tests
	NPUIndex5 = 5
	// NPUIndex8 for re-scheduler tests
	NPUIndex8 = 8
	// NPUHexKilo for const 1000,volcano frame used.
	NPUHexKilo   = 1000
	podRankIndex = "hccl/rankIndex"
	// NPU910CardName 910 card name
	NPU910CardName = "huawei.com/Ascend910"
	// AscendNPUPodRealUse for NPU pod real use cards.
	AscendNPUPodRealUse = "huawei.com/AscendReal"
	// FakeUpdateTime fake update time for test
	FakeUpdateTime = int64(11110)
)

// NPUPod test NPU pod struct
type NPUPod struct {
	Namespace, Name, NodeName, GroupName string
	Phase                                v1.PodPhase
	ReqSource                            v1.ResourceList
	Labels, Selector                     map[string]string
}

// NPUNode test NPU node struct
type NPUNode struct {
	Name                         string
	Capacity, Allocatable        v1.ResourceList
	Labels, Selector, Annotation map[string]string
	Other                        map[string]interface{}
}
