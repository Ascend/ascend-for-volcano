/*
Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.

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

Package chip310x4 is using for HuaWei A300T Ascend pin affinity schedule.

*/
package chip310x4

import (
	"errors"
	"fmt"
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/util"
)

func getCardNPUJobDefaultSelectorConfig() map[string]string {
	var defaultSchedulerConfig map[string]string
	defaultSchedulerConfig = make(map[string]string, constIntNum3)

	defaultSchedulerConfig[archSelector] = huaweiArchArm + "|" + huaweiArchX86
	defaultSchedulerConfig[acceleratorType] = cardAcceleratorType + "|" + chipAcceleratorType

	return defaultSchedulerConfig
}

// For verify npu job must config selector.
func validNPUJobSelector(job *api.JobInfo) error {
	jobSelectors := util.GetJobLabels(job)
	if len(jobSelectors) == 0 {
		msg := fmt.Errorf("%s %s getJobSelectors nil", PluginName, job.Name)
		klog.V(logErrorLev).Infof("%s.", msg.Error())
		return msg
	}

	defaultSchedulerConfig := getCardNPUJobDefaultSelectorConfig()
	klog.V(logDebugLev).Infof("%s card selector: %v default:%v.", job.Name, jobSelectors, defaultSchedulerConfig)

	if err := util.CompareNPUSelector(job, jobSelectors, defaultSchedulerConfig); err != nil {
		klog.V(logErrorLev).Infof("%v.", err)
		return err
	}

	return nil
}

// CheckSingleTrainMode Single Train job has only one task.
func CheckSingleTrainMode(job *api.JobInfo) error {

	klog.V(logDebugLev).Infof("checkSingleTrainMode job(%s).", job.Name)

	for _, task := range job.Tasks {
		taskNPU, taskError := util.GetTaskNPUNum(task, a310NPUChipName)
		if taskError != nil {
			return errors.New("no npu task")
		}

		klog.V(logDebugLev).Infof("%s check Card Mode %s require %d npu.", PluginName, task.Name, taskNPU)

		if taskNPU < constIntNum1 || taskNPU > nodeNPUNumber {
			return fmt.Errorf("%s single trainning not match task NPU number:%d", job.Name, taskNPU)
		}
	}

	return nil
}

func validJobModel(job *api.JobInfo) error {
	var err error
	taskNum := len(job.Tasks)

	if taskNum <= constIntNum1 {
		if err = CheckSingleTrainMode(job); err != nil {
			return err
		}
		return nil
	}

	return nil
}

func validJobNPUNum(job *api.JobInfo) error {
	jobNPU, err := util.GetJobReqNPUNum(job, a310NPUChipName)
	if err != nil {
		klog.V(logDebugLev).Infof("job(%s) get npu number failed", job.Name)
		return err
	}

	// The number of NPUs required by a Job greater than 0
	if jobNPU >= 1 {
		return nil
	}

	return fmt.Errorf("illegal req_npu num: %d in %s mode", jobNPU, chipAcceleratorType)
}
