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

const (
	timeTmp   = 1648556261
	checkCode = 991792085
)

type getVNPUCMDataArgs struct {
	cacheData      VNPUAllocInfCache
	cacheFunBefore func()
	cacheFunAfter  func()
}

type getVNPUCMDataTest struct {
	name string
	args getVNPUCMDataArgs
	want map[string]string
}

func buildGetVNPUCMData01TestCases(tmpPatche *gomonkey.Patches) getVNPUCMDataTest {
	test01 := getVNPUCMDataTest{
		name: "01-getVNPUCMDataTests- nil-test",
		args: getVNPUCMDataArgs{
			cacheFunBefore: func() {
				tmpPatche = gomonkey.ApplyFunc(time.Now,
					func() time.Time { return time.Unix(timeTmp, 0) })
			}, cacheFunAfter: func() {
				tmpPatche.Reset()
			},
		},
		want: map[string]string{VNPCMDataKey: `{"Nodes":null,"UpdateTime":1648556261,"CheckCode":3096267169}`},
	}
	return test01
}

func buildGetVNPUCMData02TestCases(tmpPatche *gomonkey.Patches) getVNPUCMDataTest {
	test03 := getVNPUCMDataTest{
		name: "02-getVNPUCMDataTests- no pre-deal-test",
		args: getVNPUCMDataArgs{cacheData: VNPUAllocInfCache{
			Cache: []VNPUAllocInf{{api.JobID("btg-test/mindx-dls-npu-8c"),
				"huawei.com/Ascend910-8c", "k8smaster", "Ascend910-5",
				"Ascend910-8c-180-5", false, timeTmp}}, CheckCode: checkCode},
			cacheFunBefore: func() {
				tmpPatche = gomonkey.ApplyFunc(time.Now,
					func() time.Time { return time.Unix(timeTmp, 0) })
			}, cacheFunAfter: func() {
				tmpPatche.Reset()
			},
		},
		want: map[string]string{VNPCMDataKey: `{"Nodes":null,"UpdateTime":1648556261,"CheckCode":3096267169}`},
	}
	return test03
}

func buildGetVNPUCMDataTestCases() []getVNPUCMDataTest {
	var tmpPatche *gomonkey.Patches
	testCases := []getVNPUCMDataTest{
		buildGetVNPUCMData01TestCases(tmpPatche),
		buildGetVNPUCMData02TestCases(tmpPatche),
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
