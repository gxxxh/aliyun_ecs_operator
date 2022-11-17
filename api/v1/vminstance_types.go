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

// VMInstanceSpec defines the desired state of VMInstance
type VMInstanceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	//instance info
	// +kubebuilder:pruning:PreserveUnknownFields
	Domain runtime.RawExtension `json:"domain,omitempty"`
	//request to be execute
	// +kubebuilder:pruning:PreserveUnknownFields
	LifeCycle runtime.RawExtension `json:"lifeCycle,omitempty"`
	//regionId and InstanceId, can't be empty
	RegionId   string `json:"regionId"`
	InstanceId string `json:"instanceId"`

	//SrereteRef
	SecretRef SecretRef `json:"secretRef"`
}
type SecretRef struct {
	//secretNamespace
	Namespace string `json:"namespace"`
	//secretName
	Name string `json:"name"`
}

// VMInstanceStatus defines the observed state of VMInstance
type VMInstanceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// http request status
	Status string `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:print
// VMInstance is the Schema for the vminstances API
// +kubebuilder:printcolumn:name="InstanceId",type="string",JSONPath=".spec.domain.InstanceId",description="InstanceId"
// +kubebuilder:printcolumn:name="RegionId",type="string",JSONPath=".spec.domain.RegionId",description="Region Id"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".spec.domain.Status",description="HttpStatus"
// +kubebuilder:printcolumn:name="InstanceType",type="string",JSONPath=".spec.domain.InstanceType",description="InstanceType"
// +kubebuilder:printcolumn:name="ImageId",type="string",JSONPath=".spec.domain.ImageId",description="ImageId"
type VMInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VMInstanceSpec   `json:"spec,omitempty"`
	Status VMInstanceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// VMInstanceList contains a list of VMInstance
type VMInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VMInstance `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VMInstance{}, &VMInstanceList{})
}
