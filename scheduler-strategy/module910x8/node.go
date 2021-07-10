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

Package module910x8 is using for HuaWei A800/9000 Ascend910 pin affinity schedule.

*/
package module910x8

import (
	"errors"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog"
	"reflect"
	"strings"
	time2 "time"
	"volcano.sh/volcano/pkg/scheduler/api"
	hwutil "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/util"
)

type selectNodeInf struct {
	nodeName    string
	allNPUNum   int
	leftNPUNum  int
	rightNPUNum int
}

func initSelectNodeInf(node *api.NodeInfo) selectNodeInf {
	var sNodeInf selectNodeInf
	var leftHccsTop []int
	var rightHccsTop []int

	cardIds := hwutil.GetTopFromNode(node, npu800And9000CardName, npu910CardPreName)
	klog.V(logDebugLev).Infof("%s initPriNodeGroups:%v.", PluginName, cardIds)
	for _, cardID := range cardIds {
		if cardID < npuNumPerHccs {
			leftHccsTop = append(leftHccsTop, cardID)
		} else {
			rightHccsTop = append(rightHccsTop, cardID)
		}
	}
	sNodeInf.leftNPUNum = len(leftHccsTop)
	sNodeInf.rightNPUNum = len(rightHccsTop)
	sNodeInf.allNPUNum = sNodeInf.leftNPUNum + sNodeInf.rightNPUNum

	return sNodeInf
}

// Initializes the priority group of the node.
func initPriNodeGroups(task *api.TaskInfo, nodes []*api.NodeInfo) ([]map[string]*npuPriNodeInf, error) {
	var err error
	var priNodeGroups []map[string]*npuPriNodeInf

	// for pipelined state the node npu is nil
	if len(nodes) == 0 {
		return nil, errors.New("nodes is empty")
	}

	for i := 0; i < npuNumPerHccs; i++ {
		priNodeGroups = append(priNodeGroups, make(map[string]*npuPriNodeInf, 1))
	}

	// init pri Node group
	for _, node := range nodes {
		if reflect.ValueOf(node).IsNil() {
			continue
		}

		sNodeInf := initSelectNodeInf(node)
		// set the meet node into its pri-node-list group
		addPriNodeGroupFn := func(priNodeGroup map[string]*npuPriNodeInf, groupName string) {
			klog.V(logDebugLev).Infof("%s nodeName:%s,group:%v.", PluginName, node.Name, priNodeGroup[node.Name])
			priNodeGroup[node.Name] = &npuPriNodeInf{
				Name:     groupName,
				nodeName: node.Name,
			}
			klog.V(logDebugLev).Infof("%s addPriNodeGroupFn node name:%s priNode:%v.",
				PluginName, node.Name, priNodeGroup[node.Name])
		}

		// insert into group by policy
		err = insertNodeInPriGroup(task, sNodeInf, priNodeGroups, addPriNodeGroupFn)
		if err != nil {
			continue
		}
	}
	return priNodeGroups, nil
}

func getNodeHccsArray(nodeTop []int) ([]int, []int) {
	var leftHccsArray []int
	var rightHccsArray []int

	for _, v := range nodeTop {
		if v < npuNumPerHccs {
			leftHccsArray = append(leftHccsArray, v)
			continue
		}
		rightHccsArray = append(rightHccsArray, v)
	}

	return leftHccsArray, rightHccsArray
}

func getNodeNPUNumFromAnnotation(nodeInfo *api.NodeInfo) (int, error) {
	top := hwutil.GetTopFromNode(nodeInfo, npu800And9000CardName, npu910CardPreName)
	if top == nil {
		return 0, fmt.Errorf("nil node(%s) top", nodeInfo.Name)
	}
	nodeNPUIdleNumFromTop := len(top)

	return nodeNPUIdleNumFromTop, nil
}

