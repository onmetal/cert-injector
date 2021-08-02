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
