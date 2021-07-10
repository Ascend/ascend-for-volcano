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

Package commonv910 is using for virtual HuaWei Ascend910 schedule.

*/
package commonv910

import (
	"errors"
	"fmt"
	"k8s.io/klog"
	"sort"
	"strings"
	"volcano.sh/volcano/pkg/scheduler/api"
	hwutil "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/util"
)

func initVNodesFn(nodes map[string]*api.NodeInfo) error {
	for _, node := range nodes {
		node.Others = make(map[string]interface{}, 1)
		for _, vType := range VnpuType {
			nTopStr, err := getResourceFromAnnotationFn(node.Node.Annotations, vType)
			if err != nil {
				continue
			}
			err = hwutil.SaveTopologyInMap(node.Others, nTopStr, vType)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// GetVnpuType get VnpuType
func GetVnpuType() []string {
	return VnpuType
}

func getResourceFromAnnotationFn(Annotations map[string]string, resourceName string) (string, error) {
	topStr, ok := Annotations[resourceName]
	// In case of kubernetes doesn't have some kind of resource type, but name of that type was written in
	// node annotation with a value of empty string. If topStr is empty, an error should be returned so that that type
	// of resource will be ignored.
	if !ok || topStr == "" {
		klog.V(logDebugLev).Infof("getResourceFromNodeAnnotationFn failed.")
		return "", errors.New("requested resource does not exist")
	}

	return topStr, nil
}

func getNPUsFromNodeAnnotation(annotations map[string]string, resourceName string) ([]string, error) {
	topStr, err := getResourceFromAnnotationFn(annotations, resourceName)
	if err != nil {
		klog.V(logErrorLev).Infof("getNPUsFromNodeAnnotation failed to get annotation value")
		return nil, err
	}

	prefix := strings.TrimPrefix(resourceName, npu910CardNamePrefix)
	tops := strings.Split(topStr, ",")
	sort.Strings(tops)
	for i, top := range tops {
		if !strings.HasPrefix(top, prefix) {
			klog.V(logErrorLev).Infof("getNPUsFromNodeAnnotation: vnpu name(%s) did not match its type(%s)",
				top, prefix)
			return nil, fmt.Errorf("vnpu name(%s) did not match its type(%s)", top, prefix)
		}

		if i > 0 && top == tops[i-1] {
			klog.V(logErrorLev).Infof("getNPUsFromNodeAnnotation: got duplicated npu(%s)", top)
			return nil, fmt.Errorf("got duplicated npu(%s)", top)
		}
	}

	return tops, nil
}

// Get number of devices in node annotation
func getNPUNumFromNodeAnnotation(node *api.NodeInfo, resourceName string) (int, error) {
	npuArr, err := getNPUsFromNodeAnnotation(node.Node.Annotations, resourceName)
	if err != nil {
		return 0, err
	}

	return len(npuArr), nil
}

func judgeResourceTypeByTopInfo(instance string) string {
	var vType string
	for _, vt := range VnpuType {
		v := strings.TrimPrefix(vt, npu910CardNamePrefix)
		if strings.HasPrefix(instance, v) {
			vType = vt
			break
		}
	}

	return vType
}

func getTopStrFromNodeOther(othersMap map[string]interface{}, npuCardName string) ([]string, error) {
	var topArr []string

	valueTmp, ok := othersMap[npuCardName]
	if !ok {
		klog.V(logErrorLev).Infof("%s getNodeNPUStrFromOther other nil.", npuCardName)
		return nil, errors.New("nodeTopStrArr nil")
	}

	mapStr, ok := valueTmp.(string)
	if !ok {
		klog.V(logErrorLev).Infof("%s getNodeNPUStrFromOther not string type.", npuCardName)
		return nil, errors.New("nodeTopStrArr nil")
	}

	topArr = strings.Split(mapStr, ",")
	return topArr, nil
}

// Update occupied resource info after allocate
func updateTopStrOfNodeOtherAlloc(nodeTopStrArr []string, top []string) string {
	var tmpTopStrArr []string
	var existFlag bool

	for _, nTop := range nodeTopStrArr {
		existFlag = false
		for _, tTop := range top {
			if nTop == tTop {
				existFlag = true
				break
			}
		}
		if !existFlag {
			tmpTopStrArr = append(tmpTopStrArr, nTop)
		}
	}
	klog.V(logDebugLev).Infof("updateTopStrOfNodeOtherAlloc : %v.", tmpTopStrArr)
	newNodeTopStr := strings.Join(tmpTopStrArr, ",")

	return newNodeTopStr
}

// Update occupied resource info after release
func updateTopStrOfNodeOtherRelease(nodeTopStrArr []string, top []string) string {
	var tmpTopStrArr []string

	tmpTopMap := make(map[string]int, const3)
	// add tops that already exist in node.Others to tmp map
	for _, nTop := range nodeTopStrArr {
		tmpTopMap[nTop] = 0
	}
	// add tops that been released to tmp map
	for _, tTop := range top {
		if _, ok := tmpTopMap[tTop]; ok {
			klog.V(logInfoLev).Infof("updateTopStrOfNodeOtherRelease card exists: %s.", tTop)
			continue
		}
		tmpTopMap[tTop] = 0
	}

	for k := range tmpTopMap {
		tmpTopStrArr = append(tmpTopStrArr, k)
	}

	klog.V(logDebugLev).Infof("updateTopStrOfNodeOtherRelease : %v.", tmpTopStrArr)
	newNodeTopStr := strings.Join(tmpTopStrArr, ",")

	return newNodeTopStr
}

// Update node info to node.Others
func updateNPUNodeTopology(node *api.NodeInfo, top interface{}, updateFn func([]string, []string) string) error {
	var vType string

	topArr, ok := top.([]string)
	if !ok {
		return errors.New("invalid argument")
	}

	topInstance := topArr[0]
	vType = judgeResourceTypeByTopInfo(topInstance)
	if vType == "" {
		return errors.New("invalid top content")
	}

	// get node available top from node.Others
	nodeTopStrArr, err := getTopStrFromNodeOther(node.Others, vType)
	if err != nil {
		klog.V(logErrorLev).Infof("updateNPUNodeTopology node(%s) top nil.", node.Name)
		return err
	}
	// update to node.Others
	newNodeTopStr := updateFn(nodeTopStrArr, topArr)
	err = hwutil.ReloadNewTopToNodeOther(node, newNodeTopStr, vType)
	if err != nil {
		klog.V(logErrorLev).Infof("reloadNewTopToNode failed.")
		return err
	}

	klog.V(logInfoLev).Infof("ReloadNewTopToNode %s to %s successes.", newNodeTopStr, node.Name)

	return nil
}