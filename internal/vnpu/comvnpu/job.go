/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package comvnpu is using for virtual HuaWei Ascend910 schedule.

*/
package comvnpu

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/vnpu/vnpuutil"
)

// ValidJobResource valid vJob resource,must used after init.
func (tp *VNPU) ValidJobResource(job *api.JobInfo) error {
	// Virtual npu resource count
	var vRC int
	reqReses := api.NewResource(*job.PodGroup.Spec.MinResources)
	for rType, jobNPU := range reqReses.ScalarResources {
		r := string(rType)
		if !strings.HasPrefix(r, tp.Attr.AnnoName) {
			continue
		}
		_, ok := tp.Attr.Coefficients[r]
		if !ok {
			msg := fmt.Errorf("%s request an invalid type of Vnpu resource", job.Name)
			klog.V(util.LogErrorLev).Infof("invalid request: %v.", msg)
			return msg
		}
		vRC += int(jobNPU / util.NPUHex)
	}
	klog.V(util.LogInfoLev).Infof("job(%s) requests %d Vnpu.", job.Name, vRC)

	if vRC > 1 {
		// a job can request only one Vnpu resource
		return fmt.Errorf("%s request invalid number %v", job.Name, vRC)
	}

	return nil
}

// GetNPUJobDefaultSelectorConfig get selectors for Vnpu
func (tp *VNPU) GetNPUJobDefaultSelectorConfig() map[string]string {
	var defaultSchedulerConfig map[string]string
	defaultSchedulerConfig = make(map[string]string, util.NPUIndex3)

	defaultSchedulerConfig[util.ArchSelector] = util.HuaweiArchArm + "|" + util.HuaweiArchX86

	return defaultSchedulerConfig
}

// for verify npu job must config selector
func (tp *VNPU) validNPUJobSelector(job *api.JobInfo) error {
	jobSelectors := util.GetJobSelectors(job)
	if len(jobSelectors) == 0 {
		msg := fmt.Errorf("%s getJobSelectors nil", job.Name)
		klog.V(util.LogErrorLev).Infof("%s.", msg.Error())
		return msg
	}

	klog.V(util.LogDebugLev).Infof("%s has selector: %v.", job.Name, jobSelectors)

	defaultSchedulerConfig := tp.GetNPUJobDefaultSelectorConfig()

	if err := util.CompareNPUSelector(job, jobSelectors, defaultSchedulerConfig); err != nil {
		klog.V(util.LogErrorLev).Infof("%v.", err)
		return err
	}

	return nil
}

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

// GetVJobReqNPUType get vJob req type.
func (tp *VNPU) GetVJobReqNPUType(job *api.JobInfo) (string, error) {
	tmp, getErr := util.GetReqResourceNameFromJob(job)
	if getErr != nil {
		klog.V(util.LogDebugLev).Infof("%s GetVJobReqNPUType %s %v.", tp.Name(), job.Name, getErr)
		return "", getErr
	}

	return tp.GetNPUTypeByResourceName(tmp)
}

// UpdateVJobsCacheAllocChipByJobName Update vJob allocChip in cache by job name.
func (tp *VNPU) UpdateVJobsCacheAllocChipByJobName(vJob *api.JobInfo) error {
	var cards []string
	tmp, getErr := util.GetReqResourceNameFromJob(vJob)
	if getErr != nil {
		klog.V(util.LogErrorLev).Infof("%s GetVJobReqNPUType %s %v.", tp.Name(), vJob.Name, getErr)
		return getErr
	}
	for _, vTask := range vJob.Tasks {
		chips := util.GetPodUsedNPUNames(vTask, tmp)
		cards = append(cards, chips...)
	}
	if len(cards) == 0 {
		return fmt.Errorf("%s err cards %v", vJob.UID, cards)
	}

	for k, data := range vnpuutil.VNPUAllocData.Cache {
		if data.JobUID == vJob.UID {
			vnpuutil.VNPUAllocData.Cache[k].AllocCardName = cards[0]
			vnpuutil.VNPUAllocData.Cache[k].UpdateTime = time.Now().Unix()
			vnpuutil.VNPUAllocData.CheckCode = util.MakeDataHash(vnpuutil.VNPUAllocData.Cache)
			klog.V(util.LogInfoLev).Infof("%s UpdateVJobsCacheAllocChipByJobName %s update into %v.",
				tp.Name(), vJob.UID, vnpuutil.VNPUAllocData.Cache)
			return nil
		}
	}
	return fmt.Errorf("not find %s in chache", vJob.UID)
}

