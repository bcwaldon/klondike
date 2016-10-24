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
