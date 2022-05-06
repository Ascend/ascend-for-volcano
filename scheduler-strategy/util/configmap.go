/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package util is using for HuaWei infer common Ascend pin affinity schedule.

*/
package util

import (
	"context"
	"fmt"
	"reflect"

	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/framework"
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
	_, cErr := k8s.CoreV1().ConfigMaps(cm.ObjectMeta.Namespace).Create(context.TODO(), cm, metav1.CreateOptions{})
	if cErr != nil {
		if !apierrors.IsAlreadyExists(cErr) {
			return fmt.Errorf("unable to create ConfigMap:%v", cErr)
		}

		// To reduce the cm write operations
		if !IsConfigMapChanged(k8s, cm, cmName, nameSpace) {
			klog.V(LogInfoLev).Infof("configMap not changed,no need update")
			return nil
		}

		_, err := k8s.CoreV1().ConfigMaps(cm.ObjectMeta.Namespace).Update(context.TODO(), cm, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("unable to update ConfigMap:%v", err)
		}
	}
	return nil
}

// DeleteSchedulerConfigMap Delete configMap.
func DeleteSchedulerConfigMap(ssn *framework.Session, nameSpace, cmName string) error {
	err := ssn.KubeClient().CoreV1().ConfigMaps(nameSpace).Delete(context.TODO(), cmName, metav1.DeleteOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			klog.V(LogInfoLev).Infof("Failed to delete Configmap %v in%v: %v",
				nameSpace, cmName, err)
			return err
		}
	}
	return nil
}
