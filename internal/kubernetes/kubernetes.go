package kubernetes

import (
	"context"

	"github.com/onmetal/injector/api"
	injerr "github.com/onmetal/injector/internal/errors"
	corev1 "k8s.io/api/core/v1"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Kubernetes struct {
	client.Client

	ctx      context.Context
	log      logr.Logger
	cert     *certificate.Resource
	req      ctrl.Request
	selector map[string]string
}

func New(ctx context.Context, c client.Client, l logr.Logger, cert *certificate.Resource, req ctrl.Request) (*Kubernetes, error) {
	s, err := GetService(ctx, c, req)
	if err != nil {
		return nil, err
	}
	if !isInjectNeeded(s.Annotations) {
		return nil, injerr.NotRequired()
	}
	return &Kubernetes{
		Client:   c,
		ctx:      ctx,
		log:      l,
		cert:     cert,
		req:      req,
		selector: s.Spec.Selector,
	}, nil
}

func GetService(ctx context.Context, c client.Client, req ctrl.Request) (*corev1.Service, error) {
	s := &corev1.Service{}
	err := c.Get(ctx, req.NamespacedName, s)
	return s, err
}

func isInjectNeeded(annotations map[string]string) bool {
	v, ok := annotations[api.AutoInjectAnnotationKey]
	return ok && v == api.AnnotationKeyEnabled
}
