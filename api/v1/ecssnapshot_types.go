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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ECSSnapshotSpec defines the desired state of ECSSnapshot
type ECSSnapshotSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// +kubebuilder:pruning:PreserveUnknownFields
	Domain runtime.RawExtension `json:"domain,omitempty"`
	//request to be executed
	// +kubebuilder:pruning:PreserveUnknownFields
	LifeCycle runtime.RawExtension `json:"lifeCycle,omitempty"`
	//SecretRef
	SecretRef SecretRef `json:"secretRef"`
	//metadata
	SnapshotId string `json:"snapshotId"`
	RegionId   string `json:"regionId"`
}

// ECSSnapshotStatus defines the observed state of ECSSnapshot
type ECSSnapshotStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="SnapshotId",type="string",JSONPath=".spec.domain.SnapshotId",description="snapshot id"
// +kubebuilder:printcolumn:name="SourceRegionId",type="string",JSONPath=".spec.domain.SourceRegionId",description="Sorce region Id"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".spec.domain.Status",description="snapshot status"
// +kubebuilder:printcolumn:name="SnapshotType",type="string",JSONPath=".spec.domain.SnapshotType",description="SnapshotType"
// +kubebuilder:printcolumn:name="SnapshotName",type="string",JSONPath=".spec.domain.SnapshotName",description="SnapshotName"
// ECSSnapshot is the Schema for the ecssnapshots API
type ECSSnapshot struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ECSSnapshotSpec   `json:"spec,omitempty"`
	Status ECSSnapshotStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ECSSnapshotList contains a list of ECSSnapshot
type ECSSnapshotList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ECSSnapshot `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ECSSnapshot{}, &ECSSnapshotList{})
}
