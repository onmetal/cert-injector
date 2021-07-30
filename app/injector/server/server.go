package server

import (
	"log"

	"github.com/onmetal/injector/internal/logger"
)

type Server interface {
	Run() error
}
type server struct {
	router router
	log    logger.Logger
}

func New() Server {
	l := logger.New()
	r := newRouter(l)
	r.Handlers()
	return &server{
		router: r,
		log:    l,
	}
}

func (s *server) Run() error {
	log.Printf("server started on port: %s", "8443")
	return s.router.ListenAndServe()
}
