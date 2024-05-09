/*
Copyright 2024.

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

// NFDeploySpec defines the desired state of NFDeploy
type NFDeploySpec struct {
	GitRepo string `json:"gitrepo,omitempty" yaml:"gitrepo,omitempty"`
	Chart   string `json:"chart,omitempty" yaml:"chart,omitempty"`
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
}

// NFDeployStatus defines the observed state of NFDeploy
type NFDeployStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// NFDeploy is the Schema for the nfdeploys API
type NFDeploy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NFDeploySpec   `json:"spec,omitempty"`
	Status NFDeployStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NFDeployList contains a list of NFDeploy
type NFDeployList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NFDeploy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NFDeploy{}, &NFDeployList{})
}
