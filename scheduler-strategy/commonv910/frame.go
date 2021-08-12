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

Package commonv910 is using for virtual HuaWei Ascend910 schedule.

*/
package commonv910

import (
	"errors"
	"fmt"
	"k8s.io/klog"
	"reflect"
	"strconv"
	"strings"
	"time"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/conf"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	hwutil "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/util"
)

// VnpuType types of vNPU
var VnpuType []string

func init() {
	VnpuType = append(VnpuType, npuV910CardName2c)
	VnpuType = append(VnpuType, npuV910CardName4c)
	VnpuType = append(VnpuType, npuV910CardName8c)
	VnpuType = append(VnpuType, npuV910CardName16c)
}

// Name Get plugin name for frame init
func (tp *Vnpu) Name() string {
	return PluginName
}

// OnHandlerStart Vnpu scheduler policy initial and common processing
func (tp *Vnpu) OnHandlerStart(sHandler *plugin.ScheduleHandler) {
	klog.V(logInfoLev).Infof("%v start Handler.", tp.Name())
	sHandler.AddInitNodesNPUAllocTopology(tp.Name(), initVNodesFn)
}

// IsMyTask used for identify Vnpu task, need to be implemented by vNPU plugins
func (tp *Vnpu) IsMyTask(task *api.TaskInfo) error {
	// IsMyTask choose one and one only type of plugin and return nil, so it will return error by default.
	// This function must be overwritten by whichever plugin that encapsulate Vnpu, otherwise will return error here.
	return errors.New("isMyTask is not overwritten by Vnpu")
}

// IsMyNode used for identify Vnpu node, need to be implemented by vNPU plugins
func (tp *Vnpu) IsMyNode(node *api.NodeInfo) error {
	// IsMyNode choose one and one only type of plugin and return nil, so it will return error by default.
	// This function must be overwritten by whichever plugin that encapsulate Vnpu, otherwise will return error here.
	return errors.New("isMyNode is not overwritten by Vnpu")
}

// IsMyJob used for identify Vnpu job, need to be implemented by vNPU plugins
func (tp *Vnpu) IsMyJob(job *api.JobInfo) error {
	// IsMyJob choose one and one only type of plugin and return nil, so it will return error by default.
	// This function must be overwritten by whichever plugin that encapsulate Vnpu, otherwise will return error here.
	return errors.New("isMyJob is not overwritten by Vnpu")
}

// ValidNPUJobFn check the compliance of the selector and resource request numbers
func (tp *Vnpu) ValidNPUJobFn(job *api.JobInfo) *api.ValidateResult {
	// 1.valid npu job selector
	if err := tp.validNPUJobSelector(job); err != nil {
		klog.V(logErrorLev).Infof("%s err: %v.", tp.Name(), err)
		return &api.ValidateResult{
			Pass:    false,
			Reason:  err.Error(),
			Message: fmt.Sprintf("validNPUJob err: %v", err),
		}
	}
	// 2.valid the resource type and the number of resources the job request
	if errRs := validJobResource(job); errRs != nil {
		klog.V(logErrorLev).Infof("%s err: %v.", tp.Name(), errRs)
		return &api.ValidateResult{
			Pass:    false,
			Reason:  "job resource requested error",
			Message: fmt.Sprintf("%s, err: %v", job.Name, errRs),
		}
	}

	return nil
}

// PreCheckNodeFn check whether the node matches the tag requirements of the task.
func (tp *Vnpu) PreCheckNodeFn(task *api.TaskInfo, node *api.NodeInfo, confs []conf.Configuration) error {
	schedulerConf := hwutil.GetSchedulerSelectorConfig(confs)
	if len(schedulerConf) == 0 {
		klog.V(logErrorLev).Infof("%s JobName: %s get selector nil.", tp.Name(), task.Name)
		return fmt.Errorf("%s get scheduler selector nil", node.Name)
	}

	// select node by architect
	if err := isSelectorMeetNode(task, node, schedulerConf); err != nil {
		// get scheduler selector configure failed, but need continue
		klog.V(logErrorLev).Infof("%s taskName: %s ,nodeName %s : %v.", tp.Name(), task.Name, node.Name, err)
		return fmt.Errorf("task(%s) in node(%s):%v", task.Name, node.Name, err)
	}

	return nil
}

