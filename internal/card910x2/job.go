/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
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

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/util"
)

func getCardNPUJobDefaultSelectorConfig() map[string]string {
	var defaultSchedulerConfig map[string]string
	defaultSchedulerConfig = make(map[string]string, util.NPUIndex3)

	defaultSchedulerConfig[archSelector] = huaweiArchArm + "|" + huaweiArchX86
	defaultSchedulerConfig[acceleratorType] = cardAcceleratorType + "|" + moduleAcceleratorType

	return defaultSchedulerConfig
}

// For verify npu job must config selector.
func validNPUJobSelector(job *api.JobInfo) error {
	jobSelectors := util.GetJobSelectors(job)
	if len(jobSelectors) == 0 {
		msg := fmt.Errorf("%s %s getJobSelectors nil", PluginName, job.Name)
		klog.V(logErrorLev).Infof("%s.", msg.Error())
		return msg
	}

	defaultSchedulerConfig := getCardNPUJobDefaultSelectorConfig()
	klog.V(logDebugLev).Infof("%s card selector: %v default:%v.",
		job.Name, jobSelectors, defaultSchedulerConfig)

	if err := util.CompareNPUSelector(job, jobSelectors, defaultSchedulerConfig); err != nil {
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
		taskNPU, taskError := util.GetTaskNPUNum(task, a300TNPUCardName)
		if taskError != nil {
			return errors.New("no npu task")
		}

		klog.V(logDebugLev).Infof("%s check Card Mode %s require %d npu.", PluginName, task.Name, taskNPU)

		if taskNPU != nodeNPU {
			return fmt.Errorf("distributed card train %s req npu [%d] but node [%d]", task.Name, taskNPU, nodeNPU)
		}
	}

	return nil
}

func validJobModel(job *api.JobInfo) error {
	var jobNPU int
	var err error
	var nodeNPU = util.NPUIndex2

	if jobNPU, err = util.GetJobReqNPUNum(job, a300TNPUCardName); err != nil {
		return err
	}

	if jobNPU <= nodeNPU {
		if err = util.CheckSingleTrainMode(job); err != nil {
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
	jobNPU, err := util.GetJobReqNPUNum(job, a300TNPUCardName)
	if err != nil {
		klog.V(logDebugLev).Infof("job(%s) get npu number failed", job.Name)
		return err
	}

	// only support 1,2,2*n
	if jobNPU == 1 || jobNPU%util.NPUIndex2 == 0 {
		return nil
	}

	return fmt.Errorf("illegal req_npu num: %d in %s mode", jobNPU, cardAcceleratorType)
}
