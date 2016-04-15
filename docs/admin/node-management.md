# klondike Node Management

## Adding Workers

To increase the number of workers, we simply need to increase the desired capacity of the worker AutoScaling Group and provision the new instances using ansible.
Start by fetching the current cluster state, identifying the fields `worker_asg_name`, `worker_asg_capacity` and `bastion_public_ip`:

```
python contrib/cluster-info.py $CLUSTER
```

Use the `aws` CLI tool to set the desired capacity to a higher value than was reported via `worker_asg_capacity`.
This can also be done via the AWS Web Console:

```
aws autoscaling set-desired-capacity --auto-scaling-group-name <worker_asg_name> --desired-capacity <worker_asg_capacity + N>
```

Once the new instances have booted, SSH to the bastion host:

```
ssh -i clusters/$CLUSTER/id_rsa ubuntu@<bastion_public_ip>
```

Run the `site.yml` playbook from `/home/ubuntu/klondike/`:

```
cd /home/ubuntu/klondike/
ansible-playbook -e cluster=$CLUSTER site.yml -l tag_group_worker
```

After ansible successfully provisions the new workers, the kubelet is configured to automatically self-register with the Kubernetes cluster.
Verify the workers have joined successfully using `kubectl`:

```
kubectl get nodes
```

The number of schedulable nodes (those not reporting `SchedulingDisabled`) should match the new desired capacity of the group.

## Removing Workers

To decrease the number of workers, we simply need to decrease the desired capacity of the worker AutoScaling Group.
Start by fetching the current cluster state, identifying the fields `worker_asg_name` and `worker_asg_capacity`:

```
python contrib/cluster-info.py $CLUSTER
```

Use the `aws` CLI tool to set the desired capacity to a lower value than was reported via `worker_asg_capacity`.
This can also be done via the AWS Web Console:

```
aws autoscaling set-desired-capacity --auto-scaling-group-name <worker_asg_name> --desired-capacity <worker_asg_capacity - N>
```

As EC2 instances are terminated, Kubernetes will automatically remove the nodes from the cluster.
Verify the workers have been removed successfully using `kubectl`:

```
kubectl get nodes
```

The number of schedulable nodes (those not reporting `SchedulingDisabled`) should match the new desired capacity of the group.