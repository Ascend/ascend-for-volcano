/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package test is using for HuaWei Ascend pin scheduling test.
*/
package test

import (
	"strconv"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"volcano.sh/volcano/pkg/scheduler/api"
)

// BuildNPUNode built NPU node object
func BuildNPUNode(node NPUNode) *v1.Node {
	return &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:        node.Name,
			Labels:      node.Labels,
			Annotations: node.Annotation,
		},
		Status: v1.NodeStatus{
			Capacity:    node.Capacity,
			Allocatable: node.Allocatable,
		},
	}
}

// FakeNormalTestNode fake normal test node.
func FakeNormalTestNode(name string) *api.NodeInfo {
	node := NPUNode{
		Name:        name,
		Capacity:    make(v1.ResourceList, npuIndex3),
		Allocatable: make(v1.ResourceList, npuIndex3),
		Labels:      make(map[string]string, npuIndex3),
		Selector:    make(map[string]string, npuIndex3),
		Annotation:  make(map[string]string, npuIndex3),
		Other:       make(map[string]interface{}, npuIndex3),
	}
	nodeInfo := api.NewNodeInfo(BuildNPUNode(node))
	return nodeInfo
}

// FakeNormalTestNodes fake normal test nodes.
func FakeNormalTestNodes(num int) []*api.NodeInfo {
	var nodes []*api.NodeInfo

	for i := 0; i < num; i++ {
		strNum := strconv.Itoa(i)
		nodeInfo := FakeNormalTestNode("node" + strNum)
		nodes = append(nodes, nodeInfo)
	}

	return nodes
}

// SetFakeNodeSource Set fake node the idle, Capability, Allocatable source.
func SetFakeNodeSource(nodeInf *api.NodeInfo, name string, value int) {
	idle := api.Resource{ScalarResources: map[v1.ResourceName]float64{
		v1.ResourceName(name): float64(value) * NPUHexKilo}}
	nodeInf.Idle = &idle
	Capability := api.Resource{ScalarResources: map[v1.ResourceName]float64{
		v1.ResourceName(name): float64(value) * NPUHexKilo}}
	nodeInf.Capability = &Capability
	nodeInf.Allocatable = &Capability
}
