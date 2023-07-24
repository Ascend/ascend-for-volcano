/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package rescheduling is using for HuaWei Ascend pin affinity schedule utilities.
*/
package rescheduling

import (
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

const (
	// RePropertyName name specifying re-scheduler cm
	RePropertyName = "re-scheduling"
	// CmName Name of ReSchedulerConfigmap
	CmName = "vcjob-fault-npu-cm"
	// CmNameSpace Namespace of ReSchedulerConfigmap
	CmNameSpace = "volcano-system"

	// JobRescheduleLabelKey key word of re-scheduling configuration
	JobRescheduleLabelKey = "fault-scheduling"
	// JobGraceRescheduleLabelValue Grace delete reschedule job, possible value of re-scheduling configuration
	JobGraceRescheduleLabelValue = "grace"
	// JobForceRescheduleLabelValue Force delete reschedule job, possible value of re-scheduling configuration
	JobForceRescheduleLabelValue = "force"
	// JobOffRescheduleLabelValue not delete reschedule job, possible value of re-scheduling configuration
	JobOffRescheduleLabelValue = "off"
	// GraceOverTimeKey for GraceOverTime config by user
	GraceOverTimeKey = "grace-over-time"
	// ElasticSchedulingKey for distinguishing whether a job is enabled with elastic scheduling
	ElasticSchedulingKey = "elastic-scheduling"
	// JobOnElasticScheduling job enabled with elastic scheduling
	JobOnElasticScheduling = "on"
	// JobOffElasticScheduling job not enabled with elastic scheduling
	JobOffElasticScheduling = "off"

	nodeHeartbeat         = "noded/heartbeat"
	nodeHeartbeatInterval = "noded/heartbeat-interval"

	nodeDEnableKey      = "nodeDEnable"
	nodeDEnableOnValue  = "on"
	nodeDEnableOffValue = "off"

	podRankIndex = "hccl/rankIndex"

	// CmFaultNodeKind key in configmap which saves the FaultNode cache
	CmFaultNodeKind = "fault-node"
	// CmFaultJob910bx2Kind key in configmap which saves the 910bx2 FaultJob cache
	CmFaultJob910bx2Kind = "fault-job-910bx2"
	// CmFaultJob910bx8Kind key in configmap which saves the 910bx8 FaultJob cache
	CmFaultJob910bx8Kind = "fault-job-910bx8"
	// CmFaultJob910bx16Kind key in configmap which saves the 910bx16 FaultJob cache
	CmFaultJob910bx16Kind = "fault-job-910bx16"
	// CmFaultJob910x8Kind key in configmap which saves the 910x8 FaultJob cache
	CmFaultJob910x8Kind = "fault-job-910x8"
	// CmFaultJob910x4Kind key in configmap which saves the 910x8 FaultJob cache
	CmFaultJob910x4Kind = "fault-job-910x4"
	// CmFaultJob910x2Kind key in configmap which saves the 910x8 FaultJob cache
	CmFaultJob910x2Kind = "fault-job-910x2"
	// CmFaultJob310x4Kind key in configmap which saves the 310x4 FaultJob cache
	CmFaultJob310x4Kind = "fault-job-310x4"
	// CmFaultJob310PKind key in configmap which saves the 310P FaultJob cache
	CmFaultJob310PKind = "fault-job-310P"
	// CmNodeHeartbeatKind judging node fault needs heartbeat info from former session, so should be recorded
	CmNodeHeartbeatKind = "node-heartbeat"
	// CmNodeRankTimeMapKind record map jobUID rankIndex node and times of occurrence
	CmNodeRankTimeMapKind = "node-rankIndex-Occurrence"
	// CmCheckCode Check code key
	CmCheckCode = "checkCode"

	nodeUpdateTime = 5
	// DefaultGraceOverTime time interval for grace delete
	DefaultGraceOverTime = 900
	minGraceOverTime     = 2
	maxGraceOverTime     = 3600
	maxIntervalTime      = 300
	maxRankIndex         = 1000

	// CardHealthy represents a healthy card
	CardHealthy = "Healthy"
	// CardUnhealthy represents an unhealthy card
	CardUnhealthy = "Unhealthy"
	// CardNetworkUnhealthy represents a network unhealthy card
	CardNetworkUnhealthy = "NetworkUnhealthy"
	// NodeHealthy represents node is available for scheduling
	NodeHealthy = "Healthy"
	// NodeUnhealthy represents node is unhealthy by judging heartbeat
	NodeUnhealthy = "NodeUnhealthy"
	// NodeCardUnhealthy represents node is unhealthy because of the card is unhealthy
	NodeCardUnhealthy = "CardUnhealthy"
	// NodeCardNetworkUnhealthy represents node is unhealthy because of card is network unhealthy
	NodeCardNetworkUnhealthy = "CardNetworkUnhealthy"
	// NoFaultJobsErr none fault jobs
	NoFaultJobsErr   = "none fault jobs to be restarted in cache"
	jobRestartReason = "restart for NPU malfunction"
	// NoRanksCmPre the configMap's prefix of jobs without rankTable files
	NoRanksCmPre = "env-config-"
	// JobFaultRankIDCMPre the job cm name prefix, for retraining
	JobFaultRankIDCMPre = "fault-config-"
	// JobFaultRankIDCMDataKey the job cm value key.
	JobFaultRankIDCMDataKey = "fault-npus"
	// JobRecovery Name of cm for recovery
	JobRecovery = "job-recovery"
	// DeviceFaultCmKey the key of DeviceFault info
	DeviceFaultCmKey = "huawei.com/Ascend910-Fault"
)

const (
	// SeparateNPU fault type Separate NPU
	SeparateNPU = "SeparateNPU"
	// NotHandle fault type NotHandle
	NotHandle = "NotHandle"
	// EvictType pod Evict statement
	EvictType = "Evict"
	// PreSeparateNPU fault type waiting user check
	PreSeparateNPU = "PreSeparateNPU"
	// NodeFaultCode fault type nodeUnhealthy
	NodeFaultCode = "heartbeatTimeOut"
)

// ReScheduler object for re-scheduling
type ReScheduler struct {
	*DealReSchedulerCache
	GraceDeleteTime int64
	Level           string
	Jobs            map[api.JobID]plugin.SchedulerJob
	Nodes           map[string]plugin.NPUNode
	kubeClient      kubernetes.Interface
}

// DealReSchedulerCache object with method for re-scheduler cache
type DealReSchedulerCache struct {
	*DealReSchedulerConfigmap
	FaultNodes                 []FaultNode
	FaultJobs                  []FaultJob
	NodeHeartbeats             []NodeHeartbeat
	AllocNodeRankOccurrenceMap map[api.JobID][]AllocNodeRankOccurrence
}

// DealReSchedulerConfigmap object with method for re-scheduler configmap
type DealReSchedulerConfigmap struct {
	CMName      string
	CMNameSpace string
	CMData      map[string]string
}

// AllocNodeRankOccurrence object recording node rankIndex and whether index re-allocated to new node
type AllocNodeRankOccurrence struct {
	NodeName   string
	RankIndex  string
	Occurrence int
}

// FaultCard card object for re-scheduling
type FaultCard struct {
	IsFaultCard bool
	NPUName     string
	NodeName    string
	FaultType   string
}

// FaultNode node object for re-scheduling
type FaultNode struct {
	NodeName            string
	FaultDeviceList     []FaultDeviceList
	UpdateTime          int64
	UnhealthyNPU        []string
	NetworkUnhealthyNPU []string
	IsFaultNode         bool
	NodeDEnable         bool
	NodeHealthState     string
	AllCards            []string
	FaultCards          []FaultCard
	HeartbeatInterval   int
	OldHeartbeatTime    int64
	NewHeartbeatTime    int64
	UpdateHeartbeatTime int64
}

// FaultDeviceList is the fault reason of card
type FaultDeviceList struct {
	FaultType            string `json:"fault_type"`
	NPUName              string `json:"npu_name"`
	FaultLevel           string `json:"fault_level"`
	LargeModelFaultLevel string `json:"large_model_fault_level"`
	FaultCode            string `json:"fault_code"`
}

// FaultTask object dealing with node for rescheduling
type FaultTask struct {
	Reason        []FaultReasonList
	IsFaultTask   bool
	TaskUID       api.TaskID
	TaskName      string
	TaskNamespace string
	NodeName      string
	JobName       string
	NodeRankIndex string
	UseCardName   []string
	PodCreateTime int64
	PodUID        types.UID
	faultType     string
}
type FaultReasonList struct {
	NodeName string `json:"node_name"`
	FaultDeviceList
}

// FaultJob job object for re-scheduling
type FaultJob struct {
	ReScheduleKey       string // values taken off/grace/force
	IsFaultJob          bool
	IsInSession         bool
	JobName             string
	JobUID              api.JobID
	JobNamespace        string
	JobRankIds          []string // useCardIndex + 8*NodeRankIndex
	NodeNames           []string
	FaultTasks          []FaultTask
	UpdateTime          int64
	JobRankIdCreateTime int64 // stop updating when job becomes a real fault one
	FaultTypes          []string
	DeleteExecutedFlag  bool
	ElasticScheduling   string
	ReferenceName       string
}

// NodeHeartbeat object recording nodes and their heartbeats
type NodeHeartbeat struct {
	NodeName      string
	HeartbeatTime int64
	UpdateTime    int64
}

// FaultRankIdsJobCMData used by RestoreManager for every job.
type FaultRankIdsJobCMData struct {
	FaultRankIds []string
	CreatTime    int64
}
