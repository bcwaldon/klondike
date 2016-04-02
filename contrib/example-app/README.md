This directory contains a full Kubernetes application, originating from https://github.com/kubernetes/kubernetes/tree/master/examples/k8petstore.

To deploy this application, simply pass all of the JSON files to `kubectl create`.
For example:

```
ls contrib/example-app/*.json | xargs -n1 kubectl create -f
```
