package main

import (
	"flag"
	"log"
	"os"

	"github.com/bcwaldon/farva/pkg/gateway"
)

func main() {
	fs := flag.NewFlagSet("farva", flag.ExitOnError)

	var cfg gateway.Config
	fs.StringVar(&cfg.KubeconfigFile, "kubeconfig", "", "Set this to provide an explicit path to a kubeconfig, otherwise the in-cluster config will be used.")
	fs.BoolVar(&cfg.NGINXDryRun, "nginx-dry-run", false, "Log nginx management commands rather than executing them.")

	fs.Parse(os.Args[1:])

	gw, err := gateway.New(cfg)
	if err != nil {
		log.Fatalf("Failed constructing Gateway: %v", err)
	}

	if err := gw.Start(); err != nil {
		log.Fatalf("Failed starting Gateway: %v", err)
	}

	log.Printf("Started")

	if err := gw.Refresh(); err != nil {
		log.Fatalf("Failed refreshing Gateway: %v", err)
	}

	log.Printf("Refreshed")

	select {}
}
