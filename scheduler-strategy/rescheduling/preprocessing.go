/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package rescheduling is using for HuaWei Ascend pin fault rescheduling.

*/
package rescheduling

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/util"
)

func init() {
	ReSchedulerCache = make(map[string]interface{}, util.NPUIndex3)
	reSchedulerJobController = map[string]struct{}{
		JobGraceRescheduleLabelValue: {},
		JobForceRescheduleLabelValue: {},
		JobOffRescheduleLabelValue:   {}}
}

// Delete expired data.
func updateReSchedulerDataFromSession(ssn *framework.Session) error {
	for dataID, tmpValue := range ReSchedulerCache {
		switch dataID {
		case CmJobKind:
			if err := synReSchedulerJobCache(ssn, tmpValue); err != nil {
				klog.V(util.LogErrorLev).Infof("synReSchedulerJobCache %v.", err)
			}
		case CmNodeKind:
			if err := synReSchedulerNodeCache(ssn, tmpValue); err != nil {
				klog.V(util.LogErrorLev).Infof("synReSchedulerNodeCache %v.", err)
			}
		case CmCardKind:
			if err := synReSchedulerCardCache(ssn, tmpValue); err != nil {
				klog.V(util.LogErrorLev).Infof("synReSchedulerCardCache %v.", err)
			}
		case CmNodeHeartbeatKind:
			if err := synNodeHeartbeatCache(ssn, tmpValue); err != nil {
				klog.V(util.LogErrorLev).Infof("synNodeHeartbeatCache %v.", err)
			}
		case CmJobRankIds:
			if err := synJobRankIdsCache(ssn, tmpValue); err != nil {
				klog.V(util.LogErrorLev).Infof("synJobRankIdsCache %v.", err)
			}
		default:
			klog.V(util.LogErrorLev).Infof("not support %v:%v.", dataID, tmpValue)
		}
	}
	// 2.add new node
	if err := updateNodeIntoNodesHeartbeatTmp(ssn); err != nil {
		klog.V(util.LogErrorLev).Infof("updateNodeIntoNodesHeartbeatTmp %v.", err)
	}
	return nil
}

func getJobUsedNodeRankIds(job *api.JobInfo, nodeAndPods map[string]*v1.Pod) (
	map[api.JobID]FaultRankIDRecordJobCMData, error) {
	nodeRankIds := make(map[api.JobID]FaultRankIDRecordJobCMData, util.NPUIndex3)
	var PodsName, rankIds []string
	var PodsUID []types.UID
	var PodsCreatTime []int64
	for _, task := range job.Tasks {
		tmpPod, ok := nodeAndPods[task.NodeName]
		if !ok {
			continue
		}
		rankIndexStr, err := getPodRankIndex(tmpPod)
		if err != nil {
			klog.V(util.LogErrorLev).Infof("getPodRankIndex %s %v.", tmpPod.Name, err)
			continue
		}
		rankIndex, covertErr := strconv.Atoi(rankIndexStr)
		if covertErr != nil {
			klog.V(util.LogErrorLev).Infof("%s getJobUsedNodeRankIds covert %v.", covertErr, rankIndexStr)
			continue
		}
		taskUseNPUsStr, err := getPodUsedNPUS(tmpPod, npu800And9000CardName, node910X8NPUNum)
		if err != nil {
			klog.V(util.LogInfoLev).Infof("getPodUsedNPUS %v.", err)
			continue
		}
		taskUseNPUs := util.ChangeTopToIntArray(strings.Join(taskUseNPUsStr, ","), npu800And9000CardPreName)
		for _, tmp := range taskUseNPUs {
			rankIds = append(rankIds, strconv.Itoa(tmp+rankIndex*node910X8NPUNum))
		}
		PodsName = append(PodsName, tmpPod.Name)
		PodsUID = append(PodsUID, tmpPod.UID)
		PodsCreatTime = append(PodsCreatTime, tmpPod.CreationTimestamp.Unix())
	}
	dataBuffer, err := json.Marshal(rankIds)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("marshalCacheDataToString err: %v.", err)
		return nil, err
	}
	tmpData := FaultRankIDRecordJobCMData{
		NameSpace:     job.Namespace,
		FaultRankIds:  string(dataBuffer),
		PodsName:      PodsName,
		PodsUID:       PodsUID,
		PodsCreatTime: PodsCreatTime,
		CreatTime:     time.Now().Unix(),
	}
	nodeRankIds[job.UID] = tmpData
	return nodeRankIds, nil
}

