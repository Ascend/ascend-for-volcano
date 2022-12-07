/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package ascend310p is using for HuaWei 310P Ascend pin affinity schedule.
*/
package ascend310p

import (
	"errors"
	"fmt"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/rescheduling"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// New return npu plugin.
func New(npuName string) plugin.ISchedulerPlugin {
	npuPlugin := &ascend310P{}
	npuPlugin.SetPluginName(npuName)
	npuPlugin.SetAnnoName(util.NPU310PCardName)
	npuPlugin.SetAnnoPreVal(util.NPU310PCardNamePre)
	npuPlugin.SetDefaultJobSchedulerConfig(nil)
	npuPlugin.SetMaxNodeNPUNum(maxNodeNPUNum)
	npuPlugin.InitVNPU()
	return npuPlugin
}

// PreStartAction pre-processing actions for rescheduling
func (tp *ascend310P) PreStartAction(ssn *framework.Session) error {
	klog.V(util.LogInfoLev).Infof("Entering PreStartAction of %s", util.NPU310PCardName)
	defer klog.V(util.LogInfoLev).Infof("Leaving PreStartAction of %s", util.NPU310PCardName)
	if tp == nil || ssn == nil || tp.FrameAttr.KubeClient == nil {
		return fmt.Errorf("%s handler not enabled or ssn is nil: %s", util.NPU310PCardName, util.ArgumentError)
	}
	reErr := tp.preStartRescheduling(ssn)
	vErr := tp.preStartVNPU(ssn)
	return fmt.Errorf("%s %s", reErr, vErr)
}

// PreStopAction post-processing actions for re-scheduling
func (tp *ascend310P) PreStopAction(env *plugin.ScheduleEnv) error {
	klog.V(util.LogInfoLev).Infof("enter PreStopAction of %s...", util.NPU310PCardName)
	defer klog.V(util.LogInfoLev).Infof("leave PreStopAction of %s...", util.NPU310PCardName)
	if tp == nil || tp.reHandle == nil || env == nil {
		return fmt.Errorf("%s reSchedule not enabled or nil env: %s", util.NPU310PCardName, util.ArgumentError)
	}
	if err := tp.reHandle.WriteReSchedulerCacheToEnvCache(env, rescheduling.CmFaultJob310PKind); err != nil {
		return err
	}
	return nil
}

// ValidNPUJob check job req npu num and mode
func (tp *ascend310P) ValidNPUJob() *api.ValidateResult {
	var err error
	if tp == nil {
		err := errors.New(util.ArgumentError)
		return &api.ValidateResult{Pass: false, Reason: err.Error(), Message: err.Error()}
	}
	klog.V(util.LogDebugLev).Infof("%s ValidNPUJob job(%s).", tp.GetPluginName(), tp.Name)

	switch tp.Type {
	case util.JobTypeWhole:
		return tp.NPUHandler.ValidNPUJob()
	case util.JobTypeStCut:
	case util.JobTypeDyCut:
		return tp.validDyVNPUJob()
	default:
		err = fmt.Errorf("%s no type %d", tp.Name, tp.Type)
		klog.V(util.LogDebugLev).Infof("%s ValidNPUJob %s %s.", tp.GetPluginName(), tp.Name, err)
	}

	return &api.ValidateResult{Pass: false, Reason: err.Error(), Message: err.Error()}
}
