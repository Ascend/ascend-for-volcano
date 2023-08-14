/*
Copyright(C)2023. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package module910bx8 is using for HuaWei Ascend910Bx8 pin affinity schedule.
*/
package module910bx8

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"k8s.io/api/core/v1"
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/base"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/rescheduling"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// New return npu plugin
func New(name string) base.AscendHandler {
	m := &module910bx8{}
	m.SetPluginName(name)
	m.SetAnnoName(util.NPU910CardName)
	m.SetAnnoPreVal(util.NPU910CardNamePre)
	m.SetDefaultJobSchedulerConfig(nil)
	m.SetMaxNodeNPUNum(nodeNPUNumber)
	m.SetAcceleratorValue(util.JobKind910BValue)
	m.SetArch(util.HuaweiArchX86 + util.HuaweiArchArm)
	m.netUnhealthyKey = networkUnhealthyNPU
	m.AffScoreList = [][]int{
		{util.AffScore0, util.AffScore1, util.AffScore2, util.AffScore3, util.AffScore4, util.AffScore5,
			util.AffScore6, util.AffScore7},
		{util.AffScore8, util.AffScore0, util.AffScore1, util.AffScore2, util.AffScore3, util.AffScore4,
			util.AffScore5, util.AffScore6},
		{util.AffScore8, util.AffScore8, util.AffScore0, util.AffScore1, util.AffScore2, util.AffScore3,
			util.AffScore4, util.AffScore5},
		{util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore0, util.AffScore1, util.AffScore2,
			util.AffScore3, util.AffScore4},
		{util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore0, util.AffScore1,
			util.AffScore2, util.AffScore3},
		{util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore0,
			util.AffScore1, util.AffScore2},
		{util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8,
			util.AffScore0, util.AffScore1},
		{util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8,
			util.AffScore8, util.AffScore0},
	}
	return m
}

// ValidNPUJob check job req npu num and mode
func (tp *module910bx8) ValidNPUJob() *api.ValidateResult {
	result := tp.Valid910bNPUJob()
	if result != nil {
		return result
	}
	k, ok := tp.Label[plugin.TorAffinityKey]
	if !ok {
		klog.V(util.LogInfoLev).Infof("validNPUJob job is 910x8 module")
		return nil
	}
	if k == plugin.LargeModelTag {
		if err := tp.ValidRestartNPUJob(); err != nil {
			return &api.ValidateResult{Pass: false, Reason: err.Error(), Message: err.Error()}
		}
	}
	return nil
}

func (tp *module910bx8) ValidRestartNPUJob() error {
	if tp.reHandle == nil {
		return nil
	}
	reCache := tp.reHandle.DealReSchedulerCache
	fNodes := reCache.GetRealFaultNodes()
	if !reCache.IsJobInRealFaultJobs(tp.Name) {
		return nil
	}
	err := fmt.Errorf("vaild job err")
	cm, getErr := util.GetConfigMapWithRetry(tp.FrameAttr.KubeClient, tp.NameSpace, plugin.ServerListCMPre+tp.ReferenceName)
	if getErr != nil {
		return nil
	}
	vcjob, ok := tp.Jobs[tp.Name]
	if !ok {
		return err
	}
	cmdata, ok := cm.Data[plugin.ServerListCMKey]
	if !ok {
		return nil
	}
	serverListInfo := plugin.TorListInfo{}
	marErr := json.Unmarshal([]byte(cmdata), &serverListInfo)
	if marErr != nil {
		return marErr
	}
	if serverListInfo.Status != plugin.Completed {
		return nil
	}

	serverList := serverListInfo.ServerList
	npuName := tp.GetAnnoName()
	for _, tor := range serverList {
		for _, server := range tor.Servers {
			ip, ok := server[plugin.ServerIPKey]
			if !ok {
				continue
			}
			for nodeName, node := range tp.Nodes {
				if ip != node.Address || rescheduling.IsNodeInFaultNode(fNodes, nodeName) {
					continue
				}
				nodeA, aOK := node.Annotation[npuName]
				if !aOK {
					return err
				}
				sSlice := strings.Split(nodeA, ",")
				length := len(sSlice)
				if length == 1 && sSlice[0] == "" {
					length = 0
				}
				checkErr := node.CheckNPUResourceStable(vcjob)

				if checkErr != nil || length < tp.ReqNPUNum/vcjob.GetNPUTaskNumInJob() {
					return fmt.Errorf("check Node %s err %#v and deviceInfo is %v", nodeName, checkErr, length)
				}
			}
		}
	}
	return nil
}

