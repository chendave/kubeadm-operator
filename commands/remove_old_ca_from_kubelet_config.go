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
	b64 "encoding/base64"
	"io/ioutil"
	"os"

	"github.com/go-logr/logr"
	"gopkg.in/yaml.v2"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	operatorv1 "k8s.io/kubeadm/operator/api/v1alpha1"
)

func runRemoveOldCaFromKubeletConfig(spec *operatorv1.RemoveOldCaFromKubeletConfigSpec, log logr.Logger) error {
	config := "kubelet.conf"
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

	ca_encodes := b64.StdEncoding.EncodeToString(spec.NewCaCert)
	kubeconfig.Clusters["cluster"].CertificateAuthorityData = []byte(ca_encodes)
	newContent, err := yaml.Marshal(&kubeconfig)
	if err != nil {
		log.Error(err, "Can't marshal new content of:"+config)
		return err
	}

	err = ioutil.WriteFile(kubernetesDir+config, newContent, 0600)
	if err != nil {
		log.Error(err, "Can't Write to:"+config)
		return err
	}

	nodeIp := os.Getenv("MY_NODE_IP")
	restartKubelet(nodeIp, log)

	return nil
}
