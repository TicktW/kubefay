package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"antrea.io/antrea/pkg/signals"
	"github.com/TicktW/kubefay/pkg/agent"
	crdinformers "github.com/TicktW/kubefay/pkg/client/informers/externalversions"
	"github.com/TicktW/kubefay/pkg/cni"
	"github.com/TicktW/kubefay/pkg/utils/k8s"
	"github.com/TicktW/kubefay/pkg/version"
	"github.com/spf13/cobra"
	"k8s.io/client-go/informers"
	componentbaseconfig "k8s.io/component-base/config"
	klog "k8s.io/klog/v2"

	"github.com/TicktW/kubefay/pkg/agent/cniserver"
	"github.com/TicktW/kubefay/pkg/agent/controller/ipam"
	"github.com/TicktW/kubefay/pkg/agent/interfacestore"
	"github.com/TicktW/kubefay/pkg/agent/openflow"
	"github.com/TicktW/kubefay/pkg/agent/route"
	ofconfig "github.com/TicktW/kubefay/pkg/ovs/openflow"
	"github.com/TicktW/kubefay/pkg/ovs/ovsconfig"
)

const (
	defaultOVSBridge               = "br-int"
	defaultHostGateway             = "kubefay-gw0"
	defaultHostProcPathPrefix      = "/host"
	defaultServiceCIDR             = "10.96.0.0/12"
	defaultTunnelType              = ovsconfig.VXLANTunnel
	defaultFlowCollectorAddress    = "flow-aggregator.flow-aggregator.svc:4739:tls"
	defaultFlowCollectorTransport  = "tls"
	defaultFlowCollectorPort       = "4739"
	defaultFlowPollInterval        = 5 * time.Second
	defaultActiveFlowExportTimeout = 30 * time.Second
	defaultIdleFlowExportTimeout   = 15 * time.Second
	defaultIGMPQueryInterval       = 125 * time.Second
	defaultStaleConnectionTimeout  = 5 * time.Minute
	defaultNPLPortRange            = "61000-62000"
	defaultMTU                     = 1450
	DefaultSubnetCIDRv4            = "10.192.0.0/16"
	HostProcPathPrefix             = "/"
)

type KubefayConf struct {
	tunnelType   string
	datapathType string
	serviceCIDR  string
}

const informerDefaultResync = 12 * time.Hour

func main() {

	cmd := newAgentCommand()
	cmd.Execute()

}

func newAgentCommand() *cobra.Command {
	config := KubefayConf{}

	cmd := &cobra.Command{
		Use:   "kubefay-agent",
		Short: "kubefay-agent is the agent daemon for Kubefay CNI plugin",
		Long: `
Kubefay provide a virtual network for kubernetes. 
Kubefay design as a SDN network structure.
The agent is the control plane and Open vSwtich is the data plane.
		`,
		Run: func(cmd *cobra.Command, args []string) {
			// define floags here
			fmt.Println(config.tunnelType)
			if err := validate(&config); err != nil {
				klog.Fatalf("Error running agent: %v", err)
			}
			if err := run(&config); err != nil {
				klog.Fatalf("Error running agent: %v", err)
			}
		},
		Version: version.GetFullVersionWithRuntimeInfo(),
	}

	cmd.PersistentFlags().StringVarP(&config.tunnelType, "tunnel_type", "t", "vxlan", "set default tunnel type for Kubefay network")
	cmd.PersistentFlags().StringVarP(&config.datapathType, "datapath_type", "d", "system", "set default datapath type for OVS, system or netdev")
	cmd.PersistentFlags().StringVarP(&config.serviceCIDR, "service_cidr", "s", defaultServiceCIDR, "service CIDR")

	flags := cmd.Flags()
	klog.InitFlags(nil)
	// Install log flags
	flags.AddGoFlagSet(flag.CommandLine)
	return cmd
}

func validate(config *KubefayConf) error {

	if !ovsconfig.DatapathSets.Contains(config.datapathType) {
		return fmt.Errorf("%s erorr, should be in %v", config.datapathType, ovsconfig.DatapathSets)
	}

	if !ovsconfig.TunnelSets.Contains(config.tunnelType) {
		return fmt.Errorf("%s erorr, should be in %v", config.tunnelType, ovsconfig.TunnelSets)
	}

	if _, _, err := net.ParseCIDR(config.serviceCIDR); err != nil {
		return err
	}

	return nil
}