func initNodesNPUTopologyFn(nodes map[string]*api.NodeInfo) error {
	for _, node := range nodes {
		if hwutil.IsCardModeNode(node) {
			continue
		}

		topStr, err := hwutil.GetNodeNPUAllocCards(node, npu800And9000CardName)
		if err != nil {
			klog.V(logDebugLev).Infof("%s initNodesFn :%v.", PluginName, err)
			return nil
		}

		node.Others = make(map[string]interface{}, 1)
		err = hwutil.SaveTopologyInMap(node.Others, topStr, npu800And9000CardName)
		if err != nil {
			return err
		}
	}

	return nil
}

func getNodeFaultNPUs(node *api.NodeInfo) ([]string, error) {
	npuStrings, ok := node.Node.Annotations[faultNPU]
	if !ok || len(npuStrings) == 0 {
		return nil, fmt.Errorf("%s get nil npus", node.Name)
	}

	faultNPUs := strings.Split(npuStrings, ",")
	if len(faultNPUs) > nodeNPUNumber {
		return nil, fmt.Errorf("%s get fault npus(%d)", node.Name, len(faultNPUs))
	}

	return faultNPUs, nil
}

func getInoperableNPUs(nodes map[string]*api.NodeInfo) ([]nodeFaultNPUs, error) {
	var faultNPUs []nodeFaultNPUs
	for _, nodeInfo := range nodes {
		npus, err := getNodeFaultNPUs(nodeInfo)
		if err != nil {
			klog.V(logDebugLev).Infof("%s getNodeFaultNPUs err:%v.", PluginName, err)
			continue
		}
		faultNPUs = append(faultNPUs, nodeFaultNPUs{nodeInfo.Name, npus})
	}

	if len(faultNPUs) == 0 {
		return nil, errors.New("nil inoperable NPU")
	}
	klog.V(logDebugLev).Infof("%s getNodeFaultNPUs %+v.", PluginName, faultNPUs)

	return faultNPUs, nil
}

func getFaultNodePODAndRankIndex(job *api.JobInfo, nodes map[string]*v1.Pod) (faultNPUJob, error) {
	var faultJob = faultNPUJob{
		jobName:          job.Name,
		namespace:        job.Namespace,
		taskUseRankIndex: make(map[string]string, constIntNum3),
		taskUseNode:      make(map[string]string, constIntNum3),
	}

	for _, task := range job.Tasks {
		if pod, ok := nodes[task.NodeName]; ok {
			rankIndex, err := getPodRankIndex(pod)
			if err != nil {
				klog.V(logErrorLev).Infof("%s getPodRankIndex %s %v.", PluginName, pod.Name, err)
				return faultJob, err
			}
			faultJob.taskUseRankIndex[task.Name] = rankIndex
			faultJob.taskUseNode[task.Name] = task.NodeName
		}
	}

	if len(faultJob.taskUseRankIndex) == 0 {
		return faultJob, errors.New("get nil rankIndex")
	}

	return faultJob, nil
}

func setFaultLabelOnNodeAndJob(faultNPUJobs []faultNPUJob, jobs map[string]*api.JobInfo) error {
	for _, tmpFaultNPUJob := range faultNPUJobs {
		tmpTask := hwutil.ReSchedulerTasks{
			NodeNames:   make(map[string]string, constIntNum3),
			RankIndexes: make(map[string]string, constIntNum3),
			Time:        make(map[string]int64, constIntNum3),
			NameSpace:   tmpFaultNPUJob.namespace}

		for taskName, nodeName := range tmpFaultNPUJob.taskUseNode {
			rankIndex, indexOK := tmpFaultNPUJob.taskUseRankIndex[taskName]
			if !indexOK {
				klog.V(logErrorLev).Infof("%s %s get rankIndex failed.", PluginName, taskName)
				continue
			}

			tmpTask.NodeNames[taskName] = nodeName
			tmpTask.RankIndexes[taskName] = rankIndex
			tmpTask.Time[taskName] = time2.Now().Unix()
		}

		if err := recordNPUFaultJobToBuffer(jobs, tmpFaultNPUJob, tmpTask); err != nil {
			klog.V(logErrorLev).Infof("%s recordNPUFaultJobToBuffer :%v.", PluginName, err)
			return err
		}
	}

	return nil
}
