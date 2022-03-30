/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package common is using for HuaWei common infer Ascend pin affinity schedule.
*/
package common

import (
	"fmt"
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/util"
)

func (cn *Scheduler) judgeNodeAndTaskNPU(taskNPU int, nodeNPUTopology []int) error {
	if len(nodeNPUTopology) >= taskNPU {
		return nil
	}

	var meetErr = fmt.Errorf("req npu(%d) illegal", taskNPU)
	klog.V(util.LogDebugLev).Infof("%s %v.", cn.PluginName, meetErr)
	return meetErr
}
