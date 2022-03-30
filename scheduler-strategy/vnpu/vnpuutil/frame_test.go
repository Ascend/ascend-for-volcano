/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package vnpuutil is using for virtual HuaWei Ascend910 schedule.

*/
package vnpuutil

import (
	"reflect"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"volcano.sh/volcano/pkg/scheduler/api"
)

type getVNPUCMDataArgs struct {
	cacheData      VNPUAllocInfCache
	cacheFunBefore func()
	cacheFunAfter  func()
}

type getVNPUCMDataTests []struct {
	name string
	args getVNPUCMDataArgs
	want map[string]string
}

func buildGetVNPUCMDataTestCases() getVNPUCMDataTests {
	var tmpPatche *gomonkey.Patches
	testCases := getVNPUCMDataTests{
		{
			name: "01-getVNPUCMDataTests- nil-test",
			args: getVNPUCMDataArgs{
				cacheFunBefore: func() {
					tmpPatche = gomonkey.ApplyFunc(time.Now,
						func() time.Time { return time.Unix(1648556261, 0) })
				}, cacheFunAfter: func() {
					tmpPatche.Reset()
				},
			},
			want: map[string]string{VNPCMDataKey: `{"Nodes":null,"UpdateTime":0,"CheckCode":0}`},
		},
		{
			name: "02-getVNPUCMDataTests- not-nil-test",
			args: getVNPUCMDataArgs{cacheData: VNPUAllocInfCache{
				Cache: []VNPUAllocInf{{
					JobUID:        api.JobID("btg-test/mindx-dls-npu-8c"),
					ReqNPUType:    "huawei.com/Ascend910-8c",
					NodeName:      "k8smaster",
					ReqCardName:   "Ascend910-5",
					AllocCardName: "Ascend910-8c-180-5",
					AllocFlag:     true,
					UpdateTime:    1648556261}}, CheckCode: 991792085},
				cacheFunBefore: func() {
					tmpPatche = gomonkey.ApplyFunc(time.Now,
						func() time.Time { return time.Unix(1648556261, 0) })
				}, cacheFunAfter: func() {
					tmpPatche.Reset()
				},
			},
			want: map[string]string{VNPCMDataKey: "{\"Nodes\":[{\"NodeName\":\"k8smaster\",\"Cards\":[" +
				"{\"CardName\":\"Ascend910-5\",\"Req\":[\"huawei.com/Ascend910-8c\"],\"Alloc\":" +
				"[\"Ascend910-8c-180-5\"]}]}],\"UpdateTime\":1648556261,\"CheckCode\":4009496266}"},
		},
	}
	return testCases
}

// TestVNPU_GetNodeNPUCoreInfoMap test GetNodeNPUCoreInfoMap function
func TestGetVNPUCMData(t *testing.T) {
	tests := buildGetVNPUCMDataTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.cacheFunBefore()
			got := GetVNPUCMData(tt.args.cacheData)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetNodeNPUCoreInfoMap() got = %v, want %v", got, tt.want)
			}
			tt.args.cacheFunAfter()
		})
	}
}
