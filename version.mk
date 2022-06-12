DOCKER_IMG_VERSION=0.0.1

VERSION := $(shell head -n 1 VERSION)

VERSION_LDFLAGS = -X github.com/TicktW/kubefay/pkg/version.Version=$(VERSION)
