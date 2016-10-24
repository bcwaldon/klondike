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
	"net/http"
	"time"

	"github.com/bcwaldon/klondike/src/farva/pkg/health"
	"github.com/bcwaldon/klondike/src/farva/pkg/logger"
	"github.com/bcwaldon/klondike/src/farva/pkg/logpipe"
)

type Config struct {
	RefreshInterval  time.Duration
	KubeconfigFile   string
	ClusterZone      string
	NGINXDryRun      bool
	NGINXHealthPort  int
	HTTPListenPort   int
	FarvaHealthPort  int
	AnnotationPrefix string
	FifoPath         string
}

var DefaultConfig = Config{
	HTTPListenPort:  7331,
	FarvaHealthPort: 7333,
	FifoPath:        "/nginx.fifo",
}

func DefaultHTTPReverseProxyServers(cfg *Config) []httpReverseProxyServer {
	return []httpReverseProxyServer{
		httpReverseProxyServer{
			ListenPort: cfg.NGINXHealthPort,
			Locations: []httpReverseProxyLocation{
				httpReverseProxyLocation{
					Path:          "/health",
					StaticCode:    200,
					StaticMessage: "Healthy!",
				},
			},
		},
		httpReverseProxyServer{
			ListenPort:    cfg.HTTPListenPort,
			DefaultServer: true,
			StaticCode:    444,
		},
	}
}

func DefaultReverseProxyConfig(cfg *Config) *reverseProxyConfig {
	return &reverseProxyConfig{
		HTTPServers: DefaultHTTPReverseProxyServers(cfg),
	}
}

func New(cfg Config) (*Gateway, error) {
	kc, err := newKubernetesClient(cfg.KubeconfigFile)
	if err != nil {
		return nil, err
	}

	krc := &kubernetesReverseProxyConfigGetterConfig{
		AnnotationPrefix: cfg.AnnotationPrefix,
		ClusterZone:      cfg.ClusterZone,
		ListenPort:       cfg.HTTPListenPort,
	}
	rg := newReverseProxyConfigGetter(kc, krc)

	nginxCfg := newNGINXConfig(cfg.NGINXHealthPort, cfg.ClusterZone, cfg.FifoPath, cfg.FifoPath)
	var nm NGINXManager
	if cfg.NGINXDryRun {
		nm = newLoggingNGINXManager()
	} else {
		nm = newNGINXManager(nginxCfg)
	}
	logger.Log.Infof("Using nginx config: %+v", nginxCfg)

	gw := Gateway{
		cfg: cfg,
		rg:  rg,
		nm:  nm,
	}

	return &gw, nil
}

type Gateway struct {
	cfg Config
	rg  ReverseProxyConfigGetter
	nm  NGINXManager
}

func (gw *Gateway) start() error {
	ok, err := gw.nginxIsRunning()
	if err != nil {
		return err
	} else if ok {
		return nil
	}

	rc := DefaultReverseProxyConfig(&gw.cfg)
	if err := gw.nm.SetConfig(rc); err != nil {
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
		logger.Log.Fatal(s.ListenAndServe())
	}()
}

func (gw *Gateway) nginxIsRunning() (bool, error) {
	logger.Log.Info("Checking if nginx is running")
	st, err := gw.nm.Status()
	if err != nil {
		return false, err
	}
	return st == nginxStatusRunning, nil
}

func (gw *Gateway) refresh() error {
	logger.Log.Info("Refreshing nginx config")
	rc, err := gw.rg.ReverseProxyConfig()
	if err != nil {
		return err
	}

	rc.HTTPServers = append(rc.HTTPServers, DefaultHTTPReverseProxyServers(&gw.cfg)...)

	if err := gw.nm.SetConfig(rc); err != nil {
		return err
	}
	return nil
}

func (gw *Gateway) Run() error {

	fifoLogger := logpipe.NewLogPipe(gw.cfg.FifoPath)
	if err := fifoLogger.Start(); err != nil {
		logger.Log.Fatalf(
			"Could not start fifo logger, exiting since this will cause NGINX to block: %s",
			err,
		)
	}

	if err := gw.start(); err != nil {
		return err
	}

	logger.Log.Info("Gateway started successfully, entering refresh loop")

	ticker := time.NewTicker(gw.cfg.RefreshInterval)

	for {
		if err := gw.refresh(); err != nil {
			logger.Log.Infof("Failed refreshing Gateway: %v", err)
		}

		//NOTE(bcwaldon): receive from the ticker at the
		// end of the loop to emulate do-while semantics.
		<-ticker.C
	}
}
