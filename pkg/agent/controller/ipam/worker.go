package ipam

import (
	"context"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

func (c *Controller) runWorker() {
	for c.processNextItem() {
		klog.Info("running")
	}
}

func (c *Controller) enqueueWorkQueue(obj interface{}, resourceType string, actionType string) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	allKeys := []string{resourceType + "/" + actionType, key}
	resKey := strings.Join(allKeys, ":")
	klog.Info("enq workq:", resKey)
	c.workQueue.Add(resKey)
}

// syncAddHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the Foo resource
// with the current status of the resource.
func (c *Controller) syncAddHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	ns, err := c.nsLister.Get(name)
	if err != nil {
		return err
	}
	klog.Info(ns)
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
		c.subnetclientset.IpamV1alpha1().SubNets(DefaultNamespace).Update(context.TODO(), subnet, metav1.UpdateOptions{})
	}

	return nil
}

// processNextItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextItem() bool {
	obj, shutdown := c.workQueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {

		defer c.workQueue.Done(obj)
		var key string
		var ok bool

		if key, ok = obj.(string); !ok {
			c.workQueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		if err := c.dispatcher.DoHandler(key, c); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			c.workQueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}

		c.workQueue.Forget(obj)
		klog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}
