package vsctl

import (
	"fmt"

	"github.com/TicktW/kubefay/pkg/executor"
)

type OVSBridge struct {
	exec executor.Executor
	name string
}

func NewOVSBridge(name string) (*OVSBridge, error) {
	vsctlExec, err := executor.AgentPool.Get(VSCTLNAME)
	if err != nil {
		return nil, err
	}

	return &OVSBridge{
		name: name,
		exec: vsctlExec,
	}, nil
}

func (bridge *OVSBridge) BridgeExist() (bool, error) {
	res := bridge.exec.Execute("br-exists", bridge.name)

	// br-exists BRIDGE exit 2 if BRIDGE does not exist
	if res.GetStatus() == 0 {
		return true, nil
	}

	if res.GetStatus() == 2 {
		return false, nil
	}

	return false, res.GetError()
}

func (bridge *OVSBridge) GetBridge() {
	// _, err, state := vsctlExecutor.ReadExecute("br-exists", bridgeName)

}

func (bridge *OVSBridge) Add() error {
	// add-br BRIDGE PARENT VLAN   create new fake BRIDGE in PARENT on VLAN
	res := bridge.exec.Execute("--may-exist", "add-br", bridge.name)
	return res.GetError()
}

func (bridge *OVSBridge) Delete() error {
	// del-br BRIDGE               delete BRIDGE and all of its ports
	res := bridge.exec.Execute("del-br", bridge.name)
	return res.GetError()
}

func (bridge *OVSBridge) DelPort(portName string) error {
	// add-port BRIDGE PORT        add network device PORT to BRIDGE
	// ovs-vsctl [--if-exists] del-port br0 eth1
	res := bridge.exec.Execute("--if-exists", "del-port", bridge.name, portName)
	return res.GetError()
}

func (bridge *OVSBridge) AddNormalPort(portName string) error {
	// add-port BRIDGE PORT        add network device PORT to BRIDGE
	res := bridge.exec.Execute("add-port", bridge.name, portName)
	return res.GetError()
}

func (bridge *OVSBridge) AddInternalPort(portName string) error {
	// add-port BRIDGE PORT   # 添加Internal Port
	// ovs-vsctl add-port br-int vnet0 -- set Interface vnet0 type=internal
	// # 把网卡vnet0启动并配置IP
	// ip link set vnet0 up
	// ip addr add 192.168.0.1/24 dev vnet0     add network device PORT to BRIDGE
	res := bridge.exec.Execute("add-port", bridge.name, portName, "--", "set", "Interface", portName, "type=internal")
	return res.GetError()
}

func (bridge *OVSBridge) AddPatchPort(portName string, peerName string) error {
	// add-port BRIDGE PORT     ovs-vsctl add-br br0
	// ovs-vsctl add-br br1
	// ovs-vsctl \
	// -- add-port br0 patch0 -- set interface patch0 type=patch options:peer=patch1 \
	// -- add-port br1 patch1 -- set interface patch1 type=patch options:peer=patch0   add network device PORT to BRIDGE
	res := bridge.exec.Execute("add-port", bridge.name, portName, "--", "set", "Interface", portName, "type=patch", fmt.Sprintf("options:peer=%s", peerName))
	return res.GetError()
}

func (bridge *OVSBridge) AddTunnelPort(portName string, tunnelType string, remoteIp string) error {
	// ovs-vsctl add-port br-tun vxlan-vx01 -- set Interface vxlan-vx01 type=vxlan options:remote_ip=10.1.7.22 options:key=flow
	// ovs-vsctl add-port br-tun vxlan-vx02 -- set Interface vxlan-vx02 type=vxlan options:remote_ip=10.1.7.23 options:key=flowet interface patch1 type=patch options:peer=patch0   add network device PORT to BRIDGE
	res := bridge.exec.Execute("add-port", bridge.name, portName, "--", "set", "Interface", portName, fmt.Sprintf("type=%s", tunnelType), fmt.Sprintf("options:key=flow"))
	return res.GetError()
}
