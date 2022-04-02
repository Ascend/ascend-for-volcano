/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package vnpuutil is using for virtual HuaWei Ascend910 schedule.

*/
package vnpuutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/util"
)

// IsVJobReqNPUMeetTotalResources npu card only has one kind and one.
func IsVJobReqNPUMeetTotalResources(npu string, res map[string]float64) bool {
	if len(res) == 0 || npu == "" {
		klog.V(util.LogErrorLev).Info("IsVJobReqNPUMeetTotalResources parameters error.")
		return false
	}
	temp := strings.Split(npu, "-")
	if value, ok := res[temp[0]]; ok {
		if value >= 1.0 {
			return true
		}
	}
	return false
}

// ConvertToVNPUAllocInfCacheFromCMData Convert string to VNPUAllocInfCache.
func ConvertToVNPUAllocInfCacheFromCMData(buffer string) (*VNPUAllocInfCache, error) {
	cache := VNPUAllocInfCache{}
	if unmarshalErr := json.Unmarshal([]byte(buffer), &cache); unmarshalErr != nil {
		klog.V(util.LogInfoLev).Infof("ConvertToVNPUAllocInfCacheMapFromCMData Unmarshal: %v.", unmarshalErr)
		return nil, unmarshalErr
	}
	return &cache, nil
}

// GetVNPUAllocInfDataFromCacheCM Get VNPU AllocInfData from cache configmap at session start.
func GetVNPUAllocInfDataFromCacheCM(ssn *framework.Session) (*VNPUAllocInfCache, error) {
	cmData, err := util.GetConfigMapWithRetry(ssn.KubeClient(), VNPUCMNameSpace, VNPUCacheCMName)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("GetVJobNamesFromCache :%v.", err)
		return nil, err
	}
	tmp, ok := cmData.Data[VNPCMDataKey]
	if !ok {
		klog.V(util.LogErrorLev).Infof("GetVJobNamesFromCache :%v.", err)
		return nil, err
	}

	return ConvertToVNPUAllocInfCacheFromCMData(tmp)
}

// WriteVNPUAllocInfDataIntoCacheCM Write VNPUAllocInfData into cache CM.
func WriteVNPUAllocInfDataIntoCacheCM(ssn *framework.Session) error {
	var vNPUCM = &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      VNPUCacheCMName,
			Namespace: VNPUCMNameSpace,
		},
		Data: GetVNPUCacheCMData(VNPUAllocData),
	}

	klog.V(util.LogDebugLev).Infof("Write vNPU cache into cm: %+v/%v.", vNPUCM.Namespace, vNPUCM.Name)
	if err := util.CreateOrUpdateConfigMap(ssn.KubeClient(), vNPUCM, VNPUCacheCMName, VNPUCMNameSpace); err != nil {
		klog.V(util.LogErrorLev).Infof("writevNPUAllocInfIntoCm : %v.", err)
		return err
	}
	return nil
}

// ChangeReqVNPUToCores covert the string to npu cores.
// needNPU like huawei.com/Ascend910-16c
func ChangeReqVNPUToCores(needNPU string) (int, error) {
	split := strings.Split(needNPU, "-")
	if len(split) != util.ConstIntNum2 {
		return 0, fmt.Errorf("err npu resource %s", needNPU)
	}
	tmp := split[1]
	content := tmp[:len(tmp)-1]
	return strconv.Atoi(content)
}

func updateCMCardData(dataSet, inDataSet []CardVNPUs) ([]CardVNPUs, error) {
	if len(dataSet) == 0 || len(inDataSet) == 0 {
		fmt.Printf("haha-4%+v===%+v\n", dataSet, inDataSet)
		return append(dataSet, inDataSet...), nil
	}

	tmpMap := make(map[string]CardVNPUs, util.ConstIntNum3)
	for _, firData := range dataSet {
		tmp := firData
		tmpMap[tmp.CardName] = tmp
	}
	fmt.Printf("haha-5-0%+v===%+v\n", tmpMap, inDataSet)
	for _, secData := range inDataSet {
		value, ok := tmpMap[secData.CardName]
		if !ok {
			tmpMap[secData.CardName] = secData
			fmt.Printf("haha-5-1%+v===%+v\n", tmpMap, secData)
			continue
		}
		// RemoveDuplicates
		tmp := CardVNPUs{
			CardName: secData.CardName,
			Req:      append(value.Req, secData.Req...),
			Alloc:    append(value.Alloc, secData.Alloc...),
		}
		tmpMap[secData.CardName] = tmp
		fmt.Printf("haha-3%+v\n", tmp)
	}
	var updateData []CardVNPUs
	for _, value := range tmpMap {
		tmp := value
		updateData = append(updateData, tmp)
		fmt.Printf("haha-5-2%+v===%+v\n", updateData, tmp)
	}
	return updateData, nil
}

func updateCMNodeData(dataSet []NodeVNPUs, inData NodeVNPUs) ([]NodeVNPUs, error) {
	findNode := false
	var returnValue []NodeVNPUs
	for _, nodeData := range dataSet {
		if nodeData.NodeName != inData.NodeName {
			continue
		}
		// find the exist node
		findNode = true
		fmt.Printf("haha-2-1%+v==%+v\n", nodeData.Cards, inData.Cards)
		cardSet, cardErr := updateCMCardData(nodeData.Cards, inData.Cards)
		if cardErr != nil {
			klog.V(util.LogErrorLev).Infof("updateCMNodeData %v.", cardErr)
			return nil, cardErr
		}
		tmp := nodeData
		tmp.Cards = cardSet
		returnValue = append(returnValue, tmp)
		fmt.Printf("haha-2-2%+v==%+v==%+v\n", cardSet, nodeData.Cards, inData.Cards)
		//dataSet[k].Cards = cardSet
	}
	if !findNode {
		//dataSet = append(dataSet, inData)
		returnValue = append(returnValue, inData)
	}
	return returnValue, nil
}

