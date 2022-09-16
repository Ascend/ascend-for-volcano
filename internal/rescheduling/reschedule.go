/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package rescheduling is using for HuaWei Ascend pin fault rescheduling.

*/
package rescheduling

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"k8s.io/klog"
	"volcano.sh/apis/pkg/apis/scheduling"
	"volcano.sh/volcano/pkg/cli/job"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/conf"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// GetGraceDeleteTime Get the graceful delete time from configuration
func (reScheduler *ReScheduler) GetGraceDeleteTime(Conf []conf.Configuration) (int64, error) {
	klog.V(util.LogInfoLev).Infof("enter GetGraceDeleteTime ...")
	defer klog.V(util.LogInfoLev).Infof("leave GetGraceDeleteTime ...")
	if reScheduler == nil {
		klog.V(util.LogErrorLev).Infof("GetGraceDeleteTime failed: %s, nil reScheduler", util.ArgumentError)
		return DefaultGraceOverTime, errors.New(util.ArgumentError)
	}
	if len(Conf) == 0 {
		klog.V(util.LogErrorLev).Infof("GetGraceDeleteTime failed: %s, no conf", util.ArgumentError)
		return DefaultGraceOverTime, errors.New(util.ArgumentError)
	}
	// Read configmap
	configuration, err := util.GetConfigFromSchedulerConfigMap(util.CMInitParamKey, Conf)
	if err != nil {
		klog.V(util.LogErrorLev).Info("cannot get configuration, GraceOverTime will not be changed.")
		return DefaultGraceOverTime, nil
	}
	// get grace over time by user configuration
	overTimeStr, ok := configuration.Arguments[GraceOverTimeKey]
	if !ok {
		klog.V(util.LogErrorLev).Info("set GraceOverTime failed and will not be changed, " +
			"key grace-over-time doesn't exists.")
		return DefaultGraceOverTime, nil
	}
	overTime, err := strconv.ParseInt(overTimeStr, util.Base10, util.BitSize64)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("set GraceOverTime failed and will not be changed, "+
			"grace-over-time is invalid [%#v].", overTimeStr)
		return DefaultGraceOverTime, err
	}
	// check time validity
	if !reScheduler.checkGraceDeleteTimeValid(overTime) {
		return DefaultGraceOverTime, errors.New("defaultGraceOverTime is out of range")
	}
	return overTime, nil
}

func (reScheduler *ReScheduler) setGraceOverTime(value int64) {
	reScheduler.GraceDeleteTime = value
}

// checkGraceDeleteTimeValid used by GetGraceDeleteTime for validity checking
func (reScheduler *ReScheduler) checkGraceDeleteTimeValid(overTime int64) bool {
	if overTime < minGraceOverTime || overTime > maxGraceOverTime {
		klog.V(util.LogErrorLev).Infof("GraceOverTime value should be range [2, 3600], configured is [%d], "+
			"GraceOverTime will not be changed", overTime)
		return false
	}
	// use user's configuration to set grace over time
	klog.V(util.LogInfoLev).Infof("set GraceOverTime to new value [%d].", overTime)
	return true
}

// createFaultTaskHandler Create FaultTask struct and set the corresponding values
func (reScheduler *ReScheduler) createFaultTaskHandler(job *api.JobInfo, cardName string) ([]FaultTask, error) {
	var faultTasks []FaultTask
	for _, task := range job.Tasks {
		faultTask := newFaultTaskDefault(task, job)
		// 2. updateNodeRankIndex by pod.Annotation
		tmpNodeRankIndex, err := faultTask.getNodeRankIndex(task)
		if err != nil {
			klog.V(util.LogErrorLev).Infof("getNodeRankIndex %s %#v.", task.Name, err)
		}
		faultTask.setNodeRankIndex(tmpNodeRankIndex)
		// 3. update UseCardName
		if cardName == util.NPU910CardName {
			tmpUseCardName, getErr := faultTask.getUseCardName(task, cardName, util.NPUIndex8)
			if getErr != nil {
				klog.V(util.LogErrorLev).Infof("getUseCardName %s %#v", task.Name, getErr)
			}
			faultTask.setUseCardName(tmpUseCardName)
		} else {
			tmpUseCardName, getErr := faultTask.getUseCardName(task, cardName, util.NPUIndex4)
			if getErr != nil {
				klog.V(util.LogErrorLev).Infof("getUseCardName %s %#v", task.Name, getErr)
			}
			faultTask.setUseCardName(tmpUseCardName)
		}
		isFaultTask, nodeHealthState := reScheduler.getTaskHealthState(&faultTask)
		faultTask.setIsFaultTask(isFaultTask)
		faultTask.setFaultType(nodeHealthState)
		faultTasks = append(faultTasks, faultTask)
	}
	return faultTasks, nil
}

// GetRunningJobs get all the running jobs of <UseCardName> type
func (reScheduler *ReScheduler) GetRunningJobs(
	ssn *framework.Session, cardName string, jobType string) (map[api.JobID]*api.JobInfo, error) {
	klog.V(util.LogInfoLev).Infof("enter GetRunningJobs %s-%s...", cardName, jobType)
	defer klog.V(util.LogInfoLev).Infof("leave GetRunningJobs %s-%s...", cardName, jobType)
	if reScheduler == nil || ssn == nil {
		klog.V(util.LogErrorLev).Infof("GetRunningJobs failed: %s, nil reScheduler or session",
			util.ArgumentError)
		return nil, errors.New(util.ArgumentError)
	}
	var myJobs = make(map[api.JobID]*api.JobInfo, util.MapInitNum)
	for _, jobInfo := range ssn.Jobs {
		if (jobInfo.PodGroup.Status.Phase != scheduling.PodGroupRunning) &&
			(jobInfo.PodGroup.Status.Phase != scheduling.PodGroupUnknown) { // pending jobs would not be put into cache
			klog.V(util.LogDebugLev).Infof("job %s pod group is not running but %s",
				jobInfo.Name, jobInfo.PodGroup.Status.Phase)
			continue
		}
		klog.V(util.LogDebugLev).Infof("job %s status: podGroup status %s",
			jobInfo.Name, jobInfo.PodGroup.Status.Phase)
		schedulerJob, ok := reScheduler.Jobs[jobInfo.UID]
		if !ok || schedulerJob.NPUJob == nil {
			klog.V(util.LogDebugLev).Infof("job %s not in session", jobInfo.UID)
			continue
		}
		if schedulerJob.ReqNPUNum == 0 || schedulerJob.ReqNPUName != cardName {
			klog.V(util.LogDebugLev).Infof("job %s requires npu %d, name %s", schedulerJob.JobName,
				schedulerJob.ReqNPUNum, schedulerJob.ReqNPUName)
			continue
		}
		if len(schedulerJob.Selector) == 0 {
			klog.V(util.LogErrorLev).Infof("job(%s) has no selectors.", schedulerJob.JobName)
			continue
		}
		if cardName != util.NPU910CardName {
			myJobs[jobInfo.UID] = jobInfo
			continue
		}
		accType, ok := schedulerJob.Selector[util.AcceleratorType]
		if (!ok && jobType == util.ModuleAcceleratorType) || (ok && accType == jobType) { // failed, deal as 910module
			myJobs[jobInfo.UID] = jobInfo
		}
	}
	if len(myJobs) == 0 {
		return nil, fmt.Errorf("nil %s jobs", cardName)
	}
	return myJobs, nil
}

