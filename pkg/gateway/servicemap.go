package gateway

import (
	"strings"
)

type ServiceMapper interface {
	ServiceMap() (*ServiceMap, error)
}

type ServiceMap struct {
	HTTPServiceGroups []*HTTPServiceGroup
	TCPServices       []*TCPService
}

type ServiceGroup interface {
	Namespace() string
	Name() string
}

type HTTPService struct {
	Namespace  string
	Name       string
	TargetPort int
	Endpoints  []TCPEndpoint
	Path       string
}

type HTTPServiceGroup struct {
	name      string
	namespace string
	Aliases   []string
	Services  []HTTPService
}

func (sg *HTTPServiceGroup) Namespace() string {
	return sg.namespace
}

func (sg *HTTPServiceGroup) Name() string {
	return sg.name
}

type TCPService struct {
	name       string
	namespace  string
	ListenPort int
	Endpoints  []TCPEndpoint
}

func (sg *TCPService) Namespace() string {
	return sg.namespace
}

func (sg *TCPService) Name() string {
	return sg.name
}

func CanonicalHostname(sg ServiceGroup, clusterZone string) string {
	return strings.Join([]string{sg.Name(), sg.Namespace(), clusterZone}, ".")
}

type TCPEndpoint struct {
	Name string
	IP   string
	Port int
}
