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
	"io/ioutil"
	"os"

	"github.com/go-logr/logr"
	"gopkg.in/yaml.v2"
	apiv1 "k8s.io/api/core/v1"

	operatorv1 "k8s.io/kubeadm/operator/api/v1alpha1"
)

func runModifyConfigs(spec *operatorv1.ModifyConfigsSpec, log logr.Logger) error {
	log.Info("Modifying parameters")
	for _, item := range spec.FlagsPatchs {
		log.Info("Modifying parameters of %s", item.Name)
		if !in(item.Name, componentList) {
			log.Error(nil, "Unrecognized component name")
		}

		// modify static pod's yaml file and restart static pod
		file, err := os.Open(staticPodPath + item.Name + ".yaml")
		if err != nil {
			log.Error(err, "Can't load static pod file of "+item.Name)
			return err
		}
		defer file.Close()

		content, err := ioutil.ReadAll(file)
		if err != nil {
			log.Error(err, "Can't read static pod file of "+item.Name)
			return err
		}

		pod := apiv1.Pod{}
		err = yaml.Unmarshal(content, &pod)
		if err != nil {
			log.Error(err, "Can't Unmarshal static pod file of "+item.Name)
			return err
		}

		ReplaceFlags(&pod.Spec.Containers[0].Command, item.FlagsPatch)

		newContent, err := yaml.Marshal(&pod)
		if err != nil {
			log.Error(err, "Can't marshal static pod file of "+item.Name)
			return err
		}

		err = ioutil.WriteFile(staticPodPath+item.Name+".yaml", newContent, 0600)
		if err != nil {
			log.Error(err, "Can't write static pod file of "+item.Name)
			return err
		}
		log.Info("Successfully rewrote static pod file:" + item.Name + ".yaml")

		restartStaticPod(item.Name)
	}
	return nil
}

func in(target string, strArray []string) bool {
	for _, element := range strArray {
		if target == element {
			return true
		}
	}
	return false
}
