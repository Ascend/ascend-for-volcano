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

Package topology910 test  for topology910 test.

*/

package topology910

import (
	"errors"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
	"math/rand"
	"strconv"
	"strings"
	"testing"
	schedulingv1 "volcano.sh/volcano/pkg/apis/scheduling/v1beta1"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/cache"
	"volcano.sh/volcano/pkg/scheduler/conf"
	"volcano.sh/volcano/pkg/scheduler/framework"
	"volcano.sh/volcano/pkg/scheduler/util"
)

const (
	bufferSize     = 100
	PodNamePrefix  = "p"
	osArchX86      = "x86"
	osArchArm      = "arm"
	labelSize      = 8
	annSize        = 8
	nodePoolSize   = 8
	podNameLength  = 7
	arraySize      = 2
	nodeNpuNumber7 = 7
	nodeNpuNumber6 = 6
)

type testCaseStruct struct {
	podGroups []*schedulingv1.PodGroup
	pods      []*v1.Pod
	nodes     []*api.NodeInfo
	queues    []*schedulingv1.Queue
	arguments framework.Arguments
	expected  interface{}
	name      string
}

type NodeAllocate struct {
	nodeName       string
	npuAllocateNum int
}

type (
	NodeNameList = []string
)

type testNodeInfo struct {
	nodeName       string
	nodeArch       string
	cpu, mem       string
	npuAllocateNum string
	npuTop         string
}

type testPodInfo struct {
	namespace  string
	groupName  string
	podName    string
	nodeName   string
	reqCpuNum  string
	reqMem     string
	npuNeedNum string
}

// buildNpuNodeResourceList builts resource list object
func buildNpuResourceList(cpu string, memory string, npu string) v1.ResourceList {
	npuNum, err := strconv.Atoi(npu)
	if err != nil {
		return nil
	}

	if npuNum == 0 {
		return v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse(cpu),
			v1.ResourceMemory: resource.MustParse(memory),
		}
	}

	return v1.ResourceList{
		v1.ResourceCPU:    resource.MustParse(cpu),
		v1.ResourceMemory: resource.MustParse(memory),
		npu910CardName:    resource.MustParse(npu),
	}
}

func addNodeNpuTop(Annotations map[string]string, top string) {
	Annotations[npu910CardName] = top
}

func addNodeSelector(node *v1.Node, nodeArch string) error {
	if nodeArch == osArchArm {
		node.Labels[archSelector] = huaweiArchArm
		return nil
	}

	if nodeArch == osArchX86 {
		node.Labels[archSelector] = osArchArm
		return nil
	}

	return fmt.Errorf("unKnown nodeArch %s", nodeArch)
}

// BuildNode builts npu node object
func buildNpuNode(nodeInfo testNodeInfo) *api.NodeInfo {
	nodeCapacity := buildNpuResourceList(nodeInfo.cpu, nodeInfo.mem, strconv.Itoa(nodeNpuNumber))
	nodeAlloc := buildNpuResourceList(nodeInfo.cpu, nodeInfo.mem, nodeInfo.npuAllocateNum)
	labels := make(map[string]string, labelSize)
	ann := make(map[string]string, annSize)

	v1node := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:        nodeInfo.nodeName,
			Labels:      labels,
			Annotations: ann,
		},
		Status: v1.NodeStatus{
			Capacity:    nodeCapacity,
			Allocatable: nodeAlloc,
		},
	}

	if nodeInfo.npuAllocateNum != "0" {
		addNodeNpuTop(v1node.Annotations, nodeInfo.npuTop)
	}

	errs := addNodeSelector(v1node, nodeInfo.nodeArch)
	if errs != nil {
		klog.V(logErrorLev).Infof("Test buildNpuNode arch type error: %v", errs)
		return nil
	}

	node := api.NewNodeInfo(v1node)
	return node
}

func addPodLabels(pod *v1.Pod, nodeArch string) error {
	if nodeArch == osArchArm {
		pod.Spec.NodeSelector[archSelector] = huaweiArchArm
		return nil
	}

	if nodeArch == osArchX86 {
		pod.Spec.NodeSelector[archSelector] = huaweiArchX86
		return nil
	}

	return fmt.Errorf("unKnown podArch %s", nodeArch)
}

func buildNpuPod(podInfo testPodInfo) *v1.Pod {

	pod := util.BuildPod(podInfo.namespace, podInfo.podName, podInfo.nodeName, v1.PodPending,
		buildNpuResourceList(podInfo.reqCpuNum, podInfo.reqMem, podInfo.npuNeedNum),
		podInfo.groupName, make(map[string]string, magicNumInt2), make(map[string]string, magicNumInt2))

	errs := addPodLabels(pod, osArchArm)
	if errs != nil {
		klog.V(logErrorLev).Infof("Test buildNpuPod arch type error: %v", errs)
		return nil
	}

	return pod
}

