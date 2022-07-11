/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package comvnpu is using for virtual HuaWei Ascend910 schedule.

*/
package comvnpu

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/vnpu/vnpuutil"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
)

type vnpuPlugin struct {
	Plugin               VNPUHandler
	Attr                 vnpuutil.ComVNPU
	HwNPUSchedulerPlugin plugin.HwNPUSchedulerPlugin
}

type getPluginNameByTaskInfoArgs struct {
	vTask *api.TaskInfo
}

type getPluginNameByTaskInfoTests struct {
	name    string
	fields  vnpuPlugin
	args    getPluginNameByTaskInfoArgs
	want    string
	wantErr error
}

func buildGetPluginNameByTaskInfoTestCases() []getPluginNameByTaskInfoTests {
	tasks := test.FakeNormalTestTasks(util.NPUIndex3)
	test.AddFakeTaskResReq(tasks[1], npuV910CardName2c, 1)
	test.AddFakeTaskResReq(tasks[util.NPUIndex2], npuV310PCardName2c, 1)
	testCases := []getPluginNameByTaskInfoTests{
		{
			name: "01-GetPluginNameByTaskInfoTests nil NPU",
			args: getPluginNameByTaskInfoArgs{
				vTask: tasks[0]},
			want:    "",
			wantErr: fmt.Errorf("%s nil NPU", tasks[0].Name),
		},
		{
			name: "02-GetPluginNameByTaskInfoTests 910-VNPU",
			args: getPluginNameByTaskInfoArgs{
				vTask: tasks[1]},
			want:    "910-VNPU",
			wantErr: nil,
		},
		{
			name: "03-GetPluginNameByTaskInfoTests 710-VNPU",
			args: getPluginNameByTaskInfoArgs{
				vTask: tasks[util.NPUIndex2]},
			want:    "310P-VNPU",
			wantErr: nil,
		},
	}
	return testCases
}

// TestGetPluginNameByTaskInfo test GetPluginNameByTaskInfo
func TestGetPluginNameByTaskInfo(t *testing.T) {
	tests := buildGetPluginNameByTaskInfoTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp := &VNPU{
				Plugin:               tt.fields.Plugin,
				Attr:                 tt.fields.Attr,
				HwNPUSchedulerPlugin: tt.fields.HwNPUSchedulerPlugin,
			}
			got, err := tp.GetPluginNameByTaskInfo(tt.args.vTask)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("GetPluginNameByTaskInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPluginNameByTaskInfo() got = %v, want %v", got, tt.want)
			}
		})
	}
}

type getVTaskReqNPUTypeArgs struct {
	vTask *api.TaskInfo
}

type getVTaskReqNPUTypeTests struct {
	name    string
	fields  vnpuPlugin
	args    getVTaskReqNPUTypeArgs
	want    string
	wantErr error
}

func buildGetVTaskReqNPUTypeTestCases() []getVTaskReqNPUTypeTests {
	tasks := test.FakeNormalTestTasks(util.NPUIndex3)
	test.AddFakeTaskResReq(tasks[1], "huawei.com/Ascend910-16c-2", 1)
	test.AddFakeTaskResReq(tasks[util.NPUIndex2], npuV910CardName16c, 1)
	testCases := []getVTaskReqNPUTypeTests{
		{
			name: "01-GetVTaskReqNPUType() nil task test.",
			args: getVTaskReqNPUTypeArgs{
				vTask: nil},
			fields:  vnpuPlugin{},
			want:    "",
			wantErr: errors.New("nil parameter"),
		},
		{
			name: "02-GetVTaskReqNPUType() task no npu test.",
			args: getVTaskReqNPUTypeArgs{
				vTask: tasks[0]},
			want:    "",
			wantErr: fmt.Errorf("%s nil NPU", tasks[0].Name),
		},
		{
			name: "03-GetVTaskReqNPUType() has wrong request.",
			args: getVTaskReqNPUTypeArgs{
				vTask: tasks[util.NPUIndex1]},
			want:    "",
			wantErr: errors.New("err resource"),
		},
		{
			name: "04-GetVTaskReqNPUType() success test.",
			args: getVTaskReqNPUTypeArgs{
				vTask: tasks[util.NPUIndex2]},
			want:    "huawei.com/Ascend910",
			wantErr: nil,
		},
	}
	return testCases
}

// TestGetVTaskReqNPUType test GetVTaskReqNPUType
func TestGetVTaskReqNPUType(t *testing.T) {
	tests := buildGetVTaskReqNPUTypeTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp := &VNPU{
				Plugin:               tt.fields.Plugin,
				Attr:                 tt.fields.Attr,
				HwNPUSchedulerPlugin: tt.fields.HwNPUSchedulerPlugin,
			}
			got, err := tp.GetVTaskReqNPUType(tt.args.vTask)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("GetVTaskReqNPUType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetVTaskReqNPUType() got = %v, want %v", got, tt.want)
			}
		})
	}
}

type setNPUTopologyToPodArgs struct {
	task *api.TaskInfo
	top  interface{}
}

type setNPUTopologyToPodTests struct {
	name    string
	fields  vnpuPlugin
	args    setNPUTopologyToPodArgs
	wantErr error
}

func buildSetNPUTopologyToPodTestCases() []setNPUTopologyToPodTests {
	var testTopInt []string
	testTopStr := "Ascend910-16c-123-1"
	tasks := test.FakeNormalTestTasks(util.NPUIndex2)
	test.AddFakeTaskResReq(tasks[1], npuV910CardName16c, 1)
	testCases := []setNPUTopologyToPodTests{
		{
			name: "01-SetNPUTopologyToPodFn() nil parameters test.",
			args: setNPUTopologyToPodArgs{
				task: nil, top: testTopInt},
			wantErr: errors.New("nil parameters"),
		},
		{
			name: "02-SetNPUTopologyToPodFn() no match assert test.",
			args: setNPUTopologyToPodArgs{
				task: tasks[0], top: testTopInt},
			wantErr: fmt.Errorf("set NPU topology to pod gets invalid argument"),
		},
		{
			name: "03-SetNPUTopologyToPodFn() no attr test",
			args: setNPUTopologyToPodArgs{
				task: tasks[0], top: testTopStr},
			fields:  vnpuPlugin{},
			wantErr: fmt.Errorf("%s no requets %s", tasks[0].Name, testTopStr),
		},
		{
			name: "04-SetNPUTopologyToPodFn() set success test",
			args: setNPUTopologyToPodArgs{
				task: tasks[1], top: testTopStr},
			fields: vnpuPlugin{
				Attr: vnpuutil.ComVNPU{DivideKinds: []string{npuV910CardName16c},
					HwEntity: plugin.HwEntity{AnnoPreVal: vnpuutil.NPUCardNamePrefix}},
			},
			wantErr: nil,
		},
	}
	return testCases
}

// TestSetNPUTopologyToPodFn test SetNPUTopologyToPodFn
func TestSetNPUTopologyToPodFn(t *testing.T) {
	tests := buildSetNPUTopologyToPodTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp := &VNPU{
				Plugin:               tt.fields.Plugin,
				Attr:                 tt.fields.Attr,
				HwNPUSchedulerPlugin: tt.fields.HwNPUSchedulerPlugin,
			}
			err := tp.SetNPUTopologyToPodFn(tt.args.task, tt.args.top)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("SetNPUTopologyToPodFn() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
