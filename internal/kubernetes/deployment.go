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

	"github.com/onmetal/injector/api"
	injerr "github.com/onmetal/injector/internal/errors"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (k *Kubernetes) InjectCertIntoDeployment() error {
	d, err := k.getDeployment(k.selector)
	if err != nil {
		if injerr.IsNotExist(err) {
			k.log.Info("deployment not exist")
			return nil
		}
		return err
	}
	d.Annotations[api.AdmissionWebhookAnnotationInjectKey] = api.AnnotationKeyEnabled
	d.Annotations[api.AdmissionWebhookAnnotationCertKey] = fmt.Sprintf("%s-tls", k.req.Name)
	return k.Update(k.ctx, d)
}

func (k *Kubernetes) getDeployment(selector map[string]string) (*appsv1.Deployment, error) {
	filter := &client.ListOptions{
		LabelSelector: client.MatchingLabelsSelector{Selector: labels.SelectorFromSet(selector)}}
	d := &appsv1.DeploymentList{}
	if err := k.List(k.ctx, d, filter); err != nil {
		return nil, err
	}
	if len(d.Items) == 0 {
		return nil, injerr.NotExist("deployment")
	}
	return &d.Items[0], nil
}
