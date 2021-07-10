#!/bin/sh
# Perform  test volcano-huawei-npu-scheduler plugin
# Copyright @ Huawei Technologies CO., Ltd. 2020-2021. All rights reserved

set -e

export GO111MODULE=on
export GOPROXY=http://cmc.centralrepo.rnd.huawei.com/go/
export GONOSUMDB="*"
export PATH=$GOPATH/bin:$PATH

cd "${GOPATH}"/src/volcano.sh/volcano
go mod vendor

file_input='testVolcano.txt'
file_detail_output='api.html'

echo "************************************* Start LLT Test *************************************"
mkdir -p "${GOPATH}"/src/volcano.sh/volcano/_output/test/
cd "${GOPATH}"/src/volcano.sh/volcano/_output/test/
rm -rf $file_detail_output $file_input

if  ! go test -v -race -coverprofile cov.out "${GOPATH}"/src/volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/... \
    > ./$file_input;
then
  echo '****** go test cases error! ******'
  echo 'Failed' > $file_input
else
  gocov convert cov.out | gocov-html >"$file_detail_output"
  gotestsum --junitfile unit-tests.xml "${GOPATH}"/src/volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/...
fi

echo "************************************* End   LLT Test *************************************"