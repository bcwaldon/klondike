#!/bin/bash -ex

origin=`pwd`

cd /tmp

wget https://github.com/kubernetes/kubernetes/releases/download/v1.2.0/kubernetes.tar.gz
tar -xf kubernetes.tar.gz kubernetes/server/kubernetes-server-linux-amd64.tar.gz
tar -xf kubernetes/server/kubernetes-server-linux-amd64.tar.gz kubernetes/server/bin/kubelet
mv kubernetes/server/bin/kubelet $origin/roles/kubelet/files/kubelet

rm -fr kubernetes.tar.gz
rm -fr kubernetes/server/kubernetes-server-linux-amd64.tar.gz

