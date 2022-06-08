/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package rescheduling is using for HuaWei Ascend pin fault rescheduling.

*/
package rescheduling

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	"volcano.sh/apis/pkg/apis/scheduling"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/util"
)

func convertToRankIdsMapFromCache(jobData interface{}) (map[api.JobID]FaultRankIDRecordJobCMData, error) {
	rankIds, reOk := jobData.(map[api.JobID]FaultRankIDRecordJobCMData)
	if !reOk {
		msg := fmt.Errorf("assert %v to FaultRankIDRecordJobCMData failed", jobData)
		klog.V(util.LogErrorLev).Infof("%v.", msg)
		return nil, msg
	}

	if len(rankIds) == 0 {
		msg := fmt.Errorf("rank ids is nil")
		klog.V(util.LogErrorLev).Infof("convertToRankIdsMapFromCache: %v.", msg)
		return nil, msg
	}

	return rankIds, nil
}

func marshalCacheDataToStringByReplaceSlash(jobData interface{}) (string, error) {
	data, err := util.MarshalCacheDataToString(jobData)
	if err != nil {
		return "", err
	}
	dataTemp := strings.ReplaceAll(data, "/", "_")
	return dataTemp, nil
}

func getCMJobWriteData(jobData interface{}) (string, error) {
	return marshalCacheDataToStringByReplaceSlash(jobData)
}

func getCMRankIdsWriteData(jobData interface{}) (string, error) {
	return marshalCacheDataToStringByReplaceSlash(jobData)
}

func checkJobPodStatusOK(ssn *framework.Session, job *api.JobInfo) bool {
	for _, task := range job.Tasks {
		realPod, err := util.GetRealPodByTask(ssn, task)
		if err != nil {
			klog.V(util.LogErrorLev).Infof("checkJobPodStatusOK  getRealPodByTask %v %v.", task.Name, err)
			return false
		}
		if realPod.Status.Phase == v1.PodRunning {
			continue
		}
		klog.V(util.LogErrorLev).Infof("checkJobPodStatusOK %v not ok %v.", task.Name, realPod.Status.Phase)
		return false
	}
	klog.V(util.LogInfoLev).Infof("%v checkJobPodStatusOK.", job.Name)
	return true
}

// Delete expired job data.
func synReSchedulerJobCache(ssn *framework.Session, tmpValue interface{}) error {
	jobMap, assertOk := tmpValue.(map[api.JobID]ReSchedulerTasks)
	if !assertOk {
		msg := fmt.Errorf("convert %v to map[api.JobID]ReSchedulerTasks failed", tmpValue)
		klog.V(util.LogErrorLev).Infof("synReSchedulerJobCache %v.", msg)
		return msg
	}

	newJobMap := make(map[api.JobID]ReSchedulerTasks, util.NPUIndex3)
	for jobID, reSchedulerTasksData := range jobMap {
		// 	No job
		job, ok := ssn.Jobs[jobID]
		if !ok {
			klog.V(util.LogErrorLev).Infof("delete %s from configMap due to not existence.", jobID)
			continue
		}

		// For job running, need delete fault job record.
		if util.IsJobInitial(job) {
			if checkJobPodStatusOK(ssn, job) {
				klog.V(util.LogErrorLev).Infof("delete %s from configMap due to job is ok.", jobID)
				continue
			}
			klog.V(util.LogErrorLev).Infof("%s 's status is ok, but not pod, cannot delete this job.", jobID)
		}

		// For Node doesn't last too long
		flag := false
		for _, preTime := range reSchedulerTasksData.Time {
			nowTime := time.Now().Unix()
			missTime := nowTime - preTime
			if missTime > maxIntervalTime+GraceOverTime {
				klog.V(util.LogErrorLev).Infof("delete %s from CM for overTime %v => %v.", jobID, nowTime,
					preTime)
				flag = true
				break
			}
		}
		if flag {
			continue
		}
		newJobMap[jobID] = reSchedulerTasksData
	}
	ReSchedulerCache[CmJobKind] = newJobMap
	return nil
}

func getRunningJobUsedNodes(job *api.JobInfo) (map[string]*v1.Pod, error) {
	var nodeNames = make(map[string]*v1.Pod, util.NPUIndex3)

	for _, task := range job.Tasks {
		nodeNames[task.NodeName] = task.Pod
	}
	klog.V(util.LogErrorLev).Infof("getRunningJobUsedNodes %s use %v.", job.Name, len(nodeNames))

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

	return getPodUsedNPUS(tmpPod, AscendNPUPodRealUse, node910X8NPUNum)
}

