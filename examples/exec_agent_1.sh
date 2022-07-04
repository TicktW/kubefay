kubectl -n kube-system exec -it  `kubectl -n kube-system get po | grep antrea-agent | tail -1 | awk  '{print $1;}'` -c antrea-ovs -- bash