func addJobsRankIdsIntoRankIds(jobsRankIds map[api.JobID]FaultRankIDRecordJobCMData) error {
	jobsRankIdsFromCache, getErr := getRankIDJobsFromCache()
	if getErr != nil {
		klog.V(util.LogErrorLev).Infof("addJobsRankIdsIntoCache not contain node fault :%v.", getErr)
		jobsRankIdsFromCache = make(map[api.JobID]FaultRankIDRecordJobCMData, util.NPUIndex3)
	}

	for jobID, rankIDData := range jobsRankIds {
		jobsRankIdsFromCache[jobID] = rankIDData
	}

	ReSchedulerCache[CmJobRankIds] = jobsRankIdsFromCache
	return nil
}

// For last words by fault nodes, the NPUs need record.
func addJobUseNodeNPUsIntoCard(nodeAndPods map[string]*v1.Pod) error {
	allFaultNodes, err := getFaultNodesFromCache()
	if err != nil {
		klog.V(util.LogErrorLev).Infof("getFaultNodesFromCache %v.", err)
		return err
	}
	tmpFaultNPUsMap := make(map[string]FaultNPUsOnNode, util.NPUIndex3)
	for nodeName, pod := range nodeAndPods {
		if _, ok := allFaultNodes[nodeName]; !ok {
			continue
		}
		cards, getErr := getPodUsedNPUS(pod, npu800And9000CardName, node910X8NPUNum)
		if getErr != nil {
			klog.V(util.LogErrorLev).Infof("getPodUsedNPUS %v.", getErr)
			continue
		}
		tempFaultNPUsOnNode := FaultNPUsOnNode{NodeName: nodeName, FaultNPUs: cards, NetworkUnhealthyNPUs: nil,
			UpdateTime: time.Now().Unix(),
		}
		tmpFaultNPUsMap[nodeName] = tempFaultNPUsOnNode
	}

	ReSchedulerCache[CmCardKind] = tmpFaultNPUsMap
	klog.V(util.LogErrorLev).Infof("addJobUseNodeNPUsIntoCard after :%+v.", tmpFaultNPUsMap)
	return nil
}

func writeFaultNodeRankIdsByJobInCache(ssn *framework.Session) error {
	if len(ssn.Jobs) == 0 {
		klog.V(util.LogErrorLev).Infof("writeFaultNodeRankIdsByJobInCache none jobs in ssn.")
		return nil
	}
	for _, job := range ssn.Jobs {
		nodeAndPods, getErr := getRunningJobUsedNodes(job)
		if getErr != nil {
			klog.V(util.LogErrorLev).Infof("%s getRunningJobUsedNodes %v.", job.Name, getErr)
			continue
		}
		if !isJobHasFaultNodes(nodeAndPods) {
			klog.V(util.LogErrorLev).Infof("%s isJobHasFaultNodes .", job.Name)
			continue
		}
		jobsRankIds, getRankIdsErr := getJobUsedNodeRankIds(job, nodeAndPods)
		if getRankIdsErr != nil {
			klog.V(util.LogErrorLev).Infof("%s getJobUsedNodeRankIds %+v.", job.Name, getRankIdsErr)
			continue
		}
		if addErr := addJobsRankIdsIntoRankIds(jobsRankIds); addErr != nil {
			klog.V(util.LogErrorLev).Infof("%s addJobsRankIdsIntoCache %v %+v.", job.Name, jobsRankIds, addErr)
			continue
		}
		if addNodeErr := addJobUseNodeNPUsIntoCard(nodeAndPods); addNodeErr != nil {
			klog.V(util.LogErrorLev).Infof("%s addJobUseNodeNPUsIntoCard %v.", job.Name, addNodeErr)
			continue
		}
	}
	return nil
}

