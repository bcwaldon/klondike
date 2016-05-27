package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/bcwaldon/farva/pkg/flagutil"
	"github.com/bcwaldon/farva/pkg/gateway"
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
