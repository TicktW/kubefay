package vsctl

import "net"

type OVSBridge struct {
	name string
}

func NewOVSBridge(name string) *OVSBridge {
	
	return &OVSBridge{name}
}

func (bridge *OVSBridge) BridgeExist() (bool, error) {
	_, err, state := vsctlExecutor.ReadExecute("br-exists", bridge.name)

	// br-exists BRIDGE exit 2 if BRIDGE does not exist
	if state == 0 {
		return true, nil
	}

	if state == 2 {
		return false, nil
	}

	return false, err
}

func (bridge *OVSBridge) GetBridge() {
	// _, err, state := vsctlExecutor.ReadExecute("br-exists", bridgeName)

}

func (bridge *OVSBridge) AddBridge() error {
	// add-br BRIDGE PARENT VLAN   create new fake BRIDGE in PARENT on VLAN
	_, err, _ := vsctlExecutor.WriteExecute("--may-exist", "add-br", bridge.name)
	return err
}

func (bridge *OVSBridge) DeleteBridge() error {
	// del-br BRIDGE               delete BRIDGE and all of its ports
	_, err, _ := vsctlExecutor.WriteExecute("del-br", bridge.name)
	return err
}

func (bridge *OVSBridge) AddPort(portName string) error {
	// add-port BRIDGE PORT        add network device PORT to BRIDGE
	_, err, _ := vsctlExecutor.WriteExecute("add-port", bridge.name, portName)
	return err
}
