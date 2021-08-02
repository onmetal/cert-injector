/*
Copyright (c) 2021 T-Systems International GmbH, SAP SE or an SAP affiliate company. All right reserved
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
