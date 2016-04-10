package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bcwaldon/kube-datadog/pkg/server"
)

func main() {
	var cfg server.Config
	var tags StringSliceFlag

	fs := flag.NewFlagSet("kube-datadog", flag.ExitOnError)
	fs.StringVar(&cfg.KubeletHost, "kubelet", "127.0.0.1:10255", "Address of kubelet stats API")
	fs.DurationVar(&cfg.Period, "period", 10*time.Second, "Amount of time to wait between metric collection attempts")
	fs.StringVar(&cfg.DogStatsDHost, "dogstatsd", "127.0.0.1:8125", "Address of DogStatsD endpoint (UDP)")
	fs.Var(&tags, "tags", "Set of tags to attach to all metrics (i.e. cloud:aws,cluster:prod)")

	fs.Parse(os.Args[1:])

	cfg.Tags = []string(tags)

	s, err := server.New(cfg)
	if err != nil {
		log.Fatalf("Failed initializing server: %v", err)
	}

	log.Printf("Initialized server with config: %+v", cfg)
	log.Printf("Starting metrics collection")
	s.Run()
}

type StringSliceFlag []string

func (f *StringSliceFlag) String() string {
	return fmt.Sprintf("%s", *f)
}

func (f *StringSliceFlag) Set(value string) error {
	*f = append(*f, value)
	return nil
}