func buildQueue(queueName string, weight int32) *schedulingv1.Queue {
	queue := &schedulingv1.Queue{
		ObjectMeta: metav1.ObjectMeta{
			Name: queueName,
		},
		Spec: schedulingv1.QueueSpec{
			Weight: weight,
		},
	}

	return queue
}

func buildPodGroup(podGroupName string, namespaceName string, queueName string) *schedulingv1.PodGroup {
	return &schedulingv1.PodGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podGroupName,
			Namespace: namespaceName,
		},
		Spec: schedulingv1.PodGroupSpec{
			Queue: queueName,
		},
	}
}

func buildSchedulerCache(test testCaseStruct) *cache.SchedulerCache {
	binder := &util.FakeBinder{
		Binds:   map[string]string{},
		Channel: make(chan string),
	}
	schedulerCache := &cache.SchedulerCache{
		Nodes:         make(map[string]*api.NodeInfo, magicNumInt2),
		Jobs:          make(map[api.JobID]*api.JobInfo, magicNumInt2),
		Queues:        make(map[api.QueueID]*api.QueueInfo, magicNumInt2),
		Binder:        binder,
		StatusUpdater: &util.FakeStatusUpdater{},
		VolumeBinder:  &util.FakeVolumeBinder{},

		Recorder: record.NewFakeRecorder(bufferSize),
	}

	for _, node := range test.nodes {
		schedulerCache.AddNode(node.Node)
	}
	for _, pod := range test.pods {
		schedulerCache.AddPod(pod)
	}
	for _, ss := range test.podGroups {
		schedulerCache.AddPodGroupV1beta1(ss)
	}
	for _, q := range test.queues {
		schedulerCache.AddQueueV1beta1(q)
	}

	// Add pod can reduce the idle NPU num, so need recover node for pod
	recoverSchedulerNode(schedulerCache, test.nodes)
	return schedulerCache
}

func recoverSchedulerNode(sc *cache.SchedulerCache, nodes []*api.NodeInfo) {
	if sc == nil || sc.Nodes == nil {
		return
	}
	var nodeExistFlag bool

	for key, scNode := range sc.Nodes {
		nodeExistFlag = false
		for _, node := range nodes {
			if node.Node != nil && scNode.Name == node.Name {
				scNode.Idle = api.NewResource(node.Node.Status.Allocatable)
				scNode.Used = api.EmptyResource()
				nodeExistFlag = true
				break
			}
		}
		if !nodeExistFlag {
			delete(sc.Nodes, key)
		}
	}
}

func addSsnConfig(ssn *framework.Session) {
	var config conf.Configuration
	config.Arguments = make(map[string]string, 1)
	config.Arguments[archSelector] = "huawei-arm|huawei-x86"

	ssn.Configurations = append(ssn.Configurations, config)
}

func openTestSession(schedulerCache *cache.SchedulerCache, trueValue bool, test testCaseStruct) *framework.Session {
	ssn := framework.OpenSession(schedulerCache, []conf.Tier{
		{
			Plugins: []conf.PluginOption{
				{
					Name:             PluginName,
					EnabledNodeOrder: &trueValue,
					Arguments:        test.arguments,
				},
			},
		},
	}, nil)

	// recover node ilde nums
	recoverSessionNode(ssn, schedulerCache)

	addSsnConfig(ssn)

	return ssn
}

func recoverSessionNode(ssn *framework.Session, schedulerCache *cache.SchedulerCache) {
	if ssn == nil || schedulerCache == nil {
		return
	}
	ssn.Nodes = schedulerCache.Nodes
}

func initTestCase(testCaseName string, podGroupList []*schedulingv1.PodGroup, queueList []*schedulingv1.Queue,
	podList []*v1.Pod, nodeList []*api.NodeInfo, expect interface{}) testCaseStruct {
	return testCaseStruct{
		name:      testCaseName,
		podGroups: podGroupList,
		queues:    queueList,
		pods:      podList,
		nodes:     nodeList,
		expected:  expect,
	}
}

