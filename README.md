# farva

This is the start to a Kubernetes Ingress Controller.
Assuming you've got a Kubernetes cluster with a single service, you can do this:

```
% go run cmd/farva/main.go --kubeconfig=<KUBECONFIG>

http {

    server default__nginx {
        listen 30190;
    }
    upstream default__nginx {

        server 10.1.2.5;  # nginx-aheok
        server 10.1.2.7;  # nginx-avop9
        server 10.1.2.6;  # nginx-bguu4
    }

}
```
