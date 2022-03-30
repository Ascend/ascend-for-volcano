/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package modulev910 is using for virtual HuaWei Ascend910 schedule.

*/
package modulev910

import (
	"errors"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/vnpu/vnpuutil"
)

// Name get plugin name of 910 for frame init
func (tp *ChipV910) Name() string {
	return PluginName
}

// GetResourceName get plugin NPU resource name.
func (tp *ChipV910) GetResourceName() string {
	return npu910CardName
}

// GetResourcePreVal get plugin NPU resource name prefix.
func (tp *ChipV910) GetResourcePreVal() string {
	return vnpuutil.NPUCardNamePrefix
}

// GetPluginDefaultJobSchedulerConfig get plugin default job scheduler config.
func (tp *ChipV910) GetPluginDefaultJobSchedulerConfig() map[string]string {
	defaultSchedulerConfig := make(map[string]string, util.ConstIntNum1)
	defaultSchedulerConfig[util.ArchSelector] = util.HuaweiArchArm + "|" + util.HuaweiArchX86
	defaultSchedulerConfig[util.AcceleratorType] = util.CardAcceleratorType + "|" + util.ModuleAcceleratorType
	return defaultSchedulerConfig
}

// GetDivideKinds get vNPU all type.
func (tp *ChipV910) GetDivideKinds() []string {
	var vNPUType []string
	vNPUType = append(vNPUType, npuV910CardName2c)
	vNPUType = append(vNPUType, npuV910CardName4c)
	vNPUType = append(vNPUType, npuV910CardName8c)
	vNPUType = append(vNPUType, npuV910CardName16c)
	return vNPUType
}

// GetCoefficients get vNPU all coefficients.
func (tp *ChipV910) GetCoefficients() map[string]int {
	return map[string]int{
		npuV910CardName2c:  npuV910CardCoef2c,
		npuV910CardName4c:  npuV910CardCoef4c,
		npuV910CardName8c:  npuV910CardCoef8c,
		npuV910CardName16c: npuV910CardCoef16c,
	}
}

// GetUnhealthyNameInAnnotation get the unhealthy card label key used in node annotation.
func (tp *ChipV910) GetUnhealthyNameInAnnotation() string {
	return util.Fault910NPU
}

// GetNPUCardCoreKey get source of NPU cores in node annotation.
func (tp *ChipV910) GetNPUCardCoreKey() string {
	return vnpuutil.NPU910CardCoreKey
}

// InitVNPUPlugin init plugin
func (tp *ChipV910) InitVNPUPlugin() error {
	if tp == nil {
		return errors.New("nil ChipV910 obj")
	}
	tp.HwEntity = plugin.HwEntity{
		PluginName:                tp.Name(),
		AnnoName:                  tp.GetResourceName(),
		AnnoPreVal:                tp.GetResourcePreVal(),
		DefaultJobSchedulerConfig: tp.GetPluginDefaultJobSchedulerConfig(),
		AnnoUnhealthyName:         tp.GetUnhealthyNameInAnnotation(),
	}
	tp.DivideKinds = tp.GetDivideKinds()
	// divide vNPU coefficient for each chip.
	tp.Coefficients = tp.GetCoefficients()
	tp.NPUCardCoreKey = tp.GetNPUCardCoreKey()
	return nil
}
