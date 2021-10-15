/*
Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/*

Package rescheduling is using for HuaWei Ascend pin fault rescheduling.

*/
package rescheduling

import (
	"context"
	"encoding/json"
	"fmt"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"reflect"
	"strings"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"
)

func getConfigMapWithRetry(client kubernetes.Interface, namespace, cmName string) (*v1.ConfigMap, error) {
	var cm *v1.ConfigMap
	var err error

	// There can be no delay or blocking operations in a session.
	if cm, err = client.CoreV1().ConfigMaps(namespace).Get(context.TODO(), cmName, metav1.GetOptions{}); err != nil {
		return nil, err
	}

	return cm, nil
}

func isConfigMapChanged(k8s kubernetes.Interface, cm *v1.ConfigMap) bool {
	cmData, getErr := getConfigMapWithRetry(k8s, cmNameSpace, cmName)
	if getErr != nil {
		return true
	}
	if reflect.DeepEqual(cmData, cm) {
		return false
	}

	return true
}

func createOrUpdateConfigMap(k8s kubernetes.Interface, cm *v1.ConfigMap) error {
	_, cErr := k8s.CoreV1().ConfigMaps(cm.ObjectMeta.Namespace).Create(context.TODO(), cm, metav1.CreateOptions{})
	if cErr != nil {
		if !apierrors.IsAlreadyExists(cErr) {
			return fmt.Errorf("unable to create ConfigMap:%v", cErr)
		}

		// To reduce the cm write operations
		if !isConfigMapChanged(k8s, cm) {
			klog.V(logInfoLev).Infof("configMap not changed,no need update")
			return nil
		}

		_, err := k8s.CoreV1().ConfigMaps(cm.ObjectMeta.Namespace).Update(context.TODO(), cm, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("unable to update ConfigMap:%v", err)
		}
	}
	return nil
}

func convertToReSchedulerJobsMapFromCM(buffer string) (map[string]ReSchedulerTasks, error) {
	reSchedulerJob := map[string]ReSchedulerTasks{}
	if unmarshalErr := json.Unmarshal([]byte(buffer), &reSchedulerJob); unmarshalErr != nil {
		klog.V(logErrorLev).Infof("convertToReSchedulerJobsMapFromCM: %v %v.", buffer, unmarshalErr)
		return nil, unmarshalErr
	}
	return reSchedulerJob, nil
}

func updateReSchedulerJobsFromCM(buffer string) error {
	reSchedulerJob, covErr := convertToReSchedulerJobsMapFromCM(buffer)
	if covErr != nil {
		klog.V(logErrorLev).Infof("convertToReSchedulerNodesMap: %v.", covErr)
		return covErr
	}
	var jobData = make(map[api.JobID]ReSchedulerTasks, 1)
	for dataID, data := range reSchedulerJob {
		jobIDStr := strings.Replace(dataID, "_", "/", -1)
		if strings.Count(jobIDStr, "/") != 1 {
			countErr := fmt.Errorf("%s more than one character '_'", dataID)
			return countErr
		}

		jobData[api.JobID(jobIDStr)] = data
	}

	ReSchedulerCache[CmJobKind] = jobData
	return nil
}

func updateFaultNodeFromCM(tmpData string) error {
	faultNode, covErr := convertToReSchedulerNodesMapFromCM(tmpData)
	if covErr != nil {
		klog.V(logErrorLev).Infof("convertToReSchedulerNodesMapFromCM: %v.", covErr)
		return covErr
	}

	ReSchedulerCache[CmNodeKind] = faultNode

	return nil
}

func updateFaultCardFromCM(tmpData string) error {
	faultCars, covErr := convertToReSchedulerCardsMapFromCM(tmpData)
	if covErr != nil {
		klog.V(logErrorLev).Infof("updateFaultCardFromCM: %v.", covErr)
		return covErr
	}

	ReSchedulerCache[CmCardKind] = faultCars

	return nil
}

func updateNodeHeartbeatFromCM(tmpData string) error {
	heartbeat, covErr := convertToNodeHeartbeatMapFromCM(tmpData)
	if covErr != nil {
		klog.V(logErrorLev).Infof("convertToNodeHeartbeatMapFromCM: %v.", covErr)
		return covErr
	}

	ReSchedulerCache[CmNodeHeartbeatKind] = heartbeat

	return nil
}

