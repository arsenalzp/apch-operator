package controllers

import (
	"apache-operator/api/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *ApachewebReconciler) dependentDeployment(aw v1alpha1.Apacheweb, cf corev1.ConfigMap) (appsv1.Deployment, error) {
	deployment := appsv1.Deployment{
		TypeMeta: v1.TypeMeta{
			APIVersion: appsv1.SchemeGroupVersion.String(),
			Kind:       "Deployment",
		},
		ObjectMeta: v1.ObjectMeta{
			Namespace: aw.Namespace,
			Name:      aw.Name,
			Labels:    map[string]string{"servername": aw.Spec.ServerName},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &aw.Spec.Size,
			Selector: &v1.LabelSelector{
				MatchLabels: map[string]string{"servername": aw.Spec.ServerName},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{"servername": aw.Spec.ServerName},
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
										Name: cf.Name,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	if err := ctrl.SetControllerReference(&aw, &deployment, r.Scheme); err != nil {
		return deployment, err
	}

	return deployment, nil
}