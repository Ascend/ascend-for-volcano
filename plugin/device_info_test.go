/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package rescheduling is using for HuaWei Ascend pin fault rescheduling.
*/
package plugin

import (
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

const (
	nodeTotalCoreNum256 = 256
	nodeTotalCpuNum112  = 112
	nodeTotalCoreNum128 = 128
	nodeTotalCoreNum56  = 56
	nodeTotalCoreNum64  = 64
	nodeTotalCpuNum49   = 49
	taskCoreNum16       = 16
	taskCpuNum14        = 14
	taskCpuNum28        = 28
)

type getResourceFromStrArgs struct {
	vDeviceResourceStr string
}

type getResourceFromStrTest struct {
	name string
	args getResourceFromStrArgs
	want *util.VResource
}

func buildGetResourceFromStrTests() []getResourceFromStrTest {
	tests := []getResourceFromStrTest{
		{
			name: "01-GetResourceFromRealStr invalid string",
			args: getResourceFromStrArgs{
				vDeviceResourceStr: "44",
			},
			want: nil,
		},
		{
			name: "02-GetResourceFromRealStr only core",
			args: getResourceFromStrArgs{
				vDeviceResourceStr: "vir04",
			},
			want: &util.VResource{
				Aicore: util.NPUIndex4,
				Aicpu:  util.NPUIndex4,
				DVPP:   AscendDVPPEnabledNull,
			},
		},
		{
			name: "03-GetResourceFromRealStr core and cpu",
			args: getResourceFromStrArgs{
				vDeviceResourceStr: "vir04_3c",
			},
			want: &util.VResource{
				Aicore: util.NPUIndex4,
				Aicpu:  util.NPUIndex3,
				DVPP:   AscendDVPPEnabledNull,
			},
		},
		{
			name: "04-GetResourceFromRealStr core, cpu and dvpp",
			args: getResourceFromStrArgs{
				vDeviceResourceStr: "vir04_3c_ndvpp",
			},
			want: &util.VResource{
				Aicore: util.NPUIndex4,
				Aicpu:  util.NPUIndex3,
				DVPP:   AscendDVPPEnabledOff,
			},
		},
	}
	return tests
}

func TestGetResourceFromCoreStr(t *testing.T) {
	tests := buildGetResourceFromStrTests()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetResourceFromCoreStr(tt.args.vDeviceResourceStr); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetResourceFromRealStr() = %v, want %v", got, tt.want)
			}
		})
	}
}

type IsPodWholeCardArgs struct {
	realCardName string
}

type IsPodWholeCardTest struct {
	name string
	args IsPodWholeCardArgs
	want bool
}

func buildIsPodWholeCardTest() []IsPodWholeCardTest {
	tests := []IsPodWholeCardTest{
		{
			name: "01-IsPodWholeCardTest-is whole card",
			args: IsPodWholeCardArgs{
				realCardName: "0,1",
			},
			want: true,
		},
		{
			name: "02-IsPodWholeCardTest-not whold card",
			args: IsPodWholeCardArgs{realCardName: "0-vir04"},
			want: false,
		},
	}
	return tests
}

func TestIsPodWholeCard(t *testing.T) {
	tests := buildIsPodWholeCardTest()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsPodWholeCardFromAscendCore(tt.args.realCardName); got != tt.want {
				t.Errorf("IsPodWholeCard() = %v, want %v", got, tt.want)
			}
		})
	}
}

func FakeVNPUTaskWithResSetting(name, ringController, vnpuLevel, coreNum, dvpp string) *api.TaskInfo {
	task := FakeVNPUTestTask(name, "node0", "", "")
	task.Pod.Labels[util.JobKindKey] = ringController
	task.Pod.Labels[AscendVNPULevel] = vnpuLevel
	task.Pod.Spec.Containers[0].Resources.Requests[util.AscendNPUCore] = resource.MustParse(coreNum)
	task.Pod.Labels[AscendVNPUDVPP] = dvpp
	return task
}

type VNodeTransferTaskLabelToResReqFields struct {
	Chips        map[int]*VChip
	ChipKind     string
	ServerType   string
	TotalChipNum int
	FreeChipNum  int
	TotalRes     util.VResource
}

type VNodeTransferTaskLabelToResReqArgs struct {
	task *api.TaskInfo
}

type VNodeTransferTaskLabelToResReqTests struct {
	name    string
	fields  VNodeTransferTaskLabelToResReqFields
	args    VNodeTransferTaskLabelToResReqArgs
	want    util.VResource
	wantErr bool
}

