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

Package module910x8 is using for HuaWei A800/9000 Ascend910 pin affinity schedule.

*/
package module910x8

import (
	"errors"
	"fmt"
	"k8s.io/klog"
	"volcano.sh/apis/pkg/apis/scheduling"
	"volcano.sh/volcano/pkg/scheduler/api"
	vapi "volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/rescheduling"
	hwutil "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/util"
)

// If job requires more than 8 npu, every task need 8 npu.
func checkModuleDistributeTrainMode(job *api.JobInfo, nodeNPU int) error {
	taskNum := len(job.Tasks)

	klog.V(logDebugLev).Infof("%s Module DistributeTrainMode %s has %d tasks.", PluginName, job.Name, taskNum)

	for _, task := range job.Tasks {
		taskNPU, taskError := hwutil.GetTaskNPUNum(task, npu800And9000CardName)
		if taskError != nil {
			return errors.New("no npu task")
		}

		klog.V(logDebugLev).Infof("%s Module DistributeTrain %s has %d npu.", PluginName, task.Name, taskNPU)

		if taskNPU != nodeNPU {
			return fmt.Errorf("distributeTrain %s req npu[%d] but req[%d]", task.Name, taskNPU, nodeNPU)
		}
	}

	return nil
}

func validJobModel(job *api.JobInfo) error {
	var jobNPU int
	var err error
	var nodeNPU = nodeNPUNumber

	if jobNPU, err = hwutil.GetJobReqNPUNum(job, npu800And9000CardName); err != nil {
		return err
	}

	if jobNPU <= nodeNPU {
		if err = hwutil.CheckSingleTrainMode(job); err != nil {
			return err
		}

		return nil
	}

	if err = checkModuleDistributeTrainMode(job, nodeNPU); err != nil {
		return err
	}

	return nil
}

func getModuleNPUJobDefaultSelectorConfig() map[string]string {
	var defaultSchedulerConfig map[string]string
	defaultSchedulerConfig = make(map[string]string, constIntNum3)

	defaultSchedulerConfig[archSelector] = huaweiArchArm + "|" + huaweiArchX86

	return defaultSchedulerConfig
}

// For verify npu job must config selector.
func validNPUJobSelector(job *api.JobInfo) error {
	jobSelectors := hwutil.GetJobSelectors(job)
	if len(jobSelectors) == 0 {
		msg := fmt.Errorf("%s getJobSelectors nil", job.Name)
		klog.V(logErrorLev).Infof("%s.", msg.Error())
		return msg
	}

	defaultSchedulerConfig := getModuleNPUJobDefaultSelectorConfig()

	klog.V(logDebugLev).Infof("%s module selector: %v default:%v.", job.Name, jobSelectors, defaultSchedulerConfig)

	if err := hwutil.CompareNPUSelector(job, jobSelectors, defaultSchedulerConfig); err != nil {
		klog.V(logErrorLev).Infof("%v.", err)
		return err
	}

	return nil
}

func validJobNPUNum(job *api.JobInfo) error {
	jobNPU, err := hwutil.GetJobReqNPUNum(job, npu800And9000CardName)
	if err != nil {
		klog.V(logDebugLev).Infof("job(%s) get npu number failed", job.Name)
		return err
	}

	if jobNPU == 1 ||
		jobNPU == constIntNum2 ||
		jobNPU == npuNumPerHccs ||
		jobNPU%nodeNPUNumber == 0 {
		return nil
	}

	return fmt.Errorf("illegal req_npu num:%d in %s mode", jobNPU, moduleAcceleratorType)
}

func isMyJob(job *vapi.JobInfo) error {
	_, err := hwutil.GetJobReqNPUNum(job, npu800And9000CardName)
	if err != nil {
		return errors.New(jobNoNPUCard)
	}

	if hwutil.IsJobOfCardMode(job) {
		return errors.New("job is card mode")
	}

	return nil
}

func isMyTask(task *vapi.TaskInfo) error {
	_, err := hwutil.GetTaskNPUNum(task, npu800And9000CardName)
	if err != nil {
		return errors.New(jobNoNPUCard)
	}

	if hwutil.IsTaskOfCardMode(task) {
		return errors.New("task is card mode")
	}

	return nil
}

func get910x8RunningJobs(jobs map[api.JobID]*api.JobInfo) (map[string]*api.JobInfo, error) {
	var myJobs = make(map[string]*api.JobInfo, constIntNum3)
	for _, job := range jobs {
		if job.PodGroup.Status.Phase != scheduling.PodGroupRunning {
			continue
		}
		if err := isMyJob(job); err == nil {
			myJobs[job.Name] = job
		}
	}

	if len(myJobs) == 0 {
		return nil, fmt.Errorf("nil %s jobs", PluginName)
	}

	return myJobs, nil
}

// restartFaultJob restart fault job.
func restartFaultJob(ssn *framework.Session, fJobs []rescheduling.FaultNPUJob, jobs map[string]*api.JobInfo) error {
	// 1.Get fault jobs.
	restartJobs, getErr := rescheduling.GetRestartNPUFaultJobs(fJobs, jobs)
	if getErr != nil {
		klog.V(logDebugLev).Infof("restartFaultJob %v.", getErr)
		return nil
	}
	// 2.Restart job.
	for _, restartJob := range restartJobs {
		klog.V(logInfoLev).Infof("%s need restart.", restartJob.Name)
		if err := plugin.RestartJob(ssn, restartJob, jobRestartReason); err != nil {
			klog.V(logInfoLev).Infof("RestartJob %v.", err)
			return err
		}
	}
	return nil
}
