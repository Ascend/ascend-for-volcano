/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package chip310P is using for HuaWei 310P Ascend pin affinity schedule.

*/
package chip310P

import (
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/conf"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/common"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/rescheduling"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

// Name This need by frame init plugin.
func (tp *chip310P) Name() string {
	return tp.com.PluginName
}

// New return npu plugin.
func New(npuName string) plugin.HwNPUSchedulerPlugin {
	var npuPlugin = chip310P{}
	npuPlugin.com = common.Scheduler{
		PluginName: npuName,
		AnnoName:   npuPlugin.GetResourceName(),
		AnnoPreVal: npuPlugin.GetResourcePreVal(),

		DefaultJobSchedulerConfig: npuPlugin.GetPluginDefaultJobSchedulerConfig(),
	}
	npuPlugin.re = common.ReScheduler{AnnoUnHealthy: a310PFaultNPUName,
		AnnoName: rescheduling.AscendNPUPodRealUse, IsMyJob: npuPlugin.com.IsMyJob}
	return &npuPlugin
}

// OnHandlerStart The npu scheduler policy initial and common processing.
func (tp *chip310P) OnHandlerStart(sHandler *plugin.ScheduleHandler) {
	tp.com.OnHandlerStart(sHandler)
	sHandler.AddPreHandleFaultNPU(tp.com.AnnoName, tp.re.PreHandleFaultNPUFn)
}

// ValidNPUJobFn Check the compliance of the selector and resource request numbers of job.
func (tp *chip310P) ValidNPUJobFn(job *api.JobInfo) *api.ValidateResult {
	return tp.com.ValidNPUJobFn(job)
}

// PreCheckNodeFn 310P no need to Distinguish between architecture.
func (tp *chip310P) PreCheckNodeFn(task *api.TaskInfo, node *api.NodeInfo, confs []conf.Configuration) error {
	return tp.com.PreCheckNodeFn(task, node, confs)
}

// CheckNPUResourceStableFn Check whether the node's NPU resources are stable.
func (tp *chip310P) CheckNPUResourceStableFn(node *api.NodeInfo) error {
	return tp.com.CheckNPUResourceStableFn(node)
}

// CheckNodeNPUByTaskFn Check whether the requested resource exists and are sufficient on the node.
func (tp *chip310P) CheckNodeNPUByTaskFn(task *api.TaskInfo, node *api.NodeInfo, distributeFlag bool) error {
	return tp.com.CheckNodeNPUByTaskFn(task, node, distributeFlag)
}

// GetNPUAffinityBestNodesFn to implement the interface
// GetNPUAffinityBestNodesFn Initialize a mapping between nodes and priorities.
func (tp *chip310P) GetNPUAffinityBestNodesFn(_ *api.TaskInfo, _ []*api.NodeInfo, _ bool) (map[string]int, error) {
	return nil, nil
}

// ScoreBestNPUNodesFn Used for score candidate nodes.
func (tp *chip310P) ScoreBestNPUNodesFn(scoreMap map[string]float64,
	_ map[string]int,
	_ *api.TaskInfo,
	_ []*api.NodeInfo) (map[string]float64, error) {

	return scoreMap, nil
}

// UpdateNPUNodeUsedCardFn Update used npu resources on node.
func (tp *chip310P) UpdateNPUNodeUsedCardFn(node *api.NodeInfo, top interface{}) error {
	return tp.com.UpdateNPUNodeUsedCardFn(node, top)
}

// GetReleaseNPUTopologyFn Get the release npu card id from task(pod).
func (tp *chip310P) GetReleaseNPUTopologyFn(task *api.TaskInfo) (interface{}, error) {
	return tp.com.GetReleaseNPUTopologyFn(task)
}

// UpdateReleaseNPUNodeTopologyFn Update the node using npu when release pod's npu.
func (tp *chip310P) UpdateReleaseNPUNodeTopologyFn(node *api.NodeInfo, top interface{}) error {
	return tp.com.UpdateReleaseNPUNodeTopologyFn(node, top)
}

// GetAllocatedNPUFromTopologyFn Get the pod's npu card to record in node others.
func (tp *chip310P) GetAllocatedNPUFromTopologyFn(task *api.TaskInfo, node *api.NodeInfo,
	disFlag bool) (interface{}, error) {
	return tp.com.GetAllocatedNPUFromTopologyFn(task, node, disFlag)
}

// SetNPUTopologyToPodFn Set the npu card ids into pod.
func (tp *chip310P) SetNPUTopologyToPodFn(task *api.TaskInfo, top interface{}) error {
	return tp.com.SetNPUTopologyToPodFn(task, top)
}

// IsMyTask Determine if it is the NPU task of your plug-in.
func (tp *chip310P) IsMyTask(task *api.TaskInfo) error {
	return tp.com.IsMyTask(task)
}

// IsMyNode Determine if it is the NPU node of your plug-in.
func (tp *chip310P) IsMyNode(node *api.NodeInfo) error {
	return tp.com.IsMyNode(node)
}

// IsMyJob Determine if it is the NPU job of your plug-in.
func (tp *chip310P) IsMyJob(job *api.JobInfo) error {
	return tp.com.IsMyJob(job)
}

// GetResourceName get plugin NPU resource name.
func (tp *chip310P) GetResourceName() string {
	return a310PNPUChipName
}

// GetResourcePreVal get plugin NPU resource name prefix.
func (tp *chip310P) GetResourcePreVal() string {
	return a310PNPUCardPreName
}

// GetPluginDefaultJobSchedulerConfig get plugin default job scheduler config.
func (tp *chip310P) GetPluginDefaultJobSchedulerConfig() map[string]string {
	defaultSchedulerConfig := make(map[string]string, util.NPUIndex1)
	defaultSchedulerConfig[archSelector] = huaweiArchArm + "|" + huaweiArchX86
	return defaultSchedulerConfig
}
