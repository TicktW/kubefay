package agent

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/TicktW/kubefay/pkg/agent/cniserver"
	"github.com/TicktW/kubefay/pkg/agent/config"
	"github.com/TicktW/kubefay/pkg/agent/controller/noderoute"
	"github.com/TicktW/kubefay/pkg/agent/interfacestore"
	"github.com/TicktW/kubefay/pkg/agent/openflow"
	"github.com/TicktW/kubefay/pkg/agent/openflow/cookie"
	"github.com/TicktW/kubefay/pkg/agent/route"
	"github.com/TicktW/kubefay/pkg/agent/types"
	"github.com/TicktW/kubefay/pkg/agent/util"
	"github.com/TicktW/kubefay/pkg/ovs/ovsconfig"
	"github.com/TicktW/kubefay/pkg/utils/env"
	"github.com/containernetworking/plugins/pkg/ip"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	klog "k8s.io/klog/v2"
)

const (
	// Default name of the default tunnel interface on the OVS bridge.
	defaultTunInterfaceName = "antrea-tun0"
	maxRetryForHostLink     = 5
	// ipsecPSKEnvKey is environment variable.
	ipsecPSKEnvKey          = "ANTREA_IPSEC_PSK"
	roundNumKey             = "roundNum" // round number key in externalIDs.
	initialRoundNum         = 1
	maxRetryForRoundNumSave = 5
)

const (
	// Invalid ofport_request number is in range 1 to 65,279. For ofport_request number not in the range, OVS
	// ignore the it and automatically assign a port number.
	// Here we use an invalid port number "0" to request for automatically port allocation.
	AutoAssignedOFPort = 0
	DefaultTunOFPort   = 1
	HostGatewayOFPort  = 2
	UplinkOFPort       = 3
	// 0xfffffffe is a reserved port number in OpenFlow protocol, which is dedicated for the Bridge interface.
	BridgeOFPort = 0xfffffffe
)

type Agent struct {
	client              clientset.Interface
	ovsBridgeClient     ovsconfig.OVSBridgeClient
	ofClient            openflow.Client
	ifaceStore          interfacestore.InterfaceStore
	ovsBridge           string
	hostGateway         string // name of gateway port on the OVS bridge
	mtu                 int
	tunnelType          string
	defaultSubnetCIDRv4 string
	serviceCIDR         *net.IPNet // K8s Service ClusterIP CIDR
	nodeConfig          *config.NodeConfig
	routeClient         route.Interface
	// networkReadyCh      chan<- struct{}
}

func NewAgent(
	client clientset.Interface,
	ovsBridgeClient ovsconfig.OVSBridgeClient,
	ofClient openflow.Client,
	ifaceStore interfacestore.InterfaceStore,
	ovsBridge string,
	hostGateway string,
	mtu int,
	tunnelType string,
	defaultSubnetCIDRv4 string,
	serviceCIDR *net.IPNet,
	routeClient route.Interface,
) *Agent {
	return &Agent{
		client:              client,
		ovsBridgeClient:     ovsBridgeClient,
		ofClient:            ofClient,
		ovsBridge:           ovsBridge,
		ifaceStore:          ifaceStore,
		hostGateway:         hostGateway,
		mtu:                 mtu,
		tunnelType:          tunnelType,
		defaultSubnetCIDRv4: defaultSubnetCIDRv4,
		serviceCIDR:         serviceCIDR,
		routeClient:         routeClient,
	}

}

func (agent *Agent) Initialize() error {
	klog.Info("Setting up node network")
	// wg is used to wait for the asynchronous initialization.
	var wg sync.WaitGroup

	klog.Info("agent init: init node local config")
	if err := agent.initNodeLocalConfig(); err != nil {
		return err
	}

	// if err := agent.initializeIPSec(); err != nil {
	// 	return err
	// }

	klog.Info("agent init: set up ovs bridge")
	if err := agent.setupOVSBridge(); err != nil {
		return err
	}

	wg.Add(1)
	// routeClient.Initialize() should be after i.setupOVSBridge() which
	// creates the host gateway interface.
	klog.Info("agent init: route client init")
	if err := agent.routeClient.Initialize(agent.nodeConfig, wg.Done); err != nil {
		return err
	}

	// time.Sleep(10000 * time.Second)
	// Install OpenFlow entries on OVS bridge.
	klog.Info("agent init: init openflow pipeline")
	if err := agent.initOpenFlowPipeline(); err != nil {
		return err
	}

	return nil
}

