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
package flagutil

import (
	"fmt"
	"strings"
)

type KVSliceFlag [][2]string

func (f *KVSliceFlag) String() string {
	pairs := []string{}
	for _, v := range *f {
		pairs = append(pairs, fmt.Sprintf("%s=%s", v[0], v[1]))
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
		*f = append(*f, [2]string{strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])})
	}
	return nil
}