// Write fault resource(NPUs,nodes) into cache.
func writeFaultResourceInfInCache(ssn *framework.Session, npus []FaultNPUsOnNode, nodes []FaultNodeState) error {
	// 1.Write fault NPU cards into cache.
	var cardMap = make(map[string]FaultNPUsOnNode, 1)
	for _, card := range npus {
		tmp := card
		cardMap[card.NodeName] = tmp
	}
	ReSchedulerCache[CmCardKind] = cardMap
	// 2.Write fault NPU nodes into cache.
	var nodeMap = make(map[string]FaultNodeState, 1)
	for _, nodeState := range nodes {
		tmp := nodeState
		nodeMap[nodeState.NodeName] = tmp
	}
	ReSchedulerCache[CmNodeKind] = nodeMap
	// 3.Writes the chip involved in the failed node to the cache.
	if writeRankIdsErr := writeFaultNodeRankIdsByJobInCache(ssn); writeRankIdsErr != nil {
		klog.V(util.LogErrorLev).Infof("writeFaultNodeRankIdsByJobInCache %v.", writeRankIdsErr)
		return writeRankIdsErr
	}

	return nil
}

// RecordFaultInfInCache Record the fault information(card/node) in the cache
func RecordFaultInfInCache(ssn *framework.Session, npuNumber int) error {
	// 1.Get fault NPUs and its nodes from running vcJob.
	faultNPUs, npuErr := getInoperableNPUCards(ssn.Nodes, npuNumber)
	if npuErr != nil {
		klog.V(util.LogErrorLev).Infof("getInoperableNPUCards %v.", npuErr)
	}
	// 2.Obtaining the Faulty Node from nodeD.
	faultNodes, nodeErr := getInoperableNodes(ssn.Nodes)
	if nodeErr != nil {
		klog.V(util.LogErrorLev).Infof("getInoperableNodes %v.", nodeErr)
	}

	if npuErr != nil && nodeErr != nil {
		return fmt.Errorf("%v %v", npuErr, nodeErr)
	}

	if writeErr := writeFaultResourceInfInCache(ssn, faultNPUs, faultNodes); writeErr != nil {
		klog.V(util.LogErrorLev).Infof("writeFaultResourceInfInCache %v.", writeErr)
		return writeErr
	}
	return nil
}

func isGraceDeleteJob(job *api.JobInfo) bool {
	label, getErr := GetJobFaultRescheduleLabel(job)
	if getErr != nil {
		return false
	}
	if label == JobGraceRescheduleLabelValue {
		return true
	}
	return false
}

func makeReSchedulerTasksByFaultNPUJob(tmpFaultNPUJob FaultNPUJob, jobInf *api.JobInfo) (ReSchedulerTasks, error) {
	tmpTask := ReSchedulerTasks{
		NameSpace:     tmpFaultNPUJob.namespace,
		GraceTaskFlag: isGraceDeleteJob(jobInf),
		UpdateTime:    time.Now().Unix()}

	for taskName, nodeName := range tmpFaultNPUJob.taskUseNode {
		rankIndex, indexOK := tmpFaultNPUJob.taskUseRankIndex[taskName]
		if !indexOK {
			klog.V(util.LogErrorLev).Infof("%s get rankIndex failed.", taskName)
			continue
		}

		useNPUs, npuOK := tmpFaultNPUJob.taskUseNPUs[taskName]
		if !npuOK {
			klog.V(util.LogErrorLev).Infof("%s get use NPUs failed.", taskName)
			continue
		}

		tmpTask.TaskName = append(tmpTask.TaskName, taskName)
		tmpTask.NodeNames = append(tmpTask.NodeNames, nodeName)
		tmpTask.TaskUseNPUs = append(tmpTask.TaskUseNPUs, useNPUs)
		tmpTask.RankIndexes = append(tmpTask.RankIndexes, rankIndex)
		tmpTask.Time = append(tmpTask.Time, time.Now().Unix())
	}
	return tmpTask, nil
}