// PreStartAction pre-processing actions for rescheduling
func (tp *module910bx8) PreStartAction(ssn *framework.Session) error {
	moduleFullName := util.NPU910CardName + util.Module910bx8AcceleratorType
	klog.V(util.LogInfoLev).Infof("Entering PreStartAction of %s...", moduleFullName)
	defer klog.V(util.LogInfoLev).Infof("Leaving PreStartAction of %s", moduleFullName)
	if tp == nil || ssn == nil || tp.FrameAttr.KubeClient == nil {
		return fmt.Errorf("%s handler not enabled or ssn is nil: %s", moduleFullName, util.ArgumentError)
	}
	tp.reHandle = rescheduling.New(&tp.ScheduleEnv, rescheduling.CmFaultJob910bx8Kind)
	if tp.reHandle == nil {
		klog.V(util.LogErrorLev).Infof("create new fault handler failed.")
		return fmt.Errorf("%s reSchedule not enabled: %s", moduleFullName, util.ArgumentError)
	}
	tp.reHandle.NewCommonReScheduler(rescheduling.CmFaultJob910bx8Kind)
	tp.reHandle.SynCacheFaultNodeWithSession(util.NPU910CardName)
	tp.reHandle.AddFaultNodeWithSession(util.NPU910CardName)
	tp.reHandle.SynCacheFaultJobWithSession(ssn, util.NPU910CardName, util.NPU910CardNamePre)
	tp.reHandle.SynCacheNodeRankOccMapWithSession(ssn)
	// 1. restart Fault Jobs that are recorded in cache
	if restartErr := tp.reHandle.RestartNeedForceDeleteJobs(ssn); restartErr != nil {
		klog.V(util.LogErrorLev).Infof("%s RestartNeedForceDeleteJobs: %s", moduleFullName, restartErr.Error())
	}
	// 2. get all the new 910bx8 jobs in session
	runningJobs910bx8, getRunErr := tp.reHandle.GetRunningJobs(ssn, util.Ascend910bName, util.Module910bx8AcceleratorType)
	if getRunErr != nil {
		klog.V(util.LogInfoLev).Infof("%s GetRunningJobs: %s", moduleFullName, getRunErr.Error())
	}
	// 3. get nodes of session and fault jobs of 910bx8
	err := tp.reHandle.AddFaultJobWithSession(runningJobs910bx8, util.NPU910CardName, util.NPU910CardNamePre)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s AddFaultJobWithSession", moduleFullName)
	}
	// 4. restart the fault jobs
	if restartErr := tp.reHandle.RestartFaultJobs(ssn); restartErr != nil {
		klog.V(util.LogErrorLev).Infof("%s RestartFaultJobs: %s", moduleFullName, restartErr.Error())
		return restartErr
	}
	// 5. save structure for later allocation process
	tp.reHandle.GenerateNodeRankIndexTaskMap()
	return nil
}

// PreStopAction post-processing actions for re-scheduling
func (tp *module910bx8) PreStopAction(env *plugin.ScheduleEnv) error {
	moduleFullName := util.NPU910CardName + util.Module910bx8AcceleratorType
	klog.V(util.LogInfoLev).Infof("enter PreStopAction %s...", moduleFullName)
	defer klog.V(util.LogInfoLev).Infof("leave PreStopAction %s...", moduleFullName)
	if tp == nil || tp.reHandle == nil || env == nil || tp.FrameAttr.KubeClient == nil {
		return fmt.Errorf("%s reSchedule not enabled or nil env: %s", moduleFullName, util.ArgumentError)
	}
	if err := tp.reHandle.WriteReSchedulerCacheToEnvCache(env, rescheduling.CmFaultJob910bx8Kind); err != nil {
		return err
	}
	return nil
}

// CheckNodeNPUByTask check nod npu meet task req
func (tp *module910bx8) CheckNodeNPUByTask(task *api.TaskInfo, node plugin.NPUNode) error {
	if tp == nil || task == nil || len(node.Annotation) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("CheckNodeNPUByTask err: %s", err)
		return err
	}
	taskNPUNum, err := tp.GetTaskReqNPUNum(task)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s GetTaskReqNPUNum err: %s", tp.GetPluginName(), err.Error())
		return err
	}
	nTaskNum := tp.GetNPUTaskNumInJob()
	nodeTop, err := tp.getUsableTopFromNode(node, nTaskNum > 1)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s getUsableTopFromNode err: %s", tp.GetPluginName(), err.Error())
		return err
	}

	if err = tp.JudgeNodeAndTaskNPU(taskNPUNum, nodeTop); err != nil {
		klog.V(util.LogErrorLev).Infof("%s JudgeNodeAndTaskNPU err: %s", tp.GetPluginName(), err.Error())
		return fmt.Errorf("checkNodeNPUByTask %s err: %s", util.NodeNotMeetTopologyWarning, err.Error())
	}

	if tp.reHandle != nil {
		if reErr := tp.reHandle.CheckNodeNPUByTask(task, node); reErr != nil {
			return fmt.Errorf("rescheduling CheckNodeNPUByTask %s", reErr.Error())
		}
	}
	return nil
}

