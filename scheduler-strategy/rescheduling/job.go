/*
Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/*

Package rescheduling is using for HuaWei Ascend pin fault rescheduling.

*/
package rescheduling

import (
	"encoding/json"
	"errors"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	"strings"
	time2 "time"
	"volcano.sh/apis/pkg/apis/scheduling"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"
)

func convertToReSchedulerTasksMapFromCache(jobData interface{}) (map[api.JobID]ReSchedulerTasks, error) {
	reSchedulerJob, reOk := jobData.(map[api.JobID]ReSchedulerTasks)
	if !reOk {
		msg := fmt.Errorf("assert %v to ReSchedulerTasks failed", jobData)
		klog.V(logErrorLev).Infof("%v.", msg)
		return nil, msg
	}

	if len(reSchedulerJob) == 0 {
		msg := fmt.Errorf("ReSchedulerTasks is nil")
		klog.V(logDebugLev).Infof("convertToReSchedulerTasksMapFromCache: %v.", msg)
		return nil, msg
	}

	return reSchedulerJob, nil
}

func convertToRankIdsMapFromCache(jobData interface{}) (map[api.JobID]FaultRankIDRecordJobCMData, error) {
	rankIds, reOk := jobData.(map[api.JobID]FaultRankIDRecordJobCMData)
	if !reOk {
		msg := fmt.Errorf("assert %v to FaultRankIDRecordJobCMData failed", jobData)
		klog.V(logErrorLev).Infof("%v.", msg)
		return nil, msg
	}

	if len(rankIds) == 0 {
		msg := fmt.Errorf("RankIds is nil")
		klog.V(logDebugLev).Infof("convertToRankIdsMapFromCache: %v.", msg)
		return nil, msg
	}

	return rankIds, nil
}

func getCMJobWriteData(ssn *framework.Session, jobData interface{}) (string, error) {
	reSchedulerJob, covErr := convertToReSchedulerTasksMapFromCache(jobData)
	if covErr != nil {
		klog.V(logDebugLev).Infof("getCMJobWriteData: %v.", covErr)
		return "", covErr
	}

	var data = make(map[string]string, 1)
	for jobID, faultNPUJob := range reSchedulerJob {
		buffer, err := json.Marshal(faultNPUJob)
		if err != nil {
			klog.V(logErrorLev).Infof("getCMJobWriteData  %+v err: %v.", faultNPUJob, err)
			return "", err
		}

		job, ok := ssn.Jobs[jobID]
		if !ok {
			return "", fmt.Errorf("writeFaultNPUJobsToCM ssn not has %v", jobID)
		}

		data[job.Namespace+"_"+job.Name] = string(buffer)
	}

	return marshalCacheDataToString(data)
}

func getCMRankIdsWriteData(jobData interface{}) (string, error) {
	data, err := marshalCacheDataToString(jobData)
	if err != nil {
		return "", err
	}
	dataTemp := strings.ReplaceAll(data, "/", "_")
	return dataTemp, nil
}

// Delete expired job data.
func synReSchedulerJobCache(ssn *framework.Session, tmpValue interface{}) error {
	jobMap, assertOk := tmpValue.(map[api.JobID]ReSchedulerTasks)
	if !assertOk {
		msg := fmt.Errorf("convert %v to map[api.JobID]ReSchedulerTasks failed", tmpValue)
		klog.V(logErrorLev).Infof("synReSchedulerJobCache %v.", msg)
		return msg
	}

	for jobID, reSchedulerTasksData := range jobMap {
		// 	No job
		job, ok := ssn.Jobs[jobID]
		if !ok {
			klog.V(logErrorLev).Infof("delete %s from configMap due to not existence.", jobID)
			delete(jobMap, jobID)
			continue
		}
		// For job running
		if job.PodGroup.Status.Phase == scheduling.PodGroupRunning {
			klog.V(logErrorLev).Infof("delete %s from configMap due to job is ok.", jobID)
			delete(jobMap, jobID)
			continue
		}
		// For Node doesn't last too long
		for _, preTime := range reSchedulerTasksData.Time {
			nowTime := time2.Now().Unix()
			if nowTime-preTime > maxIntervalTime {
				klog.V(logErrorLev).Infof("delete %s from CM for overTime %v => %v.", jobID, nowTime, preTime)
				delete(jobMap, jobID)
			}
		}
	}
	ReSchedulerCache[CmJobKind] = jobMap
	return nil
}

