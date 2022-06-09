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
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/vnpu/modulev910"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/vnpu/vnpuutil"
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
	job0 := test.FakeNormalTestJobByCreatTime("pg0", util.NPUIndex2, 0)
	job1 := test.FakeNormalTestJobByCreatTime("pg1", util.NPUIndex2, 1)
	job2 := test.FakeNormalTestJobByCreatTime("pg2", util.NPUIndex2, util.NPUIndex2)
	job3 := test.FakeNormalTestJobByCreatTime("pg3", util.NPUIndex2, util.NPUIndex4)

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

type recordNewVNPUJobInCacheArgs struct {
	job      *api.JobInfo
	cacheFun func()
}

type recordNewVNPUJobInCacheTests struct {
	name    string
	fields  vnpuPlugin
	args    recordNewVNPUJobInCacheArgs
	wantErr error
}

func buildRecordNewVNPUJobInCacheTestCases() []recordNewVNPUJobInCacheTests {
	testJob := test.FakeNormalTestJob("test", util.NPUIndex2)
	testCases := []recordNewVNPUJobInCacheTests{
		{
			name: "01-RecordNewVNPUJobInCache() no npu job.",
			args: recordNewVNPUJobInCacheArgs{
				job: testJob, cacheFun: func() {
					test.SetFakeJobRequestSource(testJob, noPrefixResourceType, 1)
				}},
			wantErr: errors.New("nil NPU"),
		},
		{
			name: "02-RecordNewVNPUJobInCache() no VNPUAllocData, should return nil",
			args: recordNewVNPUJobInCacheArgs{
				job: testJob, cacheFun: func() {
					test.SetFakeJobRequestSource(testJob, npuV910CardName16c, 1)
				}},
			fields:  vnpuPlugin{},
			wantErr: nil,
		},
		{
			name: "03-RecordNewVNPUJobInCache() success test",
			args: recordNewVNPUJobInCacheArgs{
				job: testJob, cacheFun: func() {
					test.SetFakeJobRequestSource(testJob, npuV910CardName16c, 1)
					addTestJobIntoVNPUAllocDataCache(testJob)
				}},
			fields:  vnpuPlugin{},
			wantErr: nil,
		},
	}
	return testCases
}

// RecordNewVNPUJobInCache test RecordNewVNPUJobInCache
func TestRecordNewVNPUJobInCache(t *testing.T) {
	tests := buildRecordNewVNPUJobInCacheTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp := &VNPU{
				Plugin:               tt.fields.Plugin,
				Attr:                 tt.fields.Attr,
				HwNPUSchedulerPlugin: tt.fields.HwNPUSchedulerPlugin,
			}
			tt.args.cacheFun()
			err := tp.RecordNewVNPUJobInCache(tt.args.job)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("RecordNewVNPUJobInCache() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
