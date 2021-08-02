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

package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/onmetal/injector/api"

	"github.com/onmetal/injector/app/injector/patch"

	v1 "k8s.io/api/admission/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()
)

const (
	jsonSpecContainers = "/spec/template/spec/containers"
	jsonSpecVolumes    = "/spec/template/spec/volumes"
)

const volumeName = "tls-certificates"

func (c *chiRouter) mutateHandler(w http.ResponseWriter, r *http.Request) {
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}
	if len(body) == 0 {
		c.log.Info("empty body")
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		c.log.Info("Content-Type=%s, expect application/json")
		http.Error(w, "invalid Content-Type, expect `application/json`", http.StatusUnsupportedMediaType)
		return
	}

	admissionResponse := &v1.AdmissionReview{}
	if _, _, err := deserializer.Decode(body, nil, admissionResponse); err != nil {
		c.log.Error("Can't decode body", err)
		admissionResponse.Response = &v1.AdmissionResponse{
			Result: &metav1.Status{Message: err.Error()}}
	} else {
		admissionResponse.Response = c.mutate(admissionResponse)
	}

	resp, err := json.Marshal(admissionResponse)
	if err != nil {
		c.log.Error("Can't encode response: %v", err)
		http.Error(w, fmt.Sprintf("could not encode response: %v", err), http.StatusInternalServerError)
	}
	if _, writeErr := w.Write(resp); writeErr != nil {
		c.log.Error("Can't write response: %v", writeErr)
		http.Error(w, fmt.Sprintf("could not write response: %v", writeErr), http.StatusInternalServerError)
	}
}

func (c *chiRouter) mutate(ar *v1.AdmissionReview) *v1.AdmissionResponse {
	deployment, err := getDeployment(ar.Request.Object.Raw)
	if err != nil {
		c.log.Error("can't unmarshal deployment from admission request", err)
		return &v1.AdmissionResponse{Allowed: false, UID: ar.Request.UID, Result: &metav1.Status{Message: err.Error()}}
	}

	if !isMutationRequired(deployment.Annotations, deployment.Spec.Template.Spec.Volumes) {
		return &v1.AdmissionResponse{Allowed: true, UID: ar.Request.UID}
	}

	secretName, ok := deployment.Annotations[api.AdmissionWebhookAnnotationCertKey]
	if !ok {
		return &v1.AdmissionResponse{Allowed: false, UID: ar.Request.UID,
			Result: &metav1.Status{Message: "secret with certs not provided"}}
	}

	body, err := mutateDeployment(secretName, deployment)
	if err != nil {
		c.log.Error("can't mutate deployment", err)
		return &v1.AdmissionResponse{Allowed: false, UID: ar.Request.UID,
			Result: &metav1.Status{Message: "can't mutate deployment"}}
	}

	return &v1.AdmissionResponse{Allowed: true, Patch: body, UID: ar.Request.UID, PatchType: getPatchType()}
}

func getDeployment(req []byte) (*appsv1.Deployment, error) {
	d := &appsv1.Deployment{}
	if err := json.Unmarshal(req, d); err != nil {
		return nil, err
	}
	return d, nil
}

func isMutationRequired(a map[string]string, volumes []corev1.Volume) bool {
	for v := range volumes {
		if volumes[v].Name == volumeName {
			return false
		}
	}
	value, ok := a[api.AdmissionWebhookAnnotationInjectKey]
	return ok && value == "true"
}

func mutateDeployment(secretName string, deployment *appsv1.Deployment) ([]byte, error) {
	var operations []patch.Operation
	containers := updateContainer(deployment.Spec.Template.Spec.Containers)
	volume := addVolume(deployment.Spec.Template.Spec.Volumes, secretName)
	operations = append(operations,
		patch.AddPatchOperation(jsonSpecContainers, containers),
		patch.AddPatchOperation(jsonSpecVolumes, volume))
	return json.Marshal(operations)
}

func updateContainer(containers []corev1.Container) []corev1.Container {
	containers[0].VolumeMounts = append(containers[0].VolumeMounts,
		corev1.VolumeMount{Name: volumeName, MountPath: "/certs", ReadOnly: true})
	return containers
}

func addVolume(volumes []corev1.Volume, name string) []corev1.Volume {
	var isOptional = true
	return append(volumes, corev1.Volume{
		Name: volumeName,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: name,
				Optional:   &isOptional,
			},
		},
	})
}

func getPatchType() *v1.PatchType {
	jsonPatchType := v1.PatchTypeJSONPatch
	return &jsonPatchType
}
