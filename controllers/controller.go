/*
Copyright 2021.

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

package controllers

import (
	"context"

	injerr "github.com/onmetal/injector/internal/errors"
	"github.com/onmetal/injector/internal/issuer"
	"github.com/onmetal/injector/internal/kubernetes"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type Reconciler struct {
	client.Client

	Scheme *runtime.Scheme
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Service{}).
		Complete(r)
}

//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqLog := log.FromContext(ctx)

	i, err := issuer.New(ctx, r.Client, reqLog, req)
	if err != nil {
		if injerr.IsNotRequired(err) {
			return ctrl.Result{}, nil
		}
		reqLog.Info("can't create issuer", "error", err)
		return ctrl.Result{}, err
	}
	err = i.Solver()
	if err != nil {
		reqLog.Info("can't register http solver", "error", err)
		return ctrl.Result{}, err
	}
	err = i.Register()
	if err != nil {
		reqLog.Info("can't register client", "error", err)
		return ctrl.Result{}, err
	}
	cert, err := i.Obtain()
	if err != nil {
		reqLog.Info("can't obtain certificate", "error", err)
		return ctrl.Result{}, err
	}
	if err := kubernetes.CreateOrUpdateSecretForCertificate(ctx, r.Client, cert, req); err != nil {
		reqLog.Info("can't create secret for certificate", "error", err)
		return ctrl.Result{}, err
	}
	reqLog.Info("reconciliation finished")
	return ctrl.Result{}, nil
}
