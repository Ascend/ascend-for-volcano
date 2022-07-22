/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package comvnpu is using for virtual HuaWei Ascend910 schedule.

*/
package comvnpu

import (
	"errors"
	"reflect"
	"testing"

	"volcano.sh/volcano/pkg/scheduler/conf"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
)

type preHandleVNPUTest struct {
	name    string
	ssn     *framework.Session
	wantErr error
}

func buildPreHandleVNPUTestCases() []preHandleVNPUTest {
	ssn1 := test.FakeNormalSSN()
	test.AddConfigIntoFakeSSN(ssn1, []conf.Configuration{{Name: util.CMInitParamKey,
		Arguments: map[string]string{util.SegmentEnable: "false"}}})
	testCases := []preHandleVNPUTest{
		{
			name:    "01-getVNPUUsedChipByReqTest jobOrder-test",
			ssn:     ssn1,
			wantErr: errors.New(util.SegmentSetFalse),
		},
	}
	return testCases
}

// TestPreHandleVNPU test PreHandleVNPU function
func TestPreHandleVNPU(t *testing.T) {
	tests := buildPreHandleVNPUTestCases()
	tp := &VNPU{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tp.PreHandleVNPU(tt.ssn)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("GetVNPUUsedChipByReq() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
