/*
Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.
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
	"strconv"
	"strings"
	"volcano.sh/volcano/pkg/scheduler/api"
	hwutil "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/util"
)

func initVNodesFn(nodes map[string]*api.NodeInfo) error {
	for key := range nodes {
		for _, vType := range VnpuType {
			nTopStr, err := getResourceFromAnnotationFn(nodes[key].Node.Annotations, vType)
			if err != nil {
				klog.V(logDebugLev).Infof("%s initVNodesFn %s %v", nodes[key].Name, vType, err)
				continue
			}
			err = hwutil.SaveTopologyInMap(nodes[key].Others, nTopStr, vType)
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
		return "", fmt.Errorf("requested %s does not exist", resourceName)
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

// return a slice of string of virtual card to be allocated
// sorted based on rules of try to allocate from physical card with the least remaining compute power
func getVCardWithLeastRemainPw(annotations map[string]interface{}, vType string) ([]string, error) {
	vAnno, exist := annotations[vType]
	if !exist {
		klog.V(logErrorLev).Infof("%v.", annotations)
		return nil, errors.New("getVCardWithLeastRemainPw no such vnpu resources")
	}

	cardByPriority, errR := getCardIDInAscRemainPwOrder(annotations)
	if errR != nil {
		return nil, errR
	}

	vNPUEachCard, errC := getVNPUByEachCard(vAnno)
	if errC != nil {
		return nil, errC
	}

	vNPUByPriority := []string{}
	for _, cardID := range cardByPriority {
		vNPUByPriority = append(vNPUByPriority, vNPUEachCard[cardID]...)
	}

	return vNPUByPriority, nil
}

// return a map of physical card ID(int) to certain vNPUs(slice of string) belong to that card
func getVNPUByEachCard(tmpData interface{}) (map[int][]string, error) {
	vNPUEachCard := map[int][]string{}

	vAnno, ok := tmpData.(string)
	if !ok {
		return nil, fmt.Errorf("%v not string", tmpData)
	}

	vAnnoList := strings.Split(vAnno, ",")
	for _, vAnnoInstance := range vAnnoList {
		vAnnoInstanceSlice := strings.Split(vAnnoInstance, "-")
		vAnnoCardID, err := strconv.Atoi(vAnnoInstanceSlice[len(vAnnoInstanceSlice)-1])
		if err != nil || vAnnoCardID < 0 || vAnnoCardID >= constIntNum8 {
			klog.V(logErrorLev).Infof("%s : %s : %d : %s.", vAnno, vAnnoInstanceSlice, vAnnoCardID, err)
			return nil, err
		}
		vNPUEachCard[vAnnoCardID] = append(vNPUEachCard[vAnnoCardID], vAnnoInstance)
	}

	return vNPUEachCard, nil
}

// change map of remain power of each card to a ascending sorted slice
func getSortedRemainPwSlice(remainPw map[int]int) []int {
	powerAvl := []struct {
		cardID int
		avl    int
	}{}

	for vCardID, vRPower := range remainPw {
		powerAvl = append(powerAvl, struct {
			cardID int
			avl    int
		}{cardID: vCardID, avl: vRPower})
	}

	sort.SliceStable(powerAvl, func(i, j int) bool {
		return powerAvl[i].avl < powerAvl[j].avl
	})

	cardsSorted := []int{}
	for _, avlInfo := range powerAvl {
		cardsSorted = append(cardsSorted, avlInfo.cardID)
	}

	return cardsSorted
}

func getVNPUsFromNodeByType(tmpData interface{}, vType string) (string, error) {
	vAnno, ok := tmpData.(string)
	if !ok {
		return "", fmt.Errorf("%v not strings", tmpData)
	}

	if _, exist := vnpuCoefficients[vType]; !exist || len(vAnno) == 0 {
		return "", fmt.Errorf("no %v npus", vType)
	}

	return vAnno, nil
}

// Get the sorted card ids, sort from least to most according to the remaining compute power of the card
func getCardIDInAscRemainPwOrder(annotations map[string]interface{}) ([]int, error) {
	remainPower := map[int]int{}

	for vType, tmp := range annotations {
		vAnno, err := getVNPUsFromNodeByType(tmp, vType)
		if err != nil {
			klog.V(logDebugLev).Infof("getVNPUsFromNodeByType %v.", err)
			continue
		}
		cPowerList := strings.Split(vType, "-")
		cPower, err := strconv.Atoi(strings.TrimSuffix(cPowerList[len(cPowerList)-1], "c"))
		if err != nil {
			klog.V(logErrorLev).Infof("%s : %s : %s.", vType, cPowerList, err)
			return nil, err
		}

		vAnnoList := strings.Split(vAnno, ",")
		for _, vAnnoInstance := range vAnnoList {
			vAnnoInstanceSlice := strings.Split(vAnnoInstance, "-")
			vAnnoCardID, err := strconv.Atoi(vAnnoInstanceSlice[len(vAnnoInstanceSlice)-1])
			if err != nil {
				klog.V(logErrorLev).Infof("%s : %s : %s.", vAnno, vAnnoInstanceSlice, err)
				return nil, err
			}
			remainPower[vAnnoCardID] += cPower
		}
	}

	remainPwID := getSortedRemainPwSlice(remainPower)

	return remainPwID, nil
}
