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
