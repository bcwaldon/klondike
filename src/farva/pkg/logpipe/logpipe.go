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
package logpipe

import (
	"bufio"
	"github.com/bcwaldon/klondike/src/farva/pkg/logger"
	"io"
	"os"
	"syscall"
)

type LogPipe struct {
	path string
}

func NewLogPipe(path string) LogPipe {
	return LogPipe{path: path}
}

func (l *LogPipe) Start() error {
	// Always remove the fifo initially, ignore error if it doesn't exist.
	os.Remove(l.path)
	if err := syscall.Mkfifo(l.path, 0777); err != nil {
		logger.Log.Errorf("Could create fifo at %s: %s", l.path, err)
		return err
	}

	go func() {
		f, err := os.Open(l.path)
		if err != nil {
			logger.Log.Errorf("Could not open fifo: %s", err)
			return
		}

		reader := bufio.NewReader(f)

		for {
			line, _, err := reader.ReadLine()
			if err != nil && err != io.EOF {
				logger.Log.Errorf("Could not read line from fifo: %s", err)
			} else {
				logger.Log.Printf("NGINX: %s", string(line))
			}
		}
	}()

	return nil
}
