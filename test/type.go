/*
Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package ascendtest is using for HuaWei Ascend pin scheduling test.

*/
package ascendtest

import v1 "k8s.io/api/core/v1"

const (
	constIntNum3 = 3
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
