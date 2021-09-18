#!/bin/bash

# Perform  rebuild volcano-huawei-npu-scheduler plugin
# Copyright @ Huawei Technologies CO., Ltd. 2020-2021. All rights reserved

#set -e
export GOPROXY=https://goproxy.cn,direct
export GOPATH=/usr1/workspace/C20_AtlasDLS_DailyBuild_Cloud/go/

unset http_proxy
unset https_proxy

cd "${GOPATH}"/src/volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/output/
kubectl delete -f volcano-v1.3.0.yaml
sleep 3
docker image rm "$(docker images|grep vc-scheduler|awk '{print $3}')"
docker image rm "$(docker images|grep vc-controller|awk '{print $3}')"

cd "${GOPATH}"/src/volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/build/
./build.sh

cd "${GOPATH}"/src/volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/output/
rm -rf plugins

docker build --no-cache -t volcanosh/vc-scheduler:v1.3.0 ./ -f ./Dockerfile-scheduler
docker build --no-cache -t volcanosh/vc-controller-manager:v1.3.0 ./ -f ./Dockerfile-controller

kubectl apply -f volcano-v1.3.0.yaml
sleep 6
kubectl get pod -A -o wide