/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package chip310p is using for HuaWei 710 Ascend pin affinity schedule.

*/
package chip310p

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/util"

	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/conf"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/common"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
)

const (
	invalidSelectorKey   = "invalid-key"
	invalidSelectorValue = "no-selector"
	nodeName             = "node0"
	jobName              = "job-01"
)

// TestCNPUName
func TestCNPUName(t *testing.T) {
	t.Run(PluginName, func(t *testing.T) {
		npu := New(PluginName)
		if npu.Name() != PluginName {
			t.Errorf("New(): new npu plugin error ")
		}
	})
}

type isMyTaskTestCase struct {
	name     string
	taskInfo *api.TaskInfo
	wantErr  error
}

func buildIsMyTaskListTestCases() []isMyTaskTestCase {
	return []isMyTaskTestCase{
		{
			name:     "01-IsMyJob() should return error when job request no npu",
			taskInfo: api.NewTaskInfo(test.BuildPodWithReqResource(a310PNPUChipName, "0")),
			wantErr:  errors.New(common.JobNoNPUCard),
		},
		{
			name:     "02-IsMyJob() should return error when job request npu less than 1",
			taskInfo: api.NewTaskInfo(test.BuildPodWithReqResource(a310PNPUChipName, "0.5")),
			wantErr:  errors.New(common.JobNoNPUCard),
		},
		{
			name:     "03-IsMyJob() should return nil when job request npu",
			taskInfo: api.NewTaskInfo(test.BuildPodWithReqResource(a310PNPUChipName, "1")),
			wantErr:  nil,
		},
	}
}

