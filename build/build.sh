#!/bin/bash
# Perform  build volcano-huawei-npu-scheduler plugin
# Copyright @ Huawei Technologies CO., Ltd. 2020-2021. All rights reserved

set -e

DEFAULT_VER='v2.0.2'
TOP_DIR=${GOPATH}/src/volcano.sh/volcano/
BASE_PATH=${GOPATH}/src/volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/
CMD_PATH=${GOPATH}/src/volcano.sh/volcano/cmd/
PKG_PATH=volcano.sh/volcano/pkg
DATE=$(date "+%Y-%m-%d %H:%M:%S")
BASE_VER=v1.3.0

function parse_version() {
    version_file="${TOP_DIR}"/service_config.ini
    if  [ -f "$version_file" ]; then
      line=$(sed -n '1p' "$version_file" 2>&1)
      version=${line#*:}
      echo ${version}
      return
    fi
    echo ${DEFAULT_VER}
}

function parse_arch() {
   arch=$(arch 2>&1)
   echo "${arch}"
}

REL_VERSION=$(parse_version)
REL_OSARCH=$(parse_arch)

function clean() {
    rm -f "${BASE_PATH}"/output/vc-controller-manager
    rm -f "${BASE_PATH}"/output/vc-scheduler
    rm -f "${BASE_PATH}"/output/*.so
}

function build() {
    echo "Build Architecture is" "${REL_OSARCH}"

    export GO111MODULE=on
    export PATH=$GOPATH/bin:$PATH

    cd "${TOP_DIR}"
    go mod tidy
    go mod download
    go mod vendor

    cd "${BASE_PATH}"/output/

    CGO_CFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2 -fPIC -ftrapv" \
    CGO_CPPFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2 -fPIC -ftrapv" \
    CGO_ENABLED=0 go build -buildmode=pie -ldflags "-s -linkmode=external -extldflags=-Wl,-z,now
    -X '${PKG_PATH}/version.Built=${DATE}' -X '${PKG_PATH}/version.Version=${BASE_VER}'" \
    -o vc-controller-manager "${CMD_PATH}"/controller-manager

    CGO_CFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2 -fPIC -ftrapv" \
    CGO_CPPFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2 -fPIC -ftrapv" \
    CC=/usr/local/musl/bin/musl-gcc \
    CGO_ENABLED=1 go build -buildmode=pie -ldflags "-s -extldflags=-Wl,-z,now
    -X '${PKG_PATH}/version.Built=${DATE}' -X '${PKG_PATH}/version.Version=${BASE_VER}'" \
    -o vc-scheduler "${CMD_PATH}"/scheduler

    CGO_CFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2 -fPIC -ftrapv" \
    CGO_CPPFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2 -fPIC -ftrapv" \
    CC=/usr/local/musl/bin/musl-gcc CGO_ENABLED=1 go build -buildmode=plugin -ldflags \
    "-s -extldflags=-Wl,-z,now" -o volcano-npu_"${REL_VERSION}_linux-${REL_OSARCH}".so \
    "${GOPATH}"/src/volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/

    if [ ! -f "${BASE_PATH}/output/volcano-npu_${REL_VERSION}_linux-${REL_OSARCH}.so" ]
    then
      echo "fail to find volcano-npu-${REL_VERSION}.so"
      exit 1
    fi
    sed -i "s/name: volcano-npu-.*/name: volcano-npu_${REL_VERSION}_linux-${REL_OSARCH}/" "${BASE_PATH}"/output/volcano-*.yaml
    sed -i "s/ value: \"volcano-npu-.*/ value: \"volcano-npu_${REL_VERSION}_linux-${REL_OSARCH}\"/" "${BASE_PATH}"/output/volcano-*.yaml
    sed -i "s/pluginName=volcano-npu-.*/pluginName=volcano-npu_${REL_VERSION}_linux-${REL_OSARCH}/" "${BASE_PATH}"/output/Dockerfile-scheduler
    chmod 400 "${BASE_PATH}"/output/*.so
    chmod 500 vc-controller-manager vc-scheduler
    chmod 400 "${BASE_PATH}"/output/Dockerfile*
    chmod 400 "${BASE_PATH}"/output/volcano-*.yaml
}

function main() {
  clean
  build
}

main "${1}"

echo ""
echo "Finished!"
echo ""