func getRunningJobUsedNodes(job *api.JobInfo) (map[string]*v1.Pod, error) {
	var nodeNames = make(map[string]*v1.Pod, constIntNum3)

	for _, task := range job.Tasks {
		nodeNames[task.NodeName] = task.Pod
	}
	klog.V(logDebugLev).Infof("getRunningJobUsedNodes %s use %v %v.", job.Name, len(nodeNames), job.Tasks)

	if len(nodeNames) == 0 {
		return nil, fmt.Errorf("%s no tasks,no use node", job.Name)
	}

	return nodeNames, nil
}

func getTaskUseNPUs(nodesTask map[string]*v1.Pod, nodeName string) ([]string, error) {
	tmpPod, ok := nodesTask[nodeName]
	if !ok {
		return nil, fmt.Errorf("not use %s", nodeName)
	}

	return getPodUsedNPUS(tmpPod, npu800And9000CardName, node910X8NPUNum)
}

func getFaultCardsFromCache() (map[string]FaultNPUsOnNode, error) {
	allFaultNPUs, cacheOk := ReSchedulerCache[CmCardKind]
	if !cacheOk {
		klog.V(logErrorLev).Infof("getFaultCardsFromCache %+v.", ReSchedulerCache)
		return nil, errors.New("nil fault NPU cache")
	}
	faultCards, ok := allFaultNPUs.(map[string]FaultNPUsOnNode)
	if !ok || len(faultCards) == 0 {
		return nil, errors.New("nil fault NPU cards")
	}

	return faultCards, nil
}

func getFaultNodesFromCache() (map[string]FaultNodeState, error) {
	allFaultNodes := ReSchedulerCache[CmNodeKind]
	faultNodes, ok := allFaultNodes.(map[string]FaultNodeState)
	if !ok || len(faultNodes) == 0 {
		return nil, errors.New("nil fault NPU cards")
	}

	return faultNodes, nil
}

func isJobHasFaultNPU(nodesTask map[string]*v1.Pod) bool {
	allFaultNPUs, err := getFaultCardsFromCache()
	if err != nil {
		klog.V(logErrorLev).Infof("isJobHasFaultNPU %+v.", err)
		return false
	}

	for nodeName, nodeFaultNPUs := range allFaultNPUs {
		taskNPUs, getErr := getTaskUseNPUs(nodesTask, nodeName)
		if getErr != nil {
			klog.V(logDebugLev).Infof("getTaskUseNPUs %v.", getErr)
			continue
		}
		if isTaskHasFaultNPU(taskNPUs, nodeFaultNPUs.FaultNPUs) {
			return true
		}
	}
	return false
}

func isJobHasNetworkUnhealthyNPU(nodesTask map[string]*v1.Pod, job *api.JobInfo) bool {
	if !IsDistributedJob(job) {
		return false
	}

	allFaultNPUs, err := getFaultCardsFromCache()
	if err != nil {
		return false
	}

	for nodeName, nodeFaultNPUs := range allFaultNPUs {
		taskNPUs, getErr := getTaskUseNPUs(nodesTask, nodeName)
		if getErr != nil {
			klog.V(logErrorLev).Infof("getTaskUseNPUs  %v.", err)
			continue
		}
		if isTaskHasFaultNPU(taskNPUs, nodeFaultNPUs.NetworkUnhealthyNPUs) {
			return true
		}
	}
	return false
}