// CheckNPUResourceStableFn check whether the resources on the node are stable
func (tp *Vnpu) CheckNPUResourceStableFn(node *api.NodeInfo) error {
	for _, vType := range VnpuType {
		nodeNPUIdleNumFromTop, err := getNPUNumFromNodeAnnotation(node, vType)
		if err != nil {
			klog.V(logInfoLev).Infof("getNodeNPUNumFromAnnotation node %s doesn't have %s.", node.Name, vType)
			continue
		}

		nodeNPUIdleNumFromIdle, err := hwutil.GetNodeNPUNumFromIdle(node, vType)
		if err != nil {
			return fmt.Errorf("getNodeNPUNumFromIdle %s : %s", nodesNoMeetNPUReqError, err)
		}

		if err = hwutil.CheckNodeNPUStabilize(nodeNPUIdleNumFromTop, nodeNPUIdleNumFromIdle); err != nil {
			return fmt.Errorf("node %s %s : %s", node.Name, nodeNotStableWarning, err)
		}
	}

	return nil
}

// CheckNodeNPUByTaskFn check whether the requested resource exists and are sufficient on the node
func (tp *Vnpu) CheckNodeNPUByTaskFn(task *api.TaskInfo, node *api.NodeInfo) error {
	total := 0
	for _, vType := range VnpuType {
		taskVnpu, taskError := hwutil.GetTaskNPUNum(task, vType)
		if taskError != nil {
			klog.V(logDebugLev).Infof("checkVNodeNPUByTask %s doesn't requests %s.", task.Name, vType)
			continue
		}

		nodeVnpu, err := getNPUNumFromNodeAnnotation(node, vType)
		if err != nil || nodeVnpu < taskVnpu {
			klog.V(logInfoLev).Infof("%s checkVNodeNPUByTask nil, node name:%s(top:%v),task req %s:%d.",
				tp.Name(), node.Name, nodeVnpu, vType, taskVnpu)
			return fmt.Errorf("%s:get Vnpu nil", nodeNotEnoughVnpuWarning)
		}

		total += nodeVnpu * vnpuCoefficients[vType]
	}

	if tp.MaxNPUNum > 0 && total > npu910CardCoef*tp.MaxNPUNum {
		return fmt.Errorf("total amount of npu (%d)c excceeded the maximum limitation of (%d x %d)c",
			total, tp.MaxNPUNum, npu910CardCoef)
	}

	return nil
}

// GetNPUAffinityBestNodesFn initialize a mapping between nodes and priorities
func (tp *Vnpu) GetNPUAffinityBestNodesFn(task *api.TaskInfo, nodes []*api.NodeInfo) (map[string]int, error) {
	klog.V(logDebugLev).Infof("Get NPU affinity best node for task %s.", task.Name)
	var bestNodesMap = make(map[string]int, const2)

	for _, node := range nodes {
		if node == nil {
			continue
		}
		bestNodesMap[node.Name] = 0
	}

	return bestNodesMap, nil
}

// ScoreBestNPUNodesFn used for score candidate nodes
func (tp *Vnpu) ScoreBestNPUNodesFn(scoreMap map[string]float64,
	bestNodes map[string]int,
	task *api.TaskInfo,
	nodes []*api.NodeInfo) (float64s map[string]float64, e error) {
	if reflect.ValueOf(scoreMap).IsNil() {
		err := errors.New("scoreBestNPUNodes's scoreMap is nil")
		klog.V(logInfoLev).Infof("%s %s %v.", nodes, tp.Name(), err)
		return nil, err
	}

	for nodeName, priority := range bestNodes {
		if _, ok := scoreMap[nodeName]; ok {
			scoreMap[nodeName] = float64(priority)
			continue
		}

		scoreMap[nodeName] = 0.0
	}

	return scoreMap, nil
}

