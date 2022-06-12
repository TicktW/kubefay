//go:build !ignore_autogenerated
// +build !ignore_autogenerated

// MIT License Copyright (C) 2022 TicktW@https://github.com/TicktW/kubefay
// Code generated by deepcopy-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SubNet) DeepCopyInto(out *SubNet) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SubNet.
func (in *SubNet) DeepCopy() *SubNet {
	if in == nil {
		return nil
	}
	out := new(SubNet)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SubNet) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SubNetList) DeepCopyInto(out *SubNetList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]SubNet, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SubNetList.
func (in *SubNetList) DeepCopy() *SubNetList {
	if in == nil {
		return nil
	}
	out := new(SubNetList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SubNetList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SubNetSpec) DeepCopyInto(out *SubNetSpec) {
	*out = *in
	if in.Namespaces != nil {
		in, out := &in.Namespaces, &out.Namespaces
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.ExternalIPs != nil {
		in, out := &in.ExternalIPs, &out.ExternalIPs
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.UsedPool != nil {
		in, out := &in.UsedPool, &out.UsedPool
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.UnusedPool != nil {
		in, out := &in.UnusedPool, &out.UnusedPool
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.DNS != nil {
		in, out := &in.DNS, &out.DNS
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SubNetSpec.
func (in *SubNetSpec) DeepCopy() *SubNetSpec {
	if in == nil {
		return nil
	}
	out := new(SubNetSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SubNetStatus) DeepCopyInto(out *SubNetStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SubNetStatus.
func (in *SubNetStatus) DeepCopy() *SubNetStatus {
	if in == nil {
		return nil
	}
	out := new(SubNetStatus)
	in.DeepCopyInto(out)
	return out
}