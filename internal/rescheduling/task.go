/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package rescheduling is using for HuaWei Ascend pin fault rescheduling.

*/
package rescheduling

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

func (fTask *FaultTask) getNodeRankIndex(task *api.TaskInfo) (string, error) {
	rankIndex, ok := task.Pod.Annotations[podRankIndex]
	if !ok {
		return "", errors.New("nil rankIndex")
	}

	index, err := strconv.Atoi(rankIndex)
	if err != nil {
		return "", fmt.Errorf("convert %#v:%#v", rankIndex, err)
	}

	if index > maxRankIndex || index < 0 {
		return "", fmt.Errorf("rankIndex:%#v out of limit", index)
	}
	klog.V(util.LogInfoLev).Infof("task %s rankIndex read from pod: %s", task.Name, rankIndex)
	return rankIndex, nil
}

func (fTask *FaultTask) getUseCardName(task *api.TaskInfo, cardName string, cardMaxNum int) ([]string, error) {
	strNpu, ok := task.Pod.Annotations[cardName]
	if !ok {
		return nil, fmt.Errorf("%s has no NPU from %s", task.Name, cardName)
	}
	taskNPUs := strings.Split(strNpu, ",")
	if len(taskNPUs) > cardMaxNum {
		err := fmt.Errorf("get err %s npus %#v", task.Name, taskNPUs)
		return nil, err
	}
	return taskNPUs, nil
}

func (fTask *FaultTask) deleteRealPodByTask(ssn *framework.Session, waitTime int64) error {
	deleteOptions := v1.DeleteOptions{
		GracePeriodSeconds: &waitTime,
		Preconditions:      v1.NewUIDPreconditions(string(fTask.TaskUID)),
	}

	err := ssn.KubeClient().CoreV1().Pods(fTask.TaskNamespace).Delete(
		context.TODO(), fTask.TaskName, deleteOptions)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("Failed to delete %s: %#v", fTask.TaskUID, err)
		return err
	}

	klog.V(util.LogInfoLev).Infof("%s==%v force terminated and removed from etcd",
		fTask.TaskName, fTask.TaskUID)
	return nil
}

func (fTask *FaultTask) getTaskUseFaultCardHealthState(fNode *FaultNode) []string {
	var nodeUseCardHealthState []string
	for _, taskUseCard := range fTask.UseCardName {
		if util.IsSliceContain(taskUseCard, fNode.UnhealthyNPU) {
			nodeUseCardHealthState = append(nodeUseCardHealthState, NodeCardUnhealthy)
			continue
		}
		if util.IsSliceContain(taskUseCard, fNode.NetworkUnhealthyNPU) {
			nodeUseCardHealthState = append(nodeUseCardHealthState, NodeCardNetworkUnhealthy)
		}
	}
	return nodeUseCardHealthState
}

func (fTask *FaultTask) setUseCardName(value []string) {
	fTask.UseCardName = value
}

func (fTask *FaultTask) setIsFaultTask(value bool) {
	fTask.IsFaultTask = value
}

func (fTask *FaultTask) setFaultType(value string) {
	fTask.faultType = value
}

func (fTask *FaultTask) setNodeRankIndex(value string) {
	fTask.NodeRankIndex = value
}

func newFaultTaskDefault(task *api.TaskInfo, job *api.JobInfo) FaultTask {
	faultTask := FaultTask{
		IsFaultTask:   false,
		TaskName:      task.Name,
		TaskUID:       task.UID,
		TaskNamespace: task.Namespace,
		NodeName:      task.NodeName,
		JobName:       job.Name,
		NodeRankIndex: "",
		UseCardName:   nil,
		PodCreateTime: task.Pod.CreationTimestamp.Unix(),
		PodUID:        task.Pod.UID,
		faultType:     NodeHealthy,
	}
	return faultTask
}
