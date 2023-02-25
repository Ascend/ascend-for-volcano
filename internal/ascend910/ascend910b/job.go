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
Package ascend910b is using for HuaWei Ascend 910B pin affinity schedule.
*/
package ascend910b

import (
	"fmt"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// Valid910bNPUJob check the 910b job req npu num and mode
func (ab *Base910b) Valid910bNPUJob() *api.ValidateResult {
	vResult := &api.ValidateResult{}
	var vErr error
	defer func() {
		if vErr != nil {
			vResult.Pass = false
			vResult.Reason = vErr.Error()
			vResult.Message = vErr.Error()
			return
		}
	}()

	// 1. check parameter.
	if ab == nil {
		vErr = fmt.Errorf("nil plugin %s", ab.GetPluginName())
		klog.V(util.LogErrorLev).Infof("ValidNPUJob err: %s.", vErr)
		return vResult
	}

	// 2.check ring-controller.atlas
	if vErr = ab.CheckJobForm(); vErr != nil {
		klog.V(util.LogErrorLev).Infof("checkJobForm: %s.", vErr)
		return vResult
	}

	// 3.check host-arch for A+X not same as A+K
	if vErr = ab.CheckJobArch(); vErr != nil {
		klog.V(util.LogErrorLev).Infof("checkJobArch: %s.", vErr)
		return vResult
	}

	// 4.check job train mode:distribute and single.
	if vErr = ab.checkJobTrainMode(); vErr != nil {
		klog.V(util.LogErrorLev).Infof("checkJobTrainMode: %s.", vErr)
		return vResult
	}

	return nil
}

// CheckJobForm to check job ring-controller.atlas for future unification.
func (ab *Base910b) CheckJobForm() error {
	// for vcJob and deployment.
	lValue, ok := ab.Label[util.JobKindKey]
	if !ok {
		return fmt.Errorf("%s not has no label:%s", ab.Name, util.JobKindKey)
	}

	if lValue != ab.GetAcceleratorValue() {
		return fmt.Errorf("%s label:%s not right %s", ab.Name, ab.GetAcceleratorValue(), lValue)
	}
	return nil
}

// CheckJobArch to check job host-arch for A+X not same as A+K.
func (ab *Base910b) CheckJobArch() error {
	// for vcJob and deployment.
	lValue, ok := ab.Selector[util.ArchSelector]
	if !ok {
		return fmt.Errorf("%s not has no selector:%s", ab.Name, util.ArchSelector)
	}

	if lValue != ab.GetArch() {
		return fmt.Errorf("%s selector:%s not right %s", ab.Name, util.ArchSelector, lValue)
	}
	return nil
}

// checkJobTrainMode to check job train mode:distribute and single.
func (ab *Base910b) checkJobTrainMode() error {
	nTaskNum := ab.GetNPUTaskNumInJob()
	if nTaskNum == 0 {
		klog.V(util.LogErrorLev).Infof("GetVTaskNumInVJob %s has no npu tasks.", ab.Name)
		return fmt.Errorf("%s no npu job", ab.Name)
	}

	if nTaskNum == 1 {
		if err := ab.checkSingleTrainMode(); err != nil {
			klog.V(util.LogErrorLev).Infof("%s checkSingleTrainMode %s: %s", ab.GetPluginName(), ab.Name, err)
			return err
		}
		return nil
	}

	if err := ab.checkModuleDistributeTrainMode(); err != nil {
		klog.V(util.LogErrorLev).Infof("%s check distribute %s err: %s", ab.GetPluginName(), ab.Name, err)
		return err
	}

	return nil
}

// CheckSingleTrainMode Single Train job has only one task.
func (ab *Base910b) checkSingleTrainMode() error {
	klog.V(util.LogDebugLev).Infof("checkSingleTrainMode job(%s) has %d tasks.", ab.Name, len(ab.Tasks))

	if ab.CheckSingleAllowNum(ab.ReqNPUNum) {
		return nil
	}
	return fmt.Errorf("%s checkSingleTrainMode %s req npu not in [1,2,4,8,16]", ab.GetPluginName(), ab.Name)
}

// If job requires more than 8 npu, every task need 8 npu.
func (ab *Base910b) checkModuleDistributeTrainMode() error {
	klog.V(util.LogDebugLev).Infof("half DistributeTrainMode %s has %d tasks.", ab.Name, len(ab.Tasks))

	for _, task := range ab.Tasks {
		if !task.IsNPUTask() {
			continue
		}

		if task.ReqNPUNum != ab.MaxNodeNPUNum {
			return fmt.Errorf("checkModuleDistributeTrainMode %s req %d not %d", task.Name, task.ReqNPUNum,
				ab.MaxNodeNPUNum)
		}
	}
	return nil
}

func (ab *Base910b) getNPUAllocPriorityArray(taskNPUNumber int) ([]int, error) {
	var priorityArray []int
	var err error
	if !ab.CheckSingleAllowNum(taskNPUNumber) {
		err = fmt.Errorf("illegal request npu number: %d", taskNPUNumber)
		klog.V(util.LogErrorLev).Infof("%s %s.", ab.GetPluginName(), err)
		return nil, err
	}

	for i := taskNPUNumber; i <= ab.MaxNodeNPUNum/util.NPUIndex2; i++ {
		priorityArray = append(priorityArray, i)
	}

	if taskNPUNumber == ab.MaxNodeNPUNum {
		priorityArray = []int{ab.MaxNodeNPUNum}
	}
	return priorityArray, nil
}

func (ab *Base910b) selectNPUFromNode(task *api.TaskInfo, node plugin.NPUNode) ([]int, error) {
	taskNPUNum, err := ab.GetTaskReqNPUNum(task)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s ScoreBestNPUNodes err: %s", ab.GetPluginName(), err)
		return nil, err
	}
	nodeTop, err := ab.GetUsableTopFromNode(node)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s ScoreBestNPUNodes err: %s", ab.GetPluginName(), err)
		return nil, err
	}
	if taskNPUNum == ab.MaxNodeNPUNum {
		if len(nodeTop) == ab.MaxNodeNPUNum {
			return nodeTop, nil
		}
		err = fmt.Errorf("%s %#v can not meet task req:%d", node.Name, nodeTop, taskNPUNum)
		klog.V(util.LogErrorLev).Infof("%s ScoreBestNPUNodes err: %s", ab.GetPluginName(), err)
		return nil, err
	}
	priorityArray, err := ab.getNPUAllocPriorityArray(taskNPUNum)
	if err != nil {
		klog.V(util.LogErrorLev).Info(err)
		return nil, err
	}
	klog.V(util.LogInfoLev).Infof("%s selectNPUFromNode %s[%d] priority:%v in %v.", ab.GetPluginName(),
		task.Name, taskNPUNum, priorityArray, nodeTop)

	leftHCCSArray, rightHCCSArray := ab.getNodeHccsArray(nodeTop)
	for _, priority := range priorityArray {
		if priority == len(leftHCCSArray) {
			return leftHCCSArray[:taskNPUNum], nil
		}
		if priority == len(rightHCCSArray) {
			return rightHCCSArray[:taskNPUNum], nil
		}
	}
	err = fmt.Errorf("node<%s> top<%v> can not meet task req<%d>", node.Name, len(nodeTop), taskNPUNum)
	klog.V(util.LogErrorLev).Infof("%s ScoreBestNPUNodes err: %s", ab.GetPluginName(), err)
	return nil, err
}
