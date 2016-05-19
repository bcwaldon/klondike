package gateway

import (
	kapi "k8s.io/kubernetes/pkg/api"
	krestclient "k8s.io/kubernetes/pkg/client/restclient"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
	kclientcmd "k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
	kclientcmdapi "k8s.io/kubernetes/pkg/client/unversioned/clientcmd/api"
)

func newKubernetesClient(kubeconfig string) (*kclient.Client, error) {
	cfg, err := getKubernetesClientConfig(kubeconfig)
	if err != nil {
		return nil, err
	}

	return kclient.New(cfg)
}

func getKubernetesClientConfig(kubeconfig string) (*krestclient.Config, error) {
	if kubeconfig == "" {
		return krestclient.InClusterConfig()
	} else {

		//NOTE(bcwaldon): must set this or the host will be
		// overridden later when the kubeconfig is loaded
		kclientcmd.DefaultCluster = kclientcmdapi.Cluster{}

		rules := &kclientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig}
		overrides := &kclientcmd.ConfigOverrides{}
		loader := kclientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides)

		return loader.ClientConfig()
	}
}

func newServiceMapper(kc *kclient.Client) ServiceMapper {
	return &apiServiceMapper{kc: kc}
}

type apiServiceMapper struct {
	kc *kclient.Client
}

func (asm *apiServiceMapper) ServiceMap() (*ServiceMap, error) {
	ingressList, err := asm.kc.Ingress(kapi.NamespaceAll).List(kapi.ListOptions{})
	if err != nil {
		return nil, err
	}

	var services []Service
	for _, ing := range ingressList.Items {
		ingServicePort := ing.Spec.Backend.ServicePort.IntValue()

		svc := Service{
			Name:      ing.Spec.Backend.ServiceName,
			Namespace: ing.ObjectMeta.Namespace,
			Endpoints: []Endpoint{},
		}

		service, err := asm.kc.Services(svc.Namespace).Get(svc.Name)
		if err != nil {
			return nil, err
		}

		for _, port := range service.Spec.Ports {
			if port.Port != ingServicePort || port.Protocol != kapi.ProtocolTCP {
				continue
			} else {
				svc.ListenPort = port.NodePort
				svc.TargetPort = port.TargetPort.IntValue()
				break
			}
		}

		endpoints, err := asm.kc.Endpoints(svc.Namespace).Get(svc.Name)
		if err != nil {
			return nil, err
		}

		for _, sub := range endpoints.Subsets {
			if sub.Ports[0].Port != svc.TargetPort || sub.Ports[0].Protocol != kapi.ProtocolTCP {
				continue
			}

			//NOTE(bcwaldon): addresses may not be guaranteed to be in the same
			// order every time we make this API call. Probably want to sort.
			for _, addr := range sub.Addresses {
				ep := Endpoint{
					Name: addr.TargetRef.Name,
					IP:   addr.IP,
					Port: sub.Ports[0].Port,
				}
				svc.Endpoints = append(svc.Endpoints, ep)
			}
		}

		services = append(services, svc)
	}

	sm := &ServiceMap{Services: services}
	return sm, nil
}
