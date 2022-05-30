$ErrorActionPreference = "Stop";

mkdir -force /host/var/run/secrets/kubernetes.io/serviceaccount
cp -force /var/run/secrets/kubernetes.io/serviceaccount/* /host/var/run/secrets/kubernetes.io/serviceaccount/
mkdir -force /host/k/kubefay/etc/
cp /k/kubefay/cni/* /host/opt/cni/bin/
cp /etc/kubefay/kubefay-agent.conf /host/k/kubefay/etc/

cp /etc/kubefay/kubefay-cni.conflist /host/etc/cni/net.d/10-kubefay.conflist
