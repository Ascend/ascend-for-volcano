/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package plugin is using for HuaWei Ascend pin affinity schedule frame.

*/
package plugin

import (
	"errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog"
	schedulerApi "k8s.io/kube-scheduler/extender/v1"
	"reflect"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"
)

func updatePodPendingReason(task *api.TaskInfo, reasonTmp string) {
	condition := v1.PodCondition{
		Type:    v1.PodScheduled,
		Status:  v1.ConditionFalse,
		Reason:  v1.PodReasonUnschedulable,
		Message: reasonTmp,
	}
	for _, tmp := range task.Pod.Status.Conditions {
		if reflect.DeepEqual(tmp, condition) {
			return
		}
	}
	task.Pod.Status.Conditions = append(task.Pod.Status.Conditions, condition)
}

func initScoreMap(nodes []*api.NodeInfo, interPodAffinityScore schedulerApi.HostPriorityList) map[string]float64 {
	for _, node := range nodes {
		if reflect.ValueOf(node).IsNil() {
			continue
		}

		interPodAffinityScore = append(interPodAffinityScore, schedulerApi.HostPriority{
			Host:  node.Name,
			Score: 0,
		})
	}
	scoreMap := make(map[string]float64, len(interPodAffinityScore))
	for _, host := range interPodAffinityScore {
		scoreMap[host.Host] = float64(host.Score)
	}
	return scoreMap
}

func (hwNPU *ScheduleHandler) getReleaseNPUTopology(task *api.TaskInfo) (interface{}, error) {
	nowNPUPlugin := hwNPU.getNPUPlugin(task)
	if nowNPUPlugin == nil {
		return nil, errors.New("get nil NPUPlugin")
	}

	return nowNPUPlugin.GetReleaseNPUTopologyFn(task)
}

// BatchNodeOrderFn Score nodes, which used by volcano frame.
func (hwNPU *ScheduleHandler) BatchNodeOrderFn(
	task *api.TaskInfo,
	nodes []*api.NodeInfo,
	disFlag bool) (map[string]float64, error) {
	var interPodAffinityScore schedulerApi.HostPriorityList

	klog.V(logInfoLev).Infof("Enter batchNodeOrderFn")
	defer klog.V(logInfoLev).Infof("leaving batchNodeOrderFn")

	if task == nil || nodes == nil {
		klog.V(logErrorLev).Infof("%s batchNodeOrderFn got null parameter(s), which is invalid.", PluginName)
		return nil, errors.New("got null parameter(s)")
	}

	// init score-map
	scoreMap := initScoreMap(nodes, interPodAffinityScore)

	// 1.If not npu task no need continue.
	if err := hwNPU.isHwNPUTask(task); err != nil {
		klog.V(logDebugLev).Infof("%s %s : %v.", PluginName, task.Name, err)
		return scoreMap, nil
	}

	// 2.Get the best node and top by A,B,C,D rules and require numbers.
	bestNodes, errGet := hwNPU.getNPUAffinityBestNodes(task, nodes, disFlag)
	if errGet != nil || len(bestNodes) == 0 {
		// get suitable node failed
		klog.V(logErrorLev).Infof("%s batchNodeOrderFn task[%s] failed[%v].",
			PluginName, task.Name, errGet)
		return scoreMap, nil
	}
	klog.V(logInfoLev).Infof("%s batchNodeOrderFn Get %s for NPU %+v.",
		PluginName, task.Name, bestNodes)

	// 3.Scored the nodes and set topology.
	scoreMap, errGet = hwNPU.scoreBestNPUNodes(task, scoreMap, bestNodes, nodes)
	if errGet != nil {
		// get suitable node failed
		klog.V(logErrorLev).Infof("%s scoreBestNPUNodes get err:%v.", PluginName, errGet)
		return scoreMap, errGet
	}

	klog.V(logInfoLev).Infof("%s Total Score for task %s/%s is: %v.", PluginName,
		task.Namespace, task.Name, scoreMap)

	return scoreMap, nil
}

// IsDistributeTask To judge whether the distributed task.
func IsDistributeTask(task *api.TaskInfo, ssn *framework.Session) bool {
	job, ok := ssn.Jobs[task.Job]
	if !ok {
		klog.V(logErrorLev).Infof("IsDistributeTask get %s not in ssn.", task.Job)
		return false
	}

	if len(job.Tasks) > 1 {
		klog.V(logDebugLev).Infof("IsDistributeTask %s get %d tasks.", task.Job, len(job.Tasks))
		return true
	}

	klog.V(logDebugLev).Infof("IsDistributeTask %s get %d task.", task.Job, len(job.Tasks))
	return false
}
