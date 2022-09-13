/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package rescheduling is using for HuaWei Ascend pin fault rescheduling.

*/
package module910x8

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/conf"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/base"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/rescheduling"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

const (
	zero              = 0
	one               = 1
	two               = 2
	three             = 3
	four              = 4
	heartbeatInterval = 5
	fakeTime          = 12345
)

type module910x8Fields struct {
	reHandle        *rescheduling.ReScheduler
	baseHandler     base.NPUHandler
	netUnhealthyKey string
	affScoreList    [][]int
}

type module910x8PreStartActionArgs struct {
	ssn              *framework.Session
	cacheFuncBefore1 func()
	cacheFuncBefore2 func()
	cacheFuncAfter1  func()
	cacheFuncAfter2  func()
	cacheFuncBefore3 func()
	cacheFuncAfter3  func()
	cacheFuncBefore4 func()
	cacheFuncAfter4  func()
}

type module910x8PreStartActionTests struct {
	name    string
	fields  module910x8Fields
	args    module910x8PreStartActionArgs
	wantErr bool
}

// buildModule910x8PreStartActionTest1 initial: no contents in reCache; 4 nodes, node0 card fault; 2 jobs,
// job0-pod0 on fault node
func buildModule910x8PreStartActionTest1() module910x8PreStartActionTests {
	ssn1 := test.FakeSSNReSchedule()
	env := fakeEnvEmpty()
	fakeEnvAddJobsAndNodesToEnv(&env)
	var tmpPatche1 *gomonkey.Patches
	var tmpPatche2 *gomonkey.Patches
	var tmpPatche3 *gomonkey.Patches
	var tmpPatche4 *gomonkey.Patches
	myArgs := buildModule910x8PreStartActionTestCacheArgs(tmpPatche1, tmpPatche2, tmpPatche3, tmpPatche4, nil)
	myArgs.ssn = ssn1
	test1 := module910x8PreStartActionTests{
		name: "01-PreStartAction()-no fault initially, add card fault and corresponding job fault in session",
		fields: module910x8Fields{
			baseHandler: base.NPUHandler{
				SchedulerPlugin:  plugin.SchedulerPlugin{},
				SchedulerJobAttr: env.Jobs["vcjob/job0"].SchedulerJobAttr,
				ScheduleEnv:      env,
				MaxNodeNPUNum:    0,
				MaxCardNPUNum:    0,
			},
			reHandle: &rescheduling.ReScheduler{
				GraceDeleteTime:      0,
				Level:                "",
				Jobs:                 nil,
				Nodes:                nil,
				DealReSchedulerCache: nil,
			},
		},
		args:    myArgs,
		wantErr: false,
	}
	test1.args.cacheFuncBefore2 = func() {
		tmpPatche2 = gomonkey.ApplyFunc(util.GetConfigMapWithRetry, func(
			_ kubernetes.Interface, _, _ string) (*v1.ConfigMap, error) {
			return nil, errors.New("")
		})
	}
	return test1
}

func buildModule910x8PreStartActionTestCacheArgs(tmpPatche1 *gomonkey.Patches,
	tmpPatche2 *gomonkey.Patches, tmpPatche3 *gomonkey.Patches, tmpPatche4 *gomonkey.Patches,
	faultCM *v1.ConfigMap) module910x8PreStartActionArgs {
	args := module910x8PreStartActionArgs{
		cacheFuncBefore1: func() {
			tmpPatche1 = gomonkey.ApplyMethod(reflect.TypeOf(&framework.Session{}), "Evict",
				func(_ *framework.Session, _ *api.TaskInfo, _ string) error { return nil })
		},
		cacheFuncBefore2: func() {
			tmpPatche2 = gomonkey.ApplyFunc(util.GetConfigMapWithRetry, func(
				_ kubernetes.Interface, _, _ string) (*v1.ConfigMap, error) {
				return faultCM, nil
			})
		},
		cacheFuncBefore3: func() {
			tmpPatche3 = gomonkey.ApplyMethod(reflect.TypeOf(&rescheduling.FaultJob{}),
				"CheckJobExistsInKubernetes", func(_ *rescheduling.FaultJob,
					_ *framework.Session) bool {
					return true
				})
		},
		cacheFuncBefore4: func() {
			tmpPatche4 = gomonkey.ApplyMethod(reflect.TypeOf(&util.NPUTask{}), "DeleteRealPodByTask",
				func(_ *util.NPUTask, _ *framework.Session, _ int64) error { return nil })
		},
		cacheFuncAfter1: func() {
			if tmpPatche1 != nil {
				tmpPatche1.Reset()
			}
		},
		cacheFuncAfter2: func() {
			if tmpPatche2 != nil {
				tmpPatche2.Reset()
			}
		},
		cacheFuncAfter3: func() {
			if tmpPatche3 != nil {
				tmpPatche3.Reset()
			}
		},
		cacheFuncAfter4: func() {
			if tmpPatche4 != nil {
				tmpPatche4.Reset()
			}
		},
	}
	return args
}

