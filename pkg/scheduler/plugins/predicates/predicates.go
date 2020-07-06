/*
Copyright 2018 The Kubernetes Authors.

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

package predicates

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins/interpodaffinity"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins/nodeaffinity"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins/nodeports"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins/nodeunschedulable"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins/tainttoleration"
	k8sframework "k8s.io/kubernetes/pkg/scheduler/framework/v1alpha1"

	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"
	"volcano.sh/volcano/pkg/scheduler/plugins/util"
)

const (
	// PluginName indicates name of volcano scheduler plugin.
	PluginName = "predicates"

	// GPUSharingPredicate is the key for enabling GPU Sharing Predicate in YAML
	GPUSharingPredicate = "predicate.GPUSharingEnable"
)

type predicatesPlugin struct {
	// Arguments given for the plugin
	pluginArguments framework.Arguments
}

// New return predicate plugin
func New(arguments framework.Arguments) framework.Plugin {
	return &predicatesPlugin{pluginArguments: arguments}
}

func (pp *predicatesPlugin) Name() string {
	return PluginName
}

type predicateEnable struct {
	gpuSharingEnable bool
}

func enablePredicate(args framework.Arguments) predicateEnable {

	/*
		   User Should give predicatesEnable in this format(predicate.GPUSharingEnable).
		   Currently supported only GPUSharing predicate checks.

		   actions: "reclaim, allocate, backfill, preempt"
		   tiers:
		   - plugins:
		     - name: priority
		     - name: gang
		     - name: conformance
		   - plugins:
		     - name: drf
		     - name: predicates
		       arguments:
				 predicate.GPUSharingEnable: true
		     - name: proportion
		     - name: nodeorder
	*/

	predicate := predicateEnable{
		gpuSharingEnable: false,
	}

	// Checks whether predicate.GPUSharingEnable is provided or not, if given, modifies the value in predicateEnable struct.
	args.GetBool(&predicate.gpuSharingEnable, GPUSharingPredicate)

	return predicate
}

