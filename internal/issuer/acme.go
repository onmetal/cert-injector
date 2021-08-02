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

package issuer

import (
	"strings"
	"time"

	"github.com/onmetal/injector/api"

	"github.com/onmetal/injector/internal/issuer/solver"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/registration"
	injerr "github.com/onmetal/injector/internal/errors"
)

const waitServiceForSwitchSecond = 45 * time.Second

func (c *certs) Solver() error {
	externalSolver := solver.NewExternalSolver(c.ctx, c.k8sClient, c.log, c.svc)
	return c.legoClient.Challenge.SetHTTP01Provider(externalSolver)
}

func (c *certs) Register() error {
	reg, err := c.legoClient.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		c.log.Info("can't register new user, %s", "error", err)
		return err
	}
	c.User.Registration = reg
	return err
}

func (c *certs) Obtain() (*certificate.Resource, error) {
	d, ok := c.svc.Annotations[api.DomainsAnnotationKey]
	if !ok {
		return nil, injerr.NotExist("domain name")
	}
	domains := strings.Split(d, ",")
	request := certificate.ObtainRequest{
		Domains: domains,
		Bundle:  true,
	}
	time.Sleep(waitServiceForSwitchSecond)
	return c.legoClient.Certificate.Obtain(request)
}

func (c *certs) Renew() (*certificate.Resource, error) {
	return c.legoClient.Certificate.Renew(*c.cert, true, true, "")
}
