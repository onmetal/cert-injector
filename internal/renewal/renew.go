package renewal

import (
	"context"
	"crypto/ecdsa"
	"fmt"

	"github.com/onmetal/injector/internal/issuer/solver"

	"github.com/go-acme/lego/v4/registration"

	"k8s.io/apimachinery/pkg/types"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-logr/logr"
	injerr "github.com/onmetal/injector/internal/errors"
	"github.com/onmetal/injector/internal/issuer"
	"github.com/onmetal/injector/internal/kubernetes"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Renewer interface {
	RegisterChallengeProvider() error
	Renew() (*certificate.Resource, error)
}

type certs struct {
	ctx        context.Context
	legoClient *lego.Client
	k8sClient  client.Client
	log        logr.Logger
	svc        *corev1.Service
	cert       *certificate.Resource
}

func New(ctx context.Context, k8sClient client.Client, l logr.Logger, req ctrl.Request) (Renewer, error) {
	service, err := kubernetes.GetService(ctx, k8sClient, req)
	if err != nil {
		return nil, err
	}
	if !isRequired(service.Annotations) {
		return nil, injerr.NotRequired()
	}
	caURL := issuer.GetConfig(issuer.CaURLAnnotationKey, service.Annotations)
	email := issuer.GetConfig(issuer.EmailAnnotationKey, service.Annotations)

	currentCertificate, err := getCurrentCertificate(ctx, k8sClient, req)
	if err != nil {
		return nil, err
	}

	privateKey, err := issuer.GetPrivateKey(ctx, k8sClient, req.Namespace)
	if err != nil {
		return nil, err
	}
	user := getUser(email, privateKey)
	config := issuer.NewConfig(user, caURL)

	legoClient, err := lego.NewClient(config)
	if err != nil {
		return nil, err
	}
	reg, err := getRegistration(legoClient)
	if err != nil {
		return nil, err
	}
	user.Registration = reg
	return &certs{
		ctx:        ctx,
		legoClient: legoClient,
		k8sClient:  k8sClient,
		svc:        service,
		log:        l,
		cert:       currentCertificate,
	}, nil
}

func isRequired(m map[string]string) bool {
	v, ok := m[solver.InjectAnnotationKey]
	return ok && v == "done"
}

func getCurrentCertificate(ctx context.Context, c client.Client, req ctrl.Request) (*certificate.Resource, error) {
	secObj := ctrl.Request{NamespacedName: types.NamespacedName{
		Namespace: req.Namespace,
		Name:      fmt.Sprintf("%s-tls", req.Name)}}
	secret, err := kubernetes.GetSecret(ctx, c, secObj)
	if err != nil {
		return nil, err
	}
	cert := new(certificate.Resource)
	cert.Certificate = secret.Data[corev1.TLSCertKey]
	cert.PrivateKey = secret.Data[corev1.TLSPrivateKeyKey]
	return cert, nil
}

func getUser(email string, privateKey *ecdsa.PrivateKey) *issuer.User {
	return &issuer.User{
		Email: email,
		Key:   privateKey,
	}
}

func getRegistration(c *lego.Client) (*registration.Resource, error) {
	return c.Registration.ResolveAccountByKey()
}
