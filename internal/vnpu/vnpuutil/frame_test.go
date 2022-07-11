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
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/conf"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
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

func buildGetVNPUCMData03TestCases(tmpPatche *gomonkey.Patches) getVNPUCMDataTest {
	test03 := getVNPUCMDataTest{
		name: "03-getVNPUCMDataTests- no pre-deal-test",
		args: getVNPUCMDataArgs{cacheData: VNPUAllocInfCache{
			Cache: []VNPUAllocInf{{api.JobID("btg-test/mindx-dls-npu-8c"),
				"huawei.com/Ascend910-8c", "k8smaster", "Ascend910-5",
				"Ascend910-8c-180-5", false, timeTmp},
				{api.JobID("btg-test/mindx-dls-npu-8c"),
					"huawei.com/Ascend310P-2c", "ubuntu-05", "Ascend310P-0",
					"Ascend310P-2c-100-0", false, timeTmp},
			}, CheckCode: checkCode},
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
		buildGetVNPUCMData03TestCases(tmpPatche),
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

type isNPUResourceStableInNodeArgs struct {
	kind    string
	tmpNode *api.NodeInfo
}

type isNPUResourceStableInNodeTest struct {
	name string
	args isNPUResourceStableInNodeArgs
	want bool
}

func buildIsNPUResourceStableInNodeTestCases() []isNPUResourceStableInNodeTest {
	node0 := test.FakeNormalTestNode("node-0")
	node1 := test.FakeNormalTestNode("node-1")
	test.SetTestNPUNodeAnnotation(node1, NPU910CardName, "Ascend91-0,Ascend91-1")
	test.SetFakeNodeIdleSource(node1, NPU910CardName, util.NPUIndex2)
	testCases := []isNPUResourceStableInNodeTest{
		{
			name: "01-IsNPUResourceStableInNode() nil NPU",
			args: isNPUResourceStableInNodeArgs{
				kind: NPU910CardName, tmpNode: node0},
			want: false,
		},
		{
			name: "02-IsNPUResourceStableInNode() success test",
			args: isNPUResourceStableInNodeArgs{
				kind: NPU910CardName, tmpNode: node1},
			want: true,
		},
	}
	return testCases
}

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