func buildVNodeTransferTaskLabelToResReqTestCase01() VNodeTransferTaskLabelToResReqTests {
	return VNodeTransferTaskLabelToResReqTests{
		name: "01-VNodeTransferTaskLabelToResReq-310p, coreNum invalid",
		fields: VNodeTransferTaskLabelToResReqFields{
			ServerType: "Ascend310P-8",
			TotalRes:   util.VResource{Aicore: nodeTotalCoreNum56, Aicpu: nodeTotalCpuNum49},
		},
		args: VNodeTransferTaskLabelToResReqArgs{
			task: FakeVNPUTaskWithResSetting("pod0", "ascend-310P", AscendVNPULevelHigh, "16",
				AscendDVPPEnabledOn),
		},
		want:    buildVResource(taskCoreNum16, taskCpuNum14, AscendDVPPEnabledNull),
		wantErr: false,
	}
}

func buildVNodeTransferTaskLabelToResReqTestCase02() VNodeTransferTaskLabelToResReqTests {
	return VNodeTransferTaskLabelToResReqTests{
		name: "02-VNodeTransferTaskLabelToResReq-910 valid",
		fields: VNodeTransferTaskLabelToResReqFields{
			ServerType: "Ascend910-32",
			TotalRes:   util.VResource{Aicore: nodeTotalCoreNum64, Aicpu: taskCpuNum28},
		},
		args: VNodeTransferTaskLabelToResReqArgs{
			task: FakeVNPUTaskWithResSetting("pod1", "ascend-910", AscendVNPULevelLow, "4",
				AscendDVPPEnabledOn),
		},
		want:    buildVResource(util.NPUIndex4, util.NPUIndex1, AscendDVPPEnabledNull),
		wantErr: false,
	}
}

func buildVNodeTransferTaskLabelToResReqTestCase03() VNodeTransferTaskLabelToResReqTests {
	return VNodeTransferTaskLabelToResReqTests{
		name: "03-VNodeTransferTaskLabelToResReq-310p coreNum valid",
		fields: VNodeTransferTaskLabelToResReqFields{
			ServerType: "Ascend310P-8",
			TotalRes:   util.VResource{Aicore: nodeTotalCoreNum64, Aicpu: nodeTotalCoreNum56},
		},
		args: VNodeTransferTaskLabelToResReqArgs{
			task: FakeVNPUTaskWithResSetting("pod2", "ascend-310P", AscendVNPULevelHigh, "4",
				AscendDVPPEnabledOn),
		},
		want:    buildVResource(util.NPUIndex4, util.NPUIndex4, AscendDVPPEnabledOn),
		wantErr: false,
	}
}

func buildVNodeTransferTaskLabelToResReqTestCase04() VNodeTransferTaskLabelToResReqTests {
	return VNodeTransferTaskLabelToResReqTests{
		name: "04-VNodeTransferTaskLabelToResReq-910 4 whole card",
		fields: VNodeTransferTaskLabelToResReqFields{
			ServerType: "Ascend910-32",
			TotalRes:   util.VResource{Aicore: nodeTotalCoreNum256, Aicpu: nodeTotalCpuNum112},
		},
		args: VNodeTransferTaskLabelToResReqArgs{
			task: FakeVNPUTaskWithResSetting("pod3", "ascend-910", AscendVNPULevelLow, "128",
				AscendDVPPEnabledOn),
		},
		want:    buildVResource(nodeTotalCoreNum128, nodeTotalCoreNum56, AscendDVPPEnabledNull),
		wantErr: false,
	}
}

func buildVNodeTransferTaskLabelToResReqTestCases() []VNodeTransferTaskLabelToResReqTests {
	tests := []VNodeTransferTaskLabelToResReqTests{
		buildVNodeTransferTaskLabelToResReqTestCase01(),
		buildVNodeTransferTaskLabelToResReqTestCase02(),
		buildVNodeTransferTaskLabelToResReqTestCase03(),
		buildVNodeTransferTaskLabelToResReqTestCase04(),
	}
	return tests
}

func TestVNodeTransferTaskLabelToResReq(t *testing.T) {
	tests := buildVNodeTransferTaskLabelToResReqTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vNode := &VNode{
				Chips:        tt.fields.Chips,
				ChipKind:     tt.fields.ChipKind,
				ServerType:   tt.fields.ServerType,
				TotalChipNum: tt.fields.TotalChipNum,
				FreeChipNum:  tt.fields.FreeChipNum,
				TotalRes:     tt.fields.TotalRes,
			}
			got, err := vNode.TransferTaskLabelToResReq(tt.args.task)
			if (err != nil) != tt.wantErr {
				t.Errorf("TransferTaskLabelToResReq() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TransferTaskLabelToResReq() got = %v, want %v", got, tt.want)
			}
		})
	}
}
