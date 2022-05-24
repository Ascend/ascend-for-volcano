/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package util is using for HuaWei Ascend9 pin affinity schedule utilities.

*/
package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"strconv"
	"strings"

	"k8s.io/klog"
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
		// cannot use strings 's Trim
		v := strings.TrimPrefix(cardStr, npuCardPreName)
		cardInt, err = strconv.Atoi(v)
		if err != nil {
			klog.V(LogErrorLev).Infof("ChangeTopToIntArray conv failed %v.", err)
			return nil
		}

		topInt = append(topInt, cardInt)
	}
	klog.V(LogDebugLev).Infof("ChangeTopToIntArray %v.", topInt)
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

func getDefaultSchedulerSelectorConfig() map[string]string {
	var defaultSchedulerConfig map[string]string
	defaultSchedulerConfig = make(map[string]string, NPUIndex3)

	defaultSchedulerConfig[ArchSelector] = HuaweiArchArm + "|" + HuaweiArchX86
	defaultSchedulerConfig[accelerator] = accelerator910Value + "|" + accelerator310Value
	defaultSchedulerConfig[AcceleratorType] = CardAcceleratorType + "|" + ModuleAcceleratorType +
		"|" + ChipAcceleratorType

	return defaultSchedulerConfig
}

// GetSchedulerSelectorConfig Get selector from volcano config file.
func GetSchedulerSelectorConfig(confs []conf.Configuration) map[string]string {
	var customerScheduler map[string]string
	customerScheduler = make(map[string]string, NPUIndex2)

	configuration, err := GetConfigFromSchedulerConfigMap(CMSelectorKey, confs)
	if err != nil {
		klog.V(LogDebugLev).Info(err)
	}
	if len(confs) != 0 && err == nil {
		klog.V(LogDebugLev).Infof("getSchedulerSelectorConfig ok[%+v].", confs)
		// get customer config selector
		for k, v := range configuration.Arguments {
			customerScheduler[k] = v
		}
		klog.V(LogDebugLev).Infof("add config SchedulerSelector ok[%+v].", customerScheduler)
	}

	// default conf cannot be covered
	defaultSchedulerConfig := getDefaultSchedulerSelectorConfig()
	for k, v := range defaultSchedulerConfig {
		// if has default selector compare string,else add
		tempStr, ok := customerScheduler[k]
		if !ok {
			customerScheduler[k] = v
			klog.V(LogDebugLev).Infof("use default config [%s]:[%s].", k, v)
			continue
		}
		// exist default key, compare content
		if strings.Contains(tempStr, v) {
			klog.V(LogDebugLev).Infof("default config has customer config [%s]:[%s].", k, v)
			continue
		}
		// append not cover
		klog.V(LogDebugLev).Infof("config key(%s) not same [%s]:[%s].", k, v, tempStr)
		customerScheduler[k] = v + "|" + tempStr
	}
	klog.V(LogDebugLev).Infof("add getSchedulerSelectorConfig ok[%+v].", customerScheduler)
	return customerScheduler
}