// initInterfaceStore initializes InterfaceStore with all OVS ports retrieved
// from the OVS bridge.
func (agent *Agent) initInterfaceStore() error {
	ovsPorts, err := agent.ovsBridgeClient.GetPortList()
	klog.Infof("ovsports %+v", ovsPorts)
	if err != nil {
		klog.Errorf("Failed to list OVS ports: %v", err)
		return err
	}

	ifaceList := make([]*interfacestore.InterfaceConfig, 0, len(ovsPorts))
	for index := range ovsPorts {
		port := &ovsPorts[index]
		ovsPort := &interfacestore.OVSPortConfig{
			PortUUID: port.UUID,
			OFPort:   port.OFPort}
		var intf *interfacestore.InterfaceConfig

		switch {
		case port.OFPort == config.HostGatewayOFPort:
			intf = &interfacestore.InterfaceConfig{
				Type:          interfacestore.GatewayInterface,
				InterfaceName: port.Name,
				OVSPortConfig: ovsPort}
			if intf.InterfaceName != agent.hostGateway {
				klog.Warningf("The discovered gateway interface name %s is different from the configured value: %s",
					intf.InterfaceName, agent.hostGateway)
				// Set the gateway interface name to the discovered name.
				agent.hostGateway = intf.InterfaceName
			}
		case port.IFType == ovsconfig.GeneveTunnel:
			fallthrough
		case port.IFType == ovsconfig.VXLANTunnel:
			fallthrough
		case port.IFType == ovsconfig.GRETunnel:
			fallthrough
		case port.IFType == ovsconfig.STTTunnel:
			intf = noderoute.ParseTunnelInterfaceConfig(port, ovsPort)
			if intf != nil && port.OFPort == config.DefaultTunOFPort &&
				intf.InterfaceName != agent.nodeConfig.DefaultTunName {
				klog.Infof("The discovered default tunnel interface name %s is different from the default value: %s",
					intf.InterfaceName, agent.nodeConfig.DefaultTunName)
				// Set the default tunnel interface name to the discovered name.
				agent.nodeConfig.DefaultTunName = intf.InterfaceName
			}
		default:
			// The port should be for a container interface.
			intf = cniserver.ParseOVSPortInterfaceConfig(port, ovsPort)
		}
		if intf != nil {
			ifaceList = append(ifaceList, intf)
		}
	}

	agent.ifaceStore.Initialize(ifaceList)
	return nil
}

// setupOVSBridge sets up the OVS bridge and create host gateway interface and tunnel port
func (agent *Agent) setupOVSBridge() error {

	klog.Info("agent init: setup ovs bridge: create bridge")
	if err := agent.ovsBridgeClient.Create(); err != nil {
		klog.Error("Failed to create OVS bridge: ", err)
		return err
	}

	klog.Info("agent init: setup ovs bridge: init interface store")
	// Initialize interface cache
	if err := agent.initInterfaceStore(); err != nil {
		return err
	}

	klog.Info("agent init: setup ovs bridge: setup tunnel interface")
	if err := agent.setupDefaultTunnelInterface(defaultTunInterfaceName, ovsconfig.TunnelType(agent.tunnelType)); err != nil {
		return err
	}

	// Set up host gateway interface
	klog.Info("agent init: setup ovs bridge: setup gateway interface")
	err := agent.setupGatewayInterface()
	if err != nil {
		return err
	}

	return nil
}

func (agent *Agent) setupDefaultTunnelInterface(tunnelPortName string, tunnelType ovsconfig.TunnelType) error {

	tunnelIface, portExists := agent.ifaceStore.GetInterface(tunnelPortName)

	// Enabling UDP checksum can greatly improve the performance for Geneve and
	// VXLAN tunnels by triggering GRO on the receiver.
	shouldEnableCsum := tunnelType == ovsconfig.GeneveTunnel || tunnelType == ovsconfig.VXLANTunnel

	// Check the default tunnel port.
	if portExists {
		if tunnelIface.TunnelInterfaceConfig.Type == tunnelType {
			klog.V(2).Infof("Tunnel port %s already exists on OVS bridge", tunnelPortName)
			// This could happen when upgrading from previous versions that didn't set it.
			if shouldEnableCsum && !tunnelIface.TunnelInterfaceConfig.Csum {
				if err := agent.enableTunnelCsum(tunnelPortName); err != nil {
					return fmt.Errorf("failed to enable csum for tunnel port %s: %v", tunnelPortName, err)
				}
				tunnelIface.TunnelInterfaceConfig.Csum = true
			}
			return nil
		}
	}

	// Create the default tunnel port and interface.
	_, err := agent.ovsBridgeClient.CreateTunnelPort(tunnelPortName, tunnelType, DefaultTunOFPort)
	if err != nil {
		klog.Errorf("Failed to create tunnel port %s type %s on OVS bridge: %v", tunnelPortName, tunnelType, err)
		return err
	}

	return nil
}

