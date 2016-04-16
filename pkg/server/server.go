package server

import (
	"log"
	"time"
)

type Config struct {
	Emitter       Emitter
	DogStatsDHost string
	Period        time.Duration
	Tags          []string
	NoPublish     bool
}

type Server struct {
	cfg       Config
	publisher Publisher
}

func (s *Server) Run() {
	for {
		if err := s.attempt(); err != nil {
			log.Printf("Failed collecting metrics: %v", err)
		}

		<-time.After(s.cfg.Period)
	}
}

func (s *Server) attempt() error {
	metrics, err := s.cfg.Emitter.Emit()
	if err != nil {
		return err
	}
	s.publisher.Publish(metrics)
	return nil
}

func New(cfg Config) (*Server, error) {
	var publisher Publisher
	if cfg.NoPublish {
		publisher = NewLogPublisher()
	} else {
		var err error
		publisher, err = NewDogstatsdPublisher(cfg.DogStatsDHost, cfg.Tags)
		if err != nil {
			return nil, err
		}
	}

	srv := &Server{
		cfg:       cfg,
		publisher: publisher,
	}
	return srv, nil
}
