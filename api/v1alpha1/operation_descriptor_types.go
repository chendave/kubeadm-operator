/*
Copyright 2019 The Kubernetes Authors.

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

// OperatorDescriptor represents an operation to be performed.
// Only one of its members may be specified.
type OperatorDescriptor struct {
	// Upgrade provide declarative support for the kubeadm upgrade workflow.
	// +optional
	Upgrade *UpgradeOperationSpec `json:"upgrade,omitempty"`

	// RenewCertificates provide declarative support for the kubeadm upgrade workflow.
	// +optional
	RenewCertificates *RenewCertificatesOperationSpec `json:"renewCertificates,omitempty"`

	// CaRotation privides specs for the whole CA rotation.
	// +optional
	CaRotaion *CaRotationOperationSpec `json:"caRotation,omitempty"`

	// ModifyConfigs provides component name and args to be modified in cluster.
	// +optional
	ModifyConfigs *ModifyConfigsSpec `json:"modifyConfigs,omitempty"`

	// CustomOperation enable definition of custom list of RuntimeTaskGroup.
	// +optional
	CustomOperation *CustomOperationSpec `json:"custom,omitempty"`
}

// UpgradeOperationSpec provide declarative support for the kubeadm upgrade workflow.
type UpgradeOperationSpec struct {
	// KubernetesVersion specifies the target kubernetes version
	KubernetesVersion string `json:"kubernetesVersion"`
	// KubeadmVersion specifies the target kubeadm version
	// +optional
	KubeadmVersion string `json:"kubeadmVersion"`
	// KubeletVersion specifies the target kubelet version
	// +optional
	KubeletVersion string `json:"kubeletVersion"`
	// KubectlVersion specifies the target kubectl version
	// +optional
	KubectlVersion string `json:"kubectlVersion"`
	Cmd            string `json:"cmd"`
	// NodeIP is the IP address of the host IP.
	NodeIP string `json:"nodeIP"`
}

// RenewCertificatesOperationSpec provide declarative support for the kubeadm upgrade workflow.
type RenewCertificatesOperationSpec struct {
	Args string `json:"args"`
	Cmd  string `json:"cmd"`
}

// CaRotationOperationSpec provides certs for ca-rotation workflow.
type CaRotationOperationSpec struct {
	// KubernetesVersion specifies the target kubernetes version
	KubernetesVersion string `json:"kubernetesVersion"`
	// PhaseNumber provides 1 or 2 to decide which phase to run
	PhaseNumber int `json:"phaseNumber"`
	// Nodelist provides all node's name
	NodeList []string `json:"nodeList,omitempty"`
	// NewCaCert provides a new trust root certificate authority
	// +optional
	NewCaCert []byte `json:"newCaCert,omitempty"`
	// NewCaKey provides private key of new root certificate authority
	// +optional
	NewCaKey []byte `json:"newCaKey,omitempty"`
	// CaBundle includes both old and new root certificate authority
	// +optional
	CaBundle []byte `json:"caBundle,omitempty"`
	// NewKubeletCerts provides client certs of all kubelets
	// +optional
	NewKubeletClientCerts map[string]*CertPair `json:"newKubeletClientCerts,omitempty"`
}

// Certipair stores cert and private key
type CertPair struct {
	// +optional
	Certificate []byte `json:"certificate,omitempty"`
	// +optonal
	PrivateKey []byte `json:"privateKey,omitempty"`
}

type ModifyConfigsSpec struct {
	// FlagsPatchs provides the added and modified args of K8s component
	// +optional
	FlagsPatchs []FlagsPatchSpec `json:"flagsPatchs,omitempty"`
}

type FlagsPatchSpec struct {
	// The name of the component whose parameters need to be modified
	Name string `json:"name"`
	// FlagsPatch provides parameters that need to be added or modified
	FlagsPatch map[string]string `json:"flagsPatch,omitempty"`
}

// CustomOperationSpec enable definition of custom list of RuntimeTaskGroup.
type CustomOperationSpec struct {
	// Workflow allows to define a custom list of RuntimeTaskGroup.
	Workflow []RuntimeTaskGroup `json:"workflow"`
}
