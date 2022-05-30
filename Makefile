SHELL              := /bin/bash
# go options
GO                 ?= go
LDFLAGS            :=
GOFLAGS            :=
BINDIR             ?= $(CURDIR)/bin
GO_FILES           := $(shell find . -type d -name '.cache' -prune -o -type f -name '*.go' -print)
GOPATH             ?= $$($(GO) env GOPATH)
DOCKER_CACHE       := $(CURDIR)/.cache
ANTCTL_BINARY_NAME ?= antctl
OVS_VERSION        := $(shell head -n 1 build/images/deps/ovs-version)
GO_VERSION         := $(shell head -n 1 build/images/deps/go-version)

DOCKER_BUILD_ARGS = --build-arg OVS_VERSION=$(OVS_VERSION)
DOCKER_BUILD_ARGS += --build-arg GO_VERSION=$(GO_VERSION)

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

.PHONY: ubuntu
ubuntu:
	@echo "===> Building antrea/antrea-ubuntu Docker image <==="
ifneq ($(NO_PULL),)
	docker build -t antrea/antrea-ubuntu:$(DOCKER_IMG_VERSION) -f build/images/Dockerfile.ubuntu $(DOCKER_BUILD_ARGS) .
else
	docker build --pull -t antrea/antrea-ubuntu:$(DOCKER_IMG_VERSION) -f build/images/Dockerfile.ubuntu $(DOCKER_BUILD_ARGS) .
endif
	docker tag antrea/antrea-ubuntu:$(DOCKER_IMG_VERSION) antrea/antrea-ubuntu
	docker tag antrea/antrea-ubuntu:$(DOCKER_IMG_VERSION) projects.registry.vmware.com/antrea/antrea-ubuntu
	docker tag antrea/antrea-ubuntu:$(DOCKER_IMG_VERSION) projects.registry.vmware.com/antrea/antrea-ubuntu:$(DOCKER_IMG_VERSION)


.PHONY: build-ubuntu
build-ubuntu:
	@echo "===> Building Antrea bins and antrea/antrea-ubuntu Docker image <==="
	docker build -t antrea/antrea-ubuntu:$(DOCKER_IMG_VERSION) -f build/images/Dockerfile.build.ubuntu $(DOCKER_BUILD_ARGS) .