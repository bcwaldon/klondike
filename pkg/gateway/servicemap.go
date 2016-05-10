package gateway

type ServiceMapper interface {
	ServiceMap() (*ServiceMap, error)
}

type ServiceMap struct {
	Services []Service
}

type Service struct {
	Namespace  string
	Name       string
	ListenPort int32
	Endpoints  map[string]string
}
