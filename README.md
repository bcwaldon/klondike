# konduct

konduct enables the deploy and manage a Kubernetes cluster using AWS CloudFormation and Ansible.
konduct can send metrics to Datadog, aggregate logs via Logstash, and securely manage secrets using Ansible Vault.

This project is opinionated in how it deploys Kubernetes.
All feedback and code contributions are welcome, but the scope of this project is only intended to reflect one method of deployment.
konduct will not grow to encompass other deployment methodologies.

konduct is licensed under Apache 2.0.
See [LICENSE](LICENSE) for more details.

## Prerequisites

The following tools must be installed locally to use konduct:

- [cfssl](https://github.com/cloudflare/cfssl)
- [awscli](https://aws.amazon.com/cli/)
- [ansible](https://www.ansible.com/) 2.0+
- [kubectl](http://kubernetes.io/docs/user-guide/kubectl-overview/) 1.2+

You must have locally-configured credentials for your AWS account working with the `aws` CLI tool.
Ensure you've chosed the appropriate profile before deploying anything!

```
export AWS_PROFILE=<YOUR-PROFILE-NAME>
```

### Installing cfssl

For OS X users, this is available via homebrew:

```
brew install cfssl
```

On other operating systems, `cfssl` must be installed manually.
Follow the [instructions on GitHub](https://github.com/cloudflare/cfssl#installation).

### Installing awscli

This tool is installed using `pip`:

```
pip install awscli
```

### Installing ansible

For OS X users, this is available via homebrew:

```
brew install ansible
```

Linux users can rely on packaging.
For example, on Fedora 23:

```
dnf install ansible
```

### Installing kubectl

For OS X users, this is available via homebrew:

```
brew install kubernetes-cli
```

Linux users can rely on packaging.
For example, on Fedora 23:

```
dnf install kubernetes-client
```

The Kubernetes project also ships binaries via Google Storage:

```
wget https://storage.googleapis.com/kubernetes-release/release/v1.2.1/bin/darwin/amd64/kubectl
```

## Quickstart

Follow these steps only after meeting the prerequisites documented above.

1. Check out the project

	```
	git clone git@github.com:bcwaldon/konduct.git
	```

2. Choose a name for your cluster, and set it in an environment variable:

	```
	export CLUSTER=<YOUR-CLUSTER-NAME>
	```

3. Create a workspace for your cluster in the git checkout directory and copy the sample config file:

	```
	cd konduct/
	mkdir -p clusters/$CLUSTER/
	cp contrib/main.yml.sample clusters/$CLUSTER/main.yml
	```

4. Update your cluster config file, `clusters/$CLUSTER/main.yml`, with all unset fields. At this time, do not set `ssh_bastion`. We will do this later.

5. From the root of the git checkout diretory, run the `configure.yml` playbook:

	```
	ansible-playbook -e cluster=$CLUSTER configure.yml
	```

6. Create the CloudFormation stack and wait for creation to complete:

	```
	./clusters/$CLUSTER/create-stack.sh
	```

7. Once your CloudFormation stack is ready, identify a public IP of one of the controller instances (identified by the resource tags `tag:KubernetesCluster : $CLUSTER` and `tag:group : controller`). Set the value of `ssh_bastion` in your cluster config and re-run the configure.yml playbook.

	**NOTE:** This is clearly an unfortunate hack. This will improve.

	```
	export SSH_BASTION=<YOUR-CONTROLLER-IP>
	sed -i "" "s/^ssh_bastion:.*$/ssh_bastion: $SSH_BASTION/g" clusters/porter/main.yml
	ansible-playbook -e cluster=$CLUSTER configure.yml
	```

8. Deploy Kubernetes to the cluster using the cluster's `ansible.cfg`. This will take a while!

	```
	ANSIBLE_CONFIG=clusters/$CLUSTER/ansible.cfg ansible-playbook -e cluster=$CLUSTER site.yml
	```

9. Deploy the cluster-level components onto Kubernetes

	```
	ANSIBLE_CONFIG=clusters/$CLUSTER/ansible.cfg ansible-playbook -e cluster=$CLUSTER cluster.yml
	```

Now your cluster is ready to go.
A kubeconfig will be available at `clusters/$CLUSTER/kubeconfig` that provides you admin access.
Validate everything is working by querying the cluster for all nodes:

```
kubectl --kubeconfig=clusters/$CLUSTER/kubeconfig get nodes
```
