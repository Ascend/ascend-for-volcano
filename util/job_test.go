/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*

Package util is using for the total variable.

*/
package util

import "testing"

type testIsSelectorMeetJobArgs struct {
	jobSelectors map[string]string
	conf         map[string]string
}

type testIsSelectorMeetJobTest struct {
	name string
	args testIsSelectorMeetJobArgs
	want bool
}

func buildTestIsSelectorMeetJobTest() []testIsSelectorMeetJobTest {
	tests := []testIsSelectorMeetJobTest{
		{
			name: "01-IsSelectorMeetJob nil jobSelector test.",
			args: testIsSelectorMeetJobArgs{jobSelectors: nil, conf: nil},
			want: true,
		},
		{
			name: "02-IsSelectorMeetJob conf no job selector test.",
			args: testIsSelectorMeetJobArgs{jobSelectors: map[string]string{"haha": "test"}, conf: nil},
			want: false,
		},
		{
			name: "03-IsSelectorMeetJob jobSelector no have conf test.",
			args: testIsSelectorMeetJobArgs{jobSelectors: map[string]string{"haha": "test"},
				conf: map[string]string{"haha": "what"}},
			want: false,
		},
	}
	return tests
}

// TestIsSelectorMeetJob test IsSelectorMeetJob.
func TestIsSelectorMeetJob(t *testing.T) {
	tests := buildTestIsSelectorMeetJobTest()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSelectorMeetJob(tt.args.jobSelectors, tt.args.conf); got != tt.want {
				t.Errorf("IsSelectorMeetJob() = %v, want %v", got, tt.want)
			}
		})
	}
}
