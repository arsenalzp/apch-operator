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
	"crypto/md5"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	discovery "k8s.io/api/discovery/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"apache-operator/api/v1alpha1"
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
//+kubebuilder:rbac:groups=core,resources=service,verbs=get;list;watch
//+kubebuilder:rbac:groups=discovery,resources=endpointslice,verbs=get;list;watch

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
	var configMap corev1.ConfigMap
	var deployment appsv1.Deployment
	var apacheWeb v1alpha1.Apacheweb

	logr := log.FromContext(ctx).WithValues("ApacheWeb", req.NamespacedName)

	logr.Info("start ApacheWeb reconciliation")

	// Get Apacheweb resource
	if err := r.Get(ctx, req.NamespacedName, &apacheWeb); err != nil {
		if errors.IsNotFound(err) {
			logr.Error(err, "ApacheWeb object not found")
			return ctrl.Result{}, err
		}

		logr.Error(err, "unable to fetch ApacheWeb")
		return ctrl.Result{}, err
	}

	switch apacheWeb.Spec.Type {
	case "lb":
		var endPointSliceList discovery.EndpointSliceList

		// Get the list of EndpointSlices
		if err := r.List(ctx, &endPointSliceList, client.InNamespace(req.Namespace)); err != nil {
			logr.Error(err, "unable to retrieve EndPointSlice list")
			return ctrl.Result{}, err
		}

		// Generate array of v1alpha1.EndPoint
		epList := genBackEndsList(apacheWeb.Spec.LoadBalancer.BackEndService, apacheWeb.Spec.LoadBalancer.Proto, endPointSliceList)

		// Get the Deployment objects
		if err := r.Get(ctx, req.NamespacedName, &deployment); err != nil && !errors.IsNotFound(err) {
			logr.Error(err, "unable to get Deployment object")
			return ctrl.Result{}, err
		}

		// Get the ConfigMap object
		if err := r.Get(ctx, req.NamespacedName, &configMap); err != nil && !errors.IsNotFound(err) {
			logr.Error(err, "unable to get ConfigMap object")
			return ctrl.Result{}, err
		}

		// Generate ConfigMap template
		newConfMap, err := r.createLbConfmap(apacheWeb, epList)
		if err != nil {
			logr.Error(err, "unable to generate Apacheweb configMap")
			return ctrl.Result{}, err
		}

		// If an old and a new ConfigMap object are equal then stop reconciliation process
		if isConfiMapsEqual(&configMap, &newConfMap) {
			logr.Info("ConfigMap wasn't changed, nothing to patch, finishing apacheWeb reconciliation")
			return ctrl.Result{}, nil
		}

		applyOpts := []client.PatchOption{client.ForceOwnership, client.FieldOwner("ApacheWeb")}

		// Generate Deployment template
		deployment, err = r.createDeployment(apacheWeb, newConfMap)
		if err != nil {
			logr.Error(err, "unable to generate Apacheweb deployment")
			return ctrl.Result{}, err
		}

		// Patch ConfigMap resource
		if err := r.Patch(ctx, &newConfMap, client.Apply, applyOpts...); err != nil {
			logr.Error(err, "unable to patch Apacheweb configMap")
			return ctrl.Result{}, err
		}

		// Patch Deployment object
		err = r.Patch(ctx, &deployment, client.Apply, applyOpts...)
		if err != nil {
			logr.Error(err, "unable to patch Apacheweb deployment")
			return ctrl.Result{}, err
		}

		apacheWeb.Status.EndPoints = epList
		for _, ep := range epList {
			if !ep.Status {
				r.recorder.Eventf(&apacheWeb, "Warning", "Deleted", "EndPoint with IPAddress %s, port %d, protocol %s, has status %t and is being deleted from the list", ep.IPAddress, ep.Port, ep.Proto, ep.Status)
			}

			r.recorder.Eventf(&apacheWeb, "Normal", "Created", "EndPoint added IPAddress %s, port %d, protocol %s, status %t", ep.IPAddress, ep.Port, ep.Proto, ep.Status)
		}

	case "web":
		// Get the Deployment objects
		if err := r.Get(ctx, req.NamespacedName, &deployment); err != nil && !errors.IsNotFound(err) {
			logr.Error(err, "unable to get Deployment object")
			return ctrl.Result{}, err
		}

		// Get the ConfigMap object
		if err := r.Get(ctx, req.NamespacedName, &configMap); err != nil && !errors.IsNotFound(err) {
			logr.Error(err, "unable to get ConfigMap object")
			return ctrl.Result{}, err
		}

		// Generate ConfigMap template
		newConfMap, err := r.createWebConfmap(apacheWeb)
		if err != nil {
			logr.Error(err, "unable to generate Apacheweb configMap")
			return ctrl.Result{}, err
		}

		// If an old and a new ConfigMap object are equal then stop reconciliation process
		if isConfiMapsEqual(&configMap, &newConfMap) {
			logr.Info("ConfigMap wasn't changed, nothing to patch, finishing apacheWeb reconciliation")
			return ctrl.Result{}, nil
		}

		applyOpts := []client.PatchOption{client.ForceOwnership, client.FieldOwner("ApacheWeb")}

		// Generate Deployment template
		deployment, err = r.createDeployment(apacheWeb, newConfMap)
		if err != nil {
			logr.Error(err, "unable to generate Apacheweb deployment")
			return ctrl.Result{}, err
		}

		// Patch ConfigMap resource
		if err := r.Patch(ctx, &newConfMap, client.Apply, applyOpts...); err != nil {
			logr.Error(err, "unable to patch Apacheweb configMap")
			return ctrl.Result{}, err
		}

		// Patch Deployment object
		err = r.Patch(ctx, &deployment, client.Apply, applyOpts...)
		if err != nil {
			logr.Error(err, "unable to patch Apacheweb deployment")
			return ctrl.Result{}, err
		}

		r.recorder.Eventf(&apacheWeb, "Normal", "Created", "Apache Web Server with Port %d, ServerName %s, DocumentRoot %s was created", apacheWeb.Spec.ServerPort, apacheWeb.Spec.ServerName, apacheWeb.Spec.WebServer.DocumentRoot)
	}

	// Update ApacheWeb status
	err := r.Status().Update(ctx, &apacheWeb)
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

	err := mgr.GetFieldIndexer().IndexField(
		context.Background(),
		&v1alpha1.Apacheweb{},
		".spec.loadBalancer.backEndService",
		func(rawObj client.Object) []string {
			apacheWeb := rawObj.(*v1alpha1.Apacheweb)
			if apacheWeb.Spec.LoadBalancer.BackEndService == "" {
				return nil
			}

			return []string{apacheWeb.Spec.LoadBalancer.BackEndService}
		})

	if err != nil {
		fmt.Printf("Index error %s\n", err)
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Apacheweb{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.ConfigMap{}).
		Watches(
			&source.Kind{Type: &discovery.EndpointSlice{}},
			handler.EnqueueRequestsFromMapFunc(r.getApacheWebWithEndPoints),
		).
		Complete(r)
}

func isConfiMapsEqual(oldConfMap, newConfMap *corev1.ConfigMap) bool {
	return md5.Sum([]byte(oldConfMap.Data["httpd.conf"])) == md5.Sum([]byte(newConfMap.Data["httpd.conf"]))
}