func (agent *Agent) enableTunnelCsum(tunnelPortName string) error {
	options, err := agent.ovsBridgeClient.GetInterfaceOptions(tunnelPortName)
	if err != nil {
		return fmt.Errorf("error getting interface options: %w", err)
	}

	updatedOptions := make(map[string]interface{})
	for k, v := range options {
		updatedOptions[k] = v
	}
	updatedOptions["csum"] = "true"
	return agent.ovsBridgeClient.SetInterfaceOptions(tunnelPortName, updatedOptions)
}

// setupGatewayInterface creates the host gateway interface which is an internal port on OVS. The ofport for host
// gateway interface is predefined, so invoke CreateInternalPort with a specific ofport_request
func (agent *Agent) setupGatewayInterface() error {
	gatewayIface, portExists := agent.ifaceStore.GetInterface(agent.hostGateway)
	klog.Infof("%s", agent.hostGateway)
	klog.Infof("%+v", portExists)
	klog.Infof("%+v", gatewayIface)
	if !portExists {
		klog.V(2).Infof("Creating gateway port %s on OVS bridge", agent.hostGateway)
		gwPortUUID, err := agent.ovsBridgeClient.CreateInternalPort(agent.hostGateway, config.HostGatewayOFPort, nil)
		if err != nil {
			klog.Errorf("Failed to create gateway port %s on OVS bridge: %v", agent.hostGateway, err)
			return err
		}
		gatewayIface = interfacestore.NewGatewayInterface(agent.hostGateway)
		gatewayIface.OVSPortConfig = &interfacestore.OVSPortConfig{PortUUID: gwPortUUID, OFPort: config.HostGatewayOFPort}
		agent.ifaceStore.AddInterface(gatewayIface)
	} else {
		klog.V(2).Infof("Gateway port %s already exists on OVS bridge", agent.hostGateway)
	}

	// Idempotent operation to set the gateway's MTU: we perform this operation regardless of
	// whether or not the gateway interface already exists, as the desired MTU may change across
	// restarts.
	klog.V(4).Infof("Setting gateway interface %s MTU to %d", agent.hostGateway, agent.mtu)

	agent.ovsBridgeClient.SetInterfaceMTU(agent.hostGateway, agent.mtu)

	var gwMAC net.HardwareAddr
	var gwLinkIdx int
	var err error
	// Host link might not be queried at once after creating OVS internal port; retry max 5 times with 1s
	// delay each time to ensure the link is ready.
	for retry := 0; retry < maxRetryForHostLink; retry++ {
		gwMAC, gwLinkIdx, err = util.SetLinkUp(agent.hostGateway)
		if err == nil {
			break
		}
		if _, ok := err.(util.LinkNotFound); ok {
			klog.V(2).Infof("Not found host link for gateway %s, retry after 1s", agent.hostGateway)
			time.Sleep(1 * time.Second)
			continue
		} else {
			return err
		}
	}

	if err != nil {
		klog.Errorf("Failed to find host link for gateway %s: %v", agent.hostGateway, err)
		return err
	}

	_, defaultNet, parseErr := net.ParseCIDR(agent.defaultSubnetCIDRv4)

	if parseErr != nil {
		klog.Errorln(parseErr)
		return parseErr
	}
	subnetID := defaultNet.IP.Mask(defaultNet.Mask)
	gwIP := &net.IPNet{IP: ip.NextIP(subnetID), Mask: defaultNet.Mask}

	agent.nodeConfig.GatewayConfig = &config.GatewayConfig{Name: agent.hostGateway, MAC: gwMAC, LinkIndex: gwLinkIdx, IPv4: gwIP.IP}
	if err := util.ConfigureLinkAddress(gwLinkIdx, gwIP); err != nil {
		return err
	}

	return nil
}

