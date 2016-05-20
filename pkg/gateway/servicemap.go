package gateway

import (
	"fmt"
)

type ServiceMapper interface {
	ServiceMap() (*ServiceMap, error)
}

type ServiceMap struct {
	ServiceGroups []ServiceGroup
}

type Service struct {
	Namespace  string
	Name       string
	TargetPort int
	Endpoints  []Endpoint
	Path       string
}

type ServiceGroup struct {
	Name      string
	Namespace string
	Aliases   []string
	Services  []Service
}

func (svg *ServiceGroup) DefaultServerName(cz string) string {
	return fmt.Sprintf(
		"%s.%s.%s",
		svg.Name,
		svg.Namespace,
		cz,
	)
}

type Endpoint struct {
	Name string
	IP   string
	Port int
}
