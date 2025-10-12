package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// ChaosDRTestSpec defines the desired state of ChaosDRTest
type ChaosDRTestSpec struct {
	// AppSelector selects the K8s app to test (e.g., app=redis)
	AppSelector map[string]string `json:"appSelector"`
	// ChaosType specifies the chaos experiment (fixed to pod-delete for prototype)
	ChaosType string `json:"chaosType"`
	// ValidationScript runs post-restore (e.g., curl healthz)
	ValidationScript string `json:"validationScript"`
}

// ChaosDRTestStatus defines the observed state of ChaosDRTest
type ChaosDRTestStatus struct {
	Success      bool   `json:"success"`
	ErrorMessage string `json:"errorMessage,omitempty"`
	BackupName   string `json:"backupName,omitempty"`
	RestoreName  string `json:"restoreName,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:path=chaodrtests,scope=Namespaced

// ChaosDRTest is the Schema for the chaodrtests API
type ChaosDRTest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ChaosDRTestSpec   `json:"spec,omitempty"`
	Status ChaosDRTestStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ChaosDRTestList contains a list of ChaosDRTest
type ChaosDRTestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ChaosDRTest `json:"items"`
}

// DeepCopyObject implements runtime.Object for ChaosDRTest
func (in *ChaosDRTest) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyObject implements runtime.Object for ChaosDRTestList
func (in *ChaosDRTestList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func init() {
	SchemeBuilder.Register(&ChaosDRTest{}, &ChaosDRTestList{})
}
