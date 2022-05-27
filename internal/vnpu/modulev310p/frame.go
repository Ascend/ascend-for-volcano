/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package modulev310p is using for virtual HuaWei 310P chips schedule.

*/
package modulev310p

import (
	"errors"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/vnpu/vnpuutil"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

// Name get plugin name of 310P for frame init
func (tp *ChipV310P) Name() string {
	return PluginName
}

// GetResourceName get plugin NPU resource name.
func (tp *ChipV310P) GetResourceName() string {
	return npu310PCardName
}

// GetResourcePreVal get plugin NPU resource name prefix.
func (tp *ChipV310P) GetResourcePreVal() string {
	return vnpuutil.NPUCardNamePrefix
}

// GetDivideKinds get vNPU all type.
func (tp *ChipV310P) GetDivideKinds() []string {
	var vNPUType []string
	vNPUType = append(vNPUType, npuV310PCardName1c)
	vNPUType = append(vNPUType, npuV310PCardName2c)
	vNPUType = append(vNPUType, npuV310PCardName4c)
	return vNPUType
}

// GetCoefficients get vNPU all coefficients.
func (tp *ChipV310P) GetCoefficients() map[string]int {
	return map[string]int{
		npuV310PCardName1c: npuV310PCardCoef1c,
		npuV310PCardName2c: npuV310PCardCoef2c,
		npuV310PCardName4c: npuV310PCardCoef4c,
	}
}

// GetUnhealthyNameInAnnotation get the unhealthy card label key used in node annotation.
func (tp *ChipV310P) GetUnhealthyNameInAnnotation() string {
	return util.Fault310PNPU
}

// GetNPUCardCoreKey get source of NPU cores in node annotation.
func (tp *ChipV310P) GetNPUCardCoreKey() string {
	return vnpuutil.NPU310PCardCoreKey
}

// GetPluginDefaultJobSchedulerConfig get plugin default job scheduler config.
func (tp *ChipV310P) GetPluginDefaultJobSchedulerConfig() map[string]string {
	defaultSchedulerConfig := make(map[string]string, util.NPUIndex1)
	defaultSchedulerConfig[util.ArchSelector] = util.HuaweiArchArm + "|" + util.HuaweiArchX86
	return defaultSchedulerConfig
}

// InitVNPUPlugin init plugin
func (tp *ChipV310P) InitVNPUPlugin() error {
	if tp == nil {
		return errors.New("nil ModuleV310P obj")
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
