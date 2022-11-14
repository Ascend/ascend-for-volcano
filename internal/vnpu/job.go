/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package vnpu is using for HuaWei Ascend pin fault rescheduling.

*/
package vnpu

import (
	"context"
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"volcano.sh/apis/pkg/apis/scheduling"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

func (vJob *VJob) initVJobStatus(divideKinds []string) error {
	if !vJob.IsJobValidVNPUJob(divideKinds) {
		return fmt.Errorf("vJob %s is not valid vnpu job, should not be put into cache", vJob.jobUID)
	}
	vJob.setVJobStatus(VJobStatusUnhandled)
	return nil
}

func (vJob *VJob) setAllocFlag(allocFlag bool) {
	vJob.allocFlag = allocFlag
}

// IsJobValidVNPUJob state check for unhandled job
func (vJob *VJob) IsJobValidVNPUJob(divideKinds []string) bool {
	if !vJob.isJobTaskNumValid() {
		return false
	}
	if !vJob.isJobResourceTypeValid(divideKinds) {
		return false
	}
	return true
}

func (vJob *VJob) isJobTaskNumValid() bool {
	return vJob.taskNum == 1
}

func (vJob *VJob) isJobResourceTypeValid(divideKinds []string) bool {
	flag := false
	for _, kind := range divideKinds {
		if kind == vJob.reqVNPUType {
			flag = true
			break
		}
	}
	if !flag {
		klog.V(util.LogErrorLev).Infof("IsVNPUJob %s %s not in %+v.", vJob.jobUID, vJob.reqVNPUType, divideKinds)
		return false
	}
	return flag
}

// IsJobReadyForPreAlloc state check for not unhandled job
func (vJob *VJob) IsJobReadyForPreAlloc(schedulerJob plugin.SchedulerJob) bool {
	if schedulerJob.PGStatus == scheduling.PodGroupInqueue {
		klog.V(util.LogInfoLev).Infof("%s IsVJobCanPreHandle Inqueue.", vJob.jobUID)
		return true
	}
	for _, podPhase := range schedulerJob.TaskStatus {
		if podPhase != v1.PodPending {
			klog.V(util.LogInfoLev).Infof("%s's task not pending %v", vJob.jobUID, podPhase)
			return false
		}
	}
	return true
}

func (vJob *VJob) isJobNew(vCache *VCache) bool {
	for jobUID, _ := range vCache.vJobs {
		if jobUID == vJob.jobUID {
			return true
		}
	}
	return false
}

func (vJob *VJob) isJobSetPreAllocFlag(vCache *VCache) bool {
	for jobUID, _ := range vCache.vJobs {
		if jobUID == vJob.jobUID {
			return vJob.allocFlag
		}
	}
	return false
}

// IsClusterResourceSufficient ensure cluster total resource is enough
func (vJob *VJob) IsClusterResourceSufficient(vNodes map[string]VNode) bool {
	VResCluster := vJob.getClusterUnsegmentedResource(vNodes)
	return VResCluster.beGreater(vJob.resourceReq)
}

func (vJob *VJob) getClusterUnsegmentedResource(vNodes map[string]VNode) VResource {
	var unsegmentedRes []VResource
	for _, vNode := range vNodes {
		unsegmentedRes = append(unsegmentedRes, vNode.nodeUnsegmentedRes)
	}
	return add(unsegmentedRes)
}

// IsJobAllocated state check for pre-allocated jobs
func (vJob *VJob) IsJobAllocated(schedulerJob plugin.SchedulerJob) bool {
	if schedulerJob.PGStatus != scheduling.PodGroupRunning {
		return false
	}
	for _, podPhase := range schedulerJob.TaskStatus {
		if podPhase != v1.PodRunning {
			return false
		}
	}
	return true
}

// IsJobNeedDeleting state check for allocated jobs
func (vJob *VJob) IsJobNeedDeleting(schedulerJob plugin.SchedulerJob, ssn *framework.Session) bool {
	if !vJob.isJobPodTerminated(schedulerJob, ssn) {
		return false
	}
	if !vJob.isJobOvertime() {
		return false
	}
	return true
}

func (vJob *VJob) isJobPodTerminated(schedulerJob plugin.SchedulerJob, ssn *framework.Session) bool {
	for _, podPhase := range schedulerJob.TaskStatus {
		if podPhase == v1.PodSucceeded || podPhase == v1.PodFailed {
			return true
		}
	}
	if !vJob.CheckJobExistsInKubernetes(ssn) {
		return true
	}
	return false
}

// CheckJobExistsInKubernetes check whether job recorded in cache can be traced in kubernetes
func (vJob *VJob) CheckJobExistsInKubernetes(ssn *framework.Session) bool {
	var existTaskNum int
	ssnJob, ok := ssn.Jobs[vJob.jobUID]
	if !ok {
		klog.V(util.LogWarningLev).Infof("vJob %s not in session", vJob.jobUID)
	}
	for _, ssnTask := range ssnJob.Tasks {
		klog.V(util.LogDebugLev).Infof("check task %s via client-go", ssnTask.Name)
		realPod, err := ssn.KubeClient().CoreV1().Pods(ssnTask.Namespace).Get(
			context.TODO(), ssnTask.Name, metav1.GetOptions{})
		if err != nil || realPod == nil {
			klog.V(util.LogInfoLev).Infof("pod %s not in kubernetes", ssnTask.Name)
			continue
		}
		existTaskNum += 1
		klog.V(util.LogDebugLev).Infof("task %s is in kubernetes", ssnTask.Name)
	}
	if existTaskNum > 0 {
		return true
	}
	return false
}

func (vJob *VJob) isJobOvertime() bool {
	now := time.Now().Unix()
	if now-vJob.updateTime > DeleteOverTime {
		klog.V(util.LogErrorLev).Infof("IsDeleteVJobOverTime %v==%v", now, vJob.updateTime)
		return true
	}
	return false
}

func (vJob *VJob) setVJobStatus(status string) {
	vJob.jobStatus = status
	vJob.updateTime = time.Now().Unix()
}

func (vJob *VJob) setAllocCardName(allocCardName string) {
	vJob.allocCardName = allocCardName
}

func (vJob *VJob) recordVJobPreSegmentInfo(vNode *VNode, vChip *VChip) {
	vJob.reqNodeName = vNode.nodeName
	vJob.reqCardName = vChip.cardName
	vJob.allocFlag = true
}

func (vJob *VJob) clearVJobPreSegmentInfo() {
	vJob.reqNodeName = ""
	vJob.reqCardName = ""
	vJob.allocFlag = false
	vJob.updateTime = time.Now().Unix()
}

func (vJob *VJob) setVJobAllocInfo(allocCardName string) {
	vJob.setAllocCardName(allocCardName)
	vJob.updateTime = time.Now().Unix()
}

// getVJobUsedNPUNames called after vJobs being allocated and pod annotation written
func (vJob *VJob) getVJobUsedNPUNames(ssnJob *api.JobInfo) (string, error) {
	var cards []string
	if vJob.jobUID != ssnJob.UID {
		return "", fmt.Errorf("vJob UID %s and schedulerJob UID %s do not match", vJob.jobUID, ssnJob.UID)
	}
	for _, schedulerTask := range ssnJob.Tasks {
		chips := getPodUsedNPUNames(schedulerTask, vJob.reqCardName)
		cards = append(cards, chips...)
	}
	if len(cards) == 0 {
		return "", fmt.Errorf("%s err cards %v", vJob.jobUID, cards)
	}
	return cards[0], nil
}

// isVJobInDeviceInfo todo function to be completed
func (vJob *VJob) isVJobInDeviceInfo(node plugin.NPUNode) bool {
	return false
}

func (vJob VJobList) Len() int {
	return len(vJob)
}

// Less for order.
func (vJob VJobList) Less(i, j int) bool {
	if i > vJob.Len() || j > vJob.Len() {
		return false
	}
	return vJob[i].createTime < vJob[j].createTime
}

// Swap for order.
func (vJob VJobList) Swap(i, j int) {
	if i > vJob.Len() || j > vJob.Len() {
		return
	}
	vJob[i], vJob[j] = vJob[j], vJob[i]
}