// TestCNPUIsMyTask
func TestCNPUIsMyTask(t *testing.T) {
	npu := New(PluginName)
	testCases := buildIsMyTaskListTestCases()
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if err := npu.IsMyTask(tt.taskInfo); !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("IsMyTask() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type nodeRelatedTestCase struct {
	name     string
	nodeInfo *api.NodeInfo
	wantErr  error
}

func buildIsMyNodeTestCases() []nodeRelatedTestCase {
	nodeInfo1 := test.BuildNodeWithFakeOther(nodeName, a310PNPUChipName, "")
	wantErr1 := errors.New(common.JobNoNPUCard)
	nodeInfo2 := test.BuildNodeWithFakeOther(nodeName, a310PNPUChipName, "Ascend310P-2")
	return []nodeRelatedTestCase{
		{
			name:     "01-IsMyNode() should return error when node has no npu annotation",
			nodeInfo: nodeInfo1,
			wantErr:  wantErr1,
		},
		{
			name:     "02-IsMyNode() should return nil when node has npu annotation",
			nodeInfo: nodeInfo2,
			wantErr:  nil,
		},
	}
}

// TestCNPUIsMyNode
func TestCNPUIsMyNode(t *testing.T) {
	npu := New(PluginName)
	testCases := buildIsMyNodeTestCases()
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if err := npu.IsMyNode(tt.nodeInfo); !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("IsMyTask() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type isMyJobTestCase struct {
	name    string
	jobInfo *api.JobInfo
	wantErr error
}

func buildIsMyJobListTestCases() []isMyJobTestCase {
	jobInfo1 := test.FakeNormalTestJob(jobName, 1)
	test.UpdateFakeJobRequestSource(jobInfo1, a310PNPUChipName, 0)
	jobInfo2 := test.FakeNormalTestJob(jobName, 1)
	test.UpdateFakeJobRequestSource(jobInfo2, a310PNPUChipName, 1)
	return []isMyJobTestCase{
		{
			name:    "01-IsMyJob() should return error when job request no npu",
			jobInfo: jobInfo1,
			wantErr: errors.New(common.JobNoNPUCard),
		},
		{
			name:    "02-IsMyJob() should return nil when job request npu",
			jobInfo: jobInfo2,
			wantErr: nil,
		},
	}
}

// TestCNPUIsMyJob
func TestCNPUIsMyJob(t *testing.T) {
	npu := New(PluginName)
	testCases := buildIsMyJobListTestCases()
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if err := npu.IsMyJob(tt.jobInfo); !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("IsMyNode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type ValidNPUJobFnTestCase struct {
	name    string
	jobInfo *api.JobInfo
	wantRes *api.ValidateResult
}

func buildValidNPUJobFnListTestCases() []ValidNPUJobFnTestCase {
	const (
		noneExistNPUChipName = "noneExistNPUChipName"
		invalidNum0          = 0
		invalidNum65         = 65
		validNum1            = 1
	)
	job1 := test.BuildFakeJobWithSelectorAndSource(jobName, archSelector, "", a310PNPUChipName, validNum1)
	wantRes1 := &api.ValidateResult{
		Pass:    false,
		Reason:  fmt.Sprintf("%s %s getJobSelectors nil", PluginName, job1.Name),
		Message: fmt.Sprintf("validNPUJob err: %s %s getJobSelectors nil", PluginName, job1.Name),
	}
	job2 := test.BuildFakeJobWithSelectorAndSource(jobName, invalidSelectorKey, invalidSelectorValue,
		a310PNPUChipName, validNum1)
	wantRes2 := &api.ValidateResult{
		Pass:    false,
		Reason:  fmt.Sprintf("%s has no selector:%s", job2.Name, archSelector),
		Message: fmt.Sprintf("validNPUJob err: %s has no selector:%s", job2.Name, archSelector),
	}
	job3 := test.BuildFakeJobWithSelectorAndSource(jobName, archSelector, huaweiArchX86, a310PNPUChipName, invalidNum0)
	wantRes3 := &api.ValidateResult{
		Pass:    false,
		Reason:  "job require npu number illegal",
		Message: fmt.Sprintf("%s, err: job no use npu", job3.Name),
	}
	job4 := test.BuildFakeJobWithSelectorAndSource(jobName, archSelector, huaweiArchX86, noneExistNPUChipName, validNum1)
	wantRes4 := &api.ValidateResult{
		Pass:    false,
		Reason:  "job require npu number illegal",
		Message: fmt.Sprintf("%s, err: job no use npu", job4.Name),
	}
	job5 := test.BuildFakeJobWithSelectorAndSource(jobName, archSelector, huaweiArchX86, a310PNPUChipName, invalidNum65)
	wantRes5 := &api.ValidateResult{
		Pass:   false,
		Reason: "job scheduler-strategy error",
		Message: fmt.Sprintf("%s, err: %s single trainning not match task NPU number:%d",
			job5.Name, job5.Name, invalidNum65),
	}
	job6 := test.BuildFakeJobWithSelectorAndSource(jobName, archSelector, huaweiArchX86, a310PNPUChipName, validNum1)
	return []ValidNPUJobFnTestCase{
		{"ValidNPUJobFn() should return error for job without certain selector key", job1, wantRes1},
		{"ValidNPUJobFn() should return error for job for job with invalid selector value", job2, wantRes2},
		{"ValidNPUJobFn() should return error for job with invalid request number 0", job3, wantRes3},
		{"ValidNPUJobFn() should return error for job with invalid request npu", job4, wantRes4},
		{"ValidNPUJobFn() should return error for job with invalid request number 65", job5, wantRes5},
		{"ValidNPUJobFn() should return nil for job with valid selector and request number", job6, nil},
	}
}

// TestCNPUValidNPUJobFn
func TestCNPUValidNPUJobFn(t *testing.T) {
	npu := New(PluginName)
	testCases := buildValidNPUJobFnListTestCases()
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if result := npu.ValidNPUJobFn(tt.jobInfo); !reflect.DeepEqual(result, tt.wantRes) {
				t.Errorf("ValidNPUJobFn() error = %v, want %v", result, tt.wantRes)
			}
		})
	}
}

type preCheckNodeFnTestCase struct {
	name     string
	taskInfo *api.TaskInfo
	nodeInfo *api.NodeInfo
	confs    []conf.Configuration
	wantErr  error
}

