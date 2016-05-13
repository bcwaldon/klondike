package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/bcwaldon/farva/pkg/monitor"
)

func main() {
	fs := flag.NewFlagSet("farva-monitor", flag.ExitOnError)

	cfg := monitor.DefaultConfig

	fs.DurationVar(&cfg.RefreshInterval, "refresh-interval", monitor.DefaultConfig.RefreshInterval, "Attempt to sync AWS entities at this interval")
	fs.IntVar(&cfg.HealthPort, "health-port", monitor.DefaultConfig.HealthPort, "Serve a health endpoint on all interfaces using this port")
	fs.StringVar(&cfg.AWSRegion, "aws-region", monitor.DefaultConfig.AWSRegion, "Required. Scope all API calls to this AWS region")
	fs.StringVar(&cfg.AWSLoadBalancer, "aws-load-balancer", monitor.DefaultConfig.AWSLoadBalancer, "Required. Name of the Elastic Load Balancer to keep in sync")

	var tags KVSliceFlag
	fs.Var(&tags, "aws-instance-tags", "Limit EC2 instances registered in the ELB to those with the provided tags (i.e. KubernetesCluster=prod,group=gateway")

	fs.Parse(os.Args[1:])

	if err := SetFlagsFromEnv(fs, "FARVA_MONITOR"); err != nil {
		log.Fatalf("Failed setting flags from env: %v", err)
	}

	cfg.AWSInstanceTags = map[string]string{}
	for _, pair := range tags {
		cfg.AWSInstanceTags[pair[0]] = pair[1]
	}

	mon, err := monitor.New(cfg)
	if err != nil {
		log.Fatalf("Monitor construction failed: %v", err)
	}

	if err := mon.Run(); err != nil {
		log.Printf("Monitor operation failed: %v", err)
	}

	log.Printf("Monitor shutting down")
}

type KVSliceFlag [][2]string

func (f *KVSliceFlag) String() string {
	pairs := []string{}
	for k, v := range *f {
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(pairs, ",")
}

func (f *KVSliceFlag) Set(value string) error {
	pairs := strings.Split(value, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid value: %v", pair)
		}
		*f = append(*f, [2]string{parts[0], parts[1]})
	}
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
