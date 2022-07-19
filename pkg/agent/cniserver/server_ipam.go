package cniserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"

	"github.com/TicktW/kubefay/pkg/agent/cniserver/ipam"
	"github.com/TicktW/kubefay/pkg/agent/cniserver/ipam/allocator"
	"github.com/TicktW/kubefay/pkg/agent/cniserver/ipam/allocator/backend/crdstore"
	"github.com/TicktW/kubefay/pkg/agent/config"
	"github.com/TicktW/kubefay/pkg/agent/util"
	cnitypes "github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/containernetworking/cni/pkg/version"

	IPAMDNS "github.com/TicktW/kubefay/pkg/agent/cniserver/ipam/allocator/dns"
	cnipb "github.com/TicktW/kubefay/pkg/rpc/cni/v1beta1"
	crdclients "github.com/TicktW/kubefay/pkg/client/clientset/versioned"
	crdinformers "github.com/TicktW/kubefay/pkg/client/informers/externalversions/ipam/v1alpha1"
	crdlister "github.com/TicktW/kubefay/pkg/client/listers/ipam/v1alpha1"
	"github.com/TicktW/kubefay/pkg/cni"
	"google.golang.org/grpc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	coreinformers "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

type IPAMCNIServer struct {
	cniSocket            string
	supportedCNIVersions map[string]bool
	serverVersion        string
	nodeConfig           *config.NodeConfig
	hostProcPathPrefix   string
	kubeClient           clientset.Interface
	crdClient            crdclients.Interface
	containerAccess      *containerAccessArbitrator
	// podlisters and informer
	podLister corelisters.PodLister
	podSynced cache.InformerSynced

	nsLister corelisters.NamespaceLister
	nsSynced cache.InformerSynced

	subnetLister crdlister.SubNetLister
	subnetSynced cache.InformerSynced
}

// NewIPAMCNIServer return a new rpc IPAMCNIServer
func NewIPAMCNIServer(
	cniSocket string,
	nodeConfig *config.NodeConfig,
	hostProcPathPrefix string,
	kubeClient clientset.Interface,
	crdClient crdclients.Interface,
	podInformer coreinformers.PodInformer,
	nsInformer coreinformers.NamespaceInformer,
	subnetInformer crdinformers.SubNetInformer,
) *IPAMCNIServer {
	return &IPAMCNIServer{
		cniSocket:            cniSocket,
		supportedCNIVersions: supportedCNIVersionSet,
		serverVersion:        cni.AntreaCNIVersion,
		nodeConfig:           nodeConfig,
		hostProcPathPrefix:   hostProcPathPrefix,
		kubeClient:           kubeClient,
		crdClient:            crdClient,
		containerAccess:      newContainerAccessArbitrator(),
		podLister:            podInformer.Lister(),
		podSynced:            podInformer.Informer().HasSynced,
		nsLister:             nsInformer.Lister(),
		nsSynced:             nsInformer.Informer().HasSynced,
		subnetLister:         subnetInformer.Lister(),
		subnetSynced:         subnetInformer.Informer().HasSynced,
	}
}

func (s *IPAMCNIServer) updateLocalIPAMSubnet(cniConfig *CNIConfig) {
	if (s.nodeConfig.GatewayConfig.IPv4 != nil) && (s.nodeConfig.PodIPv4CIDR != nil) {
		cniConfig.NetworkConfig.IPAM.Ranges = append(cniConfig.NetworkConfig.IPAM.Ranges,
			ipam.RangeSet{ipam.Range{Subnet: s.nodeConfig.PodIPv4CIDR.String(), Gateway: s.nodeConfig.GatewayConfig.IPv4.String()}})
	}
	if (s.nodeConfig.GatewayConfig.IPv6 != nil) && (s.nodeConfig.PodIPv6CIDR != nil) {
		cniConfig.NetworkConfig.IPAM.Ranges = append(cniConfig.NetworkConfig.IPAM.Ranges,
			ipam.RangeSet{ipam.Range{Subnet: s.nodeConfig.PodIPv6CIDR.String(), Gateway: s.nodeConfig.GatewayConfig.IPv6.String()}})
	}
	cniConfig.NetworkConfiguration, _ = json.Marshal(cniConfig.NetworkConfig)
}

func (s *IPAMCNIServer) loadNetworkConfig(request *cnipb.CniCmdRequest) (*CNIConfig, error) {
	cniConfig := &CNIConfig{}
	cniConfig.CniCmdArgs = request.CniArgs
	if err := json.Unmarshal(request.CniArgs.NetworkConfiguration, cniConfig); err != nil {
		return cniConfig, err
	}
	cniConfig.k8sArgs = &k8sArgs{}
	if err := cnitypes.LoadArgs(request.CniArgs.Args, cniConfig.k8sArgs); err != nil {
		return cniConfig, err
	}

	if cniConfig.MTU == 0 {
		cniConfig.MTU = s.nodeConfig.NodeMTU
	}
	klog.V(3).Infof("Load network configurations: %v", cniConfig)
	return cniConfig, nil
}

