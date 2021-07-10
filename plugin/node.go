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

Package plugin is using for HuaWei Ascend pin affinity schedule frame.

*/
package plugin

import (
	"errors"
	"fmt"
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/conf"
	"volcano.sh/volcano/pkg/scheduler/framework"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/util"
)

// Init hw npu nodes, used in npu plugins.
func (hwNPU *ScheduleHandler) initNodesNPUAllocTopology(nodes map[string]*api.NodeInfo) error {
	for cardName, initNodes := range hwNPU.InitNodesNPUAllocTopologyFns {
		if err := initNodes(nodes); err != nil {
			klog.V(logErrorLev).Infof("%s InitNodesNPUAllocTopology :%v.", cardName, err)
			return err
		}
	}

	return nil
}

// preHandleFaultNPUFn Pretreatment of NPU faults.
func (hwNPU *ScheduleHandler) preHandleFaultNPUFn(ssn *framework.Session) error {
	for pluginName, preHandleFaultNPU := range hwNPU.PreHandleFaultNPUFns {
		if err := preHandleFaultNPU(ssn); err != nil {
			klog.V(logDebugLev).Infof("%s preHandleFaultNPU :%v.", pluginName, err)
			return err
		}
	}

	return nil
}

func (hwNPU *ScheduleHandler) preCheckNode(task *api.TaskInfo, node *api.NodeInfo, confs []conf.Configuration) error {
	curNPUPlugin := hwNPU.getNPUPlugin(task)
	if curNPUPlugin == nil {
		return nil
	}

	return curNPUPlugin.PreCheckNodeFn(task, node, confs)
}

func (hwNPU *ScheduleHandler) isHwNPUTask(task *api.TaskInfo) error {
	curNPUPlugin := hwNPU.getNPUPlugin(task)
	if curNPUPlugin == nil {
		return errors.New(noneNPUPlugin)
	}

	return curNPUPlugin.IsMyTask(task)
}

func (hwNPU *ScheduleHandler) isHwNPUNode(task *api.TaskInfo, node *api.NodeInfo) error {
	curNPUPlugin := hwNPU.getNPUPlugin(task)
	if curNPUPlugin == nil {
		return nil
	}

	return curNPUPlugin.IsMyNode(node)
}

func (hwNPU *ScheduleHandler) checkNodeNPUByTask(task *api.TaskInfo, node *api.NodeInfo) error {
	curNPUPlugin := hwNPU.getNPUPlugin(task)
	if curNPUPlugin == nil {
		return nil
	}

	return curNPUPlugin.CheckNodeNPUByTaskFn(task, node)
}

func (hwNPU *ScheduleHandler) getNPUAffinityBestNodes(
	task *api.TaskInfo,
	nodes []*api.NodeInfo) (map[string]int, error) {
	curNPUPlugin := hwNPU.getNPUPlugin(task)
	if curNPUPlugin == nil {
		return nil, errors.New("get npu plugin nil")
	}

	return curNPUPlugin.GetNPUAffinityBestNodesFn(task, nodes)
}

func (hwNPU *ScheduleHandler) scoreBestNPUNodes(
	task *api.TaskInfo,
	scoreMap map[string]float64,
	bestNodes map[string]int,
	nodes []*api.NodeInfo) (map[string]float64, error) {

	curNPUPlugin := hwNPU.getNPUPlugin(task)
	if curNPUPlugin == nil {
		return nil, errors.New("get npu plugin nil")
	}

	return curNPUPlugin.ScoreBestNPUNodesFn(scoreMap, bestNodes, task, nodes)
}

func (hwNPU *ScheduleHandler) getAllocNPUFromTopology(task *api.TaskInfo, node *api.NodeInfo) (interface{}, error) {
	curNPUPlugin := hwNPU.getNPUPlugin(task)
	if curNPUPlugin == nil {
		return nil, errors.New(noneNPUPlugin)
	}

	return curNPUPlugin.GetAllocatedNPUFromTopologyFn(task, node)
}