// CheckTaskAndNodeSelectorMeet Check the selector between task and node.
func CheckTaskAndNodeSelectorMeet(tSelectors map[string]string,
	nSelector map[string]string,
	conf map[string]string) error {

	for taskKey, taskValue := range tSelectors {
		confValue, confOk := conf[taskKey]
		if !confOk {
			klog.V(LogErrorLev).Infof("conf has no task selector:%s.", taskKey)
			return fmt.Errorf("%s : conf has no:%s", NodeNoFitSelectorError, taskKey)
		}

		nodeValue, nodeOk := nSelector[taskKey]
		if !nodeOk {
			klog.V(LogErrorLev).Infof("node has no task selector:%s.", taskKey)
			return fmt.Errorf("%s : node has no:%s", NodeNoFitSelectorError, taskKey)
		}

		if !strings.Contains(confValue, taskValue) || !strings.EqualFold(taskValue, nodeValue) {
			klog.V(LogErrorLev).Infof("selector(%s) not equal: task(%s) node(%s) conf(%s).",
				taskKey, taskValue, nodeValue, confValue)
			return fmt.Errorf("%s key[%s] : task(%s) node(%s) conf(%s)",
				NodeNoFitSelectorError, taskKey, taskValue, nodeValue, confValue)
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
			klog.V(LogInfoLev).Infof("%s getRealTopAfterRelease already has cardId: %d.", npuCardPreName, tTopI)
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
			klog.V(LogDebugLev).Infof("not npu task[%s], no need selector.", task.Name)
			return nil
		}
		// npu task need selector
		klog.V(LogErrorLev).Infof("task[%s] no selector in select node[%s].", task.Name, node.Name)
		return errors.New(NodeNoFitSelectorError)
	}

	// task has selector, so node should have
	nodeSelector, errNode := GetNodeSelector(node)
	if errNode != nil {
		klog.V(LogErrorLev).Infof("GetNodeSelector task[%s] on node(%s) %v.", task.Name, node.Name, errNode)
		return errors.New(NodeNoFitSelectorError)
	}

	if err := CheckTaskAndNodeSelectorMeet(taskSelectors, nodeSelector, conf); err != nil {
		klog.V(LogErrorLev).Infof("CheckTaskAndNodeSelectorMeet %s err:%v.", node.Name, err)
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
			klog.V(LogErrorLev).Infof("%v.", msg)
			return msg
		}

		if !isSelectorContains(defValue, jobValue) {
			msg := fmt.Errorf("%s selector[%s]:[%s] not in [%s]", job.Name, defKey, jobValue, defValue)
			klog.V(LogErrorLev).Infof("%v.", msg)
			return msg
		}
	}
	return nil
}

// ValidStringMapKeyAndValue Valid map key and value.
func ValidStringMapKeyAndValue(tmpMap map[string]string, key, value string) bool {
	tmpValue, ok := tmpMap[key]
	if !ok {
		// no AcceleratorType means module
		return false
	}

	if tmpValue == value {
		return true
	}

	klog.V(LogDebugLev).Infof("valid ok .")
	return false
}

// MakeDataHash make data hash.
func MakeDataHash(data interface{}) uint32 {
	dataString, marshErr := MarshalCacheDataToString(data)
	if marshErr != nil {
		return 0
	}
	return crc32.ChecksumIEEE([]byte(dataString))
}

// MarshalCacheDataToString Marshal cache data to string.
func MarshalCacheDataToString(data interface{}) (string, error) {
	dataBuffer, err := json.Marshal(data)
	if err != nil {
		klog.V(LogErrorLev).Infof("marshalCacheDataToString err: %v.", err)
		return "", err
	}
	return string(dataBuffer), nil
}

// GetResourceFromAnnotationFn get the source from annotation.
func GetResourceFromAnnotationFn(Annotations map[string]string, resourceName string) (string, error) {
	topStr, ok := Annotations[resourceName]
	// In case of kubernetes doesn't have some kind of resource type, but name of that type was written in
	// node annotation with a value of empty string. If topStr is empty, an error should be returned so that that type
	// of resource will be ignored.
	if !ok || topStr == "" {
		return "", fmt.Errorf("requested %s does not exist", resourceName)
	}

	return topStr, nil
}

// GetConfigFromSchedulerConfigMap get config info from yaml
func GetConfigFromSchedulerConfigMap(configKey string, configurations []conf.Configuration) (*conf.Configuration,
	error) {
	if len(configurations) == 0 {
		return nil, errors.New("no configurations in scheduler configmap")
	}

	// in the new version, the configuration is obtained based on the configured name field.
	if config := getConfigurationByKey(configKey, configurations); config != nil {
		klog.V(LogDebugLev).Infof("get the configurations by name [%s] successful.", configKey)
		return config, nil
	}

	// compatible with old versions, because of the name field is not configured in the old versions.
	if configKey == CMSelectorKey {
		// if user removes configuration name and changes the order, will make mistakes.
		klog.V(LogDebugLev).Info("compatible with old versions, get the selector configuration successful.")
		return &configurations[0], nil
	}

	return nil, fmt.Errorf("cannot get configurations by name [%s], name not in configurations", configKey)
}

func getConfigurationByKey(configKey string, configurations []conf.Configuration) *conf.Configuration {
	for _, cf := range configurations {
		if cf.Name == configKey {
			return &cf
		}
	}

	return nil
}