func (s *IPAMCNIServer) decodingFailureResponse(what string) *cnipb.CniCmdResponse {
	return s.generateCNIErrorResponse(
		cnipb.ErrorCode_DECODING_FAILURE,
		fmt.Sprintf("Failed to decode %s", what),
	)
}

func (s *IPAMCNIServer) checkRequestMessage(request *cnipb.CniCmdRequest) (*CNIConfig, *cnipb.CniCmdResponse) {
	cniConfig, err := s.loadNetworkConfig(request)
	if err != nil {
		klog.Errorf("Failed to parse network configuration: %v", err)
		return nil, s.decodingFailureResponse("network config")
	}
	cniVersion := cniConfig.CNIVersion
	// Check if CNI version in the request is supported
	if !s.isCNIVersionSupported(cniVersion) {
		klog.Errorf(fmt.Sprintf("Unsupported CNI version [%s], supported CNI versions %s", cniVersion, version.All.SupportedVersions()))
		return cniConfig, s.incompatibleCniVersionResponse(cniVersion)
	}

	return cniConfig, nil
}

func (s *IPAMCNIServer) incompatibleCniVersionResponse(cniVersion string) *cnipb.CniCmdResponse {
	cniErrorCode := cnipb.ErrorCode_INCOMPATIBLE_CNI_VERSION
	cniErrorMsg := fmt.Sprintf("Unsupported CNI version [%s], supported versions %s", cniVersion, version.All.SupportedVersions())
	return s.generateCNIErrorResponse(cniErrorCode, cniErrorMsg)
}

func (s *IPAMCNIServer) isCNIVersionSupported(reqVersion string) bool {
	_, exist := s.supportedCNIVersions[reqVersion]
	return exist
}

func (s *IPAMCNIServer) generateCNIErrorResponse(cniErrorCode cnipb.ErrorCode, cniErrorMsg string) *cnipb.CniCmdResponse {
	return &cnipb.CniCmdResponse{
		Error: &cnipb.Error{
			Code:    cniErrorCode,
			Message: cniErrorMsg,
		},
	}
}

func (s *IPAMCNIServer) unsupportedFieldResponse(key string, value interface{}) *cnipb.CniCmdResponse {
	cniErrorCode := cnipb.ErrorCode_UNSUPPORTED_FIELD
	cniErrorMsg := fmt.Sprintf("Network configuration does not support key %s and value %v", key, value)
	return s.generateCNIErrorResponse(cniErrorCode, cniErrorMsg)
}

func (s *IPAMCNIServer) getSubnet(podNamespace string) (cidr string, name string, err error) {

	ns, err := s.kubeClient.CoreV1().Namespaces().Get(context.TODO(), podNamespace, metav1.GetOptions{})
	if err != nil {
		klog.Errorf("get namespace: %s err", podNamespace)
		return "", "", fmt.Errorf("get namespace: %s err", podNamespace)
	}

	subnetName := ns.Annotations["subnet"]

	if subnetName == "" {
		subnetName = "defaultnet"
	}
	subnet, err := s.crdClient.IpamV1alpha1().SubNets("default").Get(context.TODO(), subnetName, metav1.GetOptions{})
	if err != nil {
		return "", "", fmt.Errorf("get subnet: %s error", subnetName)
	}
	cidr = subnet.Spec.CIDR
	name = subnet.Name
	return
}

