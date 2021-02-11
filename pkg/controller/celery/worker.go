package celery

import (
	"context"

	examplev1alpha1 "celery-operator/pkg/apis/example/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func getWorkerCommand() []string {
	command := []string{"celery", "-A", "tasks", "worker", "--loglevel=info"}
	return command
}

func (r *ReconcileCelery) deploymentForCeleryWorker(cr *examplev1alpha1.Celery) *appsv1.Deployment {
	labels := map[string]string{
		"app": cr.Name,
	}
	matchlabels := map[string]string{
		"app":  cr.Name,
		"tier": "worker",
	}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "celery-worker",
			Namespace: cr.Namespace,
			Labels:    labels,
		},

		Spec: appsv1.DeploymentSpec{
			Replicas: &cr.Spec.WSize,
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
						Name:    "worker",
						Command: getWorkerCommand(),
					},
					},
				},
			},
		},
	}

	controllerutil.SetControllerReference(cr, dep, r.scheme)
	return dep
}

func (r *ReconcileCelery) isCeleryWorkerUp(cr *examplev1alpha1.Celery) bool {
	deployment := &appsv1.Deployment{}

	err := r.client.Get(context.TODO(), types.NamespacedName{
		Name:      "celery-worker",
		Namespace: cr.Namespace,
	}, deployment)

	if err != nil {
		log.Error(err, "Deployment worker not found")
		return false
	}
	if deployment.Status.ReadyReplicas == cr.Spec.WSize {
		return true
	}

	return false

}
