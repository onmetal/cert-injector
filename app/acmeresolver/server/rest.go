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
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/onmetal/injector/internal/logger"
)

const (
	// HTTPChallengePath is the path prefix used for http-01 challenge requests
	HTTPChallengePath = "/.well-known/acme-challenge"
)

type httpChallenge struct {
	domain, token, authKey string
	log                    logger.Logger
}

func newRouter(l logger.Logger) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.Recoverer)
	router.Use(middleware.Logger)
	newHandlers(router, l)
	return router
}

func newHandlers(r *chi.Mux, l logger.Logger) {
	c := newHTTPChallenge(l)
	p := fmt.Sprintf("%s/%s", HTTPChallengePath, c.token)
	l.Info("listening on", "path", p)
	r.Get(p, c.challenge)
	r.Post(p, c.challenge)
}

func newHTTPChallenge(l logger.Logger) httpChallenge {
	domain := os.Getenv("DOMAIN_NAME")
	if domain == "" {
		l.Info("domain not provided")
	}
	token := os.Getenv("TOKEN")
	if token == "" {
		l.Info("token not provided")
	}
	authKey := os.Getenv("AUTH_KEY")
	if authKey == "" {
		l.Info("auth key not provided")
	}
	return httpChallenge{
		domain:  domain,
		token:   token,
		authKey: authKey,
		log:     l,
	}
}

func (h *httpChallenge) challenge(w http.ResponseWriter, r *http.Request) {
	host := strings.Split(r.Host, ":")[0]
	basePath := path.Dir(r.URL.EscapedPath())
	token := path.Base(r.URL.EscapedPath())
	key := h.authKey

	log := h.log.WithValues(
		"host", host,
		"path", r.URL.EscapedPath(),
		"base_path", basePath,
		"token", token)

	log.Info("validating request")
	// verify the base path is correct
	if basePath != HTTPChallengePath {
		log.Info("invalid base_path", "expected_base_path", HTTPChallengePath)
		http.NotFound(w, r)
		return
	}

	log.Info("comparing host", "expected_host", h.domain)
	if h.domain != host {
		log.Info("invalid host", "expected_host", h.domain)
		http.NotFound(w, r)
		return
	}

	log.Info("comparing token", "expected_token", h.token)
	if h.token != token {
		// if nothing else, we return a 404 here
		log.Info("invalid token", "expected_token", h.token)
		http.NotFound(w, r)
		return
	}

	log.Info("got successful challenge request, writing key")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, key)
}
