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

func (tp *ascend310P) GetVNPUTemplate() {
	tp.vHandle.VT = vnpu.VTemplate{
		Data: tp.FrameAttr.VJobTemplate[plugin.Ascend310P],
	}
}

func (tp *ascend310P) InitVNPU() {
	tp.vHandle = &vnpu.VNPU{
		DynamicVNPU: vnpu.DynamicVNPU{
			Cache: make(map[string][]string, util.MapInitNum),
		},
	}
}

func (tp *ascend310P) checkStVJobReq() error {
	for _, vT := range tp.Tasks {
		if _, ok := tp.vHandle.VT.Data[vT.ReqNPUName]; !ok {
			return fmt.Errorf("%s req %s not in template", vT.Name, vT.ReqNPUName)
		}
		if vT.ReqNPUNum != 1 {
			return fmt.Errorf("%s req %d not 1", vT.Name, vT.ReqNPUNum)
		}
	}
	return nil
}

func (tp *ascend310P) validStVNPUJob() *api.ValidateResult {
	if reqErr := tp.checkStVJobReq(); reqErr != nil {
		return &api.ValidateResult{Pass: false, Reason: reqErr.Error(), Message: reqErr.Error()}
	}
	return nil
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
			if vTask.Status == util.TaskStatusFailed {
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
		tmp := vnpu.GetSegmentFailureTaskIDs(ssn, ns)
		if len(tmp) == 0 {
			continue
		}
		tIDs = append(tIDs, tmp...)
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
	if len(vT) == 0 {
		return nil
	}
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
	tp.GetVNPUTemplate()
	return tp.preStartDyVNPU(ssn)
}
