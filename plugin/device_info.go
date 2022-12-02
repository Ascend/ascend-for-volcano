/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package plugin is using for HuaWei Ascend pin affinity schedule.

*/
package plugin

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

const (
	// PodEventMsgAllocateFailed dp
	PodEventMsgAllocateFailed = "Allocate failed due to rpc error: code = Unknown desc = NoNPUAffinity, " +
		"which is unexpected"
	// PodEventReasonAllocateFailed dp
	PodEventReasonAllocateFailed = "UnexpectedAdmissionError"
	// PodEventTypeAllocateFailed dp
	PodEventTypeAllocateFailed = "Warning"

	// AscendVNPULevel vnpu level
	AscendVNPULevel = "vnpu-level"
	// AscendVNPULevelLow low
	AscendVNPULevelLow = "low"
	// AscendVNPULevelHigh high
	AscendVNPULevelHigh = "high"
	// AscendVNPUPrefix vir
	AscendVNPUPrefix = "vir"
	// AscendVNPUDVPP dvpp enable
	AscendVNPUDVPP = "vnpu-dvpp"
	// AscendDVPPEnabledOff off
	AscendDVPPEnabledOff = "no"
	// AscendDVPPEnabledNull null
	AscendDVPPEnabledNull = "null"
	// AscendDVPPEnabledOn on
	AscendDVPPEnabledOn = "yes"
	// AscendNDVPPValue value
	AscendNDVPPValue = "ndvpp"
	// AscendDVPPValue value
	AscendDVPPValue = "dvpp"
	// RingController ascend-310P/ascend-310/ascend-910
	RingController = "ring-controller.atlas"
	podObjectType  = "Pod"

	// Ascend310P 310P template name
	Ascend310P = "Ascend310P"
	// Ascend910 910 template name
	Ascend910 = "Ascend910"
	// RingController310P 310p ring controller name
	RingController310P = "ascend-310P"
)

// GetResourceFromStr vDeviceResourceStr like 4c.3cpu.ndvpp
func GetResourceFromStr(vDeviceResourceStr string) *util.VResource {
	klog.V(util.LogInfoLev).Infof("GetResourceFromStr parsing ")
	resources := strings.Split(vDeviceResourceStr, ".") // like 4c.3cpu.ndvpp/2c.1cpu/8c
	if len(resources) < 1 {
		klog.V(util.LogErrorLev).Infof("%s: resource %s format incorrect", util.AscendNPUPodRealUse, resources)
		return nil
	}

	// 1. get coreNum from template
	aicoreNum, err := strconv.Atoi(strings.TrimSuffix(resources[0], "c")) // like 4c
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s: aicore %s format incorrect", util.AscendNPUPodRealUse, resources[0])
		return nil
	}

	// 2. get aicpu from template
	aicpuNum := getAicpuFromTemplate(resources, aicoreNum)
	if aicpuNum == util.ErrorInt {
		klog.V(util.LogDebugLev).Infof("%s aicpu format error", vDeviceResourceStr)
		return nil
	}

	if len(resources) < util.NPUIndex3 {
		return &util.VResource{
			Aicore: aicoreNum,
			Aicpu:  aicpuNum,
			DVPP:   AscendDVPPEnabledNull,
		}
	}

	// 3. get cvpp from template
	var dvppValue string
	switch resources[util.NPUIndex2] {
	case AscendDVPPValue:
		dvppValue = AscendDVPPEnabledOn
	case AscendNDVPPValue:
		dvppValue = AscendDVPPEnabledOff
	default:
		dvppValue = AscendDVPPEnabledNull
	}
	return &util.VResource{
		Aicore: aicoreNum,
		Aicpu:  aicpuNum,
		DVPP:   dvppValue,
	}
}

func getAicpuFromTemplate(resources []string, aicoreNum int) int {
	if len(resources) < util.NPUIndex2 { // 2.1 cpu==core
		klog.V(util.LogDebugLev).Infof("high cpu requirements")
		return aicoreNum
	}
	aicpuNum, err := strconv.Atoi(strings.TrimSuffix(resources[1], util.AICPU)) // 2.2 cpu<core
	if err != nil {
		klog.V(util.LogDebugLev).Infof("aicpu format error")
		return util.ErrorInt
	}

	return aicpuNum
}

// IsPodWholeCard judge if card is whole card Ascend910-0/Ascend910-4c-100-1-1
func IsPodWholeCard(realCardName string) bool {
	temp := strings.Split(realCardName, ",")
	for _, cardName := range temp {
		singleCardTemp := strings.Split(cardName, "-")
		if len(singleCardTemp) != util.NPUIndex2 {
			return false
		}
	}
	return true
}

// GetCardPhysicsID get card id
func GetCardPhysicsID(cardNameStr string, isWholeCard bool) ([]int, error) {
	var physicsIDs []int
	// deal vnpu card
	if !isWholeCard {
		cardNameSplit := strings.Split(cardNameStr, "-")
		if len(cardNameSplit) != util.NPUIndex5 {
			return physicsIDs, fmt.Errorf("getCardPhysicsID vnpu real device <%s> format error", cardNameStr)
		}
		phyCardID, err := strconv.Atoi(cardNameSplit[util.NPUIndex3])
		if err != nil {
			return physicsIDs, fmt.Errorf("getCardPhysicsID vnpu device<%s> get physics id failed", cardNameStr)
		}
		physicsIDs = append(physicsIDs, phyCardID)
		return physicsIDs, nil
	}
	// deal whole card
	cardNameSplit := strings.Split(cardNameStr, ",")
	for _, singleCardStr := range cardNameSplit {
		phyCardID, err := GetWholeCardIDFromCardNameStr(singleCardStr)
		if err != nil {
			return physicsIDs, fmt.Errorf("getCardPhysicsID whole device<%s> get id failed", singleCardStr)
		}
		physicsIDs = append(physicsIDs, phyCardID)
	}
	return physicsIDs, nil
}