func (reScheduler *ReScheduler) is910x8Job(job plugin.SchedulerJob, cardName string) bool {
	if job.ReqNPUNum == 0 {
		klog.V(util.LogDebugLev).Infof("job %s no use npu", job.JobName)
		return false
	}
	if cardName != util.NPU910CardName {
		klog.V(util.LogDebugLev).Infof("job %s is not 910 type", job.JobName)
		return false
	}
	if job.IsJobOfCardMode() {
		klog.V(util.LogDebugLev).Infof("job %s is card mode", job.JobName)
		return false
	}
	return true
}

func (reScheduler *ReScheduler) is910x2Job(job plugin.SchedulerJob, cardName string) bool {
	if job.ReqNPUNum == 0 {
		klog.V(util.LogDebugLev).Infof("job %s no use npu", job.JobName)
		return false
	}
	if cardName != util.NPU910CardName {
		klog.V(util.LogDebugLev).Infof("job %s is not 910 type", job.JobName)
		return false
	}
	if !job.IsJobOfCardMode() {
		klog.V(util.LogDebugLev).Infof("job %s is not card mode", job.JobName)
		return false
	}
	return true
}

func (reScheduler *ReScheduler) isCard310X4Job(job plugin.SchedulerJob, cardName string) bool {
	if job.ReqNPUNum == 0 {
		klog.V(util.LogDebugLev).Infof("job %s no use npu", job.JobName)
		return false
	}
	if cardName != util.NPU310CardName {
		klog.V(util.LogDebugLev).Infof("job %s is not 910 type", job.JobName)
		return false
	}
	if !job.IsJobOfCardMode() {
		klog.V(util.LogDebugLev).Infof("job %s is not card mode", job.JobName)
		return false
	}
	return true
}

func (reScheduler *ReScheduler) isChip310X4Job(job plugin.SchedulerJob, cardName string) bool {
	if job.ReqNPUNum == 0 {
		klog.V(util.LogDebugLev).Infof("job %s no use npu", job.JobName)
		return false
	}
	if cardName != util.NPU310CardName {
		klog.V(util.LogDebugLev).Infof("job %s is not 910 type", job.JobName)
		return false
	}
	if job.IsJobOfCardMode() {
		klog.V(util.LogDebugLev).Infof("job %s is not card mode", job.JobName)
		return false
	}
	return true
}

func (reScheduler *ReScheduler) isChip310PJob(job plugin.SchedulerJob, cardName string) bool {
	if job.ReqNPUNum == 0 {
		klog.V(util.LogDebugLev).Infof("job %s no use npu", job.JobName)
		return false
	}
	if cardName != util.NPU310PCardName {
		klog.V(util.LogDebugLev).Infof("job %s is not 910 type", job.JobName)
		return false
	}
	if job.IsJobOfCardMode() {
		klog.V(util.LogDebugLev).Infof("job %s is not card mode", job.JobName)
		return false
	}
	return true
}

func (reScheduler *ReScheduler) updateNewFaultJobAttr(
	faultJob FaultJob, jobInfo *api.JobInfo, cardName string, cardPreName string) FaultJob {
	npuJob := reScheduler.Jobs[faultJob.JobUID] // 1. set the value of ReScheduleKey, grace/force/off
	tmpReScheduleKey := faultJob.GetJobFaultRescheduleLabel(&npuJob)
	faultJob.setJobFaultReScheduleLabel(tmpReScheduleKey)
	if tmpReScheduleKey == JobOffRescheduleLabelValue {
		klog.V(util.LogErrorLev).Infof("%s not set rescheduleLabel, no need reschedule.", jobInfo.Name)
		return faultJob
	}
	klog.V(util.LogInfoLev).Infof("%s set rescheduleLabel %v", jobInfo.Name, tmpReScheduleKey)
	// 2. create new FaultTask objects and update corresponding attributes
	tmpFaultTasks, err := reScheduler.createFaultTaskHandler(jobInfo, cardName)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s createFaultTaskHandler failed: %#v", jobInfo.Name, err)
	}
	faultJob.setFaultTasks(tmpFaultTasks)
	tmpNodeNames := faultJob.getJobUseNodes() // 3. update the value of Job used nodeNames
	faultJob.setNodeNames(tmpNodeNames)
	tmpIsFaultJob := faultJob.getIsFaultJob() // 4. update the value of IsFaultJob
	faultJob.setIsFaultJob(tmpIsFaultJob)
	// 4.1 additionally, jobs using nodes that are not ready in current session should also be set false
	for _, nodeName := range faultJob.NodeNames {
		if nodeName == "" {
			continue
		}
		_, ok := reScheduler.Nodes[nodeName]
		if !ok {
			klog.V(util.LogInfoLev).Infof("node %s used by job %s not in current session", nodeName, job.Name)
			faultJob.setIsFaultJob(true)
		}
	}
	if cardName == util.NPU910CardName { // 5. update JobRankIds of fault cards
		tmpJobRankIds := reScheduler.getJobRankIdsFromTasks(&faultJob, cardPreName)
		faultJob.setJobRankIds(tmpJobRankIds)
	}
	// 6. update FaultTypes of the job by status of FaultTasks bound on the job
	if faultJob.IsFaultJob {
		for _, fTask := range faultJob.FaultTasks {
			if fTask.IsFaultTask {
				faultJob.FaultTypes = append(faultJob.FaultTypes, fTask.faultType)
			}
		}
	}
	return faultJob
}

// AddFaultJobWithSession read all running jobs of given card types and create the corresponding FaultJob objects
func (reScheduler *ReScheduler) AddFaultJobWithSession(
	jobs map[api.JobID]*api.JobInfo, cardName string, cardPreName string) error {
	klog.V(util.LogInfoLev).Infof("enter AddFaultJobWithSession ... ")
	defer klog.V(util.LogInfoLev).Infof("leave AddFaultJobWithSession ... ")
	if reScheduler == nil {
		klog.V(util.LogErrorLev).Infof("AddFaultJobWithSession: %s, nil reScheduler", util.ArgumentError)
		return errors.New(util.ArgumentError)
	}
	if len(jobs) == 0 {
		klog.V(util.LogErrorLev).Infof("AddFaultJobWithSession: %s, nil job in session", util.ArgumentError)
		return errors.New(util.ArgumentError)
	}
	klog.V(util.LogDebugLev).Infof("ReSchedulerCache fault jobs before add: %#v", reScheduler.FaultJobs)
	nowTime := time.Now().Unix()
	for _, jobInfo := range jobs {
		klog.V(util.LogDebugLev).Infof("ReSchedulerCache considering job %s", jobInfo.Name)
		flagInCache := false
		for _, fJob := range reScheduler.FaultJobs {
			if fJob.JobName == jobInfo.Name {
				flagInCache = true
				break
			}
		}
		// 1. jobs already in cache: go through the continue logic
		if flagInCache {
			continue
		}
		// 2. create FaultJob objects for jobs not in cache but sent by session
		faultJob := newFaultJobDefault(jobInfo.Name, jobInfo.Namespace, jobInfo.UID, nowTime)
		faultJob = reScheduler.updateNewFaultJobAttr(faultJob, jobInfo, cardName, cardPreName)
		reScheduler.FaultJobs = append(reScheduler.FaultJobs, faultJob)
	}
	klog.V(util.LogDebugLev).Infof("ReSchedulerCache fault jobs after add: %#v", reScheduler.FaultJobs)
	return nil
}

