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
	"os/exec"

	"github.com/go-logr/logr"
	operatorv1 "k8s.io/kubeadm/operator/api/v1alpha1"
)

func runRemoveOldCaFromDisk(spec *operatorv1.RemoveOldRootCaFromDiskSpec, log logr.Logger) error {
	log.Info("Removing old ca from disk")

	cmd := exec.Command("rm", CertDir+CaBundleFilename)
	_, err := cmd.Output()
	if err != nil {
		log.Error(err, "Can't delete CA bundle")
	}

	cmd = exec.Command("cp", CertDir+NewCaCertFilename, CertDir+OriginalCaCertFilename)
	_, err = cmd.Output()
	if err != nil {
		log.Error(err, "Can't recover original ca cert file")
	}

	return nil
}
