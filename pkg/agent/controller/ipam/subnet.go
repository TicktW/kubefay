package ipam

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/TicktW/kubefay/pkg/agent/config"
	"github.com/TicktW/kubefay/pkg/agent/interfacestore"
	"github.com/TicktW/kubefay/pkg/agent/util"
	"github.com/containernetworking/plugins/pkg/ip"
	"github.com/vishvananda/netlink"
	corev1 "k8s.io/api/core/v1"

	"github.com/TicktW/kubefay/pkg/apis/ipam/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/TicktW/kubefay/pkg/utils/env"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

const (
	antreaPodIPSet         = "ANTREA-POD-IP"
	antreaPostRoutingChain = "ANTREA-POSTROUTING"
)

// addSubnetHandler add a interface with routables, iptables rules
func addSubnetHandler(c *Controller, key string) error {
	klog.Info("add subnet handler begin:", key)
	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// for default subnet, only need to install l3 fwd to gateway flow
	if name == DefaultNet {
		klog.Infof("install gateway flows for: %s", c.nodeConfig.GatewayConfig.IPv4)
		return c.ofClient.InstallIPAMGatewayFlows(c.nodeConfig.GatewayConfig.IPv4, c.nodeConfig.NodeIPAddr.IP)
	}

	// get subnet
	subnet, err := c.subnetLister.SubNets(DefaultNamespace).Get(name)
	if err != nil {
		klog.Errorf("get subnet error, %v", err)
		return err
	}

	// check link and configure ip and routes
	// if link does not exist, setup net link, it does not affect now.
	// subnet | gateway interface
	// defaultnet | antreagw-0
	// Note for ovs port, len(name) <= 14
	gwName := c.nodeConfig.GatewayConfig.Name
	fmt.Println(c.ifaceStore)
	gatewayIface, portExists := c.ifaceStore.GetInterface(gwName)
	if !portExists {
		klog.Infof("Creating gateway port %s on OVS bridge", gwName)
		gwPortUUID, err := c.ovsBridgeClient.CreateInternalPort(gwName, 0, nil)
		if err != nil {
			klog.Errorf("Failed to create gateway port %s on OVS bridge: %v", gwName, err)
			return err
		}
		gatewayIface = interfacestore.NewGatewayInterface(gwName)
		gatewayIface.OVSPortConfig = &interfacestore.OVSPortConfig{PortUUID: gwPortUUID, OFPort: config.HostGatewayOFPort}
		c.ifaceStore.AddInterface(gatewayIface)
	} else {
		klog.Infof("Gateway port %s already exists on OVS bridge", gwName)
	}

	if err := c.configureGatewayInterface(gatewayIface, gwName, subnet); err != nil {
		return err
	}

	// iptables-restore iptable rules
	_, podNet, err := net.ParseCIDR(subnet.Spec.CIDR)
	if err != nil {
		klog.Errorf("parse gateway error, %s is not a valid CIDR", subnet.Spec.CIDR)
		return err
	}

	// add nat rules for subnet
	err = c.ipt.AppendNat(antreaPostRoutingChain, podNet.String())
	if err != nil {
		return err
	}

	// install gateway flows here
	klog.Infof("install gateway flows for: %s", gatewayIface.GetIPv4Addr())
	return c.ofClient.InstallIPAMGatewayFlows(gatewayIface.GetIPv4Addr(), c.nodeConfig.NodeIPAddr.IP)
}

func (c *Controller) configureGatewayInterface(gatewayIface *interfacestore.InterfaceConfig, gwName string, subnet *v1alpha1.SubNet) error {
	var gwMAC net.HardwareAddr
	var gwLinkIdx int
	var err error
	// Host link might not be queried at once after creating OVS internal port; retry max 5 times with 1s
	// delay each time to ensure the link is ready.
	for retry := 0; retry < 10; retry++ {
		gwMAC, gwLinkIdx, err = util.SetLinkUp(gwName)
		if err == nil {
			break
		}
		if _, ok := err.(util.LinkNotFound); ok {
			klog.V(2).Infof("Not found host link for gateway %s, retry after 1s", gwName)
			time.Sleep(1 * time.Second)
			continue
		} else {
			return err
		}
	}

	if err != nil {
		klog.Errorf("Failed to find host link for gateway %s: %v", gwName, err)
		return err
	}
	gatewayIface.MAC = gwMAC

	// Allocate the gateway IP address from the Pod CIDRs if it exists. The gateway IP should be the first address
	// in the Subnet and configure on the host gateway.
	_, gatewayNet, err := net.ParseCIDR(subnet.Spec.CIDR)
	if err != nil {
		klog.Errorf("parse gateway error, %s is not a valid CIDR", subnet.Spec.CIDR)
	}

	for _, podCIDR := range []*net.IPNet{gatewayNet} {
		if err := c.allocateGatewayAddress(podCIDR, gatewayIface, gwLinkIdx); err != nil {
			return err
		}
	}
	return nil
}

