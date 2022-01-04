/*
Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package rescheduling is using for HuaWei Ascend pin fault rescheduling.

*/
package rescheduling

import (
	"context"
	"encoding/json"
	"fmt"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"reflect"
	"strconv"
	"strings"
	"time"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/util"
)

func getConfigMapWithRetry(client kubernetes.Interface, namespace, cmName string) (*v1.ConfigMap, error) {
	var cm *v1.ConfigMap
	var err error

	// There can be no delay or blocking operations in a session.
	if cm, err = client.CoreV1().ConfigMaps(namespace).Get(context.TODO(), cmName, metav1.GetOptions{}); err != nil {
		return nil, err
	}

	return cm, nil
}

func isConfigMapChanged(k8s kubernetes.Interface, cm *v1.ConfigMap) bool {
	cmData, getErr := getConfigMapWithRetry(k8s, cmNameSpace, cmName)
	if getErr != nil {
		return true
	}
	if reflect.DeepEqual(cmData, cm) {
		return false
	}

	return true
}

func createOrUpdateConfigMap(k8s kubernetes.Interface, cm *v1.ConfigMap) error {
	_, cErr := k8s.CoreV1().ConfigMaps(cm.ObjectMeta.Namespace).Create(context.TODO(), cm, metav1.CreateOptions{})
	if cErr != nil {
		if !apierrors.IsAlreadyExists(cErr) {
			return fmt.Errorf("unable to create ConfigMap:%v", cErr)
		}

		// To reduce the cm write operations
		if !isConfigMapChanged(k8s, cm) {
			klog.V(logInfoLev).Infof("configMap not changed,no need update")
			return nil
		}

		_, err := k8s.CoreV1().ConfigMaps(cm.ObjectMeta.Namespace).Update(context.TODO(), cm, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("unable to update ConfigMap:%v", err)
		}
	}
	return nil
}

func deleteSchedulerConfigMap(ssn *framework.Session, nameSpace, cmName string) error {
	err := ssn.KubeClient().CoreV1().ConfigMaps(nameSpace).Delete(context.TODO(), cmName, metav1.DeleteOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			klog.V(logErrorLev).Infof("Failed to delete Configmap %v in%v: %v",
				nameSpace, cmName, err)
			return err
		}
	}
	return nil
}

func convertToReSchedulerJobsMapFromCM(buffer string) (map[string]ReSchedulerTasks, error) {
	reSchedulerJob := make(map[string]ReSchedulerTasks, constIntNum3)
	if unmarshalErr := json.Unmarshal([]byte(buffer), &reSchedulerJob); unmarshalErr != nil {
		klog.V(logErrorLev).Infof("convertToReSchedulerJobsMapFromCM Unmarshal: %v %v.", buffer, unmarshalErr)
		return nil, unmarshalErr
	}
	return reSchedulerJob, nil
}