func buildPreCheckNodeFnTestCases() []preCheckNodeFnTestCase {
	node1 := test.FakeNormalTestNode(nodeName)
	test.SetNPUNodeLabel(node1.Node, archSelector, huaweiArchX86)
	pod1 := test.BuildPodWithReqResource(a310PNPUChipName, "1")
	test.SetTestNPUPodSelector(pod1, archSelector, huaweiArchX86)
	task1 := api.NewTaskInfo(pod1)

	node2 := test.FakeNormalTestNode(nodeName)
	test.SetNPUNodeLabel(node2.Node, archSelector, huaweiArchX86)
	pod2 := test.BuildPodWithReqResource(a310PNPUChipName, "2")
	test.SetTestNPUPodSelector(pod2, invalidSelectorKey, invalidSelectorValue)
	task2 := api.NewTaskInfo(pod2)
	return []preCheckNodeFnTestCase{
		{
			name:     "01-PreCheckNodeFn() should return nil while selector is match",
			taskInfo: task1,
			nodeInfo: node1,
			confs:    []conf.Configuration{},
			wantErr:  nil,
		},
		{
			name:     "02-PreCheckNodeFn() should return error while conf has no task selector",
			taskInfo: task2,
			nodeInfo: node2,
			confs:    []conf.Configuration{},
			wantErr:  fmt.Errorf("%s : conf has no:%s", util.NodeNoFitSelectorError, invalidSelectorKey),
		},
		{
			name:     "03-PreCheckNodeFn() should return error while selector is mismatch",
			taskInfo: task2,
			nodeInfo: node2,
			confs: []conf.Configuration{
				{
					Name:      util.CMSelectorKey,
					Arguments: map[string]string{invalidSelectorKey: invalidSelectorKey},
				},
			},
			wantErr: fmt.Errorf("%s : node has no:%s", util.NodeNoFitSelectorError, invalidSelectorKey),
		},
	}
}

// TestCnpuPreCheckNodeFnSuccess
func TestCnpuPreCheckNodeFnSuccess(t *testing.T) {
	npu := New(PluginName)
	testCases := buildPreCheckNodeFnTestCases()
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if err := npu.PreCheckNodeFn(tt.taskInfo, tt.nodeInfo, tt.confs); !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("PreCheckNodeFn(), err = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func buildCheckNPUResourceStableTestCases() []nodeRelatedTestCase {
	const npuNum = 2
	nodeInfo1 := test.BuildUnstableNode(nodeName, a310PNPUChipName, "", 0)
	nodeInfo2 := test.BuildUnstableNode(nodeName, a310PNPUChipName, "Ascend310P-2", 0)
	nodeInfo3 := test.BuildUnstableNode(nodeName, a310PNPUChipName, "Ascend310P-2", npuNum)
	nodeInfo4 := test.BuildUnstableNode(nodeName, a310PNPUChipName, "Ascend310P-2,Ascend310P-3", npuNum)
	return []nodeRelatedTestCase{
		{
			name: "01-CheckNPUResourceStableFn() should return error when insufficient npus on the" +
				"schedulable nodes in cluster",
			nodeInfo: nodeInfo1,
			wantErr: fmt.Errorf("getNodeNPUNumFromOthers %s : %s", common.NodesNoMeetNPUReqError,
				fmt.Errorf("nil node(%s) top", nodeInfo1.Name)),
		},
		{
			name:     "02-CheckNPUResourceStableFn() should return error when there's missing resource type in idle",
			nodeInfo: nodeInfo2,
			wantErr: fmt.Errorf("getNodeNPUNumFromIdle %s : %s", common.NodesNoMeetNPUReqError,
				errors.New("get node idle npu failed")),
		},
		{
			name:     "03-CheckNPUResourceStableFn() should return error when resource is mismatch",
			nodeInfo: nodeInfo3,
			wantErr: fmt.Errorf("%s : %s", common.NodeNotStableWarning,
				errors.New("node not stable for annotations(1) : idle(2)")),
		},
		{
			name:     "04-CheckNPUResourceStableFn() should return nil when node resources are stable",
			nodeInfo: nodeInfo4,
			wantErr:  nil,
		},
	}
}

// TestCnpuCheckNPUResourceStableFn
func TestCnpuCheckNPUResourceStableFn(t *testing.T) {
	npu := New(PluginName)
	testCases := buildCheckNPUResourceStableTestCases()
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if result := npu.CheckNPUResourceStableFn(tt.nodeInfo); !reflect.DeepEqual(result, tt.wantErr) {
				t.Errorf("CheckNPUResourceStable() error = %v, want %v", result, tt.wantErr)
			}
		})
	}
}

