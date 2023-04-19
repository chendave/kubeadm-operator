# Kubeadm operator

The kubeadm-operator is an experimental project, still WIP.
Do not use in production.

See [KEP](https://git.k8s.io/enhancements/keps/sig-cluster-lifecycle/kubeadm/20190916-kubeadm-operator.md) for more details.


NOTE: original author is: @fabriziopandini


## upgrade

- support kubeadm upgrade (1.22->1.23, 1.23->1.24)
- support cluster upgrade (1.22->1.23, 1.23->1.24)
- support kubectl upgrade (1.22->1.23, 1.23->1.24)
- support kubelet upgrade (1.22->1.23)
  kubeadm1.23 will set the env `KUBELET_KUBEADM_ARGS` for kubelet, and `--network-plugin=cni` flag will be set, but this flag is not recognized by kubelet 1.24, so we need to update the env to remove the flag manually before the upgrade, this will be done automatically in the future.

  This command is needed on your host to restart kubelet.
  ```
  kubectl create secret generic ssh-key-secret  --namespace=operator-system  --from-file=id_rsa=$HOME/.ssh/id_rsa
  ```

## CA rotation

CA rotation is separated into two phases here, because the manager and agents can't allways access apiserver during the whole progress. Once the cluster only trusts new CA, the certs(signed by old CA) of operator will be invalid. So CA rotation contains two phases in POC:

phase1:
  ```
  make prepare_ca_rotation
  make deploy
  kubectl create secret generic ssh-key-secret  --namespace=operator-system  --from-file=id_rsa=$HOME/.ssh/id_rsa
  kubectl create -f config/samples/operator_v1alpha1_operation_carotation_phase1.yaml
  make undeploy
  ```

Once phase1 successed, undeploy and redeploy kubeadm-operator, to sign operator's cert with new CA.

phase2:
  ```
  make deploy
  kubectl create secret generic ssh-key-secret  --namespace=operator-system  --from-file=id_rsa=$HOME/.ssh/id_rsa
  kubectl create -f config/samples/operator_v1alpha1_operation_carotation_phase2.yaml
  ```
Phase1 ends when apiserver updated certs.
Phase2 start after apiserver holding new certs.