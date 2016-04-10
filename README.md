# kube-datadog

This project collects metrics from a Kubernetes kubelet and publishes them to a DataDog statsd endpoint.
The project is built on [wercker][wercker], and publishes docker images to [quay.io][quay].

[![wercker status](https://app.wercker.com/status/ac68c9bbe93ace3c0ca5d1065748c027/m/master "wercker status")](https://app.wercker.com/project/bykey/ac68c9bbe93ace3c0ca5d1065748c027)

[wercker]: https://app.wercker.com/#applications/570abdfe3f1a8913740207e8
[quay]: https://quay.io/repository/bcwaldon/kube-datadog

## metrics

| metric name | metric value |
|-------------|--------------|
| kubelet.pod.all | number of all pods the kubelet is attempting to run |
| kubelet.pod.running | number of pods the kubelet is successfully running |
| kubelet.pod.pending | number of pods the kubelet is not yet running |
| kubelet.pod.memory.reserved | amount of memory in bytes the kubelet has reserved for its pods |
| kubelet.pod.cpu.reserved | amount of CPU in shares the kubelet has reserved for its pods |
