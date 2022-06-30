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
IPAMLDFLAGS += $(VERSION_LDFLAGS)
IPAMLDFLAGS += -X github.com/kubefay/TicktW/pkg/cni.AntreaCNISocketAddr=/var/run/antrea/cni.sock.ipam

.PHONY: all
all: build

.PHONY: build
build: build-ubuntu

.PHONY: codegen
codegen:
	@echo "===> generating code ... <==="
	$(CURDIR)/hack/codegen.sh
# protoc --go_out=plugins=grpc:. pkg/apis/cni/v1beta1/cni.proto

.PHONY: rpcgen
rpcgen:
	protoc --go_out=plugins=grpc:. pkg/rpc/cni/v1beta1/cni.proto

.PHONY: build-ubuntu
build-ubuntu:
	@echo "===> Building kubefay bins and kubefay/kubefay-ubuntu Docker image <==="
	docker build -t kubefay/kubefay-ubuntu:$(DOCKER_IMG_VERSION) $(DOCKER_BUILD_ARGS) -f build/images/Dockerfile.ubuntu .
	docker tag kubefay/kubefay-ubuntu:$(DOCKER_IMG_VERSION) kubefay/kubefay-ubuntu:latest

dev-ubuntu:
	@echo "===> Dev env setup <==="
	docker build -t kubefay/kubefay-ubuntu-dev:$(DOCKER_IMG_VERSION) $(DOCKER_BUILD_ARGS) -f build/images/Dockerfile.dev.ubuntu .
	docker tag kubefay/kubefay-ubuntu-dev:$(DOCKER_IMG_VERSION) kubefay/kubefay-ubuntu-dev:latest

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

.PHONY: dev-test
dev-test: 
	make base-ubuntu
	make build-ubuntu
	make cluster
	kubectl apply -f build/deploy/fay-agent-build.yaml

.PHONY: run-dev-docker
run-dev-docker:
	# @docker run --rm --privileged -v ${PWD}:/root/go/src/github.com/TicktW/kubefay -v /dev/net/tun:/dev/net/tun -it kubefay/kubefay-ubuntu-dev:latest bash
	@docker run --rm --privileged -v ${PWD}:/root/go/src/github.com/TicktW/kubefay -v /dev/net/tun:/dev/net/tun -it kubefay/kubefay-ubuntu:latest bash

.PHONY: bin-agent
bin-agent:
	@mkdir -p $(BINDIR)
	GOOS=linux $(GO) build -o $(BINDIR)/kubefay-agent $(GOFLAGS) -ldflags '$(LDFLAGS)' github.com/TicktW/kubefay/cmd/kubefay-agent

.PHONY: bin-cni
bin-cni:
	@mkdir -p $(BINDIR)
	GOOS=linux $(GO) build -o $(BINDIR)/kubefay-cni $(GOFLAGS) -ldflags '$(LDFLAGS)' github.com/TicktW/kubefay/cmd/kubefay-cni

.PHONY: bin-ipam-cni
bin-ipam-cni:
	@mkdir -p $(BINDIR)
	# IPAMLDFLAGS := $(LDFLAGS) + -X github.com/TicktW/kubefay/pkg/cni.AntreaCNISocketAddr=/var/run/antrea/cni.sock.ipam
	# @echo $(IPAMLDFLAGS)
	GOOS=linux $(GO) build -o $(BINDIR)/kubefay-ipam-cni $(GOFLAGS) -ldflags '$(IPAMLDFLAGS)' github.com/TicktW/kubefay/cmd/kubefay-cni

.PHONY: bin
bin:
	@make bin-cni
	@make bin-ipam-cni
	@make bin-agent

.PHONY: clean
clean:
	rm -r $(BINDIR)

.PHONY: manifest-gen
manifest-gen:
	@echo "Generating dev manifest for Kubefay"
	helm template kubefay  --dry-run ./build/helm/kubefay/

.PHONY: manifest-apply
manifest-apply:
	@echo "===> Generating dev manifest for Antrea <==="
	helm template kubefay  --dry-run ./build/helm/kubefay/ | kubectl apply -f -
	kubectl apply -f ./build/helm/kubefay/defaultnet/subnet.yaml