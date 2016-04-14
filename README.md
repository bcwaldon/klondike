# konduct

konduct will deploy and manage a Kubernetes cluster using AWS CloudFormation and Ansible.
konduct can send metrics to Datadog, aggregate logs via Logstash, and securely manage secrets using Ansible Vault.

This project is opinionated in how it deploys Kubernetes and the related services.
All feedback and code contributions are welcome, but the scope of this project is only intended to reflect one method of deployment.
konduct will not grow to encompass other deployment methodologies.

konduct is licensed under Apache 2.0.
See [LICENSE](LICENSE) for more details.

## Overview

konduct deploys three types of hosts: controllers, workers and a bastion.
The controllers operate the Kubernetes control plane, while the workers run pods scheduled to them via the Kubernetes API.

The bastion is an administrative host that facilitates the management of the konduct cluster.
An operator uses the bastion for initial cluster deployment, as well as on-going cluster management tasks.

## Deploy a New Cluster

The creation of a new cluster depends on having installed a few tools locally.
Start off by getting these installed, then continue on with cluster bootstrapping.

### Install Tools

These tools are used to generate the configuration for a new konduct cluster:

- [cfssl](https://github.com/cloudflare/cfssl)
- [awscli](https://aws.amazon.com/cli/)
- [ansible](https://www.ansible.com/) 2.0+

You must have locally-configured credentials for your AWS account working with the `aws` CLI tool.
Ensure you've chosen the appropriate profile before deploying anything!

```
export AWS_PROFILE=<YOUR-PROFILE-NAME>
```

#### Install cfssl

For OS X users, this is available via homebrew:

```
brew install cfssl
```

On other operating systems, `cfssl` must be installed manually.
Follow the [instructions on GitHub](https://github.com/cloudflare/cfssl#installation).

#### Install awscli

This tool is installed using `pip`:

```
pip install awscli
```

#### Install ansible

For OS X users, this is available via homebrew:

```
brew install ansible
```

Linux users can rely on packaging.
For example, on Fedora 23:

```
dnf install ansible
```

### Bootstrap & Deploy

1. Check out the project and `cd` into the directory:

	```
	git clone git@github.com:bcwaldon/konduct.git
	cd konduct/
	```

1. Choose a name for your cluster, and set it in an environment variable:

	```
	export CLUSTER=<YOUR-CLUSTER-NAME>
	```

1. Create a workspace for your cluster in the git checkout directory and copy the sample config file:

	```
	mkdir -p clusters/$CLUSTER/
	cp contrib/main.yml.sample clusters/$CLUSTER/main.yml
	```

1. Update your cluster config file, `clusters/$CLUSTER/main.yml`, with all unset fields. It is encouraged to use the default values, when available.

1. If the S3 bucket you chose for your cluster does not yet exist, create it now:

	```
	aws s3 mb s3://<YOUR-BUCKET>
	```

1. From the root of the git checkout directory, run the `configure.yml` playbook:

	```
	ansible-playbook -e cluster=$CLUSTER configure.yml
	```

1. Upload the created cluster configuration to the S3 bucket you chose for your cluster (see main.yml):

	```
	./contrib/s3-config push $CLUSTER <YOUR-BUCKET>
	```

1. Create the CloudFormation stack and wait for creation to complete. This can take a while!

	```
	./clusters/$CLUSTER/create-stack.sh
	```

1. Once your CloudFormation stack is ready, identify the public IP of your bastion host and SSH to it as the `ubuntu` user using the cluster's deploy key. The remaining steps should be run from the bastion host. To identify the IP of your bastion, open the AWS web console, navigate to the Bastion AutoScaling Group, and find the public IP of the group's single EC2 instance:

	```
	ssh -i clusters/$CLUSTER/id_rsa ubuntu@<YOUR-BASTION-IP>
	```

1. Finish deploying your infrastructure:

	```
	cd /home/ubuntu/konduct/
	ansible-playbook -e cluster=$CLUSTER site.yml
	```

1. Finally, deploy the cluster-level components onto Kubernetes. This may take a few attempts as we must wait for the Kubernetes API to become available:

	```
	ansible-playbook -e cluster=$CLUSTER cluster.yml
	```

Now your cluster is ready to go.
Validate everything is working by querying the cluster for all nodes:

```
kubectl get nodes
```
