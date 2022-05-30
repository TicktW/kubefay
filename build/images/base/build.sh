#!/usr/bin/env bash

# Copyright 2020 kubefay Authors
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

# This is a very simple script that builds the base image for kubefay and pushes it to
# the kubefay Dockerhub (https://hub.docker.com/u/kubefay). The image is tagged with the OVS version.

set -eo pipefail

function echoerr {
    >&2 echo "$@"
}

OVS_VERSION=$(head -n 1 ../deps/ovs-version)
CNI_BINARIES_VERSION=$(head -n 1 ../deps/cni-binaries-version)
TMP_PLUGINS_DIR=$(cd $(dirname $0);pwd)/plugins
echo $OVS_VERSION

if [ ! -d "${TMP_PLUGINS_DIR}" ];then
   echo "CNI common plugins do not exist, download and build them..."
   git clone git@github.com:containernetworking/plugins.git && pushd ${TMP_PLUGINS_DIR} && ./build_linux.sh
fi

docker build $PLATFORM_ARG --target cni-binaries \
       --cache-from kubefay/cni-binaries:$CNI_BINARIES_VERSION \
       -t kubefay/cni-binaries:$CNI_BINARIES_VERSION \
       --build-arg CNI_BINARIES_VERSION=$CNI_BINARIES_VERSION \
       --build-arg OVS_VERSION=$OVS_VERSION .


docker build $PLATFORM_ARG \
       --cache-from kubefay/cni-binaries:$CNI_BINARIES_VERSION \
       --cache-from antrea/base-ubuntu:$OVS_VERSION \
       -t kubefay/base-ubuntu:$OVS_VERSION \
       --build-arg CNI_BINARIES_VERSION=$CNI_BINARIES_VERSION \
       --build-arg OVS_VERSION=$OVS_VERSION .
