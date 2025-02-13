/*
Copyright 2021.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ValuesSpec defines the desired state of Values
type ValuesSpec struct {
	File BufferedFile `json:"file"`
}

// ValuesStatus defines the observed state of Values
type ValuesStatus struct {
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Values is the Schema for the values API
type Values struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ValuesSpec   `json:"spec,omitempty"`
	Status ValuesStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ValuesList contains a list of Values
type ValuesList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Values `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Values{}, &ValuesList{})
}
