// MIT License Copyright (C) 2022 TicktW@https://github.com/TicktW/kubefay
// Code generated by informer-gen. DO NOT EDIT.

package externalversions

import (
	"fmt"

	v1alpha1 "github.com/TicktW/kubefay/pkg/apis/agent/v1alpha1"
	ipamv1alpha1 "github.com/TicktW/kubefay/pkg/apis/ipam/v1alpha1"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	cache "k8s.io/client-go/tools/cache"
)

// GenericInformer is type of SharedIndexInformer which will locate and delegate to other
// sharedInformers based on type
type GenericInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() cache.GenericLister
}

type genericInformer struct {
	informer cache.SharedIndexInformer
	resource schema.GroupResource
}

// Informer returns the SharedIndexInformer.
func (f *genericInformer) Informer() cache.SharedIndexInformer {
	return f.informer
}

// Lister returns the GenericLister.
func (f *genericInformer) Lister() cache.GenericLister {
	return cache.NewGenericLister(f.Informer().GetIndexer(), f.resource)
}

// ForResource gives generic access to a shared informer of the matching type
// TODO extend this to unknown resources with a client pool
func (f *sharedInformerFactory) ForResource(resource schema.GroupVersionResource) (GenericInformer, error) {
	switch resource {
	// Group=agent.kubefay, Version=v1alpha1
	case v1alpha1.SchemeGroupVersion.WithResource("antreaagentinfos"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Agent().V1alpha1().AntreaAgentInfos().Informer()}, nil
	case v1alpha1.SchemeGroupVersion.WithResource("antreacontrollerinfos"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Agent().V1alpha1().AntreaControllerInfos().Informer()}, nil

		// Group=ipam.kubefay.github.com, Version=v1alpha1
	case ipamv1alpha1.SchemeGroupVersion.WithResource("subnets"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Ipam().V1alpha1().SubNets().Informer()}, nil

	}

	return nil, fmt.Errorf("no informer found for %v", resource)
}