// test job valid
func buildValidJobTests() []testCaseStruct {
	var namespaceName, queueName, podGroupName, nodeName1, nodeName2 string
	var tests []testCaseStruct
	// build test env
	namespaceName = "n1"
	queueName = "c1"
	podGroupName = "pg1"
	nodeName1 = "ubuntu-1"
	nodeName2 = "ubuntu-2"

	podOneCard := buildNpuPod(testPodInfo{namespaceName, podGroupName, "p0", nodeName1, "10", "1Gi", "1"})
	podTwoCard := buildNpuPod(testPodInfo{namespaceName, podGroupName, "p0", nodeName1, "10", "1Gi", "2"})
	podEightCard1 := buildNpuPod(testPodInfo{namespaceName, podGroupName, "p1", nodeName1, "1", "1Gi", "8"})
	podEightCard2 := buildNpuPod(testPodInfo{namespaceName, podGroupName, "p2", nodeName2, "1", "1Gi", "8"})
	podErrorCard := buildNpuPod(testPodInfo{namespaceName, podGroupName, "p2", nodeName1, "1", "1Gi", "3"})
	podCpuCard := buildNpuPod(testPodInfo{namespaceName, podGroupName, "p3", nodeName1, "1", "1Gi", "0"})
	podErrorDistributeCard := buildNpuPod(testPodInfo{namespaceName, podGroupName, "16", nodeName1, "1", "1Gi", "16"})
	podDistributeCard1 := buildNpuPod(testPodInfo{namespaceName, podGroupName, "pdis1", nodeName1, "1", "1Gi", "4"})
	podDistributeCard2 := buildNpuPod(testPodInfo{namespaceName, podGroupName, "pdis2", nodeName2, "1", "1Gi", "4"})

	nodeStand1 := buildNpuNode(testNodeInfo{nodeName1, osArchArm, "192", "755Gi", "8",
		"Ascend910-0,Ascend910-1,Ascend910-2,Ascend910-3,Ascend910-4,Ascend910-5,Ascend910-6,Ascend910-7"})
	nodeStand2 := buildNpuNode(testNodeInfo{nodeName2, osArchArm, "192", "755Gi", "8",
		"Ascend910-0,Ascend910-1,Ascend910-2,Ascend910-3,Ascend910-4,Ascend910-5,Ascend910-6,Ascend910-7"})
	nodeFault1 := buildNpuNode(testNodeInfo{nodeName2, osArchArm, "192", "755Gi", "7",
		"Ascend910-0,Ascend910-1,Ascend910-2,Ascend910-3,Ascend910-5,Ascend910-6,Ascend910-7"})
	nodeStand3 := buildNpuNode(testNodeInfo{nodeName1, osArchArm, "192", "755Gi", "8",
		"Ascend910-0"})

	queue1 := buildQueue(queueName, 1)
	podGroup := buildPodGroup(podGroupName, namespaceName, queueName)

	// make test cases
	tests = append(tests, initTestCase("one card job success test", []*schedulingv1.PodGroup{podGroup},
		[]*schedulingv1.Queue{queue1}, []*v1.Pod{podOneCard}, []*api.NodeInfo{nodeStand1}, true))
	tests = append(tests, initTestCase("eight card job success test", []*schedulingv1.PodGroup{podGroup},
		[]*schedulingv1.Queue{queue1}, []*v1.Pod{podEightCard1}, []*api.NodeInfo{nodeStand1}, true))
	tests = append(tests, initTestCase("cpu job success test", []*schedulingv1.PodGroup{podGroup},
		[]*schedulingv1.Queue{queue1}, []*v1.Pod{podCpuCard}, []*api.NodeInfo{nodeStand1}, true))
	tests = append(tests, initTestCase("error card(3) job test", []*schedulingv1.PodGroup{podGroup},
		[]*schedulingv1.Queue{queue1}, []*v1.Pod{podErrorCard}, []*api.NodeInfo{nodeStand1}, false))
	tests = append(tests, initTestCase("error card(16) job test", []*schedulingv1.PodGroup{podGroup},
		[]*schedulingv1.Queue{queue1}, []*v1.Pod{podErrorDistributeCard}, []*api.NodeInfo{nodeStand1}, false))
	tests = append(tests, initTestCase("error one card job test", []*schedulingv1.PodGroup{podGroup},
		[]*schedulingv1.Queue{queue1}, []*v1.Pod{podOneCard}, []*api.NodeInfo{nodeStand3}, true))
	tests = append(tests, initTestCase("error two card job test", []*schedulingv1.PodGroup{podGroup},
		[]*schedulingv1.Queue{queue1}, []*v1.Pod{podTwoCard}, []*api.NodeInfo{nodeStand3}, false))
	// distribute job test
	tests = append(tests, initTestCase("distribute job success test", []*schedulingv1.PodGroup{podGroup},
		[]*schedulingv1.Queue{queue1}, []*v1.Pod{podEightCard1, podEightCard2},
		[]*api.NodeInfo{nodeStand1, nodeStand2}, true))
	tests = append(tests, initTestCase("distribute job parameters err test", []*schedulingv1.PodGroup{podGroup},
		[]*schedulingv1.Queue{queue1}, []*v1.Pod{podDistributeCard1, podDistributeCard2}, []*api.NodeInfo{nodeStand1}, false))
	// fault card node test
	tests = append(tests, initTestCase("one card job success on fault node test", []*schedulingv1.PodGroup{podGroup},
		[]*schedulingv1.Queue{queue1}, []*v1.Pod{podOneCard}, []*api.NodeInfo{nodeFault1}, true))
	tests = append(tests, initTestCase("eight card job success on fault node test", []*schedulingv1.PodGroup{podGroup},
		[]*schedulingv1.Queue{queue1}, []*v1.Pod{podEightCard1}, []*api.NodeInfo{nodeFault1}, true))
	tests = append(tests, initTestCase("cpu job success on fault node test", []*schedulingv1.PodGroup{podGroup},
		[]*schedulingv1.Queue{queue1}, []*v1.Pod{podCpuCard}, []*api.NodeInfo{nodeFault1}, true))
	tests = append(tests, initTestCase("error card(3) job on fault node test", []*schedulingv1.PodGroup{podGroup},
		[]*schedulingv1.Queue{queue1}, []*v1.Pod{podErrorCard}, []*api.NodeInfo{nodeFault1}, false))
	tests = append(tests, initTestCase("error card(16) job on fault node test", []*schedulingv1.PodGroup{podGroup},
		[]*schedulingv1.Queue{queue1}, []*v1.Pod{podErrorDistributeCard}, []*api.NodeInfo{nodeFault1}, false))

	return tests
}

