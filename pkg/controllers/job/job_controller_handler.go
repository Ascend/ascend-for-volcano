/*
Copyright 2017 The Volcano Authors.

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

package job

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"volcano.sh/volcano/pkg/apis/batch/v1alpha1"

	v1 "k8s.io/api/core/v1"
	"k8s.io/api/scheduling/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
	batch "volcano.sh/volcano/pkg/apis/batch/v1alpha1"
	bus "volcano.sh/volcano/pkg/apis/bus/v1alpha1"
	"volcano.sh/volcano/pkg/apis/helpers"
	scheduling "volcano.sh/volcano/pkg/apis/scheduling/v1beta1"
	"volcano.sh/volcano/pkg/controllers/apis"
	jobcache "volcano.sh/volcano/pkg/controllers/cache"
	jobhelpers "volcano.sh/volcano/pkg/controllers/job/helpers"
)

func (cc *jobcontroller) addCommand(obj interface{}) {
	cmd, ok := obj.(*bus.Command)
	if !ok {
		klog.Errorf("obj is not Command")
		return
	}

	cc.commandQueue.Add(cmd)
}

func (cc *jobcontroller) addJob(obj interface{}) {
	job, ok := obj.(*batch.Job)
	if !ok {
		klog.Errorf("obj is not Job")
		return
	}

	req := apis.Request{
		Namespace: job.Namespace,
		JobName:   job.Name,

		Event: bus.OutOfSyncEvent,
	}

	// TODO(k82cn): if failed to add job, the cache should be refresh
	if err := cc.cache.Add(job); err != nil {
		klog.Errorf("Failed to add job <%s/%s>: %v in cache",
			job.Namespace, job.Name, err)
	}
	key := jobhelpers.GetJobKeyByReq(&req)
	queue := cc.getWorkerQueue(key)
	queue.Add(req)
}

func (cc *jobcontroller) updateJob(oldObj, newObj interface{}) {
	newJob, ok := newObj.(*batch.Job)
	if !ok {
		klog.Errorf("newObj is not Job")
		return
	}

	oldJob, ok := oldObj.(*batch.Job)
	if !ok {
		klog.Errorf("oldJob is not Job")
		return
	}

	// No need to update if ResourceVersion is not changed
	if newJob.ResourceVersion == oldJob.ResourceVersion {
		klog.V(6).Infof("No need to update because job is not modified.")
		return
	}

	if err := cc.cache.Update(newJob); err != nil {
		klog.Errorf("UpdateJob - Failed to update job <%s/%s>: %v in cache",
			newJob.Namespace, newJob.Name, err)
	}

	// NOTE: Since we only reconcile job based on Spec, we will ignore other attributes
	// For Job status, it's used internally and always been updated via our controller.
	if reflect.DeepEqual(newJob.Spec, oldJob.Spec) && newJob.Status.State.Phase == oldJob.Status.State.Phase {
		klog.V(6).Infof("Job update event is ignored since no update in 'Spec'.")
		return
	}

	req := apis.Request{
		Namespace: newJob.Namespace,
		JobName:   newJob.Name,
		Event:     bus.OutOfSyncEvent,
	}
	key := jobhelpers.GetJobKeyByReq(&req)
	queue := cc.getWorkerQueue(key)
	queue.Add(req)
}

func (cc *jobcontroller) deleteJob(obj interface{}) {
	job, ok := obj.(*batch.Job)
	if !ok {
		// If we reached here it means the Job was deleted but its final state is unrecorded.
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			klog.Errorf("Couldn't get object from tombstone %#v", obj)
			return
		}
		job, ok = tombstone.Obj.(*batch.Job)
		if !ok {
			klog.Errorf("Tombstone contained object that is not a volcano Job: %#v", obj)
			return
		}
	}

	if err := cc.cache.Delete(job); err != nil {
		klog.Errorf("Failed to delete job <%s/%s>: %v in cache",
			job.Namespace, job.Name, err)
	}
}

func (cc *jobcontroller) addPod(obj interface{}) {
	pod, ok := obj.(*v1.Pod)
	if !ok {
		klog.Errorf("Failed to convert %v to v1.Pod", obj)
		return
	}
	// Filter out pods that are not created from volcano job
	if !isControlledBy(pod, helpers.JobKind) {
		return
	}

	jobName, found := pod.Annotations[batch.JobNameKey]
	if !found {
		klog.Infof("Failed to find jobName of Pod <%s/%s>, skipping",
			pod.Namespace, pod.Name)
		return
	}

	version, found := pod.Annotations[batch.JobVersion]
	if !found {
		klog.Infof("Failed to find jobVersion of Pod <%s/%s>, skipping",
			pod.Namespace, pod.Name)
		return
	}

	dVersion, err := strconv.Atoi(version)
	if err != nil {
		klog.Infof("Failed to convert jobVersion of Pod <%s/%s> into number, skipping",
			pod.Namespace, pod.Name)
		return
	}

	if pod.DeletionTimestamp != nil {
		cc.deletePod(pod)
		return
	}

	req := apis.Request{
		Namespace: pod.Namespace,
		JobName:   jobName,

		Event:      bus.OutOfSyncEvent,
		JobVersion: int32(dVersion),
	}

	if err := cc.cache.AddPod(pod); err != nil {
		klog.Errorf("Failed to add Pod <%s/%s>: %v to cache",
			pod.Namespace, pod.Name, err)
	}
	key := jobhelpers.GetJobKeyByReq(&req)
	queue := cc.getWorkerQueue(key)
	queue.Add(req)
}

func (cc *jobcontroller) updatePod(oldObj, newObj interface{}) {
	oldPod, ok := oldObj.(*v1.Pod)
	if !ok {
		klog.Errorf("Failed to convert %v to v1.Pod", oldObj)
		return
	}

	newPod, ok := newObj.(*v1.Pod)
	if !ok {
		klog.Errorf("Failed to convert %v to v1.Pod", newObj)
		return
	}

	// Filter out pods that are not created from volcano job
	if !isControlledBy(newPod, helpers.JobKind) {
		return
	}

	if newPod.ResourceVersion == oldPod.ResourceVersion {
		return
	}

	if newPod.DeletionTimestamp != nil {
		cc.deletePod(newObj)
		return
	}

	taskName, found := newPod.Annotations[batch.TaskSpecKey]
	if !found {
		klog.Infof("Failed to find taskName of Pod <%s/%s>, skipping",
			newPod.Namespace, newPod.Name)
		return
	}

	jobName, found := newPod.Annotations[batch.JobNameKey]
	if !found {
		klog.Infof("Failed to find jobName of Pod <%s/%s>, skipping",
			newPod.Namespace, newPod.Name)
		return
	}

	version, found := newPod.Annotations[batch.JobVersion]
	if !found {
		klog.Infof("Failed to find jobVersion of Pod <%s/%s>, skipping",
			newPod.Namespace, newPod.Name)
		return
	}

	dVersion, err := strconv.Atoi(version)
	if err != nil {
		klog.Infof("Failed to convert jobVersion of Pod into number <%s/%s>, skipping",
			newPod.Namespace, newPod.Name)
		return
	}

	if err := cc.cache.UpdatePod(newPod); err != nil {
		klog.Errorf("Failed to update Pod <%s/%s>: %v in cache",
			newPod.Namespace, newPod.Name, err)
	}

	event := bus.OutOfSyncEvent
	var exitCode int32
	if oldPod.Status.Phase != v1.PodFailed &&
		newPod.Status.Phase == v1.PodFailed {
		event = bus.PodFailedEvent
		// TODO: currently only one container pod is supported by volcano
		// Once multi containers pod is supported, update accordingly.
		if len(newPod.Status.ContainerStatuses) > 0 && newPod.Status.ContainerStatuses[0].State.Terminated != nil {
			exitCode = newPod.Status.ContainerStatuses[0].State.Terminated.ExitCode
		}
	}

	if oldPod.Status.Phase != v1.PodSucceeded &&
		newPod.Status.Phase == v1.PodSucceeded {
		if cc.cache.TaskCompleted(jobcache.JobKeyByName(newPod.Namespace, jobName), taskName) {
			event = bus.TaskCompletedEvent
		}
	}

	req := apis.Request{
		Namespace: newPod.Namespace,
		JobName:   jobName,
		TaskName:  taskName,

		Event:      event,
		ExitCode:   exitCode,
		JobVersion: int32(dVersion),
	}

	key := jobhelpers.GetJobKeyByReq(&req)
	queue := cc.getWorkerQueue(key)
	queue.Add(req)
}

func (cc *jobcontroller) deletePod(obj interface{}) {
	pod, ok := obj.(*v1.Pod)
	if !ok {
		// If we reached here it means the pod was deleted but its final state is unrecorded.
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			klog.Errorf("Couldn't get object from tombstone %#v", obj)
			return
		}
		pod, ok = tombstone.Obj.(*v1.Pod)
		if !ok {
			klog.Errorf("Tombstone contained object that is not a Pod: %#v", obj)
			return
		}
	}

	// Filter out pods that are not created from volcano job
	if !isControlledBy(pod, helpers.JobKind) {
		return
	}

	taskName, found := pod.Annotations[batch.TaskSpecKey]
	if !found {
		klog.Infof("Failed to find taskName of Pod <%s/%s>, skipping",
			pod.Namespace, pod.Name)
		return
	}

	jobName, found := pod.Annotations[batch.JobNameKey]
	if !found {
		klog.Infof("Failed to find jobName of Pod <%s/%s>, skipping",
			pod.Namespace, pod.Name)
		return
	}

	version, found := pod.Annotations[batch.JobVersion]
	if !found {
		klog.Infof("Failed to find jobVersion of Pod <%s/%s>, skipping",
			pod.Namespace, pod.Name)
		return
	}

	dVersion, err := strconv.Atoi(version)
	if err != nil {
		klog.Infof("Failed to convert jobVersion of Pod <%s/%s> into number, skipping",
			pod.Namespace, pod.Name)
		return
	}

	req := apis.Request{
		Namespace: pod.Namespace,
		JobName:   jobName,
		TaskName:  taskName,

		Event:      bus.PodEvictedEvent,
		JobVersion: int32(dVersion),
	}

	if err := cc.cache.DeletePod(pod); err != nil {
		klog.Errorf("Failed to delete Pod <%s/%s>: %v in cache",
			pod.Namespace, pod.Name, err)
	}

	key := jobhelpers.GetJobKeyByReq(&req)
	queue := cc.getWorkerQueue(key)
	queue.Add(req)
}

func (cc *jobcontroller) recordJobEvent(namespace, name string, event batch.JobEvent, message string) {
	job, err := cc.cache.Get(jobcache.JobKeyByName(namespace, name))
	if err != nil {
		klog.Warningf("Failed to find job in cache when reporting job event <%s/%s>: %v",
			namespace, name, err)
		return
	}
	cc.recorder.Event(job.Job, v1.EventTypeNormal, string(event), message)

}

func (cc *jobcontroller) handleCommands() {
	for cc.processNextCommand() {
	}
}

func (cc *jobcontroller) processNextCommand() bool {
	obj, shutdown := cc.commandQueue.Get()
	if shutdown {
		return false
	}
	cmd := obj.(*bus.Command)
	defer cc.commandQueue.Done(cmd)

	if err := cc.vcClient.BusV1alpha1().Commands(cmd.Namespace).Delete(context.TODO(), cmd.Name, metav1.DeleteOptions{}); err != nil {
		if !apierrors.IsNotFound(err) {
			klog.Errorf("Failed to delete Command <%s/%s>.", cmd.Namespace, cmd.Name)
			cc.commandQueue.AddRateLimited(cmd)
		}
		return true
	}
	cc.recordJobEvent(cmd.Namespace, cmd.TargetObject.Name,
		batch.CommandIssued,
		fmt.Sprintf(
			"Start to execute command %s, and clean it up to make sure executed not more than once.", cmd.Action))
	req := apis.Request{
		Namespace: cmd.Namespace,
		JobName:   cmd.TargetObject.Name,
		Event:     bus.CommandIssuedEvent,
		Action:    bus.Action(cmd.Action),
	}

	key := jobhelpers.GetJobKeyByReq(&req)
	queue := cc.getWorkerQueue(key)
	queue.Add(req)

	return true
}

func (cc *jobcontroller) updatePodGroup(oldObj, newObj interface{}) {
	oldPG, ok := oldObj.(*scheduling.PodGroup)
	if !ok {
		klog.Errorf("Failed to convert %v to PodGroup", newObj)
		return
	}

	newPG, ok := newObj.(*scheduling.PodGroup)
	if !ok {
		klog.Errorf("Failed to convert %v to PodGroup", newObj)
		return
	}

	_, err := cc.cache.Get(jobcache.JobKeyByName(newPG.Namespace, newPG.Name))
	if err != nil && newPG.Annotations != nil {
		klog.Warningf(
			"Failed to find job in cache by PodGroup, this may not be a PodGroup for volcano job.")
	}

	if errJob := cc.setJobFailedFromScheduler(newPG); errJob == nil {
		klog.V(3).Infof("Update job(%v) in pg ok", newPG.Name)
		return
	}

	if newPG.Status.Phase != oldPG.Status.Phase {
		req := apis.Request{
			Namespace: newPG.Namespace,
			JobName:   newPG.Name,
		}
		switch newPG.Status.Phase {
		case scheduling.PodGroupUnknown:
			req.Event = bus.JobUnknownEvent
		}
		key := jobhelpers.GetJobKeyByReq(&req)
		queue := cc.getWorkerQueue(key)
		queue.Add(req)
	}
}

func (cc *jobcontroller) setJobFailedFromScheduler(newPg *scheduling.PodGroup) error {
	if v, ok := newPg.Labels["Failed"]; ok {
		if v == "scheduler" {
			cc.updatePodGroupStatus(newPg)
			return nil
		}
	}
	return errors.New("no Labels need set")
}

func (cc *jobcontroller) updatePodGroupStatus(newPg *scheduling.PodGroup) {
	jobIn, err := cc.cache.Get(jobcache.JobKeyByName(newPg.Namespace, newPg.Name))
	if err == nil {
		klog.V(3).Infof("try make job(%s) failed", newPg.Name)
		jobIn.Job.Status.State.Phase = v1alpha1.Failed
		for _, cond := range newPg.Status.Conditions {
			if cond.Type == "ScheduledFailed" {
				jobIn.Job.Status.State.Reason = cond.Message
			}
		}
		cc.updateJobStatus(jobIn.Job)
	}
}

// TODO(k82cn): add handler for PodGroup unschedulable event.

func (cc *jobcontroller) addPriorityClass(obj interface{}) {
	pc := convert2PriorityClass(obj)
	if pc == nil {
		return
	}

	cc.Mutex.Lock()
	defer cc.Mutex.Unlock()

	cc.priorityClasses[pc.Name] = pc
}

func (cc *jobcontroller) deletePriorityClass(obj interface{}) {
	pc := convert2PriorityClass(obj)
	if pc == nil {
		return
	}

	cc.Mutex.Lock()
	defer cc.Mutex.Unlock()

	delete(cc.priorityClasses, pc.Name)
}

func convert2PriorityClass(obj interface{}) *v1beta1.PriorityClass {
	var pc *v1beta1.PriorityClass
	switch t := obj.(type) {
	case *v1beta1.PriorityClass:
		pc = t
	case cache.DeletedFinalStateUnknown:
		var ok bool
		pc, ok = t.Obj.(*v1beta1.PriorityClass)
		if !ok {
			klog.Errorf("Cannot convert to *v1beta1.PriorityClass: %v", t.Obj)
			return nil
		}
	default:
		klog.Errorf("Cannot convert to *v1beta1.PriorityClass: %v", t)
		return nil
	}

	return pc
}
