/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package comvnpu is using for virtual HuaWei Ascend910 schedule.

*/
package comvnpu

import (
	"errors"
	"k8s.io/klog"
	"strings"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/util"
)

// GetNPUTypeByResourceName get vJob vnpu source name, like huawei.com/Ascend310P-4c.
func (tp *VNPU) GetNPUTypeByResourceName(tmp string) (string, error) {
	split := strings.Split(tmp, "-")
	if len(split) == 1 {
		return tmp, nil
	}
	if len(split) != util.NPUIndex2 {
		klog.V(util.LogDebugLev).Infof("GetNPUTypeByResourceName get err: %v.", split)
		return "", errors.New("err resource")
	}
	klog.V(util.LogDebugLev).Infof("GetNPUTypeByResourceName get %v.", split)
	return split[0], nil
}

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