func run(config *KubefayConf) error {
	klog.Infof("kubefay agent (version %s)", version.GetFullVersion())
	klog.Infof("kubefay config %+v", config)
	// logs.GlogSetter("5")

	klog.V(1).Infof("kubefay agent 1 (version %s)", version.GetFullVersion())
	klog.V(2).Infof("kubefay agent 2 (version %s)", version.GetFullVersion())
	klog.V(3).Infof("kubefay agent 3 (version %s)", version.GetFullVersion())
	klog.V(4).Infof("kubefay agent 4 (version %s)", version.GetFullVersion())
	klog.V(5).Infof("kubefay agent 5 (version %s)", version.GetFullVersion())

	klog.Info("create k8s clients")
	k8sConfig := componentbaseconfig.ClientConnectionConfiguration{}
	k8sClient, crdClient, err := k8s.CreateClients(k8sConfig)

	if err != nil {
		log.Fatalf("error to create k8s client: %v", err)
	}

	informerFactory := informers.NewSharedInformerFactory(k8sClient, informerDefaultResync)
	crdInformerFactory := crdinformers.NewSharedInformerFactory(crdClient, informerDefaultResync)

	subnetInformer := crdInformerFactory.Ipam().V1alpha1().SubNets()
	namespaceInformer := informerFactory.Core().V1().Namespaces()
	podInformer := informerFactory.Core().V1().Pods()

	klog.Info("connect to ovs db...")
	ovsdbAddress := ovsconfig.GetConnAddress(ovsconfig.DefaultOVSRunDir)
	ovsdbConnection, err := ovsconfig.NewOVSDBConnectionUDS(ovsdbAddress)
	if err != nil {
		return fmt.Errorf("error connecting OVSDB: %v", err)
	}
	defer ovsdbConnection.Close()

	// ovsBridgeClient := ovsconfig.NewOVSBridge(defaultOVSBridge, ovsconfig.OVSDatapathNetdev, ovsdbConnection)
	ovsBridgeClient := ovsconfig.NewOVSBridge(defaultOVSBridge, config.datapathType, ovsdbConnection)
	ovsBridgeMgmtAddr := ofconfig.GetMgmtAddress(ovsconfig.DefaultOVSRunDir, defaultOVSBridge)

	klog.Info("makeing openflow client...")
	ofClient := openflow.NewClient(defaultOVSBridge, ovsBridgeMgmtAddr,
		false,
		false,
	)

	_, serviceCIDRNet, _ := net.ParseCIDR(config.serviceCIDR)

	klog.Info("set up route client...")
	routeClient, err := route.NewClient(serviceCIDRNet, ovsconfig.TunnelType(config.tunnelType), false)
	if err != nil {
		return fmt.Errorf("error creating route client: %v", err)
	}

	// Create an ifaceStore that caches network interfaces managed by this node.
	ifaceStore := interfacestore.NewInterfaceStore()

	agentInit := agent.NewAgent(
		k8sClient,
		ovsBridgeClient,
		ofClient,
		ifaceStore,
		defaultOVSBridge,
		defaultHostGateway,
		defaultMTU,
		config.tunnelType,
		DefaultSubnetCIDRv4,
		serviceCIDRNet,
		routeClient,
	)
	err = agentInit.Initialize()
	if err != nil {
		return err
	}

	nodeConfig := agentInit.GetNodeConfig()

	networkReadyCh := make(chan struct{})
	isChaining := false
	cniServer := cniserver.New(
		cni.CNISocketAddr,
		defaultHostProcPathPrefix,
		nodeConfig,
		k8sClient,
		isChaining,
		routeClient,
		networkReadyCh)
	// err = cniServer.Initialize(ovsBridgeClient, ofClient, ifaceStore, ovsconfig.OVSDatapathNetdev)
	err = cniServer.Initialize(ovsBridgeClient, ofClient, ifaceStore, config.datapathType)
	if err != nil {
		return fmt.Errorf("error initializing CNI server: %v", err)
	}

	ipamController := ipam.NewController(
		k8sClient,
		crdClient,
		podInformer,
		namespaceInformer,
		subnetInformer,
		ovsBridgeClient,
		ofClient,
		nodeConfig,
	)

	ipamCniServer := cniserver.NewIPAMCNIServer(
		cni.IPAMCNISocketAddr,
		nodeConfig,
		defaultHostProcPathPrefix,
		k8sClient,
		crdClient,
		podInformer,
		namespaceInformer,
		subnetInformer,
	)

	stopCh := signals.RegisterSignalHandlers()

	go cniServer.Run(stopCh)

	informerFactory.Start(stopCh)
	crdInformerFactory.Start(stopCh)

	go ipamController.Run(2, stopCh)
	go ipamCniServer.Run(stopCh)

	close(networkReadyCh)
	<-stopCh
	klog.Info("Stopping Kubefay agent")
	return nil
}