// GetAllocatedNPUFromTopologyFn obtain the name of the allocated devices
func (tp *Vnpu) GetAllocatedNPUFromTopologyFn(task *api.TaskInfo, node *api.NodeInfo) (interface{}, error) {
	var allocTopologyVnpus []string

	for _, vType := range VnpuType {
		reqNum, taskError := hwutil.GetTaskNPUNum(task, vType)
		if taskError != nil {
			klog.V(logDebugLev).Infof("checkVNodeNPUByTask %s doesn't requests %s.", task.Name, vType)
			continue
		}

		vNPUByPriority, err := getVCardWithLeastRemainPw(node.Node.Annotations, vType)
		if err != nil {
			klog.V(logErrorLev).Infof("GetAllocatedNPUFromTopologyFn allocate error: %s", err)
			return allocTopologyVnpus, err
		}
		allocTopologyVnpus = vNPUByPriority[:reqNum]

		return allocTopologyVnpus, nil
	}

	klog.V(logErrorLev).Infof("allocateVnpuFromTopology task %s request no Vnpu resources.", task.Name)
	return allocTopologyVnpus, errors.New("task request no Vnpu resources")
}

// SetNPUTopologyToPodFn write the name of the allocated devices to Pod
func (tp *Vnpu) SetNPUTopologyToPodFn(task *api.TaskInfo, top interface{}) error {
	var topStr string
	var vType string

	topArr, ok := top.([]string)
	if !ok || len(topArr) < 1 {
		return errors.New("SetNPUTopologyToPod gets invalid argument")
	}

	topInstance := topArr[0]
	for _, vt := range VnpuType {
		v := strings.TrimPrefix(vt, npu910CardNamePrefix)
		if strings.HasPrefix(topInstance, v) {
			vType = vt
			break
		}
	}

	topStr = strings.Join(topArr, ",")
	klog.V(logInfoLev).Infof("%s setNPUTopologyToPod begin top:%v.", tp.Name(), top)
	task.Pod.Annotations[vType] = topStr
	task.Pod.Annotations[podPredicateTime] = strconv.FormatInt(time.Now().UnixNano(), 10)
	klog.V(logInfoLev).Infof("%s setNPUTopologyToPod %s top:%s.", tp.Name(), task.Name, topStr)

	return nil
}

// UpdateNPUNodeUsedCardFn update node others after allocate
func (tp *Vnpu) UpdateNPUNodeUsedCardFn(node *api.NodeInfo, top interface{}) error {
	if ok := updateNPUNodeTopology(node, top, updateTopStrOfNodeOtherAlloc); ok != nil {
		return errors.New("update npu node topology failed")
	}

	return nil
}

// GetReleaseNPUTopologyFn obtain allocated device info from Pod
func (tp *Vnpu) GetReleaseNPUTopologyFn(task *api.TaskInfo) (interface{}, error) {
	var vType string
	var taskTopArr []string
	var err error

	for _, vType = range VnpuType {
		taskTopArr, err = getNPUsFromNodeAnnotation(task.Pod.Annotations, vType)
		if err != nil {
			continue
		}
		break
	}

	if taskTopArr == nil {
		klog.V(logErrorLev).Infof("%s getReleaseVnpuTopology pod annotation for %s is empty.", tp.Name(), vType)
		return taskTopArr, errors.New("task pod annotation is empty")
	}

	return taskTopArr, nil
}

// UpdateReleaseNPUNodeTopologyFn update node others after release
func (tp *Vnpu) UpdateReleaseNPUNodeTopologyFn(node *api.NodeInfo, top interface{}) error {
	if ok := updateNPUNodeTopology(node, top, updateTopStrOfNodeOtherRelease); ok != nil {
		return errors.New("update npu node topology after release failed")
	}

	return nil
}
