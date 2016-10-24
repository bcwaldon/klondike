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

package server

import (
	"log"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/resource"
	"k8s.io/kubernetes/pkg/client/restclient"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	clientcmd "k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
	clientcmdapi "k8s.io/kubernetes/pkg/client/unversioned/clientcmd/api"
	"k8s.io/kubernetes/pkg/fields"
)

var trackedResources = []api.ResourceName{
	api.ResourceCPU,
	api.ResourceMemory,
}

func updateResourceList(dest, src api.ResourceList) {
	for _, res := range trackedResources {
		qty, _ := dest[res]
		qty.Add(src[res])
		dest[res] = qty
	}
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

func NewAPIEmitter(kubeconfig string) (Emitter, error) {
	cfg, err := getKubernetesClientConfig(kubeconfig)
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
	nodeList, err := e.client.Nodes().List(api.ListOptions{})
	if err != nil {
		log.Printf("failed to get nodes: %v", err)
		return
	}
	ch <- Metric{Name: "kubernetes.node.all", Value: len(nodeList.Items)}

	capacity := make(map[api.ResourceName]resource.Quantity)
	allocatable := make(map[api.ResourceName]resource.Quantity)
	scheduled := make(map[api.ResourceName]resource.Quantity)

	for _, n := range nodeList.Items {
		updateResourceList(capacity, n.Status.Capacity)
		updateResourceList(allocatable, n.Status.Allocatable)

		podList, err := e.client.Pods(api.NamespaceAll).List(api.ListOptions{
			FieldSelector: fields.OneTermEqualSelector("spec.nodeName", n.Name),
		})
		if err != nil {
			log.Printf("failed to get pods for node %v: %v", n.Name, err)
			return
		}
		for _, p := range podList.Items {
			for _, c := range p.Spec.Containers {
				updateResourceList(scheduled, c.Resources.Requests)
			}
		}
	}

	capMem := capacity[api.ResourceMemory]
	capCPU := capacity[api.ResourceCPU]
	ch <- Metric{
		Name:  "kubernetes.resource.cpu.capacity",
		Value: int(capCPU.MilliValue()),
	}
	ch <- Metric{
		Name:  "kubernetes.resource.memory.capacity",
		Value: int(capMem.Value()),
	}

	allocMem := allocatable[api.ResourceMemory]
	allocCPU := allocatable[api.ResourceCPU]
	ch <- Metric{
		Name:  "kubernetes.resource.cpu.allocatable",
		Value: int(allocCPU.MilliValue()),
	}
	ch <- Metric{
		Name:  "kubernetes.resource.memory.allocatable",
		Value: int(allocMem.Value()),
	}

	schedMem := scheduled[api.ResourceMemory]
	schedCPU := scheduled[api.ResourceCPU]
	ch <- Metric{
		Name:  "kubernetes.resource.cpu.scheduled",
		Value: int(schedCPU.MilliValue()),
	}
	ch <- Metric{
		Name:  "kubernetes.resource.memory.scheduled",
		Value: int(schedMem.Value()),
	}

	namespaceList, err := e.client.Namespaces().List(api.ListOptions{})
	if err != nil {
		log.Printf("failed to get namespaces: %v", err)
		return
	}

	ch <- Metric{Name: "kubernetes.namespace.all", Value: len(namespaceList.Items)}
}
