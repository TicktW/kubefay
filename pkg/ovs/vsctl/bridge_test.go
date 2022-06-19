package vsctl

import "testing"

func TestBridgeExist(t *testing.T) {
	BridgeExist("br-phy")
	BridgeExist("br")
}