func isJobHasFaultNodes(nodesTask map[string]*v1.Pod) bool {
	allFaultNodes, err := getFaultNodesFromCache()
	if err != nil {
		return false
	}

	for faultNodeName := range allFaultNodes {
		for useNode := range nodesTask {
			if strings.EqualFold(faultNodeName, useNode) {
				return true
			}
		}
	}
	return false
}

// GetJobFaultRescheduleLabel Get job's fault reschedule label.
func GetJobFaultRescheduleLabel(job *api.JobInfo) (string, error) {
	value, ok := job.PodGroup.Labels[jobRescheduleLabelKey]
	if !ok {
		return "", fmt.Errorf("%s no job reschedule label", job.Name)
	}

	if _, setOk := reSchedulerJobController[value]; !setOk {
		msg := fmt.Errorf("%s fault reschedule label %+v not support", job.Name, value)
		klog.V(logErrorLev).Infof("GetJobFaultRescheduleLabel %+v.", msg)
		return "", msg
	}
	return value, nil
}

// Reschedule switch, no longer refinement.
func isJobSetFaultRescheduleLabel(job *api.JobInfo) bool {
	value, err := GetJobFaultRescheduleLabel(job)
	if err != nil {
		klog.V(logDebugLev).Infof("set fault reschedule label %+v.", err)
		return false
	}

	if value == JobOffRescheduleLabelValue {
		klog.V(logInfoLev).Infof("%s set fault reschedule label %+v.", job.UID, value)
		return false
	}
	return true
}

// IsDistributedJob To judge whether the distributed job.
func IsDistributedJob(job *api.JobInfo) bool {
	if len(job.Tasks) > 1 {
		return true
	}

	return false
}

func isJobHasFaultResources(nodeAndPods map[string]*v1.Pod, job *api.JobInfo) bool {
	if isJobHasFaultNodes(nodeAndPods) {
		klog.V(logDebugLev).Infof("isJobHasFaultNodes %s use fault node.", job.Name)
		return true
	}

	if isJobHasFaultNPU(nodeAndPods) {
		klog.V(logDebugLev).Infof("isJobHasFaultNPU %s use fault npu.", job.Name)
		return true
	}

	if isJobHasNetworkUnhealthyNPU(nodeAndPods, job) {
		klog.V(logDebugLev).Infof("isJobHasNetworkUnhealthyNPU %s use NetworkUnhealthy npu.", job.Name)
		return true
	}

	klog.V(logDebugLev).Infof("isJobHasFaultResources %s no use fault resources.", job.Name)

	return false
}

// GetFaultNPUJobs Obtain information about the running jobs that uses the faulty resource.
func GetFaultNPUJobs(jobs map[string]*api.JobInfo) ([]FaultNPUJob, error) {
	var faultNPUJobs []FaultNPUJob

	for _, job := range jobs {
		if !isJobSetFaultRescheduleLabel(job) {
			klog.V(logErrorLev).Infof("%s not set rescheduleLabel, no need reschedule.", job.Name)
			continue
		}
		klog.V(logInfoLev).Infof("%s set rescheduleLabel, has fault reschedule feature.", job.Name)
		// Get running job used node.
		nodeAndPods, getErr := getRunningJobUsedNodes(job)
		if getErr != nil {
			klog.V(logDebugLev).Infof("GetFaultNPUJobs %s %v.", job.Name, getErr)
			continue
		}

		if !isJobHasFaultResources(nodeAndPods, job) {
			klog.V(logDebugLev).Infof("isJobHasFaultResources %s has no fault resources.", job.Name)
			continue
		}

		klog.V(logDebugLev).Infof("GetFaultNPUJobs %s use fault resource.", job.Name)
		faultJob, err := getFaultNodePODAndRankIndex(job, nodeAndPods)
		if err != nil {
			klog.V(logDebugLev).Infof("GetFaultNPUJobs %s %v.", job.Name, err)
			continue
		}
		faultNPUJobs = append(faultNPUJobs, faultJob)
	}

	if len(faultNPUJobs) == 0 {
		return nil, errors.New("get none faultNPUJobs")
	}

	return faultNPUJobs, nil
}

