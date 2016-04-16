package server

import (
	"log"

	"k8s.io/kubernetes/pkg/api"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	clientcmd "k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
	clientcmdapi "k8s.io/kubernetes/pkg/client/unversioned/clientcmd/api"
)

func NewAPIEmitter(kubeconfig string) (Emitter, error) {
	//NOTE(bcwaldon): must set this or the host will be
	// overridden later when the kubeconfig is loaded
	clientcmd.DefaultCluster = clientcmdapi.Cluster{}

	rules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig}
	overrides := &clientcmd.ConfigOverrides{}
	loader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides)

	cfg, err := loader.ClientConfig()
	if err != nil {
		return nil, err
	}

	client, err := client.New(cfg)
	if err != nil {
		return nil, err
	}
	emitter := &apiEmitter{client: client}
	return emitter, nil
}

type apiEmitter struct {
	client *client.Client
}

func (e *apiEmitter) Emit() (MetricsBundle, error) {
	ch := make(chan Metric)
	go func() {
		e.emit(ch)
		close(ch)
	}()

	return MetricsBundle{metrics: ch}, nil
}

func (e *apiEmitter) emit(ch chan<- Metric) {
	resp, err := e.client.Namespaces().List(api.ListOptions{})
	if err != nil {
		log.Printf("failed gathering namespace data: %v", err)
	}

	ch <- Metric{Name: "kubernetes.namespace.total", Value: len(resp.Items)}
}
