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
	"volcano.sh/volcano/pkg/apis/scheduling"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"
)

type topology910plugin struct {
	// Arguments given for the plugin
	pluginArguments framework.Arguments
}

// for frame init
func (tp *topology910plugin) Name() string {
	return PluginName
}

// New return npu plugin
func New(arguments framework.Arguments) framework.Plugin {
	return &topology910plugin{pluginArguments: arguments}
}

func initNpuSession(ssn *framework.Session) {
	// init npu node top in other, for concurrency-oriented allocation
	initNodesNpuAllocTopology(ssn.Nodes)
	// init job podGroup Labels, for recored and changed job status.
	initPgLabels(ssn.Jobs)
}

// topology910plugin's init session for frame
func (tp *topology910plugin) OnSessionOpen(ssn *framework.Session) {
	// init npu plugin
	initNpuSession(ssn)
	// check job npu resource, if illegal return failed
	ssn.AddJobValidFn(tp.Name(), func(obj interface{}) *api.ValidateResult {
		return validJobFn(obj, ssn.Configurations)
	})
	// if npu no meet the task require,the task will failed.so need to intercept in advance
	ssn.AddPredicateFn(tp.Name(), func(taskInfo *api.TaskInfo, nodeInfo *api.NodeInfo) error {
		return nodePredicate(taskInfo, nodeInfo, ssn.Configurations)
	})
	// The job who has below or equal 8 NPU,only has one pod. If over, every pod has 8s NPU.
	ssn.AddBatchNodeOrderFn(tp.Name(), func(task *api.TaskInfo, nodes []*api.NodeInfo) (map[string]float64, error) {
		score, err := batchNodeOrderFn(task, nodes)
		if err != nil {
			if setErr := setJobFailed(ssn.Jobs[task.Job], err); setErr != nil {
				klog.V(logErrorLev).Infof("%s setJobFailed err:%v", PluginName, setErr)
			}
		}
		return score, nil
	})
	// Register event handlers to update task info in PodLister & nodeMap
	// for support Concurrency
	ssn.AddEventHandler(&framework.EventHandler{
		AllocateFunc: func(event *framework.Event) {
			npuAllocateFunc(event, ssn.Nodes)
		},
		DeallocateFunc: func(event *framework.Event) {
			npuDeallocateFunc(event, ssn.Nodes)
		},
	})
}

// topology910plugin's init session for frame
func (tp *topology910plugin) OnSessionClose(ssn *framework.Session) {
	// for recording job's unscheduled reason; and update job statue
	// write pod info which need trance to other MindX DL component. The info like requestId
	for _, job := range ssn.Jobs {
		// deal pending job
		if job.PodGroup.Status.Phase == scheduling.PodGroupInqueue {
			// if all nodes not meet job require failed
			setJobFailedByNodesCase(ssn.Nodes, job)
		}
	}
}
