// MIT License Copyright (C) 2022 TicktW@https://github.com/TicktW/kubefay
// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1alpha1 "github.com/TicktW/kubefay/pkg/apis/agent/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeAntreaAgentInfos implements AntreaAgentInfoInterface
type FakeAntreaAgentInfos struct {
	Fake *FakeAgentV1alpha1
}

var antreaagentinfosResource = schema.GroupVersionResource{Group: "agent.kubefay", Version: "v1alpha1", Resource: "antreaagentinfos"}

var antreaagentinfosKind = schema.GroupVersionKind{Group: "agent.kubefay", Version: "v1alpha1", Kind: "AntreaAgentInfo"}

// Get takes name of the antreaAgentInfo, and returns the corresponding antreaAgentInfo object, and an error if there is any.
func (c *FakeAntreaAgentInfos) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.AntreaAgentInfo, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(antreaagentinfosResource, name), &v1alpha1.AntreaAgentInfo{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.AntreaAgentInfo), err
}

// List takes label and field selectors, and returns the list of AntreaAgentInfos that match those selectors.
func (c *FakeAntreaAgentInfos) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.AntreaAgentInfoList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(antreaagentinfosResource, antreaagentinfosKind, opts), &v1alpha1.AntreaAgentInfoList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.AntreaAgentInfoList{ListMeta: obj.(*v1alpha1.AntreaAgentInfoList).ListMeta}
	for _, item := range obj.(*v1alpha1.AntreaAgentInfoList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested antreaAgentInfos.
func (c *FakeAntreaAgentInfos) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(antreaagentinfosResource, opts))
}

// Create takes the representation of a antreaAgentInfo and creates it.  Returns the server's representation of the antreaAgentInfo, and an error, if there is any.
func (c *FakeAntreaAgentInfos) Create(ctx context.Context, antreaAgentInfo *v1alpha1.AntreaAgentInfo, opts v1.CreateOptions) (result *v1alpha1.AntreaAgentInfo, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(antreaagentinfosResource, antreaAgentInfo), &v1alpha1.AntreaAgentInfo{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.AntreaAgentInfo), err
}

// Update takes the representation of a antreaAgentInfo and updates it. Returns the server's representation of the antreaAgentInfo, and an error, if there is any.
func (c *FakeAntreaAgentInfos) Update(ctx context.Context, antreaAgentInfo *v1alpha1.AntreaAgentInfo, opts v1.UpdateOptions) (result *v1alpha1.AntreaAgentInfo, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(antreaagentinfosResource, antreaAgentInfo), &v1alpha1.AntreaAgentInfo{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.AntreaAgentInfo), err
}

// Delete takes name of the antreaAgentInfo and deletes it. Returns an error if one occurs.
func (c *FakeAntreaAgentInfos) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteActionWithOptions(antreaagentinfosResource, name, opts), &v1alpha1.AntreaAgentInfo{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeAntreaAgentInfos) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(antreaagentinfosResource, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.AntreaAgentInfoList{})
	return err
}

// Patch applies the patch and returns the patched antreaAgentInfo.
func (c *FakeAntreaAgentInfos) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.AntreaAgentInfo, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(antreaagentinfosResource, name, pt, data, subresources...), &v1alpha1.AntreaAgentInfo{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.AntreaAgentInfo), err
}