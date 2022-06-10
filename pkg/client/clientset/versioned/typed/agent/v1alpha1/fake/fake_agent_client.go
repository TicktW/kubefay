// MIT License Copyright (C) 2022 TicktW
// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1alpha1 "../pkg/client/clientset/versioned/typed/agent/v1alpha1"
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeAgentV1alpha1 struct {
	*testing.Fake
}

func (c *FakeAgentV1alpha1) AntreaAgentInfos() v1alpha1.AntreaAgentInfoInterface {
	return &FakeAntreaAgentInfos{c}
}

func (c *FakeAgentV1alpha1) AntreaControllerInfos() v1alpha1.AntreaControllerInfoInterface {
	return &FakeAntreaControllerInfos{c}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeAgentV1alpha1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
