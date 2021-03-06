// /*
// Copyright (c) 2021 T-Systems International GmbH, SAP SE or an SAP affiliate company. All right reserved
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// */

package renewal

import (
	"github.com/go-acme/lego/v4/certificate"
	"github.com/onmetal/injector/internal/issuer/solver"
)

func (c *certs) RegisterChallengeProvider() error {
	s := solver.New(c.ctx, c.k8sClient, c.log, c.svc)
	return c.legoClient.Challenge.SetHTTP01Provider(s)
}

func (c *certs) Renew() (*certificate.Resource, error) {
	return c.legoClient.Certificate.Renew(*c.cert, true, false, "")
}
