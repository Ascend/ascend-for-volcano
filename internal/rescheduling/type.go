/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
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

	nodeHeartbeat         = "noded/heartbeat"
	nodeHeartbeatInterval = "noded/heartbeat-interval"
	faultNPU              = "huawei.com/Ascend910-Unhealthy"
	networkUnhealthyNPU   = "huawei.com/Ascend910-NetworkUnhealthy"

	nodeDEnableKey      = "nodeDEnable"
	nodeDEnableOnValue  = "on"
	nodeDEnableOffValue = "off"

	cardMode = "card"
	modeMode = "module"
	chipMode = "chip"

	podRankIndex = "hccl/rankIndex"

	// CmFaultNodeKind key in configmap which saves the FaultNode cache
	CmFaultNodeKind = "fault-node"
	// CmFaultJob910x8Kind key in configmap which saves the 910x8 FaultJob cache
	CmFaultJob910x8Kind = "fault-job-910x8"
	// CmFaultJob910x2Kind key in configmap which saves the 910x8 FaultJob cache
	CmFaultJob910x2Kind = "fault-job-910x2"
	// CmFaultJob310x4Kind key in configmap which saves the 310x4 FaultJob cache
	CmFaultJob310x4Kind = "fault-job-310x4"
	// CmFaultJob310PKind key in configmap which saves the 310P FaultJob cache
	CmFaultJob310PKind = "fault-job-310P"
	// CmNodeHeartbeatKind judging node fault needs heartbeat info from former session, so should be recorded
	CmNodeHeartbeatKind = "node-heartbeat"
	// CmNodeRankTimeMapKind record map[jobUID]{nodeName, nodeRankIndex, Occurrence}, if occurrence=1 the rankIndex of faultNode has been assigned to other node already
	CmNodeRankTimeMapKind = "node-rankIndex-Occurrence"

	nodeUpdateTime       = 5
	defaultGraceOverTime = 900
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
	NodeUnhealthy = "Unhealthy"
	// NodeCardUnhealthy represents node is unhealthy because of the card is unhealthy
	NodeCardUnhealthy = "CardUnhealthy"
	// NodeCardNetworkUnhealthy represents node is unhealthy because of card is network unhealthy
	NodeCardNetworkUnhealthy = "CardNetworkUnhealthy"

	jobRestartReason = "restart for NPU malfunction"
	// JobFaultRankIDCMPre the job cm name prefix, for retraining
	JobFaultRankIDCMPre = "fault-config-"
	// JobFaultRankIDCMDataKey the job cm value key.
	JobFaultRankIDCMDataKey = "fault-npus"
	// JobRecovery Name of cm for recovery
	JobRecovery = "job-recovery"
)

type IReschedule interface {
	IsFaultJob(*api.JobInfo) error
	GetFaultRescheduleLevel() string
}

type ReScheduler struct {
	*DealReSchedulerCache
	GraceDeleteTime int64
	Level           string
	Jobs            map[api.JobID]plugin.SchedulerJob
	Nodes           map[string]plugin.NPUNode
	kubeClient      kubernetes.Interface
}

type DealReSchedulerCache struct {
	*DealReSchedulerConfigmap
	FaultNodes                 []FaultNode
	FaultJobs                  []FaultJob
	NodeHeartbeats             []NodeHeartbeat
	AllocNodeRankOccurrenceMap map[api.JobID][]AllocNodeRankOccurrence
}

type DealReSchedulerConfigmap struct {
	CMName      string
	CMNameSpace string
	CMData      map[string]string
}

type AllocNodeRankOccurrence struct {
	NodeName   string
	RankIndex  string
	Occurrence int
}

type FaultCard struct {
	IsFaultCard bool
	NPUName     string
	NodeName    string
	FaultType   string
}

type FaultNode struct {
	NodeName            string
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

type FaultTask struct {
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
	JobRankIdCreateTime int64
	FaultTypes          []string
	DeleteExecutedFlag  bool
}

type NodeHeartbeat struct {
	NodeName      string
	HeartbeatTime int64
	UpdateTime    int64
}

type NodeIndexTaskMap struct {
}

// FaultRankIdsJobCMData used by RestoreManager for every job.
type FaultRankIdsJobCMData struct {
	FaultRankIds []string
	CreatTime    int64
}
