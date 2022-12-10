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
	"time"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

func (reCache *DealReSchedulerCache) setFaultNodes(faultNodes []FaultNode) {
	reCache.FaultNodes = faultNodes
}

func (reCache *DealReSchedulerCache) setFaultJobs(faultJobs []FaultJob) {
	reCache.FaultJobs = faultJobs
}

func (reCache *DealReSchedulerCache) setNodeHeartbeat(nodeHeartbeat []NodeHeartbeat) {
	reCache.NodeHeartbeats = nodeHeartbeat
}

func (reCache *DealReSchedulerCache) setNodeRankOccurrenceMap(
	nodeRankOccurrenceMap map[api.JobID][]AllocNodeRankOccurrence) {
	reCache.AllocNodeRankOccurrenceMap = nodeRankOccurrenceMap
}

func (reCache DealReSchedulerCache) getFaultNodesFromCM(buffer string) ([]FaultNode, error) {
	var faultNodes []FaultNode
	if unmarshalErr := json.Unmarshal([]byte(buffer), &faultNodes); unmarshalErr != nil {
		klog.V(util.LogErrorLev).Infof("Unmarshal FaultNodes from cache failed")
		return nil, fmt.Errorf("faultNodes convert from CM error: %#v", unmarshalErr)
	}
	return faultNodes, nil
}

func (reCache DealReSchedulerCache) getFaultJobsFromCM(buffer string) ([]FaultJob, error) {
	var faultJobs []FaultJob
	if unmarshalErr := json.Unmarshal([]byte(buffer), &faultJobs); unmarshalErr != nil {
		klog.V(util.LogErrorLev).Infof("Unmarshal FaultJobs from cache failed")
		return nil, fmt.Errorf("faultJobs convert from CM failed")
	}
	return faultJobs, nil
}

func (reCache DealReSchedulerCache) getNodeHeartbeatFromCM(buffer string) ([]NodeHeartbeat, error) {
	var nodeHBs []NodeHeartbeat
	if unmarshalErr := json.Unmarshal([]byte(buffer), &nodeHBs); unmarshalErr != nil {
		klog.V(util.LogErrorLev).Infof("Unmarshal NodeHeartbeat from cache failed")
		return nil, fmt.Errorf("faultNodes convert from CM error: %#v", unmarshalErr)
	}
	return nodeHBs, nil
}

func (reCache DealReSchedulerCache) getNodeRankOccurrenceMapFromCM(
	buffer string) (map[api.JobID][]AllocNodeRankOccurrence, error) {
	var nodeRankOccMap map[api.JobID][]AllocNodeRankOccurrence
	if unmarshalErr := json.Unmarshal([]byte(buffer), &nodeRankOccMap); unmarshalErr != nil {
		klog.V(util.LogErrorLev).Infof("Unmarshal AllocNodeRankOccurrence from cache failed")
		return nil, fmt.Errorf("faultNodes convert from CM error: %#v", unmarshalErr)
	}
	return nodeRankOccMap, nil
}

// SetFaultNodesFromCM unmarshal FaultNodes from string into struct and set the value
func (reCache *DealReSchedulerCache) SetFaultNodesFromCM() error {
	if reCache == nil {
		klog.V(util.LogErrorLev).Infof("SetFaultNodesFromCM failed: %s, reCache is none", util.ArgumentError)
		return errors.New(util.ArgumentError)
	}
	faultNodeData, ok := reCache.CMData[CmFaultNodeKind]
	if !ok {
		return fmt.Errorf("reading %s data from reScheduler configmap failed", CmFaultNodeKind)
	}
	faultNodes, err := reCache.getFaultNodesFromCM(faultNodeData)
	if err != nil {
		return fmt.Errorf("getFaultNodesFromCM %#v", err)
	}
	reCache.setFaultNodes(faultNodes)
	return nil
}

// SetFaultJobsFromCM unmarshal FaultJobs from string into struct and set the value
func (reCache *DealReSchedulerCache) SetFaultJobsFromCM(jobType string) error {
	if reCache == nil {
		klog.V(util.LogErrorLev).Infof("SetFaultNodesFromCM failed: %s, reCache is none", util.ArgumentError)
		return errors.New(util.ArgumentError)
	}
	if len(jobType) == 0 {
		klog.V(util.LogErrorLev).Infof("SetFaultNodesFromCM failed: %s: jobType is none", util.ArgumentError)
		return errors.New(util.ArgumentError)
	}
	faultJobData, ok := reCache.CMData[jobType]
	if !ok {
		return fmt.Errorf("reading %s data from reScheduler configmap failed", jobType)
	}
	faultJobs, err := reCache.getFaultJobsFromCM(faultJobData)
	if err != nil {
		return fmt.Errorf("getFaultNodesFromCM %#v", err)
	}
	reCache.setFaultJobs(faultJobs)
	return nil
}