// GetRestartNPUFaultJobs Find the fault job witch need be restart.
func GetRestartNPUFaultJobs(faultNPUJobs []FaultNPUJob, jobs map[string]*api.JobInfo) ([]*api.JobInfo, error) {
	var restartJobs []*api.JobInfo
	for _, faultNPUJob := range faultNPUJobs {
		for _, job := range jobs {
			if job.Name == faultNPUJob.jobName {
				restartJobs = append(restartJobs, job)
			}
		}
	}

	if len(restartJobs) == 0 {
		return nil, errors.New("none restart jobs get")
	}

	return restartJobs, nil
}

func deleteJobRankIndex(job *api.JobInfo) {
	for _, task := range job.Tasks {
		delete(task.Pod.Annotations, podRankIndex)
		klog.V(logDebugLev).Infof("delete %s rankIndex %+v for job pending.", task.UID, task.Pod.Annotations)
	}
}

// ReleaseFaultJobTakeNodes Release the node occupied by the faulty task.
func ReleaseFaultJobTakeNodes(job *api.JobInfo) error {
	jobInterface, getErr := ReSchedulerCache[CmJobKind]
	if !getErr {
		klog.V(logDebugLev).Infof("no ReScheduler Tasks")
		return nil
	}

	jobMap, jobErr := jobInterface.(map[api.JobID]ReSchedulerTasks)
	if !jobErr {
		klog.V(logDebugLev).Infof("%v not ReSchedulerTasks map", jobInterface)
		return nil
	}

	if _, ok := jobMap[job.UID]; !ok {
		klog.V(logErrorLev).Infof("release job(%s) not find in buffer", job.Name)
		return nil
	}

	if job.PodGroup.Status.Phase == scheduling.PodGroupPending {
		klog.V(logInfoLev).Infof("delete %s from configMap due to pending.", job.UID)
		delete(jobMap, job.UID)
		ReSchedulerCache[CmJobKind] = jobMap
		deleteJobRankIndex(job)
	}

	return nil
}

func isFaultJobInCache(job *api.JobInfo) bool {
	jobMaps, ok := ReSchedulerCache[CmJobKind]
	if !ok {
		return false
	}
	faultJobs, covertOk := jobMaps.(map[api.JobID]ReSchedulerTasks)
	if !covertOk {
		return false
	}
	_, getOK := faultJobs[job.UID]
	if !getOK {
		return false
	}
	return true
}

func getPodUsedNPUS(pod *v1.Pod, cardName string, cardMaxNum int) ([]string, error) {
	strNpu, npuOK := pod.Annotations[cardName]
	if !npuOK {
		return nil, fmt.Errorf("%s has no NPU", pod.Name)
	}

	taskNPUs := strings.Split(strNpu, ",")
	if len(taskNPUs) > cardMaxNum {
		err := fmt.Errorf("get err %s npus %v", pod.Name, taskNPUs)
		return nil, err
	}
	return taskNPUs, nil
}

func getRunningTaskUseNPUs(nowTask *api.TaskInfo) ([]string, error) {
	return getPodUsedNPUS(nowTask.Pod, npu800And9000CardName, node910X8NPUNum)
}

func getTaskInfFromJobByTaskName(taskName string, job *api.JobInfo) *api.TaskInfo {
	var nowTask *api.TaskInfo
	for _, taskInfo := range job.Tasks {
		if taskInfo.Name == taskName {
			nowTask = taskInfo
			break
		}
	}
	return nowTask
}

func isTaskUseFaultNode(task *api.TaskInfo) bool {
	allFaultNodes, err := getFaultNodesFromCache()
	if err != nil {
		return false
	}

	if _, ok := allFaultNodes[task.NodeName]; ok {
		return true
	}
	return false
}

