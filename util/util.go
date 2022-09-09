/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

// Package util is using for the total variable.
package util

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"k8s.io/api/core/v1"
	"k8s.io/klog"
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

// IsMapHasNPUResource Determines whether a target string exists in the map.
func IsMapHasNPUResource(resMap map[v1.ResourceName]float64, npuName string) bool {
	for k := range resMap {
		// must contains "huawei.com/Ascend"
		if strings.Contains(string(k), npuName) {
			return true
		}
	}
	return false
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
		return getConfigurationOldVersion(configurations), nil
	}

	return nil, fmt.Errorf("cannot get configurations by name [%s], name not in configurations", configKey)
}

// getConfigurationByKey called by GetConfigFromSchedulerConfigMap
func getConfigurationByKey(configKey string, configurations []conf.Configuration) *conf.Configuration {
	for _, cf := range configurations {
		if cf.Name == configKey {
			return &cf
		}
	}

	return nil
}

// getConfigurationOldVersion called by GetConfigFromSchedulerConfigMap
func getConfigurationOldVersion(configurations []conf.Configuration) *conf.Configuration {
	// if user removes configuration name and changes the order, will make mistakes.
	klog.V(LogDebugLev).Info("compatible with old versions, get the selector configuration successful.")
	return &configurations[0]
}

// Max return the bigger one
func Max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

// Min return the smaller one
func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

// IsSliceContain judges whether keyword in tasgetSlice
func IsSliceContain(keyword interface{}, targetSlice interface{}) bool {
	if targetSlice == nil {
		klog.V(LogErrorLev).Infof(
			"unable to converts %#v of type %T to map[interface{}]struct{}", targetSlice, targetSlice)
		return false
	}
	kind := reflect.TypeOf(targetSlice).Kind()
	if kind != reflect.Slice && kind != reflect.Array {
		klog.V(LogErrorLev).Infof(
			"the input %#v of type %T isn't a slice or array", targetSlice, targetSlice)
		return false
	}

	v := reflect.ValueOf(targetSlice)
	m := make(map[interface{}]struct{}, v.Len())
	for j := 0; j < v.Len(); j++ {
		m[v.Index(j).Interface()] = struct{}{}
	}

	_, ok := m[keyword]
	return ok
}
