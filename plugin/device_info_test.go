/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package rescheduling is using for HuaWei Ascend pin fault rescheduling.
*/
package plugin

import (
	"k8s.io/apimachinery/pkg/api/resource"
	"reflect"
	"testing"

	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
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
				vDeviceResourceStr: "4c",
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
				vDeviceResourceStr: "4c.3cpu",
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
				vDeviceResourceStr: "4c.3cpu.ndvpp",
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

func TestGetResourceFromStr(t *testing.T) {
	tests := buildGetResourceFromStrTests()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetResourceFromRealStr(tt.args.vDeviceResourceStr); !reflect.DeepEqual(got, tt.want) {
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
				realCardName: "Ascend310P-0,Ascend310P-1",
			},
			want: true,
		},
		{
			name: "02-IsPodWholeCardTest-not whold card",
			args: IsPodWholeCardArgs{realCardName: "Ascend310P-4c.3cpu.ndvpp-100-1-1"},
			want: false,
		},
	}
	return tests
}

func TestIsPodWholeCard(t *testing.T) {
	tests := buildIsPodWholeCardTest()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsPodWholeCard(tt.args.realCardName); got != tt.want {
				t.Errorf("IsPodWholeCard() = %v, want %v", got, tt.want)
			}
		})
	}
}

type TransferTaskLabelToResReqArgs struct {
	task *api.TaskInfo
}

type TransferTaskLabelToResReqTests struct {
	name    string
	args    TransferTaskLabelToResReqArgs
	want    util.VResource
	wantErr bool
}

func FakeVNPUTaskWithResSetting(name, ringController, vnpuLevel, coreNum, dvpp string) *api.TaskInfo {
	task := FakeVNPUTestTask(name, "node0", "", "")
	task.Pod.Labels[util.RingController] = ringController
	task.Pod.Labels[AscendVNPULevel] = vnpuLevel
	task.Pod.Spec.Containers[0].Resources.Requests[util.AscendNPUCore] = resource.MustParse(coreNum)
	task.Pod.Labels[AscendVNPUDVPP] = dvpp
	return task
}

func buildTransferTaskLabelToResReqTestCases() []TransferTaskLabelToResReqTests {
	tests := []TransferTaskLabelToResReqTests{
		//{
		//	name: "01-TransferTaskLabelToResReq-core,aicpu,dvpp",
		//	args: TransferTaskLabelToResReqArgs{
		//		task: FakeVNPUTaskWithResSetting("pod0", "ascend-310P", AscendVNPULevelHigh, "16",
		//			AscendDVPPEnabledOn),
		//	},
		//	want: util.VResource{
		//		Aicore: util.NPUIndex4,
		//		Aicpu:  util.NPUIndex4,
		//		DVPP:   AscendDVPPEnabledOn,
		//	},
		//	wantErr: false,
		//},
		{
			name: "02-TransferTaskLabelToResReq-core,aicpu,dvpp",
			args: TransferTaskLabelToResReqArgs{
				task: FakeVNPUTaskWithResSetting("pod1", "ascend-910", AscendVNPULevelLow, "4",
					AscendDVPPEnabledOn),
			},
			want: util.VResource{
				Aicore: util.NPUIndex4,
				Aicpu:  util.NPUIndex4,
				DVPP:   AscendDVPPEnabledOn,
			},
			wantErr: false,
		},
	}
	return tests
}

func TestTransferTaskLabelToResReq(t *testing.T) {
	tests := buildTransferTaskLabelToResReqTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TransferTaskLabelToResReq(tt.args.task)
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
