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

type WebServer struct {
	DocumentRoot string `json:"documentRoot"`
	ServerAdmin  string `json:"serverAdmin"`

	// +kubebuilder:default=8080
	// +kubebuilder:validation:Minimum=4096
	ServerPort *int32 `json:"serverPort"`
}

type ProxyPath struct {
	Path          string     `json:"path"`
	EndPointsList []EndPoint `json:"endPointsList,omitempty"`
}

type EndPoint struct {
	IPAddress string `json:"ipAddress"`
	Proto     string `json:"proto"`
	Port      int32  `json:"port"`
	Status    bool   `json:"status,omitempty"`
}

type LoadBalancer struct {
	EndPointsList  []EndPoint  `json:"endPointsList,omitempty"`
	Proto          string      `json:"proto,omitempty"`
	Path           string      `json:"path,omitempty"`
	BackEndService string      `json:"backEndService,omitempty"`
	ProxyPaths     []ProxyPath `json:"proxyPaths,omitempty"`

	// +kubebuilder:default=8080
	// +kubebuilder:validation:Minimum=4096
	ServerPort *int32 `json:"serverPort"`
}

type ApachewebSpec struct {
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=5
	Size int32 `json:"size"`

	ServerName string `json:"serverName"`

	// +kubebuilder:validation:Enum={"web", "lb"}
	Type string `json:"type"`

	// +optional
	LoadBalancer *LoadBalancer `json:"loadBalancer,omitempty"`

	// +optional
	WebServer *WebServer `json:"webServer,omitempty"`
}

// ApachewebStatus defines the observed state of Apacheweb
type ApachewebStatus struct {
	// +optional
	EndPoints []EndPoint `json:"endPoints,omitempty"`

	// +optional
	ProxyPaths []ProxyPath `json:"proxyPaths,omitempty"`

	// +optional
	WebServer *WebServer `json:"webServer,omitempty"`
}

// Apacheweb is the Schema for the apachewebs API

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// +kubebuilder:printcolumn:name="Size",type=integer,JSONPath=`.spec.size`
// +kubebuilder:printcolumn:name="Server Name",type=string,JSONPath=`.spec.serverName`
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=".spec.type"
// +kubebuilder:printcolumn:name="Load Balancers",type=string,JSONPath=".spec.loadBalancer"
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
