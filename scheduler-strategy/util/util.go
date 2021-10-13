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

Package util is using for HuaWei Ascend9 pin affinity schedule utilities.

*/
package util

import (
	"errors"
	"fmt"
	"k8s.io/klog"
	"strconv"
	"strings"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/conf"
)

// ChangeTopToIntArray Change npu card ids from string to int array.
func ChangeTopToIntArray(topStr string, npuCardPreName string) []int {
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
		klog.V(logDebugLev).Infof("ChangeTopToIntArray cardStr %s.", cardStr)
		// cannot use strings 's Trim
		v := strings.TrimPrefix(cardStr, npuCardPreName)
		cardInt, err = strconv.Atoi(v)
		if err != nil {
			klog.V(logErrorLev).Infof("ChangeTopToIntArray conv failed %v.", err)
			return nil
		}

		topInt = append(topInt, cardInt)
	}

	return topInt
}

// SaveTopologyInMap Set npu card ids in annotation.
func SaveTopologyInMap(annotation map[string]interface{}, srcStr string, npuCardName string) error {
	// now only 910 card
	if annotation != nil {
		annotation[npuCardName] = srcStr
		return nil
	}

	return errors.New("nil annotation map")
}

func isSelectorMeetJob(jobSelectors, schedulerConf map[string]string) bool {
	for jobKey, jobValue := range jobSelectors {
		confValue, confOk := schedulerConf[jobKey]
		if !confOk {
			klog.V(logErrorLev).Infof("conf has no job selector key:%s.", jobKey)
			return false
		}

		if !strings.Contains(confValue, jobValue) {
			klog.V(logErrorLev).Infof("conf has no job selector value:%s.", jobValue)
			return false
		}
	}
	return true
}

func getDefaultSchedulerSelectorConfig() map[string]string {
	var defaultSchedulerConfig map[string]string
	defaultSchedulerConfig = make(map[string]string, constIntNum3)

	defaultSchedulerConfig[archSelector] = huaweiArchArm + "|" + huaweiArchX86
	defaultSchedulerConfig[accelerator] = accelerator910Value + "|" + accelerator310Value
	defaultSchedulerConfig[acceleratorType] = cardAcceleratorType + "|" + moduleAcceleratorType +
		"|" + chipAcceleratorType

	return defaultSchedulerConfig
}

// GetSchedulerSelectorConfig Get selector from volcano config file.
func GetSchedulerSelectorConfig(confs []conf.Configuration) map[string]string {
	var customerScheduler map[string]string
	customerScheduler = make(map[string]string, constIntNum2)

	if len(confs) != 0 {
		klog.V(logDebugLev).Infof("getSchedulerSelectorConfig ok[%+v].", confs)
		// get customer config selector
		for k, v := range confs[0].Arguments {
			customerScheduler[k] = v
		}
		klog.V(logDebugLev).Infof("add config SchedulerSelector ok[%+v].", customerScheduler)
	}

	// default conf cannot be covered
	defaultSchedulerConfig := getDefaultSchedulerSelectorConfig()
	for k, v := range defaultSchedulerConfig {
		// if has default selector compare string,else add
		tempStr, ok := customerScheduler[k]
		if !ok {
			customerScheduler[k] = v
			klog.V(logDebugLev).Infof("use default config [%s]:[%s].", k, v)
			continue
		}
		// exist default key, compare content
		if strings.Contains(tempStr, v) {
			klog.V(logDebugLev).Infof("default config has customer config [%s]:[%s].", k, v)
			continue
		}
		// append not cover
		klog.V(logDebugLev).Infof("config key(%s) not same [%s]:[%s].", k, v, tempStr)
		customerScheduler[k] = v + "|" + tempStr
	}
	klog.V(logDebugLev).Infof("add getSchedulerSelectorConfig ok[%+v].", customerScheduler)
	return customerScheduler
}

// CheckTaskAndNodeSelectorMeet Check the selector between task and node.
func CheckTaskAndNodeSelectorMeet(tSelectors map[string]string,
	nSelector map[string]string,
	conf map[string]string) error {

	for taskKey, taskValue := range tSelectors {
		confValue, confOk := conf[taskKey]
		if !confOk {
			klog.V(logErrorLev).Infof("conf has no task selector:%s.", taskKey)
			return fmt.Errorf("%s : conf has no:%s", nodeNoFitSelectorError, taskKey)
		}

		nodeValue, nodeOk := nSelector[taskKey]
		if !nodeOk {
			klog.V(logErrorLev).Infof("node has no task selector:%s.", taskKey)
			return fmt.Errorf("%s : node has no:%s", nodeNoFitSelectorError, taskKey)
		}

		if !strings.Contains(confValue, taskValue) || !strings.EqualFold(taskValue, nodeValue) {
			klog.V(logErrorLev).Infof("selector(%s) not equal: task(%s) node(%s) conf(%s).",
				taskKey, taskValue, nodeValue, confValue)
			return fmt.Errorf("%s key[%s] : task(%s) node(%s) conf(%s)",
				nodeNoFitSelectorError, taskKey, taskValue, nodeValue, confValue)
		}
	}

	return nil
}

