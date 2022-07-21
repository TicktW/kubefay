// MIT License Copyright (C) 2022 kubefay@https://github.com/kubefay/kubefay
// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	time "time"

	agentv1alpha1 "github.com/kubefay/kubefay/pkg/apis/agent/v1alpha1"
	versioned "github.com/kubefay/kubefay/pkg/client/clientset/versioned"
	internalinterfaces "github.com/kubefay/kubefay/pkg/client/informers/externalversions/internalinterfaces"
	v1alpha1 "github.com/kubefay/kubefay/pkg/client/listers/agent/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// AntreaControllerInfoInformer provides access to a shared informer and lister for
// AntreaControllerInfos.
type AntreaControllerInfoInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.AntreaControllerInfoLister
}

type antreaControllerInfoInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// NewAntreaControllerInfoInformer constructs a new informer for AntreaControllerInfo type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewAntreaControllerInfoInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredAntreaControllerInfoInformer(client, resyncPeriod, indexers, nil)
}

// NewFilteredAntreaControllerInfoInformer constructs a new informer for AntreaControllerInfo type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredAntreaControllerInfoInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.AgentV1alpha1().AntreaControllerInfos().List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.AgentV1alpha1().AntreaControllerInfos().Watch(context.TODO(), options)
			},
		},
		&agentv1alpha1.AntreaControllerInfo{},
		resyncPeriod,
		indexers,
	)
}

func (f *antreaControllerInfoInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredAntreaControllerInfoInformer(client, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *antreaControllerInfoInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&agentv1alpha1.AntreaControllerInfo{}, f.defaultInformer)
}

func (f *antreaControllerInfoInformer) Lister() v1alpha1.AntreaControllerInfoLister {
	return v1alpha1.NewAntreaControllerInfoLister(f.Informer().GetIndexer())
}
