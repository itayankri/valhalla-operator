/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/itayankri/valhalla-operator/internal/resource"
	"github.com/itayankri/valhalla-operator/internal/status"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientretry "k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/go-logr/logr"
	valhallav1alpha1 "github.com/itayankri/valhalla-operator/api/v1alpha1"
)

const finalizerName = "valhalla.itayankri/finalizer"

// ValhallaReconciler reconciles a Valhalla object
type ValhallaReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	log    logr.Logger
}

func NewValhallaReconciler(client client.Client, scheme *runtime.Scheme) *ValhallaReconciler {
	return &ValhallaReconciler{
		Client: client,
		Scheme: scheme,
		log:    ctrl.Log.WithName("controller").WithName("valhalla"),
	}
}

// +kubebuilder:rbac:groups=valhalla.itayankri,resources=valhallas,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=pods/exec,verbs=create
// +kubebuilder:rbac:groups="",resources=pods,verbs=update;get;list;watch
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update
// +kubebuilder:rbac:groups="batch",resources=jobs,verbs=get;list;watch;create;update
// +kubebuilder:rbac:groups="apps",resources=deployments,verbs=get;list;watch;create;update
// +kubebuilder:rbac:groups="autoscaling",resources=horizontalpodautoscalers,verbs=get;list;watch;create;update
// +kubebuilder:rbac:groups=valhalla.itayankri,resources=valhallas,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=valhalla.itayankri,resources=valhallas/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=valhalla.itayankri,resources=valhallas/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get;list;watch;create;update
// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=roles,verbs=get;list;watch;create;update
// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=rolebindings,verbs=get;list;watch;create;update

func (r *ValhallaReconciler) getValhallaInstance(ctx context.Context, namespacedName types.NamespacedName) (*valhallav1alpha1.Valhalla, error) {
	instance := &valhallav1alpha1.Valhalla{}
	err := r.Client.Get(ctx, namespacedName, instance)
	return instance, err
}

func (r *ValhallaReconciler) updateValhallaResource(ctx context.Context, instance *valhallav1alpha1.Valhalla) error {
	err := r.Client.Update(ctx, instance)
	if err != nil {
		return err
	}

	instance.Status.ObservedGeneration = instance.Generation
	return r.Client.Status().Update(ctx, instance)
}

func (r *ValhallaReconciler) getChildResources(ctx context.Context, instance *valhallav1alpha1.Valhalla) ([]runtime.Object, error) {
	pvc := &corev1.PersistentVolumeClaim{}
	if err := r.Client.Get(ctx, types.NamespacedName{
		Name:      instance.ChildResourceName(resource.PersistentVolumeClaimSuffix),
		Namespace: instance.Namespace,
	}, pvc); err != nil && !errors.IsNotFound(err) {
		return nil, err
	} else if errors.IsNotFound(err) {
		pvc = nil
	}

	job := &batchv1.Job{}
	if err := r.Client.Get(ctx, types.NamespacedName{
		Name:      instance.ChildResourceName(resource.JobSuffix),
		Namespace: instance.Namespace,
	}, job); err != nil && !errors.IsNotFound(err) {
		return nil, err
	} else if errors.IsNotFound(err) {
		job = nil
	}

	deployment := &appsv1.Deployment{}
	if err := r.Client.Get(ctx, types.NamespacedName{
		Name:      instance.ChildResourceName(resource.DeploymentSuffix),
		Namespace: instance.Namespace,
	}, deployment); err != nil && !errors.IsNotFound(err) {
		return nil, err
	} else if errors.IsNotFound(err) {
		deployment = nil
	}

	hpa := &autoscalingv1.HorizontalPodAutoscaler{}
	if err := r.Client.Get(ctx, types.NamespacedName{
		Name:      instance.ChildResourceName(resource.HorizontalPodAutoscalerSuffix),
		Namespace: instance.Namespace,
	}, hpa); err != nil && !errors.IsNotFound(err) {
		return nil, err
	} else if errors.IsNotFound(err) {
		hpa = nil
	}

	service := &corev1.Service{}
	if err := r.Client.Get(ctx, types.NamespacedName{
		Name:      instance.ChildResourceName(resource.ServiceSuffix),
		Namespace: instance.Namespace,
	}, service); err != nil && !errors.IsNotFound(err) {
		return nil, err
	} else if errors.IsNotFound(err) {
		service = nil
	}

	return []runtime.Object{pvc, job, deployment, hpa, service}, nil
}

func (r *ValhallaReconciler) initialize(ctx context.Context, instance *valhallav1alpha1.Valhalla) error {
	controllerutil.AddFinalizer(instance, finalizerName)
	return r.updateValhallaResource(ctx, instance)
}

func (r *ValhallaReconciler) updateValhallaStatusConditions(
	ctx context.Context,
	instance *valhallav1alpha1.Valhalla,
	childResources []runtime.Object,
) (time.Duration, error) {
	instance.Status.SetConditions(childResources)
	err := r.Client.Status().Update(ctx, instance)
	if err != nil {
		if errors.IsConflict(err) {
			r.log.Info("failed to update status because of conflict; requeueing...")
			return 2 * time.Second, nil
		}
		return 0, err
	}
	return 0, nil
}

