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
# IPAMLDFLAGS += -X github.com/TicktW/kubefay/pkg/cni.IPAMBuild=true

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

.PHONY: cluster-load-image
cluster-load-image:
	kind load docker-image kubefay/kubefay-ubuntu:latest kubefay/kubefay-ubuntu:latest

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

.PHONY: bin-test
bin-test:
	@mkdir -p $(BINDIR)
	GOOS=linux $(GO) build -o $(BINDIR)/kubefay-test-server  github.com/TicktW/kubefay/ci/e2e/kubefay-test-server
	GOOS=linux $(GO) build -o $(BINDIR)/kubefay-test-cli github.com/TicktW/kubefay/ci/e2e/kubefay-test-cli

.PHONY: bin
bin:
	make bin-cni
	make bin-ipam-cni
	make bin-agent
	make bin-test

.PHONY: clean
clean:
	rm -r $(BINDIR)

.PHONY: manifest-gen-kind
manifest-gen-kind:
	@echo "Generating dev manifest for Kubefay"
	helm template kubefay  --dry-run ./build/helm/kubefay/

.PHONY: manifest-gen
manifest-gen:
	@echo "Generating dev manifest for Kubefay"
	helm template kubefay --set kindCluster.enabled=false  --dry-run ./build/helm/kubefay/ > kubefay.yaml

.PHONY: manifest-apply-kind
manifest-apply-kind:
	@echo "===> Generating dev manifest for Antrea <==="
	helm template kubefay  --dry-run ./build/helm/kubefay/ | kubectl apply -f -
	kubectl apply -f ./build/helm/kubefay/defaultnet/subnet.yaml

.PHONY: manifest-apply
manifest-apply:
	@echo "===> Generating dev manifest for Antrea <==="
	helm template kubefay --set kindCluster.enabled=false --dry-run ./build/helm/kubefay/ | kubectl apply -f -
	kubectl apply -f ./build/helm/kubefay/defaultnet/subnet.yaml

.PHONY: kube-clean-pod
kube-clean-pod:
	kubectl -n kube-system delete po $$( kubectl -n kube-system get po | grep kubefay-agent | awk '{print $$1}' | xargs)

.PHONY: kube-get-pod
kube-get-pod:
	kubectl -n kube-system get po | grep kubefay-agent

.PHONY: kube-log-pod
kube-log-pod:
	kubectl -n kube-system logs $$(kubectl -n kube-system get po | grep kubefay-agent | awk '{print $$1}' | xargs | awk '{print $$1}')

.PHONY: kube-exec-agent-master
kube-exec-agent-master:
	kubectl -n kube-system exec -it  $$(kubectl -n kube-system get po -o wide | grep kubefay-agent | grep kind-control-plane | awk '{print $$1}') -- bash

.PHONY: kube-exec-agent-worker
kube-exec-agent-worker:
	kubectl -n kube-system exec -it  $$(kubectl -n kube-system get po -o wide | grep kubefay-agent | grep kind-worker | awk '{print $$1}') -- bash


.PHONY: kube-exec-agent-master-ovs
kube-exec-agent-master-ovs:
	kubectl -n kube-system exec -it  $$(kubectl -n kube-system get po -o wide | grep kubefay-agent | grep kind-control-plane | awk '{print $$1}') --container kubefay-ovs -- bash

.PHONY: kube-exec-worker-openflow
kube-exec-worker-openflow:
	kubectl -n kube-system exec -it  $$(kubectl -n kube-system get po -o wide | grep kubefay-agent | grep kind-worker | awk '{print $$1}') --container kubefay-ovs -- ovs-ofctl dump-flows br-int

.PHONY: kube-exec-master-openflow
kube-exec-master-openflow:
	kubectl -n kube-system exec -it  $$(kubectl -n kube-system get po -o wide | grep kubefay-agent | grep kind-control-plane | awk '{print $$1}') --container kubefay-ovs -- ovs-ofctl dump-flows br-int

# ks exec -it kubefay-agent-cpvwj --container kubefay-ovs -- bash
.PHONY: dev-small-round
dev-small-round:
	make bin
	make build-ubuntu
	make cluster-load-image
	make kube-clean-pod
	make kube-get-pod
	make kube-log-pod

.PHONY: dev-big-round
dev-big-round:
	make bin
	make build-ubuntu
	make cluster
	make cluster-load-image
	sleep 10
	make manifest-apply-kind
	sleep 5
	make test-app-apply
	sleep 5
	make kube-get-pod
	make kube-log-pod
	sleep 30
	make test-e2e

.PHONY: exec-kind-node
exec-kind-node:
	docker exec -it $$(docker ps | grep kindest | grep node | awk '{print $$1}' | head -1) bash

.PHONY: test-app-apply
test-app-apply:
	kubectl apply -f examples/app-master.yml
	kubectl apply -f examples/app-worker.yml
	kubectl apply -f examples/subnet-new.yml
	kubectl apply -f examples/ns-new.yml
	kubectl apply -f examples/app-master-new.yml
	kubectl apply -f examples/app-worker-new.yml

# ovs-appctl ofproto/trace br-int
.PHONY: minikube-img-load
minikube-img-load:
	minikube image load kubefay/kubefay-ubuntu:latest

.PHONY: minikube-img-reload
minikube-img-reload:
	minikube image tag kubefay/kubefay-ubuntu:latest kubefay/kubefay-ubuntu:0.0.1
	minikube image load kubefay/kubefay-ubuntu:latest

.PHONY: minikube-cluster
minikube-cluster:
	minikube start --driver kvm2 --registry-mirror=https://registry.docker-cn.com --image-mirror-country cn --extra-config=kubelet.cgroup-driver=systemd --kubernetes-version=v1.18.8 -n 1 --cni=kubefay.yaml

.PHONY: minikube-dev-round
minikube-dev-round:
	minikube delete
	make bin
	make build-ubuntu
	make minikube-cluster
	make minikube-img-load

.PHONY: test-e2e
test-e2e:
	go clean -testcache && go test ./ci/e2e/ -v