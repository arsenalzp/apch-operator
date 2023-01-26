package controllers

import (
	"apache-operator/api/v1alpha1"
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *ApachewebReconciler) apacheWebUsingEndPoints(es client.Object) []ctrl.Request {
	listOptions := []client.ListOption{
		// matching our index
		client.MatchingFields{".spec.loadBalancer.backEndService": es.GetLabels()["kubernetes.io/service-name"]},
		// in the right namespace
		client.InNamespace(es.GetNamespace()),
	}

	var apacheWebList v1alpha1.ApachewebList
	if err := r.List(context.Background(), &apacheWebList, listOptions...); err != nil {
		fmt.Printf("error getting list of resources which use EndpointSlice "+es.GetName()+"\n", err)
		return nil
	}

	result := make([]ctrl.Request, len(apacheWebList.Items))
	for i, item := range apacheWebList.Items {
		result[i].Name = item.Name
		result[i].Namespace = item.Namespace
	}

	return result
}