func updateReSchedulerData(cmData *v1.ConfigMap) error {
	if len(cmData.Data) == 0 {
		klog.V(logDebugLev).Infof("updateFaultNodeFromCM cmData is nil, nothing to do.")
		return nil
	}

	for dataID, buffer := range cmData.Data {
		if len(buffer) == 0 {
			klog.V(logDebugLev).Infof("updateFaultNodeFromCM %v is nil.", dataID)
			continue
		}

		switch dataID {
		case CmJobKind:
			if err := updateReSchedulerJobsFromCM(buffer); err != nil {
				klog.V(logErrorLev).Infof("updateReSchedulerJobsFromCM %v.", err)
			}
		case CmNodeKind:
			if err := updateFaultNodeFromCM(buffer); err != nil {
				klog.V(logErrorLev).Infof("updateFaultNodeFromCM %v.", err)
			}
		case CmCardKind:
			if err := updateFaultCardFromCM(buffer); err != nil {
				klog.V(logErrorLev).Infof("updateFaultCardFromCM: %v.", err)
			}
		case CmNodeHeartbeatKind:
			if err := updateNodeHeartbeatFromCM(buffer); err != nil {
				klog.V(logErrorLev).Infof("updateNodeHeartbeatFromCM: %v.", err)
			}
		default:
			klog.V(logErrorLev).Infof("updateReSchedulerData no support type:%v", dataID)
		}
	}

	return nil
}

func updateFaultNPUInfFromCM(ssn *framework.Session) error {
	cmData, err := getConfigMapWithRetry(ssn.KubeClient(), cmNameSpace, cmName)
	if err != nil {
		klog.V(logErrorLev).Infof("getConfigMapWithRetry :%v.", err)
		return err
	}

	updateErr := updateReSchedulerData(cmData)
	if updateErr != nil {
		klog.V(logErrorLev).Infof("updateReSchedulerData :%v.", updateErr)
		return updateErr
	}

	klog.V(logDebugLev).Infof("updateFaultNPUInfFromCM success.")
	return nil
}

// ReadFaultNPUJobsFromCM read from ConfigMap FaultNPUJobs and update to the cache.
func ReadFaultNPUJobsFromCM(ssn *framework.Session) error {
	var cmErr error
	// Get configMap failure does not affect the update cache.
	if cmErr = updateFaultNPUInfFromCM(ssn); cmErr != nil {
		klog.V(logErrorLev).Infof("updateFaultNPUInfFromCM :%v.", cmErr)
	}

	if err := updateReSchedulerDataFromSession(ssn); err != nil {
		klog.V(logErrorLev).Infof("updateReSchedulerDataFromSession :%v.", err)
		return fmt.Errorf("%v %v", cmErr, err)
	}

	return cmErr
}

func getCMWriteDate(ssn *framework.Session, reSchedulerData map[string]interface{}) map[string]string {
	var data = make(map[string]string, 1)
	for dataKind, faultData := range reSchedulerData {
		switch dataKind {
		case CmJobKind:
			jobData, err := getCMJobWriteData(ssn, faultData)
			if err != nil {
				klog.V(logErrorLev).Infof("getCMJobWriteData :%v.", err)
				continue
			}
			data[CmJobKind] = data[CmJobKind] + jobData
		case CmNodeKind:
			nodeData, err := getCMNodeWriteData(faultData)
			if err != nil {
				klog.V(logErrorLev).Infof("getCMJobWriteData :%v.", err)
				continue
			}
			data[CmNodeKind] = data[CmNodeKind] + nodeData
		case CmCardKind:
			cardData, err := getCMCardWriteData(faultData)
			if err != nil {
				klog.V(logDebugLev).Infof("getCMCardWriteData :%v.", err)
				continue
			}
			data[CmCardKind] = data[CmCardKind] + cardData
		case CmNodeHeartbeatKind:
			heartbeatData, err := getCMHeartbeatWriteData(faultData)
			if err != nil {
				klog.V(logDebugLev).Infof("getCMHeartbeatWriteData :%v.", err)
				continue
			}
			data[CmNodeHeartbeatKind] = data[CmNodeHeartbeatKind] + heartbeatData
		default:
			klog.V(logErrorLev).Infof("getCMWriteDate not support %T.", dataKind)
		}
	}

	return data
}

// WriteReSchedulerDataToCM Write FaultNPUJobs into ConfigMap.
func WriteReSchedulerDataToCM(ssn *framework.Session, reSchedulerData map[string]interface{}) error {
	var faultNPUConfigMap = &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cmName,
			Namespace: cmNameSpace,
		},
		Data: getCMWriteDate(ssn, reSchedulerData),
	}

	klog.V(logDebugLev).Infof("Write faultNPUJobs into cm: %+v.", faultNPUConfigMap)
	if err := createOrUpdateConfigMap(ssn.KubeClient(), faultNPUConfigMap); err != nil {
		klog.V(logErrorLev).Infof("createOrUpdateConfigMap : %v.", err)
		return err
	}

	return nil
}
