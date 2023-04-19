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
	"sort"
	"strings"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	apiv1 "k8s.io/api/core/v1"
	operatorv1 "k8s.io/kubeadm/operator/api/v1alpha1"
)

func runRestartControllerManager(spec *operatorv1.RestartControllerManagerSpec, log logr.Logger) error {
	log.Info("Restarting controller manager")
	file, err := os.Open(staticPodPath + "kube-controller-manager.yaml")
	if err != nil {
		log.Error(err, "Can't load static pod file of controller manager")
		return err
	}
	defer file.Close()

	content, err := ioutil.ReadAll(file)
	if err != nil {
		log.Error(err, "Can't read static pod file of controller manager")
		return err
	}

	pod := apiv1.Pod{}
	err = yaml.Unmarshal(content, &pod)
	if err != nil {
		log.Error(err, "Can't Unmarshal static pod file of controller manager")
		return err
	}

	var cmFlagsPatch map[string]string
	if !spec.RemoveOldCaInBundle && spec.WithCaBundle {
		cmFlagsPatch = map[string]string{
			"client-ca-file":            "/etc/kubernetes/pki/ca.crt.new",
			"cluster-signing-cert-file": "/etc/kubernetes/pki/ca.crt.new",
		}
	} else if spec.RemoveOldCaInBundle && !spec.WithCaBundle {
		cmFlagsPatch = map[string]string{
			"client-ca-file":            "/etc/kubernetes/pki/ca.crt",
			"cluster-signing-cert-file": "/etc/kubernetes/pki/ca.crt",
		}
	} else {
		err = errors.New("Can't recognize pattern of Spec")
		log.Error(err, "")
		return err
	}

	ReplaceFlags(&pod.Spec.Containers[0].Command, cmFlagsPatch)

	newContent, err := yaml.Marshal(&pod)
	if err != nil {
		log.Error(err, "Can't marshal static pod file of controller manager")
		return err
	}

	err = ioutil.WriteFile(staticPodPath+"kube-controller-manager.yaml", newContent, 0600)
	if err != nil {
		log.Error(err, "Can't write static pod file of controller manager")
		return err
	}
	log.Info("Rewrote static pod file: kube-controller-manager.yaml")
	return restartStaticPod("kube-controller-manager")
}

// ReplaceFlags applys the patchs to flags
func ReplaceFlags(flags *[]string, patchs map[string]string) error {
	// for existing flags, patch flags in place
	for i := range *flags {
		key := strings.Split((*flags)[i], "=")[0]

		// Remove leading "--" from flag:
		key = removeLeadingDashesFromString(key)
		if val, ok := patchs[key]; ok {
			(*flags)[i] = "--" + key + "=" + val
			delete(patchs, key)
		}
	}

	// add new lines for new flags
	for key, value := range patchs {
		*flags = append(*flags, "--"+key+"="+value)
	}
	sort.Strings((*flags)[1:])
	return nil
}

func removeLeadingDashesFromString(str string) string {
	if len(str) > 2 && str[0] == '-' && str[1] == '-' {
		return str[2:]
	}
	return str
}
