package gateway

import (
	"fmt"
	kapi "k8s.io/kubernetes/pkg/api"
	kextensions "k8s.io/kubernetes/pkg/apis/extensions"
	krestclient "k8s.io/kubernetes/pkg/client/restclient"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
	kclientcmd "k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
	kclientcmdapi "k8s.io/kubernetes/pkg/client/unversioned/clientcmd/api"
	"strings"
)

type ServiceMapperConfig struct {
	AnnotationPrefix string
}

const HostnameAliasKey = "hostname-aliases"

func (smcfg *ServiceMapperConfig) annotationKey(name string) string {
	return fmt.Sprintf("%s/%s", smcfg.AnnotationPrefix, name)
}

// Gets a list of strings at a given annotation field.
func (smcfg *ServiceMapperConfig) getAnnotationStringList(ing *kextensions.Ingress, name string) []string {
	anno := ing.ObjectMeta.GetAnnotations()
	result := make([]string, 0)
	annotationKey := smcfg.annotationKey(name)
	for key, val := range anno {
		if key == annotationKey {
			result = splitCSV(val)
		}
	}
	return result
}

var DefaultServiceMapperConfig = ServiceMapperConfig{
	AnnotationPrefix: "klondike.gateway",
}

func splitCSV(csv string) []string {
	parts := strings.Split(csv, ",")
	trimmed := []string{}
	for _, part := range parts {
		trimmed = append(trimmed, strings.TrimSpace(part))
	}
	return trimmed
}

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

func newServiceMapper(kc *kclient.Client, smcfg *ServiceMapperConfig) ServiceMapper {
	return &apiServiceMapper{kc: kc, smcfg: smcfg}
}

type apiServiceMapper struct {
	kc    *kclient.Client
	smcfg *ServiceMapperConfig
}

func setServicePorts(svc *Service, ingServicePort int, asm *apiServiceMapper) error {
	service, err := asm.kc.Services(svc.Namespace).Get(svc.Name)

	if err != nil {
		return err
	}

	for _, port := range service.Spec.Ports {
		if port.Port != ingServicePort || port.Protocol != kapi.ProtocolTCP {
			continue
		} else {
			svc.TargetPort = port.TargetPort.IntValue()
			break
		}
	}

	return nil
}

func setServiceEndpoints(svc *Service, asm *apiServiceMapper) error {
	endpoints, err := asm.kc.Endpoints(svc.Namespace).Get(svc.Name)
	if err != nil {
		return err
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

	return nil
}

func (asm *apiServiceMapper) ServiceMap() (*ServiceMap, error) {
	ingressList, err := asm.kc.Ingress(kapi.NamespaceAll).List(kapi.ListOptions{})
	if err != nil {
		return nil, err
	}

	var serviceGroups []ServiceGroup
	for _, ing := range ingressList.Items {
		var services []Service
		svg := ServiceGroup{
			Name:      ing.ObjectMeta.Name,
			Namespace: ing.ObjectMeta.Namespace,
			Services:  []Service{},
			Aliases:   asm.smcfg.getAnnotationStringList(&ing, HostnameAliasKey),
		}

		if ing.Spec.Backend != nil {
			// Default backend.
			ingServicePort := ing.Spec.Backend.ServicePort.IntValue()

			svc := Service{
				Name:      ing.Spec.Backend.ServiceName,
				Namespace: ing.ObjectMeta.Namespace,
				Endpoints: []Endpoint{},
			}
			if err := setServicePorts(&svc, ingServicePort, asm); err != nil {
				return nil, err
			}
			if err := setServiceEndpoints(&svc, asm); err != nil {
				return nil, err
			}
			services = append(services, svc)
		} else {
			// Rule-based backend.
			for _, rule := range ing.Spec.Rules {
				// We're explictly ignoring the host here.
				// ingServiceHost := rule.Host
				// Only HTTP is supported currently.
				for _, path := range rule.HTTP.Paths {
					ingServicePath := path.Path
					ingServicePort := path.Backend.ServicePort.IntValue()
					svc := Service{
						Name:      path.Backend.ServiceName,
						Namespace: ing.ObjectMeta.Namespace,
						Endpoints: []Endpoint{},
						Path:      ingServicePath,
					}
					if err := setServicePorts(&svc, ingServicePort, asm); err != nil {
						return nil, err
					}
					if err := setServiceEndpoints(&svc, asm); err != nil {
						return nil, err
					}
					services = append(services, svc)
				}
			}
		}
		svg.Services = services
		serviceGroups = append(serviceGroups, svg)
	}

	sm := &ServiceMap{ServiceGroups: serviceGroups}
	return sm, nil
}