func (agent *Agent) initOpenFlowPipeline() error {
	roundInfo := getRoundInfo(agent.ovsBridgeClient)

	// Set up all basic flows.
	ofConnCh, err := agent.ofClient.Initialize(roundInfo, agent.nodeConfig)
	if err != nil {
		klog.Errorf("Failed to initialize openflow client: %v", err)
		return err
	}

	// Set up flow entries for gateway interface, including classifier, skip spoof guard check,
	// L3 forwarding and L2 forwarding
	klog.Info("gateway out ......")
	if err := agent.ofClient.InstallGatewayFlows(); err != nil {
		klog.Errorf("Failed to setup openflow entries for gateway: %v", err)
		return err
	}

	// Set up flow entries for the default tunnel port interface.
	if err := agent.ofClient.InstallDefaultTunnelFlows(); err != nil {
		klog.Errorf("Failed to setup openflow entries for tunnel interface: %v", err)
		return err
	}

	// if !agent.enableProxy {
	// 	// Set up flow entries to enable Service connectivity. Upstream kube-proxy is leveraged to
	// 	// provide load-balancing, and the flows installed by this method ensure that traffic sent
	// 	// from local Pods to any Service address can be forwarded to the host gateway interface
	// 	// correctly. Otherwise packets might be dropped by egress rules before they are DNATed to
	// 	// backend Pods.
	// 	if err := agent.ofClient.InstallClusterServiceCIDRFlows([]*net.IPNet{agent.serviceCIDR, agent.serviceCIDRv6}); err != nil {
	// 		klog.Errorf("Failed to setup OpenFlow entries for Service CIDRs: %v", err)
	// 		return err
	// 	}
	// } else {

	// Set up flow entries to enable Service connectivity. Upstream kube-proxy is leveraged to
	// provide load-balancing, and the flows installed by this method ensure that traffic sent
	// from local Pods to any Service address can be forwarded to the host gateway interface
	// correctly. Otherwise packets might be dropped by egress rules before they are DNATed to
	// backend Pods.
	// TODO add service cidr v6, 3rd para of this func
	if err := agent.ofClient.InstallClusterServiceCIDRFlows([]*net.IPNet{agent.serviceCIDR, agent.serviceCIDR}); err != nil {
		klog.Errorf("Failed to setup OpenFlow entries for Service CIDRs: %v", err)
		return err
	}

	go func() {
		// Delete stale flows from previous round. We need to wait long enough to ensure
		// that all the flow which are still required have received an updated cookie (with
		// the new round number), otherwise we would disrupt the dataplane. Unfortunately,
		// the time required for convergence may be large and there is no simple way to
		// determine when is a right time to perform the cleanup task.
		// TODO: introduce a deterministic mechanism through which the different entities
		//  responsible for installing flows can notify the agent that this deletion
		//  operation can take place. A waitGroup can be created here and notified when
		//  full sync in agent networkpolicy controller is complete. This would signal NP
		//  flows have been synced once. Other mechanisms are still needed for node flows
		//  fullSync check.
		time.Sleep(10 * time.Second)
		klog.Info("Deleting stale flows from previous round if any")
		if err := agent.ofClient.DeleteStaleFlows(); err != nil {
			klog.Errorf("Error when deleting stale flows from previous round: %v", err)
			return
		}
		persistRoundNum(roundInfo.RoundNum, agent.ovsBridgeClient, 1*time.Second, maxRetryForRoundNumSave)
	}()

	go func() {
		for {
			if _, ok := <-ofConnCh; !ok {
				return
			}
			klog.Info("Replaying OF flows to OVS bridge")
			agent.ofClient.ReplayFlows()
			klog.Info("Flow replay completed")

			// ofClient and ovsBridgeClient have their own mechanisms to restore connections with OVS, and it could
			// happen that ovsBridgeClient's connection is not ready when ofClient completes flow replay. We retry it
			// with a timeout that is longer time than ovsBridgeClient's maximum connecting retry interval (8 seconds)
			// to ensure the flag can be removed successfully.
			err := wait.PollImmediate(200*time.Millisecond, 10*time.Second, func() (done bool, err error) {
				if err := agent.FlowRestoreComplete(); err != nil {
					return false, nil
				}
				return true, nil
			})
			// This shouldn't happen unless OVS is disconnected again after replaying flows. If it happens, we will try
			// to clean up the config again so an error log should be fine.
			if err != nil {
				klog.Errorf("Failed to clean up flow-restore-wait config: %v", err)
			}
		}
	}()

	return nil
}

func (agent *Agent) FlowRestoreComplete() error {
	// ovs-vswitchd is started with flow-restore-wait set to true for the following reasons:
	// 1. It prevents packets from being mishandled by ovs-vswitchd in its default fashion,
	//    which could affect existing connections' conntrack state and cause issues like #625.
	// 2. It prevents ovs-vswitchd from flushing or expiring previously set datapath flows,
	//    so existing connections can achieve 0 downtime during OVS restart.
	// As a result, we remove the config here after restoring necessary flows.
	klog.Info("Cleaning up flow-restore-wait config")
	if err := agent.ovsBridgeClient.DeleteOVSOtherConfig(map[string]interface{}{"flow-restore-wait": "true"}); err != nil {
		return fmt.Errorf("error when cleaning up flow-restore-wait config: %v", err)
	}
	klog.Info("Cleaned up flow-restore-wait config")
	return nil
}
func getRoundInfo(bridgeClient ovsconfig.OVSBridgeClient) types.RoundInfo {
	roundInfo := types.RoundInfo{}
	num, err := getLastRoundNum(bridgeClient)
	if err != nil {
		klog.Infof("No round number found in OVSDB, using %v", initialRoundNum)
		// We use a fixed value instead of a randomly-generated value to ensure that stale
		// flows can be properly deleted in case of multiple rapid restarts when the agent
		// is first deployed to a Node.
		num = initialRoundNum
	} else {
		roundInfo.PrevRoundNum = new(uint64)
		*roundInfo.PrevRoundNum = num
		num++
	}

	num %= 1 << cookie.BitwidthRound
	klog.Infof("Using round number %d", num)
	roundInfo.RoundNum = num

	return roundInfo
}

