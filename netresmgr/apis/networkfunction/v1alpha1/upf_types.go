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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type NfEndpoint struct {
	Ipv4Addr []string `json:"ipv4Addr"`
	Gwv4Addr string   `json:"gwv4addr"`
}

type UpfN3 struct {
	Endpoints []NfEndpoint `json:"endpoints"`
}

type UpfN4 struct {
	Endpoints []NfEndpoint `json:"endpoints"`
}

type N6Endpoint struct {
	IpEndpoints NfEndpoint `json:"ipendpoints"`
	// UE address pool
	IpAddrPool string `json:"ipaddrpool"`
}

type UpfN6 struct {
	// map of {dnn-name} to
	Endpoints map[string]N6Endpoint `json:"endpoints"`
}

type UpfN9 struct {
	Endpoints []NfEndpoint `json:"endpoints"`
}

// UpfSpec defines the desired state of Upf
type UpfSpec struct {
	UpfClassName string `json:"parent"`
	Namespace    string `json:"namespace"`
	N3           UpfN3  `json:"n3"`
	N4           UpfN4  `json:"n4"`
	N6           UpfN6  `json:"n6"`
	// +optional
	N9 UpfN9 `json:"n9"`
}

// UpfStatus defines the observed state of Upf
type UpfStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Upf is the Schema for the upfs API
type Upf struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UpfSpec   `json:"spec,omitempty"`
	Status UpfStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// UpfList contains a list of Upf
type UpfList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Upf `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Upf{}, &UpfList{})
}