type checkNodeNPUByTaskFnTestCase struct {
	name     string
	nodeInfo *api.NodeInfo
	taskInfo *api.TaskInfo
	wantErr  error
}

func buildCheckNodeNPUByTaskFnTestCases() []checkNodeNPUByTaskFnTestCase {
	return []checkNodeNPUByTaskFnTestCase{
		{
			name: "01-CheckNodeNPUByTaskFn() should return error when the requirement of the task is 0",
			nodeInfo: test.BuildNodeWithFakeOther(nodeName, a310PNPUChipName, "Ascend310P-3,Ascend310P-7,"+
				"Ascend310P-8"),
			taskInfo: api.NewTaskInfo(test.BuildPodWithReqResource(a310PNPUChipName, "0")),
			wantErr: fmt.Errorf("getTaskNPUNum %s : %s", common.NodesNoMeetNPUReqError,
				fmt.Errorf("not %s task", a310PNPUChipName)),
		},
		{
			name:     "02-CheckNodeNPUByTaskFn() should return error when the number of NPU on a node is 0",
			nodeInfo: test.BuildNodeWithFakeOther(nodeName, a310PNPUChipName, ""),
			taskInfo: api.NewTaskInfo(test.BuildPodWithReqResource(a310PNPUChipName, "5")),
			wantErr:  fmt.Errorf("%s:get npu nil", common.NodeNotEnoughNPUWarning),
		},
		{
			name: "03-CheckNodeNPUByTaskFn() should return error when  when node doesn't meet task requests",
			nodeInfo: test.BuildNodeWithFakeOther(nodeName, a310PNPUChipName, "Ascend310P-6,Ascend310P-7,"+
				"Ascend310P-8"),
			taskInfo: api.NewTaskInfo(test.BuildPodWithReqResource(a310PNPUChipName, "5")),
			wantErr:  errors.New("req npu(5) illegal"),
		},
		{
			name: "04-CheckNodeNPUByTaskFn()  should return nil when node meets task requests",
			nodeInfo: test.BuildNodeWithFakeOther(nodeName, a310PNPUChipName, "Ascend310P-6,Ascend310P-7,"+
				"Ascend310P-8,Ascend310P-9,Ascend310P-10"),
			taskInfo: api.NewTaskInfo(test.BuildPodWithReqResource(a310PNPUChipName, "5")),
			wantErr:  nil,
		},
	}
}

// TestCnpuCheckNodeNPUByTaskFn
func TestCnpuCheckNodeNPUByTaskFn(t *testing.T) {
	npu := New(PluginName)
	testCases := buildCheckNodeNPUByTaskFnTestCases()
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			err := npu.CheckNodeNPUByTaskFn(tt.taskInfo, tt.nodeInfo, true)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("CheckNodeNPUByTaskFn(), err = %v, want = %v", err, tt.wantErr)
			}
		})
	}
}

type updateNPUNodeUsedCardFnTestCase struct {
	name     string
	nodeInfo *api.NodeInfo
	npuTop   interface{}
	wantErr  error
}

