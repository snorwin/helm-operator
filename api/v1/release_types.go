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

// ReleaseSpec defines the desired state of a Helm Release
type ReleaseSpec struct {
	// ChartRef is the reference to the Helm Chart object
	ChartRef ObjectReference `json:"chart"`
	// ValuesRefs are the references to the Helm Values objects
	// +optional
	ValuesRefs []ObjectReference `json:"values,omitempty"`
}

type ObjectReference struct {
	// APIVersion is the API group and version for the resource being referenced
	APIVersion string `json:"apiVersion"`
	// Kind is the type of resource being referenced
	Kind string `json:"kind"`
	// Namespace is the name of resource being referenced
	Namespace string `json:"namespace,omitempty"`
	// Name is the name of resource being referenced
	Name string `json:"name"`
}

// ReleaseStatus defines the observed state of a Helm Release
type ReleaseStatus struct {
	// FirstDeployedTime is when the release was first deployed.
	// +optional
	// +nullable
	FirstDeployedTime metav1.Time `json:"firstDeployedTime,omitempty"`
	// LastDeployedTime is when the release was last deployed.
	// +optional
	// +nullable
	LastDeployedTime metav1.Time `json:"lastDeployedTime,omitempty"`
	// Description is human-friendly "log entry" about this release.
	// +optional
	Description string `json:"description,omitempty"`
	// Status is the current state of the release
	// +optional
	Status string `json:"status,omitempty"`
	// Notes contains the rendered templates/NOTES.txt if available
	// +optional
	Notes string `json:"notes,omitempty"`
	// Version is an int which represents the revision of the release.\
	// +optional
	Version int `json:"version,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Release is the Schema for the releases API
type Release struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ReleaseSpec   `json:"spec,omitempty"`
	Status ReleaseStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ReleaseList contains a list of Release
type ReleaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Release `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Release{}, &ReleaseList{})
}