// getJobRankIdsFromTasks record the rankIds of fault cards on nodes of the given job
func (reScheduler ReScheduler) getJobRankIdsFromTasks(fJob *FaultJob, cardName string) []string {
	var jobRankIds []string

	for _, fTask := range fJob.FaultTasks {
		nodeRankIndex, err := strconv.Atoi(fTask.NodeRankIndex)
		if err != nil {
			klog.V(util.LogInfoLev).Infof("%s setJobRankIdsFromTasks convert %#v", fTask.NodeRankIndex, err)
		}
		fNode := reScheduler.getFNodeOfGivenNameFromCache(fTask.NodeName)
		if fNode == nil {
			continue
		}
		taskUseFaultCards, err := fTask.getTaskUsedFaultCards(fNode)
		if err != nil {
			klog.V(util.LogErrorLev).Infof("taskUseFaultCards: %#v", err)
			continue
		}
		for _, taskUseFaultCard := range taskUseFaultCards {
			taskUseCardID, err := strconv.Atoi(strings.TrimPrefix(taskUseFaultCard, cardName))
			if err != nil {
				klog.V(util.LogErrorLev).Infof("convert card ID %s to int failed", taskUseFaultCard)
				continue
			}
			jobRankIds = append(jobRankIds, strconv.Itoa(taskUseCardID+nodeRankIndex*util.NPUIndex8))
		}
	}
	return jobRankIds
}

func (fTask *FaultTask) getTaskUsedFaultCards(fNode *FaultNode) ([]string, error) {
	if fTask.NodeName != fNode.NodeName {
		return nil, fmt.Errorf("fNode %s is not fTask %s's occupying node", fNode.NodeName, fTask.NodeName)
	}
	var taskUseFaultCard []string
	if fTask.faultType == NodeUnhealthy { // node unhealthy, return all cards
		return fTask.UseCardName, nil
	}
	for _, taskUseCard := range fTask.UseCardName {
		for _, fCard := range fNode.FaultCards {
			if taskUseCard == fCard.NPUName && fCard.IsFaultCard {
				taskUseFaultCard = append(taskUseFaultCard, taskUseCard)
			}
		}
	}
	return taskUseFaultCard, nil
}

// getGraceDeleteFaultJobs get jobs needed to be deleted gracefully, only fault jobs with grace label would be selected
func (reScheduler ReScheduler) getGraceDeleteFaultJobs() []FaultJob {
	var graceDeleteJobs []FaultJob
	for _, fJob := range reScheduler.FaultJobs {
		if !fJob.IsFaultJob || fJob.ReScheduleKey != JobGraceRescheduleLabelValue {
			continue
		}
		graceDeleteJobs = append(graceDeleteJobs, fJob)
	}
	return graceDeleteJobs
}

// GetNeedForceDeleteDelayingNPUJobs get fault jobs with grace label but haven't been evicted successfully
func (reScheduler *ReScheduler) GetNeedForceDeleteDelayingNPUJobs(
	schedulerJobs map[api.JobID]plugin.SchedulerJob, ssn *framework.Session) ([]plugin.SchedulerJob, error) {
	klog.V(util.LogInfoLev).Infof("enter GetNeedForceDeleteDelayingNPUJobs ... ")
	defer klog.V(util.LogInfoLev).Infof("leave GetNeedForceDeleteDelayingNPUJobs ... ")
	if reScheduler == nil {
		klog.V(util.LogErrorLev).Infof("GetNeedForceDeleteDelayingNPUJobs: %s, nil reScheduler",
			util.ArgumentError)
		return nil, errors.New(util.ArgumentError)
	}
	if len(schedulerJobs) == 0 {
		klog.V(util.LogErrorLev).Infof("GetNeedForceDeleteDelayingNPUJobs: %s, nil schedulerJobs",
			util.ArgumentError)
		return nil, errors.New(util.ArgumentError)
	}
	if ssn == nil {
		klog.V(util.LogErrorLev).Infof("GetNeedForceDeleteDelayingNPUJobs: %s, nil session",
			util.ArgumentError)
		return nil, errors.New(util.ArgumentError)
	}
	var forceJobs []plugin.SchedulerJob
	graceDeleteFaultJobs := reScheduler.getGraceDeleteFaultJobs()
	for _, fJob := range graceDeleteFaultJobs {
		jobInfo, ok := ssn.Jobs[fJob.JobUID] // if job in cache not in session, do not force delete
		if !ok {
			klog.V(util.LogErrorLev).Infof(
				"GetNeedForceDeleteDelayingNPUJobs %v not in ssn.Jobs.", fJob.JobName)
		}
		if fJob.isJobGraceDeleteSuccess(jobInfo) { // if job successfully restarted, do not force delete
			klog.V(util.LogErrorLev).Infof("%v grace deleted successful.", fJob.JobName)
			continue
		}
		if !reScheduler.isDelayingJobTimeout(&fJob) { // if job not restarted and not time out, do not force delete
			continue
		}
		klog.V(util.LogErrorLev).Infof("%v is time out for delete.", fJob.JobName)
		schedulerJob, ok := schedulerJobs[fJob.JobUID]
		if !ok {
			continue
		}
		forceJobs = append(forceJobs, schedulerJob)
	}
	if len(forceJobs) == 0 {
		klog.V(util.LogInfoLev).Infof("GetNeedForceDeleteDelayingNPUJobs get nil jobs.")
		return nil, errors.New("get none jobs")
	}
	return forceJobs, nil
}

func (reScheduler *ReScheduler) isDelayingJobTimeout(fJob *FaultJob) bool {
	nowTime := time.Now().Unix()
	createTime := fJob.JobRankIdCreateTime
	klog.V(util.LogErrorLev).Infof("isDelayingJobTimeOut now:%v create:%v.", nowTime, createTime)
	if nowTime-createTime > reScheduler.GraceDeleteTime {
		return true
	}
	return false
}

