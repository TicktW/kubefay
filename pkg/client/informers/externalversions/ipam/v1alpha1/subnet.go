// MIT License Copyright (C) 2022 kubefay@https://github.com/kubefay/kubefay
// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	time "time"

	ipamv1alpha1 "github.com/kubefay/kubefay/pkg/apis/ipam/v1alpha1"
	versioned "github.com/kubefay/kubefay/pkg/client/clientset/versioned"
	internalinterfaces "github.com/kubefay/kubefay/pkg/client/informers/externalversions/internalinterfaces"
	v1alpha1 "github.com/kubefay/kubefay/pkg/client/listers/ipam/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// SubNetInformer provides access to a shared informer and lister for
// SubNets.
type SubNetInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.SubNetLister
}

type subNetInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewSubNetInformer constructs a new informer for SubNet type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewSubNetInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredSubNetInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredSubNetInformer constructs a new informer for SubNet type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredSubNetInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.IpamV1alpha1().SubNets(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.IpamV1alpha1().SubNets(namespace).Watch(context.TODO(), options)
			},
		},
		&ipamv1alpha1.SubNet{},
		resyncPeriod,
		indexers,
	)
}

func (f *subNetInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredSubNetInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *subNetInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&ipamv1alpha1.SubNet{}, f.defaultInformer)
}

func (f *subNetInformer) Lister() v1alpha1.SubNetLister {
	return v1alpha1.NewSubNetLister(f.Informer().GetIndexer())
}
