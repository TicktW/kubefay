for node_name in $(minikube node list | awk '{print $1}' | xargs)
    do
        echo ${node_name}
        node_ip=$(minikube ssh-host -n ${node_name} | awk '{print $1}')
        minikube ssh-host -n ${node_name} --append-known=true
        minikube ssh-key -n ${node_name}

        ssh -i ~/.minikube/machines/${node_name}/id_rsa docker@${node_ip} sudo rm -f /etc/cni/net.d/87-podman-bridge.conflist
        ssh -i ~/.minikube/machines/${node_name}/id_rsa docker@${node_ip} sudo rm -f /etc/cni/net.d/1-k8s.conflist
        ssh -i ~/.minikube/machines/${node_name}/id_rsa docker@${node_ip} sudo rm -f /etc/cni/net.d/10-kindnet.conflist
done

# kubectl -n kube-system delete ds kindnet
kubectl -n kube-system delete pod $(kubectl -n kube-system get po | grep coredns | awk '{print $1}')