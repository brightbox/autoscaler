/*
Copyright 2020 The Kubernetes Authors.

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

package brightbox

import (
	b64 "encoding/base64"
	"fmt"

	"k8s.io/klog"
)

const (
	preamble = `#!/bin/bash
k8s_version="%s"
kubeadm_join_command="%s"
`
	nodeScript = `
set -e

# Retries a command on failure.
# $1 - the max number of attempts
# $2... - the command to run
retry() {
    local max_attempts="${1}"; shift
    local attempt_num=1

    until "${@}"
    do
        if [ "${attempt_num}" -eq "${max_attempts}" ]
        then
            echo "Attempt ${attempt_num} failed and there are no more attempts l
eft!"
            return 1
        else
            echo "Attempt ${attempt_num} failed! Trying again in ${attempt_num}
seconds..."
            sleep $(( attempt_num=attempt_num + 1 ))
        fi
    done
}

# Installing worker

echo "Disabling IPv6"
echo 1 > /proc/sys/net/ipv6/conf/all/disable_ipv6


echo "writing config files"
mkdir -p /etc/cni/net.d
cat <<EOF | tee /etc/cni/net.d/99-loopback.conf
{
"cniVersion": "0.3.1",
"type": "loopback"
}
EOF
cat <<EOF | tee /etc/sysctl.d/40-kubernetes.conf
net.ipv4.ip_forward=1
net.ipv6.conf.all.disable_ipv6=1
EOF
# Required because kubeadm doesn't propagate the nodeRegistration flags properly
# https://github.com/kubernetes/kubeadm/issues/1021
cat <<EOF | tee /etc/default/kubelet
KUBELET_EXTRA_ARGS=--cloud-provider=external --cgroup-driver=systemd
EOF

echo "Resetting sysctl"
retry 5 systemctl try-restart systemd-sysctl

export DEBIAN_FRONTEND=noninteractive

echo "Adding repositories"
curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add -
apt-add-repository -y ppa:gluster/glusterfs-6
apt-add-repository -y 'deb http://apt.kubernetes.io/ kubernetes-xenial main'

echo "Upgrading existing Packages"
apt-get -qq -y update
apt-get -qq -y upgrade

echo "Installing versioned packages"
apt-get -qq -y install \
	language-pack-en \
	socat \
	conntrack \
	ipset \
	debconf-utils \
	containerd \
	glusterfs-client \
	kubeadm=${k8s_version}-00 \
	kubectl=${k8s_version}-00 \
	kubelet=${k8s_version}-00

apt-mark hold kubelet kubeadm kubectl

echo "Selecting iptables version"
if update-alternatives --set iptables /usr/sbin/iptables-legacy
then
	update-alternatives --set ip6tables /usr/sbin/ip6tables-legacy || true
	update-alternatives --set arptables /usr/sbin/arptables-legacy || true
	update-alternatives --set ebtables /usr/sbin/ebtables-legacy || true
fi

echo "Loading IPVS modules"
for word in ip_vs_wrr ip_vs_sh ip_vs ip_vs_rr nf_conntrack_ipv4 br_netfilter
do
	modprobe -- ${word}
done

echo "Installing bash completion"
kubectl completion bash | tee /etc/bash_completion.d/kubectl >/dev/null 2>&1

if [ -d /run/containerd ]
then
	echo "Make containerd CRI v1 runtime use systemd cgroup"
	mkdir -p /etc/containerd
	cat <<EOF | tee /etc/containerd/config.toml
[plugins."io.containerd.grpc.v1.cri"]
systemd_cgroup = true
EOF
systemctl reload-or-restart containerd

	echo "Setting up critools"
	cat <<EOF | tee /etc/crictl.yaml
runtime-endpoint: unix:///run/containerd/containerd.sock
EOF
fi

echo "Activating kubelet bootstrap services"
systemctl enable kubelet.service

echo "Making time sync run more often"
cat <<EOF | tee /etc/systemd/timesyncd.conf
[Time]
PollIntervalMaxSec=1024
EOF
systemctl try-restart systemd-timesyncd

retry 5 systemctl enable iscsid
retry 5 systemctl start iscsid

echo "Joining cluster"
retry 5 ${kubeadm_join_command}
`
)

func defaultUserData(k8sVersion, joinCommand string) string {
	klog.V(4).Info("defaultUserData")
	klog.V(4).Infof("k8s Version: %q", k8sVersion)
	klog.V(4).Infof("Join Command: %q", joinCommand)

	return b64.StdEncoding.EncodeToString([]byte(
		fmt.Sprintf(preamble, k8sVersion, joinCommand) + nodeScript,
	))
}
