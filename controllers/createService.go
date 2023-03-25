package controllers

import (
	"apache-operator/api/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *ApachewebReconciler) createService(apacheWeb v1alpha1.Apacheweb) (corev1.Service, error) {
	// The port that will be exposed by this service
	var port int32

	if apacheWeb.Spec.WebServer != nil {
		port = *apacheWeb.Spec.WebServer.ServerPort
	}

	if apacheWeb.Spec.LoadBalancer != nil {
		port = *apacheWeb.Spec.LoadBalancer.ServerPort
	}

	service := corev1.Service{
		TypeMeta: v1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Service",
		},
		ObjectMeta: v1.ObjectMeta{
			Namespace: apacheWeb.Namespace,
			Name:      apacheWeb.Name,
			Labels:    map[string]string{"servername": apacheWeb.Spec.ServerName},
		},
		Spec: corev1.ServiceSpec{
			Type: "ClusterIP",
			Ports: []corev1.ServicePort{
				{
					Name:       "apacheweb-svc",
					Protocol:   "TCP",
					Port:       port,
					TargetPort: intstr.FromInt(int(port)),
				},
			},
			Selector: map[string]string{"servername": apacheWeb.Spec.ServerName},
		},
	}

	if err := ctrl.SetControllerReference(&apacheWeb, &service, r.Scheme); err != nil {
		return service, err
	}

	return service, nil
}
