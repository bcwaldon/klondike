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
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/model"
)

// map of prometheus metrics exposed by the kubelet to
// names that should be used when exposing the sample
// data to datadog
var metricMap = map[string]string{
	"kubelet_running_pod_count":       "kubelet.pod.all",
	"kubelet_running_container_count": "kubelet.container.all",
}

const acceptHeader = `application/vnd.google.protobuf;proto=io.prometheus.client.MetricFamily;encoding=delimited;q=0.7,text/plain;version=0.0.4;q=0.3,application/json;schema="prometheus/telemetry";version=0.0.2;q=0.2,*/*;q=0.1`

func NewKubeletEmitter(source string) Emitter {
	ep := &url.URL{Scheme: "http", Host: source, Path: "/metrics"}
	return &kubeletEmitter{
		client: &http.Client{Timeout: 5 * time.Second},
		source: ep.String(),
	}
}

type kubeletEmitter struct {
	client *http.Client
	source string
}

func (c *kubeletEmitter) Emit() (MetricsBundle, error) {
	ch := filteredSamples(collectSamples(c.client, c.source))
	return MetricsBundle{metrics: ch}, nil
}

func filteredSamples(in <-chan *model.Sample) chan Metric {
	out := make(chan Metric)
	go func() {
		for s := range in {
			pMetricName, ok := s.Metric[model.MetricNameLabel]
			if !ok {
				continue
			}

			dMetricName, ok := metricMap[string(pMetricName)]
			if !ok {
				continue
			}

			out <- Metric{Name: dMetricName, Value: int(s.Value)}
		}
		close(out)
	}()
	return out
}

func collectSamples(client *http.Client, endpoint string) <-chan *model.Sample {
	collect := func(ch chan<- *model.Sample) error {
		req, err := http.NewRequest("GET", endpoint, nil)
		if err != nil {
			return err
		}
		req.Header.Add("Accept", acceptHeader)

		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("server returned HTTP status %s", resp.Status)
		}

		decSamples := make(model.Vector, 0, 50)

		sdec := expfmt.SampleDecoder{
			Dec: expfmt.NewDecoder(resp.Body, expfmt.ResponseFormat(resp.Header)),
			Opts: &expfmt.DecodeOptions{
				Timestamp: model.TimeFromUnixNano(time.Now().UnixNano()),
			},
		}

		for {
			if err = sdec.Decode(&decSamples); err != nil {
				break
			}
			for _, s := range decSamples {
				s := s
				ch <- s
			}
			decSamples = decSamples[:0]
		}

		if err == io.EOF {
			err = nil
		}

		return err
	}

	ch := make(chan *model.Sample)

	go func() {
		if err := collect(ch); err != nil {
			log.Printf("failed collecting metrics from prometheus: %v", err)
		}
		close(ch)
	}()

	return ch
}
