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

type UpfNadResourceSpec struct {
	N3Cni    string `json:"n3cni,omitempty"`
	N3Master string `json:"n3master,omitempty"`
	N3Gw     string `json:"n3gw,omitempty"`
	N4Cni    string `json:"n4cni,omitempty"`
	N4Master string `json:"n4master,omitempty"`
	N4Gw     string `json:"n4gw,omitempty"`
	N6Cni    string `json:"n6cni,omitempty"`
	N6Master string `json:"n6master,omitempty"`
	N6Gw     string `json:"n6gw,omitempty"`
}

// NfNetworkResourceSpec defines the desired state of NfNetworkResource
type NfNetworkResourceSpec struct {
	Namespace string             `json:"namespace,omitempty"`
	UpfNad    UpfNadResourceSpec `json:"upfnad,omitempty"`
}

// NfNetworkResourceStatus defines the observed state of NfNetworkResource
type NfNetworkResourceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// NfNetworkResource is the Schema for the nfnetworkresources API
type NfNetworkResource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NfNetworkResourceSpec   `json:"spec,omitempty"`
	Status NfNetworkResourceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NfNetworkResourceList contains a list of NfNetworkResource
type NfNetworkResourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NfNetworkResource `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NfNetworkResource{}, &NfNetworkResourceList{})
}
