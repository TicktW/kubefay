K8S_VERSION=1.24.0
KUBEOPENAPI_VERSION=v0.0.0-20220328201542-3ee0da9b0b42
    
go install k8s.io/code-generator/cmd/client-gen@kubernetes-$K8S_VERSION && \
go install k8s.io/code-generator/cmd/deepcopy-gen@kubernetes-$K8S_VERSION && \
go install k8s.io/code-generator/cmd/conversion-gen@kubernetes-$K8S_VERSION && \
go install k8s.io/code-generator/cmd/lister-gen@kubernetes-$K8S_VERSION && \
go install k8s.io/code-generator/cmd/informer-gen@kubernetes-$K8S_VERSION && \
go install k8s.io/kube-openapi/cmd/openapi-gen@$KUBEOPENAPI_VERSION && \
go install k8s.io/code-generator/cmd/go-to-protobuf@kubernetes-$K8S_VERSION && \
go install k8s.io/code-generator/cmd/go-to-protobuf/protoc-gen-gogo@kubernetes-$K8S_VERSION && \
go install github.com/golang/mock/mockgen@v1.4.4 && \
go install github.com/golang/protobuf/protoc-gen-go@v1.5.2 && \
go install golang.org/x/tools/cmd/goimports

PROTOBUF_VERSION=3.0.2; ZIPNAME="protoc-${PROTOBUF_VERSION}-linux-x86_64.zip"; \
  mkdir /tmp/protoc && cd /tmp/protoc && \
  wget "https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOBUF_VERSION}/${ZIPNAME}" && \
  unzip "${ZIPNAME}" && \
  chmod -R +rX /tmp/protocc

sudo cp -rp /tmp/protoc /usr/local/bin/
sudo cp -rp /tmp/protoc/include /usr/local/include