package celery

import (
	"context"
	"fmt"
	"time"

	examplev1alpha1 "celery-operator/pkg/apis/example/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_celery")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Celery Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileCelery{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("celery-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Celery
	err = c.Watch(&source.Kind{Type: &examplev1alpha1.Celery{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Celery
	// err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
	// 	IsController: true,
	// 	OwnerType:    &examplev1alpha1.Celery{},
	// })
	// if err != nil {
	// 	return err
	// }

	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &examplev1alpha1.Celery{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &examplev1alpha1.Celery{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.PersistentVolumeClaim{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &examplev1alpha1.Celery{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileCelery implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileCelery{}

// ReconcileCelery reconciles a Celery object
type ReconcileCelery struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Celery object and makes changes based on the state read
// and what is in the Celery.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileCelery) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Celery")

	// Fetch the Celery instance
	celery := &examplev1alpha1.Celery{}
	err := r.client.Get(context.TODO(), request.NamespacedName, celery)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	var result *reconcile.Result

	// === RabbitMQ ===

	result, err = r.ensurePVC(request, celery, r.pvcForRabbitMq(celery))
	if result != nil {
		return *result, err
	}

	result, err = r.ensureDeployment(request, celery, r.deploymentForRabbitMq(celery))
	if result != nil {
		return *result, err
	}
	result, err = r.ensureService(request, celery, r.serviceForRabbitMq(celery))
	if result != nil {
		return *result, err
	}

	rabbitmqRunning := r.isRabbitMqUp(celery)

	if !rabbitmqRunning {
		// If MySQL isn't running yet, requeue the reconcile
		// to run again after a delay
		delay := time.Second * time.Duration(5)

		log.Info(fmt.Sprintf("RabbitMQ isn't running, waiting for %s", delay))
		return reconcile.Result{RequeueAfter: delay}, nil
	}

	// === Worker ===

	result, err = r.ensureDeployment(request, celery, r.deploymentForCeleryWorker(celery))
	if result != nil {
		return *result, err
	}

	workerRunning := r.isCeleryWorkerUp(celery)

	if !workerRunning {
		// If worker isn't running yet, requeue the reconcile
		// to run again after a delay
		delay := time.Second * time.Duration(5)

		log.Info(fmt.Sprintf("Celery worker isn't running, waiting for %s", delay))
		return reconcile.Result{RequeueAfter: delay}, nil
	}

	// === Flower ===

	result, err = r.ensureDeployment(request, celery, r.deploymentForFlower(celery))
	if result != nil {
		return *result, err
	}

	result, err = r.ensureService(request, celery, r.serviceForFlower(celery))
	if result != nil {
		return *result, err
	}

	flowerRunning := r.isFlowerUp(celery)

	if !flowerRunning {
		// If worker isn't running yet, requeue the reconcile
		// to run again after a delay
		delay := time.Second * time.Duration(5)

		log.Info(fmt.Sprintf("Celery flower isn't running, waiting for %s", delay))
		return reconcile.Result{RequeueAfter: delay}, nil
	}

	// === Frontend ===

	result, err = r.ensureDeployment(request, celery, r.deploymentForFrontend(celery))
	if result != nil {
		return *result, err
	}
	result, err = r.ensureService(request, celery, r.serviceForFrontend(celery))
	if result != nil {
		return *result, err
	}

	frontendRunning := r.isFrontendUp(celery)

	if !frontendRunning {
		// If Frontend isn't running yet, requeue the reconcile
		// to run again after a delay
		delay := time.Second * time.Duration(5)

		log.Info(fmt.Sprintf("Frontend isn't running, waiting for %s", delay))
		return reconcile.Result{RequeueAfter: delay}, nil
	}

	return reconcile.Result{}, nil

}
