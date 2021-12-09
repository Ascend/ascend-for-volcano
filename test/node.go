/*
Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package ascendtest is using for HuaWei Ascend pin scheduling test.

*/
package ascendtest

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
	"volcano.sh/volcano/pkg/scheduler/api"
)

// SetNPUNodeLabel set NPU node label.
func SetNPUNodeLabel(node *v1.Node, labelKey string, labelValue string) {
	if labelValue == "" {
		delete(node.Labels, labelKey)
		return
	}

	if node.Labels == nil {
		node.Labels = make(map[string]string, constIntNum3)
	}

	node.Labels[labelKey] = labelValue
}

// SetTestNPUNodeOther set NPU node other for add npu resource.
func SetTestNPUNodeOther(node *api.NodeInfo, key string, value string) {
	if node.Others == nil {
		node.Others = make(map[string]interface{}, constIntNum3)
	}

	node.Others[key] = value
}

// SetTestNPUNodeAnnotation set NPU node other for add fault npu resource.
func SetTestNPUNodeAnnotation(node *api.NodeInfo, key string, value string) {
	if node.Node.Annotations == nil {
		node.Node.Annotations = make(map[string]string, constIntNum3)
	}

	node.Node.Annotations[key] = value
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

// FakeNormalTestNodes fake normal test nodes.
func FakeNormalTestNodes(num int) []*api.NodeInfo {
	var nodes []*api.NodeInfo

	for i := 0; i < num; i++ {
		strNum := strconv.Itoa(i)
		node := NPUNode{
			Name:        "node" + strNum,
			Capacity:    make(v1.ResourceList, constIntNum3),
			Allocatable: make(v1.ResourceList, constIntNum3),
			Labels:      make(map[string]string, constIntNum3),
			Selector:    make(map[string]string, constIntNum3),
			Annotation:  make(map[string]string, constIntNum3),
			Other:       make(map[string]interface{}, constIntNum3),
		}
		nodeInfo := api.NewNodeInfo(BuildNPUNode(node))
		nodes = append(nodes, nodeInfo)
	}

	return nodes
}
