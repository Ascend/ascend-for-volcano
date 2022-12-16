/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.

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

Package util is using for HuaWei infer common Ascend pin affinity schedule.

*/
package util

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
)

// GetConfigMapWithRetry  Get config map from k8s.
func GetConfigMapWithRetry(client kubernetes.Interface, namespace, cmName string) (*v1.ConfigMap, error) {
	var cm *v1.ConfigMap
	var err error

	// There can be no delay or blocking operations in a session.
	if cm, err = client.CoreV1().ConfigMaps(namespace).Get(context.TODO(), cmName, metav1.GetOptions{}); err != nil {
		return nil, err
	}

	return cm, nil
}

// IsConfigMapChanged judge the cm wither is same. true is no change.
func IsConfigMapChanged(k8s kubernetes.Interface, cm *v1.ConfigMap, cmName, nameSpace string) bool {
	cmData, getErr := GetConfigMapWithRetry(k8s, nameSpace, cmName)
	if getErr != nil {
		return true
	}
	if reflect.DeepEqual(cmData, cm) {
		return false
	}

	return true
}

// CreateOrUpdateConfigMap Create or update configMap.
func CreateOrUpdateConfigMap(k8s kubernetes.Interface, cm *v1.ConfigMap, cmName, nameSpace string) error {
	klog.V(LogDebugLev).Infof("cmName: %s, cmNamespace: %s", cmName, cm.ObjectMeta.Namespace)
	_, cErr := k8s.CoreV1().ConfigMaps(cm.ObjectMeta.Namespace).Create(context.TODO(), cm, metav1.CreateOptions{})
	if cErr != nil {
		if !errors.IsAlreadyExists(cErr) {
			return fmt.Errorf("unable to create ConfigMap:%#v", cErr)
		}

		// To reduce the cm write operations
		if !IsConfigMapChanged(k8s, cm, cmName, nameSpace) {
			klog.V(LogInfoLev).Infof("configMap not changed,no need update")
			return nil
		}

		_, err := k8s.CoreV1().ConfigMaps(cm.ObjectMeta.Namespace).Update(context.TODO(), cm, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("unable to update ConfigMap:%#v", err)
		}
	}
	return nil
}

// UpdateConfigmapIncrementally update configmap Map data but keep the key value pair that new data does not have
func UpdateConfigmapIncrementally(kubeClient kubernetes.Interface, ns, name string,
	newData map[string]string) (map[string]string, error) {
	if len(newData) == 0 {
		return newData, fmt.Errorf("newData is empty")
	}
	oldCM, err := GetConfigMapWithRetry(kubeClient, ns, name)
	if err != nil || oldCM == nil {
		return newData, fmt.Errorf("get old configmap from kubernetes failed")
	}
	oldCMData := oldCM.Data
	if oldCMData != nil {
		for key, value := range oldCMData {
			_, ok := newData[key]
			if !ok {
				newData[key] = value // place the key-value pairs from kubernetes back
				continue
			}
		}
	}
	return newData, nil
}

// ReadCMFromKubernetes readCMFromKubernetes
func (comCM *ComConfigMap) ReadCMFromKubernetes(client kubernetes.Interface) {
	cm, err := GetConfigMapWithRetry(client, comCM.Namespace, comCM.Name)
	if err != nil {
		klog.V(LogErrorLev).Infof("ReadCMFromKubernetes: %s.", err.Error())
		return
	}
	comCM.Name = cm.Name
	comCM.Namespace = cm.Namespace
	comCM.Data = cm.Data
}

// UpdateCMToKubernetes updateCMToKubernetes
func (comCM *ComConfigMap) UpdateCMToKubernetes(client kubernetes.Interface) {
	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      comCM.Name,
			Namespace: comCM.Namespace,
		},
		Data: comCM.Data,
	}
	err := CreateOrUpdateConfigMap(client, cm, cm.Name, cm.Namespace)
	if err != nil {
		klog.V(LogErrorLev).Infof("getNodeDeviceInfoFromCM: %s.", err.Error())
	}
}

// MarshalCacheDataToString marshalCacheDataToString
func (comCM *ComConfigMap) MarshalCacheDataToString(data interface{}) (string, error) {
	dataBuffer, err := json.Marshal(data)
	if err != nil {
		klog.V(LogErrorLev).Infof("MarshalCacheDataToString err: %#v", err)
		return "", err
	}
	return string(dataBuffer), nil
}
