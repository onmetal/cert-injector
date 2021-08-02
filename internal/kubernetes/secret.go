package kubernetes

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k *Kubernetes) CreateOrUpdateSecretForCertificate() error {
	sec := k.prepareSecret()
	err := k.Create(k.ctx, sec)
	if err != nil {
		if apierr.IsAlreadyExists(err) {
			return k.Update(k.ctx, sec)
		}
		return err
	}
	return err
}

func (k *Kubernetes) prepareSecret() *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-tls", k.req.Name),
			Namespace: k.req.Namespace,
		},
		Type: corev1.SecretTypeTLS,
		Data: map[string][]byte{
			corev1.TLSCertKey:       k.cert.Certificate,
			corev1.TLSPrivateKeyKey: k.cert.PrivateKey,
		},
	}
}
