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
	"fmt"
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	discovery "k8s.io/api/discovery/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
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
	var apacheWebSvc corev1.Service

	logr := log.FromContext(ctx).WithValues("ApacheWeb", req.NamespacedName)

	logr.Info("start ApacheWeb reconciliation")

	// Get Apacheweb resource
	if err := r.Get(ctx, req.NamespacedName, &apacheWeb); err != nil {
		if errors.IsNotFound(err) {
			logr.Info("ApacheWeb resource not found. Seems, it was deleted.")
			return ctrl.Result{}, nil
		}

		logr.Error(err, "unable to fetch ApacheWeb resource")
		return ctrl.Result{}, err
	}

	if err := r.Get(ctx, req.NamespacedName, &apacheWebSvc); err != nil && !errors.IsNotFound(err) {
		logr.Error(err, "unable to fetch ApacheWeb Service")
		return ctrl.Result{}, err
	}

	// Switch between Loadbalancer and Web modes
	switch apacheWeb.Spec.Type {
	case "lb":
		var endPointSliceList discovery.EndpointSliceList

		// Define labels selector for options of list request
		labelsSelector, err := labels.Parse("kubernetes.io/service-name=" + apacheWeb.Spec.LoadBalancer.BackEndService)
		if err != nil {
			logr.Error(err, "unable to parse labels that matches EndpoinsSlice labels")
			return ctrl.Result{}, err
		}

		// Define options for a list request
		listOptions := client.ListOptions{
			LabelSelector: labelsSelector,
			Namespace:     req.Namespace,
		}

		// Get the list of EndpointSlices
		if err := r.List(ctx, &endPointSliceList, &listOptions); err != nil {
			logr.Error(err, "unable to retrieve EndPointSlice list")
			return ctrl.Result{}, err
		}

		// Generate array of v1alpha1.EndPoint
		endPointsList, endPointsLisVers := genBackEndsList(apacheWeb.Spec.LoadBalancer.BackEndService, apacheWeb.Spec.LoadBalancer.Proto, endPointSliceList)

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

		if configMap.Annotations["endPointSliceVersion"] == endPointsLisVers &&
			configMap.Annotations["apacheWebGeneration"] == strconv.FormatInt(apacheWeb.GetGeneration(), 10) {
			return ctrl.Result{}, nil
		}

		// Generate ConfigMap template
		newConfMap, err := r.createLbConfmap(apacheWeb, endPointsList, endPointsLisVers)
		if err != nil {
			logr.Error(err, "unable to generate Apacheweb configMap")
			return ctrl.Result{}, err
		}

		applyOpts := []client.PatchOption{client.ForceOwnership, client.FieldOwner("ApacheWeb")}

		// Patch ConfigMap resource
		if err := r.Patch(ctx, &newConfMap, client.Apply, applyOpts...); err != nil {
			logr.Error(err, "unable to patch Apacheweb configMap")
			return ctrl.Result{}, err
		}

		// Generate Deployment resource
		deployment, err = r.createDeployment(apacheWeb, newConfMap)
		if err != nil {
			logr.Error(err, "unable to generate Apacheweb deployment")
			return ctrl.Result{}, err
		}

		// Patch Deployment object
		if err := r.Patch(ctx, &deployment, client.Apply, applyOpts...); err != nil {
			logr.Error(err, "unable to patch Apacheweb deployment")
			return ctrl.Result{}, err
		}

		// Update Apacheweb status with new endpoints list
		apacheWeb.Status.EndPoints = endPointsList
		for _, ep := range endPointsList {
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

		applyOpts := []client.PatchOption{client.ForceOwnership, client.FieldOwner("ApacheWeb")}

		// Generate Deployment resource
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
		if err := r.Patch(ctx, &deployment, client.Apply, applyOpts...); err != nil {
			logr.Error(err, "unable to patch Apacheweb deployment")
			return ctrl.Result{}, err
		}

		apacheWeb.Status.WebServer = &v1alpha1.WebServer{
			DocumentRoot: apacheWeb.Spec.WebServer.DocumentRoot,
			ServerPort:   apacheWeb.Spec.WebServer.ServerPort,
			ServerAdmin:  apacheWeb.Spec.WebServer.ServerAdmin,
		}

		r.recorder.Eventf(&apacheWeb, "Normal", "Created", "Apache Web Server created: ServerPort %d, ServerName %s, DocumentRoot %s", *apacheWeb.Spec.WebServer.ServerPort, apacheWeb.Spec.ServerName, apacheWeb.Spec.WebServer.DocumentRoot)
	}

	// Generate Service resource
	newService, err := r.createService(apacheWeb)
	if err != nil {
		logr.Error(err, "unable to generate Apacheweb Service")
		return ctrl.Result{}, err
	}

	applyOpts := []client.PatchOption{client.ForceOwnership, client.FieldOwner("ApacheWeb")}

	// Patch Apache Web Service
	if err := r.Patch(ctx, &newService, client.Apply, applyOpts...); err != nil {
		logr.Error(err, "unable to patch Apacheweb Service")
		return ctrl.Result{}, err
	}

	// Update ApacheWeb status
	if err := r.Status().Update(ctx, &apacheWeb); err != nil {
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
			if apacheWeb.Spec.LoadBalancer == nil || apacheWeb.Spec.LoadBalancer.BackEndService == "" {
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
		Owns(&corev1.Service{}).
		Watches(
			&source.Kind{Type: &discovery.EndpointSlice{}},
			handler.EnqueueRequestsFromMapFunc(r.getApacheWebWithEndPoints),
		).
		WithEventFilter(customPredicate()).
		Complete(r)
}

// Function defines custom predicates for Create, Update and Delete events.\
func customPredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return true
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			if _, ok := e.ObjectNew.(*discovery.EndpointSlice); ok {
				return e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration()
			}

			if _, ok := e.ObjectNew.(*v1alpha1.Apacheweb); ok {
				return e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration()
			}

			return false
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return !e.DeleteStateUnknown
		},
	}
}
