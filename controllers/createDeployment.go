package controllers

import (
	"apache-operator/api/v1alpha1"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

// generate Deployment resource from the given input
// Deployment resource is used for Apache HTTPD load balancer Pods
func (r *ApachewebReconciler) createDeployment(apacheWeb v1alpha1.Apacheweb, configMap corev1.ConfigMap) (appsv1.Deployment, error) {
	//checkSum := md5.Sum([]byte(cf.Data["httpd.conf"]))
	fmt.Printf("Version of new Configmap resource %s\n", configMap.GetResourceVersion())
	deployment := appsv1.Deployment{
		TypeMeta: v1.TypeMeta{
			APIVersion: appsv1.SchemeGroupVersion.String(),
			Kind:       "Deployment",
		},
		ObjectMeta: v1.ObjectMeta{
			Namespace: apacheWeb.Namespace,
			Name:      apacheWeb.Name,
			Labels:    map[string]string{"servername": apacheWeb.Spec.ServerName},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &apacheWeb.Spec.Size,
			Selector: &v1.LabelSelector{
				MatchLabels: map[string]string{"servername": apacheWeb.Spec.ServerName},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{
						"servername": apacheWeb.Spec.ServerName,
					},

					Annotations: map[string]string{
						"configMapVersion": configMap.GetResourceVersion(),
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "webserver",
							Image: "docker.io/httpd:latest",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 80,
									Name:          "http",
									Protocol:      "TCP",
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "httpd-conf",
									MountPath: "/usr/local/apache2/conf",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "httpd-conf",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: configMap.GetName(),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	if err := ctrl.SetControllerReference(&apacheWeb, &deployment, r.Scheme); err != nil {
		return deployment, err
	}

	return deployment, nil
}
