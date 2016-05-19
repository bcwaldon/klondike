# Updating and Working with Farva

Getting Inside of Farva's Container
-----------------------------------

Farva runs in a special unschedulable node, ala kubernetes control plane. If
you run:

    kubectl get nodes

You'll see something like:

    NAME                                         STATUS                     AGE
    ip-10-157-2-133.us-west-2.compute.internal   Ready,SchedulingDisabled   1d
    ip-10-157-2-138.us-west-2.compute.internal   Ready,SchedulingDisabled   1d
    ip-10-157-2-139.us-west-2.compute.internal   Ready,SchedulingDisabled   1d
    ip-10-157-2-140.us-west-2.compute.internal   Ready,SchedulingDisabled   1d
    ip-10-157-3-138.us-west-2.compute.internal   Ready                      1d
    ip-10-157-3-139.us-west-2.compute.internal   Ready                      1d
    ip-10-157-3-140.us-west-2.compute.internal   Ready                      1d

Curious. Which one of these nodes is running farva? You might try:

    kubectl get pods

Yielding:

    NAME             READY     STATUS    RESTARTS   AGE
    ...
    some_app-3tyto   1/1       Running   0          8h
    ...

Or:

    kubectl get services

Yielding:

    NAME             CLUSTER_IP      EXTERNAL_IP   PORT(S)    SELECTOR      AGE
    ...
    kubernetes       10.255.0.1      <none>        443/TCP    <none>        1d
    some_app         10.255.156.49   nodes         8085/TCP   app=some_app  8h
    ...

The trick is that farva runs in the `kube-system` namespace. Lets try:

    kubectl get pods --namespace=kube-system

Which yields:

    NAME                                                                 READY     STATUS    RESTARTS   AGE
    ...
    datadog-agent-ip-10-157-2-133.us-west-2.compute.internal             2/2       Running   0          1d
    etcd-ip-10-157-2-140.us-west-2.compute.internal                      1/1       Running   0          1d
    gateway-ip-10-157-2-133.us-west-2.compute.internal                   1/1       Running   2          2h
    ...

There it is! The `gateway-ip-10-157` pod is our farva pod. The `gateway` suffix
comes form the service description configured in the [kube
gateway](../roles/kube-gateway/vars/main.yml#L5) role.


If you then:

    kubectl describe gateway-ip-10-157-2-133.us-west-2.compute.internal --namespace=kube-system

You'll see:

    Name:				gateway-ip-10-157-2-133.us-west-2.compute.internal
    ...
    IP:				10.157.2.133
    Replication Controllers:	<none>
    Containers:
      farva:
        Container ID:		docker://91a2808c021b02187cc1c1ba73eca4427c0a0fe4a33c962899b3dc1d36dd0b73
        Image:			quay.io/bcwaldon/farva:65e333067dfa5778e28217344ee12c36bb688741
        Image ID:			docker://sha256:5ad18d6a715486fc742bbf207c58bf6f4ed2e3b1ba3e3e2718ae3d5684689d9e
        ...

See the [adventure guide](./adventure-guide.md) for instructions on getting
to the bastion. From there we can:

    ssh core@10.157.2.133 -i "klondike/clusters/$CLUSTER/id_rsa"

At this point we're in the vm that hosts, among others, our farva pod. Lets
see the docker containers we have running:

    docker ps

Yielding:

    CONTAINER ID        IMAGE                                                                    COMMAND
    91a2808c021b        quay.io/bcwaldon/farva:65e333067dfa5778e28217344ee12c36bb688741          "./farva-gateway"

Now, lets find the PID of the running container using the CONTAINER ID just obtained:

    docker inspect 91a2808c021b

Yielding:

    [
        {
            "Id": "91a2808c021b02187cc1c1ba73eca4427c0a0fe4a33c962899b3dc1d36dd0b73",
            ...
            "Path": "./farva-gateway",
            "State": {
                "Status": "running",
                ...
                "Dead": false,
                "Pid": 8052,
                "ExitCode":
                ...
                }
        }
    ]

Take note of the pid above. We can use the `nsenter` command to enter the
container context:

    nsenter -t 8052 -i -n -p -m

Read more about `nsenter` [here](https://blog.docker.com/tag/nsenter).

Farva writes an nginx.conf periodically. You can peek at it with `cat /etc/nginx/nginx.conf`.
You can also see what processes are running:

    ps ax

Yielding:

    PID TTY      STAT   TIME COMMAND
      1 ?        Ssl    0:01 ./farva-gateway
     13 ?        Ss     0:00 nginx: master process nginx -c /etc/nginx/nginx.conf
    730 ?        S      0:00 -bash
    735 ?        S      0:00 nginx: worker process
    739 ?        R+     0:00 ps ax


Hacking a new Farva Into Place
------------------------------

Lets say we want to confirm that our custom build of farva is behaving as
expected in a staging environment. We don't want to go through the entire
process of spinning up a brand new cluster just yet, rather we'd like to patch
the version that's currently running with a custom build.

If you're branching off this repo, `quay.io` is already configured to
build and tag every commit into a corresponding container that gets pushed and
tagged using the git commit hash that triggered it. For non-contributors, the
important piece is simply rebuilding farva and pushing the resulting container.

The bastion host has a copy of the roles in this repository. The bastion's home
directory should have a `klondike` folder. You can look at the image id farva
uses in the klondike/roles/kube-gateway/vars/main.yml file:

    ...
    farva_image: "quay.io/bcwaldon/farva:65e333067dfa5778e28217344ee12c36bb688741"
    ...

You can replace this with a custom image and re-deploy farva with:

    # From /home/ubuntu/klondike
    ansible-playbook site.yml -l tag_group_gateway -e cluster=$CLUSTER
