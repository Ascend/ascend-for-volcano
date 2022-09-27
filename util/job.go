/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package util is using for the total variable.

*/
package util

import (
	"strings"

	"k8s.io/klog"
)

// IsSelectorMeetJob check the selectors
func IsSelectorMeetJob(jobSelectors, conf map[string]string) bool {
	for jobKey, jobValue := range jobSelectors {
		confValue, confOk := conf[jobKey]
		if !confOk {
			klog.V(LogErrorLev).Infof("conf has no job selector key:%s.", jobKey)
			return false
		}

		if !strings.Contains(confValue, jobValue) {
			klog.V(LogErrorLev).Infof("conf has no job selector value:%s.", jobValue)
			return false
		}
	}
	return true
}