func getLastRoundNum(bridgeClient ovsconfig.OVSBridgeClient) (uint64, error) {
	extIDs, ovsCfgErr := bridgeClient.GetExternalIDs()
	if ovsCfgErr != nil {
		return 0, fmt.Errorf("error getting external IDs: %w", ovsCfgErr)
	}
	roundNumValue, exists := extIDs[roundNumKey]
	if !exists {
		return 0, fmt.Errorf("no round number found in OVSDB")
	}
	num, err := strconv.ParseUint(roundNumValue, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("error parsing last round number %v: %w", num, err)
	}
	return num, nil
}

// initNodeLocalConfig retrieves node's subnet CIDR from node.spec.PodCIDR, which is used for IPAM and setup
// host gateway interface.
func (agent *Agent) initNodeLocalConfig() error {
	nodeName, err := env.GetNodeName()
	if err != nil {
		return err
	}
	node, err := agent.client.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
	if err != nil {
		klog.Errorf("Failed to get node from K8s with name %s: %v", nodeName, err)
		return err
	}

	ipAddr, err := noderoute.GetNodeAddr(node)
	if err != nil {
		return fmt.Errorf("failed to obtain local IP address from k8s: %w", err)
	}
	localAddr, _, err := util.GetIPNetDeviceFromIP(ipAddr)
	if err != nil {
		return fmt.Errorf("failed to get local IPNet:  %v", err)
	}

	agent.nodeConfig = &config.NodeConfig{
		Name:           nodeName,
		OVSBridge:      agent.ovsBridge,
		DefaultTunName: defaultTunInterfaceName,
		NodeIPAddr:     localAddr,
		NodeMTU:        agent.mtu,
	}

	klog.Infof("Setting Node MTU=%d", agent.mtu)

	// TODO: should use more elegant method
	// need to set default subnet
	// TODO default subnet from CRD/conf
	// var localSubnet *net.IPNet
	klog.Infof("Antrea IPAM enabled, CIDR: %s", agent.defaultSubnetCIDRv4)
	_, localSubnet, err := net.ParseCIDR(agent.defaultSubnetCIDRv4)
	if err != nil {
		return err
	}
	agent.nodeConfig.PodIPv4CIDR = localSubnet
	klog.Infof("%+v", agent.nodeConfig)
	return nil

}

// persistRoundNum will save the provided round number to OVSDB as an external ID. To account for
// transient failures, this (synchronous) function includes a retry mechanism.
func persistRoundNum(num uint64, bridgeClient ovsconfig.OVSBridgeClient, interval time.Duration, maxRetries int) {
	klog.Infof("Persisting round number %d to OVSDB", num)
	retry := 0
	for {
		err := saveRoundNum(num, bridgeClient)
		if err == nil {
			klog.Infof("Round number %d was persisted to OVSDB", num)
			return // success
		}
		klog.Errorf("Error when writing round number to OVSDB: %v", err)
		if retry >= maxRetries {
			break
		}
		time.Sleep(interval)
	}
	klog.Errorf("Unable to persist round number %d to OVSDB after %d tries", num, maxRetries+1)
}

func saveRoundNum(num uint64, bridgeClient ovsconfig.OVSBridgeClient) error {
	extIDs, ovsCfgErr := bridgeClient.GetExternalIDs()
	if ovsCfgErr != nil {
		return fmt.Errorf("error getting external IDs: %w", ovsCfgErr)
	}
	updatedExtIDs := make(map[string]interface{})
	for k, v := range extIDs {
		updatedExtIDs[k] = v
	}
	updatedExtIDs[roundNumKey] = fmt.Sprint(num)
	return bridgeClient.SetExternalIDs(updatedExtIDs)
}

func (agent *Agent) GetNodeConfig() *config.NodeConfig {
	return agent.nodeConfig
}
