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
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	vapi "volcano.sh/volcano/pkg/scheduler/api"
)

var (
	index = 0
)

// TestMNPUGetFaultNPUJobs
func TestMNPUGetFaultNPUJobs(t *testing.T) {
	Convey("Test module910x8 getFaultNPUJobs", t, func() {
		const (
			annotation = "Ascend910-0,Ascend910-1,Ascend910-2,Ascend910-3," +
				"Ascend910-4,Ascend910-5,Ascend910-6,Ascend910-7"
			firstNodeName  = "testNode1"
			secondNodeName = "testNode2"
			firstIndex     = "0"
			secondIndex    = "1"
		)
		faultNPUs := []nodeFaultNPUs{
			{"testNode1", []string{"Ascend910-3", "Ascend910-5"}},
			{"testNode2", []string{"Ascend910-0"}},
		}

		job := buildJobWithTwoTasks(firstNodeName, secondNodeName, annotation, annotation, firstIndex, secondIndex)
		jobs := map[string]*vapi.JobInfo{
			job.Name: job,
		}
		Convey("getFaultNPUJobs() should return", func() {
			var expectedResult []faultNPUJob
			taskKey1, taskKey2 := vapi.TaskID("default-npu-test-M-job-1"), vapi.TaskID("default-npu-test-M-job-2")
			expectedResult = append(expectedResult, faultNPUJob{job.Name, job.Namespace,
				map[string]string{
					job.Tasks[taskKey1].Name: firstIndex,
					job.Tasks[taskKey2].Name: secondIndex,
				},
				map[string]string{
					job.Tasks[taskKey1].Name: firstNodeName,
					job.Tasks[taskKey2].Name: secondNodeName,
				}, map[string]string{
					job.Tasks[taskKey1].Name: annotation,
					job.Tasks[taskKey2].Name: annotation,
				}})
			res, err := getFaultNPUJobs(jobs, faultNPUs)
			So(err, ShouldBeNil)
			So(res, ShouldResemble, expectedResult)
		})
	})
}

func buildJobWithTwoTasks(firstNodeName, secondNodeName, firstAnnotation, secondAnnotation, firstIndex, secondIndex string) *vapi.JobInfo {
	index++
	uid := vapi.JobID(fmt.Sprintf("xxxxxxxx-xxxx-x1xx-xxxx-xxxxxxxxxxx%d", index))
	firstTask := vapi.NewTaskInfo(buildNPUPod(MPodInfo{namespace: "default",
		groupName: fmt.Sprintf("group-M-job-%d", index),
		podName:   fmt.Sprintf("npu-test-M-job-%d", index),
		nodeName:  firstNodeName, reqCPUNum: "10", reqMem: "10Gi",
		reqNPUType: npu800And9000CardName, reqNpuNum: "8"}))
	firstTask.Pod.Annotations = map[string]string{
		npu800And9000CardName: firstAnnotation,
		podRankIndex:          firstIndex,
	}
	index++
	secondTask := vapi.NewTaskInfo(buildNPUPod(MPodInfo{namespace: "default",
		groupName: fmt.Sprintf("group-M-job-%d", index),
		podName:   fmt.Sprintf("npu-test-M-job-%d", index),
		nodeName:  secondNodeName, reqCPUNum: "15", reqMem: "15Gi",
		reqNPUType: npu800And9000CardName, reqNpuNum: "8"}))
	secondTask.Pod.Annotations = map[string]string{
		npu800And9000CardName: secondAnnotation,
		podRankIndex:          secondIndex,
	}
	tasks := []*vapi.TaskInfo{firstTask, secondTask}

	job := vapi.NewJobInfo(uid, tasks...)
	job.PodGroup = &vapi.PodGroup{}
	job.PodGroup.Status.Phase = "Running"

	return job
}
