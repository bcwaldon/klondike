package monitor

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/bcwaldon/farva/pkg/health"
)

var (
	DefaultConfig = Config{
		RefreshInterval: 30 * time.Second,
		HealthPort:      7334,
	}
)

type Config struct {
	RefreshInterval time.Duration
	HealthPort      int
	AWSRegion       string
	AWSInstanceTags map[string]string
	AWSLoadBalancer string
}

func New(cfg Config) (*Monitor, error) {
	am := newAWSManager(cfg.AWSRegion)

	mon := Monitor{
		cfg: cfg,
		am:  am,
	}

	return &mon, nil
}

type Monitor struct {
	cfg Config
	am  AWSManager
}

func (mon *Monitor) start() error {
	mon.startHTTPServer()
	return nil
}

func (mon *Monitor) startHTTPServer() {
	s := &http.Server{
		Addr:    fmt.Sprintf(":%d", mon.cfg.HealthPort),
		Handler: health.NewHandler(),
	}

	go func() {
		log.Fatal(s.ListenAndServe())
	}()
}

func (mon *Monitor) refresh() error {
	instances, err := mon.am.Instances(mon.cfg.AWSInstanceTags)
	if err != nil {
		return err
	}

	cfg := AWSLoadBalancerConfig{
		Name:      mon.cfg.AWSLoadBalancer,
		Instances: instances,
	}

	if err := mon.am.SyncLoadBalancer(cfg); err != nil {
		return err
	}

	return nil
}

func (mon *Monitor) Run() error {
	if err := mon.start(); err != nil {
		return err
	}

	log.Printf("Monitor started successfully, entering refresh loop")

	ticker := time.NewTicker(mon.cfg.RefreshInterval)

	for {
		if err := mon.refresh(); err != nil {
			log.Printf("Failed refreshing Monitor: %v", err)
		}

		//NOTE(bcwaldon): receive from the ticker at the
		// end of the loop to emulate do-while semantics.
		<-ticker.C
	}
}