// CheckNodeNPUStabilize Check node npu 's stable.
func CheckNodeNPUStabilize(nodeNPUIdleNumFromTop int, nodeNPUIdleNumFromIdle int) error {
	if nodeNPUIdleNumFromTop != nodeNPUIdleNumFromIdle {
		return fmt.Errorf("node not stable for annotations(%d) : idle(%d)",
			nodeNPUIdleNumFromTop, nodeNPUIdleNumFromIdle)
	}

	return nil
}

// ChangeIntArrToStr Covert []int to string. Like [0,1] -> "Ascend910-0,Ascend910-1".
func ChangeIntArrToStr(top []int, npuCardPreName string) string {
	var tmp int
	var str string

	i := 0
	for i, tmp = range top {
		str += npuCardPreName + strconv.Itoa(tmp)
		if i+1 < len(top) {
			str += ","
		}
	}

	return str
}

// GetRealTopAfterRelease Get npu card ids after release.
func GetRealTopAfterRelease(nodeDeviceIDs []int, taskDeviceIDs []int, npuCardPreName string) string {
	var tmpDeviceIDs []int
	tmpTopMap := make(map[int]int, nodeNPUNumber)
	// add node topology to tmp map
	for _, nTopI := range nodeDeviceIDs {
		tmpTopMap[nTopI] = 0
	}
	// add task topology to tmp map, Deduplicate the same topology
	for _, tTopI := range taskDeviceIDs {
		if _, ok := tmpTopMap[tTopI]; ok {
			klog.V(logInfoLev).Infof("%s getRealTopAfterRelease already has cardId: %d.", npuCardPreName, tTopI)
			continue
		}
		tmpTopMap[tTopI] = 0
	}
	// change tmp map to slice
	for k := range tmpTopMap {
		tmpDeviceIDs = append(tmpDeviceIDs, k)
	}
	// change int to string
	return ChangeIntArrToStr(tmpDeviceIDs, npuCardPreName)
}

// IsSelectorMeetNode Determines whether the selectors of the task and node are equal.
func IsSelectorMeetNode(task *api.TaskInfo, node *api.NodeInfo, conf map[string]string, cardName string) error {
	taskSelectors := GetTaskSelectors(task)
	if len(taskSelectors) == 0 {
		if err := IsNPUTask(task, cardName); err != nil {
			klog.V(logDebugLev).Infof("not npu task[%s], no need selector.", task.Name)
			return nil
		}
		// npu task need selector
		klog.V(logErrorLev).Infof("task[%s] no selector in select node[%s].", task.Name, node.Name)
		return errors.New(nodeNoFitSelectorError)
	}

	// task has selector, so node should have
	nodeSelector, errNode := GetNodeSelector(node)
	if errNode != nil {
		klog.V(logErrorLev).Infof("GetNodeSelector task[%s] on node(%s) %v.", task.Name, node.Name, errNode)
		return errors.New(nodeNoFitSelectorError)
	}

	if err := CheckTaskAndNodeSelectorMeet(taskSelectors, nodeSelector, conf); err != nil {
		klog.V(logErrorLev).Infof("CheckTaskAndNodeSelectorMeet %s err:%v.", node.Name, err)
		return err
	}

	return nil
}

// Determine if the selectors are exactly equal.
func isSelectorContains(defValue, jobValue string) bool {
	for _, v := range strings.Split(defValue, "|") {
		if strings.EqualFold(v, jobValue) {
			return true
		}
	}

	return false
}

// CompareNPUSelector Compare the selector.
func CompareNPUSelector(job *api.JobInfo, jobS map[string]string, defaultS map[string]string) error {
	for defKey, defValue := range defaultS {
		jobValue, jobOk := jobS[defKey]
		if !jobOk {
			msg := fmt.Errorf("%s has no selector:%s", job.Name, defKey)
			klog.V(logErrorLev).Infof("%v.", msg)
			return msg
		}

		if !isSelectorContains(defValue, jobValue) {
			msg := fmt.Errorf("%s selector[%s]:[%s] not in [%s]", job.Name, defKey, jobValue, defValue)
			klog.V(logErrorLev).Infof("%v.", msg)
			return msg
		}
	}
	return nil
}

// ValidStringMapKeyAndValue Valid map key and value.
func ValidStringMapKeyAndValue(tmpMap map[string]string, key, value string) bool {
	tmpValue, ok := tmpMap[key]
	if !ok {
		// no acceleratorType means module
		return false
	}

	if tmpValue == value {
		return true
	}

	klog.V(logDebugLev).Infof("valid ok .")
	return false
}
