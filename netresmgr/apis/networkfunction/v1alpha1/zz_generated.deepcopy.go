//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *N6Endpoint) DeepCopyInto(out *N6Endpoint) {
	*out = *in
	if in.IpEndpoints != nil {
		in, out := &in.IpEndpoints, &out.IpEndpoints
		*out = make([]NfEndpoint, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new N6Endpoint.
func (in *N6Endpoint) DeepCopy() *N6Endpoint {
	if in == nil {
		return nil
	}
	out := new(N6Endpoint)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NfEndpoint) DeepCopyInto(out *NfEndpoint) {
	*out = *in
	if in.Ipv4Addr != nil {
		in, out := &in.Ipv4Addr, &out.Ipv4Addr
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NfEndpoint.
func (in *NfEndpoint) DeepCopy() *NfEndpoint {
	if in == nil {
		return nil
	}
	out := new(NfEndpoint)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Upf) DeepCopyInto(out *Upf) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Upf.
func (in *Upf) DeepCopy() *Upf {
	if in == nil {
		return nil
	}
	out := new(Upf)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Upf) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *UpfClass) DeepCopyInto(out *UpfClass) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new UpfClass.
func (in *UpfClass) DeepCopy() *UpfClass {
	if in == nil {
		return nil
	}
	out := new(UpfClass)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *UpfClass) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *UpfClassList) DeepCopyInto(out *UpfClassList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]UpfClass, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new UpfClassList.
func (in *UpfClassList) DeepCopy() *UpfClassList {
	if in == nil {
		return nil
	}
	out := new(UpfClassList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *UpfClassList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *UpfClassSpec) DeepCopyInto(out *UpfClassSpec) {
	*out = *in
	if in.Dnn != nil {
		in, out := &in.Dnn, &out.Dnn
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new UpfClassSpec.
func (in *UpfClassSpec) DeepCopy() *UpfClassSpec {
	if in == nil {
		return nil
	}
	out := new(UpfClassSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *UpfClassStatus) DeepCopyInto(out *UpfClassStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new UpfClassStatus.
func (in *UpfClassStatus) DeepCopy() *UpfClassStatus {
	if in == nil {
		return nil
	}
	out := new(UpfClassStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *UpfList) DeepCopyInto(out *UpfList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Upf, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new UpfList.
func (in *UpfList) DeepCopy() *UpfList {
	if in == nil {
		return nil
	}
	out := new(UpfList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *UpfList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *UpfN3) DeepCopyInto(out *UpfN3) {
	*out = *in
	if in.Endpoints != nil {
		in, out := &in.Endpoints, &out.Endpoints
		*out = make([]NfEndpoint, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new UpfN3.
func (in *UpfN3) DeepCopy() *UpfN3 {
	if in == nil {
		return nil
	}
	out := new(UpfN3)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *UpfN4) DeepCopyInto(out *UpfN4) {
	*out = *in
	if in.Endpoints != nil {
		in, out := &in.Endpoints, &out.Endpoints
		*out = make([]NfEndpoint, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new UpfN4.
func (in *UpfN4) DeepCopy() *UpfN4 {
	if in == nil {
		return nil
	}
	out := new(UpfN4)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *UpfN6) DeepCopyInto(out *UpfN6) {
	*out = *in
	if in.Endpoints != nil {
		in, out := &in.Endpoints, &out.Endpoints
		*out = make(map[string]N6Endpoint, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new UpfN6.
func (in *UpfN6) DeepCopy() *UpfN6 {
	if in == nil {
		return nil
	}
	out := new(UpfN6)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *UpfN9) DeepCopyInto(out *UpfN9) {
	*out = *in
	if in.Endpoints != nil {
		in, out := &in.Endpoints, &out.Endpoints
		*out = make([]NfEndpoint, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new UpfN9.
func (in *UpfN9) DeepCopy() *UpfN9 {
	if in == nil {
		return nil
	}
	out := new(UpfN9)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *UpfSpec) DeepCopyInto(out *UpfSpec) {
	*out = *in
	in.N3.DeepCopyInto(&out.N3)
	in.N4.DeepCopyInto(&out.N4)
	in.N6.DeepCopyInto(&out.N6)
	in.N9.DeepCopyInto(&out.N9)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new UpfSpec.
func (in *UpfSpec) DeepCopy() *UpfSpec {
	if in == nil {
		return nil
	}
	out := new(UpfSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *UpfStatus) DeepCopyInto(out *UpfStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new UpfStatus.
func (in *UpfStatus) DeepCopy() *UpfStatus {
	if in == nil {
		return nil
	}
	out := new(UpfStatus)
	in.DeepCopyInto(out)
	return out
}
