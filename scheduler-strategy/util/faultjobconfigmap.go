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

Package util is using for HuaWei Ascend9 pin affinity schedule utilities and processing configmap files.

*/
package util

import (
	"context"
	"encoding/json"
	"fmt"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	clientsetretry "k8s.io/client-go/util/retry"
	"k8s.io/klog"
	"reflect"
	"strings"
	time2 "time"
	"volcano.sh/apis/pkg/apis/scheduling"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"
)

func getConfigMapWithRetry(client kubernetes.Interface, namespace, cmName string) (*v1.ConfigMap, error) {
	var cm *v1.ConfigMap
	var lastError error
	err := wait.ExponentialBackoff(wait.Backoff(clientsetretry.DefaultBackoff), func() (bool, error) {
		var err error
		cm, err = client.CoreV1().ConfigMaps(namespace).Get(context.TODO(), cmName, metav1.GetOptions{})
		if err == nil {
			return true, nil
		}
		lastError = err
		return false, nil
	})
	if err == nil {
		return cm, nil
	}
	return nil, lastError
}

func isConfigMapNeedsUpdate(k8s kubernetes.Interface, cm *v1.ConfigMap) bool {
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
		if !isConfigMapNeedsUpdate(k8s, cm) {
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

func updateReSchedulerJobs(ssn *framework.Session) error {
	klog.V(logDebugLev).Infof("updateReSchedulerJobs get buffer %v.", ReSchedulerJobs)
	for jobID, tmpValue := range ReSchedulerJobs {
		// No job
		job, ok := ssn.Jobs[jobID]
		if !ok {
			klog.V(logErrorLev).Infof("delete %s from configMap due to not existence.", jobID)
			delete(ReSchedulerJobs, jobID)
			continue
		}
		// For job running
		if job.PodGroup.Status.Phase == scheduling.PodGroupRunning {
			klog.V(logErrorLev).Infof("delete %s from configMap due to job is ok.", jobID)
			delete(ReSchedulerJobs, jobID)
			continue
		}
		// For Node doesn't last too long
		for _, preTime := range tmpValue.Time {
			nowTime := time2.Now().Unix()
			if nowTime-preTime > maxIntervalTime {
				klog.V(logErrorLev).Infof("delete %s from CM for overTime %v => %v.", jobID, nowTime, preTime)
				delete(ReSchedulerJobs, jobID)
			}
		}
	}
	return nil
}

func getFaultNPUJobsFromCM(ssn *framework.Session) error {
	cmData, err := getConfigMapWithRetry(ssn.KubeClient(), cmNameSpace, cmName)
	if err != nil {
		klog.V(logErrorLev).Infof("ReadFaultNPUJobsFromCM :%v.", err)
		return err
	}

	for jobUUID, buffer := range cmData.Data {
		jobID := strings.Replace(jobUUID, "_", "/", -1)
		if strings.Count(jobID, "/") != constIntNum1 {
			countErr := fmt.Errorf("%s more than one character '_'", jobUUID)
			return countErr
		}

		tmp := ReSchedulerTasks{}
		if unmarshalErr := json.Unmarshal([]byte(buffer), &tmp); unmarshalErr != nil {
			return unmarshalErr
		}
		ReSchedulerJobs[api.JobID(jobID)] = tmp
	}

	return nil
}

// ReadFaultNPUJobsFromCM read from ConfigMap FaultNPUJobs and update to the cache.
func ReadFaultNPUJobsFromCM(ssn *framework.Session) error {
	var cmErr error
	// Get configMap failure does not affect the update cache.
	if cmErr := getFaultNPUJobsFromCM(ssn); cmErr != nil {
		klog.V(logErrorLev).Infof("getFaultNPUJobsFromCM :%v.", cmErr)
	}

	if err := updateReSchedulerJobs(ssn); err != nil {
		klog.V(logErrorLev).Infof("getFaultNPUJobsFromCM :%v.", err)
		return fmt.Errorf("%v %v", cmErr, err)
	}

	return cmErr
}

// WriteFaultNPUJobsToCM Write FaultNPUJobs into ConfigMap.
func WriteFaultNPUJobsToCM(ssn *framework.Session, reSchedulerJobs map[api.JobID]ReSchedulerTasks) error {
	var faultNPUConfigMap = &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cmName,
			Namespace: cmNameSpace,
		},
		Data: make(map[string]string, 1),
	}

	for jobID, faultNPUJob := range reSchedulerJobs {
		buffer, err := json.Marshal(faultNPUJob)
		if err != nil {
			klog.V(logErrorLev).Infof("writeFaultNPUJobsToCM  %+v err: %v.", faultNPUJob, err)
			return err
		}
		job, ok := ssn.Jobs[jobID]
		if !ok {
			return fmt.Errorf("writeFaultNPUJobsToCM ssn not has %v", jobID)
		}
		// for jobID has '/'
		faultNPUConfigMap.Data[job.Namespace+"_"+job.Name] = string(buffer)
	}
	klog.V(logDebugLev).Infof("Write faultNPUJobs into cm: %+v.", faultNPUConfigMap)
	if err := createOrUpdateConfigMap(ssn.KubeClient(), faultNPUConfigMap); err != nil {
		klog.V(logErrorLev).Infof("createOrUpdateConfigMap : %v.", err)
		return err
	}

	return nil
}
