/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package comvnpu is using for virtual HuaWei Ascend910 schedule.

*/
package comvnpu

import (
	"strconv"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/conf"
	"volcano.sh/volcano/pkg/scheduler/util"

	util2 "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/vnpu/modulev310p"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/vnpu/vnpuutil"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
)

const (
	nodeName           = "ubuntu"
	labelSize          = 8
	annSize            = 8
	npuV910CardName16c = "huawei.com/Ascend910-16c"
)

type VNodeInfo struct {
	nodeName       string
	nodeArch       string
	cpu, mem       string
	npuAllocateNum string
	npuTop         string
}

type VPodInfo struct {
	namespace  string
	groupName  string
	podName    string
	nodeName   string
	reqCPUNum  string
	reqMem     string
	reqNPUType string
	reqNpuNum  string
}

// TestVnpuIsMyNode
func TestVnpuIsMyNode(t *testing.T) {
	convey.Convey("Test IsMyNode in commonv910", t, func() {
		vnpu := &VNPU{}
		convey.Convey("IsMyNode() should return error since it's not implemented", func() {
			node := buildNPUNode(VNodeInfo{nodeName: nodeName, nodeArch: util2.HuaweiArchArm, cpu: "192", mem: "755Gi",
				npuAllocateNum: "1", npuTop: "Ascend910-16c-118-0"})
			result := vnpu.IsMyNode(node)
			convey.So(result, convey.ShouldBeError)
		})
	})
}

// TestVnpuName
func TestVnpuName(t *testing.T) {
	convey.Convey("Test Vnpu Name", t, func() {
		const (
			pluginName = "Vnpu"
		)
		vnpu2 := &VNPU{Plugin: nil, Attr: vnpuutil.ComVNPU{
			HwEntity: plugin.HwEntity{PluginName: PluginName}}, HwNPUSchedulerPlugin: nil}
		convey.Convey("Test-1 Name() should return PluginName defined in const", func() {
			n := vnpu2.Name()
			convey.So(n, convey.ShouldEqual, pluginName)
		})
	})
}

// TestVnpuPreCheckNodeFnTaskError
func TestVnpuPreCheckNodeFnTaskError(t *testing.T) {
	convey.Convey("Test job PreCheckNodeFn TaskError", t, func() {
		var confs []conf.Configuration
		node := buildNPUNode(VNodeInfo{nodeName: nodeName, nodeArch: util2.HuaweiArchX86, cpu: "192", mem: "755Gi",
			npuAllocateNum: "1", npuTop: "Ascend910-16c-131-1"})

		convey.Convey("PreCheckNodeFn() should return error when task don't have selector", func() {
			vnpu := &VNPU{}
			pod := buildNPUPod(VPodInfo{namespace: "default", groupName: "npu-group-8",
				podName: "npu-test-10", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: npuV910CardName16c, reqNpuNum: "1"})
			delete(pod.Spec.NodeSelector, util2.ArchSelector)
			task := api.NewTaskInfo(pod)
			confs = []conf.Configuration{{Name: util2.CMInitParamKey,
				Arguments: map[string]string{util2.SegmentEnable: "false"}}}
			result := vnpu.PreCheckNodeFn(task, node, confs)
			convey.So(result, convey.ShouldNotBeNil)
		})
		convey.Convey("PreCheckNodeFn() should return nil when task don't have selector and no resource requested", func() {
			vnpu := &VNPU{}
			vnpu310P := &modulev310p.ChipV310P{}
			if getErr := vnpu310P.InitVNPUPlugin(); getErr != nil {
				return
			}
			vnpu.Attr = vnpu310P.ComVNPU
			pod := buildNPUPod(VPodInfo{namespace: "default", groupName: "npu-group-9",
				podName: "npu-test-11'", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: "", reqNpuNum: ""})
			delete(pod.Spec.NodeSelector, util2.ArchSelector)
			task := api.NewTaskInfo(pod)
			confs = []conf.Configuration{{Name: util2.CMInitParamKey,
				Arguments: map[string]string{util2.SegmentEnable: "false"}}}
			result := vnpu.PreCheckNodeFn(task, node, confs)
			convey.So(result, convey.ShouldNotBeNil)
		})
	})
}

