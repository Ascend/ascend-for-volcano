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

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/vnpu/modulev910"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/vnpu/vnpuutil"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
)

type orderVJobsByCreateTimeArgs struct {
	jobs []*api.JobInfo
}

type orderVJobsByCreateTimeTests struct {
	name    string
	fields  VNPU
	args    orderVJobsByCreateTimeArgs
	want    []*api.JobInfo
	wantErr error
}

func buildOrderVJobsByCreateTimeTestCases() []orderVJobsByCreateTimeTests {
	job0 := ascendtest.FakeNormalTestJobByCreatTime("pg0", util.ConstIntNum2, 0)
	job1 := ascendtest.FakeNormalTestJobByCreatTime("pg1", util.ConstIntNum2, 1)
	job2 := ascendtest.FakeNormalTestJobByCreatTime("pg2", util.ConstIntNum2, constIntNum2)
	job3 := ascendtest.FakeNormalTestJobByCreatTime("pg3", util.ConstIntNum2, constIntNum4)

	testCases := []orderVJobsByCreateTimeTests{
		{
			name: "01-orderVJobsByCreateTimeTests jobOrder-test",
			args: orderVJobsByCreateTimeArgs{
				jobs: []*api.JobInfo{job3, job2, job0, job1}},
			want:    []*api.JobInfo{job0, job1, job2, job3},
			wantErr: nil,
		},
		{
			name: "02-orderVJobsByCreateTimeTests jobOrder-test",
			args: orderVJobsByCreateTimeArgs{
				jobs: []*api.JobInfo{job1, job2, job0, job3}},
			want:    []*api.JobInfo{job0, job1, job2, job3},
			wantErr: nil,
		},
	}
	return testCases
}

// TestVNPU_OrderVJobsByCreateTime test WriteReSchedulerDataToCM function
func TestOrderVJobsByCreateTime(t *testing.T) {
	tests := buildOrderVJobsByCreateTimeTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp := &VNPU{
				Plugin:               tt.fields.Plugin,
				Attr:                 tt.fields.Attr,
				HwNPUSchedulerPlugin: tt.fields.HwNPUSchedulerPlugin,
			}
			got, err := tp.OrderVJobsByCreateTime(tt.args.jobs)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("OrderVJobsByCreateTime() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("OrderVJobsByCreateTime() got = %v, want %v", got, tt.want)
			}
		})
	}
}

type getVJobMeetNodeListArgs struct {
	vJob *api.JobInfo
	res  map[string]float64
	ssn  *framework.Session
}

type getVJobMeetNodeListTests []struct {
	name    string
	fields  VNPU
	args    getVJobMeetNodeListArgs
	want    []*api.NodeInfo
	wantErr error
}

func buildGetVJobMeetNodeListTestCases() getVJobMeetNodeListTests {
	job0 := ascendtest.FakeNormalTestJobByCreatTime("pg0", util.ConstIntNum2, 0)
	ascendtest.AddTestJobPodGroup(job0)

	setJobResourceReq(job0, npuV910CardName16c, float64(1))
	node0 := ascendtest.FakeNormalTestNode("node0")
	ascendtest.SetTestNPUNodeAnnotation(node0, vnpuutil.NPU910CardName, "Ascend910-1")
	ascendtest.SetTestNPUNodeAnnotation(node0, vnpuutil.NPU910CardCoreKey, "1-32c-32")
	ascendtest.SetNPUNodeLabel(node0.Node, vnpuutil.VNPUNodeLabelKey, vnpuutil.VNPUNodeLabelValue)
	ascendtest.SetNPUNodeLabel(node0.Node, util.ArchSelector, util.HuaweiArchX86)

	ssn1 := ascendtest.FakeNormalSSN()
	ascendtest.AddJobIntoFakeSSN(ssn1, job0)
	ascendtest.AddNodeIntoFakeSSN(ssn1, node0)
	vnpu := &VNPU{}
	vnpu910 := &modulev910.ChipV910{}
	if getErr := vnpu910.InitVNPUPlugin(); getErr != nil {
		return getVJobMeetNodeListTests{}
	}
	vnpu.Attr = vnpu910.ComVNPU
	testCases := getVJobMeetNodeListTests{
		{
			name: "01-getVJobMeetNodeList jobOrder-test",
			args: getVJobMeetNodeListArgs{
				vJob: job0, res: nil, ssn: ssn1},
			want:    nil,
			wantErr: fmt.Errorf("total resource not meet req %s", npuV910CardName16c),
		},
		{
			name:   "02-getVJobMeetNodeList jobOrder-test",
			fields: *vnpu,
			args: getVJobMeetNodeListArgs{
				vJob: job0, res: map[string]float64{vnpuutil.NPU910CardName: 1}, ssn: ssn1},
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
