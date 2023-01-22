/*
Copyright 2023.

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

// ApachewebSpec defines the desired state of Apacheweb

type BackEnds struct {
	Proto      string `json:"proto"`
	ServerName string `json:"serverName"`
	Port       int32  `json:"port"`
}

type LoadBalancer struct {
	BackEnds []BackEnds `json:"backEnds"`
	Path     string     `json:"path"`
}

type ApachewebSpec struct {
	// +kubebuilder:validation:Minimum=1
	Size int32 `json:"size"`

	ServerName string `json:"serverName"`

	// +kubebuilder:default=8080
	// +kubebuilder:validation:Minimum=4096
	ServerPort int32 `json:"serverPort"`

	// +kubebuilder:validation:Enum={"web", "lb"}
	Type string `json:"type"`

	// +optional
	LoadBalancer *LoadBalancer `json:"loadBalancer,omitempty"`
}

// ApachewebStatus defines the observed state of Apacheweb
type ApachewebStatus struct {
	Size         int32         `json:"size"`
	ServerName   string        `json:"serverPort"`
	ServerPort   int32         `json:"port"`
	Type         string        `json:"type"`
	LoadBalancer *LoadBalancer `json:"loadBalancer,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Apacheweb is the Schema for the apachewebs API
type Apacheweb struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApachewebSpec   `json:"spec,omitempty"`
	Status ApachewebStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ApachewebList contains a list of Apacheweb
type ApachewebList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Apacheweb `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Apacheweb{}, &ApachewebList{})
}