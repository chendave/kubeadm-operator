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

package commands

import (
	"os"
	"os/exec"

	"github.com/go-logr/logr"
	operatorv1 "k8s.io/kubeadm/operator/api/v1alpha1"
)

func runUpdateApiserverCerts(spec *operatorv1.UpdateApiserverCertsSpec, log logr.Logger) error {
	log.Info("Updating Apiserver Certs")

	// backup existing apiserver's certs
	apiserverCerts := [4]string{"apiserver.crt", "apiserver.key", "apiserver-kubelet-client.crt", "apiserver-kubelet-client.key"}
	for _, file := range apiserverCerts {
		err := os.Rename(CertDir+file, CertDir+file+".old")
		if err != nil {
			log.Error(err, "Can't backup:"+file)
			return err
		}
	}

	// Regenerate apiserver's certs with new CA
	// Assuming the new CA already generated in the previous cmd.
	cmd := exec.Command("kubeadm", "init", "phase", "certs", "apiserver")
	_, err := cmd.Output()
	if err != nil {
		log.Error(err, "Can't regenerate apiserver's certs using new root-ca")
		return err
	}

	cmd = exec.Command("kubeadm", "init", "phase", "certs", "apiserver-kubelet-client")
	_, err = cmd.Output()
	if err != nil {
		log.Error(err, "Can't regenerate apiserver-kubelet-client's certs using new root-ca")
		return err
	}

	log.Info("Updated apiserver's certs using new root-ca")

	return nil
}
