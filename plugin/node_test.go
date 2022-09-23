/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package plugin is using for HuaWei Ascend pin affinity schedule frame.

*/
package plugin

import (
	"errors"
	"reflect"
	"testing"

	"k8s.io/api/core/v1"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
)

type nodeFields struct {
	Name       string
	Capability map[v1.ResourceName]float64
	Allocate   map[v1.ResourceName]float64
	Idle       map[v1.ResourceName]float64
	Annotation map[string]string
	Label      map[string]string
}

type checkNPUResourceStableArgs struct {
	vcJob SchedulerJob
}

type checkNPUResourceStableTest struct {
	name    string
	fields  nodeFields
	args    checkNPUResourceStableArgs
	wantErr bool
}

func buildVCheckNPUResourceStableTest() []checkNPUResourceStableTest {
	tJob := SchedulerJob{handler: New(testPluginName)}
	tests := []checkNPUResourceStableTest{
		{
			name:    "01-CheckNPUResourceStable no annotation test",
			fields:  nodeFields{Name: "haha", Idle: map[v1.ResourceName]float64{testCardName: 1}, Annotation: nil},
			args:    checkNPUResourceStableArgs{vcJob: tJob},
			wantErr: true,
		},
		{
			name: "02-CheckNPUResourceStable ok test.",
			fields: nodeFields{Name: "haha", Idle: map[v1.ResourceName]float64{testCardName: 1},
				Annotation: map[string]string{testCardName: "haha"}},
			args:    checkNPUResourceStableArgs{vcJob: tJob},
			wantErr: true,
		},
	}
	return tests
}

func TestCheckNPUResourceStable(t *testing.T) {
	tests := buildVCheckNPUResourceStableTest()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := NPUNode{
				Name:       tt.fields.Name,
				Capability: tt.fields.Capability,
				Allocate:   tt.fields.Allocate,
				Idle:       tt.fields.Idle,
				Annotation: tt.fields.Annotation,
				Label:      tt.fields.Label,
			}
			if err := n.CheckNPUResourceStable(tt.args.vcJob); (err != nil) != tt.wantErr {
				t.Errorf("CheckNPUResourceStable() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type nodePredicateArgs struct {
	taskInfo *api.TaskInfo
	nodeInfo *api.NodeInfo
}

type nodePredicateTest struct {
	name    string
	fields  fields
	args    nodePredicateArgs
	wantErr bool
}

func buildNodePredicateTest() []nodePredicateTest {
	tTasks := test.FakeNormalTestTasks(1)
	tNode := test.FakeNormalTestNode("testNode")
	tests := []nodePredicateTest{
		{
			name:    "01-NodePredicate nil test.",
			fields:  fields{},
			args:    nodePredicateArgs{taskInfo: nil, nodeInfo: nil},
			wantErr: true,
		},
		{
			name: "02-NodePredicate job not in test.",
			fields: fields{ScheduleEnv: ScheduleEnv{
				Jobs: map[api.JobID]SchedulerJob{"haha": {}}}},
			args:    nodePredicateArgs{taskInfo: tTasks[0], nodeInfo: tNode},
			wantErr: false,
		},
		{
			name: "03-NodePredicate node not in test.",
			fields: fields{ScheduleEnv: ScheduleEnv{
				Jobs:  map[api.JobID]SchedulerJob{tTasks[0].Job: {}},
				Nodes: map[string]NPUNode{"haha": {}}}},
			args:    nodePredicateArgs{taskInfo: tTasks[0], nodeInfo: tNode},
			wantErr: false,
		},
	}
	return tests
}

func TestSNodePredicate(t *testing.T) {
	tests := buildNodePredicateTest()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sHandle := &ScheduleHandler{
				NPUPlugins:  tt.fields.NPUPlugins,
				ScheduleEnv: tt.fields.ScheduleEnv,
			}
			if err := sHandle.NodePredicate(tt.args.taskInfo, tt.args.nodeInfo); (err != nil) != tt.wantErr {
				t.Errorf("NodePredicate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type checkNodeDeviceInfoTestCase struct {
	name    string
	dvInfo  *NodeDeviceInfoWithDevPlugin
	wantErr error
}

func buildCheckNodeDeviceInfoTestCases() []checkNodeDeviceInfoTestCase {
	const fakeCheckCode = "fakeCheckCode"
	deviceInfo := NodeDeviceInfo{
		DeviceList: map[string]string{"huawei.com/Ascend910": "Ascend910-0,Ascend910-1",
			"huawei.com/Ascend910-NetworkUnhealthy": "",
			"huawei.com/Ascend910-Unhealthy":        ""},
		UpdateTime: 0,
	}
	checkCode := MakeDataHash(deviceInfo)

	return []checkNodeDeviceInfoTestCase{
		{
			name: "01-CheckNodeDeviceInfo return nil when deviceInfo checkCode is match",
			dvInfo: &NodeDeviceInfoWithDevPlugin{
				DeviceInfo: deviceInfo,
				CheckCode:  checkCode,
			},
			wantErr: nil,
		},
		{
			name:    "02-CheckNodeDeviceInfo return err when deviceInfo is nil",
			dvInfo:  nil,
			wantErr: errors.New("nil parameters"),
		},
		{
			name: "03-CheckNodeDeviceInfo return err when checkcode is empty",
			dvInfo: &NodeDeviceInfoWithDevPlugin{
				DeviceInfo: deviceInfo,
				CheckCode:  "",
			},
			wantErr: errors.New("checkCode is empty"),
		},
		{
			name: "04-CheckNodeDeviceInfo return err when checkcode is not match",
			dvInfo: &NodeDeviceInfoWithDevPlugin{
				DeviceInfo: deviceInfo,
				CheckCode:  fakeCheckCode,
			},
			wantErr: errors.New("checkCode is not match"),
		},
	}
}

// TestCheckNodeDeviceInfo
func TestCheckNodeDeviceInfo(t *testing.T) {
	testCases := buildCheckNodeDeviceInfoTestCases()
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if err := checkNodeDeviceInfo(tt.dvInfo); !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("checkNodeDeviceInfo() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
