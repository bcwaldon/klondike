package logpipe

import (
	"bufio"
	"github.com/bcwaldon/farva/pkg/logger"
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
