package celery

import (
	"context"

	examplev1alpha1 "celery-operator/pkg/apis/example/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *ReconcileCelery) deploymentForRabbitMq(cr *examplev1alpha1.Celery) *appsv1.Deployment {
	labels := map[string]string{
		"app": cr.Name,
	}
	matchlabels := map[string]string{
		"app":  cr.Name,
		"tier": "rabbitmq",
	}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "celery-rabbitmq",
			Namespace: cr.Namespace,
			Labels:    labels,
		},

		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: matchlabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: matchlabels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: cr.Spec.BrokerImage,
						Name:  "rabbitmq",

						Ports: []corev1.ContainerPort{{
							ContainerPort: 5672,
							Name:          "rabbitmq",
						}},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "rabbitmq-persistent-storage",
								MountPath: "/var/lib/rabbitmq",
							},
						},
					},
					},

					Volumes: []corev1.Volume{
						{
							Name: "rabbitmq-persistent-storage",
							VolumeSource: corev1.VolumeSource{

								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "rabbitmq-pv-claim",
								},
							},
						},
					},
				},
			},
		},
	}

	controllerutil.SetControllerReference(cr, dep, r.scheme)
	return dep
}

func (r *ReconcileCelery) serviceForRabbitMq(cr *examplev1alpha1.Celery) *corev1.Service {
	labels := map[string]string{
		"app": cr.Name,
	}
	matchlabels := map[string]string{
		"app":  cr.Name,
		"tier": "rabbitmq",
	}

	ser := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "celery-rabbitmq",
			Namespace: cr.Namespace,
			Labels:    labels,
		},

		Spec: corev1.ServiceSpec{
			Selector: matchlabels,

			Ports: []corev1.ServicePort{
				{
					Port: 5672,
					Name: cr.Name,
				},
			},
			Type: corev1.ServiceTypeNodePort,
		},
	}

	controllerutil.SetControllerReference(cr, ser, r.scheme)
	return ser
}

func (r *ReconcileCelery) pvcForRabbitMq(cr *examplev1alpha1.Celery) *corev1.PersistentVolumeClaim {
	labels := map[string]string{
		"app": cr.Name,
	}

	pvc := &corev1.PersistentVolumeClaim{

		ObjectMeta: metav1.ObjectMeta{
			Name:      "rabbitmq-pv-claim",
			Namespace: cr.Namespace,
			Labels:    labels,
		},

		Spec: corev1.PersistentVolumeClaimSpec{

			AccessModes: []corev1.PersistentVolumeAccessMode{
				"ReadWriteOnce",
			},

			Resources: corev1.ResourceRequirements{
				Requests: map[corev1.ResourceName]resource.Quantity{
					corev1.ResourceStorage: resource.MustParse("1Gi"),
				},
			},
		},
	}

	controllerutil.SetControllerReference(cr, pvc, r.scheme)
	return pvc
}

func (r *ReconcileCelery) isRabbitMqUp(v *examplev1alpha1.Celery) bool {
	deployment := &appsv1.Deployment{}

	err := r.client.Get(context.TODO(), types.NamespacedName{
		Name:      "celery-rabbitmq",
		Namespace: v.Namespace,
	}, deployment)

	if err != nil {
		log.Error(err, "Deployment mysql not found")
		return false
	}
	if deployment.Status.ReadyReplicas == 1 {
		return true
	}

	return false

}
