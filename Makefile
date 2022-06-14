SHELL              := /bin/bash
# go options
GO                 ?= go
LDFLAGS            :=
GOFLAGS            :=
BINDIR             ?= $(CURDIR)/bin
GO_FILES           := $(shell find . -type d -name '.cache' -prune -o -type f -name '*.go' -print)
GOPATH             ?= $$($(GO) env GOPATH)
DOCKER_CACHE       := $(CURDIR)/.cache
BUILD_DIR          := $(CURDIR)/build
DOCKER_IMG_DIR     := $(BUILD_DIR)/images
ANTCTL_BINARY_NAME ?= antctl
OVS_VERSION        := $(shell head -n 1 build/images/deps/ovs-version)
GO_VERSION         := $(shell head -n 1 build/images/deps/go-version)

DOCKER_BUILD_ARGS = --build-arg OVS_VERSION=$(OVS_VERSION)
DOCKER_BUILD_ARGS += --build-arg GO_VERSION=$(GO_VERSION)

include version.mk
LDFLAGS += $(VERSION_LDFLAGS)

.PHONY: all
all: build

.PHONY: antrea-agent
antrea-agent:
	@mkdir -p $(BINDIR)
	GOOS=linux $(GO) build -o $(BINDIR) $(GOFLAGS) -ldflags '$(LDFLAGS)' antrea.io/antrea/cmd/antrea-agent

.PHONY: antrea-controller
antrea-controller:
	@mkdir -p $(BINDIR)
	GOOS=linux $(GO) build -o $(BINDIR) $(GOFLAGS) -ldflags '$(LDFLAGS)' antrea.io/antrea/cmd/antrea-controller

.PHONY: antrea-cni
antrea-cni:
	@mkdir -p $(BINDIR)
	GOOS=linux CGO_ENABLED=0 $(GO) build -o $(BINDIR) $(GOFLAGS) -ldflags '$(LDFLAGS)' antrea.io/antrea/cmd/antrea-cni

.PHONY: build
build: build-ubuntu

.PHONY: codegen
codegen:
	@echo "===> generating code ... <==="
	$(CURDIR)/hack/codegen.sh

.PHONY: build-ubuntu
build-ubuntu:
	@echo "===> Building kubefay bins and kubefay/kubefay-ubuntu Docker image <==="
	docker build -t kubefay/kubefay-ubuntu:$(DOCKER_IMG_VERSION) $(DOCKER_BUILD_ARGS) -f build/images/Dockerfile.ubuntu .
	docker tag kubefay/kubefay-ubuntu:$(DOCKER_IMG_VERSION) kubefay/kubefay-ubuntu:latest

.PHONY: ubuntu
ubuntu:
	@echo "===> Building kubefay/kubefay-ubuntu Docker image <==="
	docker build -t kubefay/kubefay-ubuntu:$(DOCKER_IMG_VERSION) -f Dockerfile.ubuntu $(DOCKER_BUILD_ARGS) .

.PHONY: base-ubuntu
base-ubuntu:
	@echo "===> Building kubefay/base-ubuntu Docker image <==="
	cd $(DOCKER_IMG_DIR)/base && ./build.sh 
	docker tag kubefay/base-ubuntu:$(OVS_VERSION) kubefay/base-ubuntu:latest

.PHONY: ovs-ubuntu
ovs-ubuntu:
	@echo "===> Building kubefay/ovs-ubuntu Docker image <==="
	cd $(DOCKER_IMG_DIR)/ovs && ./build.sh

.PHONY: all
all: ovs-ubuntu base-ubuntu ubuntu build-ubuntu

.PHONY: cluster
cluster:
	kind delete cluster && kind create cluster --config build/cluster/kind-config.yaml 
	sleep 5
	kind get nodes | xargs ./hack/kind-fix-networking.sh
	sleep 5
	kind load docker-image kubefay/kubefay-ubuntu:latest kubefay/kubefay-ubuntu:latest
	sleep 5

.PHONY: bin
bin:
	@mkdir -p $(BINDIR)
	GOOS=linux $(GO) build -o $(BINDIR) $(GOFLAGS) -ldflags '$(LDFLAGS)' github.com/TicktW/kubefay/cmd/...
	./bin/fay-agent

.PHONY: dev
dev: 
	make base-ubuntu
	make build-ubuntu
	make cluster
	kubectl apply -f build/deploy/fay-agent-build.yaml