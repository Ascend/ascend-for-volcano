/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package comvnpu is using for virtual HuaWei Ascend910 schedule.

*/
package comvnpu

import (
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/vnpu/vnpuutil"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

const (
	// PluginName the vnpu's plugin name.
	PluginName = vnpuutil.PluginName
)

// VNPU common type
type VNPU struct {
	// abstract v910x2,v910x8,v310P.
	Plugin VNPUHandler
	// the element for vnpu, all VNPU scheduler plugin need include.
	Attr vnpuutil.ComVNPU
	plugin.HwNPUSchedulerPlugin
}

// VNPUHandler The VNPU scheduler plugin must realize interface.
type VNPUHandler interface {
	Name() string
	// InitVNPUPlugin init VNPU scheduler plugin.
	InitVNPUPlugin() error
	// GetResourceName get plugin NPU resource name.
	GetResourceName() string
	// GetResourcePreVal get plugin NPU resource name prefix.
	GetResourcePreVal() string
	// GetDivideKinds get vNPU all type.
	GetDivideKinds() []string
	// GetCoefficients get vNPU all coefficients.
	GetCoefficients() map[string]int
	// GetNPUCardCoreKey the source of NPU cores in node annotation.
	GetNPUCardCoreKey() string
	// GetUnhealthyNameInAnnotation get the chip unhealthy name ,defined in node annotation.
	GetUnhealthyNameInAnnotation() string
}
