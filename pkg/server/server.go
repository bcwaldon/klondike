package server

import (
	"log"
	"time"
)

type Config struct {
	KubeletHost   string
	DogStatsDHost string
	Period        time.Duration
	Tags          []string
}

type Server struct {
	cfg       Config
	emitter   Emitter
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
	metrics, err := s.emitter.Emit()
	if err != nil {
		return err
	}
	s.publisher.Publish(metrics)
	return nil
}

func New(cfg Config) (*Server, error) {
	publisher, err := NewDogstatsdPublisher(cfg.DogStatsDHost, cfg.Tags)
	if err != nil {
		return nil, err
	}
	emitter := NewKubeletEmitter(cfg.KubeletHost)
	srv := &Server{
		cfg:       cfg,
		publisher: publisher,
		emitter:   emitter,
	}
	return srv, nil
}
