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
	"context"
	"flag"
	"strings"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	operatorv1 "k8s.io/kubeadm/operator/api/v1alpha1"
)

func runRestartKubeproxyAndCoredns(spec *operatorv1.RestartKubeproxyAndCorednsSpec, log logr.Logger) error {
	log.Info("Restarting Kubeproxy and Coredns ")

	err := deletePodsNameContainsString("kube-proxy", "kube-system")
	if err != nil {
		log.Error(err, "Can't restart Kubeproxy")
		return err
	}

	err = deletePodsNameContainsString("coredns", "kube-system")
	if err != nil {
		log.Error(err, "Can't restart Coredns")
		return err
	}

	return nil

}

func deletePodsNameContainsString(podName string, namespace string) error {
	kubeconfig := "/etc/kubernetes/admin.conf"
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	podsClient := clientset.CoreV1().Pods(namespace)
	allPods, err := podsClient.List(context.TODO(), metav1.ListOptions{})
	for _, pod := range allPods.Items {
		if strings.Contains(pod.Name, podName) {
			podsClient.Delete(context.TODO(), pod.Name, metav1.DeleteOptions{})
		}
	}
	return nil

}