func buildUpdateNPUNodeUsedCardFnTestCases() []updateNPUNodeUsedCardFnTestCase {
	nodeInfo1 := test.BuildNodeWithFakeOther(nodeName, a310PNPUChipName, "Ascend310P-0,Ascend310P-1,"+
		"Ascend310P-2,Ascend310P-3,Ascend310P-4,Ascend310P-5,Ascend310P-6")
	npuTop1 := []int{0, util.NPUIndex3, util.NPUIndex4, util.NPUIndex5}

	nodeInfo2 := test.BuildNodeWithFakeOther(nodeName, a310PNPUChipName, "Ascend310P-0,Ascend310P-2,"+
		"Ascend310P-3,Ascend310P-1")
	npuTop2 := []int{0, 1, util.NPUIndex2, util.NPUIndex3}

	nodeInfo3 := test.BuildNodeWithFakeOther(nodeName, a310PNPUChipName, "")
	npuTop3 := []int{0, 1, util.NPUIndex2, util.NPUIndex3}
	wantErr3 := errors.New("nodeDeviceIDs nil")

	nodeInfo4 := test.BuildNodeWithFakeOther(nodeName, a310PNPUChipName, "Ascend310P-0,Ascend310P-2,"+
		"Ascend310P-3,Ascend310P-1")
	npuTop4 := []string{"0", "4"}
	wantErr4 := errors.New(common.ArgumentError)
	return []updateNPUNodeUsedCardFnTestCase{
		{
			name:     "01-UpdateNPUNodeUsedCardFn() should successfully update node.others",
			nodeInfo: nodeInfo1,
			npuTop:   npuTop1,
			wantErr:  nil,
		},
		{
			name:     "02-UpdateNPUNodeUsedCardFn() should successfully update node.others",
			nodeInfo: nodeInfo2,
			npuTop:   npuTop2,
			wantErr:  nil,
		},
		{
			name:     "03-UpdateNPUNodeUsedCardFn() should return error when node's npuTop is empty",
			nodeInfo: nodeInfo3,
			npuTop:   npuTop3,
			wantErr:  wantErr3,
		},
		{
			name:     "04-UpdateNPUNodeUsedCardFn() should return error when top's type mismatch",
			nodeInfo: nodeInfo4,
			npuTop:   npuTop4,
			wantErr:  wantErr4,
		},
	}

}

// TestCnpuUpdateNPUNodeUsedCardFn
func TestCnpuUpdateNPUNodeUsedCardFn(t *testing.T) {
	npu := New(PluginName)
	testCases := buildUpdateNPUNodeUsedCardFnTestCases()
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if result := npu.UpdateNPUNodeUsedCardFn(tt.nodeInfo, tt.npuTop); !reflect.DeepEqual(result, tt.wantErr) {
				t.Errorf("UpdateNPUNodeUsedCardFn(), result = %v, want = %v", result, tt.wantErr)
			}
		})
	}
}

type getReleaseNPUTopologyFnTestCase struct {
	name     string
	taskInfo *api.TaskInfo
	wantRes  interface{}
	wantErr  error
}

func buildGetReleaseNPUTopologyFnTestCases() []getReleaseNPUTopologyFnTestCase {
	task1 := test.BuildTestTaskWithAnnotation(a310PNPUChipName, "2", "Ascend310P-0,Ascend310P-4")
	wantRes1 := []int{0, util.NPUIndex4}

	task2 := test.BuildTestTaskWithAnnotation(a310PNPUChipName, "2", "Ascend310P0,Ascend310P4")
	wantErr2 := fmt.Errorf("%s get npu nil", task2.Name)
	return []getReleaseNPUTopologyFnTestCase{
		{
			name:     "01-GetReleaseNPUTopologyFn() should return correct card id slice",
			taskInfo: task1,
			wantRes:  wantRes1,
			wantErr:  nil,
		},
		{
			name:     "02-GetReleaseNPUTopologyFn() should return err when ",
			taskInfo: task2,
			wantRes:  nil,
			wantErr:  wantErr2,
		},
	}
}

// TestCnpuGetReleaseNPUTopologyFn
func TestCnpuGetReleaseNPUTopologyFn(t *testing.T) {
	npu := New(PluginName)
	testCases := buildGetReleaseNPUTopologyFnTestCases()
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := npu.GetReleaseNPUTopologyFn(tt.taskInfo)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("GetReleaseNPUTopologyFn() err = %v, wantErr = %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(result, tt.wantRes) {
				t.Errorf("GetReleaseNPUTopologyFn() result = %v, want = %v", result, tt.wantRes)
			}
		})

	}
}

