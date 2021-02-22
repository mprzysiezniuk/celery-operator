package celery

import (
	examplev1alpha1 "celery-operator/pkg/apis/example/v1alpha1"
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func getFlowerCommand() []string {
	command := []string{"celery", "-A", "tasks", "flower"}
	return command
}

func (r *ReconcileCelery) deploymentForFlower(cr *examplev1alpha1.Celery) *appsv1.Deployment {
	labels := map[string]string{
		"app": cr.Name,
	}
	matchlabels := map[string]string{
		"app":  cr.Name,
		"tier": "flower",
	}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "celery-flower",
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
						Image:   cr.Spec.WorkerImage,
						Name:    "flower",
						Command: getFlowerCommand(),

						Ports: []corev1.ContainerPort{{
							ContainerPort: 5555,
							Name:          "flower",
						}},
					},
					},
				},
			},
		},
	}

	controllerutil.SetControllerReference(cr, dep, r.scheme)
	return dep
}

func (r *ReconcileCelery) serviceForFlower(cr *examplev1alpha1.Celery) *corev1.Service {
	labels := map[string]string{
		"app": cr.Name,
	}
	matchlabels := map[string]string{
		"app":  cr.Name,
		"tier": "flower",
	}

	ser := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "celery-flower",
			Namespace: cr.Namespace,
			Labels:    labels,
		},

		Spec: corev1.ServiceSpec{
			Selector: matchlabels,

			Ports: []corev1.ServicePort{
				{
					Port: 5555,
					Name: cr.Name,
				},
			},
			Type: corev1.ServiceTypeNodePort,
		},
	}

	controllerutil.SetControllerReference(cr, ser, r.scheme)
	return ser
}

func (r *ReconcileCelery) isFlowerUp(v *examplev1alpha1.Celery) bool {
	deployment := &appsv1.Deployment{}

	err := r.client.Get(context.TODO(), types.NamespacedName{
		Name:      "celery-flower",
		Namespace: v.Namespace,
	}, deployment)

	if err != nil {
		log.Error(err, "Deployment flower not found")
		return false
	}
	if deployment.Status.ReadyReplicas == 1 {
		return true
	}

	return false

}