func getFaultCardsFromCache() (map[string]FaultNPUsOnNode, error) {
	allFaultNPUs, cacheOk := ReSchedulerCache[CmCardKind]
	if !cacheOk {
		klog.V(util.LogErrorLev).Infof("getFaultCardsFromCache %+v.", ReSchedulerCache)
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
		klog.V(util.LogErrorLev).Infof("isJobHasFaultNPU %+v.", err)
		return false
	}

	for nodeName, nodeFaultNPUs := range allFaultNPUs {
		taskNPUs, getErr := getTaskUseNPUs(nodesTask, nodeName)
		if getErr != nil {
			klog.V(util.LogErrorLev).Infof("getTaskUseNPUs %v.", getErr)
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
			klog.V(util.LogErrorLev).Infof("getTaskUseNPUs %v.", getErr)
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
		klog.V(util.LogErrorLev).Infof("getFaultNodesFromCache %v.", err)
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
		klog.V(util.LogErrorLev).Infof("GetJobFaultRescheduleLabel %+v.", job.PodGroup.Labels)
		return "", fmt.Errorf("%s no job reschedule label", job.Name)
	}

	if _, setOk := reSchedulerJobController[value]; !setOk {
		msg := fmt.Errorf("%s fault reschedule label %+v not support", job.Name, value)
		klog.V(util.LogErrorLev).Infof("GetJobFaultRescheduleLabel %+v.", msg)
		return "", msg
	}
	return value, nil
}

// Reschedule switch, no longer refinement.
func isJobSetFaultRescheduleLabel(job *api.JobInfo) bool {
	value, err := GetJobFaultRescheduleLabel(job)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("set fault reschedule label %+v.", err)
		return false
	}

	if value == JobOffRescheduleLabelValue {
		klog.V(util.LogInfoLev).Infof("%s set fault reschedule label %+v.", job.UID, value)
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
		klog.V(util.LogErrorLev).Infof("isJobHasFaultNodes %s use fault node.", job.Name)
		return true
	}

	if isJobHasFaultNPU(nodeAndPods) {
		klog.V(util.LogErrorLev).Infof("isJobHasFaultNPU %s use fault npu.", job.Name)
		return true
	}

	if isJobHasNetworkUnhealthyNPU(nodeAndPods, job) {
		klog.V(util.LogErrorLev).Infof("isJobHasNetworkUnhealthyNPU %s use NetworkUnhealthy npu.", job.Name)
		return true
	}

	klog.V(util.LogInfoLev).Infof("isJobHasFaultResources %s no use fault resources.", job.Name)

	return false
}

