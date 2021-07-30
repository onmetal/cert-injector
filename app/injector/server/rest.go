package server

import (
	"net/http"

	"github.com/onmetal/injector/internal/logger"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type router interface {
	ListenAndServe() error
	Handlers()
}

type chiRouter struct {
	*chi.Mux
	log logger.Logger
}

func newRouter(l logger.Logger) router {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	return &chiRouter{
		Mux: r,
		log: l,
	}
}

func (c *chiRouter) Handlers() {
	c.Route("/api/v1", func(r chi.Router) {
		r.Post("/mutate", c.mutateHandler)
	})
}

func (c *chiRouter) ListenAndServe() error {
	return http.ListenAndServeTLS(":8443", "/tmp/certs/tls.crt", "/tmp/certs/tls.key", c.Mux)
}
