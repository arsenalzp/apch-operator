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

// generate ConfigMap resource from the given input
// ConfigMap store Apace HTTPD configuration - httpd.conf file
// which is mounted to /usr/local/apache2/conf directory
func (r *ApachewebReconciler) createWebConfmap(apacheWeb v1alpha1.Apacheweb) (corev1.ConfigMap, error) {
	t := template.New(apacheWeb.Spec.Type)
	t, err := t.Parse(webTemplate)
	if err != nil {
		fmt.Printf("Unabel parse template %s", err)
		return corev1.ConfigMap{}, err
	}

	// Load balancer configuration
	webServerConfig := v1alpha1.WebServer{
		ServerPort:   apacheWeb.Spec.WebServer.ServerPort,
		DocumentRoot: apacheWeb.Spec.WebServer.DocumentRoot,
		ServerAdmin:  apacheWeb.Spec.WebServer.ServerAdmin,
	}

	var httpdConf bytes.Buffer
	if err := t.Execute(&httpdConf, webServerConfig); err != nil {
		fmt.Printf("Unabel execute parser %s", err)
		return corev1.ConfigMap{}, err
	}

	configMap := corev1.ConfigMap{
		TypeMeta: v1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "ConfigMap",
		},
		ObjectMeta: v1.ObjectMeta{
			Namespace: apacheWeb.Namespace,
			Name:      apacheWeb.Name,
			Labels: map[string]string{
				"servername": apacheWeb.Spec.ServerName,
			},
		},
		Data: map[string]string{
			"httpd.conf": httpdConf.String(),
		},
	}

	if err := ctrl.SetControllerReference(&apacheWeb, &configMap, r.Scheme); err != nil {
		return configMap, err
	}

	return configMap, nil
}