// GetFaultNPUJobs Obtain information about the running jobs that uses the faulty resource.
func GetFaultNPUJobs(jobs map[string]*api.JobInfo) ([]FaultNPUJob, error) {
	var faultNPUJobs []FaultNPUJob

	for _, job := range jobs {
		if !isJobSetFaultRescheduleLabel(job) {
			klog.V(util.LogErrorLev).Infof("%s not set rescheduleLabel, no need reschedule.", job.Name)
			continue
		}
		klog.V(util.LogInfoLev).Infof("%s set rescheduleLabel, has fault reschedule feature.", job.Name)
		// Get running job used node.
		nodeAndPods, getErr := getRunningJobUsedNodes(job)
		if getErr != nil {
			klog.V(util.LogErrorLev).Infof("GetFaultNPUJobs %s %v.", job.Name, getErr)
			continue
		}

		if !isJobHasFaultResources(nodeAndPods, job) {
			klog.V(util.LogErrorLev).Infof("isJobHasFaultResources %s has no fault resources.", job.Name)
			continue
		}

		klog.V(util.LogErrorLev).Infof("GetFaultNPUJobs %s use fault resource.", job.Name)
		faultJob, err := getFaultNodePODAndRankIndex(job, nodeAndPods)
		if err != nil {
			klog.V(util.LogErrorLev).Infof("GetFaultNPUJobs %s %v.", job.Name, err)
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
		klog.V(util.LogErrorLev).Infof("delete %s rankIndex for job pending.", task.UID)
	}
}

// ReleaseFaultJobTakeNodes Release the node occupied by the faulty task.
func ReleaseFaultJobTakeNodes(job *api.JobInfo) error {
	jobInterface, getErr := ReSchedulerCache[CmJobKind]
	if !getErr {
		klog.V(util.LogErrorLev).Infof("no ReScheduler Tasks")
		return nil
	}

	jobMap, jobErr := jobInterface.(map[api.JobID]ReSchedulerTasks)
	if !jobErr {
		klog.V(util.LogErrorLev).Infof("%v not ReSchedulerTasks map", jobInterface)
		return nil
	}

	if _, ok := jobMap[job.UID]; !ok {
		klog.V(util.LogErrorLev).Infof("release job(%s) not find in buffer", job.Name)
		return nil
	}

	if job.PodGroup.Status.Phase == scheduling.PodGroupPending {
		klog.V(util.LogInfoLev).Infof("delete %s from configMap due to pending.", job.UID)
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
		return nil, fmt.Errorf("%s has no NPU from %s", pod.Name, cardName)
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
		klog.V(util.LogInfoLev).Infof("getRunningTaskUseNPUs %v.", err)
		return false
	}

	allFaultNPUs, cardsErr := getFaultCardsFromCache()
	if cardsErr != nil {
		klog.V(util.LogInfoLev).Infof("isTaskUseFaultNPU %v.", err)
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
	var rankIndexSlice = make(map[string]struct{ UpdateTime int64 }, 1)
	now := time.Now().Unix()
	for key, rankIndex := range task.RankIndexes {
		if isFailedTaskInFaultJob(task.TaskName[key], jobInf) {
			rankIndexSlice[rankIndex] = struct{ UpdateTime int64 }{UpdateTime: now}
		}
	}
	var rankIndexMap = map[api.JobID]TaskUsedRankIndex{
		jobInf.UID: {
			FaultNodeRankIndex: rankIndexSlice,
			UpdateTime:         now,
		},
	}
	ReSchedulerCache[TmpAllocRankIndexKind] = rankIndexMap
	klog.V(util.LogErrorLev).Infof("set abnormal job used rankIndex %+v.", rankIndexMap)
	return nil
}

// Record and curing RankIndex information
func writeFaultJobInfInCache(job *api.JobInfo, task ReSchedulerTasks) error {
	if isFaultJobInCache(job) {
		return fmt.Errorf("%s already in fault jobs cache", job.UID)
	}

	var jobMap = make(map[api.JobID]ReSchedulerTasks, 1)
	jobMap[job.UID] = task
	ReSchedulerCache[CmJobKind] = jobMap
	return recordReSchedulerTaskRankIndexInCache(task, job)
}

func isDelayingJobTimeOut(dJob *api.JobInfo) bool {
	jobs, getErr := getRankIDJobsFromCache()
	if getErr != nil {
		klog.V(util.LogErrorLev).Infof("isDelayingJobTimeOut %v.", getErr)
		return false
	}
	rankIDJob, ok := jobs[dJob.UID]
	if !ok {
		klog.V(util.LogErrorLev).Infof("isDelayingJobTimeOut %s not in cache.", dJob.UID)
		return false
	}
	now := time.Now().Unix()
	klog.V(util.LogErrorLev).Infof("isDelayingJobTimeOut now:%v create:%v.", now, rankIDJob.CreatTime)
	if now-rankIDJob.CreatTime > GraceOverTime {
		return true
	}
	if rankIDJob.CreatTime-now > GraceOverTime {
		return true
	}
	return false
}

func getRankIDJobsFromCache() (map[api.JobID]FaultRankIDRecordJobCMData, error) {
	tmpValue, ok := ReSchedulerCache[CmJobRankIds]
	if !ok {
		return nil, fmt.Errorf("")
	}
	klog.V(util.LogErrorLev).Infof("getRankIdJobsFromCache %v.", tmpValue)
	jobRankIds, getErr := getJobFaultNPURankIdsByData(tmpValue)
	if getErr != nil {
		msg := fmt.Errorf("assert %v to map[string]FaultRankIDRecordJobCMData failed", tmpValue)
		klog.V(util.LogErrorLev).Infof("synJobRankIdsCache %v.", msg)
		return nil, msg
	}
	klog.V(util.LogErrorLev).Infof("getJobFaultNPURankIdsByData %+v.", jobRankIds)
	return jobRankIds, nil
}

// GetRecordJobPods Get Job Pods info.
func GetRecordJobPods(dJob *api.JobInfo) (map[string]int64, map[string]types.UID, error) {
	rankIDData, ok := ReSchedulerCache[CmJobRankIds]
	if !ok {
		return nil, nil, fmt.Errorf("none %v in cache", CmJobRankIds)
	}
	rankIdsMap, covErr := convertToRankIdsMapFromCache(rankIDData)
	if covErr != nil {
		klog.V(util.LogErrorLev).Infof("getRecordJobPods: %v.", covErr)
		return nil, nil, covErr
	}
	jobID := api.JobID(dJob.Namespace + "/" + dJob.Name)
	value, ok := rankIdsMap[jobID]
	if !ok {
		return nil, nil, fmt.Errorf("none job %v in cache", jobID)
	}

	jobPodsTime := make(map[string]int64, util.NPUIndex3)
	jobPodsUID := make(map[string]types.UID, util.NPUIndex3)
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
		klog.V(util.LogErrorLev).Infof("isJobGraceDeletedSuccess: %v pods has been deleted.", dJob.Name)
		return true
	}

	jobPodsTime, _, err := GetRecordJobPods(dJob)
	if err != nil {
		return false
	}

	restartNum := 0
	for _, task := range dJob.Tasks {
		// The new POD is inconsistent with the old one.
		if task.Pod.CreationTimestamp.Unix() != jobPodsTime[task.Pod.Name] {
			klog.V(util.LogErrorLev).Infof("pod restart success[new:%v---old:%v]",
				task.Pod.CreationTimestamp.Unix(), jobPodsTime[task.Pod.Name])
			restartNum++
		}
	}
	if restartNum == len(dJob.Tasks) {
		klog.V(util.LogErrorLev).Infof("job all pod %d restart success.", restartNum)
		return true
	}
	return false
}

// GetNeedForceDeleteDelayingJobs Get delaying jobs which need be force deleted.
func GetNeedForceDeleteDelayingJobs(ssn *framework.Session,
	dJobs map[api.JobID]ReSchedulerTasks) ([]*api.JobInfo, error) {
	var forceJobs []*api.JobInfo
	for jobID := range dJobs {
		jobInf, ok := ssn.Jobs[jobID]
		if !ok {
			klog.V(util.LogErrorLev).Infof("GetNeedForceDeleteDelayingJobs %v not in ssn.", jobID)
			continue
		}

		if isJobGraceDeletedSuccess(jobInf) {
			klog.V(util.LogErrorLev).Infof("%v grace deleted successful.", jobInf.Name)
			continue
		}

		if !isDelayingJobTimeOut(jobInf) {
			continue
		}
		klog.V(util.LogErrorLev).Infof("%v is time out for delete.", jobInf.Name)
		forceJobs = append(forceJobs, jobInf)
	}

	if len(forceJobs) == 0 {
		klog.V(util.LogErrorLev).Infof("GetNeedForceDeleteDelayingJobs get nil jobs.")
		return nil, errors.New("get none jobs")
	}
	return forceJobs, nil
}

func getFaultJobPODRankIndexMapFromCache(restartJob *api.JobInfo) (map[string]string, error) {
	podRankIndexes := make(map[string]string, util.NPUIndex3)
	for _, task := range restartJob.Tasks {
		fTask, getErr := getReSchedulerTasksFromCache(task)
		if getErr != nil {
			continue
		}

		rankIndex := ""
		for key, value := range fTask.TaskName {
			if value == task.Name {
				rankIndex = fTask.RankIndexes[key]
				break
			}
		}
		if len(rankIndex) == 0 {
			continue
		}
		podRankIndexes[task.NodeName] = rankIndex
	}
	if len(podRankIndexes) == 0 {
		return nil, fmt.Errorf("%s none rankIndex in cache", restartJob.UID)
	}
	return podRankIndexes, nil
}

// GetGraceDeleteJobsFromCache Get grace delete jobs from ReSchedulerCache.
func GetGraceDeleteJobsFromCache() (map[api.JobID]ReSchedulerTasks, error) {
	jobMap, getErr := getReSchedulerJobsMapFromCache()
	if getErr != nil {
		klog.V(util.LogErrorLev).Infof("getReSchedulerJobsMapFromCache %v.", getErr)
		return nil, getErr
	}

	deleteJobMap := make(map[api.JobID]ReSchedulerTasks, util.NPUIndex3)
	for jobUID, reSchedulerTasks := range jobMap {
		if reSchedulerTasks.GraceTaskFlag == true {
			deleteJobMap[jobUID] = reSchedulerTasks
		}
	}

	if len(deleteJobMap) == 0 {
		return nil, errors.New("none grace delete jobs")
	}
	return deleteJobMap, nil
}

func getJobFaultRankIds(job *api.JobInfo) (string, error) {
	allFaultNPUs, cacheErr := getFaultCardsFromCache()
	if cacheErr != nil {
		klog.V(util.LogErrorLev).Infof("getJobFaultRankIds %s %v.", job.Name, cacheErr)
		return "", cacheErr
	}

	klog.V(util.LogErrorLev).Infof("getJobFaultRankIds %s %v.", job.Name, allFaultNPUs)
	var rankIds []string
	rankIndexMap, err := getFaultJobPODRankIndexMapFromCache(job)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("getJobFaultRankIds %s %v.", job.Name, err)
		return "", err
	}
	for nodeName, rankIndexStr := range rankIndexMap {
		klog.V(util.LogInfoLev).Infof("getJobFaultRankIds %s %v.", nodeName, rankIndexStr)
		var faultNPUs []string
		faultNPUsOnNode, ok := allFaultNPUs[nodeName]
		if !ok {
			continue
		}
		faultNPUs = append(faultNPUs, faultNPUsOnNode.FaultNPUs...)
		faultNPUs = append(faultNPUs, faultNPUsOnNode.NetworkUnhealthyNPUs...)
		tmpDeviceID := util.ChangeTopToIntArray(strings.Join(faultNPUs, ","), npu800And9000CardPreName)
		if len(tmpDeviceID) == 0 {
			klog.V(util.LogInfoLev).Infof("%s getTopFromNode %+v nil.", nodeName, faultNPUsOnNode)
			continue
		}
		rankIndex, covertErr := strconv.Atoi(rankIndexStr)
		if covertErr != nil {
			klog.V(util.LogErrorLev).Infof("%s getJobFaultRankIds covert %v.", covertErr, rankIndexStr)
			continue
		}
		for _, tmp := range tmpDeviceID {
			rankIds = append(rankIds, strconv.Itoa(tmp+rankIndex*node910X8NPUNum))
		}
	}
	dataBuffer, err := json.Marshal(rankIds)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("marshalCacheDataToString err: %v.", err)
		return "", err
	}
	return string(dataBuffer), nil
}

