package ipam

import (
	"fmt"
	"time"

	"github.com/TicktW/kubefay/pkg/agent/cniserver"
	"github.com/TicktW/kubefay/pkg/agent/config"
	"github.com/TicktW/kubefay/pkg/agent/util/iptables"

	"github.com/TicktW/kubefay/pkg/agent/openflow"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"

	"github.com/TicktW/kubefay/pkg/agent/controller/noderoute"
	"github.com/TicktW/kubefay/pkg/agent/interfacestore"
	"github.com/TicktW/kubefay/pkg/ovs/ovsconfig"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	crdclients "github.com/TicktW/kubefay/pkg/client/clientset/versioned"
	crdinformers "github.com/TicktW/kubefay/pkg/client/informers/externalversions/ipam/v1alpha1"
	crdlister "github.com/TicktW/kubefay/pkg/client/listers/ipam/v1alpha1"
)

const controllerName = "ipam-controller"

const (
	// SuccessSynced is used as part of the Event 'reason' when a Foo is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a Foo fails
	// to sync due to a Deployment of the same name already existing.
	ErrResourceExists = "ErrResourceExists"

	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a Deployment already existing
	MessageResourceExists = "Resource %q already exists and is not managed by Foo"
	// MessageResourceSynced is the message used for an Event fired when a Foo
	// is synced successfully
	MessageResourceSynced = "Foo synced successfully"

	DefaultNamespace = "default"

	DefaultNet = "defaultnet"
)

// Controller is the main struct to manage subenet, pod, and namespace
type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface
	// ipamclientset is a clientset for our own API group
	subnetclientset crdclients.Interface

	// podlisters and informer
	podLister corelisters.PodLister
	podSynced cache.InformerSynced

	nsLister corelisters.NamespaceLister
	nsSynced cache.InformerSynced

	subnetLister crdlister.SubNetLister
	subnetSynced cache.InformerSynced

	workQueue workqueue.RateLimitingInterface

	dispatcher *Dispatcher

	ovsBridgeClient ovsconfig.OVSBridgeClient

	ofClient openflow.Client

	ipt *iptables.Client

	ifaceStore interfacestore.InterfaceStore

	nodeConfig *config.NodeConfig

	// networkConfig *config.NetworkConfig
}

// NewController will create a controller instance
func NewController(
	kubeclientset kubernetes.Interface,
	subnetclientset crdclients.Interface,
	podInformer coreinformers.PodInformer,
	nsInformer coreinformers.NamespaceInformer,
	subnetInformer crdinformers.SubNetInformer,
	ovsBridgeClient ovsconfig.OVSBridgeClient,
	ofClient openflow.Client,
	nodeConfig *config.NodeConfig,
	) *Controller {

	// Create an ifaceStore that caches network interfaces managed by this node.
	ifaceStore := interfacestore.NewInterfaceStore()

	// v4Enabled := config.IsIPv4Enabled(nodeConfig, networkConfig.TrafficEncapMode)
	// v6Enabled := config.IsIPv6Enabled(nodeConfig, networkConfig.TrafficEncapMode)
	v4Enabled := true
	v6Enabled := false
	ipt, ok := iptables.New(v4Enabled, v6Enabled)
	if ok != nil {
		fmt.Errorf("error creating IPTables instance: %v", ok)
		return nil
	}

	controller := &Controller{
		kubeclientset:   kubeclientset,
		subnetclientset: subnetclientset,
		podLister:       podInformer.Lister(),
		podSynced:       podInformer.Informer().HasSynced,
		nsLister:        nsInformer.Lister(),
		nsSynced:        nsInformer.Informer().HasSynced,
		subnetLister:    subnetInformer.Lister(),
		subnetSynced:    subnetInformer.Informer().HasSynced,
		workQueue:       workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "add_q"),
		dispatcher:      GetDispatcher(),
		ovsBridgeClient: ovsBridgeClient,
		ofClient:        ofClient,
		ipt:             ipt,
		ifaceStore:      ifaceStore,
		nodeConfig:      nodeConfig,
		// networkConfig:   networkConfig,
	}

	klog.Info("init ifacestore")
	controller.initInterfaceStore()

	klog.Info("init open flow client")
	// ofClient.installN
	klog.Info("Setting up event handlers")
	// Set up an event handler for when Foo resources change

	nsInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			klog.Info("================add namespace========================")
			controller.enqueueWorkQueue(obj, "namespace", "add")
		},
		UpdateFunc: func(old, new interface{}) {
			klog.Info("================update namespace========================")
			controller.enqueueWorkQueue(new, "namespace", "update")
		},
		DeleteFunc: func(obj interface{}) {
			klog.Info("================del namespace========================")
			controller.enqueueWorkQueue(obj, "namespace", "del")
		},
	})

	subnetInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			controller.enqueueWorkQueue(obj, "subnet", "add")
		},
		UpdateFunc: func(old, new interface{}) {
			controller.enqueueWorkQueue(new, "subnet", "update")
		},
		DeleteFunc: func(obj interface{}) {
			controller.enqueueWorkQueue(obj, "subnet", "del")
		},
	})

	return controller
}

// initInterfaceStore initializes InterfaceStore with all OVS ports retrieved
// from the OVS bridge.
// TODO support windows
// copy from antrea, change a little, need to optimize
func (c *Controller) initInterfaceStore() error {
	ovsPorts, err := c.ovsBridgeClient.GetPortList()
	if err != nil {
		klog.Errorf("Failed to list OVS ports: %v", err)
		return err
	}

	ifaceList := make([]*interfacestore.InterfaceConfig, 0, len(ovsPorts))
	// uplink name format
	// uplinkIfName := c.nodeConfig.UplinkNetConfig.Name
	for index := range ovsPorts {
		port := &ovsPorts[index]
		ovsPort := &interfacestore.OVSPortConfig{
			PortUUID: port.UUID,
			OFPort:   port.OFPort}
		var intf *interfacestore.InterfaceConfig

		switch {
		case port.IFType == "internal":
			intf = &interfacestore.InterfaceConfig{
				Type:          interfacestore.GatewayInterface,
				InterfaceName: port.Name,
				OVSPortConfig: ovsPort,
			}
		case port.IFType == ovsconfig.GeneveTunnel:
			fallthrough
		case port.IFType == ovsconfig.VXLANTunnel:
			fallthrough
		case port.IFType == ovsconfig.GRETunnel:
			fallthrough
		case port.IFType == ovsconfig.STTTunnel:
			intf = noderoute.ParseTunnelInterfaceConfig(port, ovsPort)
		default:
			// The port should be for a container interface.
			intf = cniserver.ParseOVSPortInterfaceConfig(port, ovsPort)
		}
		if intf != nil {
			ifaceList = append(ifaceList, intf)
		}
	}

	c.ifaceStore.Initialize(ifaceList)
	return nil
}

func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workQueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting ipam controller")

	// Wait for the caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.podSynced, c.nsSynced, c.subnetSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Starting workers")
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down workers")

	return nil
}
