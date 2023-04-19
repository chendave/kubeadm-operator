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
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/go-logr/logr"
	operatorv1 "k8s.io/kubeadm/operator/api/v1alpha1"

	b64 "encoding/base64"

	"gopkg.in/yaml.v2"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

var kubeConfigList = [...]string{"admin.conf", "controller-manager.conf", "scheduler.conf"}

func runUpdateUserAccount(spec *operatorv1.UpdateUserAccountSpec, log logr.Logger) error {
	log.Info("Updating User Account")

	// if Regenerate is set, create new configs with existing Certificates
	if spec.Regenerate {
		// backup the existing conf
		for _, config := range kubeConfigList {
			_ = os.Rename(kubernetesDir+config, kubernetesDir+config+".old")
			// regenerate admin.conf, controller-manager.conf, scheduler.conf
			// Pass in version to avoid pulling from internet.
			// TODO: must skip the phase of `show-join-command` post-1.26 as well
			cmd := exec.Command("kubeadm", "init", "--kubernetes-version="+spec.KubernetesVersion, config[:len(config)-5])
			_, err := cmd.CombinedOutput()
			if err != nil {
				log.Error(err, "Can't regenerate user account using new root-ca")
				return err
			}
			log.Info("Updated user account using new root-ca")
		}

	}

	//update CertificateAuthorityData of configs
	cafile, err := os.Open(CertDir + "ca.crt")
	if err != nil {
		log.Error(err, "Can't load cafile:ca.crt")
		return err
	}
	defer cafile.Close()

	ca, err := ioutil.ReadAll(cafile)
	if err != nil {
		log.Error(err, "Can't read cafile:ca.crt")
	}
	ca_encodes := b64.StdEncoding.EncodeToString(ca)

	for _, config := range kubeConfigList {
		configFile, err := os.Open(kubernetesDir + config)
		if err != nil {
			log.Error(err, "Can't load:"+config)
			return err
		}
		defer configFile.Close()

		content, err := ioutil.ReadAll(configFile)
		if err != nil {
			log.Error(err, "Can't read:"+config)
			return err
		}

		kubeconfig := clientcmdapi.Config{}
		err = yaml.Unmarshal(content, &kubeconfig)
		if err != nil {
			log.Error(err, "Can't unmarshal content:"+config)
			return err
		}

		kubeconfig.Clusters["cluster"].CertificateAuthorityData = []byte(ca_encodes)

		newContent, err := yaml.Marshal(&kubeconfig)
		if err != nil {
			log.Error(err, "Can't marshal new content of:"+config)
			return err
		}

		err = ioutil.WriteFile(kubernetesDir+config, newContent, 0600)
		if err != nil {
			log.Error(err, "Can't Write:"+config)
			return err
		}
	}
	return nil
}
