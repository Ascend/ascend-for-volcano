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