func (hwNPU *ScheduleHandler) setNPUTopologyToPod(task *api.TaskInfo, top interface{}) error {
	curNPUPlugin := hwNPU.getNPUPlugin(task)
	if curNPUPlugin == nil {
		return errors.New(noneNPUPlugin)
	}

	return curNPUPlugin.SetNPUTopologyToPodFn(task, top)
}

// For node has  mixed mode, decide which plugin need by task.
func (hwNPU *ScheduleHandler) updateNPUNodeUsedCard(task *api.TaskInfo,
	node *api.NodeInfo, useDeviceIDs interface{}) error {
	curNPUPlugin := hwNPU.getNPUPlugin(task)
	if curNPUPlugin == nil {
		return errors.New(noneNPUPlugin)
	}

	return curNPUPlugin.UpdateNPUNodeUsedCardFn(node, useDeviceIDs)
}

// For node has  mixed mode, decide which plugin need by task.
func (hwNPU *ScheduleHandler) updateReleaseNPUNodeTopology(task *api.TaskInfo,
	node *api.NodeInfo, useDeviceIDs interface{}) error {
	curNPUPlugin := hwNPU.getNPUPlugin(task)
	if curNPUPlugin == nil {
		return errors.New(noneNPUPlugin)
	}

	return curNPUPlugin.UpdateReleaseNPUNodeTopologyFn(node, useDeviceIDs)
}

func (hwNPU *ScheduleHandler) useAnnotation(node *api.NodeInfo, task *api.TaskInfo) {
	// if not npu task no need continue; only check selector before
	if err := hwNPU.isHwNPUTask(task); err != nil {
		klog.V(logDebugLev).Infof("%s useAnnotation %s : %v.", PluginName, task.Name, err)
		return
	}

	useTop, err := hwNPU.getAllocNPUFromTopology(task, node)
	if err != nil {
		klog.V(logErrorLev).Infof("alloc  %s failed:%v.", node.Name, err)
		return
	}

	err = hwNPU.setNPUTopologyToPod(task, useTop)
	if err != nil {
		klog.V(logErrorLev).Infof("%s %v.", node.Name, err)
		return
	}
	// get node available top
	err = hwNPU.updateNPUNodeUsedCard(task, node, useTop)
	if err != nil {
		klog.V(logErrorLev).Infof("%s useAnnotation node(%s) top nil.", PluginName, node.Name)
		return
	}

	return
}

// NPUAllocateFunc Allocate npu and called by volcano frame.
func (hwNPU *ScheduleHandler) NPUAllocateFunc(event *framework.Event, nodeMap map[string]*api.NodeInfo) {
	klog.V(logInfoLev).Infof("enter npu allocate")
	defer klog.V(logInfoLev).Infof("leave npu allocate")

	nodeName := event.Task.NodeName
	node, found := nodeMap[nodeName]
	if !found {
		klog.V(logWarningLev).Infof("%s npuAllocateFunc NOT EXIST node [%s].", PluginName, nodeName)
		return
	}

	hwNPU.useAnnotation(node, event.Task)
	klog.V(logDebugLev).Infof("%s useAnnotation node [%s]'s top.", PluginName, nodeName)
}

func (hwNPU *ScheduleHandler) releaseAnnotation(node *api.NodeInfo, task *api.TaskInfo) {
	// If not npu task, no need to continue;
	if err := hwNPU.isHwNPUTask(task); err != nil {
		klog.V(logDebugLev).Infof("%s releaseAnnotation %s : %v.", PluginName, task.Name, err)
		return
	}

	nowTop, err := hwNPU.getReleaseNPUTopology(task)
	if err != nil {
		klog.V(logErrorLev).Infof("alloc  %s failed:%v.", node.Name, err)
		return
	}

	// Get node available topology.
	err = hwNPU.updateReleaseNPUNodeTopology(task, node, nowTop)
	if err != nil {
		klog.V(logErrorLev).Infof("%s useAnnotation node(%s) top nil.", PluginName, node.Name)
		return
	}

	return
}

