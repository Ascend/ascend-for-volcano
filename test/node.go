/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package test is using for HuaWei Ascend pin scheduling test.

*/
package test

import (
	"strconv"
	"strings"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// SetNPUNodeLabel set NPU node label.
func SetNPUNodeLabel(node *v1.Node, labelKey string, labelValue string) {
	if labelValue == "" {
		return
	}

	if node.Labels == nil {
		node.Labels = make(map[string]string, npuIndex3)
	}

	node.Labels[labelKey] = labelValue
}

// SetTestNPUNodeOther set NPU node other for add npu resource.
func SetTestNPUNodeOther(node *api.NodeInfo, otherKey, otherValue string) {
	if node.Others == nil {
		node.Others = make(map[string]interface{}, npuIndex3)
	}

	node.Others[otherKey] = otherValue
}

// SetTestNPUNodeAnnotation set NPU node annotation for add fault npu resource.
func SetTestNPUNodeAnnotation(node *api.NodeInfo, name string, value string) {
	SetTestNPUNodeOther(node, name, value)
	if node.Node.Annotations == nil {
		node.Node.Annotations = make(map[string]string, npuIndex3)
	}

	node.Node.Annotations[name] = value

	stringSlice := strings.Split(value, ",")
	if node.Allocatable == nil || len(node.Allocatable.ScalarResources) == 0 {
		allo := api.Resource{ScalarResources: map[v1.ResourceName]float64{
			v1.ResourceName(name): float64(len(stringSlice)) * util.NPUHexKilo}}
		node.Allocatable = &allo
		return
	}
	node.Allocatable.ScalarResources[v1.ResourceName(name)] = float64(len(stringSlice)) * util.NPUHexKilo
}

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

// SetFakeNodeIdleSource Set fake node the idle source.
func SetFakeNodeIdleSource(nodeInf *api.NodeInfo, name string, value int) {
	idle := api.Resource{ScalarResources: map[v1.ResourceName]float64{
		v1.ResourceName(name): float64(value) * util.NPUHexKilo}}
	nodeInf.Idle = &idle
}

// BuildNodeWithFakeOther build node with fake node other
func BuildNodeWithFakeOther(nodeName, name string, value string) *api.NodeInfo {
	nodeInfo := FakeNormalTestNode(nodeName)
	if value != "" {
		SetTestNPUNodeOther(nodeInfo, name, value)
	}
	return nodeInfo
}

// BuildUnstableNode build unstable node
func BuildUnstableNode(nodeName, resourceName, otherNpuNum string, idleNpuNum int) *api.NodeInfo {
	nodeInfo := FakeNormalTestNode(nodeName)
	if otherNpuNum != "" {
		SetTestNPUNodeOther(nodeInfo, resourceName, otherNpuNum)
	}
	if idleNpuNum != 0 {
		SetFakeNodeIdleSource(nodeInfo, resourceName, idleNpuNum)
	}
	return nodeInfo
}
