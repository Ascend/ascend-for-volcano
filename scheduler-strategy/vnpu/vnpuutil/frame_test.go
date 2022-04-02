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
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/util"

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
		want: map[string]string{VNPCMDataKey: `{"Nodes":null,"UpdateTime":0,"CheckCode":0}`},
	}
	return test01
}

func buildGetVNPUCMData02TestCases(tmpPatche *gomonkey.Patches) getVNPUCMDataTest {
	test02 := getVNPUCMDataTest{
		name: "02-getVNPUCMDataTests- not-nil-test",
		args: getVNPUCMDataArgs{cacheData: VNPUAllocInfCache{
			Cache: []VNPUAllocInf{{api.JobID("btg-test/mindx-dls-npu-8c"),
				"huawei.com/Ascend910-8c", "k8smaster", "Ascend910-5",
				"Ascend910-8c-180-5", true, timeTmp}}, CheckCode: checkCode},
			cacheFunBefore: func() {
				tmpPatche = gomonkey.ApplyFunc(time.Now,
					func() time.Time { return time.Unix(timeTmp, 0) })
			}, cacheFunAfter: func() {
				tmpPatche.Reset()
			},
		},
		want: map[string]string{VNPCMDataKey: "{\"Nodes\":[{\"NodeName\":\"k8smaster\",\"Cards\":[" +
			"{\"CardName\":\"Ascend910-5\",\"Req\":[\"huawei.com/Ascend910-8c\"],\"Alloc\":" +
			"[\"Ascend910-8c-180-5\"]}]}],\"UpdateTime\":1648556261,\"CheckCode\":4009496266}"},
	}
	return test02
}

func buildGetVNPUCMData03TestCases(tmpPatche *gomonkey.Patches) getVNPUCMDataTest {
	test03 := getVNPUCMDataTest{
		name: "03-getVNPUCMDataTests- no pre-deal-test",
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
		want: map[string]string{VNPCMDataKey: `{"Nodes":null,"UpdateTime":0,"CheckCode":0}`},
	}
	return test03
}

func buildGetVNPUCMData04TestCases(tmpPatche *gomonkey.Patches) getVNPUCMDataTest {
	test04 := getVNPUCMDataTest{
		name: "04-getVNPUCMDataTests- no pre-deal-test",
		args: getVNPUCMDataArgs{cacheData: VNPUAllocInfCache{
			Cache: []VNPUAllocInf{{JobUID: api.JobID("btg-test/mindx-dls-vnpu-1c-1"),
				ReqNPUType:    "huawei.com/Ascend710-1c",
				NodeName:      "centos-6543",
				ReqCardName:   "Ascend710-2",
				AllocCardName: "Ascend710-1c-132-2",
				AllocFlag:     true,
				UpdateTime:    timeTmp},
				{JobUID: api.JobID("btg-test/mindx-dls-npu-2c-1"),
					ReqNPUType:    "huawei.com/Ascend710-2c",
					NodeName:      "centos-6543",
					ReqCardName:   "Ascend710-4",
					AllocCardName: "",
					AllocFlag:     true,
					UpdateTime:    timeTmp}}, CheckCode: checkCode},
			cacheFunBefore: func() {
				tmpPatche = gomonkey.ApplyFunc(time.Now,
					func() time.Time { return time.Unix(timeTmp, 0) })
				tmpPatche = gomonkey.ApplyFunc(util.MakeDataHash,
					func(data interface{}) uint32 { return checkCode })
			}, cacheFunAfter: func() {
				tmpPatche.Reset()
			},
		},
		want: map[string]string{VNPCMDataKey: "{\"Nodes\":[{\"NodeName\":\"centos-6543\",\"Cards\":[{\"CardName\":\"Ascend710-2\",\"Req\":[\"huawei.com/Ascend710-1c\"],\"Alloc\":[\"Ascend710-1c-132-2\"]},{\"CardName\":\"Ascend710-4\",\"Req\":[\"huawei.com/Ascend710-2c\"],\"Alloc\":[\"\"]}]}],\"UpdateTime\":1648556261,\"CheckCode\":991792085}"},
	}
	return test04
}

func buildGetVNPUCMData05TestCases(tmpPatche *gomonkey.Patches) getVNPUCMDataTest {
	test04 := getVNPUCMDataTest{
		name: "04-getVNPUCMDataTests- no pre-deal-test",
		args: getVNPUCMDataArgs{cacheData: VNPUAllocInfCache{
			Cache: []VNPUAllocInf{{

				JobUID:        api.JobID("btg-test/mindx-dls-vnpu-1c-1"),
				ReqNPUType:    "huawei.com/Ascend710-1c",
				NodeName:      "centos-6543",
				ReqCardName:   "Ascend710-4",
				AllocCardName: "Ascend710-1c-164-4",
				AllocFlag:     true,
				UpdateTime:    1648864184,
			},
				{
					JobUID:        api.JobID("btg - test/mindx - dls - vnpu - 1c-2"),
					ReqNPUType:    "huawei.com/Ascend710-1c",
					NodeName:      "centos-6543",
					ReqCardName:   "Ascend710-4",
					AllocCardName: "",
					AllocFlag:     true,
					UpdateTime:    1648864103,
				}, {
					JobUID:        api.JobID("btg-test/mindx-dls-vnpu-1c-3"),
					ReqNPUType:    "huawei.com/Ascend710-1c",
					NodeName:      "centos-6543",
					ReqCardName:   "Ascend710-4",
					AllocCardName: "",
					AllocFlag:     true,
					UpdateTime:    1648864106,
				}}},
			cacheFunBefore: func() {
				tmpPatche = gomonkey.ApplyFunc(time.Now,
					func() time.Time { return time.Unix(timeTmp, 0) })
				tmpPatche = gomonkey.ApplyFunc(util.MakeDataHash,
					func(data interface{}) uint32 { return checkCode })
			}, cacheFunAfter: func() {
				tmpPatche.Reset()
			},
		},
		want: map[string]string{VNPCMDataKey: ""},
	}
	return test04
}

func buildGetVNPUCMDataTestCases() []getVNPUCMDataTest {
	var tmpPatche *gomonkey.Patches
	testCases := []getVNPUCMDataTest{
		buildGetVNPUCMData01TestCases(tmpPatche),
		buildGetVNPUCMData02TestCases(tmpPatche),
		buildGetVNPUCMData03TestCases(tmpPatche),
		buildGetVNPUCMData04TestCases(tmpPatche),
		//buildGetVNPUCMData05TestCases(tmpPatche),
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
