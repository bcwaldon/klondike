package gateway

import (
	"fmt"
	"log"
	"net/http"
)

const HealthPort = 7333

type Config struct {
	KubeconfigFile string
	NGINXDryRun    bool
}

func New(cfg Config) (*Gateway, error) {
	kc, err := newKubernetesClient(cfg.KubeconfigFile)
	if err != nil {
		return nil, err
	}

	sm := newServiceMapper(kc)

	var nm NGINXManager
	if cfg.NGINXDryRun {
		nm = newLoggingNGINXManager()
	} else {
		nm = newNGINXManager()
	}

	gw := Gateway{
		sm: sm,
		nm: nm,
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Healthy")
	})

	return &gw, nil
}

type Gateway struct {
	sm ServiceMapper
	nm NGINXManager
	hs *http.Server
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

	log.Printf("Started health server on 7333")
	log.Fatal(http.ListenAndServe(":7333", nil))

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

	if err := gw.nm.WriteConfig(sm); err != nil {
		return err
	}
	if err := gw.nm.Reload(); err != nil {
		return err
	}

	return nil
}
