kubectl -n kube-system exec -it  `kubectl -n kube-system get po | grep antrea-agent | head -1 | awk  '{print $1;}'` -c antrea-ovs -- bash