func (c *Controller) allocateGatewayAddress(localSubnet *net.IPNet, gatewayIface *interfacestore.InterfaceConfig, gwLinkIdx int) error {
	if localSubnet == nil {
		return nil
	}
	subnetID := localSubnet.IP.Mask(localSubnet.Mask)
	gwIP := &net.IPNet{IP: ip.NextIP(subnetID), Mask: localSubnet.Mask}

	// Check IP address configuration on existing interface first, return if the interface has the desired address.
	// We perform this check unconditionally, even if the OVS port does not exist when this function is called
	// (i.e. portExists is false). Indeed, it may be possible for the interface to exist even if the OVS bridge does
	// not exist.
	// Configure the IP address on the interface if it does not exist.
	if err := util.ConfigureLinkAddress(gwLinkIdx, gwIP); err != nil {
		return err
	}

	// add route for the gateway ip
	route := &netlink.Route{
		Dst:       localSubnet,
		LinkIndex: gwLinkIdx,
	}
	if err := netlink.RouteReplace(route); err != nil {
		return err
	}

	gatewayIface.IPs = append(gatewayIface.IPs, gwIP.IP)
	return nil
}

// updateSubnetHandler travels the pod's subnet, add of flow for every pod
func updateSubnetHandler(c *Controller, key string) error {
	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return err
	}
	// get subnet
	subnet, err := c.subnetLister.SubNets(DefaultNamespace).Get(name)

	if subnet.Status.IPAMEvent != "POD_ADD" {
		return nil
	}

	if err != nil {
		klog.Errorf("get subnet error: %v", err)
		return err
	}

	// add pod's flow for every new pod
	newIPInfo := subnet.Spec.LastReservedIP
	klog.Infof("update subnet: %v, ---- %s", subnet, newIPInfo)
	if newIPInfo == "" {
		klog.Info("can't find last ip for subnet")
		return nil
	}
	podInfoSlice := strings.Split(newIPInfo, "/")
	// peerNodeName, podNamespace, podName, containerID, ifName, peerPodIP := podInfoSlice[0], podInfoSlice[1], podInfoSlice[2], podInfoSlice[3], podInfoSlice[4], podInfoSlice[5]
	// check link
	peerNodeName, _, _, _, _, peerPodIP := podInfoSlice[0], podInfoSlice[1], podInfoSlice[2], podInfoSlice[3], podInfoSlice[4], podInfoSlice[5]

	// if link does not exist, setup net link
	// subnet | gateway interface
	// default | antreagw-default
	localNodeName, err := env.GetNodeName()
	if err != nil {
		klog.Error(err)
		return err
	}

	klog.Infof("local node %s, peer node %s", localNodeName, peerNodeName)
	// local Node doesn't need to set up
	if peerNodeName == localNodeName {
		klog.Infof("local node %s == peer node %s", localNodeName, peerNodeName)
		return nil
	}

	peerNode, err := c.kubeclientset.CoreV1().Nodes().Get(context.TODO(), peerNodeName, metav1.GetOptions{})
	if err != nil {
		klog.Errorf("Get k8s Node: %s error", peerNodeName)
		return err
	}
	peerNodeIP, err := GetNodeAddr(peerNode)
	if err != nil {
		klog.Errorf("Get k8s Node IP: %s error", peerNodeName)
		return err
	}
	podInfo := newIPInfo

	// TODO: add support for ipsec tunnel

	klog.Info("install flows for peer node pod")
	err = c.ofClient.InstallRemotePodFlows(podInfo, net.ParseIP(peerPodIP), peerNodeIP, uint32(0))

	if err != nil {
		return fmt.Errorf("failed to install flows to Peer Pod %s: %v", peerPodIP, err)
	}

	return nil
}

// delSubnetHandler remove the gateway interface
// TODO: remove ip, routs and flows for subnet
func delSubnetHandler(c *Controller, key string) error {
	// check exists
	// remove interface
	// remove store
	klog.Info("del subnet:", key)
	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}
	// get subnet
	subnets, err := c.subnetLister.List(labels.Everything())

	if err != nil {
		klog.Error("list subnet error")
	}

	// delete subnet's gateway on every Node
	for _, subnet := range subnets {

		if subnet.Name != name {
			continue
		}

		// check link
		// if link does not exist, setup net link
		// subnet | gateway interface
		// default | antreagw-default
		gwName := "antreagw-" + name
		gwName = name
		gatewayIface, portExists := c.ifaceStore.GetInterface(gwName)
		if !portExists {
			klog.V(2).Infof("Gateway port %s does not exist on OVS bridge", gwName)
		} else {
			klog.V(2).Infof("Del gateway port %s on OVS bridge", gwName)

			// del ovs port
			err := c.ovsBridgeClient.DeletePort(gwName)
			if err != nil {
				klog.Errorf("Failed to del gateway port %s on OVS bridge: %v", gwName, err)
				return err
			}

			// del interface
			c.ifaceStore.DeleteInterface(gatewayIface)
		}
	}
	return nil
}

func GetNodeAddr(node *corev1.Node) (net.IP, error) {
	addresses := make(map[corev1.NodeAddressType]string)
	for _, addr := range node.Status.Addresses {
		addresses[addr.Type] = addr.Address
	}
	var ipAddrStr string
	if internalIP, ok := addresses[corev1.NodeInternalIP]; ok {
		ipAddrStr = internalIP
	} else if externalIP, ok := addresses[corev1.NodeExternalIP]; ok {
		ipAddrStr = externalIP
	} else {
		return nil, fmt.Errorf("node %s has neither external ip nor internal ip", node.Name)
	}
	ipAddr := net.ParseIP(ipAddrStr)
	if ipAddr == nil {
		return nil, fmt.Errorf("<%v> is not a valid ip address", ipAddrStr)
	}
	return ipAddr, nil
}
