package controllers

import (
	"apache-operator/api/v1alpha1"
	"bytes"
	"fmt"
	"text/template"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *ApachewebReconciler) dependentConfmap(aw v1alpha1.Apacheweb) (corev1.ConfigMap, error) {
	t := template.New(aw.Spec.Type)
	t, err := t.Parse(templateBody)
	if err != nil {
		fmt.Printf("Unabel parse template %s", err)
		return corev1.ConfigMap{}, err
	}

	var httpdConf bytes.Buffer
	if err := t.Execute(&httpdConf, aw); err != nil {
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