// logAndRecordOperationResult - helper function to log and record events with message and error
// it logs and records 'updated' and 'created' OperationResult, and ignores OperationResult 'unchanged'
func (r *ValhallaReconciler) logOperationResult(
	logger logr.Logger,
	ro runtime.Object,
	resource runtime.Object,
	operationResult controllerutil.OperationResult,
	err error,
) {
	if operationResult == controllerutil.OperationResultNone && err == nil {
		return
	}

	var operation string
	if operationResult == controllerutil.OperationResultCreated {
		operation = "create"
	}

	if operationResult == controllerutil.OperationResultUpdated {
		operation = "update"
	}

	if err == nil {
		msg := fmt.Sprintf("%sd resource %s of Type %T", operation, resource.(metav1.Object).GetName(), resource.(metav1.Object))
		logger.Info(msg)
	}

	if err != nil {
		msg := fmt.Sprintf("failed to %s resource %s of Type %T", operation, resource.(metav1.Object).GetName(), resource.(metav1.Object))
		logger.Error(err, msg)
	}
}

func (r *ValhallaReconciler) cleanup(ctx context.Context, instance *valhallav1alpha1.Valhalla) error {
	if controllerutil.ContainsFinalizer(instance, finalizerName) {
		instance.Status.ObservedGeneration = instance.Generation
		instance.Status.SetCondition(metav1.Condition{
			Type:    status.ConditionAvailable,
			Status:  metav1.ConditionFalse,
			Reason:  "Cleanup",
			Message: "Deleting Valhalla resources",
		})

		err := r.Client.Status().Update(ctx, instance)
		if err != nil {
			return err
		}

		controllerutil.RemoveFinalizer(instance, finalizerName)

		err = r.Client.Update(ctx, instance)
		if err != nil {
			return err
		}
	}

	instance.Status.ObservedGeneration = instance.Generation
	err := r.Client.Status().Update(ctx, instance)
	if errors.IsConflict(err) || errors.IsNotFound(err) {
		// These errors are ignored. They can happen if the CR was removed
		// before the status update call is executed.
		return nil
	}
	return err
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Valhalla object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *ValhallaReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := r.log.WithValues("valhalla", req.NamespacedName)
	logger.Info("Starting reconciliation")

	instance, err := r.getValhallaInstance(ctx, req.NamespacedName)
	if err != nil {
		if errors.IsNotFound(err) {
			// Return and don't requeue
			logger.Info("Instance not found")
			return ctrl.Result{}, nil
		}

		// Error reading the object - requeue the request.
		logger.Error(err, "Failed to fetch Valhalla instance")
		return ctrl.Result{}, err
	}

	childResources, err := r.getChildResources(ctx, instance)
	if err != nil {
		logger.Error(err, "Failed to fetch child resources", instance.Namespace, instance.Name)
		return ctrl.Result{}, err
	}

	if requeueAfter, err := r.updateValhallaStatusConditions(ctx, instance, childResources); err != nil || requeueAfter > 0 {
		return ctrl.Result{RequeueAfter: requeueAfter}, err
	}

	if !isInitialized(instance) {
		err := r.initialize(ctx, instance)
		// No need to requeue here, because
		// the update will trigger reconciliation again
		logger.Info("Valhalla Instance initialized")
		return ctrl.Result{}, err
	}

	if isBeingDeleted(instance) {
		err := r.cleanup(ctx, instance)
		if err != nil {
			logger.Error(err, "Cleanup failed for rerouce: %v/%v", instance.Namespace, instance.Name)
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	if isPaused(instance) {
		if instance.Status.Paused {
			logger.Info("Valhalla operator is paused on resource: %v/%v", instance.Namespace, instance.Name)
			return ctrl.Result{}, nil
		}
		logger.Info(fmt.Sprintf("Pausing Valhalla operator on resource: %v/%v", instance.Namespace, instance.Name))
		instance.Status.Paused = true
		err := r.updateValhallaResource(ctx, instance)
		// instance.Status.ObservedGeneration = instance.Generation
		// err := r.Client.Status().Update(ctx, instance)
		return ctrl.Result{}, err
	}

	rawInstanceSpec, err := json.Marshal(instance.Spec)
	if err != nil {
		logger.Error(err, "Failed to marshal Valhalla instance spec")
	}

	logger.Info("Reconciling Valhalla instance", "spec", string(rawInstanceSpec))

	resourceBuilder := resource.ValhallaResourceBuilder{
		Instance: instance,
		Scheme:   r.Scheme,
	}

	builders := resourceBuilder.ResourceBuilders()

	for _, builder := range builders {
		if builder.ShouldDeploy(childResources) {
			resource, err := builder.Build()
			if err != nil {
				logger.Error(err, "Failed to build resource %v for Valhalla Instance %v/%v", builder, instance.Namespace, instance.Name)
				return ctrl.Result{}, err
			}

			var operationResult controllerutil.OperationResult
			err = clientretry.RetryOnConflict(clientretry.DefaultRetry, func() error {
				var apiError error
				operationResult, apiError = controllerutil.CreateOrUpdate(ctx, r.Client, resource, func() error {
					return builder.Update(resource)
				})
				return apiError
			})
			r.logOperationResult(logger, instance, resource, operationResult, err)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	logger.Info("Finished reconciling")
	return ctrl.Result{}, nil
}

func isInitialized(instance *valhallav1alpha1.Valhalla) bool {
	return controllerutil.ContainsFinalizer(instance, finalizerName)
}

func isBeingDeleted(object metav1.Object) bool {
	return !object.GetDeletionTimestamp().IsZero()
}

func isPaused(object metav1.Object) bool {
	if object.GetAnnotations() == nil {
		return false
	}
	pausedStr, ok := object.GetAnnotations()[valhallav1alpha1.OperatorPausedAnnotation]
	if !ok {
		return false
	}
	paused, err := strconv.ParseBool(pausedStr)
	if err != nil {
		return false
	}
	return paused
}

// SetupWithManager sets up the controller with the Manager.
func (r *ValhallaReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&valhallav1alpha1.Valhalla{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&batchv1.Job{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&autoscalingv1.HorizontalPodAutoscaler{}).
		Complete(r)
}
