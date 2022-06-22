package iptool

import "github.com/TicktW/kubefay/pkg/executor"

type IPLink struct{
	name string
	exec executor.Executor
}

func NewIPLink(name string) (*IPLink, error) {
	IPLinkExec, err := executor.AgentPool.Get(IP_LINK_NAME)
	if err != nil {
		return nil, err
	}

	return &IPLink{
		name: name,
		exec: IPLinkExec,
	}, nil
}

func (iplink *IPLink) Add(linkeType string) error {
	// add-br BRIDGE PARENT VLAN   create new fake BRIDGE in PARENT on VLAN
	res := iplink.exec.Execute("add", iplink.name, "type", linkeType)
	return res.GetError()
}

func (iplink *IPLink) SetUp() error {
	// add-br BRIDGE PARENT VLAN   create new fake BRIDGE in PARENT on VLAN
	res := iplink.exec.Execute("set", iplink.name, "up")
	return res.GetError()
}

func (iplink *IPLink) SetDown() error {
	// add-br BRIDGE PARENT VLAN   create new fake BRIDGE in PARENT on VLAN
	res := iplink.exec.Execute("set", iplink.name, "down")
	return res.GetError()
}