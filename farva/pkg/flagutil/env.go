package flagutil

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

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
