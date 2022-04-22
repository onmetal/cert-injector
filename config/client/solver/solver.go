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

package solver

import (
	"context"
	"fmt"
	"log"

	ctrl "sigs.k8s.io/controller-runtime"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sclient "k8s.io/client-go/kubernetes"
)

const (
	acmeHTTPResolver = "http-resolver"
	defaultImage     = "yotsyni/acmeresolver:latest"
)

type Provider interface {
	Present(domain, token, keyAuth string) error
	CleanUp(domain, token, keyAuth string) error
}

type external struct {
	*k8sclient.Clientset

	ctx context.Context
}

func NewExternalSolver() (Provider, error) {
	config := ctrl.GetConfigOrDie()
	cs, err := k8sclient.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &external{cs, context.Background()}, nil
}

func (e *external) Present(domain, token, keyAuth string) error {
	pod := e.preparePod(domain, token)
	_, err := e.Clientset.CoreV1().Pods("default").Create(e.ctx, pod, metav1.CreateOptions{})
	return err
}

func (e *external) preparePod(domain, token string) *corev1.Pod {
	name := fmt.Sprintf("%s-%s", acmeHTTPResolver, domain)
	labels := getLabelsForPod(name)
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: name, Namespace: "default", Labels: labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "test", Image: defaultImage, Env: []corev1.EnvVar{
					{Name: "DOMAIN_NAME", Value: domain},
					{Name: "TOKEN", Value: token},
				}},
			},
			RestartPolicy: "Always",
		},
	}
}

func (e *external) CleanUp(domain, token, keyAuth string) error {
	log.Println("clean up process started")
	name := fmt.Sprintf("%s-%s", acmeHTTPResolver, domain)
	return e.Clientset.CoreV1().Pods("default").Delete(e.ctx, name, metav1.DeleteOptions{})
}

func getLabelsForPod(name string) map[string]string {
	l := make(map[string]string, 1)
	l["app"] = "http-resolver"
	l["pod"] = name
	return l
}