// SetNodeHeartbeatFromCM unmarshal NodeHeartbeat from string into struct and set the value
func (reCache *DealReSchedulerCache) SetNodeHeartbeatFromCM() error {
	if reCache == nil {
		klog.V(util.LogErrorLev).Infof("SetFaultNodesFromCM failed: %s, reCache is none", util.ArgumentError)
		return errors.New(util.ArgumentError)
	}
	nodeHBsData, ok := reCache.CMData[CmNodeHeartbeatKind]
	if !ok {
		return fmt.Errorf("reading %s data from reScheduler configmap failed", CmNodeHeartbeatKind)
	}
	nodeHBs, err := reCache.getNodeHeartbeatFromCM(nodeHBsData)
	if err != nil {
		return fmt.Errorf("getFaultNodesFromCM %#v", err)
	}
	reCache.setNodeHeartbeat(nodeHBs)
	return nil
}

// SetNodeRankOccurrenceMapFromCM unmarshal NodeRankOccurrenceMap from string into struct and set the value
func (reCache *DealReSchedulerCache) SetNodeRankOccurrenceMapFromCM() error {
	if reCache == nil {
		klog.V(util.LogErrorLev).Infof("SetFaultNodesFromCM failed: %s, reCache is none", util.ArgumentError)
		return errors.New(util.ArgumentError)
	}
	nodeRankOccMapData, ok := reCache.CMData[CmNodeRankTimeMapKind]
	if !ok {
		return fmt.Errorf("reading %s data from reScheduler configmap failed", CmNodeRankTimeMapKind)
	}
	nodeRankOccMap, err := reCache.getNodeRankOccurrenceMapFromCM(nodeRankOccMapData)
	if err != nil {
		return fmt.Errorf("getFaultNodesFromCM %#v", err)
	}
	reCache.setNodeRankOccurrenceMap(nodeRankOccMap)
	return nil
}

func (reCache *DealReSchedulerCache) marshalCacheDataToString(data interface{}) (string, error) {
	dataBuffer, err := json.Marshal(data)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("marshalCacheDataToString err: %#v", err)
		return "", err
	}
	return string(dataBuffer), nil
}

// getRealFaultJobs only return FaultJobs whose IsFaultJob is true
func (reCache DealReSchedulerCache) getRealFaultJobs() ([]FaultJob, error) {
	var realFaultJobs []FaultJob
	for _, fJob := range reCache.FaultJobs {
		if !fJob.IsFaultJob || fJob.ReScheduleKey == JobOffRescheduleLabelValue {
			continue // only save real-fault and reschedule-enabled jobs
		}
		// if and only if the task is distributional would a network unhealthy card trigger re-scheduling
		if util.IsSliceContain(NodeCardNetworkUnhealthy, fJob.FaultTypes) &&
			!util.IsSliceContain(NodeCardUnhealthy, fJob.FaultTypes) &&
			!util.IsSliceContain(NodeUnhealthy, fJob.FaultTypes) &&
			len(fJob.FaultTasks) < util.NPUIndex2 {
			continue
		}
		realFaultJobs = append(realFaultJobs, fJob)
	}
	if len(realFaultJobs) == 0 {
		klog.V(util.LogDebugLev).Infof("getRealFaultJobs %s.", NoFaultJobsErr)
		return nil, fmt.Errorf(NoFaultJobsErr)
	}
	return realFaultJobs, nil
}

// getRealFaultNodes get the nodes whose isFaultNode property takes true value
func (reCache DealReSchedulerCache) getRealFaultNodes() []FaultNode {
	var realFaultNodes []FaultNode
	for _, fNode := range reCache.FaultNodes {
		if !fNode.IsFaultNode {
			continue
		}
		realFaultNodes = append(realFaultNodes, fNode)
	}
	return realFaultNodes
}

func (reCache *DealReSchedulerCache) writeFaultNodesToCMString() (string, error) {
	realFaultNode := reCache.getRealFaultNodes()
	nodeData, err := reCache.marshalCacheDataToString(realFaultNode)
	if err != nil {
		return "", fmt.Errorf("writeFaultNodesToCM: %#v", err)
	}
	return nodeData, nil
}

func (reCache *DealReSchedulerCache) writeFaultJobsToCMString() (string, error) {
	realFaultJob, err := reCache.getRealFaultJobs()
	if err != nil {
		if err.Error() == NoFaultJobsErr {
			return "", nil
		}
		return "", fmt.Errorf("writeFaultJobsToCM: %#v", err)
	}
	jobData, err := reCache.marshalCacheDataToString(realFaultJob)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("WriteFaultJobsToCM: %#v.", err)
		return "", fmt.Errorf("writeFaultJobsToCM: %#v", err)
	}
	return jobData, nil
}

func (reCache *DealReSchedulerCache) writeNodeHeartbeatToCMString() (string, error) {
	var nodeHB NodeHeartbeat
	var nodeHBs []NodeHeartbeat
	for _, fNode := range reCache.FaultNodes {
		nodeHB = NodeHeartbeat{
			NodeName:      fNode.NodeName,
			HeartbeatTime: fNode.NewHeartbeatTime,
			UpdateTime:    fNode.UpdateHeartbeatTime,
		}
		nodeHBs = append(nodeHBs, nodeHB)
	}
	nodeHBsData, err := reCache.marshalCacheDataToString(nodeHBs)
	if err != nil {
		return "", fmt.Errorf("writeNodeHeartbeatToCM: %#v", err)
	}
	return nodeHBsData, nil
}

