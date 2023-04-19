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
	"bytes"
	b64 "encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/util/wait"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	operatorv1 "k8s.io/kubeadm/operator/api/v1alpha1"
)

func runWriteNewKubeletCert(spec *operatorv1.WriteNewKubeletCertSpec, log logr.Logger) error {
	nodename := os.Getenv("MY_NODE_NAME")
	nodeIP := os.Getenv("MY_NODE_IP")

	cafile, err := os.Open(CertDir + "ca.crt")
	if err != nil {
		log.Error(err, "Can't load cafile: ca.crt")
		return err
	}
	defer cafile.Close()

	ca, err := ioutil.ReadAll(cafile)
	if err != nil {
		log.Error(err, "Can't read cafile:ca.crt")
	}
	ca_encodes := b64.StdEncoding.EncodeToString(ca)

	configFile, err := os.Open(kubernetesDir + "kubelet.conf")
	if err != nil {
		log.Error(err, "Can't load:kubelet.conf")
		return err
	}
	defer configFile.Close()

	content, err := ioutil.ReadAll(configFile)
	if err != nil {
		log.Error(err, "Can't read:kubelet.conf")
		return err
	}

	kubeconfig := clientcmdapi.Config{}
	err = yaml.Unmarshal(content, &kubeconfig)
	if err != nil {
		log.Error(err, "Can't unmarshal content:kubelet.conf")
		return err
	}

	kubeconfig.Clusters["cluster"].CertificateAuthorityData = []byte(ca_encodes)

	if kubeconfig.AuthInfos["cluster"].ClientCertificate != "" {
		// Kubelet is using client certificates auto-rotate
		var buffer bytes.Buffer
		buffer.Write(spec.CaRotationOperation.NewKubeletClientCerts[nodename].Certificate)
		buffer.Write(spec.CaRotationOperation.NewKubeletClientCerts[nodename].PrivateKey)

		t := time.Now()
		clientCertFileName := fmt.Sprintf("kubelet-client-%04d-%02d-%02d-%02d-%02d-%02d.pem", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
		// write new cert to disk
		err := ioutil.WriteFile("/var/lib/kubelet/pki/"+clientCertFileName, buffer.Bytes(), 0600)
		if err != nil {
			log.Error(err, "ERROR while create new cert file")
			return err
		}
		// symlink current.pem to new cert
		if _, err := os.Lstat("/var/lib/kubelet/pki/kubelet-client-current.pem"); err == nil {
			os.Remove("/var/lib/kubelet/pki/kubelet-client-current.pem")
		}
		err = os.Symlink("/var/lib/kubelet/pki/"+clientCertFileName, "/var/lib/kubelet/pki/kubelet-client-current.pem")
		if err != nil {
			log.Error(err, "ERROR while symlink new cert file")
			return err
		}
	} else {
		kubeconfig.AuthInfos["user"].ClientCertificateData = spec.CaRotationOperation.NewKubeletClientCerts[nodename].Certificate
		kubeconfig.AuthInfos["user"].ClientKeyData = spec.CaRotationOperation.NewKubeletClientCerts[nodename].PrivateKey
	}

	newContent, err := yaml.Marshal(&kubeconfig)
	if err != nil {
		log.Error(err, "Can't marshal new content of kubelet.conf")
		return err
	}

	err = ioutil.WriteFile(kubernetesDir+"kubelet.conf", newContent, 0600)
	if err != nil {
		log.Error(err, "Can't Write to kubelet.conf")
		return err
	}
	// restart kubelet
	restartKubelet(nodeIP, log)

	return nil
}

func restartKubelet(nodeIP string, log logr.Logger) error {
	file, err := ioutil.TempFile(".", "restartKubelet.*.sh")
	if err != nil {
		log.Error(err, "Cannot create a temp file")
	}
	defer os.Remove(file.Name())
	script := "ssh -o StrictHostKeyChecking=no root" + "@" + nodeIP + " systemctl stop kubelet" + "\n" +
		"ssh -o StrictHostKeyChecking=no root" + "@" + nodeIP + " systemctl start kubelet"

	_, err = file.Write([]byte(script))
	if err != nil {
		log.Error(err, "failed with creating the restart-kubelet script")
	}
	err = wait.Poll(100*time.Millisecond, 30*time.Second, func() (bool, error) {
		cmd := exec.Command("sh", file.Name())
		_, err = cmd.Output()
		if err != nil {
			return false, errors.New("restart kubelet failed with error: " + err.Error())
		}
		log.Info("restart kubelet successfully!")
		return true, nil
	})
	time.Sleep(200 * time.Second)
	return nil
}