// New Initialisation of ReScheduler
func New(env *plugin.ScheduleEnv, jobType string) *ReScheduler {
	klog.V(util.LogInfoLev).Infof("Creating Fault ReScheduler for %s ...", jobType)
	defer klog.V(util.LogInfoLev).Infof("Finished creating Fault ReScheduler for %s ...", jobType)
	var faultReScheduler = ReScheduler{
		GraceDeleteTime:      0,
		DealReSchedulerCache: nil,
	}
	// 1. Initialise ReScheduler.graceDeleteTime
	klog.V(util.LogDebugLev).Infof("Initialising graceDeleteTime.")
	graceDeleteTime, err := faultReScheduler.GetGraceDeleteTime(env.FrameAttr.Conf)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("GetGraceDeleteTime %#v.", err)
	}
	faultReScheduler.setGraceOverTime(graceDeleteTime)
	// 2. Initialise ReScheduler.DealReSchedulerCache
	klog.V(util.LogDebugLev).Infof("Initialising ReSchedulerCache")
	reSchedulerCache := DealReSchedulerCache{
		FaultNodes: nil,
		FaultJobs:  nil,
	}
	// 2.1 Initialise ReScheduler.DealReSchedulerCache.Configmap
	klog.V(util.LogDebugLev).Infof("Initialising ReSchedulerCache.DealReSchedulerConfigmap")
	reSchedulerConfigmap := DealReSchedulerConfigmap{
		CMData: nil,
	}
	// 2.2 read configmap from kubernetes
	if cmErr := reSchedulerConfigmap.newReSchedulerCMFromEnv(env, jobType); cmErr != nil {
		klog.V(util.LogErrorLev).Infof("GetConfigmapFromKubernetes %#v", cmErr)
	}
	reSchedulerCache.DealReSchedulerConfigmap = &reSchedulerConfigmap
	// 2.2 Initialise ReScheduler.DealReSchedulerCache.FaultNodes by unmarshal data read from cm
	if setNodeErr := reSchedulerCache.SetFaultNodesFromCM(); setNodeErr != nil {
		klog.V(util.LogErrorLev).Infof("SetFaultNodesFromCM: %#v", setNodeErr)
	}
	// 2.3 Initialise ReScheduler.DealReSchedulerCache.NodeHeartbeats by unmarshal data read from cm
	if setHBErr := reSchedulerCache.SetNodeHeartbeatFromCM(); setHBErr != nil {
		klog.V(util.LogErrorLev).Infof("SetNodeHeartbeatFromCM: %#v", setHBErr)
	}
	// 2.4 Initialise ReScheduler.DealReSchedulerCache.AllocNodeRankOccurrenceMap by unmarshal data read from cm
	if jobType == CmFaultJob910x8Kind {
		if setNROErr := reSchedulerCache.SetNodeRankOccurrenceMapFromCM(); setNROErr != nil {
			klog.V(util.LogErrorLev).Infof("SetNodeRankOccurrenceMapFromCM: %#v", setNROErr)
		}
	} else {
		reSchedulerCache.setNodeRankOccurrenceMap(map[api.JobID][]AllocNodeRankOccurrence{})
	}
	faultReScheduler.DealReSchedulerCache = &reSchedulerCache // 2.4 set DealReSchedulerCache
	faultReScheduler.Jobs = env.Jobs                          // 3 Initialise session Jobs Nodes copying data from env
	faultReScheduler.Nodes = env.Nodes
	faultReScheduler.kubeClient = env.FrameAttr.KubeClient // 4 Initialise kubeClient copying data from env
	return &faultReScheduler
}

// New910ReScheduler initialise ReScheduler.FaultJobs for 910x8
func (reScheduler *ReScheduler) New910ReScheduler() {
	klog.V(util.LogInfoLev).Infof("Initialising New 910x8 fault scheduler fault jobs")
	defer klog.V(util.LogInfoLev).Infof("Finished initialising New 910x8 fault scheduler fault jobs")
	if reScheduler == nil {
		klog.V(util.LogErrorLev).Infof("New910ReScheduler: %s, nil reScheduler", util.ArgumentError)
		return
	}
	if setJobErr := reScheduler.DealReSchedulerCache.SetFaultJobsFromCM(CmFaultJob910x8Kind); setJobErr != nil {
		klog.V(util.LogErrorLev).Infof("SetFaultJobsFromCM: %#v", setJobErr)
	}
	return
}

// NewCommonReScheduler initialise ReScheduler.FaultJobs for non 910x8
func (reScheduler *ReScheduler) NewCommonReScheduler(jobType string) {
	klog.V(util.LogInfoLev).Infof("Initialising New %s fault scheduler fault jobs", jobType)
	defer klog.V(util.LogInfoLev).Infof("Finished initialising New %s fault scheduler fault jobs", jobType)
	if reScheduler == nil {
		klog.V(util.LogErrorLev).Infof("NewCommonReScheduler: %s, nil reScheduler", util.ArgumentError)
		return
	}
	if setJobErr := reScheduler.DealReSchedulerCache.SetFaultJobsFromCM(jobType); setJobErr != nil {
		klog.V(util.LogErrorLev).Infof("SetFaultJobsFromCM: %#v", setJobErr)
	}
	return
}

// SynCacheFaultNodeWithSession Synchronise FaultNodes in cache by updating the information using current session
func (reScheduler *ReScheduler) SynCacheFaultNodeWithSession(cardName string) {
	klog.V(util.LogInfoLev).Infof("enter SynCacheFaultNodeWithSession ...")
	defer klog.V(util.LogInfoLev).Infof("leave SynCacheFaultNodeWithSession ...")
	klog.V(util.LogInfoLev).Infof("ReSchedulerCache fault nodes before sync: %#v", reScheduler.FaultNodes)
	if reScheduler == nil {
		klog.V(util.LogErrorLev).Infof("SynCacheFaultNodeWithSession: %s, nil reScheduler", util.ArgumentError)
		return
	}
	var updatedFaultNodes []FaultNode
	for _, faultNode := range reScheduler.FaultNodes {
		// 1. nodes not in session should be kept in cache
		if !faultNode.isNodeInSessionByNpuNodes(reScheduler.Nodes) {
			klog.V(util.LogInfoLev).Infof("node %s in cache is not in session, keep without updating.",
				faultNode.NodeName)
			updatedFaultNodes = append(updatedFaultNodes, faultNode)
			continue
		}
		// 2. update attributes of cached FaultNodes utilising new information read from current session
		npuNode, _ := reScheduler.Nodes[faultNode.NodeName]
		// 2.1 read oldNodeHeartbeat value from cached nodeHeartbeat objects
		faultNode.setOldNodeHeartbeatTime(reScheduler.getLastNodeHeartbeatByNodeNameFromCache(npuNode.Name))
		// 2.2 update information sent by session NPUNodes
		faultNode.updateFaultNodesFromDeviceInfo(&npuNode, cardName)
		if err := faultNode.updateFaultNodesAttr(&npuNode); err != nil {
			klog.V(util.LogErrorLev).Infof("updateFaultNodesAttr: %#v", err)
		}
		updatedFaultNodes = append(updatedFaultNodes, faultNode)
	}
	reScheduler.setFaultNodes(updatedFaultNodes)
	klog.V(util.LogDebugLev).Infof("ReSchedulerCache fault nodes after sync: %#v", reScheduler.FaultNodes)
}

// SynCacheFaultJobWithSession Synchronise FaultJobs in cache by updating the information using current session
func (reScheduler *ReScheduler) SynCacheFaultJobWithSession(
	ssn *framework.Session, cardName string, cardPreName string) {
	klog.V(util.LogInfoLev).Infof("enter SynCacheFaultJobWithSession...")
	defer klog.V(util.LogInfoLev).Infof("leave SynCacheFaultJobWithSession...")
	if reScheduler == nil {
		klog.V(util.LogErrorLev).Infof("SynCacheFaultJobWithSession: %s, nil reScheduler", util.ArgumentError)
		return
	}
	var updatedFaultJobs []FaultJob
	nowTime := time.Now().Unix()
	for _, faultJob := range reScheduler.FaultJobs {
		// 1. cache Jobs exceeded max waiting time should be deleted and treated as normal new jobs
		if nowTime-faultJob.JobRankIdCreateTime > maxIntervalTime+reScheduler.GraceDeleteTime {
			klog.V(util.LogDebugLev).Infof("delete %s from CM for overTime %v => %v.",
				faultJob.JobName, nowTime, faultJob.JobRankIdCreateTime)
			continue
		}

		if !faultJob.isJobInSession(reScheduler.Jobs) {
			klog.V(util.LogDebugLev).Infof("faultJob name: %s not in session", faultJob.JobName)
			if !faultJob.CheckJobExistsInKubernetes(ssn) { // 1.1 delete jobs not in session or k8s
				klog.V(util.LogDebugLev).Infof(
					"delete %s from re-scheduler cache due to not existence in session and k8s.", faultJob.JobName)
				continue
			}
			faultJob.UpdateTime = nowTime
			faultJob.IsInSession = false
			updatedFaultJobs = append(updatedFaultJobs, faultJob) // 1.2 keep jobs not in session but in k8s
			continue
		}
		// 2. cache Jobs turned normal in session should be deleted ,meaning it has been restarted
		jobInfo, _ := ssn.Jobs[faultJob.JobUID]
		if faultJob.isJobGraceDeleteSuccess(jobInfo) {
			klog.V(util.LogErrorLev).Infof("%v grace deleted successful.", faultJob.JobName)
			if jobInfo.PodGroup.Status.Phase == scheduling.PodGroupRunning { // new job successfully running
				klog.V(util.LogErrorLev).Infof("job %s new pods running, delete from cache", jobInfo.Name)
				continue
			}
		}
		updatedFaultJobs = append(updatedFaultJobs, faultJob)
	}
	reScheduler.setFaultJobs(updatedFaultJobs)
	klog.V(util.LogDebugLev).Infof("ReSchedulerCache fault jobs sync: %#v", reScheduler.FaultJobs)
}

