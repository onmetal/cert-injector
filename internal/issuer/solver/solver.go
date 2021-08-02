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
	"log"
	"os"
	"time"

	"github.com/onmetal/injector/api"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const acmeHTTPResolver = "acmeresolver-"
const defaultImage = "yotsyni/acmeresolver:latest"

const waitForServiceSwitchSecond = 45 * time.Second

type Provider interface {
	Present(domain, token, keyAuth string) error
	CleanUp(domain, token, keyAuth string) error
}

type external struct {
	client.Client

	ctx   context.Context
	log   logr.Logger
	svc   *corev1.Service
	image string
}

func NewExternalSolver(ctx context.Context, c client.Client, l logr.Logger, svc *corev1.Service) Provider {
	image := defaultImage
	if os.Getenv("RESOLVER_CUSTOM_IMAGE") != "" {
		image = os.Getenv("RESOLVER_CUSTOM_IMAGE")
	}
	return &external{
		Client: c,
		ctx:    ctx,
		log:    l,
		svc:    svc,
		image:  image,
	}
}

func (e *external) Present(domain, token, keyAuth string) error {
	if err := e.changeServiceSelector(); err != nil {
		return err
	}
	pod := e.preparePod(domain, token, keyAuth)
	if err := e.Create(e.ctx, pod); err != nil {
		return err
	}
	time.Sleep(waitForServiceSwitchSecond)
	return nil
}

func (e *external) changeServiceSelector() error {
	e.svc.Spec.Selector["acmesolver"] = "true"
	return e.Client.Update(e.ctx, e.svc)
}

func (e *external) preparePod(domain, token, keyAuth string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: acmeHTTPResolver, Namespace: e.svc.Namespace, Labels: e.svc.Spec.Selector},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "test", Image: e.image, Env: []corev1.EnvVar{
					{Name: "DOMAIN_NAME", Value: domain},
					{Name: "TOKEN", Value: token},
					{Name: "AUTH_KEY", Value: keyAuth}}}},
			RestartPolicy: "Always"}}
}

func (e *external) CleanUp(domain, token, keyAuth string) error {
	log.Println("clean up process started")

	pods := &corev1.PodList{}
	filter := &client.ListOptions{
		LabelSelector: client.MatchingLabelsSelector{
			Selector: labels.SelectorFromSet(e.svc.Spec.Selector)}}
	if err := e.List(e.ctx, pods, filter); err != nil {
		e.log.Info("can't list pods with labels", "error", err)
		return err
	}
	if len(pods.Items) == 0 {
		e.log.Info("no pods for deletion found")
		return e.reverseServiceSelector()
	}
	if err := e.Delete(e.ctx, &pods.Items[0]); err != nil {
		e.log.Info("can't delete acme resolver pod", "error", err)
		return err
	}
	return e.reverseServiceSelector()
}

func (e *external) reverseServiceSelector() error {
	e.svc.ObjectMeta.Annotations[api.InjectAnnotationKey] = "done"
	_, ok := e.svc.Spec.Selector["acmesolver"]
	if ok {
		delete(e.svc.Spec.Selector, "acmesolver")
	}
	e.log.Info("reverting service labels")
	return e.Client.Update(e.ctx, e.svc)
}
