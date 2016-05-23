package gateway

import (
	"fmt"
)

type ServiceMapper interface {
	ServiceMap() (*ServiceMap, error)
}

type ServiceMap struct {
	HTTPServiceGroups []HTTPServiceGroup
}

type HTTPService struct {
	Namespace  string
	Name       string
	TargetPort int
	Endpoints  []Endpoint
	Path       string
}

type HTTPServiceGroup struct {
	Aliases   []string
	Name      string
	Namespace string
	Services  []HTTPService
}

func (svg *HTTPServiceGroup) DefaultServerName(cz string) string {
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
