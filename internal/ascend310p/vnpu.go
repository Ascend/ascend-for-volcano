/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package ascend310p is using for HuaWei 310P Ascend pin affinity schedule.
*/
package ascend310p

import (
	"fmt"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/vnpu"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

func (tp *ascend310P) GetTemplate() {
	tp.vHandle.VT = vnpu.VTemplate{
		Data: map[string]util.VResource{
			"vir01":          {Aicore: 1, Aicpu: 1, DVPP: "null"},
			"vir02":          {Aicore: 2, Aicpu: 2, DVPP: "null"},
			"vir02_1c":       {Aicore: 2, Aicpu: 1, DVPP: "null"},
			"vir04":          {Aicore: 4, Aicpu: 4, DVPP: "null"},
			"vir04_3c":       {Aicore: 4, Aicpu: 3, DVPP: "null"},
			"vir04_3c_ndvpp": {Aicore: 4, Aicpu: 3, DVPP: "no"},
			"vir04_4c_dvpp":  {Aicore: 4, Aicpu: 4, DVPP: "yes"},
		},
	}
}

func (tp *ascend310P) InitVNPU() {
	tp.vHandle = &vnpu.VNPU{
		DynamicVNPU: vnpu.DynamicVNPU{
			Cache: make(map[string][]string, util.MapInitNum),
		},
	}
	tp.GetTemplate()
}

func (tp *ascend310P) checkDyVJobReq() error {
	for _, vT := range tp.Tasks {
		if vT.ReqNPUNum == 1 || vT.ReqNPUNum == util.NPUIndex2 || vT.ReqNPUNum == util.NPUIndex4 ||
			vT.ReqNPUNum%util.NPUIndex8 == 0 {
			continue
		}
		return fmt.Errorf("%s req err %d", vT.Name, vT.ReqNPUNum)
	}
	return nil
}

func (tp *ascend310P) validDyVNPUJob() *api.ValidateResult {
	if reqErr := tp.checkDyVJobReq(); reqErr != nil {
		return &api.ValidateResult{Pass: false, Reason: reqErr.Error(), Message: reqErr.Error()}
	}
	return nil
}

func (tp *ascend310P) getAllDyJobs() map[api.JobID]plugin.SchedulerJob {
	jobMap := make(map[api.JobID]plugin.SchedulerJob, util.MapInitNum)
	for jobID, vJob := range tp.Jobs {
		if vJob.VJob.Type == util.JobTypeDyCut {
			jobMap[jobID] = vJob
		}
	}
	return jobMap
}

func getFailedDyTasksFromJobs(vJobs map[api.JobID]plugin.SchedulerJob) map[api.TaskID]util.NPUTask {
	vTasks := make(map[api.TaskID]util.NPUTask, util.MapInitNum)
	for _, vJob := range vJobs {
		for tID, vTask := range vJob.Tasks {
			if vTask.Status == util.TASK_STAUS_FAILD {
				vTasks[tID] = vTask
			}
		}
	}
	return vTasks
}

func getDyFailedNamespaces(vT map[api.TaskID]util.NPUTask) map[string]struct{} {
	nsMap := make(map[string]struct{}, util.MapInitNum)
	for _, nT := range vT {
		nsMap[nT.NameSpace] = struct{}{}
	}
	return nsMap
}

func getAllDyFailedTasks(ssn *framework.Session, nsMap map[string]struct{}) []api.TaskID {
	tIDs := make([]api.TaskID, util.MapInitNum)
	for ns := range nsMap {
		tIDs = append(tIDs, vnpu.GetSegmentFailureTaskIDs(ssn, ns)...)
	}
	return tIDs
}

func getDyFailedTaskIDsInFaileds(allIDS []api.TaskID, vT map[api.TaskID]util.NPUTask) []api.TaskID {
	tIDs := make([]api.TaskID, util.MapInitNum)
	for _, tID := range allIDS {
		if _, ok := vT[tID]; !ok {
			klog.V(util.LogErrorLev).Infof("getDyFailedTaskIDsInFaileds taskID(%s) not in tasks.", tID)
			continue
		}
		tIDs = append(tIDs, tID)
	}
	return tIDs
}

func getDyFailedTasksFromFailed(ssn *framework.Session, vT map[api.TaskID]util.NPUTask) []api.TaskID {
	nsMap := getDyFailedNamespaces(vT)

	allIDS := getAllDyFailedTasks(ssn, nsMap)

	return getDyFailedTaskIDsInFaileds(allIDS, vT)
}

func (tp *ascend310P) getRestartDyTasksFromJobs(vJobs map[api.JobID]plugin.SchedulerJob,
	ssn *framework.Session) []util.NPUTask {
	vTasks := getFailedDyTasksFromJobs(vJobs)
	fTIDs := getDyFailedTasksFromFailed(ssn, vTasks)
	if len(fTIDs) == 0 {
		return nil
	}
	nSlice := make([]util.NPUTask, util.MapInitNum)
	for _, tID := range fTIDs {
		vT, ok := vTasks[tID]
		if !ok {
			klog.V(util.LogErrorLev).Infof("getRestartDyTasksFromJobs taskID(%s) not found.", tID)
			continue
		}
		nSlice = append(nSlice, vT)
	}
	return nSlice
}

func (tp *ascend310P) getAllNeedRestartDyTasks(ssn *framework.Session) []util.NPUTask {
	vJobs := tp.getAllDyJobs()
	if len(vJobs) == 0 {
		return nil
	}
	return tp.getRestartDyTasksFromJobs(vJobs, ssn)
}

func (tp *ascend310P) preStartDyVNPU(ssn *framework.Session) error {
	nTasks := tp.getAllNeedRestartDyTasks(ssn)
	if len(nTasks) == 0 {
		return nil
	}
	for _, nT := range nTasks {
		if delErr := nT.ForceDeletePodByTaskInf(ssn); delErr != nil {
			klog.V(util.LogErrorLev).Infof("ForceDeletePodByTaskInf %s: %s.", nT.Name, delErr)
			continue
		}
	}
	return nil
}

func (tp *ascend310P) preStartVNPU(ssn *framework.Session) error {
	return tp.preStartDyVNPU(ssn)
}
