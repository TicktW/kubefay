.PHONY: build-ubuntu
build-ubuntu:
	@echo "===> Building kubefay bins and kubefay/kubefay-ubuntu Docker image <==="
	docker build -t kubefay/kubefay-ubuntu:$(DOCKER_IMG_VERSION) -f Dockerfile.build.ubuntu $(DOCKER_BUILD_ARGS)

.PHONY: ubuntu
ubuntu:
	@echo "===> Building kubefay/kubefay-ubuntu Docker image <==="
	docker build -t kubefay/kubefay-ubuntu:$(DOCKER_IMG_VERSION) -f Dockerfile.ubuntu $(DOCKER_BUILD_ARGS) .

.PHONY: base-ubuntu
build-ubuntu:
	@echo "===> Building kubefay/base-ubuntu Docker image <==="
	cd base && ./build.sh

.PHONY: ovs-ubuntu
ovs-ubuntu:
	@echo "===> Building kubefay/ovs-ubuntu Docker image <==="
	cd ovs && ./build.sh

.PHONY: all
all: ovs-ubuntu base-ubuntu ubuntu build-ubuntu
