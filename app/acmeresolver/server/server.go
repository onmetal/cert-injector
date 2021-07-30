package server

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/onmetal/injector/internal/logger"
)

type Server interface {
	Run() error
}

type server struct {
	router *chi.Mux
	log    logger.Logger
	port   string
}

func New() Server {
	port := ":8080"
	if os.Getenv("PORT") != "" {
		port = fmt.Sprintf(":%s", os.Getenv("PORT"))
	}
	l := logger.New()
	r := newRouter(l)
	return &server{
		router: r,
		log:    l,
		port:   port,
	}
}

func (s *server) Run() error {
	s.log.Info("server started", "port", s.port)
	return http.ListenAndServe(s.port, s.router)
}
