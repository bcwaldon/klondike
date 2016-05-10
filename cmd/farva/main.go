package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"text/template"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/restclient"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	clientcmd "k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
	clientcmdapi "k8s.io/kubernetes/pkg/client/unversioned/clientcmd/api"
)

var (
	nginxTemplateData = `
http {
{{ range $svc := .Services }}
    server {{ $svc.Namespace }}__{{ $svc.Name }} {
        listen {{ $svc.ListenPort }};
    }
    upstream {{ $svc.Namespace }}__{{ $svc.Name }} {
    {{ range $name, $ep := $svc.Endpoints }}
        server {{ $ep }};  # {{ $name }}
    {{- end }}
    }
{{ end }}
}
`

	nginxTemplate = template.Must(template.New("nginx").Parse(nginxTemplateData))
)

func main() {
	fs := flag.NewFlagSet("farva", flag.ExitOnError)
	kubeconfig := fs.String("kubeconfig", "", "Set this to provide an explicit path to a kubeconfig, otherwise the in-cluster config will be used.")
	fs.Parse(os.Args[1:])

	kc, err := NewKubernetesClient(*kubeconfig)
	if err != nil {
		log.Fatalf("Failed building Kubernetes client: %v", err)
	}

	services, err := getServices(kc)
	if err != nil {
		log.Fatalf("Failed getting Kubernetes services: %v", err)
	}

	cfg, err := renderConfig(services)
	if err != nil {
		log.Fatalf("Failed rendering config: %v", err)
	}

	fmt.Printf("%s", cfg)
}

type Service struct {
	Namespace  string
	Name       string
	ListenPort int32
	Endpoints  map[string]string
}

func renderConfig(services []Service) ([]byte, error) {
	config := struct {
		Services []Service
	}{
		Services: services,
	}

	var buf bytes.Buffer
	if err := nginxTemplate.Execute(&buf, config); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func getKubernetesClientConfig(kubeconfig string) (*restclient.Config, error) {
	if kubeconfig == "" {
		return restclient.InClusterConfig()
	} else {

		//NOTE(bcwaldon): must set this or the host will be
		// overridden later when the kubeconfig is loaded
		clientcmd.DefaultCluster = clientcmdapi.Cluster{}

		rules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig}
		overrides := &clientcmd.ConfigOverrides{}
		loader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides)

		return loader.ClientConfig()
	}
}

func NewKubernetesClient(kubeconfig string) (*client.Client, error) {
	cfg, err := getKubernetesClientConfig(kubeconfig)
	if err != nil {
		return nil, err
	}

	return client.New(cfg)
}

func getServices(kc *client.Client) ([]Service, error) {
	ingressList, err := kc.Ingress(api.NamespaceAll).List(api.ListOptions{})
	if err != nil {
		return nil, err
	}

	var services []Service
	for _, ing := range ingressList.Items {
		ingServicePort := int32(ing.Spec.Backend.ServicePort.IntValue())

		svc := Service{
			Name:      ing.Spec.Backend.ServiceName,
			Namespace: ing.ObjectMeta.Namespace,
			Endpoints: map[string]string{},
		}

		service, err := kc.Services(svc.Namespace).Get(svc.Name)
		if err != nil {
			return nil, err
		}

		for _, port := range service.Spec.Ports {
			if port.Port != ingServicePort || port.Protocol != api.ProtocolTCP {
				continue
			}
			svc.ListenPort = port.NodePort
		}

		endpoints, err := kc.Endpoints(svc.Namespace).Get(svc.Name)
		if err != nil {
			return nil, err
		}

		for _, sub := range endpoints.Subsets {
			if sub.Ports[0].Port != ingServicePort || sub.Ports[0].Protocol != api.ProtocolTCP {
				continue
			}

			for _, addr := range sub.Addresses {
				svc.Endpoints[addr.TargetRef.Name] = addr.IP
			}
		}

		services = append(services, svc)
	}

	return services, nil
}
