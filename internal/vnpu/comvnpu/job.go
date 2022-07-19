/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package comvnpu is using for virtual HuaWei Ascend910 schedule.

*/
package comvnpu

import (
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/util"
)

// IsMyJob used for identify Vnpu job, need to be implemented by vNPU plugins
func (tp *VNPU) IsMyJob(vJob *api.JobInfo) error {
	klog.V(util.LogDebugLev).Infof("%s IsMyJob %s, no need", tp.Name(), vJob.Name)
	return nil
}

// ValidNPUJobFn check the compliance of the selector and resource request numbers
func (tp *VNPU) ValidNPUJobFn(job *api.JobInfo) *api.ValidateResult {
	klog.V(util.LogDebugLev).Infof("%s ValidNPUJobFn %s, no need", tp.Name(), job.Name)
	return nil
}
