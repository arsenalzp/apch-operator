/*
Copyright 2023.

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

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"apache-operator/api/v1alpha1"
	apachewebv1alpha1 "apache-operator/api/v1alpha1"
)

// ApachewebReconciler reconciles a Apacheweb object
type ApachewebReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=apacheweb.arsenal.dev,resources=apachewebs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apacheweb.arsenal.dev,resources=apachewebs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apacheweb.arsenal.dev,resources=apachewebs/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=events,verbs=create;patch
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Apacheweb object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *ApachewebReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var confMap corev1.ConfigMap
	var deployment appsv1.Deployment
	var apacheWeb v1alpha1.Apacheweb

	logr := log.FromContext(ctx).WithValues("ApacheWeb", req.NamespacedName)

	logr.Info("start ApacheWeb reconciliation")
	if err := r.Get(ctx, req.NamespacedName, &apacheWeb); err != nil {
		if errors.IsNotFound(err) {
			logr.Error(err, "ApacheWeb not found")
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}

		logr.Error(err, "unable to fetch ApacheWeb")
		return ctrl.Result{}, err
	}

	// Try to get the resource
	err := r.Get(ctx, req.NamespacedName, &confMap)
	if err != nil {
		if errors.IsNotFound(err) {
			confMap, err = r.dependentConfmap(apacheWeb)
			if err != nil {
				logr.Error(err, "unable to create Apacheweb configMap")
				return ctrl.Result{}, err
			}

			// Create the resource
			if err := r.Create(ctx, &confMap); err != nil {
				return ctrl.Result{}, err
			}

			// Record the creation event
			r.recorder.Eventf(&apacheWeb, "Normal", "Created", "ConfigMap %s created", confMap.Name)
		}
		return ctrl.Result{}, err
	}

	// Try to get the resource
	err = r.Get(ctx, req.NamespacedName, &deployment)
	if err != nil {
		if errors.IsNotFound(err) {
			deployment, err = r.dependentDeployment(apacheWeb, confMap)
			if err != nil {
				logr.Error(err, "unable to create Apacheweb deployment")
				return ctrl.Result{}, err
			}

			// Create the resource
			if err := r.Create(ctx, &deployment); err != nil {
				return ctrl.Result{}, err
			}

			// Record the creation event
			r.recorder.Eventf(&apacheWeb, "Normal", "Created", "Deployment %s created", deployment.Name)
		}
		return ctrl.Result{}, err
	}

	applyOpts := []client.PatchOption{client.ForceOwnership, client.FieldOwner("ApacheWeb")}

	// Create Deployment template
	deployment, err = r.dependentDeployment(apacheWeb, confMap)
	if err != nil {
		logr.Error(err, "unable to create Apacheweb deployment")
		return ctrl.Result{}, err
	}

	// Patch Deployment resource
	err = r.Patch(ctx, &deployment, client.Apply, applyOpts...)
	if err != nil {
		logr.Error(err, "unable to patch Apacheweb deployment")
		return ctrl.Result{}, err
	}

	// Create ConfigMap template
	confMap, err = r.dependentConfmap(apacheWeb)
	if err != nil {
		logr.Error(err, "unable to create Apacheweb configMap")
		return ctrl.Result{}, err
	}

	// Patch ConfigMap resource
	err = r.Patch(ctx, &confMap, client.Apply, applyOpts...)
	if err != nil {
		logr.Error(err, "unable to patch Apacheweb configMap")
		return ctrl.Result{}, err
	}

	// Update ApacheWeb status
	err = r.Status().Update(ctx, &apacheWeb)
	if err != nil {
		logr.Error(err, "unable to update Apacheweb status")
		return ctrl.Result{}, err
	}

	logr.Info("finish apacheWeb reconciliation")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ApachewebReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Configure Event Recorder for Apacheweb resource
	r.recorder = mgr.GetEventRecorderFor("apacheweb-controller")

	return ctrl.NewControllerManagedBy(mgr).
		For(&apachewebv1alpha1.Apacheweb{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.ConfigMap{}).
		Complete(r)
}
