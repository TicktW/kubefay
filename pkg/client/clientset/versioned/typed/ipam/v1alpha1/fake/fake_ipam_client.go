// MIT License Copyright (C) 2022 TicktW@https://github.com/TicktW/kubefay
// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1alpha1 "github.com/TicktW/kubefay/pkg/client/clientset/versioned/typed/ipam/v1alpha1"
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeIpamV1alpha1 struct {
	*testing.Fake
}

func (c *FakeIpamV1alpha1) SubNets(namespace string) v1alpha1.SubNetInterface {
	return &FakeSubNets{c, namespace}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeIpamV1alpha1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}