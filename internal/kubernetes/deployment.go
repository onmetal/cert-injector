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

	"k8s.io/apimachinery/pkg/types"

	"github.com/onmetal/injector/app/injector/server"
	injerr "github.com/onmetal/injector/internal/errors"
	appsv1 "k8s.io/api/apps/v1"
)

const (
	autoInjectAnnotationKey     = "cert.injector.ko/auto-inject"
	deploymentNameAnnotationKey = "cert.injector.ko/deployment-name"
	injectEnabled               = "true"
)

func (k *Kubernetes) InjectCertIntoDeployment() error {
	if !isInjectNeeded(k.annotations) {
		return injerr.NotRequired()
	}
	d, err := k.getDeployment()
	if err != nil {
		if injerr.IsNotExist(err) {
			k.log.Info("deployment not exist")
			return nil
		}
		if injerr.IsNotFound(err) {
			k.log.Info("deployment name not found")
			return nil
		}
		return err
	}
	d.Annotations[server.AdmissionWebhookAnnotationInjectKey] = injectEnabled
	d.Annotations[server.AdmissionWebhookAnnotationCertKey] = fmt.Sprintf("%s-tls", k.req.Name)
	return k.Update(k.ctx, d)
}

func (k *Kubernetes) getDeployment() (*appsv1.Deployment, error) {
	name, ok := k.annotations[deploymentNameAnnotationKey]
	if !ok {
		return nil, injerr.NotFound()
	}
	obj := types.NamespacedName{
		Namespace: k.req.Namespace,
		Name:      name,
	}
	d := &appsv1.Deployment{}
	if err := k.Get(k.ctx, obj, d); err != nil {
		return nil, err
	}
	return d, nil
}

func isInjectNeeded(annotations map[string]string) bool {
	v, ok := annotations[autoInjectAnnotationKey]
	return ok && v == injectEnabled
}
