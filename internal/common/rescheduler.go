/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

// Package common is using for HuaWei common infer Ascend pin affinity schedule.
package common

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/klog"
	"volcano.sh/apis/pkg/apis/scheduling"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/rescheduling"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/util"
)

// PreHandleFaultNPUFn PreHandleFaultNPUFn
func (re *ReScheduler) PreHandleFaultNPUFn(ssn *framework.Session) error {
	klog.V(util.LogDebugLev).Infof("%s enter preHandleFaultNPUFn.", re.AnnoUnHealthy)
	defer klog.V(util.LogDebugLev).Infof("%s leave preHandleFaultNPUFn.", re.AnnoUnHealthy)

	// 1.get fault npu and node.
	fNode, err := re.getInoperableNPUCards(ssn.Nodes)
	if err != nil {
		klog.V(util.LogDebugLev).Infof("%s preHandleFaultNPUFn %v.", re.AnnoUnHealthy, err)
	}
	// 2.Determine if it is a  jobs.
	tasks, jobGetErr := re.getFaultTasks(ssn.Jobs, fNode)
	if jobGetErr != nil {
		klog.V(util.LogDebugLev).Infof("PreHandleFaultNPUFn %s : %v.", re.AnnoUnHealthy, jobGetErr)
		return nil
	}

	for _, task := range tasks {
		if task == nil {
			continue
		}
		klog.V(util.LogDebugLev).Infof("evict task %+v", task)
		if err := ssn.Evict(task, jobRestartReason); err != nil {
			klog.V(util.LogDebugLev).Infof("%s Failed to restart %s : %v", re.AnnoUnHealthy, task.UID, err)
			re.updatePodReason(task, err.Error())
			return err
		}
	}
	return nil
}
func (re *ReScheduler) updatePodReason(task *api.TaskInfo, reasonTmp string) {
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

func (re *ReScheduler) getNodeFaultNPUs(node *api.NodeInfo) ([]string, error) {
	npuStrings, ok := node.Node.Annotations[re.AnnoUnHealthy]
	if !ok || len(npuStrings) == 0 {
		return nil, fmt.Errorf("%s get nil %s ", node.Name, re.AnnoUnHealthy)
	}

	faultNPUs := strings.Split(npuStrings, ",")
	return faultNPUs, nil
}

func (re *ReScheduler) getInoperableNPUCards(nodes map[string]*api.NodeInfo) ([]rescheduling.FaultNPUsOnNode, error) {
	var faultNPUs []rescheduling.FaultNPUsOnNode

	for _, nodeInfo := range nodes {
		npus, err := re.getNodeFaultNPUs(nodeInfo)
		if err != nil {
			klog.V(util.LogDebugLev).Infof("getNodeFaultNPUs err:%v.", err)
			continue
		}

		faultNPUs = append(faultNPUs, rescheduling.FaultNPUsOnNode{
			NodeName: nodeInfo.Name, FaultNPUs: npus, UpdateTime: time.Now().Unix()})
	}

	if len(faultNPUs) == 0 {
		return nil, fmt.Errorf("%v nil inoperable NPU", reflect.ValueOf(nodes).MapKeys())
	}
	klog.V(util.LogDebugLev).Infof("getInoperableNPUCards %+v.", faultNPUs)

	return faultNPUs, nil
}

func (re *ReScheduler) getFaultTasks(jobs map[api.JobID]*api.JobInfo,
	nodes []rescheduling.FaultNPUsOnNode) ([]*api.TaskInfo, error) {
	var myTasks []*api.TaskInfo

	for _, job := range jobs {
		if err := re.IsMyJob(job); err != nil {
			continue
		}
		if job.PodGroup.Status.Phase != scheduling.PodGroupRunning {
			continue
		}
		klog.V(util.LogDebugLev).Infof("%s Myjobs  PodGroupRunning ", job.Name)
		for _, task := range job.Tasks {
			if !re.isFaultSchedulingTask(task) {
				klog.V(util.LogDebugLev).Infof("isFaultSchedulingTask false taskname info:%s.", task.Name)
				continue
			}
			if re.isFaultTask(task, nodes) {
				myTasks = append(myTasks, task)
			}
		}
	}

	if len(myTasks) == 0 {
		klog.V(util.LogDebugLev).Infof("len myTasks is 0.")
		return nil, fmt.Errorf("nil %s jobs", re.AnnoUnHealthy)
	}

	klog.V(util.LogDebugLev).Infof("job getFaultTasks myTasks len %d", len(myTasks))
	return myTasks, nil
}

func (re *ReScheduler) isFaultTask(task *api.TaskInfo, nodes []rescheduling.FaultNPUsOnNode) bool {
	klog.V(util.LogDebugLev).Infof("isFaultTask task :%s.   nodes %+v", task.NodeName, nodes)
	for _, node := range nodes {
		if node.NodeName == task.NodeName {
			anno, exist := task.Pod.Annotations[re.AnnoName]
			if !exist {
				klog.V(util.LogDebugLev).Infof("get %s Annotations %s failed.", task.Pod, re.AnnoName)
				continue
			}
			klog.V(util.LogDebugLev).Infof("isFaultTask  nodes %+v", node.FaultNPUs)
			if re.checkFaultNPU(anno, node.FaultNPUs) {
				klog.V(util.LogDebugLev).Infof("checkFaultNPU pod annotation.")
				return true
			}
		}
	}
	return false
}

func (re *ReScheduler) checkFaultNPU(anno string, faultNPUs []string) bool {
	used := strings.Split(anno, ",")
	for _, us := range used {
		for _, ft := range faultNPUs {
			if us == ft {
				klog.V(util.LogDebugLev).Infof("checkFaultNPU true, nodes %+v", faultNPUs)
				return true
			}
		}
	}
	klog.V(util.LogDebugLev).Infof("checkFaultNPU false, nodes %+v", faultNPUs)
	return false
}

func (re *ReScheduler) isFaultSchedulingTask(task *api.TaskInfo) bool {
	l, ex := task.Pod.Labels[faultSchedulingLabel]
	if !ex {
		return false
	}
	return l == onFaultSchedulingLabel
}
