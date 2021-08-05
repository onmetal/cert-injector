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
