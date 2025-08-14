// +groupName=example.com
// +kubebuilder:rbac:groups=example.com,resources=timeprinters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=example.com,resources=timeprinters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=example.com,resources=timeprinters/finalizers,verbs=update

//go:generate controller-gen object rbac:roleName=timeprintersRole crd paths=. output:crd:dir=../../config/crd/bases output:rbac:dir=../../config

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TimePrinterSpec defines the desired state
type TimePrinterSpec struct {
	IntervalSeconds int `json:"intervalSeconds"`
}

type TimePrinterStatus struct {
	StartTime string `json:"startTime,omitempty"`
}

// TimePrinter is the Schema for the timeprinters API
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Namespaced,shortName=tp,singular=timeprinter
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Interval",type="integer",JSONPath=".spec.intervalSeconds"
// +kubebuilder:printcolumn:name="Start",type="string",JSONPath=".status.startTime"
type TimePrinter struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   TimePrinterSpec   `json:"spec"`
	Status TimePrinterStatus `json:"status,omitempty"`
}

// TimePrinterList contains a list of TimePrinter
// +kubebuilder:object:root=true
type TimePrinterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []TimePrinter `json:"items"`
}
