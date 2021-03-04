/*
Copyright(C) 2020. Huawei Technologies Co.,Ltd. All rights reserved.

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

Package topology910 is using for HuaWei Ascend910 pin affinity schedule.

*/
package topology910

import (
	"k8s.io/klog"
	"strconv"
	"strings"
	"volcano.sh/volcano/pkg/scheduler/conf"
)

const (
	// PluginName use in frame
	PluginName        = "topology910"
	npu910CardPreName = "Ascend910-"
	podPredicateTime  = "predicate-time"
	nodeNpuNumber     = 8
	// npu 910 card resource is named by npu device-plugin
	npu910CardName        = "huawei.com/Ascend910"
	logErrorLev           = 1
	logWarningLev         = 2
	logInfoLev            = 3
	logDebugLev           = 4
	npuHex                = 1000
	archSelector          = "host-arch"
	huaweiArchArm         = "huawei-arm"
	huaweiArchX86         = "huawei-x86"
	accelerator           = "accelerator"
	acceleratorValue      = "huawei.com-Ascend910"
	acceleratorType       = "accelerator-type"
	cardAcceleratorType   = "card"
	moduleAcceleratorType = "module"
	npuNumPerHccs         = 4
	magicNumInt1          = 1
	magicNumInt2          = 2
	magicNumInt3          = 3
)

// set 910 card topology in
func saveTopologyInMap(annotation map[string]interface{}, srcStr string) {
	// now only 910 card
	annotation[npu910CardName] = srcStr
}

func getDefaultSchedulerSelectorConfig() map[string]string {
	var defaultSchedulerConfig map[string]string
	defaultSchedulerConfig = make(map[string]string, magicNumInt3)

	defaultSchedulerConfig[archSelector] = huaweiArchArm + "|" + huaweiArchX86
	defaultSchedulerConfig[accelerator] = acceleratorValue
	defaultSchedulerConfig[acceleratorType] = cardAcceleratorType + "|" + moduleAcceleratorType

	return defaultSchedulerConfig
}

func getSchedulerSelectorConfig(confs []conf.Configuration) map[string]string {
	var customerScheduler map[string]string
	customerScheduler = make(map[string]string, magicNumInt2)

	if len(confs) != 0 {
		klog.V(logDebugLev).Infof("%s getSchedulerSelectorConfig ok[%+v]", PluginName, confs)
		// get customer config selector
		for k, v := range confs[0].Arguments {
			customerScheduler[k] = v
		}
		return nil
	}

	// default conf cannot be covered
	defaultSchedulerConfig := getDefaultSchedulerSelectorConfig()
	for k, v := range defaultSchedulerConfig {
		// if has default selector compare string,else add
		tempStr, ok := customerScheduler[k]
		if !ok {
			customerScheduler[k] = v
			klog.V(logDebugLev).Infof("%s use default config [%s]:[%s]", PluginName, k, v)
			continue
		}
		// exist default key, compare content
		if strings.Contains(tempStr, v) {
			klog.V(logDebugLev).Infof("%s default config has customer config [%s]:[%s]", PluginName, k, v)
			continue
		}
		// append not cover
		klog.V(logDebugLev).Infof("%s config key(%s) not same [%s]:[%s]",
			PluginName, k, v, tempStr)
		customerScheduler[k] = v + "|" + tempStr
	}

	return customerScheduler
}

func getTopToIntArray(topStr string) []int {
	var topInt []int
	var cardInt int
	var cardStr string
	var err error
	var topStrArray []string

	if topStr == "" {
		return []int{}
	}

	topStrArray = strings.Split(topStr, ",")
	for _, cardStr = range topStrArray {
		klog.V(logDebugLev).Infof("%s getTopFromNode cardStr %s", PluginName, cardStr)
		// cannot use strings 's Trim
		v := strings.TrimPrefix(cardStr, npu910CardPreName)
		klog.V(logDebugLev).Infof("%s getTopFromNode cardStr2 %s", PluginName, v)
		cardInt, err = strconv.Atoi(v)
		if err != nil {
			klog.V(logErrorLev).Infof("%s getTopFromNode conv failed %v", PluginName, err)
			return nil
		}

		topInt = append(topInt, cardInt)
	}

	return topInt
}

// Covert []int to string. Like [0,1] -> "Ascend910-0,Ascend910-1"
func changeIntArrToStr(top []int) string {
	var tmp int
	var str string

	i := 0
	for i, tmp = range top {
		str += npu910CardPreName + strconv.Itoa(tmp)
		if i+1 < len(top) {
			str += ","
		}
	}

	return str
}
