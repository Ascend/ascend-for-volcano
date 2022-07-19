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
