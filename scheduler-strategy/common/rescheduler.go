/*
Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.
*/

// common Common pkg to scheduler infer
package common

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog"
	"reflect"
	"strings"
	"time"
	"volcano.sh/apis/pkg/apis/scheduling"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/rescheduling"
)

// PreHandleFaultNPUFn PreHandleFaultNPUFn
func (re *ReScheduler) PreHandleFaultNPUFn(ssn *framework.Session) error {
	klog.V(LogDebugLev).Infof("%s enter preHandleFaultNPUFn.", re.AnnoUnHealthy)
	defer klog.V(LogDebugLev).Infof("%s leave preHandleFaultNPUFn.", re.AnnoUnHealthy)

	// 1.get fault npu and node.
	fNode, err := re.getInoperableNPUCards(ssn.Nodes)
	if err != nil {
		klog.V(LogDebugLev).Infof("%s preHandleFaultNPUFn %v.", re.AnnoUnHealthy, err)
	}
	// 2.Determine if it is a  jobs.
	tasks, jobGetErr := re.getFaultTasks(ssn.Jobs, fNode)
	if jobGetErr != nil {
		klog.V(LogDebugLev).Infof("%s : %v.getFaultTasks ", re.AnnoUnHealthy, jobGetErr)
		return nil
	}

	for _, task := range tasks {
		if task == nil {
			continue
		}
		klog.V(LogErrorLev).Infof("evict task %+v", task)
		if err := ssn.Evict(task, jobRestartReason); err != nil {
			klog.V(LogErrorLev).Infof("%s Failed to restart %s : %v", re.AnnoUnHealthy, task.UID, err)
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
		return nil, fmt.Errorf("%s get nil npus", node.Name)
	}

	faultNPUs := strings.Split(npuStrings, ",")
	return faultNPUs, nil
}

func (re *ReScheduler) getInoperableNPUCards(nodes map[string]*api.NodeInfo) ([]rescheduling.FaultNPUsOnNode, error) {
	var faultNPUs []rescheduling.FaultNPUsOnNode

	for _, nodeInfo := range nodes {
		npus, err := re.getNodeFaultNPUs(nodeInfo)
		if err != nil {
			klog.V(LogDebugLev).Infof("getNodeFaultNPUs err:%v.", err)
		}

		faultNPUs = append(faultNPUs, rescheduling.FaultNPUsOnNode{nodeInfo.Name, npus,
			nil, time.Now().Unix()})
	}

	if len(faultNPUs) == 0 {
		return nil, fmt.Errorf("%v nil inoperable NPU", reflect.ValueOf(nodes).MapKeys())
	}
	klog.V(LogDebugLev).Infof("getInoperableNPUCards %+v.", faultNPUs)

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
		klog.V(LogDebugLev).Infof("%s Myjobs  PodGroupRunning ", job.Name)
		for _, task := range job.Tasks {
			if !re.isFaultSchedulingTask(task) {
				klog.V(LogDebugLev).Infof("isFaultSchedulingTask false taskname info:%s.", task.Name)
				continue
			}
			if re.isFaultTask(task, nodes) {
				myTasks = append(myTasks, task)
			}
		}
	}

	if len(myTasks) == 0 {
		klog.V(LogDebugLev).Infof("len myTasks is 0.")
		return nil, fmt.Errorf("nil %s jobs", re.AnnoUnHealthy)
	}

	klog.V(LogDebugLev).Infof("job getFaultTasks myTasks len %d", len(myTasks))
	return myTasks, nil
}

func (re *ReScheduler) isFaultTask(task *api.TaskInfo, nodes []rescheduling.FaultNPUsOnNode) bool {
	klog.V(LogDebugLev).Infof("isFaultTask task :%s.   nodes %+v", task.NodeName, nodes)
	for _, node := range nodes {
		if node.NodeName == task.NodeName {
			anno, exist := task.Pod.Annotations[re.AnnoName]
			if !exist {
				klog.V(LogDebugLev).Infof("Annotations false pod anno info:%s.", anno)
				continue
			}
			klog.V(LogDebugLev).Infof("Annotations anno  :%+v.   nodes %+v", anno, node.FaultNPUs)
			if re.checkFaultNPU(anno, node.FaultNPUs) {
				klog.V(LogDebugLev).Infof("checkFaultNPU false pod anno info:%+v.", anno)
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
				klog.V(LogDebugLev).Infof("checkFaultNPU true :%+v.   nodes %+v", anno, faultNPUs)
				return true
			}
		}
	}
	klog.V(LogDebugLev).Infof("checkFaultNPU false :%+v.   nodes %+v", anno, faultNPUs)
	return false
}

func (re *ReScheduler) isFaultSchedulingTask(task *api.TaskInfo) bool {
	l, ex := task.Pod.Labels[faultSchedulingLabel]
	if !ex {
		return false
	}
	return l == onFaultSchedulingLabel
}
