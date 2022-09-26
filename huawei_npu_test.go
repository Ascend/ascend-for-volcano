/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package main is using for HuaWei Ascend pin affinity schedule.

*/
package main

import (
	"testing"

	"volcano.sh/volcano/pkg/scheduler/framework"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
)

type handlerStartTest struct {
	name string
	want *plugin.ScheduleHandler
}

func buildTestHandlerStartTestCases() []handlerStartTest {
	testCases := []handlerStartTest{
		{
			name: "HandlerStart ok test",
			want: &plugin.ScheduleHandler{},
		},
	}
	return testCases
}

func TestHandlerStart(t *testing.T) {
	tests := buildTestHandlerStartTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HandlerStart(); got == nil {
				t.Errorf("HandlerStart() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	type args struct {
		arguments framework.Arguments
	}
	tests := []struct {
		name string
		args args
		want framework.Plugin
	}{
		{
			name: "New ok test",
			args: args{arguments: framework.Arguments{PluginName: "haha"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.arguments); got == nil {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

type fields struct {
	Scheduler *plugin.ScheduleHandler
	Arguments framework.Arguments
}

type args struct {
	ssn            *framework.Session
	cacheFunBefore func()
	cacheFunAfter  func()
}

type onSessionOpenTest struct {
	name   string
	fields fields
	args   args
}

func buildOnSessionOpenTestCases() []onSessionOpenTest {
	tests := []onSessionOpenTest{
		{
			name:   "OnSessionOpen test ssn nil ok",
			fields: fields{Scheduler: HandlerStart()},
			args:   args{ssn: nil, cacheFunBefore: func() {}, cacheFunAfter: func() {}},
		},
	}
	return tests
}

func TestOnSessionOpen(t *testing.T) {
	tests := buildOnSessionOpenTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp := &huaweiNPUPlugin{
				Scheduler: tt.fields.Scheduler,
				Arguments: tt.fields.Arguments,
			}
			tt.args.cacheFunBefore()
			tp.OnSessionOpen(tt.args.ssn)
			tt.args.cacheFunAfter()
		})
	}
}

type onSessionCloseTest struct {
	name   string
	fields fields
	args   args
}

func buildOnSessionCloseTestCases() []onSessionCloseTest {
	testSsn := test.FakeNormalSSN()
	tests := []onSessionCloseTest{
		{
			name:   "OnSessionCloseTestCases test ok",
			fields: fields{Scheduler: HandlerStart()},
			args:   args{ssn: testSsn},
		},
	}
	return tests
}

func TestOnSessionClose(t *testing.T) {
	tests := buildOnSessionCloseTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp := &huaweiNPUPlugin{
				Scheduler: tt.fields.Scheduler,
				Arguments: tt.fields.Arguments,
			}
			tp.OnSessionClose(tt.args.ssn)
		})
	}
}