// TestVnpuPreCheckNodeFnNodeError
func TestVnpuPreCheckNodeFnNodeError(t *testing.T) {
	convey.Convey("Test job PreCheckNodeFn NodeError", t, func() {
		vnpu := &VNPU{}
		var confs []conf.Configuration
		node := buildNPUNode(VNodeInfo{nodeName: nodeName, nodeArch: util2.HuaweiArchX86, cpu: "192", mem: "755Gi",
			npuAllocateNum: "1", npuTop: "Ascend910-16c-119-1"})

		convey.Convey("PreCheckNodeFn() should return error when node don't have label", func() {
			task := api.NewTaskInfo(buildNPUPod(VPodInfo{namespace: "default", groupName: "npu-group-10",
				podName: "npu-test-12", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: npuV910CardName16c, reqNpuNum: "1"}))
			test.SetNPUNodeLabel(node.Node, util2.ArchSelector, "")
			confs = []conf.Configuration{{Name: util2.CMInitParamKey,
				Arguments: map[string]string{util2.SegmentEnable: "false"}}}
			result := vnpu.PreCheckNodeFn(task, node, confs)
			convey.So(result, convey.ShouldNotBeNil)
		})
		convey.Convey("PreCheckNodeFn() should return error when selectors mismatch with labels", func() {
			task := api.NewTaskInfo(buildNPUPod(VPodInfo{namespace: "default", groupName: "npu-group-11",
				podName: "npu-test-13", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: npuV910CardName16c, reqNpuNum: "1"}))
			// build a node with mismatch selector
			nodeArm := buildNPUNode(VNodeInfo{nodeName: nodeName, nodeArch: util2.HuaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "2", npuTop: "Ascend910-16c-140-1,Ascend910-16c-141-1"})
			confs = []conf.Configuration{{Name: util2.CMInitParamKey,
				Arguments: map[string]string{util2.SegmentEnable: "false"}}}
			result := vnpu.PreCheckNodeFn(task, nodeArm, confs)
			convey.So(result, convey.ShouldNotBeNil)
		})
	})
}

// TestVnpuPreCheckNodeFnSuccess
func TestVnpuPreCheckNodeFnSuccess(t *testing.T) {
	convey.Convey("Test job PreCheckNodeFn Success", t, func() {
		vnpu := &VNPU{}
		var confs []conf.Configuration
		node := buildNPUNode(VNodeInfo{nodeName: nodeName, nodeArch: util2.HuaweiArchX86, cpu: "192", mem: "755Gi",
			npuAllocateNum: "1", npuTop: "Ascend910-16c-132-1"})

		convey.Convey("PreCheckNodeFn() should return nil when selectors match with labels", func() {
			// build a task with no selector
			task := api.NewTaskInfo(buildNPUPod(VPodInfo{namespace: "default", groupName: "npu-group-12",
				podName: "npu-test-14", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: npuV910CardName16c, reqNpuNum: "1"}))
			confs = []conf.Configuration{{Name: util2.CMInitParamKey,
				Arguments: map[string]string{util2.SegmentEnable: "false"}}}
			confs = []conf.Configuration{{Name: util2.CMInitParamKey,
				Arguments: map[string]string{util2.SegmentEnable: "false"}}}
			result := vnpu.PreCheckNodeFn(task, node, confs)
			convey.So(result, convey.ShouldNotBeNil)
		})
	})
}

// TestVnpuCheckNodeNPUByTaskFn
func TestVnpuCheckNodeNPUByTaskFn(t *testing.T) {
	convey.Convey("Test job CheckNodeNPUByTaskFn", t, func() {
		vnpu := &VNPU{}
		pod := buildNPUPod(VPodInfo{namespace: "default", groupName: "npu-group-13",
			podName: "npu-test-15", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
			reqNPUType: npuV910CardName16c, reqNpuNum: "1"})
		task := api.NewTaskInfo(pod)

		convey.Convey("CheckNodeNPUByTaskFn() should return nil", func() {
			node := buildNPUNode(VNodeInfo{nodeName: nodeName, nodeArch: util2.HuaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "0", npuTop: ""})
			result := vnpu.CheckNodeNPUByTaskFn(task, node, true)
			convey.So(result, convey.ShouldBeNil)
		})
		convey.Convey("CheckNodeNPUByTaskFn() should return error when node meets task requests", func() {
			node := buildNPUNode(VNodeInfo{nodeName: nodeName, nodeArch: util2.HuaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "1", npuTop: "Ascend910-16c-199-1"})
			result := vnpu.CheckNodeNPUByTaskFn(task, node, true)
			convey.So(result, convey.ShouldBeNil)
		})
	})
}