type updateReleaseNPUNodeTopologyFnTestCase struct {
	name     string
	nodeInfo *api.NodeInfo
	npuTop   interface{}
	wantErr  error
}

func buildUpdateReleaseNPUNodeTopologyFnTestCases() []updateReleaseNPUNodeTopologyFnTestCase {
	nodeInfo1 := test.BuildNodeWithFakeOther(nodeName, a310PNPUChipName, "")
	npuTop1 := []int{1, util.NPUIndex2}
	wantErr1 := fmt.Errorf("%s has nil npu", nodeInfo1.Name)

	nodeInfo2 := test.BuildNodeWithFakeOther(nodeName, a310PNPUChipName, "Ascend310P-0,Ascend310P-2,"+
		"Ascend310P-61,Ascend310P-1")
	npuTop2 := []int{1, util.NPUIndex2}

	nodeInfo3 := test.BuildNodeWithFakeOther(nodeName, a310PNPUChipName, "Ascend310P-0,Ascend310P-2,"+
		"Ascend310P-61,Ascend310P-1")
	npuTop3 := []string{"0", "3"}
	wantErr3 := errors.New(common.ArgumentError)

	nodeInfo4 := test.BuildNodeWithFakeOther(nodeName, a310PNPUChipName, "Ascend310P-0Ascend310P-2")
	npuTop4 := []int{1, util.NPUIndex2}
	wantErr4 := fmt.Errorf("%s has nil npu", nodeInfo4.Name)

	return []updateReleaseNPUNodeTopologyFnTestCase{
		{
			name:     "01-UpdateNPUNodeUsedCardFn() should return err when node has nil npu",
			nodeInfo: nodeInfo1,
			npuTop:   npuTop1,
			wantErr:  wantErr1,
		},
		{
			name:     "02-UpdateNPUNodeUsedCardFn() should successfully update node.others",
			nodeInfo: nodeInfo2,
			npuTop:   npuTop2,
			wantErr:  nil,
		},
		{
			name:     "03-UpdateNPUNodeUsedCardFn() should return err when top type is mismatch",
			nodeInfo: nodeInfo3,
			npuTop:   npuTop3,
			wantErr:  wantErr3,
		},
		{
			name:     "04-UpdateNPUNodeUsedCardFn() should return error when node's others is wrong",
			nodeInfo: nodeInfo4,
			npuTop:   npuTop4,
			wantErr:  wantErr4,
		},
	}
}

