/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package common is using for HuaWei infer common Ascend pin affinity schedule.

*/
package common

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/vnpu/vnpuutil"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
)

type updateReleaseNPUNodeTopologyFnArgs struct {
	node *api.NodeInfo
	top  interface{}
}

type updateReleaseNPUNodeTopologyFnTests struct {
	name    string
	fields  Scheduler
	args    updateReleaseNPUNodeTopologyFnArgs
	wantErr error
}

func buildUpdateReleaseNPUNodeTopologyFnTestCases() []updateReleaseNPUNodeTopologyFnTests {
	nodeOne := test.FakeNormalTestNode("node-1")
	test.SetTestNPUNodeAnnotation(nodeOne, vnpuutil.NPU310PCardName, "huawei/Ascend310P-2")
	nodeTwo := test.FakeNormalTestNode("node-2")
	test.SetTestNPUNodeAnnotation(nodeTwo, vnpuutil.NPU310PCardName, "Ascend310P-2")
	scheduler := Scheduler{
		PluginName: vnpuutil.PluginName, AnnoName: vnpuutil.NPU310PCardName, AnnoPreVal: "Ascend310P-",
		DefaultJobSchedulerConfig: nil,
	}
	testCases := []updateReleaseNPUNodeTopologyFnTests{
		{
			name: "01-UpdateReleaseNPUNodeTopologyFn() invalid argument test .",
			args: updateReleaseNPUNodeTopologyFnArgs{
				node: nodeOne, top: nil},
			fields:  scheduler,
			wantErr: errors.New(ArgumentError),
		},
		{
			name: "02-UpdateReleaseNPUNodeTopologyFn() node nil npu test",
			args: updateReleaseNPUNodeTopologyFnArgs{
				node: nodeOne, top: []int{1, util.NPUIndex2}},
			fields:  scheduler,
			wantErr: fmt.Errorf("%s has nil npu", nodeOne.Name),
		},
		{
			name: "03-UpdateReleaseNPUNodeTopologyFn() success test.",
			args: updateReleaseNPUNodeTopologyFnArgs{
				node: nodeTwo, top: []int{1, util.NPUIndex2}},
			fields:  scheduler,
			wantErr: nil,
		},
	}
	return testCases
}

// TestUpdateReleaseNPUNodeTopologyFn test UpdateReleaseNPUNodeTopologyFn
func TestUpdateReleaseNPUNodeTopologyFn(t *testing.T) {
	tests := buildUpdateReleaseNPUNodeTopologyFnTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cn := &Scheduler{
				PluginName:                tt.fields.PluginName,
				AnnoName:                  tt.fields.AnnoName,
				AnnoPreVal:                tt.fields.AnnoPreVal,
				DefaultJobSchedulerConfig: tt.fields.DefaultJobSchedulerConfig,
			}
			err := cn.UpdateReleaseNPUNodeTopologyFn(tt.args.node, tt.args.top)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("UpdateReleaseNPUNodeTopologyFn() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type updateNPUNodeUsedCardFnArgs struct {
	node *api.NodeInfo
	top  interface{}
}

type updateNPUNodeUsedCardFnTests struct {
	name    string
	fields  Scheduler
	args    updateNPUNodeUsedCardFnArgs
	wantErr error
}

func buildUpdateNPUNodeUsedCardFnTestCases() []updateNPUNodeUsedCardFnTests {
	nodeOne := test.FakeNormalTestNode("node-1")
	test.SetTestNPUNodeAnnotation(nodeOne, vnpuutil.NPU310PCardName, "huawei/Ascend310P-2")
	nodeTwo := test.FakeNormalTestNode("node-2")
	test.SetTestNPUNodeAnnotation(nodeTwo, vnpuutil.NPU310PCardName, "Ascend310P-2")
	scheduler := Scheduler{
		PluginName: vnpuutil.PluginName, AnnoName: vnpuutil.NPU310PCardName, AnnoPreVal: "Ascend310P-",
		DefaultJobSchedulerConfig: nil,
	}
	testCases := []updateNPUNodeUsedCardFnTests{
		{
			name: "01-UpdateNPUNodeUsedCardFn() invalid argument test .",
			args: updateNPUNodeUsedCardFnArgs{
				node: nodeOne, top: nil},
			fields:  scheduler,
			wantErr: errors.New(ArgumentError),
		},
		{
			name: "02-UpdateNPUNodeUsedCardFn() nodeDeviceIDs nil test",
			args: updateNPUNodeUsedCardFnArgs{
				node: nodeOne, top: []int{1, util.NPUIndex2}},
			fields:  scheduler,
			wantErr: errors.New("nodeDeviceIDs nil"),
		},
		{
			name: "03-UpdateNPUNodeUsedCardFn() success test.",
			args: updateNPUNodeUsedCardFnArgs{
				node: nodeTwo, top: []int{1, util.NPUIndex2}},
			fields:  scheduler,
			wantErr: nil,
		},
	}
	return testCases
}

// TestUpdateNPUNodeUsedCardFn test UpdateNPUNodeUsedCardFn
func TestUpdateNPUNodeUsedCardFn(t *testing.T) {
	tests := buildUpdateNPUNodeUsedCardFnTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cn := &Scheduler{
				PluginName:                tt.fields.PluginName,
				AnnoName:                  tt.fields.AnnoName,
				AnnoPreVal:                tt.fields.AnnoPreVal,
				DefaultJobSchedulerConfig: tt.fields.DefaultJobSchedulerConfig,
			}
			err := cn.UpdateNPUNodeUsedCardFn(tt.args.node, tt.args.top)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("UpdateNPUNodeUsedCardFn() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
