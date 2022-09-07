/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package main is using for HuaWei Ascend pin affinity schedule.

*/
package main

import (
	"github.com/agiledragon/gomonkey/v2"
	"reflect"
	"testing"
	"volcano.sh/volcano/pkg/scheduler/api"
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
	testSsn := test.FakeNormalSSN()
	var tmpPatche *gomonkey.Patches
	var tmpPatche2 *gomonkey.Patches
	var tmpPatche3 *gomonkey.Patches
	var tmpPatche4 *gomonkey.Patches
	tests := []onSessionOpenTest{
		{
			name:   "OnSessionOpen test ssn nil ok",
			fields: fields{Scheduler: HandlerStart()},
			args:   args{ssn: nil, cacheFunBefore: func() {}, cacheFunAfter: func() {}},
		},
		{
			name:   "OnSessionOpen test ssn ok",
			fields: fields{Scheduler: HandlerStart()},
			args: args{ssn: testSsn, cacheFunBefore: func() {
				tmpPatche = gomonkey.ApplyMethod(reflect.TypeOf(new(framework.Session)), "AddJobValidFn",
					func(_ *framework.Session, _ string, _ api.ValidateExFn) {
						return
					})
				tmpPatche2 = gomonkey.ApplyMethod(reflect.TypeOf(new(framework.Session)), "AddPredicateFn",
					func(_ *framework.Session, _ string, _ api.PredicateFn) {
						return
					})
				tmpPatche3 = gomonkey.ApplyMethod(reflect.TypeOf(new(framework.Session)), "AddBatchNodeOrderFn",
					func(_ *framework.Session, _ string, _ api.BatchNodeOrderFn) {
						return
					})
				tmpPatche4 = gomonkey.ApplyMethod(reflect.TypeOf(new(framework.Session)), "AddEventHandler",
					func(_ *framework.Session, _ *framework.EventHandler) {
						return
					})
			}, cacheFunAfter: func() {
				tmpPatche.Reset()
				tmpPatche2.Reset()
				tmpPatche3.Reset()
				tmpPatche4.Reset()
			}},
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
