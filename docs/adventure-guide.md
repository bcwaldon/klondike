# Exploring the Klondike

A Confused Guide with Jacob Straszysnki
---------------------------------------

In this walkthrough we're going to work our way from the bastion host that
arbiters access to our cluster, to a node, which functions as the "physical"
host onto which kubernetes' virtual "pods" are scheduled, and finally to a pod.

Our goal:

  ssh into THE bastion -> ssh into A controller -> curl an nginx pod


Where's my Bastion?
-------------------

Let's set up some environment variables first:

```bash
# One of the kubernetes clusters at the time of writing.
export CLUSTER=cascade

# The utilities in ./contrib use s3 to store various artifacts that should
# live in the same AWS profile as the cluster itself.
export AWS_PROFILE=staging
```

Klondike's [./contrib](contrib) folder provides some useful commands for
obtaining cluster metadata as well as pulling the build artifacts generated
during a klondike gold rush. Let's figure out the bastion's IP first:


```bash
python ./contrib/cluster-info.py $CLUSTER
```

Output:

```
stack_name: klondike-cluster-cascade
stack_status: CREATE_COMPLETE
bastion_public_ip: ***.***.***.***
worker_asg_capacity: 1
worker_count: 1
worker_asg_name: klondike-cluster-cascade-WorkerASG-WAUUXDJRP8LF
```

Our bastion's IP address is `***.***.***.***`.

SSH into the Bastion!
---------------------

When the `cascade` cluster was [./README.md](created) we uploaded artifacts to
S3. Let's download them - they include a private key that we can use to SSH
into the cluster.

```bash
./contrib/s3-config pull $CLUSTER klondike-cluster-$CLUSTER
```

Output:

```bash
download: s3://klondike-cluster-cascade/klondike-cluster-cascade.tar.gz to clusters/klondike-cluster-cascade.tar.gz
Pulled clusters/klondike-cluster-cascade.tar.gz from s3://klondike-cluster-cascade/klondike-cluster-cascade.tar.gz
Unpacked clusters/cascade/ from clusters/klondike-cluster-cascade.tar.gz
```

At this point, we'll have a `private key` and `kubeconfig` inside of a local
`clusters` folder.

Lets finally ssh into that bastion! Substitute accordingly using the output
you got from earlier commands:

```bash
ssh -i ./clusters/cascade/id_rsa ubuntu@***.***.***.***
```

The Bastion is a Lonely Place, Find me a Controller
---------------------------------------------------

First, lets find the ip address of a `node` in our cluster. *From the bastion:*

```bash
# From ubuntu@ip-10-0-255-209
kubectl get nodes
```

Output:

```
NAME                                        STATUS                     AGE
ip-10-0-255-4.us-west-2.compute.internal    Ready,SchedulingDisabled   23d
ip-10-0-255-5.us-west-2.compute.internal    Ready,SchedulingDisabled   23d
ip-10-0-255-6.us-west-2.compute.internal    Ready,SchedulingDisabled   23d
ip-10-0-99-176.us-west-2.compute.internal   Ready                      23d
```

You'll notice that 3 of the 4 nodes indicate `SchedulingDisabled`. This means
that Kubernetes will not try to automatically schedule pods to the first three
hosts in that list. These happen to be the nodes in the Kubernetes control
plane.

Lets SSH into a control plane node. These share the network we'll be using
for our ingress controllers and it would be useful to assert that a pod
is routable.

*Dare we SSH?* We dare.

```bash
# From ubuntu@ip-10-0-255-209
ssh core@ip-10-0-255-4.us-west-2.compute.internal -i klondike/clusters/cascade/id_rsa
```

The bastion is provisioned with the `klondike/clusters/cascade` folder in
the home directory. It's contents should be identical to those in your
local clusters folders after using `./contrib/s3-pull`.

Check Pod Connectivity
----------------------

One of our pods is running nginx, and should be listening on port 80 so we
can do a quick and dirty check using curl to see if we're routable. First
we'll need to figure out some pod ips.

*From your local system, execute:*

```bash
kubectl --kubeconfig clusters/cascade/kubeconfig describe pods
```

Output:

```
...
Labels:       app=nginx
Status:       Running
Reason:
Message:
IP:       10.1.2.6
...
```

You're output won't match the above exactly, but as long as there's some
service listening on port 80, you should be able to curl it:

```bash
# From core@ip-10-0-255-4
curl 10.1.2.6
```

Output:

```
<!DOCTYPE html>
<html>
<head>
<title>Welcome to nginx!</title>
<style>
....
```
