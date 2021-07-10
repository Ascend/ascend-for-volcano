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

Package commonv910 is using for virtual HuaWei Ascend910 schedule.

*/
package commonv910

import (
	"errors"
	"fmt"
	"k8s.io/klog"
	"strings"
	"volcano.sh/volcano/pkg/scheduler/api"
	hwutil "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/util"
)

func validJobResource(job *api.JobInfo) error {
	// Virtual npu resource count
	var vRC int
	vR := map[string]struct{}{
		npuV910CardName2c:  {},
		npuV910CardName4c:  {},
		npuV910CardName8c:  {},
		npuV910CardName16c: {},
	}

	for rType, jobNPU := range job.TotalRequest.ScalarResources {
		r := string(rType)
		if !strings.HasPrefix(r, npu910CardName) {
			continue
		}
		_, ok := vR[r]
		if !ok {
			msg := fmt.Errorf("%s request an invalid type of Vnpu resource", job.Name)
			klog.V(logErrorLev).Infof("invalid request: %v.", msg)
			return errors.New("invalid Vnpu resource requested")
		}
		vRC += int(jobNPU / npuHex)
	}
	klog.V(logInfoLev).Infof("job(%s) requests %d Vnpu.", job.Name, vRC)

	if vRC > 1 {
		// a job can request at most one Vnpu resource
		return errors.New("invalid number of devices requested")
	}

	return nil
}

// GetNPUJobDefaultSelectorConfig get selectors for Vnpu
func (tp *Vnpu) GetNPUJobDefaultSelectorConfig() map[string]string {
	var defaultSchedulerConfig map[string]string
	defaultSchedulerConfig = make(map[string]string, const3)

	defaultSchedulerConfig[archSelector] = huaweiArchArm + "|" + huaweiArchX86

	return defaultSchedulerConfig
}

// for verify npu job must config selector
func (tp *Vnpu) validNPUJobSelector(job *api.JobInfo) error {
	jobSelectors := hwutil.GetJobSelectors(job)
	if len(jobSelectors) == 0 {
		msg := fmt.Errorf("%s getJobSelectors nil", job.Name)
		klog.V(logErrorLev).Infof("%s.", msg.Error())
		return msg
	}

	klog.V(logDebugLev).Infof("%s has selector: %v.", job.Name, jobSelectors)

	defaultSchedulerConfig := tp.GetNPUJobDefaultSelectorConfig()

	if err := hwutil.CompareNPUSelector(job, jobSelectors, defaultSchedulerConfig); err != nil {
		klog.V(logErrorLev).Infof("%v.", err)
		return err
	}

	return nil
}