// SynCacheNodeRankOccMapWithSession Synchronise FaultJobs in cache by updating the information using current session
func (reScheduler *ReScheduler) SynCacheNodeRankOccMapWithSession() {
	klog.V(util.LogInfoLev).Infof("enter SynCacheNodeRankOccMapWithSession ...")
	defer klog.V(util.LogInfoLev).Infof("leave SynCacheNodeRankOccMapWithSession ...")
	if reScheduler == nil {
		klog.V(util.LogErrorLev).Infof("SynCacheNodeRankOccMapWithSession: %s, nil reScheduler",
			util.ArgumentError)
		return
	}
	klog.V(util.LogDebugLev).Infof("NodeRankOccMap before sync: %#v", reScheduler.AllocNodeRankOccurrenceMap)
	newNodeRankOccMap := make(map[api.JobID][]AllocNodeRankOccurrence, util.MapInitNum)
	for jobUID, NodeRankOcc := range reScheduler.AllocNodeRankOccurrenceMap {
		for _, fJob := range reScheduler.FaultJobs {
			if jobUID != fJob.JobUID {
				continue
			}
			if !fJob.IsFaultJob { // delete none faultJobs
				continue
			}
			newNodeRankOccMap[jobUID] = NodeRankOcc // only add faultJobs in the re-scheduling process
		}
	}
	reScheduler.AllocNodeRankOccurrenceMap = newNodeRankOccMap
	klog.V(util.LogDebugLev).Infof("NodeRankOccMap after sync: %#v", reScheduler.AllocNodeRankOccurrenceMap)
}

// AddFaultNodeWithSession Add FaultNode objects for new nodes in session not in cache
func (reScheduler *ReScheduler) AddFaultNodeWithSession(cardName string) {
	klog.V(util.LogInfoLev).Infof("enter AddFaultNodeWithSession ...")
	defer klog.V(util.LogInfoLev).Infof("leave AddFaultNodeWithSession ...")
	if reScheduler == nil {
		klog.V(util.LogErrorLev).Infof("AddFaultNodeWithSession: %s, nil reScheduler", util.ArgumentError)
		return
	}
	klog.V(util.LogDebugLev).Infof("ReSchedulerCache fault nodes before add: %#v", reScheduler.FaultNodes)
	newNodes := make(map[string]plugin.NPUNode, util.MapInitNum)
	nowTime := time.Now().Unix()
	for npuNodeName, npuNode := range reScheduler.Nodes {
		flag := false
		for _, fNode := range reScheduler.FaultNodes {
			if npuNodeName == fNode.NodeName {
				flag = true
				break
			}
		}
		if flag {
			klog.V(util.LogInfoLev).Infof("node %s is already in session", npuNodeName)
			continue // 1. skip nodes already in cached FaultNodes
		}
		newNodes[npuNodeName] = npuNode // create new for those not in cache
	}
	for name, npuNode := range newNodes {
		klog.V(util.LogInfoLev).Infof("Updating node from node name: %s, from node value: %s",
			name, npuNode.Name)
		// 0. Initialise faultNode
		faultNode := newFaultNodeDefault(npuNode.Name, nowTime)
		faultNode.OldHeartbeatTime = reScheduler.getLastNodeHeartbeatByNodeNameFromCache(npuNode.Name)
		faultNode.UpdateHeartbeatTime = reScheduler.getLastNodeHeartUpdateTimeByNodeNameFromCache(npuNode.Name)
		faultNode.updateFaultNodesFromDeviceInfo(&npuNode, cardName)
		if err := faultNode.updateFaultNodesAttr(&npuNode); err != nil {
			klog.V(util.LogErrorLev).Infof("node %s updateFaultNodesAttr err: %#v", npuNode.Name, err)
		}
		reScheduler.FaultNodes = append(reScheduler.FaultNodes, faultNode)
	}
	klog.V(util.LogDebugLev).Infof("ReSchedulerCache fault nodes after add: %#v", reScheduler.FaultNodes)
}

// RestartNeedForceDeleteJobs Restart jobs that need to be force deleted
func (reScheduler *ReScheduler) RestartNeedForceDeleteJobs(ssn *framework.Session) error {
	klog.V(util.LogInfoLev).Infof("enter RestartNeedForceDeleteJobs...")
	defer klog.V(util.LogInfoLev).Infof("leave RestartNeedForceDeleteJobs...")
	if reScheduler == nil {
		klog.V(util.LogErrorLev).Infof("RestartNeedForceDeleteJobs failed: %s, nil reScheduler",
			util.ArgumentError)
		return errors.New(util.ArgumentError)
	}
	if ssn == nil {
		klog.V(util.LogErrorLev).Infof("RestartNeedForceDeleteJobs: %s, nil session", util.ArgumentError)
		return errors.New(util.ArgumentError)
	}
	needDeleteNPUJobs, err := reScheduler.GetNeedForceDeleteDelayingNPUJobs(reScheduler.Jobs, ssn)
	if err != nil {
		return err
	}
	klog.V(util.LogDebugLev).Infof("GetNeedForceDeleteDelayingNPUJobs: %#v", needDeleteNPUJobs)
	for _, schedulerJob := range needDeleteNPUJobs {
		for _, faultJob := range reScheduler.FaultJobs {
			if schedulerJob.JobName != faultJob.JobUID {
				continue
			}
			if deleteErr := faultJob.ForceDeleteJob(ssn, &schedulerJob); deleteErr != nil {
				klog.V(util.LogErrorLev).Infof("%s ForceDeleteJob: %#v", schedulerJob.JobName, deleteErr)
			}
		}
	}
	return nil
}

