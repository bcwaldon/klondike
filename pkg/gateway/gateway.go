package gateway

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/bcwaldon/farva/pkg/health"
)

type Config struct {
	RefreshInterval  time.Duration
	KubeconfigFile   string
	ClusterZone      string
	NGINXDryRun      bool
	NGINXHealthPort  int
	FarvaHealthPort  int
	AnnotationPrefix string
}

const DefaultFarvaHealthPort = 7333

func New(cfg Config) (*Gateway, error) {
	kc, err := newKubernetesClient(cfg.KubeconfigFile)
	if err != nil {
		return nil, err
	}

	smcfg := &ServiceMapperConfig{
		AnnotationPrefix: cfg.AnnotationPrefix,
	}
	sm := newServiceMapper(kc, smcfg)

	nginxCfg := newNGINXConfig(cfg.NGINXHealthPort, cfg.ClusterZone)
	var nm NGINXManager
	if cfg.NGINXDryRun {
		nm = newLoggingNGINXManager()
	} else {
		nm = newNGINXManager(nginxCfg)
	}
	log.Printf("Using nginx config: %+v", nginxCfg)

	gw := Gateway{
		cfg: cfg,
		sm:  sm,
		nm:  nm,
	}

	return &gw, nil
}

type Gateway struct {
	cfg Config
	sm  ServiceMapper
	nm  NGINXManager
}

func (gw *Gateway) start() error {
	ok, err := gw.nginxIsRunning()
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

	gw.startHTTPServer()

	return nil
}

func (gw *Gateway) startHTTPServer() {
	s := &http.Server{
		Addr:    fmt.Sprintf(":%d", gw.cfg.FarvaHealthPort),
		Handler: health.NewHandler(),
	}

	go func() {
		log.Fatal(s.ListenAndServe())
	}()
}

func (gw *Gateway) nginxIsRunning() (bool, error) {
	log.Printf("Checking if nginx is running")
	st, err := gw.nm.Status()
	if err != nil {
		return false, err
	}
	return st == nginxStatusRunning, nil
}

func (gw *Gateway) refresh() error {
	log.Printf("Refreshing nginx config")
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

func (gw *Gateway) Run() error {
	if err := gw.start(); err != nil {
		return err
	}

	log.Printf("Gateway started successfully, entering refresh loop")

	ticker := time.NewTicker(gw.cfg.RefreshInterval)

	for {
		if err := gw.refresh(); err != nil {
			log.Printf("Failed refreshing Gateway: %v", err)
		}

		//NOTE(bcwaldon): receive from the ticker at the
		// end of the loop to emulate do-while semantics.
		<-ticker.C
	}
}
