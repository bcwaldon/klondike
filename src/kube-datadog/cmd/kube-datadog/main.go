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
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/bcwaldon/klondike/src/kube-datadog/pkg/server"
)

func main() {
	fs := flag.NewFlagSet("kube-datadog", flag.ExitOnError)

	var cfg server.Config
	fs.DurationVar(&cfg.Period, "period", 10*time.Second, "Amount of time to wait between metric collection attempts")
	fs.StringVar(&cfg.DogStatsDHost, "dogstatsd", "127.0.0.1:8125", "Address of DogStatsD endpoint (UDP)")
	fs.BoolVar(&cfg.NoPublish, "no-publish", false, "Log metrics instead of publishing them to DogStatsD")

	source := fs.String("source", "", "Pull metrics from one of two sources: kubelet or api")
	kubelet := fs.String("kubelet", "127.0.0.1:10255", "If --mode=kubelet, address of kubelet stats API")
	kubeconfig := fs.String("kubeconfig", "", "If --mode=api, set this to provide an explicit path to a kubeconfig, otherwise the in-cluster config will be used.")

	var tags StringSliceFlag
	fs.Var(&tags, "tags", "Set of tags to attach to all metrics (i.e. cloud:aws,cluster:prod)")

	fs.Parse(os.Args[1:])

	if err := SetFlagsFromEnv(fs, "KUBE_DATADOG"); err != nil {
		log.Fatalf("Failed setting flags from env: %v", err)
	}

	var err error
	if *source == "kubelet" {
		cfg.Emitter = server.NewKubeletEmitter(*kubelet)
	} else if *source == "api" {
		cfg.Emitter, err = server.NewAPIEmitter(*kubeconfig)
	} else {
		err = fmt.Errorf("invalid source %q", *source)
	}
	if err != nil {
		log.Fatalf("Failed constructing emitter: %v", err)
	}

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
	vals := strings.Split(value, ",")
	*f = append(*f, vals...)
	return nil
}

func SetFlagsFromEnv(fs *flag.FlagSet, prefix string) error {
	var err error
	fs.VisitAll(func(f *flag.Flag) {
		key := prefix + "_" + strings.ToUpper(strings.Replace(f.Name, "-", "_", -1))
		if val := os.Getenv(key); val != "" {
			if serr := fs.Set(f.Name, val); serr != nil {
				err = fmt.Errorf("invalid value %q for %s: %v", val, key, serr)
			}
		}
	})
	return err
}