// RestartFaultJobs Restart fault jobs by its corresponding strategy  grace,force,off
func (reScheduler *ReScheduler) RestartFaultJobs(ssn *framework.Session) error {
	klog.V(util.LogInfoLev).Infof("enter RestartFaultJobs...")
	defer klog.V(util.LogInfoLev).Infof("leave RestartFaultJobs...")
	if reScheduler == nil {
		klog.V(util.LogErrorLev).Infof("RestartFaultJobs failed: %s, nil reScheduler",
			util.ArgumentError)
		return errors.New(util.ArgumentError)
	}
	if ssn == nil {
		klog.V(util.LogErrorLev).Infof("RestartFaultJobs: %s, nil session", util.ArgumentError)
		return errors.New(util.ArgumentError)
	}
	// 1. Get fault jobs, only faultJobs that haven't been evicted yet should be put into list
	realFaultJobs, err := reScheduler.getRealFaultJobs()
	restartFaultJobs := reScheduler.getJobsToBeRestarted(realFaultJobs) // each job only triggers restart once
	var newCacheJobs []FaultJob
	var flag bool
	for _, fJob := range reScheduler.FaultJobs {
		flag = false
		for _, restartFJob := range restartFaultJobs {
			if fJob.JobName == restartFJob.JobName {
				flag = true
				break
			}
		}
		if !flag {
			newCacheJobs = append(newCacheJobs, fJob) // jobs no need to restart directly put back to cache FaultJobs
		}
	}
	if err != nil {
		return fmt.Errorf("restartFaultJobs: %#v", err)
	}
	klog.V(util.LogDebugLev).Infof("Jobs to be restarted: %#v", restartFaultJobs)
	// 2. Restart fault jobs
	for _, restartFaultJob := range restartFaultJobs {
		schedulerJob, ok := reScheduler.Jobs[restartFaultJob.JobUID]
		if !ok {
			klog.V(util.LogInfoLev).Infof("restartFaultJob %s not in session, has already been deleted",
				schedulerJob.JobName)
			continue
		}
		klog.V(util.LogInfoLev).Infof("%s need restart.", restartFaultJob.JobName)
		if restartErr := restartFaultJob.restartSingleFaultJob(
			ssn, reScheduler.kubeClient, &schedulerJob, jobRestartReason); restartErr != nil {
			klog.V(util.LogErrorLev).Infof("RestartJob %#v.", restartErr)
		} else {
			restartFaultJob.DeleteExecutedFlag = true
		}
		newCacheJobs = append(newCacheJobs, restartFaultJob) // modify restartFlag and put modified fJob into cache
	}
	reScheduler.setFaultJobs(newCacheJobs)
	return nil
}

// ScoreBestNPUNodes add scores on scoreMap for normal nodes used by re-scheduling tasks
func (reScheduler *ReScheduler) ScoreBestNPUNodes(task *api.TaskInfo, scoreMap map[string]float64) error {
	if reScheduler == nil {
		klog.V(util.LogErrorLev).Infof("ScoreBestNPUNodes: %s, nil reScheduler", util.ArgumentError)
		return errors.New(util.ArgumentError)
	}
	if task == nil {
		klog.V(util.LogErrorLev).Infof("ScoreBestNPUNodes: %s, nil task", util.ArgumentError)
		return errors.New(util.ArgumentError)
	}
	if len(scoreMap) == 0 {
		mgs := fmt.Errorf("add score by fault NPU task scoreMap is nil")
		klog.V(util.LogErrorLev).Infof("%#v.", mgs)
		return mgs
	}
	klog.V(util.LogDebugLev).Infof("enter rescheduling ScoreBestNPUNodes %s...", task.Name)
	klog.V(util.LogDebugLev).Infof("node score map before add rescheduling weights %#v", scoreMap)
	defer klog.V(util.LogDebugLev).Infof("leave rescheduling ScoreBestNPUNodes ...")
	curfTask := reScheduler.getFaultTaskOfGivenTaskNameFromCache(task.Name) // 1. get faultTask object
	if curfTask == nil {
		return fmt.Errorf("task %s is not in rescheduler cache", task.Name)
	}
	fJob := reScheduler.getFaultJobOfGivenTaskInfoFromCache(task) // 2. get faultJob object given the faultTask object
	if !fJob.IsFaultJob {                                         // skip adding re-scheduling score for normal jobs
		return fmt.Errorf("task %s belongs to job %s which is not a fault job",
			task.Name, fJob.JobName)
	}
	hasOldNodeFlag := false
	for nodeName, _ := range scoreMap {
		for _, jobUseNode := range fJob.NodeNames {
			if nodeName == jobUseNode {
				klog.V(util.LogDebugLev).Infof("node %s is previously used by job", nodeName)
				hasOldNodeFlag = true
			}
		}
	}
	if !hasOldNodeFlag {
		klog.V(util.LogDebugLev).Infof("no old node, no modifications on scoreMap")
		return nil
	}
	for nodeName, _ := range scoreMap {
		scoreMap[nodeName] = float64(0)
		for _, jobUseNode := range fJob.NodeNames {
			if nodeName == jobUseNode {
				klog.V(util.LogDebugLev).Infof("assign high score to old node %s", nodeName)
				scoreMap[nodeName] = float64(util.NPUIndex8 * util.NPUIndex8)
				break
			}
		}
	}
	return nil
}

// UseAnnotation reallocate rankIndex for the re-scheduling jobs
func (reScheduler *ReScheduler) UseAnnotation(task *api.TaskInfo, node *plugin.NPUNode) error {
	if reScheduler == nil {
		klog.V(util.LogErrorLev).Infof("UseAnnotation failed: %s, nil reScheduler",
			util.ArgumentError)
		return errors.New(util.ArgumentError)
	}
	if task == nil {
		klog.V(util.LogErrorLev).Infof("UseAnnotation failed: %s, nil task",
			util.ArgumentError)
		return errors.New(util.ArgumentError)
	}
	if node == nil {
		klog.V(util.LogErrorLev).Infof("UseAnnotation failed: %s, nil node",
			util.ArgumentError)
		return errors.New(util.ArgumentError)
	}
	klog.V(util.LogInfoLev).Infof("enter rescheduling UseAnnotation ...(%s, %s)", task.Name, node.Name)
	defer klog.V(util.LogInfoLev).Infof("leave rescheduling UseAnnotation ...(%s, %s)", task.Name, node.Name)
	if reScheduler.AllocNodeRankOccurrenceMap == nil || len(reScheduler.AllocNodeRankOccurrenceMap) == 0 {
		return nil
	}
	fJob := reScheduler.getFaultJobOfGivenTaskInfoFromCache(task)
	if fJob == nil {
		return fmt.Errorf("no fJob %s in reScheduler cache", task.Job)
	}
	nodeRankTimes := reScheduler.AllocNodeRankOccurrenceMap[fJob.JobUID]
	// 1. if given node is in the nodeRankTime, keep it ,node is used by the fault job before
	for _, nodeRankTime := range nodeRankTimes {
		if node.Name == nodeRankTime.NodeName {
			klog.V(util.LogInfoLev).Infof("set old node rank index <%s>/<%s>/<%s>", node.Name,
				task.Name, nodeRankTime.RankIndex)
			task.Pod.Annotations[podRankIndex] = nodeRankTime.RankIndex
			return nil
		}
	}
	// 2. node not in the nodeRankTime, the fault nodes' rankIndex previously used should be assigned to new nodes
	newNodeRankTimes := reScheduler.useAnnotationSetNewNodeRank(nodeRankTimes, node, task)
	reScheduler.AllocNodeRankOccurrenceMap[fJob.JobUID] = newNodeRankTimes
	return nil
}

