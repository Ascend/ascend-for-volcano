/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package rescheduling is using for HuaWei Ascend pin fault rescheduling.

*/
package rescheduling

import (
	"github.com/agiledragon/gomonkey/v2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"reflect"
	"testing"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

const emptyCheckCode = "6d0c413224f9882c8342fe8bed0389875231fbb1af67a12629f0617257b533d4"

type DealReSchedulerConfigmapCreateEmptyReCMFields struct {
	CMName      string
	CMNameSpace string
	CMData      map[string]string
}

type DealReSchedulerConfigmapCreateEmptyReCMArgs struct {
	kubeClient      kubernetes.Interface
	jobType         string
	cacheFuncBefore func()
	cacheFuncAfter  func()
}

type DealReSchedulerConfigmapCreateEmptyReCMTests struct {
	name    string
	fields  DealReSchedulerConfigmapCreateEmptyReCMFields
	args    DealReSchedulerConfigmapCreateEmptyReCMArgs
	want    map[string]string
	wantErr bool
}

func buildTestDealReSchedulerConfigmapCreateEmptyReCMTests() []DealReSchedulerConfigmapCreateEmptyReCMTests {
	var tmpPatche *gomonkey.Patches
	resultMap := map[string]string{
		CmCheckCode:           emptyCheckCode,
		CmFaultNodeKind:       "",
		CmFaultJob910x8Kind:   "",
		CmNodeHeartbeatKind:   "",
		CmNodeRankTimeMapKind: "",
	}
	test1 := DealReSchedulerConfigmapCreateEmptyReCMTests{
		name: "01-DealReSchedulerConfigmapCreateEmptyReCM()-success",
		fields: DealReSchedulerConfigmapCreateEmptyReCMFields{
			CMName:      CmName,
			CMNameSpace: CmNameSpace,
			CMData:      nil,
		},
		args: DealReSchedulerConfigmapCreateEmptyReCMArgs{
			kubeClient: nil,
			jobType:    CmFaultJob910x8Kind,
			cacheFuncBefore: func() {
				tmpPatche = gomonkey.ApplyFunc(util.CreateOrUpdateConfigMap, func(_ kubernetes.Interface,
					_ *v1.ConfigMap, _, _ string) error {
					return nil
				})
			},
			cacheFuncAfter: func() {
				if tmpPatche != nil {
					tmpPatche.Reset()
				}
			},
		},
		want:    resultMap,
		wantErr: false,
	}
	tests := []DealReSchedulerConfigmapCreateEmptyReCMTests{
		test1,
	}
	return tests
}

// TestDealReSchedulerConfigmapCreateEmptyReCM test for creating empty configmap
func TestDealReSchedulerConfigmapCreateEmptyReCM(t *testing.T) {
	tests := buildTestDealReSchedulerConfigmapCreateEmptyReCMTests()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.cacheFuncBefore()
			dealCM := &DealReSchedulerConfigmap{
				CMName:      tt.fields.CMName,
				CMNameSpace: tt.fields.CMNameSpace,
				CMData:      tt.fields.CMData,
			}
			got, err := dealCM.createEmptyReCM(tt.args.kubeClient, tt.args.jobType)
			if (err != nil) != tt.wantErr {
				t.Errorf("createEmptyReCM() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createEmptyReCM() got = %v, want %v", got, tt.want)
			}
			tt.args.cacheFuncAfter()
		})
	}
}
