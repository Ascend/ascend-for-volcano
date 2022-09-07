#!/bin/bash
# Perform  test volcano-huawei-npu-scheduler plugin
# Copyright @ Huawei Technologies CO., Ltd. 2020-2022. All rights reserved

set -e

export GO111MODULE=on
export GONOSUMDB="*"
export PATH=$GOPATH/bin:$PATH

cd "${GOPATH}"/src/volcano.sh/volcano
go get github.com/agiledragon/gomonkey/v2@v2.3.0
go mod vendor

file_input='testVolcano.txt'
file_detail_output='api.html'

echo "************************************* Start LLT Test *************************************"
mkdir -p "${GOPATH}"/src/volcano.sh/volcano/_output/test/
cd "${GOPATH}"/src/volcano.sh/volcano/_output/test/
rm -f $file_detail_output $file_input

if  ! go test -v -race -gcflags=all=-l -coverprofile cov.out "${GOPATH}"/src/volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/... \
    > ./$file_input;
then
  echo '****** go test cases error! ******'
  echo 'Failed' > $file_input
  exit 1
else
  gocov convert cov.out | gocov-html >"$file_detail_output"
  gotestsum --junitfile unit-tests.xml "${GOPATH}"/src/volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/...
fi

echo "************************************* End   LLT Test *************************************"
exit 0