func (reScheduler *ReScheduler) useAnnotationSetNewNodeRank(nodeRankTimes []AllocNodeRankOccurrence,
	node *plugin.NPUNode, task *api.TaskInfo) []AllocNodeRankOccurrence {
	var newNodeRankTimes []AllocNodeRankOccurrence
	bindNodeFlag := false
	for _, nodeRankTime := range nodeRankTimes { // check nodes, corresponding index recorded to be used by faultJob
		fNode := reScheduler.getFNodeOfGivenNameFromCache(nodeRankTime.NodeName) // faultNode object given nodeName
		if fNode == nil {
			klog.V(util.LogDebugLev).Infof("node %s not in cache", node.Name)
			newNodeRankTimes = append(newNodeRankTimes, nodeRankTime)
			continue
		}
		if !fNode.IsFaultNode { // if node is not faultNode, the rankIndex shouldn't be occupied
			klog.V(util.LogDebugLev).Infof("node %s is not fault node", fNode.NodeName)
			newNodeRankTimes = append(newNodeRankTimes, nodeRankTime)
			continue
		}
		if nodeRankTime.Occurrence == 1 { // if node is faultNode but rankIndex already assigned to other new nodes
			klog.V(util.LogDebugLev).Infof("rankIndex %s occupied by other node %s",
				nodeRankTime.RankIndex, nodeRankTime.NodeName)
			newNodeRankTimes = append(newNodeRankTimes, nodeRankTime)
			continue
		}
		if bindNodeFlag { // if old node in nodeRankIndex reach this stage, its rIndex will be assigned to cur node
			newNodeRankTimes = append(newNodeRankTimes, nodeRankTime)
			break
		}
		nodeRankTime.Occurrence++ // indicate the rankIndex has been bound so should not be used by other nodes
		newNodeRankTimes = append(newNodeRankTimes, nodeRankTime)
		klog.V(util.LogDebugLev).Infof(
			"Assigning rankIndex %s to node %s...(use annotation task %s)",
			nodeRankTime.RankIndex, node.Name, task.Name)
		task.Pod.Annotations[podRankIndex] = nodeRankTime.RankIndex
		bindNodeFlag = true
	}
	return newNodeRankTimes
}

// GenerateNodeRankIndexTaskMap get the nodeName, rankIndex, and Occurrence of nodes in a job
func (reScheduler *ReScheduler) GenerateNodeRankIndexTaskMap() {
	klog.V(util.LogInfoLev).Infof("enter GenerateNodeRankIndexTaskMap ...")
	defer klog.V(util.LogInfoLev).Infof("leave GenerateNodeRankIndexTaskMap ...")
	if reScheduler == nil {
		klog.V(util.LogErrorLev).Infof("GenerateNodeRankIndexTaskMap failed: %s, nil reScheduler",
			util.ArgumentError)
		return
	}
	nodeRankIndexTaskMap := make(map[api.JobID][]AllocNodeRankOccurrence, util.MapInitNum)
	for _, fJob := range reScheduler.FaultJobs {
		if fJob.DeleteExecutedFlag {
			var nodeRankTimes []AllocNodeRankOccurrence
			for _, fTask := range fJob.FaultTasks {
				nodeRankTime := AllocNodeRankOccurrence{
					NodeName:   fTask.NodeName,
					RankIndex:  fTask.NodeRankIndex,
					Occurrence: 0,
				}
				nodeRankTimes = append(nodeRankTimes, nodeRankTime)
			}
			nodeRankIndexTaskMap[fJob.JobUID] = nodeRankTimes
		}
	}
	reScheduler.AllocNodeRankOccurrenceMap = nodeRankIndexTaskMap
}

// CheckNodeNPUByTask used in the predicate process of task and node
func (reScheduler *ReScheduler) CheckNodeNPUByTask(task *api.TaskInfo, vcNode plugin.NPUNode) error {
	if reScheduler == nil {
		klog.V(util.LogErrorLev).Infof("CheckNodeNPUByTask failed: %s, nil reScheduler",
			util.ArgumentError)
		return errors.New(util.ArgumentError)
	}
	if task == nil {
		klog.V(util.LogErrorLev).Infof("CheckNodeNPUByTask failed: %s, nil task",
			util.ArgumentError)
		return errors.New(util.ArgumentError)
	}
	if len(vcNode.Name) == 0 {
		klog.V(util.LogErrorLev).Infof("CheckNodeNPUByTask failed: %s, nil node",
			util.ArgumentError)
		return errors.New(util.ArgumentError)
	}
	klog.V(util.LogDebugLev).Infof("enter rescheduling CheckNodeNPUByTask ...(%s, %s)", task.Name, vcNode.Name)
	defer klog.V(util.LogDebugLev).Infof("leave rescheduling CheckNodeNPUByTask ...(%s, %s)",
		task.Name, vcNode.Name)
	curFTask := reScheduler.getFaultTaskOfGivenTaskNameFromCache(task.Name)
	curFJob := reScheduler.getFaultJobOfGivenTaskInfoFromCache(task)
	curSchedulerJob := reScheduler.getSchedulerJobOfGivenUIDFromReScheduler(task.Job)
	if curFTask == nil {
		klog.V(util.LogDebugLev).Infof("task %s not in rescheduler cache", task.Name)
		return nil // cannot return error, or will be stuck for new job scheduling
	}
	if curFJob == nil {
		return fmt.Errorf("task %s does not have corresponding job in cache", task.Name)
	}
	// 0. if fTask's corresponding faultJobs' previously used normal nodes haven't be released,
	// stuck all scheduling process
	if err := reScheduler.checkNodeFJobNormNodeRelease(curFJob, curSchedulerJob); err != nil {
		return err
	}
	// 1. jobs should not be scheduled to faultNodes
	if err := reScheduler.checkNodeCurNodeIsFault(curFJob, vcNode, task); err != nil {
		return err
	}
	// 3. non faultJobs should not occupy normal nodes previously used by distributional
	if err := reScheduler.checkNodeNewJobUseFJobNormNode(curFTask, vcNode, task); err != nil {
		return err
	}
	// 4. faultJobs should not be assigned to nodes' whose rankIndex already occupied by some other new nodes
	if err := reScheduler.checkNodeRankIndexOccupied(curFJob, vcNode); err != nil {
		return err
	}
	klog.V(util.LogDebugLev).Infof("CheckNodeNPUByTask node %s passed rescheduling predicate for task %s",
		vcNode.Name, task.Name)
	return nil
}

// 0. stuck scheduling as long as normal nodes used by re-scheduling jobs not released
func (reScheduler *ReScheduler) checkNodeFJobNormNodeRelease(
	curFJob *FaultJob, curSchedulerJob plugin.SchedulerJob) error {
	if !curFJob.IsFaultJob { // if current job is not in faultJob cache, skip
		klog.V(util.LogDebugLev).Infof("job %s is not a fault job", curFJob.JobName)
		return nil
	}
	for _, fTask := range curFJob.FaultTasks {
		fNode := reScheduler.getFNodeOfGivenNameFromCache(fTask.NodeName)
		if fNode == nil { // if fNode is nil then the node should be a fault one
			klog.V(util.LogDebugLev).Infof("node %s does not exist in cache", fTask.NodeName)
			continue
		}
		if !fNode.IsFaultNode { // if normal node in faultJob hasn't been released, return error
			if !fNode.isNodeInSessionByNpuNodes(reScheduler.Nodes) { // case1. node not released and not sent by ssn
				return fmt.Errorf("cache fault <job/task>: <%s/%s>  "+
					"normal node %s hasn't been release, waiting for next session",
					curFJob.JobName, fTask.TaskName, fTask.NodeName)
			}
			npuNode := reScheduler.getNPUNodeOfGiveNodeNameFromReScheduler(fNode.NodeName)
			if err := npuNode.CheckNPUResourceStableReScheduling(curSchedulerJob); err != nil {
				return fmt.Errorf("cache fault <job/task>: <%s/%s>  normal node %s "+
					"resource still unstable, waiting for next session: %#v",
					curFJob.JobName, fTask.TaskName, fTask.NodeName, err) // case2.node sent by ssn but still unstable
			}
		}
	}
	klog.V(util.LogDebugLev).Infof("checkNodeFJobNormNodeRelease: check ok, fault job %s task length: %d",
		curFJob.JobName, len(curFJob.FaultTasks))
	return nil
}

