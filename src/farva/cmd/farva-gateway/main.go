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
package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/bcwaldon/klondike/src/farva/pkg/flagutil"
	"github.com/bcwaldon/klondike/src/farva/pkg/gateway"
)

func main() {
	fs := flag.NewFlagSet("farva-gateway", flag.ExitOnError)

	var cfg gateway.Config
	fs.DurationVar(&cfg.RefreshInterval, "refresh-interval", 30*time.Second, "Attempt to build and reload a new nginx config at this interval")
	fs.StringVar(&cfg.KubeconfigFile, "kubeconfig", "", "Set this to provide an explicit path to a kubeconfig, otherwise the in-cluster config will be used.")
	fs.BoolVar(&cfg.NGINXDryRun, "nginx-dry-run", false, "Log nginx management commands rather than executing them.")
	fs.IntVar(&cfg.NGINXHealthPort, "nginx-health-port", gateway.DefaultNGINXConfig.HealthPort, "Port to listen on for nginx health checks.")
	fs.IntVar(&cfg.FarvaHealthPort, "farva-health-port", gateway.DefaultConfig.FarvaHealthPort, "Port to listen on for farva health checks.")
	fs.IntVar(&cfg.HTTPListenPort, "http-listen-port", gateway.DefaultConfig.HTTPListenPort, "Port to listen on for HTTP traffic.")
	fs.StringVar(&cfg.FifoPath, "fifo-path", gateway.DefaultConfig.FifoPath, "Location of nginx stderr and stdout logging fifo.")
	fs.StringVar(&cfg.ClusterZone, "cluster-zone", "", "Use this DNS zone for routing of traffic to Kubernetes")
	fs.StringVar(&cfg.AnnotationPrefix, "annotation-prefix", gateway.DefaultKubernetesReverseProxyConfigGetterConfig.AnnotationPrefix, "Forms the lookup key for additional gateway configuration annotations.")

	fs.Parse(os.Args[1:])

	if err := flagutil.SetFlagsFromEnv(fs, "FARVA_GATEWAY"); err != nil {
		log.Fatalf("Failed setting flags from env: %v", err)
	}

	gw, err := gateway.New(cfg)
	if err != nil {
		log.Fatalf("Gateway construction failed: %v", err)
	}

	if err := gw.Run(); err != nil {
		log.Printf("Gateway operation failed: %v", err)
	}

	log.Printf("Gateway shutting down")
}
