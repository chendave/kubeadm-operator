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

package operations

import (
	operatorv1 "k8s.io/kubeadm/operator/api/v1alpha1"
)

func setupModifyConfigs() map[string]string {
	return map[string]string{
		"node-role.kubernetes.io/control-plane": "",
	}
}

func planModifyConfigs(operation *operatorv1.Operation, spec *operatorv1.ModifyConfigsSpec) *operatorv1.RuntimeTaskGroupList {
	var items []operatorv1.RuntimeTaskGroup

	t1 := createBasicTaskGroup(operation, "01", "modify-cluster-configs")
	setCPSelector(&t1)
	t1.Spec.Template.Spec.Commands = append(t1.Spec.Template.Spec.Commands,
		operatorv1.CommandDescriptor{
			ModifyConfigs: &operatorv1.ModifyConfigsSpec{FlagsPatchs: spec.FlagsPatchs},
		},
	)
	items = append(items, t1)

	return &operatorv1.RuntimeTaskGroupList{
		Items: items,
	}
}