func isTaskUseFaultNPU(task *api.TaskInfo, job *api.JobInfo) bool {
	taskUseNPUs, err := getRunningTaskUseNPUs(task)
	if err != nil {
		klog.V(logInfoLev).Infof("getRunningTaskUseNPUs %v.", err)
		return false
	}

	allFaultNPUs, cardsErr := getFaultCardsFromCache()
	if cardsErr != nil {
		klog.V(logInfoLev).Infof("isTaskUseFaultNPU %v.", err)
		return false
	}

	if nodeFaultNPUs, ok := allFaultNPUs[task.NodeName]; ok {
		if isTaskHasFaultNPU(taskUseNPUs, nodeFaultNPUs.FaultNPUs) {
			return true
		}
		if IsDistributedJob(job) {
			isTaskHasFaultNPU(taskUseNPUs, nodeFaultNPUs.NetworkUnhealthyNPUs)
			return true
		}
	}
	return false
}

func isFailedTaskInFaultJob(taskName string, job *api.JobInfo) bool {
	nowTask := getTaskInfFromJobByTaskName(taskName, job)
	if nowTask == nil {
		return false
	}
	// 1.Check whether a faulty node is used
	if isTaskUseFaultNode(nowTask) {
		return true
	}
	// 2.Check whether faulty NPUs is used
	if isTaskUseFaultNPU(nowTask, job) {
		return true
	}
	return false
}

func recordReSchedulerTaskRankIndexInCache(task ReSchedulerTasks, jobInf *api.JobInfo) error {
	var rankIndexSlice = make(map[string]struct{}, 1)
	for taskName, rankIndex := range task.RankIndexes {
		if isFailedTaskInFaultJob(taskName, jobInf) {
			rankIndexSlice[rankIndex] = struct{}{}
		}
	}
	var rankIndexMap = map[api.JobID]TaskUsedRankIndex{
		jobInf.UID: {
			FaultNodeRankIndex: rankIndexSlice,
			UpdateTime:         time2.Now().Unix(),
		},
	}
	ReSchedulerCache[TmpAllocRankIndexKind] = rankIndexMap
	klog.V(logDebugLev).Infof("set abnormal job used rankIndex %+v.", rankIndexMap)
	return nil
}

// Record and curing RankIndex information
func writeFaultJobInfInCache(jobs map[string]*api.JobInfo, fJob FaultNPUJob, task ReSchedulerTasks) error {
	job, ok := jobs[fJob.jobName]
	if !ok {
		return fmt.Errorf("%s not found in fault jobs", fJob.jobName)
	}

	if isFaultJobInCache(job) {
		return fmt.Errorf("%s already in fault jobs cache", job.UID)
	}

	var jobMap = make(map[api.JobID]ReSchedulerTasks, 1)
	jobMap[job.UID] = task
	ReSchedulerCache[CmJobKind] = jobMap

	return recordReSchedulerTaskRankIndexInCache(task, job)
}

// GetDelayingDeleteJobs Get delaying delete jobs, only "grace" can be.
func GetDelayingDeleteJobs(restartJobs map[api.JobID]*api.JobInfo) ([]*api.JobInfo, error) {
	var dJob []*api.JobInfo
	if len(restartJobs) == 0 {
		return nil, errors.New("no restartJobs")
	}
	for _, job := range restartJobs {
		label, getErr := GetJobFaultRescheduleLabel(job)
		if getErr != nil {
			label = JobOffRescheduleLabelValue
			continue
		}
		if label == JobGraceRescheduleLabelValue {
			dJob = append(dJob, job)
		}
	}
	return dJob, nil
}

