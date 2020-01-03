# Cluster Autoscaler for Brightbox Cloud

This cloud provider implements the autoscaling function for
Brightbox Cloud. The autoscaler should work on any Kubernetes
clusters running on Brightbox Cloud, however the approach
is tailored to clusters built with the [Kubernetes Cluster
Builder](https://github.com/brightbox/kubernetes-cluster)

# How Autoscaler works on Brightbox Cloud

The autoscaler looks for [Server
Groups](https://www.brightbox.com/docs/guides/cli/server-groups/) named
after the cluster-name option passed to the autoscaler  (--cluster-name).

A group named with a suffix of the cluster-name
(e.g. k8s-worker.k8s-test.cluster.local) is a candidate to be a scaling
group. The autoscaler will then check the description to see if it is
a pair of integers separated by a colon (e.g. 1:4). If it finds those
numbers then they will become the minimum and maximum server size for
that group, and autoscaler will attempt to scale the group between those sizes.

The type of server, the image used  and the target zone will be
dynamically determined from the existing members. If these differ, or
there are no existing servers, autoscaler will log an error and will not
scale that group.

A group named precisely the same as the cluster-name
(e.g. k8s-test.cluster.local) is considered to be the default cluster
group and all autoscaled servers created are placed within it as well
as the scaling group.

## Cluster configuration

If you are using the [Kubernetes Cluster
Builder](https://github.com/brightbox/kubernetes-cluster) set the
`worker_min` and `worker_max` values to scale the worker group, and the
`storage_min` and `storage_max` values to scale the storage group.

The Cluster Builder will ensure the group name and description are
updated with the correct values

# Autoscaler Brightbox cloudprovider configuration

The Brightbox Cloud cloudprovider is configured via Environment Variables
suppied to the autoscaler pod. The easiest way to do this is to create
a secret containing the variables within the `kube-system` namespace.

```
---
apiVersion: v1
kind: Secret
type: Opaque
data:
  BRIGHTBOX_API_URL: <base 64 of api URL>
  BRIGHTBOX_CLIENT: <bas64 of Brighbox Cloud client id>
  BRIGHTBOX_CLIENT_SECRET: <base64 of Brightbox Cloud client id secret>
  BRIGHTBOX_KUBE_JOIN_COMMAND: <base64 of cluster join command>
  BRIGHTBOX_KUBE_VERSION: <base 64 of installed k8s version>
metadata:
  name: brightbox-credentials
  namespace: kube-system
```

The join command can be obtained from the kubeadm token command

```
$ kubeadm token create --ttl 0 --description 'Cluster autoscaling token' --print-join-command
```

[Brightbox API
Clients](https://www.brightbox.com/docs/guides/manager/api-clients/)
can be created in the [Brightbox
Manager](https://www.brightbox.com/docs/guides/manager/)

## Cluster Configuration

The [Kubernetes Cluster
Builder](https://github.com/brightbox/kubernetes-cluster) creates a
`brightbox-credentials` secret in the `kube-system` namespace ready
to use.
