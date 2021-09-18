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

Package rescheduling is using for HuaWei Ascend pin affinity schedule utilities.

*/
package rescheduling

const (
	// CmJobKind the fault job data record in configmap.
	CmJobKind = "job"
	// CmNodeKind the fault node data record in configmap.
	CmNodeKind = "node"
	// CmCardKind the fault card data record in configmap.
	CmCardKind      = "card"
	logErrorLev     = 1
	logInfoLev      = 3
	logDebugLev     = 4
	constIntNum2    = 2
	constIntNum3    = 3
	node910X8NPUNum = 8
	maxIntervalTime = 300
	maxRankIndex    = 1000
	archSelector    = "host-arch"
	huaweiArchX86   = "huawei-x86"
	huaweiArchArm   = "huawei-arm"
	cmNameSpace     = "volcano-system"
	cmName          = "vcjob-fault-npu-cm"
	// node inoperable interval time(s)
	nodeUpdateTime        = 15
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
	nodeDEnableKey             = "nodeDEnable"
	nodeDEnableOnValue         = "on"
	nodeDEnableOffValue        = "off"
	npu800And9000CardName      = "huawei.com/Ascend910"
	npu800And9000CardPreName   = "Ascend910-"
)

// ReSchedulerTasks record the tasks using the failed NPU.
// The key in ReSchedulerCache is jobID
type ReSchedulerTasks struct {
	// Key is taskName.
	NodeNames   map[string]string
	RankIndexes map[string]string
	Time        map[string]int64
	TaskUseNPUs map[string]string
	NameSpace   string
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

var reSchedulerJobController = make(map[string]struct{}, constIntNum3)
