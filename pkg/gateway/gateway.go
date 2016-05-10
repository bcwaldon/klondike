package gateway

import "log"

func New(kubeconfig string) (*Gateway, error) {
	kc, err := newKubernetesClient(kubeconfig)
	if err != nil {
		return nil, err
	}

	sm := newServiceMapper(kc)
	nm := newNGINXManager()

	gw := Gateway{
		sm: sm,
		nm: nm,
	}

	return &gw, nil
}

type Gateway struct {
	sm ServiceMapper
	nm *NGINXManager
}

func (gw *Gateway) Start() error {
	ok, err := gw.isRunning()
	if err != nil {
		return err
	} else if ok {
		return nil
	}

	if err := gw.nm.WriteConfig(&ServiceMap{}); err != nil {
		return err
	}
	if err := gw.nm.Start(); err != nil {
		return err
	}
	return nil
}

func (gw *Gateway) isRunning() (bool, error) {
	st, err := gw.nm.Status()
	if err != nil {
		return false, err
	}
	return st == nginxStatusRunning, nil
}

func (gw *Gateway) Refresh() error {
	sm, err := gw.sm.ServiceMap()
	if err != nil {
		return err
	}

	log.Printf("Using ServiceMap: %+v", sm)

	if err := gw.nm.WriteConfig(sm); err != nil {
		return err
	}
	if err := gw.nm.Reload(); err != nil {
		return err
	}

	return nil
}
