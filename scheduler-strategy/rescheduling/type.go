/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package rescheduling is using for HuaWei Ascend pin affinity schedule utilities.

*/
package rescheduling

import (
	"k8s.io/apimachinery/pkg/types"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/util"
)

const (
	// CmJobKind the fault job data record in configmap.
	CmJobKind = "job"
	// CmNodeKind the fault node data record in configmap.
	CmNodeKind = "node"
	// CmCardKind the fault card data record in configmap.
	CmCardKind = "card"
	// CmNodeHeartbeatKind the node  heartbeat record in configmap.
	CmNodeHeartbeatKind = "heartbeat"
	// CmJobRankIds the jobs rankIds for .
	CmJobRankIds = "rankIds"
	// TmpAllocRankIndexKind used for allocated rankIndex in one session.
	TmpAllocRankIndexKind = "allocRankIndex"
	node910X8NPUNum       = 8
	maxIntervalTime       = 300
	maxRankIndex          = 1000
	cmNameSpace           = "volcano-system"
	cmName                = "vcjob-fault-npu-cm"
	// node inoperable interval time(s)
	nodeUpdateTime        = 5
	graceOverTime         = 900
	nodeHeartbeat         = "noded/heartbeat"
	nodeHeartbeatInterval = "noded/heartbeat-interval"
	faultNPU              = "huawei.com/Ascend910-Unhealthy"
	networkUnhealthyNPU   = "huawei.com/Ascend910-NetworkUnhealthy"
	podRankIndex          = "hccl/rankIndex"
	jobRescheduleLabelKey = "fault-scheduling"
	// JobGraceRescheduleLabelValue Grace delete reschedule job.
	JobGraceRescheduleLabelValue = "grace"
	// JobForceRescheduleLabelValue Force delete reschedule job.
	JobForceRescheduleLabelValue = "force"
	// JobOffRescheduleLabelValue not delete reschedule job.
	JobOffRescheduleLabelValue = "off"
	// JobFaultRankIDCMDataKey the job cm value key.
	JobFaultRankIDCMDataKey = "fault-npus"
	// JobFaultRankIDCMPre the job cm name prefix.
	JobFaultRankIDCMPre      = "fault-config-"
	nodeDEnableKey           = "nodeDEnable"
	nodeDEnableOnValue       = "on"
	nodeDEnableOffValue      = "off"
	npu800And9000CardName    = "huawei.com/Ascend910"
	npu800And9000CardPreName = "Ascend910-"
	// GraceOverTimeKey for GraceOverTime config by user
	GraceOverTimeKey = "grace-over-time"
)

// GraceOverTime for rescheduling units: seconds
var GraceOverTime int64 = graceOverTime

// ReSchedulerTasks record the tasks using the failed NPU.
// The key in ReSchedulerCache is jobID
type ReSchedulerTasks struct {
	// All values are stored in order by taskName.
	TaskName      []string
	NodeNames     []string
	RankIndexes   []string
	Time          []int64
	TaskUseNPUs   []string
	NameSpace     string
	GraceTaskFlag bool
	UpdateTime    int64
}

// FaultNodeState record the inoperable node in k8s.
// The key in ReSchedulerCache is node name.
type FaultNodeState struct {
	NodeName string
	// Inoperable node's code.
	HealthCode int
	// Fault node code update Time.
	UpdateTime int64
	// nodeD record Time.
	Heartbeat int64
	// nodeD Heartbeat interval time.
	HeartbeatInterval int
}

// FaultNPUsOnNode Record the node's corresponding fault chips.
type FaultNPUsOnNode struct {
	// Fault NPUs's node name.
	NodeName string
	// Fault NPUs that in one nodes.
	FaultNPUs []string
	// network unhealthy NPUs that in one nodes.
	NetworkUnhealthyNPUs []string
	// Fault card update Time.
	UpdateTime int64
}

// ReSchedulerCache record the inoperable jobs/nodes/pods in cm and buffer.
// The structure is job:JobID:ReSchedulerTasks/node:NodeName:FaultNodeState/card:NodeName:reserved.
var ReSchedulerCache map[string]interface{}

type faultNPUJobBase struct {
	jobName   string
	namespace string
	// task name:rank index
	taskUseRankIndex map[string]string
	// task name:node name
	taskUseNode map[string]string
}

// FaultNPUJob While the program is running, record fault job information.
type FaultNPUJob struct {
	faultNPUJobBase
	// task name:task annotation
	taskUseNPUs map[string]string
}

// NormalNodeHeartbeat Record the heartbeat of a node to determine whether it is healthy.
// map key is nodeName.
type NormalNodeHeartbeat struct {
	// nodeD send.
	NodeDHeartbeat int64
	// The time recorded by the node where volcano is located when NodeDHeartbeat changed.
	UpdateHeartbeatTime int64
	// nodeD Heartbeat interval time, need multiply by 3.
	HeartbeatInterval int
	// The time recorded last update.
	UpdateTime int64
}

// TaskUsedRankIndex Record the fault node used rankIndex.
// This is used for write pod's rankIndex when task used new node.
type TaskUsedRankIndex struct {
	// nodeD Heartbeat interval time, need multiply by 3.
	FaultNodeRankIndex map[string]struct{ UpdateTime int64 }
	// The time recorded last update.
	UpdateTime int64
}

var reSchedulerJobController = make(map[string]struct{}, util.ConstIntNum3)

// FaultRankIDRecordJobCMData record in volcano fault cm, key is job's uuid.
type FaultRankIDRecordJobCMData struct {
	NameSpace    string
	FaultRankIds string
	// key is podName,value is pod create time
	PodsName      []string
	PodsUID       []types.UID
	PodsCreatTime []int64
	// this the record create time.
	CreatTime int64
}

// FaultRankIdsJobCMData used by RestoreManager for every job.
type FaultRankIdsJobCMData struct {
	FaultRankIds string
	CreatTime    int64
}
