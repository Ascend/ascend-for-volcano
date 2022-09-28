/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package test is using for HuaWei Ascend testing.

*/
package test

import (
	"reflect"

	"github.com/agiledragon/gomonkey/v2"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/rescheduling"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

// PatchNew go monkey patch
func PatchNew() *gomonkey.Patches {
	return gomonkey.ApplyFunc(rescheduling.New, func(_ *plugin.ScheduleEnv,
		_ string) *rescheduling.ReScheduler {
		return &rescheduling.ReScheduler{GraceDeleteTime: rescheduling.DefaultGraceOverTime}
	})
}

// PatchNewComRes go monkey patch
func PatchNewComRes() *gomonkey.Patches {
	return gomonkey.ApplyMethod(reflect.TypeOf(&rescheduling.ReScheduler{}), "NewCommonReScheduler",
		func(_ *rescheduling.ReScheduler, _ string) { return })
}

// PatchSynNode go monkey patch
func PatchSynNode() *gomonkey.Patches {
	return gomonkey.ApplyMethod(reflect.TypeOf(&rescheduling.ReScheduler{}), "SynCacheFaultNodeWithSession",
		func(_ *rescheduling.ReScheduler, _ string) { return })
}

// PatchAddNode go monkey patch
func PatchAddNode() *gomonkey.Patches {
	return gomonkey.ApplyMethod(reflect.TypeOf(&rescheduling.ReScheduler{}),
		"AddFaultNodeWithSession",
		func(_ *rescheduling.ReScheduler, _ string) { return })
}

// PatchSynJob go monkey patch
func PatchSynJob() *gomonkey.Patches {
	return gomonkey.ApplyMethod(reflect.TypeOf(&rescheduling.ReScheduler{}),
		"SynCacheFaultJobWithSession",
		func(_ *rescheduling.ReScheduler, _ *framework.Session, _, _ string) { return })
}

// PatchForce go monkey patch
func PatchForce() *gomonkey.Patches {
	return gomonkey.ApplyMethod(reflect.TypeOf(&rescheduling.ReScheduler{}),
		"SynCacheFaultJobWithSession",
		func(_ *rescheduling.ReScheduler, _ *framework.Session, _, _ string) { return })
}

// PatchGetRun go monkey patch
func PatchGetRun() *gomonkey.Patches {
	return gomonkey.ApplyMethod(reflect.TypeOf(&rescheduling.ReScheduler{}),
		"GetRunningJobs",
		func(_ *rescheduling.ReScheduler, _ *framework.Session, _, _ string) (map[api.JobID]*api.JobInfo,
			error) {
			return map[api.JobID]*api.JobInfo{"job1": &api.JobInfo{}}, nil
		})
}

// PatchAddJob go monkey patch
func PatchAddJob() *gomonkey.Patches {
	return gomonkey.ApplyMethod(reflect.TypeOf(&rescheduling.ReScheduler{}),
		"AddFaultJobWithSession",
		func(_ *rescheduling.ReScheduler, _ map[api.JobID]*api.JobInfo, _, _ string) error { return nil })
}

// PatchRestart go monkey patch
func PatchRestart() *gomonkey.Patches {
	return gomonkey.ApplyMethod(reflect.TypeOf(&rescheduling.ReScheduler{}),
		"RestartFaultJobs",
		func(_ *rescheduling.ReScheduler, _ *framework.Session) error { return nil })
}

// PatchReset go monkey patch reset
func PatchReset(patch *gomonkey.Patches) {
	if patch != nil {
		patch.Reset()
	}
}
