/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package module910x8 is using for HuaWei A800/9000 Ascend910 pin affinity schedule.

*/
package module910x8

//type restartFaultJobArgs struct {
//	ssn   *framework.Session
//	fJobs []rescheduling.FaultNPUJob
//	jobs  map[string]*api.JobInfo
//}
//
//type restartFaultJobTests struct {
//	name    string
//	args    restartFaultJobArgs
//	wantErr bool
//}
//
//func buildRestartFaultJobTestCases() []restartFaultJobTests {
//	ssn := test.FakeNormalSSN()
//	fakeJob1 := test.FakeNormalTestJob("pg0", util.NPUIndex3)
//	fakeJob2 := test.FakeNormalTestJob("pg1", util.NPUIndex3)
//	mapJob1 := map[string]*api.JobInfo{fakeJob1.Name: fakeJob1}
//	mapJob2 := map[string]*api.JobInfo{fakeJob2.Name: fakeJob2}
//
//	testCases := []restartFaultJobTests{
//		{
//			name: "test-RestartFaultJobTestCases()\ncase0: no fault job configured"
//			args: restartFaultJobArgs{jobs: mapJob1, fJobs: },
//		}
//	}
//
//	return testCases
//}
//
//func Test_restartFaultJob(t *testing.T) {
//	tests := buildRestartFaultJobTestCases()
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if err := restartFaultJob(tt.args.ssn, tt.args.fJobs, tt.args.jobs); (err != nil) != tt.wantErr {
//				t.Errorf("restartFaultJob() error = %v, wantErr %v", err, tt.wantErr)
//			}
//		})
//	}
//}