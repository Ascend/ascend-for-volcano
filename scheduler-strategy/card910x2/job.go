/*
Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package card910x2 is using for HuaWei A300T Ascend pin affinity schedule.

*/
package card910x2

import (
	"errors"
	"fmt"
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	hwutil "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/util"
)

func getCardNPUJobDefaultSelectorConfig() map[string]string {
	var defaultSchedulerConfig map[string]string
	defaultSchedulerConfig = make(map[string]string, constIntNum3)

	defaultSchedulerConfig[archSelector] = huaweiArchArm + "|" + huaweiArchX86
	defaultSchedulerConfig[acceleratorType] = cardAcceleratorType + "|" + moduleAcceleratorType

	return defaultSchedulerConfig
}

// For verify npu job must config selector.
func validNPUJobSelector(job *api.JobInfo) error {
	jobSelectors := hwutil.GetJobSelectors(job)
	if len(jobSelectors) == 0 {
		msg := fmt.Errorf("%s %s getJobSelectors nil", PluginName, job.Name)
		klog.V(logErrorLev).Infof("%s.", msg.Error())
		return msg
	}

	defaultSchedulerConfig := getCardNPUJobDefaultSelectorConfig()
	klog.V(logDebugLev).Infof("%s card selector: %v default:%v.",
		job.Name, jobSelectors, defaultSchedulerConfig)

	if err := hwutil.CompareNPUSelector(job, jobSelectors, defaultSchedulerConfig); err != nil {
		klog.V(logErrorLev).Infof("%v.", err)
		return err
	}

	return nil
}

// When job requires more than 8 npu, every task need 8 npu.
func checkCardDistributeTrainMode(job *api.JobInfo, nodeNPU int) error {
	taskNum := len(job.Tasks)

	klog.V(logDebugLev).Infof("%s check Card Mode %s has %d tasks.", PluginName, job.Name, taskNum)

	for _, task := range job.Tasks {
		taskNPU, taskError := hwutil.GetTaskNPUNum(task, a300TNPUCardName)
		if taskError != nil {
			return errors.New("no npu task")
		}

		klog.V(logDebugLev).Infof("%s check Card Mode %s require %d npu.", PluginName, task.Name, taskNPU)

		if taskNPU != nodeNPU {
			return fmt.Errorf("CardDistributeTrain %s req npu [%d] but node [%d]", task.Name, taskNPU, nodeNPU)
		}
	}

	return nil
}

func validJobModel(job *api.JobInfo) error {
	var jobNPU int
	var err error
	var nodeNPU = constIntNum2

	if jobNPU, err = hwutil.GetJobReqNPUNum(job, a300TNPUCardName); err != nil {
		return err
	}

	if jobNPU <= nodeNPU {
		if err = hwutil.CheckSingleTrainMode(job); err != nil {
			return err
		}
		return nil
	}

	if err = checkCardDistributeTrainMode(job, nodeNPU); err != nil {
		return err
	}

	return nil
}

func validJobNPUNum(job *api.JobInfo) error {
	jobNPU, err := hwutil.GetJobReqNPUNum(job, a300TNPUCardName)
	if err != nil {
		klog.V(logDebugLev).Infof("job(%s) get npu number failed", job.Name)
		return err
	}

	// only support 1,2,2*n
	if jobNPU == 1 || jobNPU%constIntNum2 == 0 {
		return nil
	}

	return fmt.Errorf("illegal req_npu num: %d in %s mode", jobNPU, cardAcceleratorType)
}
