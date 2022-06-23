package ipam

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

func addNamespaceHandler(c *Controller, key string) error {
	klog.Info("add namespace:", key)
	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	ns, err := c.nsLister.Get(name)
	if err != nil {
		return err
	}
	if subnet, ok := ns.Annotations["subnet"]; ok {
		klog.Info(subnet)

		subnet, err := c.subnetLister.SubNets(DefaultNamespace).Get(subnet)
		if err != nil {
			klog.Error(err)
			return err
		}

		for _, subns := range subnet.Spec.Namespaces {
			if subns == name {
				klog.Warning(fmt.Sprintf("namespace:%s has existed in subnet %s", subns, subnet.Name))
				return nil
			}
		}

		subnet = subnet.DeepCopy()
		subnet.Spec.Namespaces = append(subnet.Spec.Namespaces, ns.Name)
		subnet.Spec.LastReservedIP = ""
		c.subnetclientset.IpamV1alpha1().SubNets(DefaultNamespace).Update(context.TODO(), subnet, metav1.UpdateOptions{})

	} else {
		// TODO: cache default subnet
		subnet, err := c.subnetLister.SubNets(DefaultNamespace).Get(DefaultNet)
		if err != nil {
			klog.Error(err)
			return err
		}

		for _, subns := range subnet.Spec.Namespaces {
			if subns == name {
				klog.Warning(fmt.Sprintf("namespace:%s has existed in subnet %s", subns, subnet.Name))
				return nil
			}
		}

		subnet = subnet.DeepCopy()
		subnet.Spec.Namespaces = append(subnet.Spec.Namespaces, ns.Name)
		// make d difference between cmdAdd and namespace add
		subnet.Spec.LastReservedIP = ""
		c.subnetclientset.IpamV1alpha1().SubNets(DefaultNamespace).Update(context.TODO(), subnet, metav1.UpdateOptions{})
	}
	return nil
}

func updateNamespaceHandler(c *Controller, key string) error {
	klog.Info("update namespace:", key)
	// use same logic to deal with update
	addNamespaceHandler(c, key)
	return nil
}

func delNamespaceHandler(c *Controller, key string) error {
	klog.Info("del namespace:", key)
	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	ns, err := c.nsLister.Get(name)
	if err != nil {
		return err
	}

	if subnet, ok := ns.Annotations["subnet"]; ok {
		klog.Info(subnet)

		subnet, err := c.subnetLister.SubNets(DefaultNamespace).Get(subnet)
		if err != nil {
			klog.Error(err)
			return err
		}

		subnet = subnet.DeepCopy()

		klog.Info("del before", subnet.Spec.Namespaces)
		subnet.Spec.Namespaces = delItem(subnet.Spec.Namespaces, name)
		klog.Info("after del:", subnet.Spec.Namespaces)

		// subnet.Spec.Namespaces = append(subnet.Spec.Namespaces, ns.Name)
		c.subnetclientset.IpamV1alpha1().SubNets(DefaultNamespace).Update(context.TODO(), subnet, metav1.UpdateOptions{})
	}
	return nil
}

func delItem(vs []string, s string) []string {
	for i := 0; i < len(vs); i++ {
		if s == vs[i] {
			vs = append(vs[:i], vs[i+1:]...)
			i = i - 1
		}
	}
	return vs
}
