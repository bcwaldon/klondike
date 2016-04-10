package server

import (
	"fmt"
	"log"
	"strings"

	"github.com/DataDog/datadog-go/statsd"
)

type Publisher interface {
	Publish(MetricsBundle)
}

func NewLogPublisher() Publisher {
	return &logPublisher{}
}

type logPublisher struct{}

func (p *logPublisher) Publish(b MetricsBundle) {
	metrics := b.Metrics()
	for m := range metrics {
		log.Printf("%s: %v; %s\n", m.Name, m.Value, strings.Join(m.Tags, ","))
	}
}

func NewDogstatsdPublisher(dest string, tags []string) (Publisher, error) {
	c, err := statsd.New(dest)
	if err != nil {
		return nil, fmt.Errorf("failed constructing publisher: %v", err)
	}
	c.Tags = append(c.Tags, tags...)
	return &dogstatsdPublisher{client: c}, nil
}

type dogstatsdPublisher struct {
	client *statsd.Client
}

func (p *dogstatsdPublisher) Publish(b MetricsBundle) {
	metrics := b.Metrics()
	for m := range metrics {
		if err := p.client.Gauge(m.Name, float64(m.Value), nil, 1); err != nil {
			log.Printf("Failed sending metric to datadog: %v", err)
			continue
		}
	}
}
