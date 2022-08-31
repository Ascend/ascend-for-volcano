/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package plugin is using for HuaWei Ascend pin affinity schedule frame.

*/
package plugin

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/conf"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/rescheduling"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/util"
)

func (hwNPU *ScheduleHandler) initNodesNPUAllocTopology(ssn *framework.Session) error {
	if initErr := util.InitPluginNodeCache(ssn); initErr != nil {
		klog.V(logErrorLev).Infof("InitNodesNPUAllocTopology :%v.", initErr)
		return initErr
	}

	for cardName, initNodes := range hwNPU.InitNodesNPUAllocTopologyFns {
		if err := initNodes(ssn.Nodes); err != nil {
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

func (hwNPU *ScheduleHandler) isHwNPUNode(task *api.TaskInfo, node *api.NodeInfo) error {
	curNPUPlugin := hwNPU.getNPUPlugin(task)
	if curNPUPlugin == nil {
		return nil
	}
	if !hwNPU.IsPluginRegistered(curNPUPlugin.Name()) {
		plugErr := fmt.Errorf("%s not registered", curNPUPlugin.Name())
		klog.V(logErrorLev).Infof("isHwNPUNode :%v.", plugErr)
		return plugErr
	}
	return curNPUPlugin.IsMyNode(node)
}

func (hwNPU *ScheduleHandler) checkNodeNPUByTask(task *api.TaskInfo, node *api.NodeInfo, distributeFlag bool) error {
	curNPUPlugin := hwNPU.getNPUPlugin(task)
	if curNPUPlugin == nil {
		return nil
	}

	return curNPUPlugin.CheckNodeNPUByTaskFn(task, node, distributeFlag)
}

func (hwNPU *ScheduleHandler) getNPUAffinityBestNodes(
	task *api.TaskInfo, nodes []*api.NodeInfo, disFlag bool) (map[string]int, error) {
	curNPUPlugin := hwNPU.getNPUPlugin(task)
	if curNPUPlugin == nil {
		return nil, errors.New("get npu plugin nil")
	}

	return curNPUPlugin.GetNPUAffinityBestNodesFn(task, nodes, disFlag)
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

func (hwNPU *ScheduleHandler) getAllocNPUsFromNode(
	task *api.TaskInfo, node *api.NodeInfo, disFlag bool) (interface{}, error) {
	curNPUPlugin := hwNPU.getNPUPlugin(task)
	if curNPUPlugin == nil {
		return nil, errors.New(noneNPUPlugin)
	}

	return curNPUPlugin.GetAllocatedNPUFromTopologyFn(task, node, disFlag)
}

func (hwNPU *ScheduleHandler) setNPUTopologyToPod(task *api.TaskInfo, top interface{}) error {
	curNPUPlugin := hwNPU.getNPUPlugin(task)
	if curNPUPlugin == nil {
		return errors.New(noneNPUPlugin)
	}
	// sleep for pod not be same time create.
	time.Sleep(time.Millisecond)
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

func (hwNPU *ScheduleHandler) useAnnotation(node *api.NodeInfo, task *api.TaskInfo, distributeFlag bool) {
	// if not npu task no need continue; only check selector before
	if err := hwNPU.isHwNPUTask(task); err != nil {
		klog.V(logDebugLev).Infof("%s useAnnotation %s : %v.", PluginName, task.Name, err)
		return
	}

	useTop, err := hwNPU.getAllocNPUsFromNode(task, node, distributeFlag)
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
	// set pod rankIndex
	if err = rescheduling.SetFaultJobPodIndex(task, node); err != nil {
		klog.V(logInfoLev).Infof("%s useAnnotation setFaultJobPodIndex %v.", task.Name, err)
	}
	return
}

// NPUAllocateFunc Allocate npu and called by volcano frame.
func (hwNPU *ScheduleHandler) NPUAllocateFunc(event *framework.Event, ssn *framework.Session) {
	klog.V(logInfoLev).Infof("enter npu allocate")
	defer klog.V(logInfoLev).Infof("leave npu allocate")

	nodeName := event.Task.NodeName
	node, found := ssn.Nodes[nodeName]
	if !found {
		klog.V(logWarningLev).Infof("%s npuAllocateFunc NOT EXIST node [%s].", PluginName, nodeName)
		return
	}

	hwNPU.useAnnotation(node, event.Task, IsDistributeTask(event.Task, ssn))
	klog.V(logDebugLev).Infof("%s %v useAnnotation node [%s]'s top.", PluginName, event.Task.Name, nodeName)
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
	if !hwNPU.IsPluginRegistered(curNPUPlugin.Name()) {
		plugErr := fmt.Errorf("%s not registered", curNPUPlugin.Name())
		klog.V(logErrorLev).Infof("checkNPUResourceStable :%v.", plugErr)
		return plugErr
	}
	return curNPUPlugin.CheckNPUResourceStableFn(node)
}

// ClusterNodePredicate Predicate node by volcano frame.
func (hwNPU *ScheduleHandler) ClusterNodePredicate(task *api.TaskInfo, ssn *framework.Session) error {
	for pluginName, clusterNodePredicate := range hwNPU.ClusterNodePredicateFns {
		if err := clusterNodePredicate(task, ssn); err != nil {
			klog.V(logErrorLev).Infof("%s clusterNodePredicate :%v.", pluginName, err)
			return err
		}
	}

	return nil
}

func handlingPreCheckNodeErr(preErr error) error {
	if strings.Contains(preErr.Error(), util.SegmentNoEnable) {
		klog.V(logInfoLev).Infof("%s preCheckNode %v.", PluginName, preErr)
		return nil
	}
	klog.V(logErrorLev).Infof("%s preCheckNode %v.", PluginName, preErr)
	return preErr
}

// NodePredicate Predicate node by volcano frame.
func (hwNPU *ScheduleHandler) NodePredicate(task *api.TaskInfo, node *api.NodeInfo, ssn *framework.Session) error {
	klog.V(logInfoLev).Infof("enter node(%s) predicate", node.Name)
	defer klog.V(logInfoLev).Infof("leave node(%s) predicate", node.Name)

	if task == nil || node == nil || ssn == nil {
		klog.V(logErrorLev).Infof("%s got null parameter(s), which is invalid", PluginName)
		return fmt.Errorf("got null parameter(s)")
	}

	// select node by architect
	if err := hwNPU.preCheckNode(task, node, ssn.Configurations); err != nil {
		// get scheduler selector configure failed, but need continue
		preErr := fmt.Errorf("%s in %s:%v", task.Name, node.Name, err)
		return handlingPreCheckNodeErr(preErr)
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
		klog.V(logInfoLev).Infof("%s checkNPUResourceStable %v ,cannot be selected.", PluginName, err)
		return fmt.Errorf("checkNPUResourceStable %v", err)
	}

	if err := hwNPU.checkNodeNPUByTask(task, node, IsDistributeTask(task, ssn)); err != nil {
		// node doesn't have enough npu for the task
		klog.V(logInfoLev).Infof("%s checkNodeNPUByTask %s:%v ,cannot be selected.", PluginName, node.Name, err)
		return fmt.Errorf("checkNodeNPUByTask  %s : %s %v", node.Name, nodesNoMeetNPUReqError, err)
	}
	klog.V(logInfoLev).Infof("%s NodePredicate %s select successes.", PluginName, node.Name)
	return nil
}

// initHandleFaultNPUInf handle NPU fault chip.(Find fault npu,node,pod,RankIndex)
func (hwNPU *ScheduleHandler) initHandleFaultNPUInf(ssn *framework.Session) error {
	if err := rescheduling.ReadFaultNPUJobsFromCM(ssn); err != nil {
		klog.V(logErrorLev).Infof("readFaultNPUJobsFromCM :%v.", err)
	}

	npuErr := hwNPU.preHandleFaultNPUFn(ssn)
	if npuErr != nil {
		klog.V(logErrorLev).Infof("initHandleFaultNPUInf :%v.", npuErr)
	}

	return nil
}

// preHandleVNPUFn handle NPU fault chip.(Find fault npu,node,pod,RankIndex)
func (hwNPU *ScheduleHandler) preHandleVNPUFn(ssn *framework.Session) error {
	for pluginName, preHandleVNPUFn := range hwNPU.PreHandleVNPUFns {
		if err := preHandleVNPUFn(ssn); err != nil {
			if err.Error() != util.SegmentNoEnable {
				klog.V(logDebugLev).Infof("%s preHandleFaultNPU :%v.", pluginName, err)
				return err
			}
			klog.V(logErrorLev).Infof("%s preHandleFaultNPU :%v.", pluginName, err)
			if unRegErr := hwNPU.UnRegisterNPUScheduler(pluginName); unRegErr != nil {
				klog.V(logErrorLev).Infof("preHandleVNPUFn :%v.", unRegErr)
				return unRegErr
			}
			return err
		}
	}
	return nil
}

// vJobRunHandleFn handle NPU fault chip.(Find fault npu,node,pod,RankIndex)
func (hwNPU *ScheduleHandler) vJobRunHandleFn(ssn *framework.Session) error {
	for pluginName, vJobRunHandle := range hwNPU.VJobRunHandleFns {
		if err := vJobRunHandle(ssn); err != nil {
			klog.V(logDebugLev).Infof("%s vJobRunHandleFn :%v.", pluginName, err)
			return err
		}
	}
	return nil
}
