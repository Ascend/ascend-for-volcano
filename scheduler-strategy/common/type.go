/*
Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package common is using for HuaWei common infer Ascend pin affinity schedule.

*/
package common

import "volcano.sh/volcano/pkg/scheduler/api"

// common type
const (
	NodeNPUNumber = 64
	ConstIntNum1  = 1

	LogErrorLev = 1
	LogInfoLev  = 3
	LogDebugLev = 4

	PodPredicateTime = "predicate-time"

	NodesNoMeetNPUReqError  = "insufficient npus on the schedulable nodes in cluster"
	NodeNotStableWarning    = "the npus on this node are unstable"
	NodeNotEnoughNPUWarning = "insufficient number of available npus on this node"

	JobNoNPUCard           = "job no use npu"
	ArgumentError          = "invalid argument"
	jobRestartReason       = "restart for NPU malfunction"
	faultSchedulingLabel   = "fault-scheduling"
	onFaultSchedulingLabel = "grace"
)

// Scheduler common scheduler
type Scheduler struct {
	// PluginName plugin name
	PluginName string
	// AnnoName annonation map Name
	AnnoName string
	// AnnoPreVal annonation pre val
	AnnoPreVal string
	// DefaultJobSchedulerConfig label map
	DefaultJobSchedulerConfig map[string]string
}

// ReScheduler rescheduler struct
type ReScheduler struct {
	// IsMyJob judge job
	IsMyJob func(*api.JobInfo) error
	// AnnoUnHealthy unhealthy key
	AnnoUnHealthy string
	// AnnoName annonation map Name
	AnnoName string
}
