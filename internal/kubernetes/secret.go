package kubernetes

import (
	"context"
	"fmt"

	"github.com/go-acme/lego/v4/certificate"
	corev1 "k8s.io/api/core/v1"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateOrUpdateSecretForCertificate(ctx context.Context, c client.Client, cert *certificate.Resource, req ctrl.Request) error {
	sec := prepareSecret(cert, req)
	err := c.Create(ctx, sec)
	if err != nil {
		if apierr.IsAlreadyExists(err) {
			return c.Update(ctx, sec)
		}
		return err
	}
	return err
}

func prepareSecret(cert *certificate.Resource, r ctrl.Request) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-tls", r.Name),
			Namespace: r.Namespace,
		},
		Type: corev1.SecretTypeTLS,
		Data: map[string][]byte{
			corev1.TLSCertKey:       cert.Certificate,
			corev1.TLSPrivateKeyKey: cert.PrivateKey,
		},
	}
}