// TestCnpuUpdateReleaseNPUNodeTopologyFn
func TestCnpuUpdateReleaseNPUNodeTopologyFn(t *testing.T) {
	npu := New(PluginName)
	testCases := buildUpdateReleaseNPUNodeTopologyFnTestCases()
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if err := npu.UpdateReleaseNPUNodeTopologyFn(tt.nodeInfo, tt.npuTop); !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("GetReleaseNPUTopologyFn() err = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

type getAllocatedNPUFromTopologyFnTestCase struct {
	name     string
	taskInfo *api.TaskInfo
	nodeInfo *api.NodeInfo
	wantRes  interface{}
	wantErr  error
}

func buildGetAllocatedNPUFromTopologyFnTestCases() []getAllocatedNPUFromTopologyFnTestCase {
	task1 := api.NewTaskInfo(test.BuildPodWithReqResource(a310PNPUChipName, "2"))
	node1 := test.BuildNodeWithFakeOther(nodeName, a310PNPUChipName, "Ascend310P-0,Ascend310P-3")
	wantRes1 := []int{0, util.NPUIndex3}

	task2 := api.NewTaskInfo(test.BuildPodWithReqResource(a310PNPUChipName, "1"))
	node2 := test.BuildNodeWithFakeOther(nodeName, a310PNPUChipName, "Ascend310P-0,Ascend310P-3")
	wantRes2 := []int{0}

	task3 := api.NewTaskInfo(test.BuildPodWithReqResource(a310PNPUChipName, "0"))
	node3 := test.BuildNodeWithFakeOther(nodeName, a310PNPUChipName, "Ascend310P-0,Ascend310P-3")
	var wantRes3 interface{}
	wantErr3 := errors.New("no npu task")

	task4 := api.NewTaskInfo(test.BuildPodWithReqResource(a310PNPUChipName, "5"))
	node4 := test.BuildNodeWithFakeOther(nodeName, a310PNPUChipName, "Ascend310P-0,Ascend310P-3")
	var wantRes4 []int
	topology := []int{0, util.NPUIndex3}
	wantErr4 := fmt.Errorf("%v-%v not meet request number[5]", topology, topology)

	return []getAllocatedNPUFromTopologyFnTestCase{
		{
			name:     "01-GetAllocatedNPUFromTopologyFn() should return correct result when reqNpuNum is match",
			taskInfo: task1,
			nodeInfo: node1,
			wantRes:  wantRes1,
			wantErr:  nil,
		},
		{
			name: "02-GetAllocatedNPUFromTopologyFn() should return correct result when reqNpuNum is less" +
				"than node others",
			taskInfo: task2,
			nodeInfo: node2,
			wantRes:  wantRes2,
			wantErr:  nil,
		},
		{
			name:     "03-GetAllocatedNPUFromTopologyFn() should return error when reqNpuNum of pod is 0",
			taskInfo: task3,
			nodeInfo: node3,
			wantRes:  wantRes3,
			wantErr:  wantErr3,
		},
		{
			name: "04-GetAllocatedNPUFromTopologyFn() should return error when reqNpuNum is " +
				"greater than node others",
			taskInfo: task4,
			nodeInfo: node4,
			wantRes:  wantRes4,
			wantErr:  wantErr4,
		},
	}
}

// TestCnpuGetAllocatedNPUFromTopologyFn
func TestCnpuGetAllocatedNPUFromTopologyFn(t *testing.T) {
	npu := New(PluginName)
	testCases := buildGetAllocatedNPUFromTopologyFnTestCases()
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := npu.GetAllocatedNPUFromTopologyFn(tt.taskInfo, tt.nodeInfo, false)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("GetAllocatedNPUFromTopologyFn() err = %v, wantErr = %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(result, tt.wantRes) {
				t.Errorf("GetAllocatedNPUFromTopologyFn() result = %v, wantRes = %v", result, tt.wantRes)
			}
		})
	}
}

type setNPUTopologyToPodFnTestCase struct {
	name     string
	taskInfo *api.TaskInfo
	npuTop   interface{}
	wantErr  error
}

func buildSetNPUTopologyToPodFnTestCases() []setNPUTopologyToPodFnTestCase {
	task1 := api.NewTaskInfo(test.BuildPodWithReqResource(a310PNPUChipName, "4"))
	npuTop1 := []string{"0", "4"}
	wantErr1 := errors.New(common.ArgumentError)

	task2 := api.NewTaskInfo(test.BuildPodWithReqResource(a310PNPUChipName, "4"))
	npuTop2 := []int{0, util.NPUIndex4}

	return []setNPUTopologyToPodFnTestCase{
		{
			name:     "01-SetNPUTopologyToPodFn() should return error when top is of wrong type",
			taskInfo: task1,
			npuTop:   npuTop1,
			wantErr:  wantErr1,
		},
		{
			name:     "02-SetNPUTopologyToPodFn() should write correct info in pod annotation",
			taskInfo: task2,
			npuTop:   npuTop2,
			wantErr:  nil,
		},
	}
}

// TestCnpuSetNPUTopologyToPodFn
func TestCnpuSetNPUTopologyToPodFn(t *testing.T) {
	npu := New(PluginName)
	testCases := buildSetNPUTopologyToPodFnTestCases()
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if err := npu.SetNPUTopologyToPodFn(tt.taskInfo, tt.npuTop); !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("TestCnpuSetNPUTopologyToPodFn() err = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}