func buildModule910x8PreStartActionTest2() module910x8PreStartActionTests {
	ssn1 := test.FakeSSNReSchedule()
	env := fakeEnvEmpty()
	fakeEnvAddJobsAndNodesToEnv(&env)
	fakeEnvAddCacheFaultNodeToEnv(&env)
	fakeEnvAddCacheFaultJobToEnv(&env, []string{"job0", "node0", "node1"}, time.Now().Unix(), 0)

	var tmpPatche1 *gomonkey.Patches
	var tmpPatche2 *gomonkey.Patches
	var tmpPatche3 *gomonkey.Patches
	var tmpPatche4 *gomonkey.Patches
	reHandle := fakeReSchedulerNew(env)
	faultCM := fakeFaultCM(env)
	myArgs := buildModule910x8PreStartActionTestCacheArgs(tmpPatche1, tmpPatche2, tmpPatche3, tmpPatche4, faultCM)
	myArgs.ssn = ssn1
	test6 := module910x8PreStartActionTests{
		name: "02-PreStartAction()-with fault node and job in cm",
		fields: module910x8Fields{
			baseHandler: base.NPUHandler{
				SchedulerPlugin:  plugin.SchedulerPlugin{},
				SchedulerJobAttr: env.Jobs["vcjob/job0"].SchedulerJobAttr,
				ScheduleEnv:      env,
				MaxNodeNPUNum:    0,
				MaxCardNPUNum:    0,
			},
			reHandle: &reHandle,
		},
		args:    myArgs,
		wantErr: false,
	}
	return test6
}

func buildModule910x8PreStartActionTest4() module910x8PreStartActionTests {
	ssn1 := test.FakeSSNReSchedule()
	env := fakeEnvEmpty()
	fakeEnvAddJobsAndNodesToEnv(&env)
	fakeEnvAddCacheFaultNodeToEnv(&env)
	fakeEnvAddCacheFaultJobToEnv(&env, []string{"job0", "node0", "node1"}, 0, time.Now().Unix()-1)
	var tmpPatche1 *gomonkey.Patches
	var tmpPatche2 *gomonkey.Patches
	var tmpPatche3 *gomonkey.Patches
	var tmpPatche4 *gomonkey.Patches
	reHandle := rescheduling.ReScheduler{
		GraceDeleteTime: rescheduling.DefaultGraceOverTime,
		Level:           "",
		Jobs:            env.Jobs,
		Nodes:           env.Nodes,
		DealReSchedulerCache: &rescheduling.DealReSchedulerCache{
			FaultNodes: nil,
			FaultJobs:  nil,
			DealReSchedulerConfigmap: &rescheduling.DealReSchedulerConfigmap{
				CMName:      rescheduling.CmName,
				CMNameSpace: rescheduling.CmNameSpace,
				CMData: map[string]string{rescheduling.CmFaultNodeKind: env.Cache.Data[rescheduling.
					RePropertyName][rescheduling.CmFaultNodeKind],
					rescheduling.CmFaultJob910x8Kind: env.Cache.Data[rescheduling.
						RePropertyName][rescheduling.CmFaultJob910x8Kind]},
			},
		},
	}
	faultCM := fakeFaultCM(env)
	myArgs := buildModule910x8PreStartActionTestCacheArgs(tmpPatche1, tmpPatche2, tmpPatche3, tmpPatche4, faultCM)
	myArgs.ssn = ssn1
	test7 := module910x8PreStartActionTests{
		name: "04-PreStartAction()-with fault node and job in cm and faultJob not in session",
		fields: module910x8Fields{
			baseHandler: fakeBaseHandlerEmpty(env),
			reHandle:    &reHandle,
		},
		args:    myArgs,
		wantErr: false,
	}
	return test7
}

