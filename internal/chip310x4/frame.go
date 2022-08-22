/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package chip310x4 is using for HuaWei 310 Ascend pin affinity schedule.

*/
package chip310x4

import (
	"errors"
	"fmt"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/conf"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/card310x4"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/common"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/rescheduling"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

// Name This need by frame init plugin.
func (tp *chip310x4) Name() string {
	return tp.com.Name()
}

// New return npu plugin.
func New(npuName string) plugin.HwNPUSchedulerPlugin {
	var npuPlugin = chip310x4{}
	npuPlugin.com = common.Scheduler{
		PluginName:                npuName,
		AnnoName:                  npuPlugin.GetResourceName(),
		AnnoPreVal:                npuPlugin.GetResourcePreVal(),
		DefaultJobSchedulerConfig: npuPlugin.GetPluginDefaultJobSchedulerConfig(),
	}
	npuPlugin.re = common.ReScheduler{AnnoUnHealthy: a310FaultNPUName,
		IsMyJob: npuPlugin.com.IsMyJob, AnnoName: rescheduling.AscendNPUPodRealUse}
	return &npuPlugin
}

// OnHandlerStart The npu scheduler policy initial and common processing.
func (tp *chip310x4) OnHandlerStart(sHandler *plugin.ScheduleHandler) {
	tp.com.OnHandlerStart(sHandler)
	sHandler.AddPreHandleFaultNPU(tp.com.AnnoName, tp.re.PreHandleFaultNPUFn)
}

// ValidNPUJobFn Check the compliance of the selector and resource request numbers of job.
func (tp *chip310x4) ValidNPUJobFn(job *api.JobInfo) *api.ValidateResult {
	return tp.com.ValidNPUJobFn(job)
}

// PreCheckNodeFn 310 no need to Distinguish between architecture.
func (tp *chip310x4) PreCheckNodeFn(task *api.TaskInfo, node *api.NodeInfo, confs []conf.Configuration) error {
	return tp.com.PreCheckNodeFn(task, node, confs)
}

// CheckNPUResourceStableFn Check whether the node's NPU resources are stable.
func (tp *chip310x4) CheckNPUResourceStableFn(node *api.NodeInfo) error {
	return tp.com.CheckNPUResourceStableFn(node)
}

// CheckNodeNPUByTaskFn Check whether the requested resource exists and are sufficient on the node.
func (tp *chip310x4) CheckNodeNPUByTaskFn(task *api.TaskInfo, node *api.NodeInfo, flag bool) error {
	return tp.com.CheckNodeNPUByTaskFn(task, node, flag)
}

// GetNPUAffinityBestNodesFn to implement the interface
// GetNPUAffinityBestNodesFn Initialize a mapping between nodes and priorities.
func (tp *chip310x4) GetNPUAffinityBestNodesFn(_ *api.TaskInfo, _ []*api.NodeInfo, _ bool) (map[string]int, error) {
	return nil, nil
}

// ScoreBestNPUNodesFn Used for score candidate nodes.
func (tp *chip310x4) ScoreBestNPUNodesFn(scoreMap map[string]float64,
	_ map[string]int,
	_ *api.TaskInfo,
	_ []*api.NodeInfo) (map[string]float64, error) {

	return scoreMap, nil
}

// UpdateNPUNodeUsedCardFn Update used npu resources on node.
func (tp *chip310x4) UpdateNPUNodeUsedCardFn(node *api.NodeInfo, top interface{}) error {
	return tp.com.UpdateNPUNodeUsedCardFn(node, top)
}

// GetReleaseNPUTopologyFn Get the release npu card id from task(pod).
func (tp *chip310x4) GetReleaseNPUTopologyFn(task *api.TaskInfo) (interface{}, error) {
	return tp.com.GetReleaseNPUTopologyFn(task)
}

// UpdateReleaseNPUNodeTopologyFn Update the node using npu when release pod's npu.
func (tp *chip310x4) UpdateReleaseNPUNodeTopologyFn(node *api.NodeInfo, top interface{}) error {
	return tp.com.UpdateReleaseNPUNodeTopologyFn(node, top)
}

// GetAllocatedNPUFromTopologyFn Get the pod's npu card to record in node others.
func (tp *chip310x4) GetAllocatedNPUFromTopologyFn(task *api.TaskInfo, node *api.NodeInfo, _ bool) (interface{},
	error) {
	taskNPUNumber, taskError := util.GetTaskNPUNum(task, a310NPUChipName)
	if taskError != nil {
		return nil, errors.New("no npu task")
	}

	priorityArray, err := tp.getNPUAllocPriorityArray()
	if err != nil {
		return nil, err
	}

	nodeTop := util.GetTopFromNodeOthers(node, a310NPUChipName, a310NPUCardPreName)
	if nodeTop == nil {
		klog.V(logErrorLev).Infof("not npu node[%s], no need to continue.", node.Name)
		return nil, err
	}
	klog.V(logInfoLev).Infof("%s %s[%d] priority:%v in %v.", PluginName,
		task.Name, taskNPUNumber, priorityArray, nodeTop)

	cardNumGroups := tp.getCardNumGroupsFromTop(nodeTop)
	npuNumberIndex := tp.getNPUIndex(cardNumGroups)
	var selectedNPU []int
	reqNPUNum := taskNPUNumber
	for _, priority := range priorityArray {
		curGroups, ok := npuNumberIndex[priority]
		if !ok {
			continue
		}
		if len(curGroups) >= reqNPUNum {
			selectedNPU = append(selectedNPU, curGroups[:reqNPUNum]...)
			return selectedNPU, nil
		}
		selectedNPU = append(selectedNPU, curGroups...)
		reqNPUNum -= len(curGroups)
	}
	err = fmt.Errorf("node %v not meet req: %d", nodeTop, taskNPUNumber)
	klog.V(logErrorLev).Infof("%s %s.", PluginName, err.Error())
	return nil, err
}

// SetNPUTopologyToPodFn Set the npu card ids into pod.
func (tp *chip310x4) SetNPUTopologyToPodFn(task *api.TaskInfo, top interface{}) error {
	return tp.com.SetNPUTopologyToPodFn(task, top)
}

// IsMyTask Determine if it is the NPU task of your plug-in.
func (tp *chip310x4) IsMyTask(task *api.TaskInfo) error {
	if err := tp.com.IsMyTask(task); err != nil {
		return err
	}

	if card310x4.IsTaskOfCardModeFromLabel(task) {
		return errors.New(modeNotChip)
	}

	return nil
}

// IsMyNode Determine if it is the NPU node of your plug-in.
func (tp *chip310x4) IsMyNode(node *api.NodeInfo) error {
	return tp.com.IsMyNode(node)
}

// IsMyJob Determine if it is the NPU job of your plug-in.
func (tp *chip310x4) IsMyJob(job *api.JobInfo) error {

	if err := tp.com.IsMyJob(job); err != nil {
		return err
	}

	if card310x4.IsJobOfCardModeFromLabel(job) {
		return errors.New(modeNotChip)
	}

	return nil
}

// GetResourceName get plugin NPU resource name.
func (tp *chip310x4) GetResourceName() string {
	return a310NPUChipName
}

// GetResourcePreVal get plugin NPU resource name prefix.
func (tp *chip310x4) GetResourcePreVal() string {
	return a310NPUCardPreName
}

// GetPluginDefaultJobSchedulerConfig get plugin default job scheduler config.
func (tp *chip310x4) GetPluginDefaultJobSchedulerConfig() map[string]string {
	defaultSchedulerConfig := make(map[string]string, util.NPUIndex1)
	defaultSchedulerConfig[archSelector] = huaweiArchArm + "|" + huaweiArchX86
	defaultSchedulerConfig[acceleratorType] = cardAcceleratorType + "|" + chipAcceleratorType
	return defaultSchedulerConfig
}
