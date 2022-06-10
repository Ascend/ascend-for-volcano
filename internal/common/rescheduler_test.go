/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

// Package common is using for HuaWei common infer Ascend pin affinity schedule.
package common

import (
	"errors"
	"reflect"
	"testing"
	"volcano.sh/apis/pkg/apis/scheduling"

	"github.com/agiledragon/gomonkey/v2"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/vnpu/vnpuutil"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
)

type preHandleFaultNPUFnArgs struct {
	ssn            *framework.Session
	cacheFunBefore func()
	cacheFunAfter  func()
}

type preHandleFaultNPUFnTests struct {
	name    string
	fields  ReScheduler
	args    preHandleFaultNPUFnArgs
	wantErr error
}

func buildPreHandleFaultNPUFnTestCases() []preHandleFaultNPUFnTests {
	jobOne := test.FakeNormalTestJob("test-1", 1)
	test.SetFakeJobRequestSource(jobOne, vnpuutil.NPU310PCardName, 1*util.NPUHex)
	test.SetFakeNPUJobUseAnnotationInPod(jobOne, vnpuutil.NPU310PCardName, "Ascend310P-1")
	test.SetFakeNPUJobLabelInTask(jobOne, faultSchedulingLabel, onFaultSchedulingLabel)
	test.SetTestJobPodGroupStatus(jobOne, scheduling.PodGroupRunning)
	nodeOne := test.FakeNormalTestNode("node-1")
	test.SetTestNPUNodeAnnotation(nodeOne, util.Fault310PNPU, "Ascend310P-1")
	test.SetTestNPUNodeAnnotation(nodeOne, vnpuutil.NPU310PCardName, "huawei.com/Ascend310P-2")
	ssnTest := test.FakeNormalSSN()
	test.SetFakeNPUJobUseNodeNameInTask(jobOne, nodeOne.Name)
	test.AddJobIntoFakeSSN(ssnTest, jobOne)
	test.AddNodeIntoFakeSSN(ssnTest, nodeOne)
	scheduler := Scheduler{
		PluginName: vnpuutil.PluginName, AnnoName: vnpuutil.NPU310PCardName, AnnoPreVal: util.CommCardPreName,
		DefaultJobSchedulerConfig: nil,
	}
	var tmpPatch *gomonkey.Patches
	testCases := []preHandleFaultNPUFnTests{
		{
			name: "01-VJobRunHandle() success test.",
			args: preHandleFaultNPUFnArgs{
				ssn: ssnTest, cacheFunBefore: func() {
					tmpPatch = gomonkey.ApplyMethod(reflect.TypeOf(new(framework.Session)), "Evict",
						func(_ *framework.Session, _ *api.TaskInfo, _ string) error { return errors.New("haha") })
				}, cacheFunAfter: func() {
					if tmpPatch != nil {
						tmpPatch.Reset()
					}
				}},
			fields: ReScheduler{AnnoUnHealthy: util.Fault310PNPU, AnnoName: scheduler.AnnoName,
				IsMyJob: scheduler.IsMyJob},
			wantErr: errors.New("haha"),
		},
	}
	return testCases
}

// TestPreHandleFaultNPUFn test PreHandleFaultNPUFn
func TestPreHandleFaultNPUFn(t *testing.T) {
	tests := buildPreHandleFaultNPUFnTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := &ReScheduler{
				IsMyJob:       tt.fields.IsMyJob,
				AnnoUnHealthy: tt.fields.AnnoUnHealthy,
				AnnoName:      tt.fields.AnnoName,
			}
			tt.args.cacheFunBefore()
			err := re.PreHandleFaultNPUFn(tt.args.ssn)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("PreHandleFaultNPUFn() error = %v, wantErr %v", err, tt.wantErr)
			}
			tt.args.cacheFunAfter()
		})
	}
}