func (reCache *DealReSchedulerCache) writeNodeRankOccurrenceMapToCMString() (string, error) {
	nodeRankOccMapData, err := reCache.marshalCacheDataToString(reCache.AllocNodeRankOccurrenceMap)
	if err != nil {
		return "", fmt.Errorf("writeNodeRankOccurrenceMapToCM: %#v", err)
	}
	return nodeRankOccMapData, nil
}

func (reCache *DealReSchedulerCache) writeJobRankIndexToCMString(fJob *FaultJob) (*FaultRankIdsJobCMData, string,
	error) {
	faultRankIds := &FaultRankIdsJobCMData{
		FaultRankIds: fJob.JobRankIds,
		CreatTime:    time.Now().Unix(),
	}
	faultRankIdsData, err := reCache.marshalCacheDataToString(faultRankIds)
	if err != nil {
		return nil, "", fmt.Errorf("%s writeJobRankIndexToCMString: %#v", fJob.JobName, err)
	}
	return faultRankIds, faultRankIdsData, nil
}

// WriteReSchedulerCacheToEnvCache write the modifications on cache data to env to update re-scheduling configmap
func (reCache *DealReSchedulerCache) WriteReSchedulerCacheToEnvCache(env *plugin.ScheduleEnv, jobType string) error {
	if reCache == nil || env == nil {
		return errors.New(util.ArgumentError)
	}
	env.Cache.Names[RePropertyName] = CmName
	env.Cache.Namespaces[RePropertyName] = CmNameSpace
	fNodeString, err := reCache.writeFaultNodesToCMString()
	if err != nil {
		klog.V(util.LogErrorLev).Infof("WriteReSchedulerCacheToEnvCache: %#v", err)
	}
	fJobString, err := reCache.writeFaultJobsToCMString()
	if err != nil {
		klog.V(util.LogInfoLev).Infof("WriteReSchedulerCacheToEnvCache: %#v", err)
	}
	nodeHBString, err := reCache.writeNodeHeartbeatToCMString()
	if err != nil {
		klog.V(util.LogErrorLev).Infof("WriteReSchedulerCacheToEnvCache: %#v", err)
	}
	nodeRankOccurrenceMapString, err := reCache.writeNodeRankOccurrenceMapToCMString()
	if err != nil {
		klog.V(util.LogErrorLev).Infof("WriteReSchedulerCacheToEnvCache: %#v", err)
	}
	cmData, ok := env.Cache.Data[RePropertyName]
	if !ok {
		cmData := make(map[string]string, util.MapInitNum)
		env.Cache.Data[RePropertyName] = cmData
	}
	cmData[CmFaultNodeKind] = fNodeString
	cmData[jobType] = fJobString
	cmData[CmNodeHeartbeatKind] = nodeHBString
	cmData[CmNodeRankTimeMapKind] = nodeRankOccurrenceMapString
	_, ok = cmData[CmCheckCode]
	if ok {
		delete(cmData, CmCheckCode) // if check code exists, delete and create new
	}
	checkCode := plugin.MakeDataHash(cmData)
	klog.V(util.LogDebugLev).Infof("cm checkCode: %s, calc checkCode: %s, check equal: %v", checkCode,
		plugin.MakeDataHash(cmData), checkCode == plugin.MakeDataHash(cmData))
	cmData[CmCheckCode] = checkCode
	if jobType != CmFaultJob910x8Kind && jobType != CmFaultJob910x4Kind {
		return nil
	}
	if err := reCache.writeRecoveryCacheToEnv(env); err != nil {
		klog.V(util.LogErrorLev).Infof("WriteReSchedulerCacheToEnvCache: %#v", err)
	}
	return nil
}

func (reCache *DealReSchedulerCache) writeRecoveryCacheToEnv(env *plugin.ScheduleEnv) error {
	for _, fJob := range reCache.FaultJobs { // configmap for recovery
		if fJob.IsFaultJob {
			env.Cache.Names[JobRecovery] = JobFaultRankIDCMPre + fJob.JobName
			env.Cache.Namespaces[JobRecovery] = fJob.JobNamespace
			jobRankIndex, jobRankIndexString, err := reCache.writeJobRankIndexToCMString(&fJob)
			if err != nil {
				return err
			}
			cmRecData, ok := env.Cache.Data[JobRecovery]
			if !ok {
				cmRecData := make(map[string]string, util.MapInitNum)
				env.Cache.Data[JobRecovery] = cmRecData
			}
			cmRecData[JobFaultRankIDCMDataKey] = jobRankIndexString
			klog.V(util.LogDebugLev).Infof("fault configMap string to calculate checkCode: <%s>", jobRankIndexString)
			checkCode := plugin.MakeDataHash(jobRankIndex)
			klog.V(util.LogDebugLev).Infof("checkCode for fault configMap: %s", checkCode)
			cmRecData[CmCheckCode] = checkCode
		}
	}
	return nil
}