func updateReSchedulerJobsFromCM(buffer string) error {
	reSchedulerJob, covErr := convertToReSchedulerJobsMapFromCM(buffer)
	if covErr != nil {
		klog.V(logErrorLev).Infof("convertToReSchedulerNodesMap: %v.", covErr)
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

func updateFaultNodeFromCM(tmpData string) error {
	faultNode, covErr := convertToReSchedulerNodesMapFromCM(tmpData)
	if covErr != nil {
		klog.V(logErrorLev).Infof("convertToReSchedulerNodesMapFromCM: %v.", covErr)
		return covErr
	}

	ReSchedulerCache[CmNodeKind] = faultNode

	return nil
}

func updateFaultCardFromCM(tmpData string) error {
	faultCars, covErr := convertToReSchedulerCardsMapFromCM(tmpData)
	if covErr != nil {
		klog.V(logErrorLev).Infof("updateFaultCardFromCM: %v.", covErr)
		return covErr
	}

	ReSchedulerCache[CmCardKind] = faultCars

	return nil
}

func updateNodeHeartbeatFromCM(tmpData string) error {
	heartbeat, covErr := convertToNodeHeartbeatMapFromCM(tmpData)
	if covErr != nil {
		klog.V(logErrorLev).Infof("convertToNodeHeartbeatMapFromCM: %v.", covErr)
		return covErr
	}

	ReSchedulerCache[CmNodeHeartbeatKind] = heartbeat

	return nil
}

func updateJobRankIdsFromCM(tmpData string) error {
	rankIds, covErr := convertToJobRankIdsMapFromCM(tmpData)
	if covErr != nil {
		klog.V(logErrorLev).Infof("convertToJobRankIdsMapFromCM: %v.", covErr)
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
		klog.V(logDebugLev).Infof("updateFaultNodeFromCM cmData is nil, nothing to do.")
		return nil
	}

	for dataID, buffer := range cmData.Data {
		if len(buffer) == 0 {
			klog.V(logDebugLev).Infof("updateFaultNodeFromCM %v is nil.", dataID)
			continue
		}

		switch dataID {
		case CmJobKind:
			if err := updateReSchedulerJobsFromCM(buffer); err != nil {
				klog.V(logErrorLev).Infof("updateReSchedulerJobsFromCM %v.", err)
			}
		case CmNodeKind:
			if err := updateFaultNodeFromCM(buffer); err != nil {
				klog.V(logErrorLev).Infof("updateFaultNodeFromCM %v.", err)
			}
		case CmCardKind:
			if err := updateFaultCardFromCM(buffer); err != nil {
				klog.V(logErrorLev).Infof("updateFaultCardFromCM: %v.", err)
			}
		case CmNodeHeartbeatKind:
			if err := updateNodeHeartbeatFromCM(buffer); err != nil {
				klog.V(logErrorLev).Infof("updateNodeHeartbeatFromCM: %v.", err)
			}
		case CmJobRankIds:
			if err := updateJobRankIdsFromCM(buffer); err != nil {
				klog.V(logErrorLev).Infof("updateJobRankIdsFromCM: %v.", err)
			}
		default:
			klog.V(logErrorLev).Infof("updateReSchedulerData no support type:%v", dataID)
		}
	}
	return nil
}

func updateFaultNPUInfFromCM(ssn *framework.Session) error {
	cmData, err := getConfigMapWithRetry(ssn.KubeClient(), cmNameSpace, cmName)
	if err != nil {
		klog.V(logErrorLev).Infof("getConfigMapWithRetry :%v.", err)
		return err
	}

	updateErr := updateReSchedulerData(cmData)
	if updateErr != nil {
		klog.V(logErrorLev).Infof("updateReSchedulerData :%v.", updateErr)
		return updateErr
	}

	klog.V(logDebugLev).Infof("updateFaultNPUInfFromCM success: %+v.", ReSchedulerCache)
	return nil
}

// ReadFaultNPUJobsFromCM read from ConfigMap FaultNPUJobs and update to the cache.
func ReadFaultNPUJobsFromCM(ssn *framework.Session) error {
	var cmErr error
	// Get configMap failure does not affect the update cache.
	if cmErr = updateFaultNPUInfFromCM(ssn); cmErr != nil {
		klog.V(logErrorLev).Infof("updateFaultNPUInfFromCM :%v.", cmErr)
	}

	if err := updateReSchedulerDataFromSession(ssn); err != nil {
		klog.V(logErrorLev).Infof("updateReSchedulerDataFromSession :%v.", err)
		return fmt.Errorf("%v %v", cmErr, err)
	}

	return cmErr
}

func getCMWriteDate(reSchedulerData map[string]interface{}) map[string]string {
	var data = make(map[string]string, 1)
	for dataKind, faultData := range reSchedulerData {
		switch dataKind {
		case CmJobKind:
			jobData, err := getCMJobWriteData(faultData)
			if err != nil {
				klog.V(logErrorLev).Infof("getCMJobWriteData :%v.", err)
				continue
			}
			data[CmJobKind] = jobData
		case CmNodeKind:
			nodeData, err := getCMNodeWriteData(faultData)
			if err != nil {
				klog.V(logErrorLev).Infof("getCMJobWriteData :%v.", err)
				continue
			}
			data[CmNodeKind] = nodeData
		case CmCardKind:
			cardData, err := getCMCardWriteData(faultData)
			if err != nil {
				klog.V(logDebugLev).Infof("getCMCardWriteData :%v.", err)
				continue
			}
			data[CmCardKind] = cardData
		case CmNodeHeartbeatKind:
			heartbeatData, err := getCMHeartbeatWriteData(faultData)
			if err != nil {
				klog.V(logDebugLev).Infof("getCMHeartbeatWriteData :%v.", err)
				continue
			}
			data[CmNodeHeartbeatKind] = heartbeatData
		case CmJobRankIds:
			rankIdsData, err := getCMRankIdsWriteData(faultData)
			if err != nil {
				klog.V(logDebugLev).Infof("getCMRankIdsWriteData :%v.", err)
				continue
			}
			data[CmJobRankIds] = rankIdsData
		default:
			klog.V(logErrorLev).Infof("getCMWriteDate not support %v %v.", dataKind, faultData)
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

	klog.V(logDebugLev).Infof("Write faultNPUJobs into cm: %+v.", faultNPUConfigMap)
	if err := createOrUpdateConfigMap(ssn.KubeClient(), faultNPUConfigMap); err != nil {
		klog.V(logErrorLev).Infof("createOrUpdateConfigMap : %v.", err)
		return err
	}

	return nil
}

func getJobFaultRankIds(job *api.JobInfo) (string, error) {
	allFaultNPUs, cacheErr := getFaultCardsFromCache()
	if cacheErr != nil {
		klog.V(logErrorLev).Infof("getJobFaultRankIds %s %v.", job.Name, cacheErr)
		return "", cacheErr
	}

	klog.V(logDebugLev).Infof("getJobFaultRankIds %s %v.", job.Name, allFaultNPUs)
	var rankIds []string
	rankIndexMap, err := getFaultJobPODRankIndexMapFromCache(job)
	if err != nil {
		klog.V(logErrorLev).Infof("getJobFaultRankIds %s %v.", job.Name, err)
		return "", err
	}
	for nodeName, rankIndexStr := range rankIndexMap {
		klog.V(logInfoLev).Infof("getJobFaultRankIds %s %v.", nodeName, rankIndexStr)
		var faultNPUs []string
		faultNPUsOnNode, ok := allFaultNPUs[nodeName]
		if !ok {
			continue
		}
		faultNPUs = append(faultNPUs, faultNPUsOnNode.FaultNPUs...)
		faultNPUs = append(faultNPUs, faultNPUsOnNode.NetworkUnhealthyNPUs...)
		tmpDeviceID := util.ChangeTopToIntArray(strings.Join(faultNPUs, ","), npu800And9000CardPreName)
		if len(tmpDeviceID) == 0 {
			klog.V(logInfoLev).Infof("%s getTopFromNode %+v nil.", nodeName, faultNPUsOnNode)
			continue
		}
		rankIndex, covertErr := strconv.Atoi(rankIndexStr)
		if covertErr != nil {
			klog.V(logErrorLev).Infof("%s getJobFaultRankIds covert %v.", covertErr, rankIndexStr)
			continue
		}
		for _, tmp := range tmpDeviceID {
			rankIds = append(rankIds, strconv.Itoa(tmp+rankIndex*node910X8NPUNum))
		}
	}
	dataBuffer, err := json.Marshal(rankIds)
	if err != nil {
		klog.V(logErrorLev).Infof("marshalCacheDataToString err: %v.", err)
		return "", err
	}
	return string(dataBuffer), nil
}

func getFaultNPUJobCMData(job *api.JobInfo) (*FaultRankIdsJobCMData, error) {
	fRankIds, getErr := getJobFaultRankIds(job)
	if getErr != nil {
		klog.V(logErrorLev).Infof("getJobFaultRankIds err: %v.", getErr)
		return nil, getErr
	}

	faultRankIds := &FaultRankIdsJobCMData{
		FaultRankIds: fRankIds,
		CreatTime:    time.Now().Unix(),
	}
	return faultRankIds, nil
}

func getJobFaultNPURankIDCMData(job *api.JobInfo) (map[string]string, error) {
	faultRankIdsMap := make(map[string]string, constIntNum3)

	faultRankIds, getErr := getFaultNPUJobCMData(job)
	if getErr != nil {
		return nil, getErr
	}
	dataBuffer, err := json.Marshal(faultRankIds)
	if err != nil {
		klog.V(logErrorLev).Infof("marshalCacheDataToString err: %v.", err)
		return nil, err
	}
	faultRankIdsMap[JobFaultRankIDCMDataKey] = string(dataBuffer)

	return faultRankIdsMap, nil
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

// WriteJobFaultRankIDIntoCM Record into job cm.
func WriteJobFaultRankIDIntoCM(ssn *framework.Session, job *api.JobInfo, cmData map[string]string) error {
	var faultRankIdsCM = &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      JobFaultRankIDCMPre + job.Name,
			Namespace: job.Namespace,
		},
		Data: cmData,
	}
	klog.V(logDebugLev).Infof("WriteJobFaultRankIDIntoCacheAndCM cm is : %+v.", faultRankIdsCM)
	if err := createOrUpdateConfigMap(ssn.KubeClient(), faultRankIdsCM); err != nil {
		klog.V(logErrorLev).Infof("WriteJobFaultRankIDIntoCacheAndCM : %v.", err)
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
		klog.V(logErrorLev).Infof("getJobFaultNPURankIdCMData : %v.", getErr)
		return getErr
	}
	// write into cache
	if writeErr := WriteJobFaultRankIDIntoCache(job); writeErr != nil {
		klog.V(logErrorLev).Infof("WriteJobFaultRankIDIntoCache : %v.", writeErr)
		return writeErr
	}
	// write into job cm
	if writeCMErr := WriteJobFaultRankIDIntoCM(ssn, job, faultRankIdsMap); writeCMErr != nil {
		klog.V(logErrorLev).Infof("WriteJobFaultRankIDIntoCache : %v.", writeCMErr)
		return writeCMErr
	}

	return nil
}
