package crdstore

import (
	"context"
	"fmt"

	// "fmt"
	"io/ioutil"
	"math"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/containernetworking/plugins/plugins/ipam/host-local/backend"
	"github.com/kubefay/kubefay/pkg/apis/ipam/v1alpha1"
	crdclients "github.com/kubefay/kubefay/pkg/client/clientset/versioned"
	crdlister "github.com/kubefay/kubefay/pkg/client/listers/ipam/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/klog"
)

const lastIPFilePrefix = "last_reserved_ip."
const LineBreak = "\r\n"

var defaultSubnet = "defaultnet"

// Store is a simple disk-backed store that creates one file per IP
// address in a given directory. The contents of the file are the container ID.
type Store struct {
	*FileLock
	podName         string
	podNamespace    string
	subnetName      string
	subnetclientset crdclients.Interface
	Subnet          *v1alpha1.SubNet
	k8sClient       clientset.Interface
	podLister       corelisters.PodLister
	nsLister        corelisters.NamespaceLister
	subnetLister    crdlister.SubNetLister
}

// Store implements the Store interface
var _ backend.Store = &Store{}

func New(
	network,
	podName string,
	podNamespace string,
	crdClientset crdclients.Interface,
	k8sClient clientset.Interface,
	podLister corelisters.PodLister,
	nsLister corelisters.NamespaceLister,
	subnetLister crdlister.SubNetLister,
) (*Store, error) {

	ns, err := nsLister.Get(podNamespace)
	if err != nil {
		klog.Errorf("get namespace: %s err", podNamespace)
		return nil, fmt.Errorf("get namespace: %s err", podNamespace)
	}

	subnetName := ns.Annotations["subnet"]
	if subnetName == "" {
		subnetName = "defaultnet"
	}
	subnetObj, err := subnetLister.SubNets("default").Get(subnetName)
	if err != nil {
		return nil, fmt.Errorf("get subnet: %s error", subnetName)
	}

	dir := filepath.Join("/tmp/", subnetObj.Name, network)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	lk, err := NewFileLock(dir)
	if err != nil {
		return nil, err
	}

	return &Store{lk, podName, podNamespace, subnetObj.Name, crdClientset, subnetObj, k8sClient, podLister, nsLister, subnetLister}, nil
}

// Reserve IP from CRD Subnet.
// if the requsested IP has been allocated, return false, nil
// if error, return false, err
// if reserve IP succeed, return true, nil
func (s *Store) Reserve(id string, ifname string, requestedIP net.IP, rangeID string) (bool, error) {
	// init Subnet CRD
	copySubnet := s.Subnet.DeepCopy()
	if copySubnet.Spec.UsedPool == nil {
		copySubnet.Spec.UsedPool = make(map[string]string)
	}

	usedNum := len(copySubnet.Spec.UsedPool)
	_, network, err := net.ParseCIDR(copySubnet.Spec.CIDR)
	if err != nil {
		return false, err
	}
	ones, bits := network.Mask.Size()
	capNum := int(math.Pow(2, float64(bits-ones)) - 3)

	usedNumStr := strconv.Itoa(usedNum)
	capNumStr := strconv.Itoa(capNum)

	if usedNum == capNum {
		return false, fmt.Errorf("the IP of Subnet:%s has used up. used IPs: %d, cap: %d", copySubnet.Name, usedNum, capNum)
	}

	// whether IP has been used
	if _, ok := copySubnet.Spec.UsedPool[requestedIP.To4().String()]; ok == true {
		return false, nil
	}

	pod, err := s.k8sClient.CoreV1().Pods(s.podNamespace).Get(context.TODO(), s.podName, metav1.GetOptions{})
	if err != nil {
		return false, err
	}

	ipInfo := pod.Spec.NodeName + "/" + s.podNamespace + "/" + s.podName + "/" + strings.TrimSpace(id) + "/" + ifname
	copySubnet.Spec.UsedPool[requestedIP.String()] = ipInfo
	copySubnet.Spec.LastReservedIP = ipInfo + "/" + requestedIP.String()
	copySubnet.Status.PoolStatus = usedNumStr + "/" + capNumStr
	copySubnet.Status.IPAMEvent = "POD_ADD"
	_, err = s.subnetclientset.IpamV1alpha1().SubNets("default").Update(context.TODO(), copySubnet, metav1.UpdateOptions{})
	klog.Infof("update subnet..........%+v", copySubnet)
	if err != nil {
		klog.Info("=============================")
		klog.Error(err)
		return false, err
	}

	return true, nil
}