func (tp *module910bx8) ScoreBestNPUNodes(task *api.TaskInfo, nodes []*api.NodeInfo, sMap map[string]float64) error {
	if tp == nil || task == nil || len(nodes) == 0 || len(sMap) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("ScoreBestNPUNodes %s.", err)
		return err
	}
	taskNPUNum, getErr := tp.GetTaskReqNPUNum(task)
	if getErr != nil {
		klog.V(util.LogErrorLev).Infof("%s GetTaskReqNPUNum %s: %s", tp.GetPluginName(), task.Name, getErr)
		return getErr
	}
	for _, node := range nodes {
		if reflect.ValueOf(node).IsNil() {
			continue
		}
		nNode, ok := tp.Nodes[node.Name]
		if !ok {
			klog.V(util.LogWarningLev).Infof("%s %s ScoreBestNPUNodes %s is not npu node",
				tp.GetPluginName(), task.Name, node.Name)
			continue
		}
		nTaskNum := tp.GetNPUTaskNumInJob()
		cardIds, err := tp.getUsableTopFromNode(nNode, nTaskNum > 1)
		if err != nil {
			klog.V(util.LogWarningLev).Infof("%s ScoreBestNPUNodes getErr: %s", tp.GetPluginName(), err)
			continue
		}
		bestScore, err := tp.getNodeBestScore(taskNPUNum, cardIds)
		if err != nil {
			klog.V(util.LogWarningLev).Infof("%s ScoreBestNPUNodes getErr: %s", tp.GetPluginName(), err)
			continue
		}
		healthyNPUNum, ok := nNode.Allocate[v1.ResourceName(tp.GetAnnoName())]
		if !ok {
			klog.V(util.LogWarningLev).Infof("%s ScoreBestNPUNodes node<%s> get allocate npu failed",
				tp.GetPluginName(), node.Name)
			continue
		}
		sMap[node.Name] = float64(tp.MaxNodeNPUNum * (int(healthyNPUNum/util.NPUHexKilo) - bestScore))
	}
	reErr := tp.reHandle.ScoreBestNPUNodes(task, sMap)
	if reErr != nil {
		klog.V(util.LogErrorLev).Infof(
			"%s rescheduling ScoreBestNPUNodes failed :%s.", SchedulerName, reErr.Error())
	}
	klog.V(util.LogInfoLev).Infof("%s ScoreBestNPUNodes task<%s> sMap<%v>", tp.GetPluginName(),
		task.Name, sMap)
	return nil
}

// UseAnnotation select npu for task from node
func (tp *module910bx8) UseAnnotation(task *api.TaskInfo, node plugin.NPUNode) *plugin.NPUNode {
	if tp == nil || task == nil || len(node.Annotation) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("UseAnnotation %s.", err)
		return nil
	}
	klog.V(util.LogDebugLev).Infof("%s UseAnnotation task<%s> node<%s> resource<%s> Annotation: %#v",
		tp.GetPluginName(), task.Name, node.Name, tp.GetAnnoName(), node.Annotation)
	selectedNPU, err := tp.selectNPUFromNode(task, node)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s UseAnnotation err:%s.", tp.GetPluginName(), err)
		return nil
	}
	klog.V(util.LogInfoLev).Infof("%s UseAnnotation %s select %v.", tp.GetPluginName(), task.Name, selectedNPU)

	tp.SetNPUTopologyToPodFn(task, selectedNPU)
	newNode := tp.UpdateNodeInfo(node, selectedNPU)
	if tp.reHandle != nil {
		if reErr := tp.reHandle.UseAnnotation(task, newNode); reErr != nil {
			klog.V(util.LogErrorLev).Infof("%s rescheduling UseAnnotation: %s", SchedulerName, reErr.Error())
			return nil
		}
	}
	return newNode
}

func (tp *module910bx8) selectNPUFromNode(task *api.TaskInfo, node plugin.NPUNode) ([]int, error) {
	taskNPUNum, err := tp.GetTaskReqNPUNum(task)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s GetTaskReqNPUNum err: %s", tp.GetPluginName(), err.Error())
		return nil, err
	}
	nTaskNum := tp.GetNPUTaskNumInJob()
	nodeTop, err := tp.getUsableTopFromNode(node, nTaskNum > 1)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s getUsableTopFromNode err: %s", tp.GetPluginName(), err.Error())
		return nil, err
	}
	if len(nodeTop) < taskNPUNum {
		err = fmt.Errorf("node<%s> top<%v> can not meet task req<%d>", node.Name, len(nodeTop), taskNPUNum)
		klog.V(util.LogErrorLev).Infof("ScoreBestNPUNodes err: %s", err)
		return nil, err
	}
	return nodeTop[:taskNPUNum], nil
}

// ReleaseAnnotation Release used resource.
func (tp *module910bx8) ReleaseAnnotation(_ *api.TaskInfo, node plugin.NPUNode) *plugin.NPUNode {
	return &node
}
