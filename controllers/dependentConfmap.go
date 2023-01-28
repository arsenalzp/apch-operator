package controllers

import (
	"apache-operator/api/v1alpha1"
	"bytes"
	"fmt"
	"text/template"

	corev1 "k8s.io/api/core/v1"
	discovery "k8s.io/api/discovery/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

type LoadBalancer struct {
	EndPointsList []v1alpha1.EndPoint
	Path          string
	Type          string
	ServerPort    int32
}

func (r *ApachewebReconciler) dependentConfmap(aw v1alpha1.Apacheweb, es discovery.EndpointSlice) (corev1.ConfigMap, error) {
	var be LoadBalancer

	t := template.New(aw.Spec.Type)
	t, err := t.Parse(templateBody)
	if err != nil {
		fmt.Printf("Unabel parse template %s", err)
		return corev1.ConfigMap{}, err
	}

	// Load balancer configuration
	be = LoadBalancer{
		EndPointsList: genBackEndsList(aw.Spec.LoadBalancer.Proto, es),
		Path:          aw.Spec.LoadBalancer.Path,
		Type:          aw.Spec.Type,
		ServerPort:    aw.Spec.ServerPort,
	}

	var httpdConf bytes.Buffer
	if err := t.Execute(&httpdConf, be); err != nil {
		fmt.Printf("Unabel execute parser %s", err)
		return corev1.ConfigMap{}, err
	}

	configMap := corev1.ConfigMap{
		TypeMeta: v1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "ConfigMap",
		},
		ObjectMeta: v1.ObjectMeta{
			Namespace: aw.Namespace,
			Name:      aw.Name,
			Labels:    map[string]string{"servername": aw.Spec.ServerName},
		},
		Data: map[string]string{
			"httpd.conf": httpdConf.String(),
		},
	}

	if err := ctrl.SetControllerReference(&aw, &configMap, r.Scheme); err != nil {
		return configMap, err
	}

	return configMap, nil
}