// TestVnpuUpdateReleaseNPUNodeTopologyFn
func TestVnpuUpdateReleaseNPUNodeTopologyFn(t *testing.T) {
	convey.Convey("Test UpdateReleaseNPUNodeTopologyFn", t, func() {
		vnpu := &VNPU{}
		node := buildNPUNode(VNodeInfo{nodeName: nodeName, nodeArch: util2.HuaweiArchX86, cpu: "192", mem: "755Gi",
			npuAllocateNum: "3", npuTop: "Ascend910-16c-112-1,Ascend910-16c-113-1,Ascend910-16c-114-1"})
		convey.Convey("UpdateReleaseNPUNodeTopologyFn() should return error when content of top is invalid", func() {
			top := []string{"Ascend910-5c-111-0"}
			result := vnpu.UpdateReleaseNPUNodeTopologyFn(node, top)
			convey.So(result, convey.ShouldBeError)
		})
		convey.Convey("UpdateReleaseNPUNodeTopologyFn() should successfully update node.Others", func() {
			node.Others = map[string]interface{}{
				npuV910CardName16c: "Ascend910-16c-111-0,Ascend910-16c-112-0,Ascend910-16c-113-0",
			}
			top := []string{"Ascend910-16c-114-0"}
			result := vnpu.UpdateReleaseNPUNodeTopologyFn(node, top)
			convey.So(result, convey.ShouldNotBeNil)
		})
	})
}

// TestVnpuinitVNodesFn
func TestVnpuinitVNodesFn(t *testing.T) {
	convey.Convey("Test initVNodesFn", t, func() {
		convey.Convey("", func() {
			vnpu := &VNPU{}
			node1 := buildNPUNode(VNodeInfo{nodeName: "ubuntu-1", nodeArch: util2.HuaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "1", npuTop: "Ascend910-16c-117-0"})
			node2 := buildNPUNode(VNodeInfo{nodeName: "ubuntu-2", nodeArch: util2.HuaweiArchArm, cpu: "192", mem: "755Gi",
				npuAllocateNum: "1", npuTop: "Ascend910-16c-118-0"})
			node3 := buildNPUNode(VNodeInfo{nodeName: "ubuntu-3", nodeArch: util2.HuaweiArchArm, cpu: "192", mem: "755Gi",
				npuAllocateNum: "1", npuTop: "Ascend910-16c-119-0"})
			test.SetTestNPUNodeAnnotation(node3, npuV910CardName16c, "")
			nodes := map[string]*api.NodeInfo{
				"1": node1,
				"2": node2,
				"3": node3,
			}
			result := vnpu.InitVNodesFn(nodes)
			convey.So(result, convey.ShouldBeNil)
		})
	})
}

func buildNPUPod(podInfo VPodInfo) *v1.Pod {

	pod := util.BuildPod(podInfo.namespace, podInfo.podName, podInfo.nodeName, v1.PodPending,
		buildNPUResourceList(podInfo.reqCPUNum, podInfo.reqMem,
			v1.ResourceName(podInfo.reqNPUType), podInfo.reqNpuNum),
		podInfo.groupName, make(map[string]string, util2.NPUIndex2),
		make(map[string]string, util2.NPUIndex2))

	test.SetTestNPUPodSelector(pod, util2.ArchSelector, util2.HuaweiArchX86)

	return pod
}

func buildNPUResourceList(cpu string, memory string, npuResourceType v1.ResourceName, npu string) v1.ResourceList {
	npuNum, err := strconv.Atoi(npu)
	if err != nil {
		return nil
	}

	if npuNum == 0 {
		return v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse(cpu),
			v1.ResourceMemory: resource.MustParse(memory),
		}
	}

	return v1.ResourceList{
		v1.ResourceCPU:    resource.MustParse(cpu),
		v1.ResourceMemory: resource.MustParse(memory),
		npuResourceType:   resource.MustParse(npu),
	}
}

func buildNPUNode(nodeInfo VNodeInfo) *api.NodeInfo {
	nodeCapacity := buildNPUResourceList(nodeInfo.cpu, nodeInfo.mem, npuV910CardName16c, strconv.Itoa(1))
	nodeAlloc := buildNPUResourceList(nodeInfo.cpu, nodeInfo.mem, npuV910CardName16c, nodeInfo.npuAllocateNum)
	labels := make(map[string]string, labelSize)
	ann := make(map[string]string, annSize)

	v1node := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:        nodeInfo.nodeName,
			Labels:      labels,
			Annotations: ann,
		},
		Status: v1.NodeStatus{
			Capacity:    nodeCapacity,
			Allocatable: nodeAlloc,
		},
	}

	if nodeInfo.npuAllocateNum != "0" {
		v1node.Annotations[npuV910CardName16c] = nodeInfo.npuTop
	}

	test.SetNPUNodeLabel(v1node, util2.ArchSelector, nodeInfo.nodeArch)
	node := api.NewNodeInfo(v1node)
	if nodeInfo.npuAllocateNum != "0" {
		node.Others = map[string]interface{}{
			npuV910CardName16c: nodeInfo.npuTop,
		}
	}
	return node
}
