// api/v1/timeprinter_types.go
package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// TimePrinterSpec defines the desired state
type TimePrinterSpec struct {
	IntervalSeconds int `json:"intervalSeconds"`
}

// TimePrinter is the Schema for the timeprinters API
type TimePrinter struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec TimePrinterSpec `json:"spec"`
}

// TimePrinterList contains a list of TimePrinter
type TimePrinterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []TimePrinter `json:"items"`
}

func (t *TimePrinter) DeepCopyObject() runtime.Object {
	return t
}

func (t *TimePrinterList) DeepCopyObject() runtime.Object {
	return t
}