// GetWholeCardIDFromCardNameStr get card physics id from Ascend910-0
func GetWholeCardIDFromCardNameStr(cardNameStr string) (int, error) {
	idStr := strings.Split(cardNameStr, "-")
	if len(idStr) < util.NPUIndex2 {
		return util.ErrorInt, fmt.Errorf("getCardIDFromCardNameStr %s format incorrect", cardNameStr)
	}
	id, err := strconv.Atoi(idStr[util.NPUIndex1])
	if err != nil {
		return util.ErrorInt, fmt.Errorf("getCardIDFromCardNameStr %s %v", cardNameStr, err)
	}
	return id, nil
}

// TransferTaskLabelToResReq transfer 4c.3cpu.ndvpp to resource
func TransferTaskLabelToResReq(task *api.TaskInfo) (util.VResource, error) {
	resReq := util.VResource{
		DVPP: AscendDVPPEnabledNull,
	}
	coreNum, err := getAiCoreNumFromTask(task)
	if err != nil {
		return resReq, fmt.Errorf("task %s AscendNPUCore read failed", task.Name)
	}

	cpuNum, err := getAiCpuNum(task, coreNum)
	if err != nil {
		return resReq, fmt.Errorf("task %s AscendCPUNum get failed", task.Name)
	}

	dvppVal := getDVPPEnable(task)

	resReq = util.VResource{
		Aicore: coreNum,
		Aicpu:  cpuNum,
		DVPP:   dvppVal,
	}

	return resReq, nil
}

func getAiCoreNumFromTask(task *api.TaskInfo) (int, error) {
	for _, container := range task.Pod.Spec.Containers {
		coreNum, ok := container.Resources.Requests[util.AscendNPUCore]
		if !ok {
			return 0, errors.New("getAiCoreNumFromTask get resource requests failed")
		}
		return int(coreNum.Value()), nil
	}
	return 0, fmt.Errorf("getAiCoreNumFromTask get resource requests failed")
}

func getDVPPEnable(task *api.TaskInfo) string {
	dvppVal, ok := task.Pod.Labels[AscendVNPUDVPP]
	if !ok {
		return AscendDVPPEnabledNull
	}
	return dvppVal
}

func getAiCpuNum(task *api.TaskInfo, coreNum int) (int, error) {
	ringControllerType, ok := task.Pod.Labels[RingController] // ascend-910/ascend-310P/ascend-310
	if !ok {
		return 0, fmt.Errorf("getAiCpuNum get label %s failed", RingController)
	}
	vnpuLevel, ok := task.Pod.Labels[AscendVNPULevel]
	if !ok {
		vnpuLevel = AscendVNPULevelLow
	}
	if ringControllerType == RingController310P && (coreNum == util.NPUIndex2 || coreNum == util.
		NPUIndex4) && vnpuLevel == AscendVNPULevelLow {
		return coreNum - 1, nil // 310P with 4core/2core low
	}
	return coreNum, nil
}

// GetSegmentFailurePod get segmentation failed pod from pod event
func GetSegmentFailurePod(ssn *framework.Session, namespace string) []*v1.Pod {
	var faultPods []*v1.Pod
	events, err := ssn.KubeClient().CoreV1().Events(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil || len(events.Items) < 1 {
		klog.V(util.LogDebugLev).Infof("GetSegmentFailurePod get error or no event")
		return faultPods
	}

	for _, event := range events.Items {
		if !isEventSegmentFailurePod(event) {
			continue
		}

		faultPod := getPodFromKubernetes(ssn, event.InvolvedObject.Name, namespace)
		if faultPod == nil {
			continue
		}

		faultPods = append(faultPods, faultPod)
	}
	return faultPods
}

func isEventSegmentFailurePod(event v1.Event) bool {
	if event.InvolvedObject.Kind != podObjectType {
		klog.V(util.LogDebugLev).Infof("GetSegmentFailurePod %s not pod but %s, continue",
			event.InvolvedObject.Name, event.InvolvedObject.Kind)
		return false
	}

	if event.Type != PodEventTypeAllocateFailed || event.Reason != PodEventReasonAllocateFailed ||
		event.Message != PodEventMsgAllocateFailed {
		klog.V(util.LogDebugLev).Infof("GetSegmentFailurePod pod event not segmentation error")
		return false
	}
	return true
}

func getPodFromKubernetes(ssn *framework.Session, name, namespace string) *v1.Pod {
	faultPod, err := ssn.KubeClient().CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		klog.V(util.LogDebugLev).Infof("GetSegmentFailurePod get pod<%s> from kubernetes failed", faultPod)
		return nil
	}
	klog.V(util.LogInfoLev).Infof("in getPodEvent pod %s segmentation fault event", name)
	return faultPod
}