// GetVNPUAllocInfFromCacheByJobInf Get vJob allocInf from cache by jobInf.
func (tp *VNPU) GetVNPUAllocInfFromCacheByJobInf(vJob *api.JobInfo) (*vnpuutil.VNPUAllocInf, error) {
	for _, data := range vnpuutil.VNPUAllocData.Cache {
		if data.JobUID == vJob.UID {
			return &data, nil
		}
	}
	return nil, fmt.Errorf("%s not in cache", vJob.Name)
}

// IsVJobHasPredistribution judge vJob whether has been pre-alloc.
func (tp *VNPU) IsVJobHasPredistribution(vJob *api.JobInfo) bool {
	data, getERR := tp.GetVNPUAllocInfFromCacheByJobInf(vJob)
	if getERR != nil {
		klog.V(util.LogErrorLev).Infof("%s IsVJobHasPredistribution: %v.", tp.Name(), getERR)
		return false
	}
	if !data.AllocFlag {
		return false
	}
	if data.AllocCardName == "" {
		return false
	}
	return true
}

// IsVJobOverWaitTime judge vJob whether has been wait over time, after pre-alloc.
func (tp *VNPU) IsVJobOverWaitTime(vJob *api.JobInfo) bool {
	data, getERR := tp.GetVNPUAllocInfFromCacheByJobInf(vJob)
	if getERR != nil {
		klog.V(util.LogErrorLev).Infof("%s IsVJobOverWaitTime: %v.", tp.Name(), getERR)
		return false
	}
	diffTime := time.Now().Unix() - data.UpdateTime
	if diffTime > vnpuutil.JobPendingWaitTime {
		klog.V(util.LogErrorLev).Infof("%s IsVJobOverWaitTime %s: %v==%v.", tp.Name(), vJob.Name,
			time.Now().Unix(), data.UpdateTime)
		return true
	}
	return false
}

// DeleteCacheVJobByInfo delete vJob from cache.
func (tp *VNPU) DeleteCacheVJobByInfo(vJob *api.JobInfo) error {
	var cache []vnpuutil.VNPUAllocInf
	for _, data := range vnpuutil.VNPUAllocData.Cache {
		if data.JobUID == vJob.UID {
			continue
		}
		tmp := data
		cache = append(cache, tmp)
	}
	vnpuutil.VNPUAllocData.Cache = cache
	vnpuutil.VNPUAllocData.CheckCode = util.MakeDataHash(vnpuutil.VNPUAllocData.Cache)
	klog.V(util.LogDebugLev).Infof("%s DeleteCacheVJobByInfo %v.", tp.Name(), vnpuutil.VNPUAllocData.Cache)
	return nil
}

// GetPluginNameByJobInfo get vPlugin name by jobInfo.
func (tp *VNPU) GetPluginNameByJobInfo(job *api.JobInfo) (string, error) {
	reqNpuType, typeErr := tp.GetVJobReqNPUType(job)
	if typeErr != nil {
		klog.V(util.LogDebugLev).Infof("%s GetPluginNameByJobInfo %s %v.", tp.Name(), job.Name, typeErr)
		return "", typeErr
	}

	var pluginName string
	var pluginErr error
	switch reqNpuType {
	case vnpuutil.NPU310PCardName:
		pluginName = vnpuutil.PluginNameBy310PVNPU
	case vnpuutil.NPU910CardName:
		pluginName = vnpuutil.PluginNameBy910VNPU
	default:
		pluginName = vnpuutil.UNKnownPluginName
		pluginErr = fmt.Errorf("%s resource %v not support", job.Name, reqNpuType)
		klog.V(util.LogErrorLev).Infof("%s GetPluginNameByJobInfo %v.", tp.Name(), pluginErr)
	}
	return pluginName, pluginErr
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
