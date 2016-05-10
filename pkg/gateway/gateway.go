package gateway

func New(kubeconfig string) (*Gateway, error) {
	kc, err := newKubernetesClient(kubeconfig)
	if err != nil {
		return nil, err
	}

	sm := newServiceMapper(kc)
	return &Gateway{sm: sm}, nil
}

type Gateway struct {
	sm ServiceMapper
}

func (gw *Gateway) Render() ([]byte, error) {
	sm, err := gw.sm.ServiceMap()
	if err != nil {
		return nil, err
	}

	cfg, err := renderNginxConfig(sm)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
