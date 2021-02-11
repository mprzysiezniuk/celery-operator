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

func getFrontendCommand() []string {
	command := []string{"python", "app.py"}
	return command
}

func (r *ReconcileCelery) deploymentForFrontend(cr *examplev1alpha1.Celery) *appsv1.Deployment {
	labels := map[string]string{
		"app": cr.Name,
	}
	matchlabels := map[string]string{
		"app":  cr.Name,
		"tier": "frontend",
	}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "celery-frontend",
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
						Image:   cr.Spec.FrontendImage,
						Name:    "frontend",
						Command: getFrontendCommand(),

						Ports: []corev1.ContainerPort{{
							ContainerPort: 5000,
							Name:          "frontend",
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

func (r *ReconcileCelery) serviceForFrontend(cr *examplev1alpha1.Celery) *corev1.Service {
	labels := map[string]string{
		"app": cr.Name,
	}
	matchlabels := map[string]string{
		"app":  cr.Name,
		"tier": "frontend",
	}

	ser := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "celery-frontend",
			Namespace: cr.Namespace,
			Labels:    labels,
		},

		Spec: corev1.ServiceSpec{
			Selector: matchlabels,

			Ports: []corev1.ServicePort{
				{
					Port: 5000,
					Name: cr.Name,
				},
			},
			Type: corev1.ServiceTypeNodePort,
		},
	}

	controllerutil.SetControllerReference(cr, ser, r.scheme)
	return ser
}

func (r *ReconcileCelery) isFrontendUp(v *examplev1alpha1.Celery) bool {
	deployment := &appsv1.Deployment{}

	err := r.client.Get(context.TODO(), types.NamespacedName{
		Name:      "celery-frontend",
		Namespace: v.Namespace,
	}, deployment)

	if err != nil {
		log.Error(err, "Deployment frontend not found")
		return false
	}
	if deployment.Status.ReadyReplicas == 1 {
		return true
	}

	return false

}