// NPUDeallocateFunc Free assigned npu, if allocate failed by volcano frame.
func (hwNPU *ScheduleHandler) NPUDeallocateFunc(event *framework.Event, nodeMap map[string]*api.NodeInfo) {
	klog.V(logInfoLev).Infof("enter npu deallocate")
	defer klog.V(logInfoLev).Infof("leave npu deallocate")

	nodeName := event.Task.NodeName
	node, found := nodeMap[nodeName]
	if !found {
		klog.V(logWarningLev).Infof("%s npuDeallocateFunc from NOT EXIST node [%s].", PluginName, nodeName)
		return
	}

	hwNPU.releaseAnnotation(node, event.Task)
	klog.V(logDebugLev).Infof("%s releaseAnnotation node [%s]'s top.", PluginName, nodeName)
}

func (hwNPU *ScheduleHandler) checkNPUResourceStable(task *api.TaskInfo, node *api.NodeInfo) error {
	curNPUPlugin := hwNPU.getNPUPlugin(task)
	if curNPUPlugin == nil {
		return errors.New(noneNPUPlugin)
	}

	return curNPUPlugin.CheckNPUResourceStableFn(node)
}

// NodePredicate Predicate node by volcano frame.
func (hwNPU *ScheduleHandler) NodePredicate(task *api.TaskInfo, node *api.NodeInfo, conf []conf.Configuration) error {
	klog.V(logInfoLev).Infof("enter node predicate")
	defer klog.V(logInfoLev).Infof("leave NodePredicate")

	if task == nil || node == nil {
		klog.V(logErrorLev).Infof("%s got null parameter(s), which is invalid", PluginName)
		return fmt.Errorf("got null parameter(s)")
	}

	// select node by architect
	if err := hwNPU.preCheckNode(task, node, conf); err != nil {
		// get scheduler selector configure failed, but need continue
		klog.V(logErrorLev).Infof("%s taskName: %s ,nodeName %s : %v.", PluginName, task.Name, node.Name, err)
		return fmt.Errorf("%s in %s:%v", task.Name, node.Name, err)
	}

	// if not npu task no need continue; only check selector before
	if err := hwNPU.isHwNPUTask(task); err != nil {
		klog.V(logDebugLev).Infof("%s isHwNPUTask %s : %v.", PluginName, task.Name, err)
		return nil
	}
	// if not npu node, node should exclude
	if err := hwNPU.isHwNPUNode(task, node); err != nil {
		klog.V(logDebugLev).Infof("%s %s : %v.", PluginName, node.Name, err)
		return fmt.Errorf("isNPUNode %s :%s", nodesNoMeetNPUReqError, err)
	}
	// check resource stabilize
	if err := hwNPU.checkNPUResourceStable(task, node); err != nil {
		// npu on node are not stable, node cannot be selected.
		klog.V(logInfoLev).Infof("%s checkNPUResourceStable %s : %v ,cannot be selected.", PluginName,
			node.Name, err)
		return fmt.Errorf("checkNPUResourceStable %s : %v", node.Name, err)
	}

	if err := hwNPU.checkNodeNPUByTask(task, node); err != nil {
		// node doesn't have enough npu for the task
		klog.V(logInfoLev).Infof("%s checkNodeNPUByTask %s:%v ,cannot be selected.", PluginName, node.Name, err)
		return fmt.Errorf("checkNodeNPUByTask %s : %v", node.Name, err)
	}

	return nil
}

// PreHandleFaultNPU handle NPU fault chip.(Find fault npu,node,pod,RankIndex)
func (hwNPU *ScheduleHandler) PreHandleFaultNPU(ssn *framework.Session) error {
	if err := util.ReadFaultNPUJobsFromCM(ssn); err != nil {
		klog.V(logErrorLev).Infof("readFaultNPUJobsFromCM :%v.", err)
	}

	return hwNPU.preHandleFaultNPUFn(ssn)
}