func isDelayingJobTimeOut(dJob *api.JobInfo) bool {
	jobs, getErr := getRankIdJobsFromCache()
	if getErr != nil {
		klog.V(logErrorLev).Infof("isDelayingJobTimeOut %v.", getErr)
		return false
	}
	rankIdJob, ok := jobs[dJob.UID]
	if !ok {
		klog.V(logErrorLev).Infof("isDelayingJobTimeOut %s not in cache.", dJob.UID)
		return false
	}
	now := time2.Now().Unix()
	klog.V(logDebugLev).Infof("isDelayingJobTimeOut now:%v create:%v.", now, rankIdJob.CreatTime)
	if now-rankIdJob.CreatTime > GraceOverTime {
		return true
	}
	if rankIdJob.CreatTime-now > GraceOverTime {
		return true
	}
	return false
}

func getRankIdJobsFromCache() (map[api.JobID]FaultRankIDRecordJobCMData, error) {
	tmpValue, ok := ReSchedulerCache[CmJobRankIds]
	if !ok {
		return nil, fmt.Errorf("")
	}
	klog.V(logErrorLev).Infof("getRankIdJobsFromCache %v.", tmpValue)
	jobRankIds, getErr := getJobFaultNPURankIdsByData(tmpValue)
	if getErr != nil {
		msg := fmt.Errorf("assert %v to map[string]FaultRankIDRecordJobCMData failed", tmpValue)
		klog.V(logErrorLev).Infof("synJobRankIdsCache %v.", msg)
		return nil, msg
	}
	klog.V(logErrorLev).Infof("getJobFaultNPURankIdsByData %v.", jobRankIds)
	return jobRankIds, nil
}

func isJobHasBeenGraceDeleted(ssn *framework.Session, dJob *api.JobInfo) bool {
	// 1.get all job RankIds.
	jobsRankIds, getErr := getRankIdJobsFromCache()
	if getErr != nil {
		return false
	}
	klog.V(logDebugLev).Infof("%v getRankIdJobsFromCache %v.", dJob.UID, jobsRankIds)
	// 2.Job is in cache.
	jobRankIds, ok := jobsRankIds[dJob.UID]
	if !ok {
		return false
	}
	klog.V(logDebugLev).Infof("%s jobRankIds %v.", dJob.UID, jobRankIds)
	// 3.Job has been grace delete or not.Judge by exists and creat time.
	jobK8sPods, podErr := getJobPodsInfoFromK8s(ssn, dJob)
	if podErr != nil {
		klog.V(logErrorLev).Infof("getJobPodsInfoFromK8s %v.", podErr)
		return true
	}
	klog.V(logDebugLev).Infof("%s getJobPodsInfoFromK8s: %v.", dJob.UID, len(jobK8sPods))
	for k, podName := range jobRankIds.PodsName {
		tmpPod, getOk := jobK8sPods[podName]
		if !getOk {
			klog.V(logErrorLev).Infof("k8s not has pod  %v.", podName)
			return true
		}
		if tmpPod.CreationTimestamp.Unix() != jobRankIds.PodsCreatTime[k] {
			klog.V(logErrorLev).Infof("isJobHasBeenGraceDeleted new pod:%v---old pod:%v.",
				tmpPod.CreationTimestamp.Unix(), jobRankIds.PodsCreatTime[k])
			return true
		}
	}
	klog.V(logInfoLev).Infof("isJobHasBeenGraceDeleted pods %v not grace delete.", jobRankIds.PodsName)
	return false
}

// GetRecordJobPods Get Job Pods info.
func GetRecordJobPods(dJob *api.JobInfo) (map[string]int64, map[string]types.UID, error) {
	rankIdData, ok := ReSchedulerCache[CmJobRankIds]
	if !ok {
		return nil, nil, fmt.Errorf("none %v in cache", CmJobRankIds)
	}
	rankIdsMap, covErr := convertToRankIdsMapFromCache(rankIdData)
	if covErr != nil {
		klog.V(logDebugLev).Infof("getRecordJobPods: %v.", covErr)
		return nil, nil, covErr
	}
	jobId := api.JobID(dJob.Namespace + "/" + dJob.Name)
	value, ok := rankIdsMap[jobId]
	if !ok {
		return nil, nil, fmt.Errorf("none job %v in cache", jobId)
	}

	jobPodsTime := make(map[string]int64, constIntNum3)
	jobPodsUID := make(map[string]types.UID, constIntNum3)
	for k, v := range value.PodsName {
		tmpTime := value.PodsCreatTime[k]
		tmpUID := value.PodsUID[k]
		jobPodsTime[v] = tmpTime
		jobPodsUID[v] = tmpUID
	}
	return jobPodsTime, jobPodsUID, nil
}

