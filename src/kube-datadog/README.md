# kube-datadog

This project collects metrics from a Kubernetes kubelet and publishes them to a DataDog statsd endpoint.
The project is built on [wercker][wercker], and docker images are published to [quay.io][quay].

[![wercker status](https://app.wercker.com/status/ac68c9bbe93ace3c0ca5d1065748c027/m/master "wercker status")](https://app.wercker.com/project/bykey/ac68c9bbe93ace3c0ca5d1065748c027)

[wercker]: https://app.wercker.com/#applications/570abdfe3f1a8913740207e8
[quay]: https://quay.io/repository/bcwaldon/kube-datadog

kube-datadog is licensed under Apache 2.0.
See [LICENSE](LICENSE) for more details.

## metrics

When run with `--source=kubelet`:

| metric name | metric value |
|-------------|--------------|
| kubelet.pod.all | number of all pods the kubelet is attempting to run |
| kubelet.container.all | number of all containers the kubelet is attempting to run |

When run with `--source=api`:

| metric name | metric value |
|-------------|--------------|
| kubernetes.namespace.all | number of namespaces in cluster |
| kubernetes.resource.cpu.capacity | amount of CPU in cluster |
| kubernetes.resource.cpu.allocatable | amount of CPU in cluster minus that reserved for the system components |
| kubernetes.resource.cpu.scheduled | amount of CPU claimed by pod resource requests |
| kubernetes.resource.memory.capacity | amount of memory in cluster |
| kubernetes.resource.memory.allocatable | amount of memory in cluster minus that reserved for the system components |
| kubernetes.resource.memory.scheduled | amount of memory claimed by pod resource requests |
