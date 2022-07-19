kind delete cluster
sleep 5
kind create cluster --config ./config-3nodes.yml
sleep 1
kind get nodes | xargs ./hack/kind-fix-networking.sh
sleep 1
kind load docker-image projects.registry.vmware.com/antrea/antrea-ubuntu:latest