func buildModule910x8PreStartActionTest3() module910x8PreStartActionTests {
	ssn1 := test.FakeSSNReSchedule()
	env := fakeEnvEmpty()
	fakeEnvAddJobsAndNodesToEnv(&env)
	fakeEnvAddCacheFaultNodeToEnv(&env)
	fakeEnvAddCacheFaultJobToEnv(&env, []string{"job2", "node0", "node1"}, time.Now().Unix()-1, time.Now().Unix()-1)

	var tmpPatche1 *gomonkey.Patches
	var tmpPatche2 *gomonkey.Patches
	var tmpPatche3 *gomonkey.Patches
	var tmpPatche4 *gomonkey.Patches
	reHandle := fakeReSchedulerNew(env)
	faultCM := fakeFaultCM(env)
	myArgs := buildModule910x8PreStartActionTestCacheArgs(tmpPatche1, tmpPatche2, tmpPatche3, tmpPatche4, faultCM)
	myArgs.ssn = ssn1
	test2 := module910x8PreStartActionTests{
		name: "03-PreStartAction()-with fault node and job in cm job not in session",
		fields: module910x8Fields{
			baseHandler: fakeBaseHandlerEmpty(env),
			reHandle:    &reHandle,
		},
		args:    myArgs,
		wantErr: false,
	}
	return test2
}

func buildModule910x8PreStartActionTests() []module910x8PreStartActionTests {
	return []module910x8PreStartActionTests{
		buildModule910x8PreStartActionTest1(),
		buildModule910x8PreStartActionTest2(),
		buildModule910x8PreStartActionTest3(),
		buildModule910x8PreStartActionTest4(),
	}
}

// TestModule910x8PreStartAction test for 910x8 re-scheduling preStartAction
func TestModule910x8PreStartAction(t *testing.T) {
	tests := buildModule910x8PreStartActionTests()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.cacheFuncBefore1()
			tt.args.cacheFuncBefore2()
			if tt.args.cacheFuncBefore3 != nil {
				tt.args.cacheFuncBefore3()
			}
			tt.args.cacheFuncBefore4()
			tp := &module910x8{
				NPUHandler:      tt.fields.baseHandler,
				netUnhealthyKey: tt.fields.netUnhealthyKey,
				affScoreList:    tt.fields.affScoreList,
				reHandle:        tt.fields.reHandle,
			}
			if err := tp.PreStartAction(tt.args.ssn); (err != nil) != tt.wantErr {
				t.Errorf("PreStartAction() error = %v, wantErr %v", err, tt.wantErr)
			}
			tt.args.cacheFuncAfter1()
			tt.args.cacheFuncAfter2()
			if tt.args.cacheFuncAfter3 != nil {
				tt.args.cacheFuncAfter3()
			}
			tt.args.cacheFuncBefore4()
		})
	}
}

func fakeEnvEmpty() plugin.ScheduleEnv {
	schedulerCache := plugin.ScheduleCache{
		Names:      map[string]string{util.RePropertyCacheName: rescheduling.CmName},
		Namespaces: map[string]string{util.RePropertyCacheName: rescheduling.CmNameSpace},
		Data: map[string]map[string]string{
			util.RePropertyCacheName: {rescheduling.CmFaultNodeKind: "", rescheduling.CmFaultJob910x8Kind: ""},
		},
	}
	frameAttr := plugin.VolcanoFrame{
		Conf: []conf.Configuration{
			{
				Name:      util.CMInitParamKey,
				Arguments: map[string]string{rescheduling.GraceOverTimeKey: "800"},
			},
		},
	}
	env := plugin.ScheduleEnv{
		Jobs:      make(map[api.JobID]plugin.SchedulerJob, util.NPUIndex2),
		Nodes:     make(map[string]plugin.NPUNode, util.NPUIndex4),
		FrameAttr: frameAttr,
		Cache:     schedulerCache,
	}
	return env
}

