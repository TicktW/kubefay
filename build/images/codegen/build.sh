GO_VERSION=$(head -n 1 ../deps/go-version)

docker build -t kubefay/codegen --build-arg GO_VERSION=${GO_VERSION} --build-arg GOPROXY=${GOPROXY} .