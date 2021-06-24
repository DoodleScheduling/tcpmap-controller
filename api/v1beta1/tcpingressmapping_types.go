/*


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

package v1beta1

import (
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// TCPIngressMappingSpec defines the desired state of TCPIngressMapping
type TCPIngressMappingSpec struct {
	// +required
	BackendService BackendService `json:"backendService"`

	// +optional
	FrontendService *FrontendService `json:"frontendService,omitempty"`

	// +optional
	TCPConfigMap *TCPConfigMap `json:"tcpConfigMap,omitempty"`
}

type TCPConfigMap struct {
	// +required
	Name string `json:"name"`

	// +optional
	Namespace string `json:"namespace"`
}

type BackendService struct {
	// +required
	Name string `json:"name"`

	// +required
	Port intstr.IntOrString `json:"port"`

	// +optional
	Namespace string `json:"namespace"`
}

type FrontendService struct {
	// +required
	Name string `json:"name"`

	// +optional
	Port string `json:"port"`

	// +optional
	Namespace string `json:"namespace"`
}

// TCPIngressMappingStatus defines the observed state of TCPIngressMapping
type TCPIngressMappingStatus struct {
	// Conditions holds the conditions for the VaultBinding.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
	// +optional
	ElectedPort int32 `json:"electedPort"`
}

const (
	ReadyCondition                    = "Ready"
	FrontendServiceNotFoundReason     = "FrontendServiceNotFound"
	BackendServiceNotFoundReason      = "BackendServiceNotFound"
	TCPConfigMapNotFoundReason        = "TCPConfigMapNotFound"
	FailedRegisterFrontendPortReason  = "FailedRegisterFrontendPort"
	FailedRegisterConfigMapPortReason = "FailedRegisterConfigMapPort"
	BackendPortNotFoundReason         = "BackendPortNotFound"
	NoPortElectedReason               = "NoPortElected"
	PortReadyReason                   = "PortReady"
)

// ConditionalResource is a resource with conditions
type conditionalResource interface {
	GetStatusConditions() *[]metav1.Condition
}

// setResourceCondition sets the given condition with the given status,
// reason and message on a resource.
func setResourceCondition(resource conditionalResource, condition string, status metav1.ConditionStatus, reason, message string) {
	conditions := resource.GetStatusConditions()

	newCondition := metav1.Condition{
		Type:    condition,
		Status:  status,
		Reason:  reason,
		Message: message,
	}

	apimeta.SetStatusCondition(conditions, newCondition)
}

// TCPIngressMappingNotReady
func TCPIngressMappingNotReady(clone TCPIngressMapping, reason, message string) TCPIngressMapping {
	setResourceCondition(&clone, ReadyCondition, metav1.ConditionFalse, reason, message)
	return clone
}

// TCPIngressMappingReady
func TCPIngressMappingReady(clone TCPIngressMapping, reason, message string) TCPIngressMapping {
	setResourceCondition(&clone, ReadyCondition, metav1.ConditionTrue, reason, message)
	return clone
}

// GetStatusConditions returns a pointer to the Status.Conditions slice
func (in *TCPIngressMapping) GetStatusConditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=tcpmap
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].status",description=""
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].message",description=""
// +kubebuilder:printcolumn:name="Port",type="integer",JSONPath=".status.electedPort",description=""
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description=""

// TCPIngressMapping is the Schema for the TCPIngressMappings API
type TCPIngressMapping struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TCPIngressMappingSpec   `json:"spec,omitempty"`
	Status TCPIngressMappingStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TCPIngressMappingList contains a list of TCPIngressMapping
type TCPIngressMappingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TCPIngressMapping `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TCPIngressMapping{}, &TCPIngressMappingList{})
}