func getCMNodeVNPUsDataFromVNPUAllocInfCache(cacheData VNPUAllocInfCache) ([]NodeVNPUs, error) {
	var nodeData []NodeVNPUs
	var updateErr error
	for _, tmp := range cacheData.Cache {
		if !tmp.AllocFlag {
			continue
		}
		tmpNode := NodeVNPUs{
			NodeName: tmp.NodeName,
			Cards: []CardVNPUs{{
				CardName: tmp.ReqCardName,
				Req:      []string{tmp.ReqNPUType},
				Alloc:    []string{tmp.AllocCardName}}},
		}
		if nodeData, updateErr = updateCMNodeData(nodeData, tmpNode); updateErr != nil {
			klog.V(util.LogErrorLev).Infof("getCMNodeVNPUsDataFromVNPUAllocInfCache %v.", updateErr)
			return nil, updateErr
		}
		fmt.Printf("haha-1%+v===%+v\n", nodeData, tmpNode)
	}
	if len(nodeData) == 0 {
		return nil, errors.New("no configmap data")
	}
	return nodeData, nil
}

func changeVNPUAllocInfCacheToCMData(cacheData VNPUAllocInfCache) (VNPUCM, error) {
	cmData := VNPUCM{}
	nodeData, err := getCMNodeVNPUsDataFromVNPUAllocInfCache(cacheData)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("changeVNPUAllocInfCacheToCMData: %v.", err)
		return VNPUCM{}, nil
	}

	cmData.Nodes = nodeData
	cmData.UpdateTime = time.Now().Unix()
	cmData.CheckCode = util.MakeDataHash(cmData.Nodes)
	return cmData, nil
}

// GetVNPUCMData Get VNPU CM data for write into config map.
func GetVNPUCMData(cacheData VNPUAllocInfCache) map[string]string {
	cmData, changeErr := changeVNPUAllocInfCacheToCMData(cacheData)
	if changeErr != nil {
		klog.V(util.LogErrorLev).Infof("GetVNPUCMData err: %v.", changeErr)
		return nil
	}

	tmp, err := util.MarshalCacheDataToString(cmData)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("marshalCacheDataToString err: %v.", err)
		return nil
	}
	dataBuffer := make(map[string]string, util.ConstIntNum3)
	dataBuffer[VNPCMDataKey] = tmp
	return dataBuffer
}

// GetVNPUCacheCMData get the cache configmap data.
func GetVNPUCacheCMData(cacheData VNPUAllocInfCache) map[string]string {
	tmp, err := util.MarshalCacheDataToString(cacheData)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("marshalCacheDataToString err: %v.", err)
		return nil
	}
	cacheBuffer := make(map[string]string, util.ConstIntNum3)
	cacheBuffer[VNPCMDataKey] = tmp
	return cacheBuffer
}

// IsNPUCardNodeByCardName judge the node has card or not.
func IsNPUCardNodeByCardName(cardName string, tmpNode *api.NodeInfo) bool {
	if tmpNode == nil {
		return false
	}
	// only read
	for resName := range tmpNode.Allocatable.ScalarResources {
		tmp := string(resName)
		if strings.Contains(tmp, cardName) {
			// the VNPU func before has check the resource type.
			return true
		}
	}
	return false
}

// IsNPUResourceStableInNode judge the node resource whether is stable.
func IsNPUResourceStableInNode(kind string, tmpNode *api.NodeInfo) bool {
	if tmpNode == nil {
		klog.V(util.LogErrorLev).Infof("IsNPUResourceStableInNode parameters nil.")
		return false
	}
	k8sNum, k8sOK := tmpNode.Allocatable.ScalarResources[v1.ResourceName(kind)]
	if !k8sOK {
		klog.V(util.LogErrorLev).Infof("IsNPUResourceStableInNode %s no %v in k8s.", tmpNode.Name, kind)
		return false
	}
	annoSrings, anOK := tmpNode.Node.Annotations[kind]
	if !anOK {
		klog.V(util.LogErrorLev).Infof("IsNPUResourceStableInNode %s no %v in annotation.", tmpNode.Name, kind)
		return false
	}
	annoSlice := strings.Split(annoSrings, ",")
	annoNum := len(annoSlice)
	if int(k8sNum/util.NPUHex) != annoNum {
		klog.V(util.LogErrorLev).Infof("%s IsNPUResourceStableInNode  %s not stable (%v != %v).", tmpNode.Name,
			kind, k8sNum, annoNum)
		return false
	}
	return true
}

// IsVJobRunning check whether the job is running or not.
func IsVJobRunning(job *api.JobInfo) bool {
	if len(job.Tasks) > util.ConstIntNum2 {
		klog.V(util.LogInfoLev).Infof("%s has wrong tasks %#v", job.UID, job.Tasks)
		return false
	}
	for _, task := range job.Tasks {
		if task.Pod.Status.Phase != v1.PodRunning {
			klog.V(util.LogInfoLev).Infof("%s's task not running %v", job.UID, task.Pod.Status.Phase)
			return false
		}
	}
	return true
}

// IsVJobPending check whether the job is pending or not.
func IsVJobPending(job *api.JobInfo) bool {
	if len(job.Tasks) > util.ConstIntNum2 {
		klog.V(util.LogErrorLev).Infof("%s has wrong tasks %+v", job.UID, job.Tasks)
		return false
	}
	for _, task := range job.Tasks {
		if task.Pod.Status.Phase != v1.PodPending {
			klog.V(util.LogInfoLev).Infof("%s's task not pending %v", job.UID, task.Pod.Status.Phase)
			return false
		}
	}
	return true
}
