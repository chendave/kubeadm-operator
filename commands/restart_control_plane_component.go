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
	"errors"
	"flag"
	"os"
	"os/exec"
	"time"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	operatorv1 "k8s.io/kubeadm/operator/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	staticPodPath = "/etc/kubernetes/manifests/"
	kubernetesDir = "/etc/kubernetes/"
)

var componentList = []string{"kube-apiserver", "kube-scheduler"}

func runRestartControlPlaneComponet(spec *operatorv1.RestartControlPlaneComponentSpec, log logr.Logger) error {
	log.Info("Restarting Control Plane Component " + spec.ComponentName)
	for _, component := range componentList {
		if spec.ComponentName == component {
			return restartStaticPod(component)
		}
	}
	log.Error(nil, "Can't recognize control plane name:"+spec.ComponentName)
	return errors.New("Can't recognize control plane name:" + spec.ComponentName)

}

func restartStaticPod(name string) error {
	cmd := exec.Command("mv", staticPodPath+name+".yaml", kubernetesDir+name+".yaml")
	_, err := cmd.Output()
	if err != nil {
		log.Log.Error(err, "Can't restart static pod:"+name)
		return err
	}
	time.Sleep(2 * time.Second)

	cmd = exec.Command("mv", kubernetesDir+name+".yaml", staticPodPath+name+".yaml")
	_, err = cmd.Output()
	if err != nil {
		log.Log.Error(err, "Can't restart static pod:"+name)
		return err
	}
	time.Sleep(2 * time.Second)

	// Wait for the static pod ready again.
	err = wait.Poll(time.Second, 200*time.Second, func() (done bool, err error) {
		done = checkIfComponentReady(name)
		return true, nil
	})
	return nil

}

func checkIfComponentReady(podName string) bool {
	nodeName := os.Getenv("MY_NODE_NAME")
	podFullName := podName + "-" + nodeName
	return checkIfPodReady(podFullName, "kube-system")
}

func checkIfPodReady(podFullName string, namespace string) bool {
	kubeconfig := "/etc/kubernetes/admin.conf"
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return false
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return false
	}

	podsClient := clientset.CoreV1().Pods(namespace)
	pod, err := podsClient.Get(context.TODO(), podFullName, metav1.GetOptions{})
	if pod.Status.ContainerStatuses[0].Ready == true {
		return true
	}
	return false
}
