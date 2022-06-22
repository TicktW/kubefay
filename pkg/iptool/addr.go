package iptool

import "github.com/TicktW/kubefay/pkg/executor"

type IPAddr struct {
	exec executor.Executor
}

func NewIPAddr(name string) (*IPAddr, error) {
	IPLinkExec, err := executor.AgentPool.Get(IP_LINK_NAME)
	if err != nil {
		return nil, err
	}

	return &IPAddr{
		exec: IPLinkExec,
	}, nil
}

func (ipaddr *IPAddr) Add(addr string, dev string) error {
	// ip addr add 192.168.0.1/24 dev vnet0
	res := ipaddr.exec.Execute("add", addr, "dev", dev)
	return res.GetError()
}