func (pp *predicatesPlugin) OnSessionOpen(ssn *framework.Session) {
	pl := util.NewPodLister(ssn)
	pods, _ := pl.List(labels.NewSelector())
	nodeMap, nodeSlice := util.GenerateNodeMapAndSlice(ssn.Nodes)

	predicate := enablePredicate(pp.pluginArguments)

	kubeClient := ssn.KubeClient()
	// Register event handlers to update task info in PodLister & nodeMap
	ssn.AddEventHandler(&framework.EventHandler{
		AllocateFunc: func(event *framework.Event) {
			pod := pl.UpdateTask(event.Task, event.Task.NodeName)

			nodeName := event.Task.NodeName
			node, found := nodeMap[nodeName]
			if !found {
				klog.Errorf("predicates, update pod %s/%s allocate to NOT EXIST node [%s]", pod.Namespace, pod.Name, nodeName)
				return
			}

			if predicate.gpuSharingEnable && api.GetGPUResourceOfPod(pod) > 0 {
				nodeInfo, _ := ssn.Nodes[nodeName]
				id := predicateGPU(pod, nodeInfo)
				if id < 0 {
					klog.Errorf("The node %s can't place the pod %s in ns %s", pod.Spec.NodeName, pod.Name, pod.Namespace)
					return
				}
				patch := api.AddGPUIndexPatch(id)
				pod, err := kubeClient.CoreV1().Pods(pod.Namespace).Patch(context.TODO(), pod.Name, types.JSONPatchType, []byte(patch), metav1.PatchOptions{})
				if err != nil {
					klog.Errorf("Patch pod %s failed with patch %s: %v", pod.Name, patch, err)
					return
				}
				dev, _ := nodeInfo.GPUDevices[id]
				dev.PodMap[string(pod.UID)] = pod
				klog.V(4).Infof("predicates with gpu sharing, update pod %s/%s allocate to node [%s]", pod.Namespace, pod.Name, nodeName)
			}

			node.AddPod(pod)
			klog.V(4).Infof("predicates, update pod %s/%s allocate to node [%s]", pod.Namespace, pod.Name, nodeName)
		},
		DeallocateFunc: func(event *framework.Event) {
			pod := pl.UpdateTask(event.Task, "")
			nodeName := event.Task.NodeName
			node, found := nodeMap[nodeName]
			if !found {
				klog.Errorf("predicates, update pod %s/%s allocate from NOT EXIST node [%s]", pod.Namespace, pod.Name, nodeName)
				return
			}

			if predicate.gpuSharingEnable && api.GetGPUResourceOfPod(pod) > 0 {
				// deallocate pod gpu id
				id := api.GetGPUIndex(pod)
				patch := api.RemoveGPUIndexPatch()
				_, err := kubeClient.CoreV1().Pods(pod.Namespace).Patch(context.TODO(), pod.Name, types.JSONPatchType, []byte(patch), metav1.PatchOptions{})
				if err != nil {
					klog.Errorf("Patch pod %s failed with patch %s: %v", pod.Name, patch, err)
					return
				}

				nodeInfo, _ := ssn.Nodes[nodeName]
				if dev, ok := nodeInfo.GPUDevices[id]; ok {
					delete(dev.PodMap, string(pod.UID))
				}

				klog.V(4).Infof("predicates with gpu sharing, update pod %s/%s deallocate from node [%s]", pod.Namespace, pod.Name, nodeName)
			}

			node.RemovePod(pod)
			klog.V(4).Infof("predicates, update pod %s/%s deallocate from node [%s]", pod.Namespace, pod.Name, nodeName)

		},
	})

	// Initialize k8s plugins
	// TODO: Add more predicates, k8s.io/kubernetes/pkg/scheduler/framework/plugins/legacy_registry.go
	handle := framework.NewFrameworkHandle(pods, nodeSlice)
	// 1. NodeUnschedulable
	plugin, _ := nodeunschedulable.New(nil, handle)
	nodeUnscheduleFilter := plugin.(*nodeunschedulable.NodeUnschedulable)
	// 2. NodeAffinity
	plugin, _ = nodeaffinity.New(nil, handle)
	nodeAffinityFilter := plugin.(*nodeaffinity.NodeAffinity)
	// 3. NodePorts
	plugin, _ = nodeports.New(nil, handle)
	nodePortFilter := plugin.(*nodeports.NodePorts)
	// 4. TaintToleration
	plugin, _ = tainttoleration.New(nil, handle)
	tolerationFilter := plugin.(*tainttoleration.TaintToleration)
	// 5. InterPodAffinity
	plugin, _ = interpodaffinity.New(nil, handle)
	podAffinityFilter := plugin.(*interpodaffinity.InterPodAffinity)

	ssn.AddPredicateFn(pp.Name(), func(task *api.TaskInfo, node *api.NodeInfo) error {
		nodeInfo, found := nodeMap[node.Name]
		if !found {
			fmt.Errorf("failed to predicates, node info for %s not found", node.Name)
		}

		if node.Allocatable.MaxTaskNum <= len(nodeInfo.Pods()) {
			klog.V(4).Infof("NodePodNumber predicates Task <%s/%s> on Node <%s> failed",
				task.Namespace, task.Name, node.Name)
			return api.NewFitError(task, node, api.NodePodNumberExceeded)
		}

		state := k8sframework.NewCycleState()
		// CheckNodeUnschedulable
		status := nodeUnscheduleFilter.Filter(context.TODO(), state, task.Pod, nodeInfo)
		if !status.IsSuccess() {
			return fmt.Errorf("plugin %s predicates failed %s", nodeunschedulable.Name, status.Message())
		}

		// Check NodeAffinity
		status = nodeAffinityFilter.Filter(context.TODO(), state, task.Pod, nodeInfo)
		if !status.IsSuccess() {
			return fmt.Errorf("plugin %s predicates failed %s", nodeaffinity.Name, status.Message())
		}

		// Check NodePorts
		nodePortFilter.PreFilter(context.TODO(), state, task.Pod)
		status = nodePortFilter.Filter(context.TODO(), state, nil, nodeInfo)
		if !status.IsSuccess() {
			return fmt.Errorf("plugin %s predicates failed %s", nodeaffinity.Name, status.Message())
		}

		// PodToleratesNodeTaints: TaintToleration
		status = tolerationFilter.Filter(context.TODO(), state, task.Pod, nodeInfo)
		if !status.IsSuccess() {
			return fmt.Errorf("plugin %s predicates failed %s", tainttoleration.Name, status.Message())
		}

		// InterPodAffinity Predicate
		status = podAffinityFilter.PreFilter(context.TODO(), state, task.Pod)
		if !status.IsSuccess() {
			return fmt.Errorf("plugin %s pre-predicates failed %s", interpodaffinity.Name, status.Message())
		}

		status = podAffinityFilter.Filter(context.TODO(), state, task.Pod, nodeInfo)
		if !status.IsSuccess() {
			return fmt.Errorf("plugin %s predicates failed %s", interpodaffinity.Name, status.Message())
		}

		if predicate.gpuSharingEnable {
			// CheckGPUSharingPredicate
			fit, err := checkNodeGPUSharingPredicate(task.Pod, node)
			if err != nil {
				return err
			}

			klog.V(4).Infof("checkNodeGPUSharingPredicate predicates Task <%s/%s> on Node <%s>: fit %v",
				task.Namespace, task.Name, node.Name, fit)
		}

		return nil
	})
}

func (pp *predicatesPlugin) OnSessionClose(ssn *framework.Session) {}