// go test for validJobFn
func TestValidJobFn(t *testing.T) {
	framework.RegisterPluginBuilder(PluginName, New)
	defer framework.CleanupPluginBuilders()
	var successNum int
	var failedNum int

	successNum = 0
	failedNum = 0
	tests := buildValidJobTests()

	for i, test := range tests {
		// open session
		schedulerCache := buildSchedulerCache(test)
		ssn := openTestSession(schedulerCache, true, test)

		// test validJobFn
		for _, jobInfo := range ssn.Jobs {
			err := validJobFn(jobInfo, ssn.Configurations)
			if err != nil && err.Pass != test.expected {
				fmt.Printf("case %d : %-70s failed: %v\n", i, test.name, err.Reason)
				t.Errorf("")
				failedNum++
				continue
			}
			successNum++
			fmt.Printf("case %d : %-70s success\n", i, test.name)
		}
		framework.CloseSession(ssn)
	}
	fmt.Printf("+++Taltol : %d ,success : %d failed : %d\n", successNum+failedNum, successNum, failedNum)
}

// test node NpuPredicate
func buildNpuPredicateTests() []testCaseStruct {
	var namespaceName, queueName, podGroupName, nodeName1, nodeName2 string
	var tests []testCaseStruct
	// build test env
	namespaceName = "n1"
	queueName = "c1"
	podGroupName = "pg1"
	nodeName1 = "ubuntu-1"
	nodeName2 = "ubuntu-2"

	npuPod := buildNpuPod(testPodInfo{namespaceName, podGroupName, "p0", nodeName1, "10", "1Gi", "8"})
	cpuPod := buildNpuPod(testPodInfo{namespaceName, podGroupName, "p3", nodeName1, "1", "1Gi", "0"})

	npuNodeStand1 := buildNpuNode(testNodeInfo{nodeName1, osArchArm, "192", "755Gi", "8",
		"Ascend910-0,Ascend910-1,Ascend910-2,Ascend910-3,Ascend910-4,Ascend910-5,Ascend910-6,Ascend910-7"})
	stableNode := buildNpuNode(testNodeInfo{nodeName2, osArchArm, "192", "755Gi", "8",
		"Ascend910-0,Ascend910-1,Ascend910-2,Ascend910-3,Ascend910-4,Ascend910-5,Ascend910-6"})
	cpuNodeStand1 := buildNpuNode(testNodeInfo{nodeName1, osArchArm, "192", "755Gi", "0", ""})

	queue1 := buildQueue(queueName, 1)
	podGroup := buildPodGroup(podGroupName, namespaceName, queueName)

	// make test cases
	tests = append(tests, initTestCase("npu node success test", []*schedulingv1.PodGroup{podGroup},
		[]*schedulingv1.Queue{queue1}, []*v1.Pod{npuPod}, []*api.NodeInfo{npuNodeStand1}, nil))
	tests = append(tests, initTestCase("cpu job success test", []*schedulingv1.PodGroup{podGroup},
		[]*schedulingv1.Queue{queue1}, []*v1.Pod{cpuPod}, []*api.NodeInfo{npuNodeStand1}, nil))
	tests = append(tests, initTestCase("cpu node success test", []*schedulingv1.PodGroup{podGroup},
		[]*schedulingv1.Queue{queue1}, []*v1.Pod{cpuPod}, []*api.NodeInfo{cpuNodeStand1}, nil))
	tests = append(tests, initTestCase("node stable success test", []*schedulingv1.PodGroup{podGroup},
		[]*schedulingv1.Queue{queue1}, []*v1.Pod{npuPod}, []*api.NodeInfo{stableNode}, errors.New("no meet")))

	return tests
}

