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

package commands

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	"k8s.io/apimachinery/pkg/util/wait"
	operatorv1 "k8s.io/kubeadm/operator/api/v1alpha1"
)

const bianryDownloadPath = "https://storage.googleapis.com/kubernetes-release/release/%s/bin/linux/%s/%s"

func runUpgradeKubectlAndKubelet(spec *operatorv1.UpgradeKubeletAndKubeactlCommandSpec, log logr.Logger) error {
	file, err := ioutil.TempFile(".", "upgrade.*.sh")
	if err != nil {
		log.Error(err, "Cannot create a temp file")
	}
	defer os.Remove(file.Name())

	kubectlPath := fmt.Sprintf(bianryDownloadPath, spec.KubectlVersion, runtime.GOARCH, "kubectl")
	kubeletPath := fmt.Sprintf(bianryDownloadPath, spec.KubeletVersion, runtime.GOARCH, "kubelet")
	script := "apt update && apt install wget -y \n" +
		"wget " + kubectlPath + " -O /usr/bin/kubectl-" + spec.KubectlVersion + "\n" +
		"wget " + kubeletPath + " -O /usr/bin/kubelet-" + spec.KubeletVersion + "\n" +
		"chmod +x /usr/bin/kubectl-" + spec.KubectlVersion + "\n" +
		"chmod +x /usr/bin/kubelet-" + spec.KubeletVersion + "\n" +
		"cp -f /usr/bin/kubectl-" + spec.KubectlVersion + " /usr/bin/kubectl" + "\n" +
		"ssh -o StrictHostKeyChecking=no root" + "@" + spec.NodeIP + " systemctl stop kubelet" + "\n" +
		"cp -f /usr/bin/kubelet-" + spec.KubeletVersion + " /usr/bin/kubelet" + "\n" +
		"ssh -o StrictHostKeyChecking=no root" + "@" + spec.NodeIP + " systemctl start kubelet"

	_, err = file.Write([]byte(script))
	if err != nil {
		log.Error(err, "failed with creating the upgrade script")
	}
	err = wait.Poll(100*time.Millisecond, 30*time.Second, func() (bool, error) {
		cmd := exec.Command("sh", file.Name())
		_, err = cmd.Output()
		if err != nil {
			return false, errors.New("update failed with error: " + err.Error())
		}
		log.Info("upgrade kubectl and kubelet binary successfully!")
		return true, nil
	})
	return err
}