func fakeEnvAddJobsAndNodesToEnv(env *plugin.ScheduleEnv) {
	job0 := fakeSchedulerJobEmptyTask("job0", "vcjob")
	fakeSchedulerJobAddTask(&job0, "pod0", "vcjob", test.NPUIndex8)
	fakeSchedulerJobAddTask(&job0, "pod1", "vcjob", test.NPUIndex8)
	job1 := fakeSchedulerJobEmptyTask("job1", "vcjob")
	fakeSchedulerJobAddTask(&job1, "pod0", "vcjob", test.NPUIndex8)
	fakeSchedulerJobAddTask(&job1, "pod1", "vcjob", test.NPUIndex8)
	node0 := fakeNPUNodeUnhealthy("node0", []string{"Ascend910-0"}, []string{})
	node1 := fakeNPUNodeUnhealthy("node1", []string{}, []string{})
	node2 := fakeNPUNodeUnhealthy("node2", []string{}, []string{})
	node3 := fakeNPUNodeUnhealthy("node3", []string{}, []string{})
	env.Nodes = map[string]plugin.NPUNode{
		"node0": node0,
		"node1": node1,
		"node2": node2,
		"node3": node3,
	}
	env.Jobs = map[api.JobID]plugin.SchedulerJob{
		"vcjob/job0": job0,
		"vcjob/job1": job1,
	}
}

func fakeEnvAddCacheFaultJobToEnv(env *plugin.ScheduleEnv, paras []string, rankIdCreateTime int64,
	podCreateTime int64) {
	if len(paras) < util.NPUIndex3 {
		return
	}
	jobName := paras[zero]
	node0 := paras[one]
	node1 := paras[two]
	faultTask1 := fakeReSchedulerFaultTask(true, []string{"pod0", "vcjob", node0, jobName, "0"}, podCreateTime,
		"ppppppppppppp")
	faultTask2 := fakeReSchedulerFaultTask(false, []string{"pod1", "vcjob", node1, jobName, "1"}, podCreateTime,
		"ppppppppppppp")
	faultJob := fakeReSchedulerFaultJobEmptyTask(jobName, "vcjob", time.Now().Unix()-1,
		rankIdCreateTime, true)
	fakeReSchedulerFaultJobAddTask(&faultJob, faultTask1)
	fakeReSchedulerFaultJobAddTask(&faultJob, faultTask2)
	faultJobString := dealMarshal([]rescheduling.FaultJob{faultJob})
	env.Cache.Data[rescheduling.RePropertyName][rescheduling.CmFaultJob910x8Kind] = faultJobString
}

func fakeEnvAddCacheFaultNodeToEnv(env *plugin.ScheduleEnv) {
	now := time.Now().Unix()
	fCard1 := fakeReSchedulerFaultCard("Ascend910-0", "node0", true, rescheduling.CardUnhealthy)
	fCard2 := fakeReSchedulerFaultCard("Ascend910-1", "node0", false, rescheduling.CardHealthy)
	fCard3 := fakeReSchedulerFaultCard("Ascend910-2", "node0", false, rescheduling.CardHealthy)
	fCard4 := fakeReSchedulerFaultCard("Ascend910-3", "node0", false, rescheduling.CardHealthy)
	fCard5 := fakeReSchedulerFaultCard("Ascend910-4", "node0", false, rescheduling.CardHealthy)
	fCard6 := fakeReSchedulerFaultCard("Ascend910-5", "node0", false, rescheduling.CardHealthy)
	fCard7 := fakeReSchedulerFaultCard("Ascend910-6", "node0", false, rescheduling.CardHealthy)
	fCard8 := fakeReSchedulerFaultCard("Ascend910-7", "node0", false, rescheduling.CardHealthy)
	fNode := fakeReSchedulerFaultNodeEmptyCard("node0", []string{"Ascend910-0"}, []string{},
		heartbeatInterval, []int64{now - 1, fakeTime, now - 1})
	fakeReSchedulerFaultNodeAddFaultCards(&fNode, fCard1)
	fakeReSchedulerFaultNodeAddFaultCards(&fNode, fCard2)
	fakeReSchedulerFaultNodeAddFaultCards(&fNode, fCard3)
	fakeReSchedulerFaultNodeAddFaultCards(&fNode, fCard4)
	fakeReSchedulerFaultNodeAddFaultCards(&fNode, fCard5)
	fakeReSchedulerFaultNodeAddFaultCards(&fNode, fCard6)
	fakeReSchedulerFaultNodeAddFaultCards(&fNode, fCard7)
	fakeReSchedulerFaultNodeAddFaultCards(&fNode, fCard8)
	faultNodes := []rescheduling.FaultNode{fNode}
	faultNodesMarshaled := dealMarshal(faultNodes)
	env.Cache.Data[rescheduling.RePropertyName][rescheduling.CmFaultNodeKind] = faultNodesMarshaled
}

