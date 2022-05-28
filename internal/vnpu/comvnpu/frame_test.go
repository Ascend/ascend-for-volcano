/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package comvnpu is using for virtual HuaWei Ascend910 schedule.

*/
package comvnpu

import (
	"fmt"
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
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/vnpu/modulev910"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/vnpu/vnpuutil"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
)

const (
	nodeName             = "ubuntu"
	labelSize            = 8
	annSize              = 8
	constIntNum2         = 2
	constIntNum3         = 3.0
	constIntNum4         = 4
	npuV910CardName16c   = "huawei.com/Ascend910-16c"
	npuV910CardName2c    = "huawei.com/Ascend910-2c"
	npuV310PCardName2c   = "huawei.com/Ascend310P-2c"
	invalidResourceType  = "huawei.com/Ascend910-5c"
	noPrefixResourceType = "no-prefix"
	validNum             = 1000
	invalidNum           = 2000
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

// TestVnpuIsMyTask
func TestVnpuIsMyTask(t *testing.T) {
	convey.Convey("Test IsMyTask in commonv910", t, func() {
		vnpu := &VNPU{}
		convey.Convey("IsMyTask() should return error since it's not implemented", func() {
			pod := buildNPUPod(VPodInfo{namespace: "default", groupName: "npu-4p", podName: "npu-4p",
				nodeName: nodeName, reqCPUNum: "10", reqMem: "20Gi", reqNPUType: npuV910CardName16c, reqNpuNum: "4"})
			task := api.NewTaskInfo(pod)
			result := vnpu.IsMyTask(task)
			convey.So(result, convey.ShouldBeNil)
		})
	})
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

// TestVnpuIsMyJob
func TestVnpuIsMyJob(t *testing.T) {
	convey.Convey("Test IsMyJob in commonv910", t, func() {
		vnpu := &VNPU{}
		var tasks []*api.TaskInfo
		uid := api.JobID("xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx1")

		convey.Convey("IsMyJob() should return error since it's not implemented", func() {
			tasks = append(tasks, api.NewTaskInfo(
				buildNPUPod(VPodInfo{namespace: "default", groupName: "npu-group-0",
					podName: "npu-test-1", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
					reqNPUType: npuV910CardName16c, reqNpuNum: "1"})))
			tasks = append(tasks, api.NewTaskInfo(
				buildNPUPod(VPodInfo{namespace: "default", groupName: "npu-group-1",
					podName: "npu-test-2", nodeName: nodeName, reqCPUNum: "10", reqMem: "5Gi",
					reqNPUType: npuV910CardName16c, reqNpuNum: "2"})))
			job := api.NewJobInfo(uid, tasks...)
			test.AddTestJobPodGroup(job)
			result := vnpu.IsMyJob(job)
			convey.So(result, convey.ShouldNotBeNil)
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

// TestVnpuValidNPUJobFnInvalidSelector
func TestVnpuValidNPUJobFnInvalidSelector(t *testing.T) {
	convey.Convey("Test job ValidNPUJobFn InvalidSelector", t, func() {
		const (
			invalidSelectorKey   = "invalid-key"
			invalidSelectorValue = "no-selector"
			validNum             = 1000
		)
		vnpu := &VNPU{}
		var tasks []*api.TaskInfo
		uid := api.JobID("xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx2")

		convey.Convey("ValidNPUJobFn() should return error for job without certain selector key", func() {
			pod := buildNPUPod(VPodInfo{namespace: "default", groupName: "npu-group-2",
				podName: "npu-test-5", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: npuV910CardName16c, reqNpuNum: "1"})
			delete(pod.Spec.NodeSelector, util2.ArchSelector)
			setPodSelector(pod, invalidSelectorKey, invalidSelectorValue)
			tasks = append(tasks, api.NewTaskInfo(pod))
			job := api.NewJobInfo(uid, tasks...)
			setJobResourceReq(job, npuV910CardName16c, float64(validNum))
			result := vnpu.ValidNPUJobFn(job)
			convey.So(result, convey.ShouldNotBeNil)
		})
		convey.Convey("ValidNPUJobFn() should return error for job with invalid selector value", func() {
			pod := buildNPUPod(VPodInfo{namespace: "default", groupName: "npu-group-3",
				podName: "npu-test-6", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: npuV910CardName16c, reqNpuNum: "1"})
			setPodSelector(pod, util2.ArchSelector, invalidSelectorValue)
			tasks = append(tasks, api.NewTaskInfo(pod))
			job := api.NewJobInfo(uid, tasks...)
			test.AddTestJobPodGroup(job)
			setJobResourceReq(job, npuV910CardName16c, float64(validNum))
			result := vnpu.ValidNPUJobFn(job)
			convey.So(result, convey.ShouldNotBeNil)
		})
	})
}

// TestVnpuValidNPUJobFnInvalidReq
func TestVnpuValidNPUJobFnInvalidReq(t *testing.T) {
	convey.Convey("Test job ValidNPUJobFn InvalidReq", t, func() {
		vnpu := &VNPU{}
		vnpu910 := &modulev910.ChipV910{}
		if getErr := vnpu910.InitVNPUPlugin(); getErr != nil {
			return
		}
		vnpu.Attr = vnpu910.ComVNPU
		var tasks []*api.TaskInfo
		uid := api.JobID("xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx3")

		convey.Convey("ValidNPUJobFn() should return error for job with invalid resource prefix", func() {
			pod := buildNPUPod(VPodInfo{namespace: "default", groupName: "npu-group-4", podName: "npu-test-7",
				nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi", reqNPUType: npuV910CardName16c, reqNpuNum: "1"})
			setPodSelector(pod, util2.ArchSelector, util2.HuaweiArchArm)
			tasks = append(tasks, api.NewTaskInfo(pod))
			job := api.NewJobInfo(uid, tasks...)
			setJobResourceReq(job, noPrefixResourceType, float64(validNum))
			test.AddTestJobPodGroup(job)
			result := vnpu.ValidNPUJobFn(job)
			convey.So(result, convey.ShouldNotBeEmpty)
		})
		convey.Convey("ValidNPUJobFn() should return error for job with invalid npu request type",
			func() {
				pod := buildNPUPod(VPodInfo{namespace: "default", groupName: "npu-group-5",
					podName: "npu-test-8", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
					reqNPUType: npuV910CardName16c, reqNpuNum: "1"})
				setPodSelector(pod, util2.ArchSelector, util2.HuaweiArchX86)
				tasks = append(tasks, api.NewTaskInfo(pod))
				job := api.NewJobInfo(uid, tasks...)
				setJobResourceReq(job, invalidResourceType, float64(validNum))
				test.AddTestJobPodGroup(job)
				result := vnpu.ValidNPUJobFn(job)
				convey.So(result, convey.ShouldNotBeNil)
			})
		convey.Convey("ValidNPUJobFn() should return error for job with invalid npu request num", func() {
			pod := buildNPUPod(VPodInfo{namespace: "default", groupName: "npu-group-6",
				podName: "npu-test-9", nodeName: nodeName, reqCPUNum: "20", reqMem: "5Gi",
				reqNPUType: npuV910CardName16c, reqNpuNum: "1"})
			setPodSelector(pod, util2.ArchSelector, util2.HuaweiArchX86)
			tasks = append(tasks, api.NewTaskInfo(pod))
			job := api.NewJobInfo(uid, tasks...)
			test.AddTestJobPodGroup(job)
			setJobResourceReq(job, npuV910CardName16c, float64(invalidNum))
			result := vnpu.ValidNPUJobFn(job)
			convey.So(result, convey.ShouldNotBeNil)
		})
	})
}

// TestVnpuValidNPUJobFnSuccess
func TestVnpuValidNPUJobFnSuccess(t *testing.T) {
	convey.Convey("Test job ValidNPUJobFn Success", t, func() {
		vnpu := &VNPU{}
		vnpu310P := &modulev310p.ChipV310P{}
		if getErr := vnpu310P.InitVNPUPlugin(); getErr != nil {
			return
		}
		vnpu.Attr = vnpu310P.ComVNPU
		var tasks []*api.TaskInfo
		uid := api.JobID("xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx4")

		convey.Convey("ValidNPUJobFn() should return nil for job with valid selectors", func() {
			tasks = append(tasks, api.NewTaskInfo(buildNPUPod(
				VPodInfo{namespace: "default", groupName: "npu-group-7",
					podName: "npu-test-3", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
					reqNPUType: npuV310PCardName2c, reqNpuNum: "1"})))
			job := api.NewJobInfo(uid, tasks...)
			setJobResourceReq(job, npuV310PCardName2c, float64(1))
			test.AddTestJobPodGroup(job)
			result := vnpu.ValidNPUJobFn(job)
			convey.So(result, convey.ShouldBeNil)
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
			setNodeLabel(node.Node, util2.ArchSelector, "")
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

// TestVnpuCheckNPUResourceStableFn
func TestVnpuCheckNPUResourceStableFn(t *testing.T) {
	convey.Convey("Test job CheckNPUResourceStableFn", t, func() {
		vnpu := &VNPU{}
		vnpu910 := &modulev910.ChipV910{}
		if getErr := vnpu910.InitVNPUPlugin(); getErr != nil {
			return
		}
		vnpu.Attr = vnpu910.ComVNPU
		convey.Convey("CheckNPUResourceStableFn() should return error when there's missing resource type in idle", func() {
			// build a node with 2 Ascend-910-16c in annotation and no idle Ascend-910-16c
			node := buildNPUNode(VNodeInfo{nodeName: nodeName, nodeArch: util2.HuaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "3", npuTop: "Ascend910-16c-142-1,Ascend910-16c-143-1"})
			result := vnpu.CheckNPUResourceStableFn(node)
			convey.So(result, convey.ShouldBeError)
		})
		convey.Convey("CheckNPUResourceStableFn() should return error when node resources are unstable", func() {
			// build a node with 2 Ascend-910-16c in annotation and 1 idle Ascend-910-16c
			node := buildNPUNode(VNodeInfo{nodeName: nodeName, nodeArch: util2.HuaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "1", npuTop: "Ascend910-16c-144-1,Ascend910-16c-145-1"})
			result := vnpu.CheckNPUResourceStableFn(node)
			convey.So(result, convey.ShouldBeError)
		})
		convey.Convey("CheckNPUResourceStableFn() should return nil when node resources are stable", func() {
			// build a node with 2 Ascend-910-16c in annotation and 1 idle Ascend-910-16c
			node := buildNPUNode(VNodeInfo{nodeName: nodeName, nodeArch: util2.HuaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "2", npuTop: "Ascend910-16c-146-1,Ascend910-16c-147-1"})
			result := vnpu.CheckNPUResourceStableFn(node)
			convey.So(result, convey.ShouldBeNil)
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

// TestVnpuGetNPUAffinityBestNodesFn
func TestVnpuGetNPUAffinityBestNodesFn(t *testing.T) {
	convey.Convey("Test GetNPUAffinityBestNodesFn", t, func() {
		const (
			nodeNameCase1 = "node1"
			nodeNameCase2 = "node2"
			nodeNameCase3 = "node3"
		)
		bestNodesMap := map[string]int{
			nodeNameCase1: 0,
			nodeNameCase2: 0,
			nodeNameCase3: 0,
		}

		vnpu := &VNPU{}
		pod := buildNPUPod(VPodInfo{namespace: "default", groupName: "npu-group-14",
			podName: "npu-test-16", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
			reqNPUType: npuV910CardName16c, reqNpuNum: "1"})
		task := api.NewTaskInfo(pod)
		node1 := buildNPUNode(VNodeInfo{nodeName: nodeNameCase1, nodeArch: util2.HuaweiArchX86, cpu: "192", mem: "755Gi",
			npuAllocateNum: "3", npuTop: "Ascend910-16c-110-1,Ascend910-16c-111-1,Ascend910-16c-112-1"})
		node2 := buildNPUNode(VNodeInfo{nodeName: nodeNameCase2, nodeArch: util2.HuaweiArchArm, cpu: "192", mem: "755Gi",
			npuAllocateNum: "3", npuTop: "Ascend910-16c-113-1,Ascend910-16c-114-1,Ascend910-16c-115-1"})
		node3 := buildNPUNode(VNodeInfo{nodeName: nodeNameCase3, nodeArch: util2.HuaweiArchArm, cpu: "192", mem: "755Gi",
			npuAllocateNum: "3", npuTop: "Ascend910-16c-116-1,Ascend910-16c-117-1,Ascend910-16c-118-1"})

		convey.Convey("GetNPUAffinityBestNodesFn() should return correct bestNodesMap", func() {
			nodes := []*api.NodeInfo{node1, node2, node3, nil}
			result, err := vnpu.GetNPUAffinityBestNodesFn(task, nodes, false)
			convey.So(result, convey.ShouldResemble, bestNodesMap)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestVnpuScoreBestNPUNodesFn
func TestVnpuScoreBestNPUNodesFn(t *testing.T) {
	const constNumber3 = 3.0
	convey.Convey("Test ScoreBestNPUNodesFn", t, func() {
		vnpu := &VNPU{}
		const (
			nodeNameCase1 = "node1"
			nodeNameCase2 = "node2"
		)
		bestNodesMap := map[string]int{
			nodeNameCase1: 0,
			nodeNameCase2: 0,
		}
		node1 := buildNPUNode(VNodeInfo{nodeName: nodeNameCase1, nodeArch: util2.HuaweiArchX86, cpu: "192", mem: "755Gi",
			npuAllocateNum: "3", npuTop: "Ascend910-16c-103-1,Ascend910-16c-104-1,Ascend910-16c-105-1"})
		node2 := buildNPUNode(VNodeInfo{nodeName: nodeNameCase2, nodeArch: util2.HuaweiArchArm, cpu: "192", mem: "755Gi",
			npuAllocateNum: "3", npuTop: "Ascend910-16c-106-1,Ascend910-16c-107-1,Ascend910-16c-108-1"})

		convey.Convey("ScoreBestNPUNodesFn() should return correct scoreMap", func() {
			scoreMap := map[string]float64{
				nodeNameCase1: constNumber3,
			}
			expectResult := map[string]float64{
				nodeNameCase1: 0.0,
				nodeNameCase2: 0.0,
			}
			task := api.TaskInfo{}
			nodes := []*api.NodeInfo{node1, node2}
			result, err := vnpu.ScoreBestNPUNodesFn(scoreMap, bestNodesMap, &task, nodes)
			convey.So(result, convey.ShouldResemble, expectResult)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestVnpuGetAllocatedNPUFromTopologyFn
func TestVnpuGetAllocatedNPUFromTopologyFn(t *testing.T) {
	convey.Convey("Test GetAllocatedNPUFromTopologyFn", t, func() {
		vnpu := &VNPU{}
		pod := buildNPUPod(VPodInfo{namespace: "default", groupName: "npu-group-15",
			podName: "npu-test-17", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
			reqNPUType: npuV910CardName16c, reqNpuNum: "1"})
		task := api.NewTaskInfo(pod)

		convey.Convey("GetAllocatedNPUFromTopologyFn() should return name of the allocated devices", func() {
			node := buildNPUNode(VNodeInfo{nodeName: nodeName, nodeArch: util2.HuaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "1", npuTop: "Ascend910-16c-133-1"})
			result, err := vnpu.GetAllocatedNPUFromTopologyFn(task, node, false)
			convey.So(result, convey.ShouldBeNil)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("GetAllocatedNPUFromTopologyFn() should return error when node annotation is empty", func() {
			node := buildNPUNode(VNodeInfo{nodeName: nodeName, nodeArch: util2.HuaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "1", npuTop: ""})
			result, err := vnpu.GetAllocatedNPUFromTopologyFn(task, node, false)
			convey.So(result, convey.ShouldBeNil)
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("GetAllocatedNPUFromTopologyFn() should return error when task request no vNPU", func() {
			podNoReq := buildNPUPod(VPodInfo{namespace: "default", groupName: "npu-group-16",
				podName: "npu-test-18", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
				reqNPUType: npuV910CardName16c, reqNpuNum: "0"})
			taskNoReq := api.NewTaskInfo(podNoReq)
			node := buildNPUNode(VNodeInfo{nodeName: nodeName, nodeArch: util2.HuaweiArchX86, cpu: "192", mem: "755Gi",
				npuAllocateNum: "1", npuTop: ""})
			result, err := vnpu.GetAllocatedNPUFromTopologyFn(taskNoReq, node, false)
			convey.So(result, convey.ShouldBeNil)
			convey.So(err, convey.ShouldBeError)
		})
	})
}

// TestVnpuSetNPUTopologyToPodFn
func TestVnpuSetNPUTopologyToPodFn(t *testing.T) {
	convey.Convey("Test SetNPUTopologyToPodFn", t, func() {
		vnpu := &VNPU{}
		pod := buildNPUPod(VPodInfo{namespace: "default", groupName: "npu-group-17",
			podName: "npu-test-19", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
			reqNPUType: npuV910CardName16c, reqNpuNum: "1"})
		task := api.NewTaskInfo(pod)

		convey.Convey("SetNPUTopologyToPodFn() should return error when top isn't of slice of string", func() {
			var top []int
			result := vnpu.SetNPUTopologyToPodFn(task, top)
			convey.So(result, convey.ShouldBeError)
		})
		convey.Convey("SetNPUTopologyToPodFn() should write top into pod annotation", func() {
			var top []int
			result := vnpu.SetNPUTopologyToPodFn(task, top)
			convey.So(result, convey.ShouldNotBeNil)
			topFromPod := task.Pod.Annotations[npuV910CardName2c]
			convey.So(topFromPod, convey.ShouldEqual, "")
		})
	})
}

// TestVnpuUpdateNPUNodeUsedCardFn
func TestVnpuUpdateNPUNodeUsedCardFn(t *testing.T) {
	convey.Convey("Test UpdateNPUNodeUsedCardFn", t, func() {
		vnpu := &VNPU{}
		node := buildNPUNode(VNodeInfo{nodeName: nodeName, nodeArch: util2.HuaweiArchX86, cpu: "192", mem: "755Gi",
			npuAllocateNum: "3", npuTop: "Ascend910-16c-109-1,Ascend910-16c-110-1,Ascend910-16c-111-1"})

		convey.Convey("UpdateNPUNodeUsedCardFn() should return error when top isn't of slice of string", func() {
			var top []int
			result := vnpu.UpdateNPUNodeUsedCardFn(node, top)
			convey.So(result, convey.ShouldBeError)
		})
		convey.Convey("UpdateNPUNodeUsedCardFn() should return error when no certain vnpu type in node.Others", func() {
			top := []string{"Ascend910-2c-113-0"}
			result := vnpu.UpdateNPUNodeUsedCardFn(node, top)
			convey.So(result, convey.ShouldBeError)
		})
		convey.Convey("UpdateNPUNodeUsedCardFn() should return error when node.Others has invalid content", func() {
			node.Others = map[string]interface{}{npuV910CardName16c: 0}
			top := []string{"Ascend910-16c-112-0"}
			result := vnpu.UpdateNPUNodeUsedCardFn(node, top)
			convey.So(result, convey.ShouldBeError)
		})
		convey.Convey("UpdateNPUNodeUsedCardFn() should successfully update node.Others", func() {
			node.Others = map[string]interface{}{
				npuV910CardName16c: "Ascend910-16c-111-0,Ascend910-16c-112-0",
			}
			top := []string{"Ascend910-16c-111-0"}
			result := vnpu.UpdateNPUNodeUsedCardFn(node, top)
			convey.So(result, convey.ShouldNotBeNil)
			expectResult := "Ascend910-16c-111-0,Ascend910-16c-112-0"
			convey.So(node.Others[npuV910CardName16c], convey.ShouldResemble, expectResult)
		})
	})
}

// TestVnpuGetReleaseNPUTopologyFn
func TestVnpuGetReleaseNPUTopologyFn(t *testing.T) {
	convey.Convey("Test GetReleaseNPUTopologyFn", t, func() {
		vnpu := &VNPU{}
		var expectResult []string

		pod := buildNPUPod(VPodInfo{namespace: "default", groupName: "npu-group-18",
			podName: "npu-test-20", nodeName: nodeName, reqCPUNum: "10", reqMem: "10Gi",
			reqNPUType: npuV910CardName16c, reqNpuNum: "1"})
		task := api.NewTaskInfo(pod)
		convey.Convey("GetReleaseNPUTopologyFn() should return error when pod has no vnpu annotation", func() {
			taskTopArr, err := vnpu.GetReleaseNPUTopologyFn(task)
			convey.So(err, convey.ShouldBeError)
			convey.So(taskTopArr, convey.ShouldResemble, expectResult)
		})
		convey.Convey("GetReleaseNPUTopologyFn() should return error when pod has empty annotation", func() {
			task.Pod.Annotations[npuV910CardName16c] = ""
			taskTopArr, err := vnpu.GetReleaseNPUTopologyFn(task)
			convey.So(err, convey.ShouldBeError)
			convey.So(taskTopArr, convey.ShouldResemble, expectResult)
		})
		convey.Convey("GetReleaseNPUTopologyFn() should return error when pod has mismatch annotation", func() {
			topStr := "Ascend910-2c-190-1"
			task.Pod.Annotations[npuV910CardName16c] = topStr
			taskTopArr, err := vnpu.GetReleaseNPUTopologyFn(task)
			convey.So(err, convey.ShouldBeError)
			convey.So(taskTopArr, convey.ShouldResemble, expectResult)
		})
		convey.Convey("GetReleaseNPUTopologyFn() should return nil when pod has correct annotation", func() {
			topStr := "Ascend910-16c-180-1"
			task.Pod.Annotations[npuV910CardName16c] = topStr
			expectResult = append(expectResult, topStr)
			taskTopArr, err := vnpu.GetReleaseNPUTopologyFn(task)
			convey.So(err, convey.ShouldBeNil)
			convey.So(taskTopArr, convey.ShouldResemble, expectResult)
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
			setNodeAnnotation(node3.Node, npuV910CardName16c, "")
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

func setPodSelector(pod *v1.Pod, selectorKey string, selectorValue string) {
	pod.Spec.NodeSelector[selectorKey] = selectorValue
}

func setNodeLabel(node *v1.Node, labelKey string, labelValue string) {
	if labelValue == "" {
		delete(node.Labels, labelKey)
		return
	}
	node.Labels[labelKey] = labelValue
}

func setNodeAnnotation(node *v1.Node, annKey string, annValue string) {
	if annValue == "" {
		delete(node.Annotations, annKey)
		return
	}
	node.Annotations[annKey] = annValue
}

func setJobResourceReq(job *api.JobInfo, res string, num float64) {
	if len(job.TotalRequest.ScalarResources) == 0 {
		job.TotalRequest.ScalarResources = make(map[v1.ResourceName]float64, constIntNum2)
	}
	job.TotalRequest.ScalarResources[v1.ResourceName(res)] = num
	var minRes = make(v1.ResourceList, constIntNum3)
	for k, v := range job.TotalRequest.ScalarResources {
		minRes[k] = resource.MustParse(fmt.Sprintf("%f", v))
	}
	if job.PodGroup == nil {
		test.AddTestJobPodGroup(job)
	}
	job.PodGroup.Spec.MinResources = &minRes
}

func buildNPUPod(podInfo VPodInfo) *v1.Pod {

	pod := util.BuildPod(podInfo.namespace, podInfo.podName, podInfo.nodeName, v1.PodPending,
		buildNPUResourceList(podInfo.reqCPUNum, podInfo.reqMem,
			v1.ResourceName(podInfo.reqNPUType), podInfo.reqNpuNum),
		podInfo.groupName, make(map[string]string, constIntNum2),
		make(map[string]string, constIntNum2))

	setPodSelector(pod, util2.ArchSelector, util2.HuaweiArchX86)

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
	nodeCapacity := buildNPUResourceList(nodeInfo.cpu, nodeInfo.mem, npuV910CardName16c, strconv.Itoa(constIntNum2))
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

	setNodeLabel(v1node, util2.ArchSelector, nodeInfo.nodeArch)

	node := api.NewNodeInfo(v1node)
	if nodeInfo.npuAllocateNum != "0" {
		node.Others = map[string]interface{}{
			npuV910CardName16c: nodeInfo.npuTop,
		}
	}
	return node
}
