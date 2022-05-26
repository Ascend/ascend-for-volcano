/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package card310x4 is using for HuaWei A300T Ascend pin affinity schedule.

*/
package card310x4

import (
	"errors"
	"fmt"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/util"
)

func getCardNPUJobDefaultSelectorConfig() map[string]string {
	var defaultSchedulerConfig map[string]string
	defaultSchedulerConfig = make(map[string]string, util.NPUIndex3)

	defaultSchedulerConfig[archSelector] = huaweiArchArm + "|" + huaweiArchX86
	defaultSchedulerConfig[acceleratorType] = cardAcceleratorType + "|" + chipAcceleratorType

	return defaultSchedulerConfig
}

// For verify npu job must config selector.
func validNPUJobSelector(job *api.JobInfo) error {
	jobSelectors := util.GetJobLabels(job)
	if len(jobSelectors) == 0 {
		msg := fmt.Errorf("%s %s GetJobLabels nil", PluginName, job.Name)
		klog.V(util.LogDebugLev).Infof("%s.", msg.Error())
		return msg
	}

	defaultSchedulerConfig := getCardNPUJobDefaultSelectorConfig()
	klog.V(util.LogDebugLev).Infof("%s card selector: %v default:%v.", job.Name, jobSelectors, defaultSchedulerConfig)

	if err := util.CompareNPUSelector(job, jobSelectors, defaultSchedulerConfig); err != nil {
		klog.V(util.LogDebugLev).Infof("%v.", err)
		return err
	}

	return nil
}

// CheckSingleTrainMode Single Train job has only one task.
func CheckSingleTrainMode(job *api.JobInfo) error {
	klog.V(util.LogDebugLev).Infof("checkSingleTrainMode job(%s).", job.Name)

	for _, task := range job.Tasks {
		taskNPU, taskError := util.GetTaskNPUNum(task, a310NPUCardName)
		if taskError != nil {
			return errors.New("no npu task")
		}

		klog.V(util.LogDebugLev).Infof("%s check Card Mode %s require %d npu.", PluginName, task.Name, taskNPU)

		if taskNPU < util.NPUIndex1 || taskNPU > cardNPUNumber {
			return fmt.Errorf("%s single trainning not match task NPU number:%d", job.Name, taskNPU)
		}
	}

	return nil
}

// every task need 4 npu.
func checkCardDistributeTrainMode(job *api.JobInfo, nodeNPU int) error {
	taskNum := len(job.Tasks)

	klog.V(util.LogDebugLev).Infof("%s check Card Mode %s has %d tasks.", PluginName, job.Name, taskNum)

	for _, task := range job.Tasks {
		taskNPU, taskError := util.GetTaskNPUNum(task, a310NPUCardName)
		if taskError != nil {
			return errors.New("no npu task")
		}

		klog.V(util.LogDebugLev).Infof("%s check Card Mode %s require %d npu.", PluginName, task.Name, taskNPU)

		if taskNPU < util.NPUIndex1 || taskNPU > cardNPUNumber {
			return fmt.Errorf("distributed card train %s req npu [%d] but node [%d]", task.Name, taskNPU, nodeNPU)
		}
	}

	return nil
}

func validJobModel(job *api.JobInfo) error {
	var err error
	taskNum := len(job.Tasks)
	var nodeNPU = cardNPUNumber

	if taskNum <= util.NPUIndex1 {
		if err = CheckSingleTrainMode(job); err != nil {
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
	jobNPU, err := util.GetJobReqNPUNum(job, a310NPUCardName)
	if err != nil {
		klog.V(util.LogDebugLev).Infof("job(%s) get npu number failed", job.Name)
		return err
	}

	// The number of NPUs required by a Job greater than 0
	if jobNPU >= 1 {
		return nil
	}

	return fmt.Errorf("illegal req_npu num: %d in %s mode", jobNPU, cardAcceleratorType)
}

// IsJobOfCardModeFromLabel judge job is card mode or card mod by label.
func IsJobOfCardModeFromLabel(job *api.JobInfo) bool {
	jobSelectors := util.GetJobLabels(job)
	if len(jobSelectors) == 0 {
		klog.V(util.LogDebugLev).Infof("job(%s) has no selectors.", job.Name)
		return false
	}

	return util.ValidStringMap(jobSelectors, acceleratorType, cardAcceleratorType)
}
