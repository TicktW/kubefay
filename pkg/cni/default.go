// Copyright 2020 Antrea Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build !windows
// +build !windows

package cni

import (
	"os"
	"path/filepath"

	"k8s.io/klog/v2"
)

// CNISocketAddr is the UNIX socket used by the CNI Protobuf / gRPC service.
var CNISocketAddr = "/var/run/kubefay/cni.sock"
var IPAMCNISocketAddr = "/var/run/kubefay/cni.ipam.sock"

// same code to build two binaries both of kubefay-cni and kubefay-ipam-cni
// IPAMBuild should be marked at compile time
var CNIRequestAddr string

func init() {
	path, _ := os.Executable()
	_, exec := filepath.Split(path)
	if exec == "kubefay-cni" {
		CNIRequestAddr = CNISocketAddr
	} else if exec == "kubefay-ipam-cni" {
		CNIRequestAddr = IPAMCNISocketAddr
	} else {
		klog.V(5).Infof("not in CNI request mode, CNIRequestAddr is: %s", CNIRequestAddr)
	}
}