// 1. fault nodes should not be bound to any task
func (reScheduler *ReScheduler) checkNodeCurNodeIsFault(
	curFJob *FaultJob, vcNode plugin.NPUNode, task *api.TaskInfo) error {
	if len(curFJob.FaultTasks) == 1 {
		return nil
	}
	for _, fNode := range reScheduler.FaultNodes {
		if vcNode.Name == fNode.NodeName && fNode.IsFaultNode {
			if len(curFJob.FaultTasks) > 1 {
				return fmt.Errorf("task %s cannot be assigned to node %s because it's in faultNode list",
					task.Name, vcNode.Name)
			}
			if fNode.NodeHealthState == NodeCardNetworkUnhealthy {
				return nil
			}
			return fmt.Errorf("task %s cannot be assigned to node %s "+
				"because it neither healthy nor networkUnhealthy", task.Name, vcNode.Name)
		}
	}
	return nil
}

// 2. new jobs cannot take normal nodes used by old distributional jobs
func (reScheduler *ReScheduler) checkNodeNewJobUseFJobNormNode(
	curFTask *FaultTask, vcNode plugin.NPUNode, task *api.TaskInfo) error {
	realFaultJobs, err := reScheduler.getRealFaultJobs()
	if err != nil {
		klog.V(util.LogDebugLev).Infof("none real fault jobs")
		return nil
	}
	// 3. non faultJobs should not occupy normal nodes previously used by distributional
	// faultJobs in the re-scheduling process
	for _, fJob := range realFaultJobs {
		for _, fJobUseNode := range fJob.NodeNames {
			if fJobUseNode == vcNode.Name && curFTask.JobName != fJob.JobName && len(fJob.FaultTasks) > 1 {
				klog.V(util.LogDebugLev).Infof("task %s cannot use node %s occupied by faultJob %s",
					task.Name, vcNode.Name, fJob.JobName)
				return fmt.Errorf("task %s cannot use node %s occupied by faultJob %s",
					task.Name, vcNode.Name, fJob.JobName)
			}
		}
	}
	return nil
}

// 3. check if current node's rankIndex is occupied already
func (reScheduler *ReScheduler) checkNodeRankIndexOccupied(curFJob *FaultJob, vcNode plugin.NPUNode) error {
	nodeRankTimes, ok := reScheduler.AllocNodeRankOccurrenceMap[curFJob.JobUID]
	if ok {
		for _, nodeRankTime := range nodeRankTimes {
			if vcNode.Name == nodeRankTime.NodeName && nodeRankTime.Occurrence != 0 {
				return fmt.Errorf("fault node %s rankIndex %s cannot be assigned to node %s, "+
					"since it has been assigned to other node with occurrence %d",
					nodeRankTime.NodeName, nodeRankTime.RankIndex, vcNode.Name, nodeRankTime.Occurrence)
			}
		}
	}
	return nil
}

func (reScheduler ReScheduler) getFaultTaskOfGivenTaskNameFromCache(taskName string) *FaultTask {
	for _, fJob := range reScheduler.FaultJobs {
		for _, fTask := range fJob.FaultTasks {
			if fTask.TaskName == taskName {
				return &fTask
			}
		}
	}
	return nil
}

func (reScheduler ReScheduler) getNPUNodeOfGiveNodeNameFromReScheduler(nodeName string) *plugin.NPUNode {
	for _, npuNode := range reScheduler.Nodes {
		if npuNode.Name == nodeName {
			return &npuNode
		}
	}
	return nil
}

func (reScheduler ReScheduler) getSchedulerJobOfGivenUIDFromReScheduler(jobUID api.JobID) plugin.SchedulerJob {
	return reScheduler.Jobs[jobUID]
}

func (reScheduler ReScheduler) getFaultJobOfGivenTaskInfoFromCache(taskInfo *api.TaskInfo) *FaultJob {
	for _, fJob := range reScheduler.FaultJobs {
		if fJob.JobUID == taskInfo.Job {
			return &fJob
		}
	}
	return nil
}

func (reScheduler ReScheduler) getLastNodeHeartbeatByNodeNameFromCache(nodeName string) int64 {
	for _, nodeHB := range reScheduler.NodeHeartbeats {
		if nodeHB.NodeName == nodeName {
			klog.V(util.LogDebugLev).Infof("getLastNodeHeartbeatByNodeNameFromCache: %s, %d",
				nodeName, nodeHB.HeartbeatTime)
			return nodeHB.HeartbeatTime
		}
	}
	return 0
}

func (reScheduler ReScheduler) getLastNodeHeartUpdateTimeByNodeNameFromCache(nodeName string) int64 {
	for _, nodeHB := range reScheduler.NodeHeartbeats {
		if nodeHB.NodeName == nodeName {
			klog.V(util.LogDebugLev).Infof("getLastNodeHeartbeatByNodeNameFromCache: %s, %d",
				nodeName, nodeHB.HeartbeatTime)
			return nodeHB.UpdateTime
		}
	}
	return 0
}

// getTaskHealthState return true when unhealthy
func (reScheduler ReScheduler) getTaskHealthState(fTask *FaultTask) (bool, string) {
	klog.V(util.LogDebugLev).Infof("task %s getTaskHealthState", fTask.TaskName)
	var nodeUseCardHealthState []string
	realFaultNode := reScheduler.getRealFaultNodes()
	if fTask.NodeName == "" {
		return false, NodeHealthy // tasks has not yet been scheduled
	}
	_, ok := reScheduler.Nodes[fTask.NodeName] // task used node not in session
	if !ok {
		return true, NodeUnhealthy
	}
	for _, fNode := range realFaultNode {
		if fNode.NodeName == fTask.NodeName {
			if !fNode.IsFaultNode { // if task used node isFaultNode is false, return healthy
				return false, NodeHealthy
			}
			if fNode.NodeHealthState == NodeUnhealthy { // if task used node is nodeUnhealthy, return
				return true, NodeUnhealthy
			}
			nodeUseCardHealthState = fTask.getTaskUseFaultCardHealthState(&fNode) // get fault NPUs on task used node
		}
	}
	if util.IsSliceContain(NodeCardUnhealthy, nodeUseCardHealthState) { // if has unhealthy npu, return in advance
		return true, NodeCardUnhealthy
	}
	if util.IsSliceContain(NodeCardNetworkUnhealthy, nodeUseCardHealthState) { // if has networkUnhealthy npu, return
		return true, NodeCardNetworkUnhealthy
	}
	return false, NodeHealthy
}

func (reScheduler ReScheduler) getJobsToBeRestarted(realFaultJobs []FaultJob) []FaultJob {
	var restartFaultJobs []FaultJob
	for _, fJob := range realFaultJobs {
		if fJob.DeleteExecutedFlag {
			continue
		}
		restartFaultJobs = append(restartFaultJobs, fJob)
	}
	return restartFaultJobs
}

func (reScheduler ReScheduler) getFNodeOfGivenNameFromCache(nodeName string) *FaultNode {
	for _, fNode := range reScheduler.FaultNodes {
		if fNode.NodeName == nodeName {
			return &fNode
		}
	}
	return nil
}
