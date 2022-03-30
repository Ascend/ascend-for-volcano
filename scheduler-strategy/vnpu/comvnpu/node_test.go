/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package comvnpu is using for virtual HuaWei Ascend910 schedule.

*/
package comvnpu

import (
	"github.com/agiledragon/gomonkey/v2"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"reflect"
	"testing"
	"volcano.sh/volcano/pkg/scheduler/framework"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/util"

	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/vnpu/vnpuutil"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
)

type getNodeNPUCoreInfoMapArgs struct {
	vNode *api.NodeInfo
}

type getNodeNPUCoreInfoMapTests []struct {
	name    string
	fields  VNPU
	args    getNodeNPUCoreInfoMapArgs
	want    map[string]vNPUCoreInfo
	wantErr error
}

func buildGetNodeNPUCoreInfoMapTestCases() getNodeNPUCoreInfoMapTests {
	const maxCoreNum = 32
	nodeInf := ascendtest.FakeNormalTestNode("vNode")
	ascendtest.SetTestNPUNodeAnnotation(nodeInf, vnpuutil.NPU910CardCoreKey, "0-32c-32c")
	testCases := getNodeNPUCoreInfoMapTests{
		{
			name: "01-orderVJobsByCreateTimeTests jobOrder-test",
			fields: VNPU{
				Attr: vnpuutil.ComVNPU{NPUCardCoreKey: vnpuutil.NPU910CardCoreKey,
					HwEntity: plugin.HwEntity{
						AnnoName: vnpuutil.NPU910CardName, AnnoPreVal: vnpuutil.NPUCardNamePrefix}},
			},
			args:    getNodeNPUCoreInfoMapArgs{vNode: nodeInf},
			want:    map[string]vNPUCoreInfo{"Ascend910-0": {ChipID: 0, AllCore: maxCoreNum, UnCutCore: maxCoreNum}},
			wantErr: nil,
		},
	}
	return testCases
}

// TestVNPU_GetNodeNPUCoreInfoMap test GetNodeNPUCoreInfoMap function
func TestVNPU_GetNodeNPUCoreInfoMap(t *testing.T) {
	tests := buildGetNodeNPUCoreInfoMapTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp := &VNPU{
				Plugin:               tt.fields.Plugin,
				Attr:                 tt.fields.Attr,
				HwNPUSchedulerPlugin: tt.fields.HwNPUSchedulerPlugin,
			}

			got, err := tp.GetNodeNPUCoreInfoMap(tt.args.vNode)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("GetNodeNPUCoreInfoMap() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetNodeNPUCoreInfoMap() got = %v, want %v", got, tt.want)
			}
		})
	}
}

type getVNPUUsedChipByReqArgs struct {
	needNPU string
	vNode   *api.NodeInfo
}

type getVNPUUsedChipByReqTests []struct {
	name    string
	fields  VNPU
	args    getVNPUUsedChipByReqArgs
	want    string
	wantErr error
}

func buildGetVNPUUsedChipByReqTestCases() getVNPUUsedChipByReqTests {
	nodeInf := ascendtest.FakeNormalTestNode("vNode")
	ascendtest.SetTestNPUNodeAnnotation(nodeInf, vnpuutil.NPU910CardCoreKey, "0-32c-32c,1-32c-30c")
	ascendtest.SetTestNPUNodeAnnotation(nodeInf, vnpuutil.NPU910CardName, "Ascend91-0,Ascend91-1")
	testCases := getVNPUUsedChipByReqTests{
		{
			name: "01-getVNPUUsedChipByReqTest jobOrder-test",
			fields: VNPU{
				Attr: vnpuutil.ComVNPU{NPUCardCoreKey: vnpuutil.NPU910CardCoreKey,
					HwEntity: plugin.HwEntity{AnnoName: vnpuutil.NPU910CardName, AnnoPreVal: vnpuutil.NPUCardNamePrefix}},
			},
			args:    getVNPUUsedChipByReqArgs{needNPU: "huawei.com/Ascend910-16c", vNode: nodeInf},
			want:    "Ascend910-1",
			wantErr: nil,
		},
	}
	return testCases
}

