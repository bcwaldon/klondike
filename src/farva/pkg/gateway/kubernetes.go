/*
Copyright 2016 Planet Labs

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package gateway

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/bcwaldon/klondike/src/farva/pkg/logger"
	kapi "k8s.io/kubernetes/pkg/api"
	kextensions "k8s.io/kubernetes/pkg/apis/extensions"
	krestclient "k8s.io/kubernetes/pkg/client/restclient"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
	kclientcmd "k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
	kclientcmdapi "k8s.io/kubernetes/pkg/client/unversioned/clientcmd/api"
	"strings"
)

type kubernetesReverseProxyConfigGetterConfig struct {
	AnnotationPrefix string
	ClusterZone      string
	ListenPort       int
}

const HostnameAliasKey = "hostname-aliases"

func (krc *kubernetesReverseProxyConfigGetterConfig) annotationKey(name string) string {
	return fmt.Sprintf("%s/%s", krc.AnnotationPrefix, name)
}

// Gets a list of strings at a given annotation field.
func (krc *kubernetesReverseProxyConfigGetterConfig) getAnnotationStringList(ing *kextensions.Ingress, name string) []string {
	anno := ing.ObjectMeta.GetAnnotations()
	result := make([]string, 0)
	annotationKey := krc.annotationKey(name)
	for key, val := range anno {
		if key == annotationKey {
			result = splitCSV(val)
		}
	}
	return result
}

var DefaultKubernetesReverseProxyConfigGetterConfig = kubernetesReverseProxyConfigGetterConfig{
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

func newReverseProxyConfigGetter(kc *kclient.Client, krc *kubernetesReverseProxyConfigGetterConfig) ReverseProxyConfigGetter {
	return &kubernetesReverseProxyConfigGetter{
		kc:  kc,
		krc: krc,
	}
}

type kubernetesReverseProxyConfigGetter struct {
	kc  *kclient.Client
	krc *kubernetesReverseProxyConfigGetterConfig
}

func (rcg *kubernetesReverseProxyConfigGetter) getServiceTargetPort(svcNamespace, svcName string, svcPort int) (int, error) {
	svc, err := rcg.kc.Services(svcNamespace).Get(svcName)
	if err != nil {
		return 0, err
	}

	for _, port := range svc.Spec.Ports {
		if port.Port != svcPort || port.Protocol != kapi.ProtocolTCP {
			continue
		} else {
			return port.TargetPort.IntValue(), nil
		}
	}

	return 0, fmt.Errorf("coult not find port matching %d for service %s in namespace %s", svcPort, svcName, svcNamespace)
}

func (rcg *kubernetesReverseProxyConfigGetter) getServiceEndpoints(svcNamespace, svcName string, svcTargetPort int) ([]reverseProxyUpstreamServer, error) {
	endpoints, err := rcg.kc.Endpoints(svcNamespace).Get(svcName)
	if err != nil {
		return nil, err
	}

	ups := []reverseProxyUpstreamServer{}

	for _, sub := range endpoints.Subsets {
		if sub.Ports[0].Port != svcTargetPort || sub.Ports[0].Protocol != kapi.ProtocolTCP {
			logger.Log.WithFields(logrus.Fields{
				"SourcePort":            sub.Ports[0].Port,
				"SourceProtocol":        sub.Ports[0].Protocol,
				"ServiceTargetPort":     svcTargetPort,
				"ServiceTargetProtocol": kapi.ProtocolTCP,
			}).Info("Ignoring endpoint")
			continue
		}

		//NOTE(bcwaldon): addresses may not be guaranteed to be in the same
		// order every time we make this API call. Probably want to sort.
		for _, addr := range sub.Addresses {
			up := reverseProxyUpstreamServer{
				Name: addr.TargetRef.Name,
				Host: addr.IP,
				Port: sub.Ports[0].Port,
			}
			logger.Log.WithFields(logrus.Fields{
				"Name": addr.TargetRef.Name,
				"Host": addr.IP,
				"Port": sub.Ports[0].Port,
			}).Info("Adding upstream")
			ups = append(ups, up)
		}
	}

	return ups, nil
}

func (rcg *kubernetesReverseProxyConfigGetter) ReverseProxyConfig() (*reverseProxyConfig, error) {
	rp := reverseProxyConfig{}

	ingressList, err := rcg.kc.Ingress(kapi.NamespaceAll).List(kapi.ListOptions{})
	if err != nil {
		return nil, err
	}

	// NOTE(bcwaldon): treat Ingress objects w/o rules as HTTP for now. This will
	// eventually be treated as a TCP-only service.
	for _, ing := range ingressList.Items {
		if ing.Spec.Backend != nil {
			ing.Spec.Rules = []kextensions.IngressRule{
				kextensions.IngressRule{
					IngressRuleValue: kextensions.IngressRuleValue{
						HTTP: &kextensions.HTTPIngressRuleValue{
							Paths: []kextensions.HTTPIngressPath{
								kextensions.HTTPIngressPath{
									Path:    "/",
									Backend: *ing.Spec.Backend,
								},
							},
						},
					},
				},
			}
		}

		if err := rcg.addHTTPIngressToReverseProxyConfig(&rp, &ing); err != nil {
			return nil, fmt.Errorf("failed building reverse proxy config for ingress %s in namespace %s: %v", ing.ObjectMeta.Name, ing.ObjectMeta.Name, err)
		}
	}

	return &rp, nil
}

func (rcg *kubernetesReverseProxyConfigGetter) addHTTPIngressToReverseProxyConfig(rp *reverseProxyConfig, ing *kextensions.Ingress) error {
	ingNamespace := ing.ObjectMeta.Namespace
	ingName := ing.ObjectMeta.Name

	for _, rule := range ing.Spec.Rules {
		srv := httpReverseProxyServer{
			Name:       CanonicalHostname(ingName, ingNamespace, rcg.krc.ClusterZone),
			AltNames:   rcg.krc.getAnnotationStringList(ing, HostnameAliasKey),
			ListenPort: rcg.krc.ListenPort,
			Locations:  []httpReverseProxyLocation{},
		}

		logger.Log.WithFields(logrus.Fields{
			"Name":        srv.Name,
			"AltNames":    srv.AltNames,
			"ListentPort": srv.ListenPort,
		}).Info("Generating new reverse proxy server")

		for _, path := range rule.HTTP.Paths {

			svcName := path.Backend.ServiceName
			svcPort := path.Backend.ServicePort.IntValue()

			up := httpReverseProxyUpstream{
				Name: strings.Join([]string{ingNamespace, ingName, svcName}, "__"),
			}

			svcTargetPort, err := rcg.getServiceTargetPort(ingNamespace, svcName, svcPort)
			if err != nil {
				return err
			}

			up.Servers, err = rcg.getServiceEndpoints(ingNamespace, svcName, svcTargetPort)
			if err != nil {
				return err
			}

			if len(up.Servers) == 0 {
				logger.Log.WithFields(logrus.Fields{
					"svcName":       svcName,
					"ingNamespace":  ingNamespace,
					"svcTargetPort": svcTargetPort,
				}).Infof("No servers found for upstream, using StaticCode for %s", path.Path)
				srv.Locations = append(srv.Locations, httpReverseProxyLocation{
					Path:       path.Path,
					StaticCode: 503,
				})
			} else {
				rp.HTTPUpstreams = append(rp.HTTPUpstreams, up)
				srv.Locations = append(srv.Locations, httpReverseProxyLocation{
					Path:     path.Path,
					Upstream: up.Name,
				})
			}
		}

		rp.HTTPServers = append(rp.HTTPServers, srv)
	}

	return nil
}

func CanonicalHostname(name, namespace, clusterZone string) string {
	return strings.Join([]string{name, namespace, clusterZone}, ".")
}
