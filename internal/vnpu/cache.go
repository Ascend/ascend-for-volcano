/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package vnpu is using for HuaWei Ascend pin fault rescheduling.

*/
package vnpu

import (
	"encoding/json"
	"fmt"
	"sort"

	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// InitVConfigMap init VConfigMap
func (vCache *VCache) InitVConfigMap(client kubernetes.Interface) {
	vCache.vConfigMap.ReadCMFromKubernetes(client)
}

// InitVJobs init VJobs
func (vCache *VCache) InitVJobs(env *plugin.ScheduleEnv, divideKinds []string) error {
	klog.V(util.LogInfoLev).Info("enter InitVJobs...")
	if err := vCache.ReadVJobsFromVCM(); err != nil {
		klog.V(util.LogInfoLev).Info("InitVJobs from configmap failed")
	}
	if err := vCache.initVJobsFromSession(env, divideKinds); err != nil {
		klog.V(util.LogErrorLev).Info("InitVJobs from session failed")
		return err
	}
	klog.V(util.LogInfoLev).Info("InitVJobs success")
	return nil
}

// ReadVJobsFromVCM read vJobs from cache
func (vCache *VCache) ReadVJobsFromVCM() error {
	cacheCMData, ok := vCache.vConfigMap.Data[VNPUCMDataKey]
	if !ok {
		klog.V(util.LogErrorLev).Infof("ReadVJobsFromVCM: no %s data", VNPUCMDataKey)
		return fmt.Errorf("%s not in configmap", VNPUCMDataKey)
	}
	vJobs, err := vCache.getVJobsFromCM(cacheCMData)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("ReadVJobsFromVCM: %s", err.Error())
		return err
	}
	vCache.setVJobs(vJobs)
	return nil
}

func (vCache *VCache) getVJobsFromCM(buffer string) (map[api.JobID]VJob, error) {
	vJobs := make(map[api.JobID]VJob, util.MapInitNum)
	if unmarshalErr := json.Unmarshal([]byte(buffer), &vJobs); unmarshalErr != nil {
		klog.V(util.LogErrorLev).Infof("Unmarshal FaultNodes from cache failed")
		return nil, fmt.Errorf("faultNodes convert from CM error: %#v", unmarshalErr)
	}
	return vJobs, nil
}

// initVJobsFromSession create new vJobs from session and updates old vJobs
func (vCache *VCache) initVJobsFromSession(env *plugin.ScheduleEnv, divideKinds []string) error {
	for jobUID, schedulerJob := range env.Jobs {
		vJob, ok := vCache.vJobs[jobUID]
		if ok {
			klog.V(util.LogInfoLev).Infof("job %s is already in cache, skip initialisation", vJob.jobUID)
			continue
		}
		vJob = vCache.initVJobFromSchedulerJob(schedulerJob)
		if err := vJob.initVJobStatus(divideKinds); err != nil {
			klog.V(util.LogInfoLev).Infof("job %s is not vnpu job, do not put into vCache", vJob.jobUID)
			continue
		}
		vCache.addOrUpdateVJobToCache(vJob)
	}
	return nil
}

func (vCache *VCache) initVJobFromSchedulerJob(schedulerJob plugin.SchedulerJob) VJob {
	vJob := VJob{
		jobUID:        schedulerJob.JobName,
		jobStatus:     "", // need to be updated later
		reqVNPUType:   schedulerJob.ReqNPUName,
		reqNodeName:   "",
		reqCardName:   "",
		taskNum:       len(schedulerJob.Tasks),
		allocCardName: "",
		allocFlag:     false,
		resourceReq:   VResource{},
		createTime:    schedulerJob.CreateTime,
		allocTime:     0,
	}
	return vJob
}

func (vCache *VCache) getJobTaskStatus(ssnJob *api.JobInfo) []v1.PodPhase {
	var jobTaskStatus []v1.PodPhase
	for _, task := range ssnJob.Tasks {
		jobTaskStatus = append(jobTaskStatus, task.Pod.Status.Phase)
	}
	return jobTaskStatus
}

// WriteVJobsToVCM update VJobs into vCache.vConfigMap.Data, and will be later updated to k8s by vConfigMap's method
func (vCache *VCache) WriteVJobsToVCM() {
	vCache.vConfigMap.Data = vCache.getVCacheCMData()
}

func (vCache *VCache) getVCacheCMData() map[string]string {
	tmp, err := vCache.vConfigMap.MarshalCacheDataToString(vCache.vJobs)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("marshalCacheDataToString err: %v.", err)
		return nil
	}
	cacheBuffer := make(map[string]string, util.NPUIndex3)
	cacheBuffer[VNPUCMDataKey] = tmp
	return cacheBuffer
}

// GetAndSortNotPreAllocVJobsFromCache get and sort NotPreAlloc VJobs from cache
func (vCache *VCache) GetAndSortNotPreAllocVJobsFromCache(env *plugin.ScheduleEnv) []VJob {
	klog.V(util.LogInfoLev).Infof("GetAndSort jobs of %s in cache", VJobStatusNotPreSegmented)
	vJobsToBePreAlloc := vCache.getNeedAllocVJobsFromCache(env)
	vJobsToBePreAllocSorted := vCache.sortNeedAllocVJobsByAllocTime(vJobsToBePreAlloc)
	return vJobsToBePreAllocSorted
}

func (vCache *VCache) getNeedAllocVJobsFromCache(env *plugin.ScheduleEnv) []VJob {
	var vJobsToBePreAlloc []VJob
	for jobUID, vJob := range vCache.vJobs {
		_, ok := env.Jobs[jobUID]
		if !ok || vJob.jobStatus != VJobStatusNotPreSegmented {
			klog.V(util.LogDebugLev).Infof("vJob %s not in session, so skip preAlloc check", jobUID)
			continue
		}
		vJobsToBePreAlloc = append(vJobsToBePreAlloc, vJob)
	}
	return vJobsToBePreAlloc
}

func (vCache *VCache) sortNeedAllocVJobsByAllocTime(vJobs []VJob) []VJob {
	tempVJobs := VJobList(vJobs)
	sort.Sort(tempVJobs)
	return tempVJobs
}

func (vCache *VCache) getNeedDeleteJobsFromCache() []VJob {
	var vJobToBeDeleted []VJob
	for _, vJob := range vCache.vJobs {
		if vJob.jobStatus == VJobStatusDestroying {
			vJobToBeDeleted = append(vJobToBeDeleted, vJob)
		}
	}
	return vJobToBeDeleted
}

func (vCache *VCache) addOrUpdateVJobToCache(vJob VJob) {
	vCache.vJobs[vJob.jobUID] = vJob
}

func (vCache *VCache) deleteVJobFromCache(vJobUID api.JobID) {
	delete(vCache.vJobs, vJobUID)
}

func (vCache *VCache) setVJobs(vJobs map[api.JobID]VJob) {
	vCache.vJobs = vJobs
}