func (s *IPAMCNIServer) CmdAdd(ctx context.Context, request *cnipb.CniCmdRequest) (*cnipb.CniCmdResponse, error) {
	klog.Infof("IPAM Received CmdAdd request %v", request)
	cniConfig, response := s.checkRequestMessage(request)
	if response != nil {
		return response, nil
	}
	klog.Errorf("cniConfig %v", cniConfig)
	args := request.CniArgs

	crdStore, err := crdstore.New(
		cniConfig.CNIVersion,
		string(cniConfig.k8sArgs.K8S_POD_NAME),
		string(cniConfig.k8sArgs.K8S_POD_NAMESPACE),
		s.crdClient,
		s.kubeClient,
		s.podLister,
		s.nsLister,
		s.subnetLister,
	)

	if err != nil {
		return nil, err
	}

	ipamConf, confVersion, err := allocator.LoadIPAMConfigCRD(cniConfig.CNIVersion, crdStore.Subnet)
	if err != nil {
		return nil, err
	}

	// init result obj
	result := &current.Result{}

	// parese dns file to struct obj if it is given
	if ipamConf.ResolvConf != "" {
		dns, err := IPAMDNS.ParseResolvConf(ipamConf.ResolvConf)
		if err != nil {
			return nil, err
		}
		result.DNS = *dns
	}

	// Keep the allocators we used, so we can release all IPs if an error
	// occurs after we start allocating
	allocs := []*allocator.IPAllocator{}
	// Store all requested IPs in a map, so we can easily remove ones we use
	// and error if some remain
	pod, err := s.podLister.Pods(string(cniConfig.k8sArgs.K8S_POD_NAMESPACE)).Get(string(cniConfig.k8sArgs.K8S_POD_NAME))
	if err != nil {
		return nil, err
	}

	for idx, rangeset := range ipamConf.Ranges {
		allocator := allocator.NewIPAllocator(&rangeset, crdStore, idx)

		var requestedIP net.IP

		if requestedIPv4, ok := pod.Annotations["IPv4"]; ok {
			requestedIP = net.ParseIP(requestedIPv4)
		}

		ipConf, err := allocator.Get(args.ContainerId, args.Ifname, requestedIP)
		if err != nil {
			// Deallocate all already allocated IPs
			for _, alloc := range allocs {
				_ = alloc.Release(args.ContainerId, args.Ifname)
			}
			return nil, fmt.Errorf("failed to allocate for range %d: %v", idx, err)
		}

		allocs = append(allocs, allocator)

		result.IPs = append(result.IPs, ipConf)
	}

	defaultV4RouteDst := "0.0.0.0/0"
	_, defaultV4RouteDstNet, err := net.ParseCIDR(defaultV4RouteDst)
	if err != nil {
		return nil, err
	}
	result.Routes = append(result.Routes, &cnitypes.Route{Dst: *defaultV4RouteDstNet, GW: result.IPs[0].Gateway})
	newResult, err := result.GetAsVersion(confVersion)
	if err != nil {
		return nil, err
	}
	var resultBytes bytes.Buffer
	_ = newResult.PrintTo(&resultBytes)
	klog.Infof("CmdAdd for container %v succeeded", cniConfig.ContainerId)
	return &cnipb.CniCmdResponse{CniResult: resultBytes.Bytes()}, nil
}

// CmdDel delete ip at subnet pool, remove podflows
func (s *IPAMCNIServer) CmdDel(_ context.Context, request *cnipb.CniCmdRequest) (
	*cnipb.CniCmdResponse, error) {
	klog.Infof("Received CmdDel request %v", request)

	cniConfig, response := s.checkRequestMessage(request)
	if response != nil {
		return response, nil
	}

	// del ip in store
	crdStore, err := crdstore.New(
		cniConfig.CNIVersion,
		string(cniConfig.k8sArgs.K8S_POD_NAME),
		string(cniConfig.k8sArgs.K8S_POD_NAMESPACE),
		s.crdClient,
		s.kubeClient,
		s.podLister,
		s.nsLister,
		s.subnetLister,
	)

	if err != nil {
		return nil, err
	}

	if err := crdStore.ReleaseByID(cniConfig.ContainerId, cniConfig.Ifname); err != nil {
		return nil, err
	}

	klog.Infof("CmdDel for container %v succeeded", cniConfig.ContainerId)
	return &cnipb.CniCmdResponse{CniResult: []byte("")}, nil
}

func (s *IPAMCNIServer) CmdCheck(_ context.Context, request *cnipb.CniCmdRequest) (
	*cnipb.CniCmdResponse, error) {
	klog.Infof("Received CmdCheck request %v", request)

	cniConfig, response := s.checkRequestMessage(request)
	if response != nil {
		return response, nil
	}

	infraContainer := cniConfig.getInfraContainer()
	s.containerAccess.lockContainer(infraContainer)
	defer s.containerAccess.unlockContainer(infraContainer)

	klog.Infof("CmdCheck for container %v succeeded", cniConfig.ContainerId)
	return &cnipb.CniCmdResponse{CniResult: []byte("")}, nil
}

func (s *IPAMCNIServer) Run(stopCh <-chan struct{}) error {
	klog.Info("Starting IPAM CNI server")
	defer klog.Info("Shutting down IPAM CNI server")

	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, s.podSynced, s.nsSynced, s.subnetSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	listener, err := util.ListenLocalSocket(s.cniSocket)
	if err != nil {
		klog.Fatalf("Failed to bind on IPAM %s: %v", s.cniSocket, err)
	}
	rpcServer := grpc.NewServer()

	cnipb.RegisterCniServer(rpcServer, s)
	klog.Info("IPAM CNI server is listening ...")
	go func() {
		if err := rpcServer.Serve(listener); err != nil {
			klog.Errorf("IPAM CNI server:Failed to serve connections: %v", err)
		}
	}()
	<-stopCh
	return nil
}
