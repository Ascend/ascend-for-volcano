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
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/scheduler-strategy/util"
)

// IsSelectorMeetNode Determines whether the selectors of the task and node are equal.
func IsSelectorMeetNode(task *api.TaskInfo, node *api.NodeInfo, defaultConf, conf map[string]string, _ string) error {

	// task has selector, so node should have
	nodeSelector, errNode := util.GetNodeSelector(node)
	if errNode != nil {
		klog.V(logErrorLev).Infof("GetNodeSelector task[%s] on node(%s) %v.", task.Name, node.Name, errNode)
		return errors.New(nodeNoFitSelectorError)
	}

	if err := util.CheckTaskAndNodeSelectorMeet(defaultConf, nodeSelector, conf); err != nil {
		klog.V(logErrorLev).Infof("CheckTaskAndNodeSelectorMeet %s err:%v.", node.Name, err)
		return err
	}

	return nil
}
