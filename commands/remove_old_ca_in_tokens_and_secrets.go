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

func runRemoveOldCaInTokensAndSecrets(spec *operatorv1.RemoveOldCaInTokensAndSecretsSpec, log logr.Logger) error {
	log.Info("Removing old ca in tokens and secrets")
	cmd := exec.Command("bash", "-c", "base64_encoded_ca=\"$(base64 -w0 /etc/kubernetes/pki/ca.crt)\"\n"+
		"for namespace in $(kubectl get namespace --no-headers -o name | cut -d / -f 2 ); do\n"+
		"for token in $(kubectl get secrets --namespace \"$namespace\" --field-selector type=kubernetes.io/service-account-token -o name); do\n"+
		"kubectl get $token --namespace \"$namespace\" -o yaml | /bin/sed \"s/\\(ca.crt:\\).*/\\1 ${base64_encoded_ca}/\" | kubectl apply -f - \n"+
		"done\n"+
		"done")
	_, err := cmd.Output()
	if err != nil {
		log.Error(err, "Can't remove old CA in tokens and secrets")
	}

	return nil
}