func testNodeByJobNpuPredicateFn(ssn *framework.Session, node *api.NodeInfo) error {
	for _, jobInfo := range ssn.Jobs {
		for _, task := range jobInfo.Tasks {
			err := nodePredicate(task, node, ssn.Configurations)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// go test for NpuPredicateFn
func TestNpuPredicateFn(t *testing.T) {
	framework.RegisterPluginBuilder(PluginName, New)
	defer framework.CleanupPluginBuilders()

	var successNum int
	var failedNum int

	successNum = 0
	failedNum = 0
	tests := buildNpuPredicateTests()

	for i, test := range tests {
		// open session
		schedulerCache := buildSchedulerCache(test)
		ssn := openTestSession(schedulerCache, true, test)

		// test validJobFn
		var nodes []*api.NodeInfo
		for _, node := range test.nodes {
			nodes = append(nodes, node)
			err := testNodeByJobNpuPredicateFn(ssn, node)
			if (err != nil) && (test.expected == nil) {
				fmt.Printf("case %d : %-70s failed: %v,%v\n", i, test.name, err, test.expected)
				t.Errorf("")
				failedNum++
				continue
			}
			successNum++
			fmt.Printf("case %d : %-70s success\n", i, test.name)
		}

		framework.CloseSession(ssn)
	}
	fmt.Printf("+++Taltol : %d ,success : %d failed : %d\n", successNum+failedNum, successNum, failedNum)
}

func buildAnnotationRandom(leftHccsNum int, rightHccsNum int) string {
	var ann []int
	leftHccs := []int{0, 1, 2, 3}
	rightHccs := []int{4, 5, 6, 7}
	var randCard1 int
	var randCard2 int

	generateRandCard := func() {
		randCard1 = rand.Int() % npuNumPerHccs
		for {
			if randCard2 = (rand.Int() + 1) % npuNumPerHccs; randCard1 != randCard2 {
				break
			}
		}
	}
	generateHccs := func(hccs *[]int, hccsNum int) {
		generateRandCard()
		switch hccsNum {
		case 1:
			ann = append(ann, (*hccs)[randCard1])
		case 2:
			ann = append(ann, (*hccs)[randCard1])
			ann = append(ann, (*hccs)[randCard2])
		case 3:
			ann = append(ann, (*hccs)[:randCard1]...)
			ann = append(ann, (*hccs)[randCard1+1:]...)
		case 4:
			ann = append(ann, *hccs...)
		default:
		}
	}
	generateHccs(&leftHccs, leftHccsNum)
	generateHccs(&rightHccs, rightHccsNum)

	return changeIntArrToStr(ann)
}

func addNpuNodePool(nodeName string, npuAllocatecNum int, nodePool map[string]*api.NodeInfo) {
	// map: key type:nodeName_leftHccsNum_rightHccsNum
	for leftHccsNum := 0; leftHccsNum <= npuNumPerHccs; leftHccsNum++ {
		for rightHccsNum := 0; rightHccsNum <= npuNumPerHccs; rightHccsNum++ {
			if (leftHccsNum + rightHccsNum) > npuAllocatecNum {
				break
			}
			key := nodeName + "_" + strconv.Itoa(leftHccsNum) + "_" + strconv.Itoa(rightHccsNum)
			annotation := buildAnnotationRandom(leftHccsNum, rightHccsNum)
			p := buildNpuNode(testNodeInfo{nodeName, osArchArm, "96", "64Gi", strconv.Itoa(npuAllocatecNum), annotation})
			nodePool[key] = p
		}
	}
}

func buildNpuNodePool(nodeAlloList []NodeAllocate) map[string]*api.NodeInfo {
	nodePool := make(map[string]*api.NodeInfo, nodePoolSize)
	for _, nodAlloc := range nodeAlloList {
		addNpuNodePool(nodAlloc.nodeName, nodAlloc.npuAllocateNum, nodePool)
	}
	return nodePool
}

func buildNpuPodPool(namespaceName string, podGroupName string, nodeAlloList []NodeAllocate) map[string]*v1.Pod {
	// map: key type pi_nodeName, i=pod need npu number
	podPool := make(map[string]*v1.Pod, nodeNpuNumber)
	for podNo := 0; podNo <= nodeNpuNumber; podNo++ {
		for _, nodAlloc := range nodeAlloList {
			podName := PodNamePrefix + strconv.Itoa(podNo) + "_" + nodAlloc.nodeName
			p := buildNpuPod(testPodInfo{namespaceName, podGroupName, podName, nodAlloc.nodeName,
				"1", "1Gi", strconv.Itoa(podNo)})
			podPool[podName] = p
		}
	}
	return podPool
}

func getTaskInfo(podName string) (string, string, error) {
	// podName format: namespace/pi_nodeName, i = needcardnum
	// retrun : needcardnum, nodename
	if len(podName) < podNameLength {
		return "", "", errors.New("input podName Format error: length not enough")
	}
	strArray1 := strings.Split(podName, "/")
	if len(strArray1) < arraySize {
		return "", "", errors.New("input podName Format error: not include '/' ")
	}
	strArray2 := strings.Split(strArray1[1], "_")
	if len(strArray2) < arraySize {
		return "", "", errors.New("input podName Format error: not include '_' ")
	}
	needCardnum := strings.TrimPrefix(strArray2[0], PodNamePrefix)
	if _, err := strconv.Atoi(needCardnum); err != nil {
		return "", "", errors.New("podName Format error: needCardNum is not 'int' type after 'p' and before '_' ")
	}
	return needCardnum, strArray2[1], nil
}

func buildNodeFromName(nodeInfoList []string, nodPool map[string]*api.NodeInfo, taskName string,
	nodeNameForPod string, nodList *[]*api.NodeInfo) (interface{}, error) {

	expect := map[string]map[string]float64{taskName: {}}

	for _, nodeInfo := range nodeInfoList {
		node, ok := nodPool[nodeInfo]
		if !ok {
			return testCaseStruct{}, errors.New("node not find in nodPool")
		}
		*nodList = append(*nodList, node)

		if node.Name == nodeNameForPod {
			expect[taskName][node.Name] = 1
		} else {
			expect[taskName][node.Name] = 0
		}
	}

	nodeExist := false
	for _, score := range expect[taskName] {
		if score > 0 {
			nodeExist = true
		}
	}
	if !nodeExist {
		expect = nil
	}

	return expect, nil
}

func initBatchOrderTestCase(testName string, podGroup *schedulingv1.PodGroup, queue *schedulingv1.Queue,
	nodPool map[string]*api.NodeInfo, podPool map[string]*v1.Pod,
	taskName string, nodeInfoList []string) (testCaseStruct, error) {
	var nodList []*api.NodeInfo

	// podName: pi_nodename
	needCardNum, nodeNameForPod, err := getTaskInfo(taskName)
	if err != nil {
		return testCaseStruct{}, errors.New("taskName format error")
	}
	if needCardNum != "1" && needCardNum != "2" && needCardNum != "4" && needCardNum != "8" {
		return testCaseStruct{}, errors.New("need Card Num is not right")
	}
	if taskName == "" || len(nodeInfoList) == 0 {
		return testCaseStruct{}, errors.New("test case para is not right")
	}

	podName := PodNamePrefix + needCardNum + "_" + nodeNameForPod
	testCaseName := needCardNum + "cards -" + testName
	podGroupList := []*schedulingv1.PodGroup{podGroup}
	queueList := []*schedulingv1.Queue{queue}
	podList := []*v1.Pod{podPool[podName]}

	expect, err := buildNodeFromName(nodeInfoList, nodPool, taskName, nodeNameForPod, &nodList)
	if err != nil {
		return testCaseStruct{}, err
	}

	return initTestCase(testCaseName, podGroupList, queueList, podList, nodList, expect), nil
}

func buildBatchNodeOrderFnTests() []testCaseStruct {
	// build test env
	var tests []testCaseStruct
	namespaceName := "ns1"
	queueName := "q1"
	podGroupName := "pg1"
	nodAllocate := []NodeAllocate{
		{"n1", nodeNpuNumber},
		{"n2", nodeNpuNumber},
		{"n3", nodeNpuNumber},
		{"n4", nodeNpuNumber7},
		{"n5", nodeNpuNumber6},
		{"n0", 0},
	}
	queue := buildQueue(queueName, 1)
	podGroup := buildPodGroup(podGroupName, namespaceName, queueName)
	nodPool := buildNpuNodePool(nodAllocate)
	podPool := buildNpuPodPool(namespaceName, podGroupName, nodAllocate)

	addBatchOrderTestCase := func(tests *[]testCaseStruct, taskName string, nodeInfoList []string) {
		testCaseStruct, err := initBatchOrderTestCase(strconv.Itoa(len(*tests)+1), podGroup, queue,
			nodPool, podPool, taskName, nodeInfoList)
		if err != nil {
			fmt.Printf("addBatchOrderTestCase error: %v", err)
			return
		}
		*tests = append(*tests, testCaseStruct)
	}
	// make test cases
	// 1 card
	addBatchOrderTestCase(&tests, "ns1/p1_n0", NodeNameList{"n1_0_0"})
	addBatchOrderTestCase(&tests, "ns1/p1_n1", NodeNameList{"n1_0_1"})
	addBatchOrderTestCase(&tests, "ns1/p1_n1", NodeNameList{"n1_4_4"})
	addBatchOrderTestCase(&tests, "ns1/p1_n0", NodeNameList{"n1_0_0", "n2_0_0", "n3_0_0"})
	addBatchOrderTestCase(&tests, "ns1/p1_n1", NodeNameList{"n1_1_0", "n2_1_1", "n3_1_2"})
	addBatchOrderTestCase(&tests, "ns1/p1_n1", NodeNameList{"n1_1_1", "n2_0_0", "n3_1_2"})
	addBatchOrderTestCase(&tests, "ns1/p1_n1", NodeNameList{"n1_1_2", "n2_1_3", "n3_1_4"})
	addBatchOrderTestCase(&tests, "ns1/p1_n1", NodeNameList{"n1_1_2", "n2_0_2", "n3_1_4"})
	addBatchOrderTestCase(&tests, "ns1/p1_n1", NodeNameList{"n1_1_4", "n2_0_2", "n3_0_3"})
	addBatchOrderTestCase(&tests, "ns1/p1_n1", NodeNameList{"n1_3_0", "n2_3_2", "n3_3_3"})
	addBatchOrderTestCase(&tests, "ns1/p1_n1", NodeNameList{"n1_3_2", "n2_2_0", "n3_3_3"})
	addBatchOrderTestCase(&tests, "ns1/p1_n1", NodeNameList{"n1_3_3", "n2_2_0", "n3_2_2"})
	addBatchOrderTestCase(&tests, "ns1/p1_n1", NodeNameList{"n1_3_3", "n2_3_4", "n4_1_3"})
	addBatchOrderTestCase(&tests, "ns1/p1_n1", NodeNameList{"n1_3_2", "n2_3_4", "n4_3_3"})
	addBatchOrderTestCase(&tests, "ns1/p1_n1", NodeNameList{"n1_2_0", "n2_2_2", "n3_0_0"})
	addBatchOrderTestCase(&tests, "ns1/p1_n1", NodeNameList{"n1_2_2", "n2_2_4", "n3_0_0"})
	addBatchOrderTestCase(&tests, "ns1/p1_n1", NodeNameList{"n1_4_0", "n2_4_4", "n4_1_0"})
	addBatchOrderTestCase(&tests, "ns1/p1_n1", NodeNameList{"n1_4_4", "n4_1_2", "n5_1_0"})
	addBatchOrderTestCase(&tests, "ns1/p1_n4", NodeNameList{"n1_0_0", "n4_1_2", "n5_1_0"})
	addBatchOrderTestCase(&tests, "ns1/p1_n5", NodeNameList{"n1_0_0", "n4_0_0", "n5_4_0"})
	// 2 cards
	addBatchOrderTestCase(&tests, "ns1/p2_n0", NodeNameList{"n1_0_0"})
	addBatchOrderTestCase(&tests, "ns1/p2_n0", NodeNameList{"n1_0_1"})
	addBatchOrderTestCase(&tests, "ns1/p2_n1", NodeNameList{"n1_0_2"})
	addBatchOrderTestCase(&tests, "ns1/p2_n0", NodeNameList{"n1_0_0", "n2_0_1", "n3_1_1"})
	addBatchOrderTestCase(&tests, "ns1/p2_n1", NodeNameList{"n1_2_0", "n2_2_1", "n3_2_2"})
	addBatchOrderTestCase(&tests, "ns1/p2_n1", NodeNameList{"n1_2_1", "n2_0_0", "n3_2_2"})
	addBatchOrderTestCase(&tests, "ns1/p2_n1", NodeNameList{"n1_4_0", "n2_4_1", "n3_4_3"})
	addBatchOrderTestCase(&tests, "ns1/p2_n1", NodeNameList{"n1_2_0", "n2_4_1", "n3_4_3"})
	addBatchOrderTestCase(&tests, "ns1/p2_n1", NodeNameList{"n1_4_1", "n2_0_0", "n3_4_3"})
	addBatchOrderTestCase(&tests, "ns1/p2_n1", NodeNameList{"n1_2_1", "n2_0_0", "n3_4_3"})
	addBatchOrderTestCase(&tests, "ns1/p2_n1", NodeNameList{"n1_4_3", "n2_0_0", "n3_3_3"})
	addBatchOrderTestCase(&tests, "ns1/p2_n1", NodeNameList{"n1_2_1", "n2_4_0", "n3_2_2"})
	addBatchOrderTestCase(&tests, "ns1/p2_n1", NodeNameList{"n1_2_0", "n2_4_0", "n3_4_0"})
	addBatchOrderTestCase(&tests, "ns1/p2_n1", NodeNameList{"n1_2_3", "n2_4_0", "n3_1_3"})
	addBatchOrderTestCase(&tests, "ns1/p2_n1", NodeNameList{"n1_3_0", "n2_3_1", "n4_2_0"})
	addBatchOrderTestCase(&tests, "ns1/p2_n4", NodeNameList{"n1_1_0", "n4_3_1", "n5_2_0"})
	addBatchOrderTestCase(&tests, "ns1/p2_n5", NodeNameList{"n1_0_0", "n4_1_1", "n5_3_3"})
	// 4 cards
	addBatchOrderTestCase(&tests, "ns1/p4_n0", NodeNameList{"n1_0_0"})
	addBatchOrderTestCase(&tests, "ns1/p4_n0", NodeNameList{"n1_0_2"})
	addBatchOrderTestCase(&tests, "ns1/p4_n1", NodeNameList{"n1_0_4"})
	addBatchOrderTestCase(&tests, "ns1/p4_n1", NodeNameList{"n1_2_4"})
	addBatchOrderTestCase(&tests, "ns1/p4_n0", NodeNameList{"n1_2_0", "n2_2_2", "n3_3_2"})
	addBatchOrderTestCase(&tests, "ns1/p4_n1", NodeNameList{"n1_4_0", "n2_4_1", "n3_4_2"})
	addBatchOrderTestCase(&tests, "ns1/p4_n1", NodeNameList{"n1_4_1", "n2_3_1", "n3_4_2"})
	addBatchOrderTestCase(&tests, "ns1/p4_n1", NodeNameList{"n1_4_2", "n2_4_4", "n3_4_3"})
	addBatchOrderTestCase(&tests, "ns1/p4_n1", NodeNameList{"n1_4_1", "n2_4_2", "n4_4_0"})
	addBatchOrderTestCase(&tests, "ns1/p4_n4", NodeNameList{"n1_3_1", "n2_3_2", "n4_4_2"})
	addBatchOrderTestCase(&tests, "ns1/p4_n4", NodeNameList{"n1_0_0", "n4_4_1", "n5_4_0"})
	addBatchOrderTestCase(&tests, "ns1/p4_n5", NodeNameList{"n1_0_0", "n4_3_2", "n5_4_2"})
	// 8 cards
	addBatchOrderTestCase(&tests, "ns1/p8_n0", NodeNameList{"n1_0_0"})
	addBatchOrderTestCase(&tests, "ns1/p8_n0", NodeNameList{"n1_4_3"})
	addBatchOrderTestCase(&tests, "ns1/p8_n0", NodeNameList{"n1_4_3", "n2_4_2", "n3_4_1"})
	addBatchOrderTestCase(&tests, "ns1/p8_n1", NodeNameList{"n1_4_4", "n2_4_3", "n3_4_2"})
	addBatchOrderTestCase(&tests, "ns1/p8_n2", NodeNameList{"n2_4_4", "n4_4_3", "n5_4_1"})
	// other exceptions should be test by Predicate function
	return tests
}

func testNodeByJobBatchNodeOrderFn(
	ssn *framework.Session,
	nodes []*api.NodeInfo,
	test *testCaseStruct,
	testNo int) error {
	var testCaseName string
	for _, jobInfo := range ssn.Jobs {
		for _, task := range jobInfo.Tasks {
			taskID := fmt.Sprintf("%s/%s", task.Namespace, task.Name)
			score, err := batchNodeOrderFn(task, nodes)
			if err != nil {
				fmt.Println(` "case" + strconv.Itoa(testNo) + " : " + test.name + " task:" + taskID + " on node:" 
				+ task.NodeName + " has err:" + err.Error()`)
			}
			scoreErr := testScore(test, taskID, task, score, testNo)
			if scoreErr != nil {
				return scoreErr
			}

			testCaseName = fmt.Sprintf("%s task:%s on node: %s test ", test.name, taskID, task.NodeName)
			fmt.Printf("case %d : %-70s success\n", testNo, testCaseName)

			npuAllocateFunc(&framework.Event{
				Task: task,
			}, ssn.Nodes)
			npuDeallocateFunc(&framework.Event{
				Task: task,
			}, ssn.Nodes)
		}
	}

	return nil
}

func testScore(test *testCaseStruct, taskID string, task *api.TaskInfo, score map[string]float64, testNo int) error {
	expType, ok := test.expected.(map[string]map[string]float64)
	if !ok {
		return errors.New("test case expect type is error")
	}

	tasks, taskOk := expType[taskID]
	if !taskOk {
		return nil
	}

	if expectScore, scoreOk := tasks[task.NodeName]; scoreOk && expectScore != score[task.NodeName] {
		fmt.Println("case " + strconv.Itoa(testNo) + " : " + test.name + " task:" +
			taskID + " on node:" + task.NodeName + " test failed, expect have score:" +
			strconv.Itoa(int(expectScore)) + ",but get:" + strconv.Itoa(int(score[task.NodeName])))
		return errors.New("test case failed")
	}

	return nil
}

// go test for BatchNodeOrderFn
func TestBatchNodeOrderFn(t *testing.T) {
	framework.RegisterPluginBuilder(PluginName, New)
	defer framework.CleanupPluginBuilders()

	var successNum int
	var failedNum int

	successNum = 0
	failedNum = 0
	tests := buildBatchNodeOrderFnTests()

	for i, test := range tests {
		// open session
		schedulerCache := buildSchedulerCache(test)
		ssn := openTestSession(schedulerCache, true, test)
		// test BatchNodeOrderFn
		var nodes []*api.NodeInfo
		for _, node := range ssn.Nodes {
			nodes = append(nodes, node)
		}
		err := testNodeByJobBatchNodeOrderFn(ssn, nodes, &test, i)
		framework.CloseSession(ssn)
		if err != nil {
			t.Errorf("%v", err)
			failedNum++
			continue
		}
		successNum++
	}
	fmt.Printf("+++Taltol : %d ,success : %d failed : %d\n", successNum+failedNum, successNum, failedNum)
}
