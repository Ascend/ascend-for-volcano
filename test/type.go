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
	npuIndex3 = 3
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
