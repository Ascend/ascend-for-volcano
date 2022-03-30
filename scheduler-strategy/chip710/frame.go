/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package chip710 is using for HuaWei 710 Ascend pin affinity schedule.

*/
package chip710

import (
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/conf"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/common"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/util"
)

// Name This need by frame init plugin.
func (tp *chip710) Name() string {
	return tp.com.PluginName
}

// New return npu plugin.
func New(npuName string) plugin.HwNPUSchedulerPlugin {
	var npuPlugin = chip710{}
	npuPlugin.com = common.Scheduler{
		PluginName: npuName,
		AnnoName:   npuPlugin.GetResourceName(),
		AnnoPreVal: npuPlugin.GetResourcePreVal(),

		DefaultJobSchedulerConfig: npuPlugin.GetPluginDefaultJobSchedulerConfig(),
	}
	npuPlugin.re = common.ReScheduler{AnnoUnHealthy: a710FaultNPUName,
		AnnoName: npuPlugin.com.AnnoName, IsMyJob: npuPlugin.com.IsMyJob}
	return &npuPlugin
}

// OnHandlerStart The npu scheduler policy initial and common processing.
func (tp *chip710) OnHandlerStart(sHandler *plugin.ScheduleHandler) {
	tp.com.OnHandlerStart(sHandler)
	sHandler.AddPreHandleFaultNPU(tp.com.AnnoName, tp.re.PreHandleFaultNPUFn)
}

// ValidNPUJobFn Check the compliance of the selector and resource request numbers of job.
func (tp *chip710) ValidNPUJobFn(job *api.JobInfo) *api.ValidateResult {
	return tp.com.ValidNPUJobFn(job)
}

// PreCheckNodeFn 710 no need to Distinguish between architecture.
func (tp *chip710) PreCheckNodeFn(task *api.TaskInfo, node *api.NodeInfo, confs []conf.Configuration) error {
	return tp.com.PreCheckNodeFn(task, node, confs)
}

// CheckNPUResourceStableFn Check whether the node's NPU resources are stable.
func (tp *chip710) CheckNPUResourceStableFn(node *api.NodeInfo) error {
	return tp.com.CheckNPUResourceStableFn(node)
}

// CheckNodeNPUByTaskFn Check whether the requested resource exists and are sufficient on the node.
func (tp *chip710) CheckNodeNPUByTaskFn(task *api.TaskInfo, node *api.NodeInfo, distributeFlag bool) error {
	return tp.com.CheckNodeNPUByTaskFn(task, node, distributeFlag)
}

// GetNPUAffinityBestNodesFn to implement the interface
// GetNPUAffinityBestNodesFn Initialize a mapping between nodes and priorities.
func (tp *chip710) GetNPUAffinityBestNodesFn(_ *api.TaskInfo, _ []*api.NodeInfo, _ bool) (map[string]int, error) {
	return nil, nil
}

// ScoreBestNPUNodesFn Used for score candidate nodes.
func (tp *chip710) ScoreBestNPUNodesFn(scoreMap map[string]float64,
	_ map[string]int,
	_ *api.TaskInfo,
	_ []*api.NodeInfo) (map[string]float64, error) {

	return scoreMap, nil
}

// UpdateNPUNodeUsedCardFn Update used npu resources on node.
func (tp *chip710) UpdateNPUNodeUsedCardFn(node *api.NodeInfo, top interface{}) error {
	return tp.com.UpdateNPUNodeUsedCardFn(node, top)
}

// GetReleaseNPUTopologyFn Get the release npu card id from task(pod).
func (tp *chip710) GetReleaseNPUTopologyFn(task *api.TaskInfo) (interface{}, error) {
	return tp.com.GetReleaseNPUTopologyFn(task)
}

// UpdateReleaseNPUNodeTopologyFn Update the node using npu when release pod's npu.
func (tp *chip710) UpdateReleaseNPUNodeTopologyFn(node *api.NodeInfo, top interface{}) error {
	return tp.com.UpdateReleaseNPUNodeTopologyFn(node, top)
}

// GetAllocatedNPUFromTopologyFn Get the pod's npu card to record in node others.
func (tp *chip710) GetAllocatedNPUFromTopologyFn(task *api.TaskInfo, node *api.NodeInfo,
	disFlag bool) (interface{}, error) {
	return tp.com.GetAllocatedNPUFromTopologyFn(task, node, disFlag)
}

// SetNPUTopologyToPodFn Set the npu card ids into pod.
func (tp *chip710) SetNPUTopologyToPodFn(task *api.TaskInfo, top interface{}) error {
	return tp.com.SetNPUTopologyToPodFn(task, top)
}

// IsMyTask Determine if it is the NPU task of your plug-in.
func (tp *chip710) IsMyTask(task *api.TaskInfo) error {
	return tp.com.IsMyTask(task)
}

// IsMyNode Determine if it is the NPU node of your plug-in.
func (tp *chip710) IsMyNode(node *api.NodeInfo) error {
	return tp.com.IsMyNode(node)
}

// IsMyJob Determine if it is the NPU job of your plug-in.
func (tp *chip710) IsMyJob(job *api.JobInfo) error {
	return tp.com.IsMyJob(job)
}

// GetResourceName get plugin NPU resource name.
func (tp *chip710) GetResourceName() string {
	return a710NPUChipName
}

// GetResourcePreVal get plugin NPU resource name prefix.
func (tp *chip710) GetResourcePreVal() string {
	return a710NPUCardPreName
}

// GetPluginDefaultJobSchedulerConfig get plugin default job scheduler config.
func (tp *chip710) GetPluginDefaultJobSchedulerConfig() map[string]string {
	defaultSchedulerConfig := make(map[string]string, util.ConstIntNum1)
	defaultSchedulerConfig[archSelector] = huaweiArchArm + "|" + huaweiArchX86
	return defaultSchedulerConfig
}
