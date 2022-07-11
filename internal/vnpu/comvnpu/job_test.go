/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package comvnpu is using for virtual HuaWei Ascend910 schedule.

*/
package comvnpu

import (
	"fmt"
	"reflect"
	"testing"

	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/vnpu/modulev910"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/vnpu/vnpuutil"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
)

type getVJobMeetNodeListArgs struct {
	vJob *api.JobInfo
	res  map[string]int
	ssn  *framework.Session
}

type getVJobMeetNodeListTest struct {
	name    string
	fields  VNPU
	args    getVJobMeetNodeListArgs
	want    []*api.NodeInfo
	wantErr error
}

func buildGetVJobMeetNodeListTestCases() []getVJobMeetNodeListTest {
	const npuCoreNum = 32
	job0 := test.FakeNormalTestJobByCreatTime("pg0", util.NPUIndex2, 0)
	test.AddTestJobPodGroup(job0)

	test.SetFakeJobRequestSource(job0, npuV910CardName16c, 1)
	node0 := test.FakeNormalTestNode("node0")
	test.SetTestNPUNodeAnnotation(node0, vnpuutil.NPU910CardName, "Ascend910-1")
	test.SetTestNPUNodeAnnotation(node0, vnpuutil.NPU910CardCoreKey, "1-32c-32")
	test.SetFakeNodeIdleSource(node0, vnpuutil.NPU910CardName, 1)
	test.SetNPUNodeLabel(node0.Node, vnpuutil.VNPUNodeLabelKey, vnpuutil.VNPUNodeLabelValue)
	test.SetNPUNodeLabel(node0.Node, util.ArchSelector, util.HuaweiArchX86)

	ssn1 := test.FakeNormalSSN()
	test.AddJobIntoFakeSSN(ssn1, job0)
	test.AddNodeIntoFakeSSN(ssn1, node0)
	vnpu := &VNPU{}
	vnpu910 := &modulev910.ChipV910{}
	if getErr := vnpu910.InitVNPUPlugin(); getErr != nil {
		return nil
	}
	vnpu.Attr = vnpu910.ComVNPU
	testCases := []getVJobMeetNodeListTest{
		{
			name: "01-getVJobMeetNodeList jobOrder-test",
			args: getVJobMeetNodeListArgs{
				vJob: job0, res: nil, ssn: ssn1},
			want:    nil,
			wantErr: fmt.Errorf("total resource map[] not meet req %s", npuV910CardName16c),
		},
		{
			name:   "02-getVJobMeetNodeList jobOrder-test",
			fields: *vnpu,
			args: getVJobMeetNodeListArgs{
				vJob: job0, res: map[string]int{node0.Name: npuCoreNum}, ssn: ssn1},
			want:    []*api.NodeInfo{node0},
			wantErr: nil,
		},
	}
	return testCases
}

func TestGetVJobMeetNodeList(t *testing.T) {
	tests := buildGetVJobMeetNodeListTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp := &VNPU{
				Plugin:               tt.fields.Plugin,
				Attr:                 tt.fields.Attr,
				HwNPUSchedulerPlugin: tt.fields.HwNPUSchedulerPlugin,
			}
			got, err := tp.GetVJobMeetNodeList(tt.args.vJob, tt.args.res, tt.args.ssn)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("GetVJobMeetNodeList() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetVJobMeetNodeList() got = %v, want %v", got, tt.want)
			}
		})
	}
}
