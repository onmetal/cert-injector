package renewal

import (
	"github.com/go-acme/lego/v4/certificate"
	"github.com/onmetal/injector/internal/issuer/solver"
)

func (c *certs) Solver() error {
	externalSolver := solver.NewExternalSolver(c.ctx, c.k8sClient, c.log, c.svc)
	return c.legoClient.Challenge.SetHTTP01Provider(externalSolver)
}

func (c *certs) Renew() (*certificate.Resource, error) {
	return c.legoClient.Certificate.Renew(*c.cert, true, false, "")
}
