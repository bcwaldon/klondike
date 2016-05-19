# Using kubectl

The `kubectl` CLI tool is used to interact with the Kubernetes API.
This guide assists a user of Kubernetes in the installation, configuration and usage of `kubectl`.

## Installation

For those using OS X, kubectl can be installed via [Homebrew](http://brew.sh/).
Assuming Homebrew is available, run the following:

    brew install kubernetes-cli

kubectl is also packaged for many popular Linux distros.
On Fedora:

    dnf install kubernetes-client

Alternatively, kubectl binary builds are available on Google storage from the Kubernetes team.
Download the latest binary `${VERSION}` (i.e. v1.2.2) for your `${OS}` (i.e. linux, darwin) and `${ARCH}` (i.e. amd64):

    wget https://storage.googleapis.com/kubernetes-release/release/${VERSION}/bin/${OS}/${ARCH}/kubectl

## Configuration

Once kubectl is available, it must be configured to communicate with a Kubernetes cluster.
klondike generates a kubeconfig per cluster containing the necessary cluster configuration and user credentials.
Ask your cluster administrator for this kubeconfig if you do not yet have it.

Once the kubeconfig is available on your workstation, simply run `kubectl` with the `--kubeconfig` flag.
To verify everything is functioning properly, run the following command:

    kubectl --kubeconfig=${PATH_TO_KUBECONFIG} version

The `--kubeconfig` flag may also be provided via an environment variable:

    export KUBECONFIG=${PATH_TO_KUBECONFIG}
    kubectl version

A kubeconfig may actually contain configuration for multiple users and clusters.
Many kubeconfigs may be combined into one by hand, but this process is not defined here.
Be careful making manual edits to your kubeconfig!

The default location kubectl looks for a kubeconfig is `~/.kube/config`.
Feel free to place the klondike kubeconfig here, but this may not be a big time saver if you are interacting with multiple Kubernetes clusters and are not willing to combine multiple configs into this single file.

One last note - `kubectl config` allows a user to interact with a kubeconfig programmatically.
This can be used to modify a kubeconfig, but perhaps of more use is the ability to switch between default clusters in a given kubeconfig.

## Useful kubectl Commands

Here's where to start if you aren't familiar with kubectl commands:

- `kubectl version` not only tells you the version of the kubectl binary, but of the components running in the cluster itself. It's reaching out to the Kubernetes API in a given cluster, which can be a useful step in debugging whether or not your local config is correct.
- `kubectl get pods` lists the running pods in the cluster. A pod is the unit of work in the cluster, representing the actual containers. Run `kubectl get` to see what other types of objects exist in the system.
- `kubectl describe pods <pod>` pulls multiple types of resources together to give the user a report of what's going on with a given pod. Start here when pods are acting funky.
- `kubectl get pods <pod> -o yaml` dumps the config of a pod in yaml format, which can be submitted back using `kubectl replace -f <file>` to quickly iterate on the definition of a pod.
- `kubectl --all-namespaces` and `kubectl --namespace=<namespace>` allow the user to interact with objects across all namespaces, or a specific namespace. The default namespace is actually defined in the kubeconfig, but this is typically "default".