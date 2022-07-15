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