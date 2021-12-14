/*
Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package common is using for HuaWei A300T Ascend pin affinity schedule.
*/
package common

import (
	"fmt"
	"k8s.io/klog"
)

func (cn *CommonScheduler) judgeNodeAndTaskNPU(taskNPU int, nodeNPUTopology []int) error {
	if len(nodeNPUTopology) >= taskNPU {
		return nil
	}

	var meetErr = fmt.Errorf("req npu(%d) illegal", taskNPU)
	klog.V(LogErrorLev).Infof("%s %v.", cn.PluginName, meetErr)
	return meetErr
}
