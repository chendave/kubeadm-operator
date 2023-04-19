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

	var argPatch map[string]string
	if !spec.RemoveOldCaInBundle && spec.WithCaBundle {
		argPatch = map[string]string{
			"--client-ca-file":            "/etc/kubernetes/pki/ca.crt.new",
			"--cluster-signing-cert-file": "/etc/kubernetes/pki/ca.crt.new",
		}
	} else if spec.RemoveOldCaInBundle && !spec.WithCaBundle {
		argPatch = map[string]string{
			"--client-ca-file":            "/etc/kubernetes/pki/ca.crt",
			"--cluster-signing-cert-file": "/etc/kubernetes/pki/ca.crt",
		}
	} else {
		err = errors.New("Can't recognize pattern of Spec")
		log.Error(err, "")
		return err
	}

	ReplaceArgs(&pod.Spec.Containers[0].Command, argPatch)

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

// ReplaceArgs applys the patchs to args
func ReplaceArgs(args *[]string, patchs map[string]string) error {
	for i := range *args {
		key := strings.Split((*args)[i], "=")[0]
		if val, ok := patchs[key]; ok {
			(*args)[i] = key + "=" + val
		}
	}
	return nil
}
