/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package vnpuutil is using for virtual HuaWei Ascend910 schedule.

*/
package vnpuutil

import (
	"errors"
	"reflect"
	"testing"
	"volcano.sh/volcano/pkg/scheduler/conf"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/util"
)

type checkVNPUSegmentEnableByConfigArgs struct {
	configurations []conf.Configuration
}

type checkVNPUSegmentEnableByConfigTest struct {
	name string
	args checkVNPUSegmentEnableByConfigArgs
	want error
}

func buildCheckVNPUSegmentEnableByConfigTestCases() []checkVNPUSegmentEnableByConfigTest {
	confs := []conf.Configuration{{Name: util.CMInitParamKey,
		Arguments: map[string]string{util.SegmentEnable: "false"}}}
	testCases := []checkVNPUSegmentEnableByConfigTest{
		{
			name: "01-CheckVNPUSegmentEnableByConfig() nil conf",
			args: checkVNPUSegmentEnableByConfigArgs{
				configurations: nil},
			want: errors.New(util.SegmentNoEnable),
		},
		{
			name: "02-CheckVNPUSegmentEnableByConfig() success test",
			args: checkVNPUSegmentEnableByConfigArgs{
				configurations: confs},
			want: errors.New(util.SegmentSetFalse),
		},
	}
	return testCases
}

// TestCheckVNPUSegmentEnableByConfig test CheckVNPUSegmentEnableByConfig
func TestCheckVNPUSegmentEnableByConfig(t *testing.T) {
	tests := buildCheckVNPUSegmentEnableByConfigTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckVNPUSegmentEnableByConfig(tt.args.configurations); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CheckVNPUSegmentEnableByConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
