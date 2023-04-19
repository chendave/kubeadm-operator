/*
Copyright 2023 The Kubernetes Authors.

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

package operations

import (
	"crypto"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io"
	"math"
	"math/big"
	"os"
	"time"

	"k8s.io/kubeadm/operator/commands"

	operatorv1 "k8s.io/kubeadm/operator/api/v1alpha1"
)

const (
	certDir                = commands.CertDir
	newCaCertFilename      = commands.NewCaCertFilename
	newCaKeyFilename       = commands.NewCaKeyFilename
	caBundleFilename       = commands.CaBundleFilename
	originalCaCertFilename = commands.OriginalCaCertFilename
	originalCaKeyFilename  = commands.OriginalCaKeyFilename
)

type CertPair struct {
	// +optional
	Certificate []byte `json:"certificate,omitempty"`
	// +optonal
	PrivateKey []byte `json:"privateKey,omitempty"`
}

func setupCaRotation() map[string]string {
	return map[string]string{}
}

func planCaRotation(operation *operatorv1.Operation, spec *operatorv1.CaRotationOperationSpec) *operatorv1.RuntimeTaskGroupList {
	var items []operatorv1.RuntimeTaskGroup

	// Some work should be done before this:
	// 1. make prepare_ca_rotation
	//    this backs up old ca, and creates a set of new rootCA pair
	//    this will update all update tokens and secrets
	// 2. kubectl create secret generic ssh-key-secret  --namespace=operator-system  --from-file=id_rsa=$HOME/.ssh/id_rsa
	//    this agent will use this to ssh node, in order to restart kubelet on node

	// Manager load new CA-pair
	loadNewRootCa(spec)

	// run Phase1 of ca_rotation
	if spec.PhaseNumber == 1 {
		// Manager: create New certs for all kubelets' client key
		if spec.NewKubeletClientCerts == nil { // prevent duplicate generation
			newKubeletClientCerts := make(map[string]*CertPair)
			spec.NewKubeletClientCerts = make(map[string]*operatorv1.CertPair)
			// Create New certs
			createAllNewKubeletClientCerts(spec.NewCaCert, spec.NewCaKey, spec.NodeList, newKubeletClientCerts)

			for key, value := range newKubeletClientCerts {
				spec.NewKubeletClientCerts[key] = &operatorv1.CertPair{}
				spec.NewKubeletClientCerts[key].Certificate = value.Certificate
				spec.NewKubeletClientCerts[key].PrivateKey = value.PrivateKey
			}
		}

		// CPN read the new CA-pair in CaRotationOperationSpec and write it to disk
		t01 := createBasicTaskGroup(operation, "01", "write-new-root-ca-to-disk-cp-n")
		setCPNSelector(&t01)
		t01.Spec.Template.Spec.Commands = append(t01.Spec.Template.Spec.Commands,
			operatorv1.CommandDescriptor{
				WriteNewRootCaToDisk: &operatorv1.WriteNewRootCaToDiskSpec{
					CaRotationOperation: spec,
				},
			},
		)
		items = append(items, t01)
		// CP1 restart controller manager
		t02 := createBasicTaskGroup(operation, "02", "restart-controller-manager-cp-1")
		setCP1Selector(&t02)
		t02.Spec.Template.Spec.Commands = append(t02.Spec.Template.Spec.Commands,
			operatorv1.CommandDescriptor{
				RestartControllerManager: &operatorv1.RestartControllerManagerSpec{
					WithCaBundle:        true,
					RemoveOldCaInBundle: false,
				},
			},
		)
		items = append(items, t02)

		// CPN restart controller manager
		t03 := createBasicTaskGroup(operation, "03", "restart-controller-manager-cp-n")
		setCPNSelector(&t03)
		t03.Spec.Template.Spec.Commands = append(t03.Spec.Template.Spec.Commands,
			operatorv1.CommandDescriptor{
				RestartControllerManager: &operatorv1.RestartControllerManagerSpec{
					WithCaBundle:        true,
					RemoveOldCaInBundle: false,
				},
			},
		)
		items = append(items, t03)

		// CP1 update tokens and secrets
		// This has been done in Makefile, so leave empty here

		// Restart kube-proxy and coredns
		t04 := createBasicTaskGroup(operation, "04", "restart-kubeproxy-and-coredns")
		setCP1Selector(&t04)
		t04.Spec.Template.Spec.Commands = append(t04.Spec.Template.Spec.Commands,
			operatorv1.CommandDescriptor{
				RestartKubeproxyAndCoredns: &operatorv1.RestartKubeproxyAndCorednsSpec{},
			},
		)
		items = append(items, t04)

		// CP1 restart apiserver
		t05 := createBasicTaskGroup(operation, "05", "restart-kube-apiserver-cp-1")
		setCP1Selector(&t05)
		t05.Spec.Template.Spec.Commands = append(t05.Spec.Template.Spec.Commands,
			operatorv1.CommandDescriptor{
				RestartControlPlaneComponent: &operatorv1.RestartControlPlaneComponentSpec{
					ComponentName: "kube-apiserver",
				},
			},
		)
		items = append(items, t05)

		// CP1 restart kube-scheduler
		t06 := createBasicTaskGroup(operation, "06", "restart-kube-scheduler-cp-1")
		setCP1Selector(&t06)
		t06.Spec.Template.Spec.Commands = append(t06.Spec.Template.Spec.Commands,
			operatorv1.CommandDescriptor{
				RestartControlPlaneComponent: &operatorv1.RestartControlPlaneComponentSpec{
					ComponentName: "kube-scheduler",
				},
			},
		)
		items = append(items, t06)

		// CPN restart apiserver
		t07 := createBasicTaskGroup(operation, "07", "restart-kube-apiserver-cp-n")
		setCPNSelector(&t07)
		t07.Spec.Template.Spec.Commands = append(t07.Spec.Template.Spec.Commands,
			operatorv1.CommandDescriptor{
				RestartControlPlaneComponent: &operatorv1.RestartControlPlaneComponentSpec{
					ComponentName: "kube-apiserver",
				},
			},
		)
		items = append(items, t07)

		// CPN restart kube-scheduler
		t08 := createBasicTaskGroup(operation, "08", "restart-kube-scheduler-cp-n")
		setCPNSelector(&t08)
		t08.Spec.Template.Spec.Commands = append(t08.Spec.Template.Spec.Commands,
			operatorv1.CommandDescriptor{
				RestartControlPlaneComponent: &operatorv1.RestartControlPlaneComponentSpec{
					ComponentName: "kube-scheduler",
				},
			},
		)
		items = append(items, t08)

		// CP1 update user account(admin.conf scheduler.conf controller-manager.conf)
		t09 := createBasicTaskGroup(operation, "09", "update-user-account-cp-1")
		setCP1Selector(&t09)
		t09.Spec.Template.Spec.Commands = append(t09.Spec.Template.Spec.Commands,
			operatorv1.CommandDescriptor{
				UpdateUserAccount: &operatorv1.UpdateUserAccountSpec{
					Regenerate:        true,
					KubernetesVersion: spec.KubernetesVersion,
				},
			},
		)
		items = append(items, t09)

		// CPN update user account(admin.conf scheduler.conf controller-manager.conf)
		t10 := createBasicTaskGroup(operation, "10", "update-user-account-cp-n")
		setCPNSelector(&t10)
		t10.Spec.Template.Spec.Commands = append(t10.Spec.Template.Spec.Commands,
			operatorv1.CommandDescriptor{
				UpdateUserAccount: &operatorv1.UpdateUserAccountSpec{
					Regenerate:        true,
					KubernetesVersion: spec.KubernetesVersion,
				},
			},
		)
		items = append(items, t10)

		// UPdate CCM's CA(skipped nowadays)

		// Update aggrator-apiserver and webhook apps's CA (skipped nowadays)

		// All nodes: write kubelet's client certs and restart kubelet
		t11 := createBasicTaskGroup(operation, "11", "write-client-certs-of-kubelets")
		t11.Spec.Template.Spec.Commands = append(t11.Spec.Template.Spec.Commands,
			operatorv1.CommandDescriptor{
				WriteNewKubeletCert: &operatorv1.WriteNewKubeletCertSpec{
					CaRotationOperation: spec,
				},
			},
		)
		items = append(items, t11)
		// CP1 create New Certs for apiserver, and restart API server with new certs
		t12 := createBasicTaskGroup(operation, "12", "update-apiserver-certs-cp-1")
		setCP1Selector(&t12)
		t12.Spec.Template.Spec.Commands = append(t12.Spec.Template.Spec.Commands,
			operatorv1.CommandDescriptor{
				UpdateApiserverCerts: &operatorv1.UpdateApiserverCertsSpec{},
			},
		)
		items = append(items, t12)

		t13 := createBasicTaskGroup(operation, "13", "restart-apiserver-cp-1")
		setCP1Selector(&t13)
		t13.Spec.Template.Spec.Commands = append(t13.Spec.Template.Spec.Commands,
			operatorv1.CommandDescriptor{
				RestartControlPlaneComponent: &operatorv1.RestartControlPlaneComponentSpec{
					ComponentName: "kube-apiserver",
				},
			},
		)
		items = append(items, t13)

		// CPN create New Certs for apiserver, and restart API server with new certs
		t14 := createBasicTaskGroup(operation, "14", "update-apiserver-certs-cp-n")
		setCPNSelector(&t14)
		t14.Spec.Template.Spec.Commands = append(t14.Spec.Template.Spec.Commands,
			operatorv1.CommandDescriptor{
				UpdateApiserverCerts: &operatorv1.UpdateApiserverCertsSpec{},
			},
		)
		items = append(items, t14)

		t15 := createBasicTaskGroup(operation, "15", "restart-apiserver-cp-n")
		setCPNSelector(&t15)
		t15.Spec.Template.Spec.Commands = append(t15.Spec.Template.Spec.Commands,
			operatorv1.CommandDescriptor{
				RestartControlPlaneComponent: &operatorv1.RestartControlPlaneComponentSpec{
					ComponentName: "kube-apiserver",
				},
			},
		)
		items = append(items, t15)
	}

	// run phase 2 of ca_rotation
	if spec.PhaseNumber == 2 {
		// Restart apiserver and kube-scheduler
		t16 := createBasicTaskGroup(operation, "16", "restart-scheduler-cp-1")
		setCP1Selector(&t16)
		t16.Spec.Template.Spec.Commands = append(t16.Spec.Template.Spec.Commands,
			operatorv1.CommandDescriptor{
				RestartControlPlaneComponent: &operatorv1.RestartControlPlaneComponentSpec{
					ComponentName: "kube-scheduler",
				},
			},
		)
		items = append(items, t16)

		// Restart apiserver and kube-scheduler
		t17 := createBasicTaskGroup(operation, "17", "restart-scheduler-cp-n")
		setCPNSelector(&t17)
		t17.Spec.Template.Spec.Commands = append(t17.Spec.Template.Spec.Commands,
			operatorv1.CommandDescriptor{
				RestartControlPlaneComponent: &operatorv1.RestartControlPlaneComponentSpec{
					ComponentName: "kube-scheduler",
				},
			},
		)
		items = append(items, t17)

		// Annotate any DaemonSets and Deployments to trigger pod replacement in a safer rolling fashion.(skipped now)

		// Update bootstrap-token if present(skipped now)

		// Verify the status of cluster(skipped now)

		// Remove old CA in cert-dir
		t18 := createBasicTaskGroup(operation, "18", "remove-old-ca-from-disk-cp-1")
		setCP1Selector(&t18)
		t18.Spec.Template.Spec.Commands = append(t18.Spec.Template.Spec.Commands,
			operatorv1.CommandDescriptor{
				RemoveOldRootCaFromDisk: &operatorv1.RemoveOldRootCaFromDiskSpec{
					CaRotationOperation: spec,
				},
			},
		)
		items = append(items, t18)

		t19 := createBasicTaskGroup(operation, "19", "remove-old-ca-from-disk-cp-n")
		setCPNSelector(&t19)
		t19.Spec.Template.Spec.Commands = append(t19.Spec.Template.Spec.Commands,
			operatorv1.CommandDescriptor{
				RemoveOldRootCaFromDisk: &operatorv1.RemoveOldRootCaFromDiskSpec{
					CaRotationOperation: spec,
				},
			},
		)
		items = append(items, t19)

		// CP1 delete old CA in token and secrets
		t20 := createBasicTaskGroup(operation, "20", "remove-old-ca-in-tokens-and-secrets")
		setCP1Selector(&t20)
		t20.Spec.Template.Spec.Commands = append(t20.Spec.Template.Spec.Commands,
			operatorv1.CommandDescriptor{
				RemoveOldCaInTokensAndSecrets: &operatorv1.RemoveOldCaInTokensAndSecretsSpec{},
			},
		)
		items = append(items, t20)

		// CP1 Remove old CA in controller plane
		// Remove old ca in {admin.conf, scheduler.conf, controller-manager.conf}
		t21 := createBasicTaskGroup(operation, "21", "remove-old-ca-in-user-account-cp-1")
		setCP1Selector(&t21)
		t21.Spec.Template.Spec.Commands = append(t21.Spec.Template.Spec.Commands,
			operatorv1.CommandDescriptor{
				UpdateUserAccount: &operatorv1.UpdateUserAccountSpec{
					Regenerate:        false,
					KubernetesVersion: spec.KubernetesVersion,
				},
			},
		)
		items = append(items, t21)

		// Recover args in kube-controller-manager.yaml and restart kube-controller-manager.yaml
		t22 := createBasicTaskGroup(operation, "22", "restart-controller-manager-cp-1")
		setCP1Selector(&t22)
		t22.Spec.Template.Spec.Commands = append(t22.Spec.Template.Spec.Commands,
			operatorv1.CommandDescriptor{
				RestartControllerManager: &operatorv1.RestartControllerManagerSpec{
					WithCaBundle:        false,
					RemoveOldCaInBundle: true,
				},
			},
		)
		items = append(items, t22)
		// Restart apiserver and kube-scheduler
		t23 := createBasicTaskGroup(operation, "23", "restart-apiserver-cp-1")
		setCP1Selector(&t23)
		t23.Spec.Template.Spec.Commands = append(t23.Spec.Template.Spec.Commands,
			operatorv1.CommandDescriptor{
				RestartControlPlaneComponent: &operatorv1.RestartControlPlaneComponentSpec{
					ComponentName: "kube-apiserver",
				},
			},
		)
		items = append(items, t23)
		t24 := createBasicTaskGroup(operation, "24", "restart-scheduler-cp-1")
		setCP1Selector(&t24)
		t24.Spec.Template.Spec.Commands = append(t24.Spec.Template.Spec.Commands,
			operatorv1.CommandDescriptor{
				RestartControlPlaneComponent: &operatorv1.RestartControlPlaneComponentSpec{
					ComponentName: "kube-scheduler",
				},
			},
		)
		items = append(items, t24)

		// CPN Remove old CA in controller plane
		// Remove old ca in {admin.conf, scheduler.conf, controller-manager.conf}
		t25 := createBasicTaskGroup(operation, "25", "remove-old-ca-in-user-account-cp-n")
		setCPNSelector(&t25)
		t25.Spec.Template.Spec.Commands = append(t25.Spec.Template.Spec.Commands,
			operatorv1.CommandDescriptor{
				UpdateUserAccount: &operatorv1.UpdateUserAccountSpec{
					Regenerate:        false,
					KubernetesVersion: spec.KubernetesVersion,
				},
			},
		)
		items = append(items, t25)
		// Recover args in kube-controller-manager.yaml and restart kube-controller-manager.yaml
		t26 := createBasicTaskGroup(operation, "26", "restart-controller-manager-cp-n")
		setCPNSelector(&t26)
		t26.Spec.Template.Spec.Commands = append(t26.Spec.Template.Spec.Commands,
			operatorv1.CommandDescriptor{
				RestartControllerManager: &operatorv1.RestartControllerManagerSpec{
					WithCaBundle:        false,
					RemoveOldCaInBundle: true,
				},
			},
		)
		items = append(items, t26)
		// Restart apiserver and kube-scheduler
		t27 := createBasicTaskGroup(operation, "27", "restart-apiserver-cp-n")
		setCPNSelector(&t27)
		t27.Spec.Template.Spec.Commands = append(t27.Spec.Template.Spec.Commands,
			operatorv1.CommandDescriptor{
				RestartControlPlaneComponent: &operatorv1.RestartControlPlaneComponentSpec{
					ComponentName: "kube-apiserver",
				},
			},
		)
		items = append(items, t27)

		t28 := createBasicTaskGroup(operation, "28", "restart-scheduler-cp-n")
		setCPNSelector(&t28)
		t28.Spec.Template.Spec.Commands = append(t28.Spec.Template.Spec.Commands,
			operatorv1.CommandDescriptor{
				RestartControlPlaneComponent: &operatorv1.RestartControlPlaneComponentSpec{
					ComponentName: "kube-scheduler",
				},
			},
		)
		items = append(items, t28)

		t29 := createBasicTaskGroup(operation, "29", "remove-old-ca-on-nodes")
		t29.Spec.Template.Spec.Commands = append(t29.Spec.Template.Spec.Commands,
			operatorv1.CommandDescriptor{
				RemoveOldCaFromKubeletConfig: &operatorv1.RemoveOldCaFromKubeletConfigSpec{
					NewCaCert: spec.NewCaCert,
				},
			},
		)
		items = append(items, t29)
	}
	return &operatorv1.RuntimeTaskGroupList{
		Items: items,
	}
}

func createAllNewKubeletClientCerts(caCertRaw []byte, caKeyRaw []byte, nodeList []string, newCerts map[string]*CertPair) error {
	// parse caCert and caKey from []bytes
	caCertBlock, _ := pem.Decode(caCertRaw)
	caCert, err := x509.ParseCertificate(caCertBlock.Bytes)
	if err != nil {
		return err
	}
	caKeyBlock, _ := pem.Decode(caKeyRaw)
	caKey, err := x509.ParsePKCS1PrivateKey(caKeyBlock.Bytes)
	if err != nil {
		return err
	}

	for _, node := range nodeList {
		key, err := rsa.GenerateKey(cryptorand.Reader, 2048)
		if err != nil {
			return err
		}
		cert, err := newSignedCert(node, key, caCert, caKey)
		if err != nil {
			return err
		}
		encodedCert, encodedKey := encodeCertPair(cert, key)
		newCerts[node] = &CertPair{}
		newCerts[node].Certificate = encodedCert
		newCerts[node].PrivateKey = encodedKey
	}
	return nil
}

func newSignedCert(nodeName string, key crypto.Signer, caCert *x509.Certificate, caKey crypto.Signer) (*x509.Certificate, error) {
	// prepare args to sign cert
	serial, err := cryptorand.Int(cryptorand.Reader, new(big.Int).SetInt64(math.MaxInt64))
	if err != nil {
		return nil, err
	}
	commonName := "system:node:" + nodeName
	organization := []string{"system:nodes"}
	keyUsage := x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature
	notAfter := time.Now().Add(time.Hour * 24 * 365).UTC()
	extKeyUsage := []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}

	certTmpl := x509.Certificate{
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: organization,
		},
		SerialNumber:          serial,
		NotBefore:             caCert.NotBefore,
		NotAfter:              notAfter,
		KeyUsage:              keyUsage,
		ExtKeyUsage:           extKeyUsage,
		BasicConstraintsValid: true,
		IsCA:                  false,
	}
	certDERBytes, err := x509.CreateCertificate(cryptorand.Reader, &certTmpl, caCert, key.Public(), caKey)
	if err != nil {
		return nil, err
	}
	return x509.ParseCertificate(certDERBytes)
}

func encodeCertPair(cert *x509.Certificate, key crypto.PrivateKey) ([]byte, []byte) {
	certBlock := pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	}
	var keyBlock *pem.Block
	switch t := key.(type) {
	case *rsa.PrivateKey:
		keyBlock = &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(t),
		}
	}
	return pem.EncodeToMemory(&certBlock), pem.EncodeToMemory(keyBlock)
}

func loadNewRootCa(spec *operatorv1.CaRotationOperationSpec) error {
	//log.Info("Loading new root CA on controller plane - 1.")
	newCaCertFile, err := os.Open(certDir + newCaCertFilename)
	if err != nil {
		return err
	}
	defer newCaCertFile.Close()
	spec.NewCaCert, err = io.ReadAll(newCaCertFile)
	if err != nil {
		return err
	}
	newCaKeyFile, err := os.Open(certDir + newCaKeyFilename)
	if err != nil {
		return err
	}
	defer newCaKeyFile.Close()
	spec.NewCaKey, err = io.ReadAll(newCaKeyFile)
	if err != nil {
		return err
	}

	caBundleFile, err := os.Open(certDir + caBundleFilename)
	if err != nil {
		return err
	}
	defer caBundleFile.Close()
	spec.CaBundle, err = io.ReadAll(caBundleFile)
	if err != nil {
		return err
	}

	//log.Info("Successfully loaded new root CA on controller plane - 1.")

	return nil
}