func fakeReSchedulerFaultNodeEmptyCard(nodeName string, unhealthyNPU []string, netUnhealthyNPU []string,
	HBinterval int, updateTimes []int64) rescheduling.FaultNode {
	var isFault bool
	if len(updateTimes) < util.NPUIndex3 {
		return rescheduling.FaultNode{}
	}
	updateTime := updateTimes[zero]
	oldHBTime := updateTimes[one]
	updateHBTime := updateTimes[two]
	hState := rescheduling.NodeHealthy
	if len(netUnhealthyNPU) > 0 {
		isFault = true
		hState = rescheduling.NodeCardNetworkUnhealthy
	}
	if len(unhealthyNPU) > 0 {
		isFault = true
		hState = rescheduling.NodeCardUnhealthy
	}

	faultNode := rescheduling.FaultNode{
		NodeName:            nodeName,
		UpdateTime:          updateTime,
		UnhealthyNPU:        unhealthyNPU,
		NetworkUnhealthyNPU: netUnhealthyNPU,
		IsFaultNode:         isFault,
		NodeDEnable:         true,
		NodeHealthState:     hState,
		AllCards: []string{"Ascend910-0,Ascend910-1,Ascend910-2,Ascend910-3,Ascend910-4,Ascend910-5," +
			"Ascend910-6,Ascend910-7"},
		FaultCards:          make([]rescheduling.FaultCard, util.MapInitNum),
		HeartbeatInterval:   HBinterval,
		OldHeartbeatTime:    oldHBTime,
		UpdateHeartbeatTime: updateHBTime,
	}
	return faultNode
}

func fakeReSchedulerFaultCard(name, nodeName string, isFault bool, faultType string) rescheduling.FaultCard {
	faultCard := rescheduling.FaultCard{
		IsFaultCard: isFault,
		NPUName:     name,
		NodeName:    nodeName,
		FaultType:   faultType,
	}
	return faultCard
}

func fakeReSchedulerFaultNodeAddFaultCards(fNode *rescheduling.FaultNode, fCard rescheduling.FaultCard) {
	fNode.FaultCards = append(fNode.FaultCards, fCard)
}

func fakeReSchedulerFaultJobEmptyTask(jobName, jobNamespace string, updateTime int64,
	rankCreateTime int64, isFault bool) rescheduling.FaultJob {
	faultJob := rescheduling.FaultJob{
		ReScheduleKey:       "grace",
		IsFaultJob:          isFault,
		IsInSession:         true,
		JobName:             jobName,
		JobUID:              api.JobID(jobNamespace + "/" + jobName),
		JobNamespace:        jobNamespace,
		JobRankIds:          []string{},
		NodeNames:           []string{"node0", "node1"},
		FaultTasks:          []rescheduling.FaultTask{},
		UpdateTime:          updateTime,
		JobRankIdCreateTime: rankCreateTime,
	}
	return faultJob
}

func fakeReSchedulerFaultJobAddTask(fJob *rescheduling.FaultJob, fTask rescheduling.FaultTask) {
	fJob.FaultTasks = append(fJob.FaultTasks, fTask)
}

func fakeReSchedulerFaultTask(isFault bool, paras []string,
	podCreateTime int64, podUID types.UID) rescheduling.FaultTask {
	if len(paras) < test.NPUIndex5 {
		return rescheduling.FaultTask{}
	}
	name := paras[zero]
	ns := paras[one]
	nodeName := paras[two]
	jobName := paras[three]
	rankIndex := paras[four]
	faultTask := rescheduling.FaultTask{
		IsFaultTask:   isFault,
		TaskUID:       api.TaskID(`"` + ns + `"-"` + name + `"`),
		TaskName:      name,
		TaskNamespace: ns,
		NodeName:      nodeName,
		JobName:       jobName,
		NodeRankIndex: rankIndex,
		UseCardName:   []string{"Ascend910-1,Ascend910-2,Ascend910-3,Ascend910-4,Ascend910-5,Ascend910-6,Ascend910-7"},
		PodCreateTime: podCreateTime,
		PodUID:        podUID,
	}
	return faultTask
}

