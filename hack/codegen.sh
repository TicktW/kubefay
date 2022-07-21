# Generate clientset and apis code with K8s codegen tools.
# code generations tools shoule be installed before this script running

PROJECT_PKG="github.com/kubefay/kubefay"
# export to make a env for child processes
export GOPATH=$(go env GOPATH)

# NOTE: input path is addr base on GOPATH, output package and go-header-file are system paths.
# project root dir must be in GOPATH/src
echo "====================client-gen======================="
$GOPATH/bin/client-gen \
  --clientset-name versioned \
  --input-base "${PROJECT_PKG}/pkg/apis/" \
  --input "agent/v1alpha1" \
  --input "ipam/v1alpha1" \
  --output-package "${PROJECT_PKG}/pkg/client/clientset" \
  --go-header-file ${GOPATH}/src/${PROJECT_PKG}/hack/boilerplate/license_header.txt \
  -v 5


echo "====================lister-gen======================="

$GOPATH/bin/lister-gen \
  --input-dirs "${PROJECT_PKG}/pkg/apis/agent/v1alpha1" \
  --input-dirs "${PROJECT_PKG}/pkg/apis/ipam/v1alpha1" \
  --output-package "${PROJECT_PKG}/pkg/client/listers" \
  --go-header-file ${GOPATH}/src/${PROJECT_PKG}/hack/boilerplate/license_header.txt \
  -v 5

echo "===================informer-gen======================="

$GOPATH/bin/informer-gen \
  --input-dirs "${PROJECT_PKG}/pkg/apis/agent/v1alpha1" \
  --input-dirs "${PROJECT_PKG}/pkg/apis/ipam/v1alpha1" \
  --versioned-clientset-package "${PROJECT_PKG}/pkg/client/clientset/versioned" \
  --listers-package "${PROJECT_PKG}/pkg/client/listers" \
  --output-package "${PROJECT_PKG}/pkg/client/informers" \
  --go-header-file ${GOPATH}/src/${PROJECT_PKG}/hack/boilerplate/license_header.txt \
  -v 5

echo "===================deepcopy-gen======================="
$GOPATH/bin/deepcopy-gen \
  --input-dirs "${PROJECT_PKG}/pkg/apis/agent/v1alpha1" \
  --input-dirs "${PROJECT_PKG}/pkg/apis/ipam/v1alpha1" \
  --go-header-file ${GOPATH}/src/${PROJECT_PKG}/hack/boilerplate/license_header.txt \
  -O zz_generated.deepcopy \
  -v 5

# $GOPATH/bin/conversion-gen  \
#   --input-dirs "${PROJECT_PKG}/pkg/apis/controlplane/v1beta2,${PROJECT_PKG}/pkg/apis/controlplane/" \
#   --input-dirs "${PROJECT_PKG}/pkg/apis/stats/v1alpha1,${PROJECT_PKG}/pkg/apis/stats/" \
#   -O zz_generated.conversion
