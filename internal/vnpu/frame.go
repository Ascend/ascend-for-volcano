/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package rescheduling is using for HuaWei Ascend pin fault rescheduling.

*/
package vnpu

import (
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/conf"
	"volcano.sh/volcano/pkg/scheduler/framework"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// CheckVNPUSegmentEnableByConfig Check VNPU segmentEnable by init plugin parameters.
func CheckVNPUSegmentEnableByConfig(configurations []conf.Configuration) bool {
	configuration, err := util.GetConfigFromSchedulerConfigMap(util.CMInitParamKey, configurations)
	if err != nil {
		klog.V(util.LogDebugLev).Info("cannot get configuration, segmentEnable.")
		return false
	}
	// get segmentEnable by user configuration
	segmentEnable, ok := configuration.Arguments[util.SegmentEnable]
	if !ok {
		klog.V(util.LogDebugLev).Info("checkVNPUSegmentEnable doesn't exist presetVirtualDevice.")
		return false
	}
	if segmentEnable == "false" {
		return true
	}
	return false
}

// CheckVNPUSegmentEnable Check VNPU segmentEnable by init plugin parameters.
func CheckVNPUSegmentEnable(ssn *framework.Session) bool {
	if len(ssn.Configurations) == 0 {
		klog.V(util.LogDebugLev).Info("no configurations, segmentEnable will not be changed.")
		return false
	}

	return CheckVNPUSegmentEnableByConfig(ssn.Configurations)
}
