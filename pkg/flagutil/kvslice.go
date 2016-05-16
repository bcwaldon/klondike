package flagutil

import (
	"fmt"
	"strings"
)

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
		*f = append(*f, [2]string{strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])})
	}
	return nil
}