// TestVNPU_GetVNPUUsedChipByReq test GetNodeNPUCoreInfoMap function
func TestVNPU_GetVNPUUsedChipByReq(t *testing.T) {
	tests := buildGetVNPUUsedChipByReqTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp := &VNPU{
				Plugin:               tt.fields.Plugin,
				Attr:                 tt.fields.Attr,
				HwNPUSchedulerPlugin: tt.fields.HwNPUSchedulerPlugin,
			}
			got, err := tp.GetVNPUUsedChipByReq(tt.args.needNPU, tt.args.vNode)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("GetVNPUUsedChipByReq() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetVNPUUsedChipByReq() got = %v, want %v", got, tt.want)
			}
		})
	}
}

type preHandleVNPUArgs struct {
	ssn            *framework.Session
	cacheFunBefore func()
	cacheFunAfter  func()
}

type preHandleVNPUTests []struct {
	name    string
	fields  VNPU
	args    preHandleVNPUArgs
	wantErr error
}

func buildPreHandleVNPUTestCases() preHandleVNPUTests {
	node0 := ascendtest.FakeNormalTestNode("node0")
	ascendtest.SetTestNPUNodeAnnotation(node0, vnpuutil.NPU910CardCoreKey, "0-32c-32c,1-32c-30c")
	ascendtest.SetTestNPUNodeAnnotation(node0, vnpuutil.NPU910CardName, "Ascend91-0,Ascend91-1")
	job0 := ascendtest.FakeNormalTestJobByCreatTime("pg0", util.ConstIntNum2, 0)
	job1 := ascendtest.FakeNormalTestJobByCreatTime("pg1", util.ConstIntNum2, 1)

	ssn1 := ascendtest.FakeNormalSSN()
	ascendtest.AddJobIntoFakeSSN(ssn1, job0)
	ascendtest.AddJobIntoFakeSSN(ssn1, job1)
	ascendtest.AddNodeIntoFakeSSN(ssn1, node0)
	var tmpPatche *gomonkey.Patches
	var tmpPatche1 *gomonkey.Patches
	testCases := preHandleVNPUTests{
		{
			name: "01-getVNPUUsedChipByReqTest jobOrder-test",
			fields: VNPU{
				Attr: vnpuutil.ComVNPU{NPUCardCoreKey: vnpuutil.NPU910CardCoreKey,
					HwEntity: plugin.HwEntity{AnnoName: vnpuutil.NPU910CardName, AnnoPreVal: vnpuutil.NPUCardNamePrefix}},
			},
			args: preHandleVNPUArgs{ssn: ssn1,
				cacheFunBefore: func() {
					tmpPatche = gomonkey.ApplyFunc(util.CreateOrUpdateConfigMap,
						func(k8s kubernetes.Interface, cm *v1.ConfigMap, cmName, cmNameSpac string) error {
							return nil
						})
					tmpPatche1 = gomonkey.ApplyFunc(util.GetConfigMapWithRetry,
						func(k8s kubernetes.Interface, cmNameSpac, cmName string) (*v1.ConfigMap, error) {
							var cm = v1.ConfigMap{Data: make(map[string]string, constIntNum4)}
							cm.Data[vnpuutil.VNPCMDataKey] = "{\"Nodes\":[{\"NodeName\":\"k8smaster\",\"Cards\":[" +
								"{\"CardName\":\"Ascend910-5\",\"Req\":[\"huawei.com/Ascend910-8c\"],\"Alloc\":" +
								"[\"Ascend910-8c-180-5\"]}]}],\"UpdateTime\":1648556261,\"CheckCode\":4009496266}"
							return &cm, nil
						})

				}, cacheFunAfter: func() {
					tmpPatche.Reset()
					tmpPatche1.Reset()
				}},
			wantErr: nil,
		},
	}
	return testCases
}

// TestVNPU_PreHandleVNPU test PreHandleVNPU function
func TestVNPU_PreHandleVNPU(t *testing.T) {
	tests := buildPreHandleVNPUTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp := &VNPU{
				Plugin:               tt.fields.Plugin,
				Attr:                 tt.fields.Attr,
				HwNPUSchedulerPlugin: tt.fields.HwNPUSchedulerPlugin,
			}
			tt.args.cacheFunBefore()
			err := tp.PreHandleVNPU(tt.args.ssn)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("GetVNPUUsedChipByReq() error = %v, wantErr %v", err, tt.wantErr)
			}
			tt.args.cacheFunAfter()
		})
	}
}
