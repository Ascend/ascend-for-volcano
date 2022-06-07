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

func TestVNPU_GetPluginNameByTaskInfo(t *testing.T) {
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

type getReleaseNPUTopologyFnArgs struct {
	vTask *api.TaskInfo
}

type getReleaseNPUTopologyFnTests struct {
	name    string
	fields  vnpuPlugin
	args    getReleaseNPUTopologyFnArgs
	want    interface{}
	wantErr error
}

func buildGetReleaseNPUTopologyFnTestCases() []getReleaseNPUTopologyFnTests {
	var expectResult []string
	tasks := test.FakeNormalTestTasks(util.NPUIndex4)
	test.AddFakeTaskResReq(tasks[1], npuV910CardName16c, 1)
	test.SetTestNPUPodAnnotation(tasks[1].Pod, npuV910CardName16c, "")
	test.AddFakeTaskResReq(tasks[util.NPUIndex2], npuV910CardName16c, 1)
	test.SetTestNPUPodAnnotation(tasks[util.NPUIndex2].Pod, npuV910CardName16c, "Ascend910-2c-190-1")
	test.AddFakeTaskResReq(tasks[util.NPUIndex3], npuV910CardName16c, 1)
	test.SetTestNPUPodAnnotation(tasks[util.NPUIndex3].Pod, npuV910CardName16c, "Ascend910-16c-190-1")
	testCases := []getReleaseNPUTopologyFnTests{
		{
			name: "01-GetReleaseNPUTopologyFn() should return error when pod has no vnpu annotation",
			args: getReleaseNPUTopologyFnArgs{
				vTask: tasks[0]},
			fields:  vnpuPlugin{},
			want:    nil,
			wantErr: fmt.Errorf("%s nil NPU", tasks[0].Name),
		},
		{
			name: "02-GetReleaseNPUTopologyFn() should return error when pod has empty annotation",
			args: getReleaseNPUTopologyFnArgs{
				vTask: tasks[1]},
			want:    expectResult,
			wantErr: errors.New("task pod annotation is empty"),
		},
		{
			name: "03-GetReleaseNPUTopologyFn() should return error when pod has mismatch annotation",
			args: getReleaseNPUTopologyFnArgs{
				vTask: tasks[util.NPUIndex2]},
			want:    expectResult,
			wantErr: errors.New("task pod annotation is empty"),
		},
		{
			name: "04-GetReleaseNPUTopologyFn() should return nil when pod has correct annotation",
			args: getReleaseNPUTopologyFnArgs{
				vTask: tasks[util.NPUIndex3]},
			want:    []string{"Ascend910-16c-190-1"},
			wantErr: nil,
		},
	}
	return testCases
}

func TestVNPU_GetReleaseNPUTopologyFn(t *testing.T) {
	tests := buildGetReleaseNPUTopologyFnTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp := &VNPU{
				Plugin:               tt.fields.Plugin,
				Attr:                 tt.fields.Attr,
				HwNPUSchedulerPlugin: tt.fields.HwNPUSchedulerPlugin,
			}
			got, err := tp.GetReleaseNPUTopologyFn(tt.args.vTask)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("GetReleaseNPUTopologyFn() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetReleaseNPUTopologyFn() got = %v, want %v", got, tt.want)
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

func TestVNPU_GetVTaskReqNPUType(t *testing.T) {
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

type isMyTaskArgs struct {
	vTask *api.TaskInfo
}

type isMyTaskTests struct {
	name    string
	fields  vnpuPlugin
	args    isMyTaskArgs
	wantErr error
}

func buildIsMyTaskTestCases() []isMyTaskTests {
	tasks := test.FakeNormalTestTasks(util.NPUIndex3)
	test.AddFakeTaskResReq(tasks[1], "huawei.com/Ascend910-16c-2", 1)
	test.AddFakeTaskResReq(tasks[util.NPUIndex2], npuV910CardName16c, 1)
	testCases := []isMyTaskTests{
		{
			name: "01-IsMyTask() nil task test.",
			args: isMyTaskArgs{
				vTask: nil},
			fields:  vnpuPlugin{},
			wantErr: errors.New("nil task"),
		},
		{
			name: "02-IsMyTask() task no npu test.",
			args: isMyTaskArgs{
				vTask: tasks[0]},
			wantErr: fmt.Errorf("%s nil NPU", tasks[0].Name),
		},
		{
			name: "03-IsMyTask() has no request test.",
			args: isMyTaskArgs{
				vTask: tasks[util.NPUIndex1]},
			wantErr: errors.New("err resource"),
		},
		{
			name: "04-IsMyTask() success test.",
			args: isMyTaskArgs{
				vTask: tasks[util.NPUIndex2]},
			wantErr: nil,
		},
	}
	return testCases
}

func TestVNPU_IsMyTask(t *testing.T) {
	tests := buildIsMyTaskTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp := &VNPU{
				Plugin:               tt.fields.Plugin,
				Attr:                 tt.fields.Attr,
				HwNPUSchedulerPlugin: tt.fields.HwNPUSchedulerPlugin,
			}
			err := tp.IsMyTask(tt.args.vTask)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("IsMyTask() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type isSelectorMeetNodeArgs struct {
	task *api.TaskInfo
	node *api.NodeInfo
	conf map[string]string
}

type isSelectorMeetNodeTests struct {
	name    string
	fields  vnpuPlugin
	args    isSelectorMeetNodeArgs
	wantErr error
}

func buildIsSelectorMeetNodeTestCases() []isSelectorMeetNodeTests {
	tasks := test.FakeNormalTestTasks(util.NPUIndex2)
	test.AddFakeTaskResReq(tasks[1], npuV910CardName16c, 1)
	nodes := test.FakeNormalTestNodes(util.NPUIndex2)
	test.SetNPUNodeLabel(nodes[0].Node, util.ArchSelector, util.HuaweiArchArm)
	test.SetTestNPUPodSelector(tasks[0].Pod, util.ArchSelector, util.HuaweiArchX86)
	test.SetNPUNodeLabel(nodes[1].Node, util.ArchSelector, util.HuaweiArchArm)
	test.SetTestNPUPodSelector(tasks[1].Pod, util.ArchSelector, util.HuaweiArchArm)
	cons := map[string]string{util.ArchSelector: util.HuaweiArchArm}
	testCases := []isSelectorMeetNodeTests{
		{
			name: "01-IsSelectorMeetNode() nil parameters test.",
			args: isSelectorMeetNodeArgs{
				task: nil, node: nil, conf: nil},
			fields:  vnpuPlugin{},
			wantErr: errors.New("nil parameters"),
		},
		{
			name: "02-IsSelectorMeetNode() no match selector test.",
			args: isSelectorMeetNodeArgs{
				task: tasks[0], node: nodes[0], conf: cons},
			wantErr: fmt.Errorf("no matching label on this node key[%s] : task(%s) node(%s) conf(%s)",
				util.ArchSelector, util.HuaweiArchX86, util.HuaweiArchArm, util.HuaweiArchArm),
		},
		{
			name: "03-IsSelectorMeetNode() match selector test",
			args: isSelectorMeetNodeArgs{
				task: tasks[1], node: nodes[1], conf: cons},
			wantErr: nil,
		},
	}
	return testCases
}

func TestVNPU_IsSelectorMeetNode(t *testing.T) {
	tests := buildIsSelectorMeetNodeTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp := &VNPU{
				Plugin:               tt.fields.Plugin,
				Attr:                 tt.fields.Attr,
				HwNPUSchedulerPlugin: tt.fields.HwNPUSchedulerPlugin,
			}
			err := tp.IsSelectorMeetNode(tt.args.task, tt.args.node, tt.args.conf)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("IsSelectorMeetNode() error = %v, wantErr %v", err, tt.wantErr)
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

func TestVNPU_SetNPUTopologyToPodFn(t *testing.T) {
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
