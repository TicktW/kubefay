package agent

import (
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/TicktW/kubefay/pkg/agent/types"
	"github.com/TicktW/kubefay/pkg/agent/util"
	"github.com/TicktW/kubefay/pkg/ovs/ovsconfig"
	"github.com/TicktW/kubefay/pkg/agent/openflow"
	"github.com/TicktW/kubefay/pkg/agent/openflow/cookie"
	"github.com/containernetworking/plugins/pkg/ip"
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
	ovsBridge           string
	hostGateway         string // name of gateway port on the OVS bridge
	mtu                 int
	defaultTunnelType   string
	defaultSubnetCIDRv4 string
	serviceCIDR         *net.IPNet // K8s Service ClusterIP CIDR
	// networkReadyCh      chan<- struct{}
}

func NewAgent(
	client clientset.Interface,
	ovsBridgeClient ovsconfig.OVSBridgeClient,
	ofClient openflow.Client,
	ovsBridge string,
	hostGateway string,
	mtu int,
	defaultTunnelType string,
	defaultSubnetCIDRv4 string,
	serviceCIDR *net.IPNet,
) *Agent {
	return &Agent{
		client:              client,
		ovsBridgeClient:     ovsBridgeClient,
		ofClient:            ofClient,
		ovsBridge:           ovsBridge,
		hostGateway:         hostGateway,
		mtu:                 mtu,
		defaultTunnelType:   defaultTunnelType,
		defaultSubnetCIDRv4: defaultSubnetCIDRv4,
		serviceCIDR:         serviceCIDR,
	}

}

func (agent *Agent) Initialize() error {
	klog.Info("Setting up node network")
	// wg is used to wait for the asynchronous initialization.
	var wg sync.WaitGroup

	// if err := agent.initNodeLocalConfig(); err != nil {
	// 	return err
	// }

	// if err := agent.initializeIPSec(); err != nil {
	// 	return err
	// }

	if err := agent.setupOVSBridge(); err != nil {
		return err
	}

	// Install OpenFlow entries on OVS bridge.
	if err := agent.initOpenFlowPipeline(); err != nil {
		return err
	}

	return nil
}

// setupOVSBridge sets up the OVS bridge and create host gateway interface and tunnel port
func (agent *Agent) setupOVSBridge() error {
	if err := agent.ovsBridgeClient.Create(); err != nil {
		klog.Error("Failed to create OVS bridge: ", err)
		return err
	}

	// if err := agent.prepareOVSBridge(); err != nil {
	// 	return err
	// }

	// // Initialize interface cache
	// if err := agent.initInterfaceStore(); err != nil {
	// 	return err
	// }

	if err := agent.setupDefaultTunnelInterface(defaultTunInterfaceName, ovsconfig.TunnelType(agent.defaultTunnelType)); err != nil {
		return err
	}

	// Set up host gateway interface
	err := agent.setupGatewayInterface()
	if err != nil {
		return err
	}

	return nil
}

func (agent *Agent) setupDefaultTunnelInterface(tunnelPortName string, tunnelType ovsconfig.TunnelType) error {

	// Create the default tunnel port and interface.
	_, err := agent.ovsBridgeClient.CreateTunnelPort(tunnelPortName, tunnelType, DefaultTunOFPort)
	if err != nil {
		klog.Errorf("Failed to create tunnel port %s type %s on OVS bridge: %v", tunnelPortName, tunnelType, err)
		return err
	}

	return nil
}

// setupGatewayInterface creates the host gateway interface which is an internal port on OVS. The ofport for host
// gateway interface is predefined, so invoke CreateInternalPort with a specific ofport_request
func (agent *Agent) setupGatewayInterface() error {

	// Idempotent operation to set the gateway's MTU: we perform this operation regardless of
	// whether or not the gateway interface already exists, as the desired MTU may change across
	// restarts.
	klog.V(4).Infof("Setting gateway interface %s MTU to %d", agent.hostGateway, agent.mtu)

	agent.ovsBridgeClient.SetInterfaceMTU(agent.hostGateway, agent.mtu)

	// var gwMAC net.HardwareAddr
	var gwLinkIdx int
	var err error
	// Host link might not be queried at once after creating OVS internal port; retry max 5 times with 1s
	// delay each time to ensure the link is ready.
	for retry := 0; retry < maxRetryForHostLink; retry++ {
		_, gwLinkIdx, err = util.SetLinkUp(agent.hostGateway)
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

	if err := util.ConfigureLinkAddress(gwLinkIdx, gwIP); err != nil {
		return err
	}

	return nil
}

func (agent *Agent) initOpenFlowPipeline() error {
	roundInfo := getRoundInfo(agent.ovsBridgeClient)

	// Set up all basic flows.
	ofConnCh, err := agent.ofClient.Initialize(roundInfo, agent.nodeConfig, agent.networkConfig.TrafficEncapMode)
	if err != nil {
		klog.Errorf("Failed to initialize openflow client: %v", err)
		return err
	}

	// On Windows platform, host network flows are needed for host traffic.
	if err := agent.initHostNetworkFlows(); err != nil {
		klog.Errorf("Failed to install openflow entries for host network: %v", err)
		return err
	}

	// On Windows platform, extra flows are needed to perform SNAT for the
	// traffic to external network.
	if err := agent.initExternalConnectivityFlows(); err != nil {
		klog.Errorf("Failed to install openflow entries for external connectivity: %v", err)
		return err
	}

	// Set up flow entries for gateway interface, including classifier, skip spoof guard check,
	// L3 forwarding and L2 forwarding
	if err := agent.ofClient.InstallGatewayFlows(); err != nil {
		klog.Errorf("Failed to setup openflow entries for gateway: %v", err)
		return err
	}

	if agent.networkConfig.TrafficEncapMode.SupportsEncap() {
		// Set up flow entries for the default tunnel port interface.
		if err := agent.ofClient.InstallDefaultTunnelFlows(); err != nil {
			klog.Errorf("Failed to setup openflow entries for tunnel interface: %v", err)
			return err
		}
	}

	if !agent.enableProxy {
		// Set up flow entries to enable Service connectivity. Upstream kube-proxy is leveraged to
		// provide load-balancing, and the flows installed by this method ensure that traffic sent
		// from local Pods to any Service address can be forwarded to the host gateway interface
		// correctly. Otherwise packets might be dropped by egress rules before they are DNATed to
		// backend Pods.
		if err := agent.ofClient.InstallClusterServiceCIDRFlows([]*net.IPNet{agent.serviceCIDR, agent.serviceCIDRv6}); err != nil {
			klog.Errorf("Failed to setup OpenFlow entries for Service CIDRs: %v", err)
			return err
		}
	} else {
		// Set up flow entries to enable Service connectivity. The agent proxy handles
		// ClusterIP Services while the upstream kube-proxy is leveraged to handle
		// any other kinds of Services.
		if err := agent.ofClient.InstallClusterServiceFlows(); err != nil {
			klog.Errorf("Failed to setup default OpenFlow entries for ClusterIP Services: %v", err)
			return err
		}
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