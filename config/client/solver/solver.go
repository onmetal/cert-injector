package solver

import (
	"context"
	"fmt"
	ctrl "sigs.k8s.io/controller-runtime"

	"log"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sclient "k8s.io/client-go/kubernetes"
)

const acmeHTTPResolver = "http-resolver"
const defaultImage = "yotsyni/acmeresolver:latest"

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
			Name: name, Namespace: "default", Labels: labels},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "test", Image: defaultImage, Env: []corev1.EnvVar{
					{Name: "DOMAIN_NAME", Value: domain},
					{Name: "TOKEN", Value: token}}}},
			RestartPolicy: "Always"}}
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
