package issuer

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/onmetal/injector/api"
	injerr "github.com/onmetal/injector/internal/errors"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
)

const defaultEmail = "your@email.local"

type Issuer interface {
	Register() error
	Solver() error
	Obtain() (*certificate.Resource, error)
	Renew() (*certificate.Resource, error)
}

type certs struct {
	ctx        context.Context
	legoClient *lego.Client
	k8sClient  client.Client
	log        logr.Logger
	svc        *corev1.Service
	User       *User
	cert       *certificate.Resource
}

// User - user or account type that implements acme.User
type User struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *User) GetEmail() string {
	return u.Email
}

func (u User) GetRegistration() *registration.Resource {
	return u.Registration
}

func (u *User) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

func New(ctx context.Context, k8sClient client.Client, l logr.Logger, req ctrl.Request) (Issuer, error) {
	service, err := getService(ctx, k8sClient, req)
	if err != nil {
		return nil, err
	}
	if !isRequired(service.Annotations) {
		return nil, injerr.NotRequired()
	}
	caURL := getConfig(api.CaURLAnnotationKey, service.Annotations)
	email := getConfig(api.EmailAnnotationKey, service.Annotations)

	privateKey, err := createPrivateKey()
	if err != nil {
		return nil, err
	}
	user := newUser(email, privateKey)
	config := newConfig(user, caURL)

	legoClient, err := lego.NewClient(config)
	if err != nil {
		return nil, err
	}

	return &certs{
		ctx:        ctx,
		legoClient: legoClient,
		k8sClient:  k8sClient,
		log:        l,
		svc:        service,
		User:       user,
	}, nil
}

func getService(ctx context.Context, c client.Client, req ctrl.Request) (*corev1.Service, error) {
	s := &corev1.Service{}
	err := c.Get(ctx, req.NamespacedName, s)
	return s, err
}

func isRequired(m map[string]string) bool {
	v, ok := m[api.InjectAnnotationKey]
	return ok && v == "true"
}

func getConfig(s string, m map[string]string) string {
	switch s {
	case api.CaURLAnnotationKey:
		v, ok := m[api.CaURLAnnotationKey]
		if !ok {
			return lego.LEDirectoryStaging
		}
		return v
	case api.EmailAnnotationKey:
		v, ok := m[api.EmailAnnotationKey]
		if !ok {
			return defaultEmail
		}
		return v
	default:
		return ""
	}
}

func createPrivateKey() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

func newUser(email string, privateKey crypto.PrivateKey) *User {
	return &User{
		Email: email,
		key:   privateKey,
	}
}

func newConfig(u *User, caURL string) *lego.Config {
	config := lego.NewConfig(u)
	config.CADirURL = caURL
	return config
}
