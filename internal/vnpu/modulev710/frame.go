/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package modulev710 is using for virtual HuaWei 710 chips schedule.

*/
package modulev710

import (
	"errors"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/vnpu/vnpuutil"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

// Name get plugin name of 710 for frame init
func (tp *ChipV710) Name() string {
	return PluginName
}

// GetResourceName get plugin NPU resource name.
func (tp *ChipV710) GetResourceName() string {
	return npu710CardName
}

// GetResourcePreVal get plugin NPU resource name prefix.
func (tp *ChipV710) GetResourcePreVal() string {
	return vnpuutil.NPUCardNamePrefix
}

// GetDivideKinds get vNPU all type.
func (tp *ChipV710) GetDivideKinds() []string {
	var vNPUType []string
	vNPUType = append(vNPUType, npuV710CardName1c)
	vNPUType = append(vNPUType, npuV710CardName2c)
	vNPUType = append(vNPUType, npuV710CardName4c)
	return vNPUType
}

// GetCoefficients get vNPU all coefficients.
func (tp *ChipV710) GetCoefficients() map[string]int {
	return map[string]int{
		npuV710CardName1c: npuV710CardCoef1c,
		npuV710CardName2c: npuV710CardCoef2c,
		npuV710CardName4c: npuV710CardCoef4c,
	}
}

// GetUnhealthyNameInAnnotation get the unhealthy card label key used in node annotation.
func (tp *ChipV710) GetUnhealthyNameInAnnotation() string {
	return util.Fault710NPU
}

// GetNPUCardCoreKey get source of NPU cores in node annotation.
func (tp *ChipV710) GetNPUCardCoreKey() string {
	return vnpuutil.NPU710CardCoreKey
}

// GetPluginDefaultJobSchedulerConfig get plugin default job scheduler config.
func (tp *ChipV710) GetPluginDefaultJobSchedulerConfig() map[string]string {
	defaultSchedulerConfig := make(map[string]string, util.NPUIndex1)
	defaultSchedulerConfig[util.ArchSelector] = util.HuaweiArchArm + "|" + util.HuaweiArchX86
	return defaultSchedulerConfig
}

// InitVNPUPlugin init plugin
func (tp *ChipV710) InitVNPUPlugin() error {
	if tp == nil {
		return errors.New("nil ModuleV710 obj")
	}
	tp.HwEntity = plugin.HwEntity{
		PluginName:                tp.Name(),
		AnnoName:                  tp.GetResourceName(),
		AnnoPreVal:                tp.GetResourcePreVal(),
		AnnoUnhealthyName:         tp.GetUnhealthyNameInAnnotation(),
		DefaultJobSchedulerConfig: tp.GetPluginDefaultJobSchedulerConfig(),
	}
	tp.DivideKinds = tp.GetDivideKinds()
	// divide vNPU coefficient for each chip.
	tp.Coefficients = tp.GetCoefficients()
	tp.NPUCardCoreKey = tp.GetNPUCardCoreKey()
	return nil
}