// TODO: modified
// LastReservedIP returns the last reserved IP if exists
func (s *Store) LastReservedIP(rangeID string) (net.IP, error) {
	lastIPInfo := s.Subnet.Spec.LastReservedIP

	if lastIPInfo == "" {
		return nil, nil
	}

	podInfoSlice := strings.Split(lastIPInfo, "/")

	return net.ParseIP(podInfoSlice[5]), nil
}

func (s *Store) Release(ip net.IP) error {
	return os.Remove(GetEscapedPath(s.subnetName, ip.String()))
}

func (s *Store) FindByKey(id string, ifname string, match string) (bool, error) {
	found := false

	err := filepath.Walk(s.subnetName, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return nil
		}
		if strings.TrimSpace(string(data)) == match {
			found = true
		}
		return nil
	})
	return found, err

}

func (s *Store) FindByID(id string, ifname string) bool {
	s.Lock()
	defer s.Unlock()

	found := false
	match := strings.TrimSpace(id) + LineBreak + ifname
	found, err := s.FindByKey(id, ifname, match)

	// Match anything created by this id
	if !found && err == nil {
		match := strings.TrimSpace(id)
		found, err = s.FindByKey(id, ifname, match)
	}

	return found
}

func (s *Store) ReleaseByKey(id string, ifname string, match string) (bool, error) {
	found := false
	copySubnet := s.Subnet.DeepCopy()
	if copySubnet.Spec.UsedPool == nil {
		return false, fmt.Errorf("the used pool for subnet %s is empty", copySubnet.Name)
	}

	key := ""
	for i, IPInfo := range copySubnet.Spec.UsedPool {
		if match == IPInfo {
			key = i
			continue
		}
	}

	if key == "" {
		return found, fmt.Errorf("can not find ipinfo in used pool %s", match)
	}

	delete(copySubnet.Spec.UsedPool, key)
	_, err := s.subnetclientset.IpamV1alpha1().SubNets("default").Update(context.TODO(), copySubnet, metav1.UpdateOptions{})
	if err != nil {
		klog.Error(err)
		return false, err
	}

	found = true
	return found, nil
}

// N.B. This function eats errors to be tolerant and
// release as much as possible
func (s *Store) ReleaseByID(id string, ifname string) error {
	pod, err := s.k8sClient.CoreV1().Pods(s.podNamespace).Get(context.TODO(), s.podName, metav1.GetOptions{})
	// match := strings.TrimSpace(id) + LineBreak + ifname

	match := pod.Spec.NodeName + "/" + s.podNamespace + "/" + s.podName + "/" + strings.TrimSpace(id) + "/" + ifname
	found, err := s.ReleaseByKey(id, ifname, match)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("can not find pod in subnet")
	}
	return nil
}

// GetByID returns the IPs which have been allocated to the specific ID
func (s *Store) GetByID(id string, ifname string) []net.IP {
	var ips []net.IP

	match := strings.TrimSpace(id) + "/" + ifname
	// matchOld for backwards compatibility
	// matchOld := strings.TrimSpace(id)

	for ip, val := range s.Subnet.Spec.UsedPool {
		if val == match {
			ips = append(ips, net.ParseIP(ip))
		}
	}

	return ips
}

func GetEscapedPath(dataDir string, fname string) string {
	if runtime.GOOS == "windows" {
		fname = strings.Replace(fname, ":", "_", -1)
	}
	return filepath.Join(dataDir, fname)
}
