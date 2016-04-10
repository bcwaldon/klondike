package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	regexMemory = regexp.MustCompile("^([0-9].*)Mi$")
	regexCPU    = regexp.MustCompile("^([0-9].*)m$")

	size_b  = 1
	size_kb = size_b * 1024
	size_mb = size_kb * 1024
)

func parseComputeResource(value string, re *regexp.Regexp) (int, error) {
	groups := re.FindStringSubmatch(value)
	if len(groups) < 2 {
		return 0, fmt.Errorf("could not find submatch in %s", value)
	}
	return strconv.Atoi(groups[1])
}

func parseComputeResourceCPU(value string) (int, error) {
	return parseComputeResource(value, regexCPU)
}

func parseComputeResourceMemory(value string) (int, error) {
	val, err := parseComputeResource(value, regexMemory)
	if err != nil {
		return 0, err
	}
	return val * size_mb, nil
}

func NewKubeletEmitter(source string) Emitter {
	return &kubeletEmitter{
		source: &url.URL{Scheme: "http", Host: source},
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

type Emitter interface {
	Emit() (MetricsBundle, error)
}

type kubeletEmitter struct {
	source *url.URL
	client *http.Client
}

func (c *kubeletEmitter) Emit() (MetricsBundle, error) {
	endpoint, err := c.source.Parse("/pods")
	if err != nil {
		return MetricsBundle{}, fmt.Errorf("failed fetching pods: %v", err)
	}
	resp, err := c.client.Get(endpoint.String())
	if err != nil {
		return MetricsBundle{}, fmt.Errorf("failed fetching pods: %v", err)
	}

	var pl kubeletPodList
	if err := json.NewDecoder(resp.Body).Decode(&pl); err != nil {
		return MetricsBundle{}, fmt.Errorf("failed parsing pods: %v", err)
	}
	defer resp.Body.Close()

	metrics, err := pl.metrics()
	if err != nil {
		return MetricsBundle{}, fmt.Errorf("failed gathering pod metrics: %v", err)
	}

	return MetricsBundle{metrics: metrics}, nil
}

type kubeletPodList struct {
	Kind       string
	APIVersion string
	Items      []struct {
		Metadata struct {
			Name      string
			Namespace string
		}
		Status struct {
			Phase string
		}
		Spec struct {
			Containers []struct {
				Resources struct {
					Requests struct {
						CPU    string
						Memory string
					}
				}
			}
		}
	}
}

func (l *kubeletPodList) valid() error {
	if l.APIVersion != "v1" {
		return fmt.Errorf("invalid 'apiVersion' value: %s", l.APIVersion)
	} else if l.Kind != "PodList" {
		return fmt.Errorf("invalid 'kind' value: %s", l.Kind)
	}

	return nil
}

func (l *kubeletPodList) metrics() (chan Metric, error) {
	if err := l.valid(); err != nil {
		return nil, err
	}

	ch := make(chan Metric)
	go func() {
		l.emit(ch)
		close(ch)
	}()

	return ch, nil
}

func (l *kubeletPodList) emit(ch chan Metric) {
	ch <- Metric{Name: "kubelet.pod.all", Value: len(l.Items)}

	statii := map[string]int{
		"running": 0,
		"pending": 0,
	}

	var reservedMemory, reservedCPU int

	for _, i := range l.Items {
		status := strings.ToLower(i.Status.Phase)
		count, ok := statii[status]
		if !ok {
			log.Printf("got unrecognized pod status %q", status)
			continue
		}
		statii[status] = count + 1

		for _, c := range i.Spec.Containers {
			rr := c.Resources.Requests

			if rr.Memory != "" {
				contMemory, err := parseComputeResourceMemory(rr.Memory)
				if err != nil {
					log.Printf("failed parsing memory reservation: %v", err)
				} else {
					reservedMemory += contMemory
				}
			}
			if rr.CPU != "" {
				contCPU, err := parseComputeResourceCPU(rr.CPU)
				if err != nil {
					log.Printf("failed parsing CPU reservation: %v", err)
				} else {
					reservedCPU += contCPU
				}
			}
		}
	}

	for status, count := range statii {
		ch <- Metric{Name: fmt.Sprintf("kubelet.pod.%s", status), Value: count}
	}

	ch <- Metric{Name: "kubelet.pod.memory.reserved", Value: reservedMemory}
	ch <- Metric{Name: "kubelet.pod.cpu.reserved", Value: reservedCPU}
}