// SetFaultInNodeAndJobs Recorded the information about the faulty task in the cache.
func SetFaultInNodeAndJobs(ssn *framework.Session, fNPUJobs []FaultNPUJob, jobs map[string]*api.JobInfo) error {
	for _, tmpFaultNPUJob := range fNPUJobs {
		jobInf, ok := jobs[tmpFaultNPUJob.jobName]
		if !ok {
			klog.V(util.LogErrorLev).Infof("SetFaultInNodeAndJobs %s not found in fault jobs", tmpFaultNPUJob.jobName)
			continue
		}
		tmpTask, makeErr := makeReSchedulerTasksByFaultNPUJob(tmpFaultNPUJob, jobInf)
		if makeErr != nil {
			klog.V(util.LogErrorLev).Infof("%s SetFaultInNodeAndJobs %v", jobInf.Name, makeErr)
			continue
		}
		if err := writeFaultJobInfInCache(jobInf, tmpTask); err != nil {
			klog.V(util.LogErrorLev).Infof("SetFaultInNodeAndJobs :%v.", err)
			continue
		}
		// 3.Record into job-fault-configMap.
		if cmErr := WriteJobFaultRankIDIntoCacheAndCM(ssn, jobInf); cmErr != nil {
			klog.V(util.LogErrorLev).Infof("SetFaultInNodeAndJobs %v.", cmErr)
		}
	}
	klog.V(util.LogErrorLev).Infof("SetFaultInNodeAndJobs after %+v.", ReSchedulerCache)
	return nil
}

func getCMWriteDate(reSchedulerData map[string]interface{}) map[string]string {
	var data = make(map[string]string, 1)
	for dataKind, faultData := range reSchedulerData {
		switch dataKind {
		case CmJobKind:
			jobData, err := getCMJobWriteData(faultData)
			if err != nil {
				klog.V(util.LogErrorLev).Infof("getCMJobWriteData :%v.", err)
				continue
			}
			data[CmJobKind] = jobData
		case CmNodeKind:
			nodeData, err := getCMNodeWriteData(faultData)
			if err != nil {
				klog.V(util.LogErrorLev).Infof("getCMJobWriteData :%v.", err)
				continue
			}
			data[CmNodeKind] = nodeData
		case CmCardKind:
			cardData, err := getCMCardWriteData(faultData)
			if err != nil {
				klog.V(util.LogErrorLev).Infof("getCMCardWriteData :%v.", err)
				continue
			}
			data[CmCardKind] = cardData
		case CmNodeHeartbeatKind:
			heartbeatData, err := getCMHeartbeatWriteData(faultData)
			if err != nil {
				klog.V(util.LogErrorLev).Infof("getCMHeartbeatWriteData :%v.", err)
				continue
			}
			data[CmNodeHeartbeatKind] = heartbeatData
		case CmJobRankIds:
			rankIdsData, err := getCMRankIdsWriteData(faultData)
			if err != nil {
				klog.V(util.LogErrorLev).Infof("getCMRankIdsWriteData :%v.", err)
				continue
			}
			data[CmJobRankIds] = rankIdsData
		default:
			klog.V(util.LogErrorLev).Infof("getCMWriteDate not support %v.", dataKind)
		}
	}

	return data
}

// WriteReSchedulerDataToCM Write FaultNPUJobs into ConfigMap.
func WriteReSchedulerDataToCM(ssn *framework.Session, reSchedulerData map[string]interface{}) error {
	var faultNPUConfigMap = &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cmName,
			Namespace: cmNameSpace,
		},
		Data: getCMWriteDate(reSchedulerData),
	}

	klog.V(util.LogErrorLev).Infof("Write faultNPUJobs into cm: %+v/%v.",
		faultNPUConfigMap.Namespace, faultNPUConfigMap.Name)
	if err := util.CreateOrUpdateConfigMap(ssn.KubeClient(), faultNPUConfigMap, cmName, cmNameSpace); err != nil {
		klog.V(util.LogErrorLev).Infof("CreateOrUpdateConfigMap : %v.", err)
		return err
	}

	return nil
}

