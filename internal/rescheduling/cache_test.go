/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package rescheduling is using for HuaWei Ascend pin fault rescheduling.

*/
package rescheduling

import (
	"encoding/json"
	"testing"
)

func fakeInvalidReSchedulerCMData() *DealReSchedulerConfigmap {
	return &DealReSchedulerConfigmap{
		CMName:      CmName,
		CMNameSpace: CmNameSpace,
		CMData:      nil,
	}
}

func dealMarshal(data interface{}) string {
	dataString, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	return string(dataString)
}

func fakeNormalReSchedulerCMData() *DealReSchedulerConfigmap {
	FaultNodes := []FaultNode{
		*FakeTestFaultNodeNodeHealthy("node0"),
		*FakeTestFaultNodeNodeHealthy("node1"),
	}
	FaultJobs := []FaultJob{
		*FakeTestFaultJob([]string{"node0", "node1"}, []string{"0", "9"},
			nil, "job1", "test"),
	}

	fNodeBuffer := dealMarshal(FaultNodes)
	fJobBuffer := dealMarshal(FaultJobs)

	return &DealReSchedulerConfigmap{
		CMName:      CmName,
		CMNameSpace: CmNameSpace,
		CMData: map[string]string{
			CmFaultNodeKind:     string(fNodeBuffer),
			CmFaultJob910x8Kind: string(fJobBuffer),
		},
	}
}

type DealReSchedulerCacheSetFaultNodesFromCMTests struct {
	name    string
	fields  *DealReSchedulerCache
	wantErr bool
}

func buildDealReSchedulerCacheSetFaultNodesFromCMTests() []DealReSchedulerCacheSetFaultNodesFromCMTests {
	field1 := FakeReSchedulerCache()
	field1.DealReSchedulerConfigmap = fakeInvalidReSchedulerCMData()
	field2 := FakeReSchedulerCache()
	field2.DealReSchedulerConfigmap = fakeNormalReSchedulerCMData()

	test1 := DealReSchedulerCacheSetFaultNodesFromCMTests{
		name:    "01-DealReSchedulerCache_SetFaultNodesFromCM()  invalid cache structure",
		fields:  field1,
		wantErr: true,
	}
	test2 := DealReSchedulerCacheSetFaultNodesFromCMTests{
		name:    "01-DealReSchedulerCache_SetFaultNodesFromCM()  succeed",
		fields:  field2,
		wantErr: false,
	}
	testCases := []DealReSchedulerCacheSetFaultNodesFromCMTests{
		test1,
		test2,
	}
	return testCases
}

func TestDealReSchedulerCacheSetFaultNodesFromCM(t *testing.T) {
	tests := buildDealReSchedulerCacheSetFaultNodesFromCMTests()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reCache := &DealReSchedulerCache{
				FaultNodes:               tt.fields.FaultNodes,
				FaultJobs:                tt.fields.FaultJobs,
				DealReSchedulerConfigmap: tt.fields.DealReSchedulerConfigmap,
			}
			if err := reCache.SetFaultNodesFromCM(); (err != nil) != tt.wantErr {
				t.Errorf("SetFaultNodesFromCM() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
