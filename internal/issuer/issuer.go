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
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"

	apierr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	"github.com/onmetal/injector/internal/kubernetes"

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
const privateKeySecretName = "le-issuer"

type Issuer interface {
	Register() error
	Solver() error
	Obtain() (*certificate.Resource, error)
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
	Key          crypto.PrivateKey
}

func (u *User) GetEmail() string {
	return u.Email
}

func (u User) GetRegistration() *registration.Resource {
	return u.Registration
}

func (u *User) GetPrivateKey() crypto.PrivateKey {
	return u.Key
}

func New(ctx context.Context, k8sClient client.Client, l logr.Logger, req ctrl.Request) (Issuer, error) {
	service, err := kubernetes.GetService(ctx, k8sClient, req)
	if err != nil {
		return nil, err
	}
	if !isRequired(service.Annotations) {
		return nil, injerr.NotRequired()
	}
	caURL := GetConfig(api.CaURLAnnotationKey, service.Annotations)
	email := GetConfig(api.EmailAnnotationKey, service.Annotations)

	var privateKey *ecdsa.PrivateKey
	privateKey, err = GetPrivateKey(ctx, k8sClient, req.Namespace)
	if err != nil && apierr.IsNotFound(err) {
		privateKey, err = createPrivateKey(ctx, k8sClient, req.Namespace)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	user := newUser(email, privateKey)
	config := NewConfig(user, caURL)

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

func isRequired(m map[string]string) bool {
	v, ok := m[api.InjectAnnotationKey]
	return ok && v == api.AnnotationKeyEnabled
}

func GetPrivateKey(ctx context.Context, c client.Client, namespace string) (*ecdsa.PrivateKey, error) {
	objKey := ctrl.Request{NamespacedName: types.NamespacedName{
		Namespace: namespace,
		Name:      privateKeySecretName,
	}}
	secretPrivateKey, err := kubernetes.GetSecret(ctx, c, objKey)
	if err != nil {
		return nil, err
	}
	return x509.ParseECPrivateKey(secretPrivateKey.Data[corev1.TLSPrivateKeyKey])
}

func createPrivateKey(ctx context.Context, c client.Client, namespace string) (*ecdsa.PrivateKey, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	data, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return nil, err
	}
	preparedSecret := preparePrivateKeySecret(data, namespace)
	if createErr := kubernetes.CreateSecret(ctx, c, preparedSecret); createErr != nil {
		return privateKey, createErr
	}
	return privateKey, err
}

func preparePrivateKeySecret(data []byte, namespace string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: privateKeySecretName, Namespace: namespace},
		Immutable:  pointer.Bool(true),
		Data:       map[string][]byte{corev1.TLSPrivateKeyKey: data},
		Type:       corev1.SecretTypeOpaque,
	}
}

func GetConfig(s string, m map[string]string) string {
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

func newUser(email string, privateKey *ecdsa.PrivateKey) *User {
	return &User{
		Email: email,
		Key:   privateKey,
	}
}

func NewConfig(u *User, caURL string) *lego.Config {
	config := lego.NewConfig(u)
	config.CADirURL = caURL
	return config
}
