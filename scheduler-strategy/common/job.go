/*
Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package chip710 is using for HuaWei A300T Ascend pin affinity schedule.

*/
package common

import (
	"errors"
	"fmt"
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/util"
)

func (cn *CommonScheduler) getCardNPUJobDefaultSelectorConfig() map[string]string {
	return cn.DefaultJobSchedulerConfig
}

// For verify npu job must config selector.
func (cn *CommonScheduler) validNPUJobSelector(job *api.JobInfo) error {
	jobSelectors := util.GetJobLabels(job)
	if len(jobSelectors) == 0 {
		msg := fmt.Errorf("%s %s getJobSelectors nil", cn.PluginName, job.Name)
		klog.V(LogErrorLev).Infof("%s.", msg.Error())
		return msg
	}

	defaultSchedulerConfig := cn.getCardNPUJobDefaultSelectorConfig()
	klog.V(LogDebugLev).Infof("%s card selector: %v default:%v.", job.Name, jobSelectors, defaultSchedulerConfig)

	if err := util.CompareNPUSelector(job, jobSelectors, defaultSchedulerConfig); err != nil {
		klog.V(LogErrorLev).Infof("%v.", err)
		return err
	}

	return nil
}

func (cn *CommonScheduler) validJobModel(job *api.JobInfo) error {
	klog.V(LogDebugLev).Infof("validJobModel job(%s).", job.Name)

	for _, task := range job.Tasks {
		taskNPU, taskError := util.GetTaskNPUNum(task, cn.AnnoName)
		if taskError != nil {
			return errors.New("no npu task")
		}
		klog.V(LogDebugLev).Infof("%s check Card Mode %s require %d npu.", cn.PluginName, task.Name, taskNPU)

		if taskNPU < ConstIntNum1 || taskNPU > NodeNPUNumber {
			return fmt.Errorf("%s single trainning not match task NPU number:%d", job.Name, taskNPU)
		}
	}

	return nil
}

func (cn *CommonScheduler) validJobNPUNum(job *api.JobInfo) error {
	_, err := util.GetJobReqNPUNum(job, cn.AnnoName)
	return err
}
