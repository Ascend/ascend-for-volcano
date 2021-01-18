# Copyright 2019 The Volcano Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# 202006028:add support for arm64.
# Huawei technologies Co.,Ltd.

BIN_DIR=_output/bin
RELEASE_DIR=_output/release
REL_OSARCH_ROOT=linux/
REPO_PATH=volcano.sh/volcano
IMAGE_PREFIX=volcanosh/vc

include Makefile.def

.EXPORT_ALL_VARIABLES:

REL_OSARCH=$(REL_OSARCH_ROOT)
ARMARCH=arm64
X86ARCH=amd64
MACHINE_ARCH=$(shell uname -m)
ifeq ($(MACHINE_ARCH),aarch64)
        REL_OSARCH=$(REL_OSARCH_ROOT)$(ARMARCH)
else
        REL_OSARCH=$(REL_OSARCH_ROOT)$(X86ARCH)
endif

all: vc-scheduler vc-controller-manager vc-webhook-manager vcctl command-lines

init:
	mkdir -p ${BIN_DIR}
	mkdir -p ${RELEASE_DIR}

vc-scheduler: init
	go build -ldflags ${LD_FLAGS} -o=${BIN_DIR}/vc-scheduler ./cmd/scheduler

vc-controller-manager: init
	go build -ldflags ${LD_FLAGS} -o=${BIN_DIR}/vc-controller-manager ./cmd/controller-manager

vc-webhook-manager: init
	go build -ldflags ${LD_FLAGS} -o=${BIN_DIR}/vc-webhook-manager ./cmd/webhook-manager

vcctl: init
	go build -ldflags ${LD_FLAGS} -o=${BIN_DIR}/vcctl ./cmd/cli

image_bins: init
	go get github.com/mitchellh/gox
	CGO_ENABLED=0 gox -osarch=${REL_OSARCH} -ldflags ${LD_FLAGS} -output ${BIN_DIR}/${REL_OSARCH}/vcctl ./cmd/cli
	for name in controller-manager scheduler webhook-manager; do\
		CGO_ENABLED=0 gox -osarch=${REL_OSARCH} -ldflags ${LD_FLAGS} -output ${BIN_DIR}/${REL_OSARCH}/vc-$$name ./cmd/$$name; \
	done

images: image_bins
	for name in controller-manager scheduler webhook-manager; do\
		cp ${BIN_DIR}/${REL_OSARCH}/vc-$$name ./installer/dockerfile/$$name/; \
		docker build --no-cache -t $(IMAGE_PREFIX)-$$name:$(TAG) ./installer/dockerfile/$$name; \
		rm installer/dockerfile/$$name/vc-$$name; \
	done

webhook-manager-base-image:
	docker build --no-cache -t $(IMAGE_PREFIX)-webhook-manager-base:$(TAG) ./installer/dockerfile/webhook-manager/ -f ./installer/dockerfile/webhook-manager/Dockerfile.base;

generate-code:
	./hack/update-gencode.sh

unit-test:
	go clean -testcache
	go list ./... | grep -v e2e | xargs go test -v -race

e2e-test-kind:
	./hack/run-e2e-kind.sh

generate-yaml: init
	./hack/generate-yaml.sh TAG=${RELEASE_VER}

release-env:
	./hack/build-env.sh release

dev-env:
	./hack/build-env.sh dev

release: images generate-yaml
	./hack/publish.sh

clean:
	rm -rf _output/
	rm -f *.log

verify:
	hack/verify-gofmt.sh
	hack/verify-golint.sh
	hack/verify-gencode.sh
	hack/verify-vendor.sh

verify-generated-yaml:
	./hack/check-generated-yaml.sh

command-lines:
	go build -ldflags ${LD_FLAGS} -o=${BIN_DIR}/vcancel ./cmd/cli/vcancel
	go build -ldflags ${LD_FLAGS} -o=${BIN_DIR}/vresume ./cmd/cli/vresume
	go build -ldflags ${LD_FLAGS} -o=${BIN_DIR}/vsuspend ./cmd/cli/vsuspend
	go build -ldflags ${LD_FLAGS} -o=${BIN_DIR}/vjobs ./cmd/cli/vjobs
	go build -ldflags ${LD_FLAGS} -o=${BIN_DIR}/vqueues ./cmd/cli/vqueues
	go build -ldflags ${LD_FLAGS} -o=${BIN_DIR}/vsub ./cmd/cli/vsub
