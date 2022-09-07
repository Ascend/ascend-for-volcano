/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package test is using for HuaWei Ascend pin fault rescheduling.

*/
package test

import (
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// FakeSchedulerJobAttrByJob fake scheduler attr by job
func FakeSchedulerJobAttrByJob(job *api.JobInfo) util.SchedulerJobAttr {
	attr := util.SchedulerJobAttr{
		ComJob: util.ComJob{
			JobName:   job.UID,
			NameSpace: job.Namespace,
			Selector:  nil,
			Label:     nil,
		},
	}
	name, num, err := plugin.GetVCJobReqNPUTypeFromJobInfo(job)
	if err != nil {
		return attr
	}
	NPUJob := &util.NPUJob{
		ReqNPUName: name,
		ReqNPUNum:  num,
		Tasks:      plugin.GetJobNPUTasks(job),
	}
	attr.NPUJob = NPUJob
	return attr
}