// WriteJobFaultRankIDIntoCM Record into job cm.
func WriteJobFaultRankIDIntoCM(ssn *framework.Session, job *api.JobInfo, cmData map[string]string) error {
	var faultRankIdsCM = &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      JobFaultRankIDCMPre + job.Name,
			Namespace: job.Namespace,
		},
		Data: cmData,
	}
	klog.V(util.LogErrorLev).Infof("WriteJobFaultRankIDIntoCacheAndCM cm is : %v/%v.",
		faultRankIdsCM.Namespace, faultRankIdsCM.Name)
	if err := util.CreateOrUpdateConfigMap(ssn.KubeClient(), faultRankIdsCM, cmName, cmNameSpace); err != nil {
		klog.V(util.LogErrorLev).Infof("WriteJobFaultRankIDIntoCacheAndCM : %v.", err)
		return err
	}
	return nil
}

// WriteJobFaultRankIDIntoCacheAndCM Write job fault RankID into configmap, every job has own cm.
func WriteJobFaultRankIDIntoCacheAndCM(ssn *framework.Session, job *api.JobInfo) error {
	if !isGraceDeleteJob(job) {
		return fmt.Errorf("%v not GraceDeleteJob", job.UID)
	}

	faultRankIdsMap, getErr := getJobFaultNPURankIDCMData(job)
	if getErr != nil {
		klog.V(util.LogErrorLev).Infof("getJobFaultNPURankIdCMData : %v.", getErr)
		return getErr
	}
	// write into cache
	if writeErr := WriteJobFaultRankIDIntoCache(job); writeErr != nil {
		klog.V(util.LogErrorLev).Infof("WriteJobFaultRankIDIntoCache : %v.", writeErr)
		return writeErr
	}
	// write into job cm
	if writeCMErr := WriteJobFaultRankIDIntoCM(ssn, job, faultRankIdsMap); writeCMErr != nil {
		klog.V(util.LogErrorLev).Infof("WriteJobFaultRankIDIntoCache : %v.", writeCMErr)
		return writeCMErr
	}

	return nil
}

// ReadFaultNPUJobsFromCM read from ConfigMap FaultNPUJobs and update to the cache.
func ReadFaultNPUJobsFromCM(ssn *framework.Session) error {
	var cmErr error
	// Get configMap failure does not affect the update cache.
	if cmErr = UpdateFaultNPUInfFromCM(ssn); cmErr != nil {
		klog.V(util.LogErrorLev).Infof("updateFaultNPUInfFromCM :%v.", cmErr)
	}

	if err := updateReSchedulerDataFromSession(ssn); err != nil {
		klog.V(util.LogErrorLev).Infof("updateReSchedulerDataFromSession :%v.", err)
		return fmt.Errorf("%v %v", cmErr, err)
	}

	return cmErr
}

func updateFaultNodeFromCM(tmpData string) error {
	faultNode, covErr := convertToReSchedulerNodesMapFromCM(tmpData)
	if covErr != nil {
		klog.V(util.LogErrorLev).Infof("convertToReSchedulerNodesMapFromCM: %v.", covErr)
		return covErr
	}

	ReSchedulerCache[CmNodeKind] = faultNode

	return nil
}

func updateFaultCardFromCM(tmpData string) error {
	faultCars, covErr := convertToReSchedulerCardsMapFromCM(tmpData)
	if covErr != nil {
		klog.V(util.LogErrorLev).Infof("updateFaultCardFromCM: %v.", covErr)
		return covErr
	}

	ReSchedulerCache[CmCardKind] = faultCars

	return nil
}

func updateNodeHeartbeatFromCM(tmpData string) error {
	heartbeat, covErr := convertToNodeHeartbeatMapFromCM(tmpData)
	if covErr != nil {
		klog.V(util.LogErrorLev).Infof("convertToNodeHeartbeatMapFromCM: %v.", covErr)
		return covErr
	}

	ReSchedulerCache[CmNodeHeartbeatKind] = heartbeat

	return nil
}