func getJobPodsAndCreateTimeByRankIDJobs(job *api.JobInfo) (FaultRankIDRecordJobCMData, error) {
	var podNames []string
	var podCreatTimes []int64
	var podsUID []types.UID
	rankIds := FaultRankIDRecordJobCMData{}
	for _, task := range job.Tasks {
		podNames = append(podNames, task.Pod.Name)
		podCreatTimes = append(podCreatTimes, task.Pod.CreationTimestamp.Unix())
		podsUID = append(podsUID, task.Pod.UID)
	}
	rankIds.PodsName = podNames
	rankIds.PodsCreatTime = podCreatTimes
	rankIds.PodsUID = podsUID
	return rankIds, nil
}

// WriteJobFaultRankIDIntoCache Record into cache for recording in volcano-cm later.
func WriteJobFaultRankIDIntoCache(job *api.JobInfo) error {
	faultRankIds, getErr := getFaultNPUJobCMData(job)
	if getErr != nil {
		return getErr
	}
	rankIds, podErr := getJobPodsAndCreateTimeByRankIDJobs(job)
	if podErr != nil {
		return podErr
	}
	rankIds.CreatTime = faultRankIds.CreatTime
	rankIds.NameSpace = job.Namespace
	rankIds.FaultRankIds = faultRankIds.FaultRankIds
	var rankIDMap = map[api.JobID]FaultRankIDRecordJobCMData{
		job.UID: rankIds,
	}
	ReSchedulerCache[CmJobRankIds] = rankIDMap

	return nil
}

func getFaultNPUJobCMData(job *api.JobInfo) (*FaultRankIdsJobCMData, error) {
	fRankIds, getErr := getJobFaultRankIds(job)
	if getErr != nil {
		klog.V(util.LogErrorLev).Infof("getJobFaultRankIds err: %v.", getErr)
		return nil, getErr
	}

	faultRankIds := &FaultRankIdsJobCMData{
		FaultRankIds: fRankIds,
		CreatTime:    time.Now().Unix(),
	}
	return faultRankIds, nil
}

func getJobFaultNPURankIDCMData(job *api.JobInfo) (map[string]string, error) {
	faultRankIdsMap := make(map[string]string, util.NPUIndex3)

	faultRankIds, getErr := getFaultNPUJobCMData(job)
	if getErr != nil {
		klog.V(util.LogErrorLev).Infof("getFaultNPUJobCMData err: %v.", getErr)
		return nil, getErr
	}
	dataBuffer, err := json.Marshal(faultRankIds)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("marshalCacheDataToString err: %v.", err)
		return nil, err
	}
	faultRankIdsMap[JobFaultRankIDCMDataKey] = string(dataBuffer)

	return faultRankIdsMap, nil
}
