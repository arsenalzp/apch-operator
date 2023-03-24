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
func (r *ApachewebReconciler) createWebConfmap(aw v1alpha1.Apacheweb) (corev1.ConfigMap, error) {
	t := template.New(aw.Spec.Type)
	t, err := t.Parse(webTemplate)
	if err != nil {
		fmt.Printf("Unabel parse template %s", err)
		return corev1.ConfigMap{}, err
	}

	// Load balancer configuration
	webServerConfig := v1alpha1.WebServer{
		ServerPort:   aw.Spec.WebServer.ServerPort,
		DocumentRoot: aw.Spec.WebServer.DocumentRoot,
		ServerAdmin:  aw.Spec.WebServer.ServerAdmin,
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
			Namespace: aw.Namespace,
			Name:      aw.Name,
			Labels: map[string]string{
				"servername": aw.Spec.ServerName,
			},
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
