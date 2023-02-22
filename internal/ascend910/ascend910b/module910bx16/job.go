/*
Copyright(C)2020-2023. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package module910bx16 is using for HuaWei A800/9000 Ascend910B A+X pin affinity schedule.
*/
package module910bx16

import (
	"fmt"

	"k8s.io/klog"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// checkJobForm to check job ring-controller.atlas for future unification.
func (tp *module910bx16) checkJobForm() error {
	// for vcJob and deployment.
	lValue, ok := tp.Label[util.AcceleratorType]
	if !ok {
		return fmt.Errorf("%s not has no label:%s", tp.Name, util.AcceleratorType)
	}

	if lValue != module910BAcceleratorValue {
		return fmt.Errorf("%s label:%s not right %s", tp.Name, util.AcceleratorType, lValue)
	}
	return nil
}

// checkJobArch to check job host-arch for A+X not same as A+K.
func (tp *module910bx16) checkJobArch() error {
	// for vcJob and deployment.
	lValue, ok := tp.Selector[util.ArchSelector]
	if !ok {
		return fmt.Errorf("%s not has no selector:%s", tp.Name, util.ArchSelector)
	}

	if lValue != util.HuaweiArchX86 {
		return fmt.Errorf("%s selector:%s not right %s", tp.Name, util.ArchSelector, lValue)
	}
	return nil
}

// checkJobTrainMode to check job train mode:distribute and single.
func (tp *module910bx16) checkJobTrainMode() error {
	nTaskNum := tp.GetVTaskNumInVJob()
	if nTaskNum == 0 {
		klog.V(util.LogErrorLev).Infof("GetVTaskNumInVJob %s has no npu tasks.", tp.Name)
		return fmt.Errorf("%s no npu job", tp.Name)
	}

	if nTaskNum == 1 {
		if err := tp.checkSingleTrainMode(); err != nil {
			klog.V(util.LogErrorLev).Infof("%s checkSingleTrainMode %s: %s", tp.GetPluginName(), tp.Name, err)
			return err
		}
		return nil
	}

	if err := tp.checkModuleDistributeTrainMode(); err != nil {
		klog.V(util.LogErrorLev).Infof("%s check distribute %s err: %s", tp.GetPluginName(), tp.Name, err)
		return err
	}

	return nil
}

// CheckSingleTrainMode Single Train job has only one task.
func (tp *module910bx16) checkSingleTrainMode() error {
	klog.V(util.LogDebugLev).Infof("checkSingleTrainMode job(%s) has %d tasks.", tp.Name, len(tp.Tasks))

	jobNPU := tp.ReqNPUNum
	if jobNPU == 1 || jobNPU == util.NPUIndex2 || jobNPU == util.NPUIndex4 || jobNPU == util.
		NPUIndex8 || jobNPU == util.NPUIndex16 {
		return nil
	}
	return fmt.Errorf("%s checkSingleTrainMode %s req npu not in [1,2,4,8,16]", tp.GetPluginName(), tp.Name)
}

// If job requires more than 8 npu, every task need 8 npu.
func (tp *module910bx16) checkModuleDistributeTrainMode() error {
	klog.V(util.LogDebugLev).Infof("half DistributeTrainMode %s has %d tasks.", tp.Name, len(tp.Tasks))

	for _, task := range tp.Tasks {
		if !task.IsVNPUTask() {
			continue
		}

		if task.ReqNPUNum != tp.MaxNodeNPUNum {
			return fmt.Errorf("checkModuleDistributeTrainMode %s req %d not %d", task.Name, task.ReqNPUNum,
				tp.MaxNodeNPUNum)
		}
	}
	return nil
}

func (tp *module910bx16) judgeNodeAndTaskNPU(taskNPU int, nodeTop []int) error {
	var reFlag = false

	sNodeInf := initSelectNodeInf(nodeTop)

	switch taskNPU {
	case 1, util.NPUIndex2, util.NPUIndex4, util.NPUIndex8:
		reFlag = (sNodeInf.LeftNPUNum >= taskNPU) || (sNodeInf.RightNPUNum >= taskNPU)
	case tp.MaxNodeNPUNum:
		reFlag = sNodeInf.AllNPUNum == tp.MaxNodeNPUNum
	default:
		// single pod(task) cannot require npu not belong to mode
		// this kind job has been deal with job logical
		reFlag = false
	}

	if reFlag {
		return nil
	}
	meetErr := fmt.Errorf("%v not meet req npu(%d)", nodeTop, taskNPU)
	klog.V(util.LogErrorLev).Infof("cardIDs:%#v not meet task reqNum<%d>.", nodeTop, taskNPU)
	return meetErr
}

func getNPUAllocPriorityArray(taskNPUNumber int) ([]int, error) {
	var priorityArray []int
	var err error

	switch taskNPUNumber {
	case 1:
		// priority:1>2>3>4>5>6>7>8
		priorityArray = []int{1, util.NPUIndex2, util.NPUIndex3, util.NPUIndex4, util.NPUIndex5, util.NPUIndex6,
			util.NPUIndex7, util.NPUIndex8}
	case util.NPUIndex2:
		// priority：2>3>4>5>6>7>8
		priorityArray = []int{util.NPUIndex2, util.NPUIndex3, util.NPUIndex4, util.NPUIndex5, util.NPUIndex6,
			util.NPUIndex7, util.NPUIndex8}
	case util.NPUIndex4:
		// priority：4>5>6>7>8
		priorityArray = []int{util.NPUIndex4, util.NPUIndex5, util.NPUIndex6,
			util.NPUIndex7, util.NPUIndex8}
	case util.NPUIndex8:
		// priority：16
		priorityArray = []int{util.NPUIndex8}
	case nodeNPUNumber:
		priorityArray = []int{nodeNPUNumber}
	default:
		// For normal,can not be here. The pre function validate job has done this.
		err = fmt.Errorf("illegal request npu number: %d", taskNPUNumber)
	}

	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s %s.", SchedulerName, err.Error())
		return priorityArray, err
	}

	return priorityArray, nil
}
