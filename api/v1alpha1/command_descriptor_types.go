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

// CommandDescriptor represents a command to be performed.
// Only one of its members may be specified.
type CommandDescriptor struct {

	// +optional
	KubeadmRenewCertificates *KubeadmRenewCertsCommandSpec `json:"kubeadmRenewCertificates,omitempty"`

	// +optional
	KubeadmUpgradeApply *KubeadmUpgradeApplyCommandSpec `json:"kubeadmUpgradeApply,omitempty"`

	// +optional
	KubeadmUpgradeNode *KubeadmUpgradeNodeCommandSpec `json:"kubeadmUpgradeNode,omitempty"`

	// +optional
	Preflight *PreflightCommandSpec `json:"preflight,omitempty"`

	// +optional
	UpgradeKubeadm *UpgradeKubeadmCommandSpec `json:"upgradeKubeadm,omitempty"`

	// +optional
	UpgradeKubeletAndKubeactl *UpgradeKubeletAndKubeactlCommandSpec `json:"upgradeKubeletAndKubeactl,omitempty"`

	// +optional
	KubectlDrain *KubectlDrainCommandSpec `json:"kubectlDrain,omitempty"`

	// +optional
	KubectlUncordon *KubectlUncordonCommandSpec `json:"kubectlUncordon,omitempty"`

	// +optional
	WriteNewRootCaToDisk *WriteNewRootCaToDiskSpec `json:"writeNewRootCaToDisk,omitempty"`

	// +optional
	RestartControllerManager *RestartControllerManagerSpec `json:"restartControllerManager,omitempty"`

	// +optional
	RestartControlPlaneComponent *RestartControlPlaneComponentSpec `json:"restartControlPlaneComponent,omitempty"`

	// +optional
	RestartKubeproxyAndCoredns *RestartKubeproxyAndCorednsSpec `json:"restartKubeproxyAndCoredns,omitempty"`

	// +optional
	UpdateUserAccount *UpdateUserAccountSpec `json:"updateUserAccount,omitempty"`

	// +optional
	UpdateApiserverCerts *UpdateApiserverCertsSpec `json:"updateApiserverCerts,omitempty"`

	// +optinal
	RemoveOldRootCaFromDisk *RemoveOldRootCaFromDiskSpec `json:"removeOldRootCaFromDisk,omitempty"`

	// +optional
	RemoveOldCaInTokensAndSecrets *RemoveOldCaInTokensAndSecretsSpec `json:"removeOldCaInTokensAndSecrets,omitempty"`

	// +optional
	WriteNewKubeletCert *WriteNewKubeletCertSpec `json:"writeNewKubeletCert,omitempty"`

	// +optional
	RemoveOldCaFromKubeletConfig *RemoveOldCaFromKubeletConfigSpec `json:"removeOldCaOnNodes,omitempty"`

	// Pass provide a dummy command for testing the kubeadm-operator workflow.
	// +optional
	Pass *PassCommandSpec `json:"pass,omitempty"`

	// Fail provide a dummy command for testing the kubeadm-operator workflow.
	// +optional
	Fail *FailCommandSpec `json:"fail,omitempty"`

	// Wait pauses the execution on the next command for a given number of seconds.
	// +optional
	Wait *WaitCommandSpec `json:"wait,omitempty"`
}

// PreflightCommandSpec provides...
type PreflightCommandSpec struct {

	// INSERT ADDITIONAL SPEC FIELDS -
	// Important: Run "make" to regenerate code after modifying this file
}

// UpgradeKubeadmCommandSpec provides kubeadm's target version to upgrade.
type UpgradeKubeadmCommandSpec struct {
	Version string `json:"version"`
}

// KubeadmUpgradeApplyCommandSpec provides the binary and the target version to upgrade.
type KubeadmUpgradeApplyCommandSpec struct {
	Version string `json:"version"`
	Cmd     string `json:"cmd"`
}

// KubeadmUpgradeNodeCommandSpec provides...
type KubeadmUpgradeNodeCommandSpec struct {

	// INSERT ADDITIONAL SPEC FIELDS -
	// Important: Run "make" to regenerate code after modifying this file
}

// KubectlDrainCommandSpec provides...
type KubectlDrainCommandSpec struct {

	// INSERT ADDITIONAL SPEC FIELDS -
	// Important: Run "make" to regenerate code after modifying this file
}

// KubectlUncordonCommandSpec provides...
type KubectlUncordonCommandSpec struct {

	// INSERT ADDITIONAL SPEC FIELDS -
	// Important: Run "make" to regenerate code after modifying this file
}

// UpgradeKubeletAndKubeactlCommandSpec provides target version to upgrade to.
type UpgradeKubeletAndKubeactlCommandSpec struct {
	KubeletVersion string `json:"kubeletVersion"`
	KubectlVersion string `json:"kubectlVersion"`
	NodeIP         string `json:"nodeIP"`
}

// KubeadmRenewCertsCommandSpec provides...
type KubeadmRenewCertsCommandSpec struct {
	Args string `json:"args"`
	Cmd  string `json:"cmd"`
}

// WriteNewRootCaToDiskSpec provides fields to distribute new root ca to all controller planes.
type WriteNewRootCaToDiskSpec struct {
	// +optional
	CaRotationOperation *CaRotationOperationSpec `json:"caRotationOperation,omitempty"`
}

// RemoveOldRootCaFromDiskSpec provides fields to replace old root ca on disk.
type RemoveOldRootCaFromDiskSpec struct {
	// +optional
	CaRotationOperation *CaRotationOperationSpec `json:"caRotationOperation,omitempty"`
}

// RestartControllerManagerSpec provides args to decide whether use ca bundle or only new ca.
type RestartControllerManagerSpec struct {
	// +optional
	WithCaBundle bool `json:"withCaBundle"`
	// +optional
	RemoveOldCaInBundle bool `json:"removeOldCaInBundle"`
}

// RestartControlPlaneComponentSpec provides name of control plane component to restart
type RestartControlPlaneComponentSpec struct {
	ComponentName string `json:"componentName,omitempty"`
}

// RestartKubeproxyAndCorednsSpec provides fields help restart kube-proxy and coredns.
type RestartKubeproxyAndCorednsSpec struct {
}

type UpdateUserAccountSpec struct {
	// +optional
	Regenerate bool `json:"update,omitempty"`
	// KubernetesVersion specifies the target kubernetes version
	KubernetesVersion string `json:"kubernetesVersion"`
}

type UpdateApiserverCertsSpec struct {
}

type RemoveOldCaInTokensAndSecretsSpec struct {
}

type CreateAllNewKubeletClientCertsSpec struct {
	// +optional
	CaRotationOperation *CaRotationOperationSpec `json:"caRotationOperation,omitempty"`
	// +optional
	NodeList []string `json:"nodeList,omitempty"`
}

type WriteNewKubeletCertSpec struct {
	// +optional
	CaRotationOperation *CaRotationOperationSpec `json:"caRotationOperation,omitempty"`
}

type RemoveOldCaFromKubeletConfigSpec struct {
	// +optional
	NewCaCert []byte `json:"newCaCert,omitempty"`
}

// PassCommandSpec provide a dummy command for testing the kubeadm-operator workflow.
type PassCommandSpec struct {
}

// FailCommandSpec provide a dummy command for testing the kubeadm-operator workflow.
type FailCommandSpec struct {
}

// WaitCommandSpec pauses the execution on the next command for a given number of seconds.
type WaitCommandSpec struct {
	// Seconds to pause before next command.
	// +optional
	Seconds int32 `json:"seconds,omitempty"`
}
