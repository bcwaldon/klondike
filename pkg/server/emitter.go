package server

import (
	"fmt"
	"regexp"
	"strconv"
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

type Emitter interface {
	Emit() (MetricsBundle, error)
}