func updateJobRankIdsFromCM(tmpData string) error {
	rankIds, covErr := convertToJobRankIdsMapFromCM(tmpData)
	if covErr != nil {
		klog.V(util.LogErrorLev).Infof("convertToJobRankIdsMapFromCM: %v.", covErr)
		return covErr
	}
	var rankIDData = make(map[api.JobID]FaultRankIDRecordJobCMData, 1)
	for dataID, data := range rankIds {
		jobIDStr := strings.Replace(dataID, "_", "/", -1)
		if strings.Count(jobIDStr, "/") != 1 {
			countErr := fmt.Errorf("%s more than one character '_'", dataID)
			return countErr
		}

		rankIDData[api.JobID(jobIDStr)] = data
	}

	ReSchedulerCache[CmJobRankIds] = rankIDData

	return nil
}

func updateReSchedulerData(cmData *v1.ConfigMap) error {
	if len(cmData.Data) == 0 {
		klog.V(util.LogErrorLev).Infof("updateFaultNodeFromCM cmData is nil, nothing to do.")
		return nil
	}

	for dataID, buffer := range cmData.Data {
		if len(buffer) == 0 {
			klog.V(util.LogErrorLev).Infof("updateFaultNodeFromCM %v is nil.", dataID)
			continue
		}

		switch dataID {
		case CmJobKind:
			if err := updateReSchedulerJobsFromCM(buffer); err != nil {
				klog.V(util.LogErrorLev).Infof("updateReSchedulerJobsFromCM %v.", err)
			}
		case CmNodeKind:
			if err := updateFaultNodeFromCM(buffer); err != nil {
				klog.V(util.LogErrorLev).Infof("updateFaultNodeFromCM %v.", err)
			}
		case CmCardKind:
			if err := updateFaultCardFromCM(buffer); err != nil {
				klog.V(util.LogErrorLev).Infof("updateFaultCardFromCM: %v.", err)
			}
		case CmNodeHeartbeatKind:
			if err := updateNodeHeartbeatFromCM(buffer); err != nil {
				klog.V(util.LogErrorLev).Infof("updateNodeHeartbeatFromCM: %v.", err)
			}
		case CmJobRankIds:
			if err := updateJobRankIdsFromCM(buffer); err != nil {
				klog.V(util.LogErrorLev).Infof("updateJobRankIdsFromCM: %v.", err)
			}
		default:
			klog.V(util.LogErrorLev).Infof("updateReSchedulerData no support type:%v", dataID)
		}
	}
	return nil
}

// UpdateFaultNPUInfFromCM update fault data in cm.
func UpdateFaultNPUInfFromCM(ssn *framework.Session) error {
	cmData, err := util.GetConfigMapWithRetry(ssn.KubeClient(), cmNameSpace, cmName)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("getConfigMapWithRetry :%v.", err)
		return err
	}

	updateErr := updateReSchedulerData(cmData)
	if updateErr != nil {
		klog.V(util.LogErrorLev).Infof("updateReSchedulerData :%v.", updateErr)
		return updateErr
	}

	klog.V(util.LogErrorLev).Infof("updateFaultNPUInfFromCM success.")
	return nil
}

func convertToReSchedulerJobsMapFromCM(buffer string) (map[string]ReSchedulerTasks, error) {
	reSchedulerJob := make(map[string]ReSchedulerTasks, util.NPUIndex3)
	if unmarshalErr := json.Unmarshal([]byte(buffer), &reSchedulerJob); unmarshalErr != nil {
		klog.V(util.LogInfoLev).Infof("convertToReSchedulerJobsMapFromCM Unmarshal: %v.", unmarshalErr)
		return nil, unmarshalErr
	}
	return reSchedulerJob, nil
}

func updateReSchedulerJobsFromCM(buffer string) error {
	reSchedulerJob, covErr := convertToReSchedulerJobsMapFromCM(buffer)
	if covErr != nil {
		klog.V(util.LogErrorLev).Infof("convertToReSchedulerNodesMap: %v.", covErr)
		return covErr
	}
	var jobData = make(map[api.JobID]ReSchedulerTasks, 1)
	for dataID, data := range reSchedulerJob {
		jobIDStr := strings.Replace(dataID, "_", "/", -1)
		if strings.Count(jobIDStr, "/") != 1 {
			countErr := fmt.Errorf("%s more than one character '_'", dataID)
			return countErr
		}

		jobData[api.JobID(jobIDStr)] = data
	}

	ReSchedulerCache[CmJobKind] = jobData
	return nil
}
