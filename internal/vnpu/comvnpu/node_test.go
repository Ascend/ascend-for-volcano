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
	"strings"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"volcano.sh/volcano/pkg/scheduler/conf"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/vnpu/vnpuutil"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
)

type preHandleVNPUArgs struct {
	ssn            *framework.Session
	cacheFunBefore func()
	cacheFunAfter  func()
}

type preHandleVNPUTest struct {
	name    string
	fields  VNPU
	args    preHandleVNPUArgs
	wantErr error
}

func buildPreHandleVNPUTestCases() []preHandleVNPUTest {
	node0 := test.FakeNormalTestNode("node0")
	test.SetTestNPUNodeAnnotation(node0, vnpuutil.NPU910CardCoreKey, "0-32c-32c,1-32c-30c")
	test.SetTestNPUNodeAnnotation(node0, vnpuutil.NPU910CardName, "Ascend91-0,Ascend91-1")
	job0 := test.FakeNormalTestJobByCreatTime("pg0", util.NPUIndex2, 0)
	job1 := test.FakeNormalTestJobByCreatTime("pg1", util.NPUIndex2, 1)

	ssn1 := test.FakeNormalSSN()
	test.AddJobIntoFakeSSN(ssn1, job0)
	test.AddJobIntoFakeSSN(ssn1, job1)
	test.AddNodeIntoFakeSSN(ssn1, node0)
	test.AddConfigIntoFakeSSN(ssn1, []conf.Configuration{{Name: util.CMInitParamKey,
		Arguments: map[string]string{util.SegmentEnable: "false"}}})
	var tmpPatche *gomonkey.Patches
	var tmpPatche1 *gomonkey.Patches
	testCases := []preHandleVNPUTest{
		{
			name: "01-getVNPUUsedChipByReqTest jobOrder-test",
			fields: VNPU{
				Attr: vnpuutil.ComVNPU{HwEntity: plugin.HwEntity{
					AnnoName: vnpuutil.NPU910CardName, AnnoPreVal: vnpuutil.NPUCardNamePrefix}},
			},
			args: preHandleVNPUArgs{ssn: ssn1,
				cacheFunBefore: func() {
					tmpPatche = gomonkey.ApplyFunc(util.CreateOrUpdateConfigMap,
						func(k8s kubernetes.Interface, cm *v1.ConfigMap, cmName, cmNameSpac string) error {
							return nil
						})
					tmpPatche1 = gomonkey.ApplyFunc(util.GetConfigMapWithRetry,
						func(k8s kubernetes.Interface, cmNameSpac, cmName string) (*v1.ConfigMap, error) {
							var cm = v1.ConfigMap{Data: make(map[string]string, util.NPUIndex4)}
							tmp := `{"Nodes":[{"NodeName":"k8smaster","Cards":[{"CardName":"Ascend910-5",
"Req":["huawei.com/Ascend910-8c"],"Alloc":["Ascend910-8c-180-5"]}]}],"UpdateTime":1648556261,"CheckCode":4009496266}`
							tmp = strings.Replace(tmp, "\n", "", -1)
							cm.Data[vnpuutil.VNPCMDataKey] = tmp
							return &cm, nil
						})

				}, cacheFunAfter: func() {
					if tmpPatche != nil {
						tmpPatche.Reset()
					}
					if tmpPatche1 != nil {
						tmpPatche1.Reset()
					}
				}},
			wantErr: errors.New(util.SegmentSetFalse),
		},
	}
	return testCases
}

// TestPreHandleVNPU test PreHandleVNPU function
func TestPreHandleVNPU(t *testing.T) {
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