func fakeSchedulerJobEmptyTask(jobName, namespace string) plugin.SchedulerJob {
	job0 := plugin.SchedulerJob{
		SchedulerJobAttr: util.SchedulerJobAttr{
			ComJob: util.ComJob{
				JobName:   api.JobID(jobName),
				NameSpace: namespace,
				Selector:  map[string]string{util.AcceleratorType: util.ModuleAcceleratorType},
				Label: map[string]string{
					rescheduling.JobRescheduleLabelKey: rescheduling.JobGraceRescheduleLabelValue,
				},
			},
			NPUJob: &util.NPUJob{
				ReqNPUName: util.NPU910CardName,
				ReqNPUNum:  0,
				Tasks:      make(map[string]util.NPUTask, util.NPUIndex2),
			},
		},
	}
	return job0
}

func fakeSchedulerJobAddTask(sJob *plugin.SchedulerJob, taskName, ns string, reqNPUNum int) {
	task := util.NPUTask{
		TaskName:   taskName,
		ReqNPUName: util.NPU910CardName,
		ReqNPUNum:  reqNPUNum,
		Selector:   nil,
	}
	sJob.Tasks[`"`+ns+`"`+"-"+`"`+taskName+`"`] = task
	sJob.ReqNPUNum += reqNPUNum
}

func fakeNPUNodeUnhealthy(nodeName string, unHealthyCard []string, networkUnhealthyCard []string) plugin.NPUNode {
	allCard := []string{"Ascend910-0", "Ascend910-1,Ascend910-2,Ascend910-3,Ascend910-4,Ascend910-5,Ascend910-6," +
		"Ascend910-7"}
	var faultCards []string
	faultCards = append(faultCards, unHealthyCard...)
	faultCards = append(faultCards, networkUnhealthyCard...)
	var healthyCards []string
	var flag bool
	for _, card := range allCard {
		flag = false
		for _, fCardName := range faultCards {
			if card == fCardName {
				flag = true
				break
			}
		}
		if !flag {
			healthyCards = append(healthyCards, card)
		}
	}

	node0 := plugin.NPUNode{
		Name:       nodeName,
		Capability: nil,
		Allocate:   nil,
		Idle:       nil,
		Annotation: map[string]string{
			util.NPU910CardName: strings.Join(healthyCards, ","),
			util.NPU910CardName + "-" + rescheduling.CardUnhealthy:        strings.Join(unHealthyCard, ","),
			util.NPU910CardName + "-" + rescheduling.CardNetworkUnhealthy: strings.Join(networkUnhealthyCard, ","),
		},
		Label: nil,
	}
	return node0
}

func dealMarshal(data interface{}) string {
	dataString, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	return string(dataString)
}

func fakeBaseHandlerEmpty(env plugin.ScheduleEnv) base.NPUHandler {
	return base.NPUHandler{
		SchedulerPlugin:  plugin.SchedulerPlugin{},
		SchedulerJobAttr: env.Jobs["vcjob/job0"].SchedulerJobAttr,
		ScheduleEnv:      env,
		MaxNodeNPUNum:    0,
		MaxCardNPUNum:    0,
	}
}

func fakeFaultCM(env plugin.ScheduleEnv) *v1.ConfigMap {
	cmData := make(map[string]string, util.MapInitNum)
	cmData[rescheduling.CmFaultNodeKind] = env.Cache.Data[rescheduling.RePropertyName][rescheduling.CmFaultNodeKind]
	cmData[rescheduling.CmFaultJob910x8Kind] =
		env.Cache.Data[rescheduling.RePropertyName][rescheduling.CmFaultJob910x8Kind]
	cmData[rescheduling.CmNodeHeartbeatKind] = ""
	cmData[rescheduling.CmNodeRankTimeMapKind] = ""
	var faultCM = &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rescheduling.CmName,
			Namespace: rescheduling.CmNameSpace,
		},
		Data: cmData,
	}
	return faultCM
}

func fakeReSchedulerNew(env plugin.ScheduleEnv) rescheduling.ReScheduler {
	reScheduler := rescheduling.ReScheduler{
		GraceDeleteTime: 0,
		Level:           "",
		Jobs:            env.Jobs,
		Nodes:           env.Nodes,
		DealReSchedulerCache: &rescheduling.DealReSchedulerCache{
			FaultNodes:               nil,
			FaultJobs:                nil,
			DealReSchedulerConfigmap: nil,
		},
	}
	return reScheduler
}
