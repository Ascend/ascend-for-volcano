/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package vnpu is using for HuaWei Ascend pin vnpu rescheduling.
*/
package vnpu

import (
	"context"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"strings"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// GetSegmentFailureTaskIDs get segmentation failed pod from pod event
func GetSegmentFailureTaskIDs(ssn *framework.Session, namespace string) []api.TaskID {
	var faultTIDs []api.TaskID
	events, err := ssn.KubeClient().CoreV1().Events(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil || len(events.Items) < 1 {
		klog.V(util.LogDebugLev).Infof("GetSegmentFailureTaskIDs get error or no event")
		return nil
	}

	for _, event := range events.Items {
		if !isEventSegmentFailurePod(event) {
			continue
		}

		faultPod := getPodFromKubernetes(ssn, event.InvolvedObject.Name, namespace)
		if faultPod == nil {
			continue
		}
		faultTIDs = append(faultTIDs, api.TaskID(faultPod.UID))
	}
	return faultTIDs
}

func isEventSegmentFailurePod(event v1.Event) bool {
	if event.InvolvedObject.Kind != podObjectType {
		klog.V(util.LogDebugLev).Infof("GetSegmentFailureTaskIDs %s not pod but %s, continue",
			event.InvolvedObject.Name, event.InvolvedObject.Kind)
		return false
	}

	if event.Type != PodEventTypeAllocateFailed || event.Reason != PodEventReasonAllocateFailed ||
		!strings.Contains(event.Message, PodEventMsgAllocateFailed) {
		klog.V(util.LogDebugLev).Infof("GetSegmentFailureTaskIDs pod event not segmentation error")
		return false
	}
	return true
}

func getPodFromKubernetes(ssn *framework.Session, name, namespace string) *v1.Pod {
	faultPod, err := ssn.KubeClient().CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		klog.V(util.LogDebugLev).Infof("GetSegmentFailureTaskIDs get pod<%s> from kubernetes failed", faultPod)
		return nil
	}
	klog.V(util.LogInfoLev).Infof("in getPodEvent pod %s segmentation fault event", name)
	return faultPod
}
