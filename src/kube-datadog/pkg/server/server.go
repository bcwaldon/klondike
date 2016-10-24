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
