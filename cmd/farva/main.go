package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/bcwaldon/farva/pkg/gateway"
)

func main() {
	fs := flag.NewFlagSet("farva", flag.ExitOnError)
	kubeconfig := fs.String("kubeconfig", "", "Set this to provide an explicit path to a kubeconfig, otherwise the in-cluster config will be used.")
	fs.Parse(os.Args[1:])

	gw, err := gateway.New(*kubeconfig)
	if err != nil {
		log.Fatalf("Failed constructing Gateway: %v", err)
	}

	cfg, err := gw.Render()
	if err != nil {
		log.Fatalf("Failed rendering config: %v", err)
	}

	fmt.Printf("%s", cfg)
}
