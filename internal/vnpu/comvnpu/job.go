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
	"sort"
	"strings"
	"time"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

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

func (tp *VNPU) isJobSetPreAllocFlag(job *api.JobInfo) bool {
	for _, v := range vnpuutil.VNPUAllocData.Cache {
		if v.JobUID == job.UID {
			return v.AllocFlag
		}
	}
	return false
}

func (tp *VNPU) isNewVNPUJob(job *api.JobInfo) bool {
	for _, v := range vnpuutil.VNPUAllocData.Cache {
		if v.JobUID == job.UID {
			return false
		}
	}
	return true
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

func (tp *VNPU) getVJobReqInfFromJobInfo(job *api.JobInfo) (*vnpuutil.VNPUAllocInf, error) {
	reqNpuName, typeErr := util.GetReqResourceNameFromJob(job)
	if typeErr != nil {
		klog.V(util.LogErrorLev).Infof("%s getVJobReqInfFromJobInfo %s %v.", tp.Name(), job.Name, typeErr)
		return nil, typeErr
	}
	var tmp = vnpuutil.VNPUAllocInf{
		JobUID:        job.UID,
		ReqNPUType:    reqNpuName,
		NodeName:      "",
		ReqCardName:   "",
		AllocCardName: "",
		AllocFlag:     false,
		UpdateTime:    time.Now().Unix(),
	}
	klog.V(util.LogErrorLev).Infof("%s getVJobReqInfFromJobInfo %s %+v.", tp.Name(), job.Name, tmp)
	return &tmp, nil
}

// CheckJobNeedPreAlloc Check the vJob whether need do pre-Alloc or not.
func (tp *VNPU) CheckJobNeedPreAlloc(job *api.JobInfo) error {
	if tp.isNewVNPUJob(job) {
		klog.V(util.LogDebugLev).Infof("%s CheckJobNeedPreAlloc new %s need to preAlloc.", tp.Name(), job.Name)
		return nil
	}
	if tp.isJobSetPreAllocFlag(job) {
		preErr := fmt.Errorf("%s has been preAlloc", job.Name)
		klog.V(util.LogDebugLev).Infof("%s isJobSetPreAllocFlag %v.", tp.Name(), preErr)
		return preErr
	}
	klog.V(util.LogDebugLev).Infof("%s CheckJobNeedPreAlloc %s need to preAlloc.", tp.Name(), job.Name)
	// over time is deal in valid job
	return nil
}

// RecordNewVNPUJobInCache deal new VNPU job from session.
func (tp *VNPU) RecordNewVNPUJobInCache(job *api.JobInfo) error {
	if checkErr := tp.CheckJobNeedPreAlloc(job); checkErr != nil {
		return checkErr
	}

	vNPUAllocInf, allocErr := tp.getVJobReqInfFromJobInfo(job)
	if allocErr != nil {
		return allocErr
	}

	if err := tp.AddOrUpdateVNPUAllocInfIntoCache(vNPUAllocInf); err != nil {
		return err
	}
	return nil
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

// DealVJobLegality IsJobInitial has been called before.
func (tp *VNPU) DealVJobLegality(vJob *api.JobInfo) error {
	// 1.Only unallocated VJob can do these.
	// 2.whether the job has predistribution flag
	if tp.IsVJobHasPredistribution(vJob) {
		klog.V(util.LogDebugLev).Infof("%s DealVJobLegality %s not pre-distribution.", tp.Name(), vJob.UID)
		return nil
	}
	if util.IsJobRunningByInfo(vJob) {
		klog.V(util.LogDebugLev).Infof("%s DealVJobLegality %s has running.", tp.Name(), vJob.UID)
		return nil
	}
	// 3.check the job whether over the max wait time.
	if !tp.IsVJobOverWaitTime(vJob) {
		return nil
	}
	// 4.delete vJob if over time.
	return tp.DeleteCacheVJobByInfo(vJob)
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

// IsVNPUJob judge the job is vJob or not.
func (tp *VNPU) IsVNPUJob(job *api.JobInfo) bool {
	// 1.init vnp
	pluginName, nameErr := tp.GetPluginNameByJobInfo(job)
	if nameErr != nil {
		klog.V(util.LogDebugLev).Infof("%s IsVNPUJob %s %v.", tp.Name(), job.Name, nameErr)
		return false
	}
	if pluginErr := tp.InitVNPUPluginByType(pluginName); pluginErr != nil {
		klog.V(util.LogErrorLev).Infof("%s IsVNPUJob :%v.", vnpuutil.PluginName, pluginErr)
		return false
	}

	reqNpuType, getErr := util.GetReqResourceNameFromJob(job)
	if getErr != nil {
		klog.V(util.LogErrorLev).Infof("%s IsVNPUJob %s %v.", tp.Name(), job.Name, getErr)
		return false
	}
	// 2.vnp job.
	flag := false
	for _, kind := range tp.Attr.DivideKinds {
		if kind == reqNpuType {
			flag = true
			break
		}
	}
	if !flag {
		klog.V(util.LogErrorLev).Infof("%s IsVNPUJob %s %s not in %+v.", tp.Name(), job.Name, reqNpuType,
			tp.Attr.DivideKinds)
		return false
	}
	// 3.valid vJob require vNPU Number.
	num, numErr := util.GetJobReqResourceNumFromJobPG(job, reqNpuType)
	if numErr != nil {
		klog.V(util.LogErrorLev).Infof("%s IsVNPUJob %s %+v.", tp.Name(), job.Name, numErr)
		return false
	}
	if num != 1 {
		klog.V(util.LogErrorLev).Infof("%s IsVNPUJob %s req %+v==%d.", tp.Name(), job.Name, reqNpuType, num)
		return false
	}
	tp.setVPUPluginToVNPUBack()
	return true
}

type vJobsList []*api.JobInfo

// Len for order.
func (vJob vJobsList) Len() int {
	return len(vJob)
}

// Less for order.
func (vJob vJobsList) Less(i, j int) bool {
	if i > vJob.Len() || j > vJob.Len() {
		return false
	}
	return vJob[i].CreationTimestamp.Unix() < vJob[j].CreationTimestamp.Unix()
}

// Swap for order.
func (vJob vJobsList) Swap(i, j int) {
	if i > vJob.Len() || j > vJob.Len() {
		return
	}
	vJob[i], vJob[j] = vJob[j], vJob[i]
}

// OrderVJobsByCreateTime Order the jobs by create time(all kinds vJobs are in one list).
func (tp *VNPU) OrderVJobsByCreateTime(jobs []*api.JobInfo) ([]*api.JobInfo, error) {
	tempVJobs := vJobsList(jobs)
	sort.Sort(tempVJobs)
	return tempVJobs, nil
}

// GetVJobNeedVNPU return VNPU name for number is 1.
func (tp *VNPU) GetVJobNeedVNPU(vJob *api.JobInfo) (string, error) {
	if tp == nil {
		return "", errors.New(vnpuutil.PluginUninitializedError)
	}

	reqReses := api.NewResource(*vJob.PodGroup.Spec.MinResources)
	for value := range reqReses.ScalarResources {
		tmp := string(value)
		if strings.Contains(tmp, tp.Attr.AnnoName) {
			// the VNPU func before has check the resource type.
			return tmp, nil
		}
	}
	klog.V(util.LogErrorLev).Infof("%s GetVJobNeedVNPU %s %#v has no %v.", tp.Name(),
		vJob.Name, vJob.TotalRequest.ScalarResources, tp.Attr.AnnoName)
	return "", fmt.Errorf("%s not has NPUS", vJob.Name)
}

// IsVJobReqNPUMeetTotalResources npu card only has one kind and one.
func (tp *VNPU) IsVJobReqNPUMeetTotalResources(npu string, res map[string]int) bool {
	if len(res) == 0 || npu == "" {
		klog.V(util.LogErrorLev).Info("IsVJobReqNPUMeetTotalResources parameters error.")
		return false
	}
	chipCore, coverErr := tp.coverReqNPUTypeToCoreNum(npu)
	if coverErr != nil {
		klog.V(util.LogErrorLev).Infof("%s IsVNPUNodeMeetReqResource %v.", tp.Name(), coverErr)
		return false
	}
	for _, value := range res {
		if value >= chipCore {
			return true
		}
	}
	return false
}

// GetVJobMeetNodeList Get vJob meet nodeMap
func (tp *VNPU) GetVJobMeetNodeList(vJob *api.JobInfo, res map[string]int,
	ssn *framework.Session) ([]*api.NodeInfo, error) {
	// 1.Get vJob need VNPU.
	jobNeedNPUType, getErr := tp.GetVJobNeedVNPU(vJob)
	if getErr != nil {
		klog.V(util.LogErrorLev).Infof("%s GetVJobNeedVNPU %v.", tp.Name(), getErr)
		return nil, getErr
	}
	// 2. check the cluster total res meet VJob require.
	if !tp.IsVJobReqNPUMeetTotalResources(jobNeedNPUType, res) {
		err := fmt.Errorf("total resource %+v not meet req %s", res, jobNeedNPUType)
		klog.V(util.LogErrorLev).Infof("%s GetVJobMeetNodeList %s %v.", tp.Name(), vJob.Name, err)
		return nil, err
	}
	// 3.Get node list by req VNPU.
	nodeList, listErr := tp.GetNodeListByReqVNPU(jobNeedNPUType, ssn)
	if listErr != nil {
		klog.V(util.LogErrorLev).Infof("%s GetNodeListByReqVNPU %v.", tp.Name(), listErr)
		return nil, listErr
	}
	if len(nodeList) == 0 {
		return nil, errors.New("none node meet")
	}
	return nodeList, nil
}

// IsMyJob used for identify Vnpu job, need to be implemented by vNPU plugins
func (tp *VNPU) IsMyJob(vJob *api.JobInfo) error {
	if tp.IsVNPUJob(vJob) {
		return nil
	}
	return fmt.Errorf("%s not VNPU job", vJob.Name)
}

// ValidNPUJobFn check the compliance of the selector and resource request numbers
func (tp *VNPU) ValidNPUJobFn(job *api.JobInfo) *api.ValidateResult {
	// 1.valid npu job selector
	if err := tp.validNPUJobSelector(job); err != nil {
		klog.V(util.LogErrorLev).Infof("%s err: %v.", tp.Name(), err)
		return &api.ValidateResult{
			Pass:    false,
			Reason:  err.Error(),
			Message: fmt.Sprintf("validNPUJob err: %v", err),
		}
	}
	// 2.valid the resource type and the number of resources the job request
	if errRs := tp.ValidJobResource(job); errRs != nil {
		klog.V(util.LogErrorLev).Infof("%s err: %v.", tp.Name(), errRs)
		return &api.ValidateResult{
			Pass:    false,
			Reason:  "job resource requested error",
			Message: fmt.Sprintf("%s, err: %v", job.Name, errRs),
		}
	}
	// 3.Valid the vJob's legality
	if checkErr := tp.DealVJobLegality(job); checkErr != nil {
		klog.V(util.LogErrorLev).Infof("%s DealVJobLegality: %v.", tp.Name(), checkErr)
		return &api.ValidateResult{
			Pass:    false,
			Reason:  "vJob legality error",
			Message: fmt.Sprintf("%s, err: %v", job.Name, checkErr),
		}
	}
	return nil
}

// GetVNPUAllocInfFromCacheByJobID Get VNPU allocInf from cache by using JobID.
func (tp *VNPU) GetVNPUAllocInfFromCacheByJobID(jobUID api.JobID) (*vnpuutil.VNPUAllocInf, error) {
	for _, data := range vnpuutil.VNPUAllocData.Cache {
		if data.JobUID == jobUID {
			return &data, nil
		}
	}
	return nil, fmt.Errorf("%s not in cache", jobUID)
}
