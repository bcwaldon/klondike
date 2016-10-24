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