func isJobGraceDeletedSuccess(dJob *api.JobInfo) bool {
	if len(dJob.Tasks) == 0 {
		// old pod has been deleted.
		klog.V(logDebugLev).Infof("isJobGraceDeletedSuccess: %v pods has been deleted.", dJob.Name)
		return true
	}

	jobPodsTime, _, err := GetRecordJobPods(dJob)
	if err != nil {
		return false
	}

	for _, task := range dJob.Tasks {
		// The new POD is inconsistent with the old one.
		if task.Pod.CreationTimestamp.Unix() != jobPodsTime[task.Pod.Name] {
			klog.V(logDebugLev).Infof("pod restart success[new:%v---old:%v]",
				task.Pod.CreationTimestamp.Unix(), jobPodsTime[task.Pod.Name])
			return true
		}
	}
	return false
}

func deleteRedundantRankIdCM(ssn *framework.Session, nameSpace string, dJob api.JobID) error {
	temp := strings.Split(string(dJob), "/")
	configMapName := JobFaultRankIdCMPre + temp[1]
	klog.V(logInfoLev).Infof("deleteRedundantRankIdCM %v:%v.", nameSpace, configMapName)
	if err := deleteSchedulerConfigMap(ssn, nameSpace, configMapName); err != nil {
		return err
	}
	return nil
}

// GetNeedForceDeleteDelayingJobs Get delaying jobs which need be force deleted.
func GetNeedForceDeleteDelayingJobs(ssn *framework.Session, dJobs []*api.JobInfo) ([]*api.JobInfo, error) {
	if len(dJobs) == 0 {
		msg := errors.New("none jobs")
		klog.V(logDebugLev).Infof("getNeedForceDeleteDelayingJobs %v.", msg)
		return nil, msg
	}
	var forceJobs []*api.JobInfo
	for _, dJob := range dJobs {
		klog.V(logInfoLev).Infof("IsDelayingJobNeedForceDelete %v.", dJob.Name)
		// Check whether the task has been deleted
		if !isJobHasBeenGraceDeleted(ssn, dJob) {
			klog.V(logErrorLev).Infof("%v first be grace Deleted.", dJob.Name)
			// 3.Record into job-fault-configMap.
			if cmErr := WriteJobFaultRankIDIntoCacheAndCM(ssn, dJob); cmErr != nil {
				klog.V(logErrorLev).Infof("GetDelayingDeleteJobs %v.", cmErr)
			}
			continue
		}
		klog.V(logDebugLev).Infof("%v not first grace Deleted.", dJob.Name)
		if isJobGraceDeletedSuccess(dJob) {
			klog.V(logDebugLev).Infof("%v grace deleted successful.", dJob.Name)
			continue
		}
		if isDelayingJobTimeOut(dJob) {
			forceJobs = append(forceJobs, dJob)
		}
	}

	return forceJobs, nil
}

func getFaultJobPODRankIndexMapFromCache(restartJob *api.JobInfo) (map[string]string, error) {
	podRankIndexes := make(map[string]string, constIntNum3)
	for _, task := range restartJob.Tasks {
		fTask, getErr := getReSchedulerTasksFromCache(task)
		if getErr != nil {
			continue
		}
		rankIndex, ok := fTask.RankIndexes[task.Name]
		if !ok {
			continue
		}
		podRankIndexes[task.NodeName] = rankIndex
	}
	if len(podRankIndexes) == 0 {
		return nil, fmt.Errorf("%s none rankIndex in cache", restartJob.UID)
	}
	return podRankIndexes, nil
}
