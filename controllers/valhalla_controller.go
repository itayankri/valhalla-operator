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

	"github.com/itayankri/valhalla-operator/internal/resource"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
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
		log:    ctrl.Log.WithName("controllers").WithName("Valhalla"),
	}
}

// +kubebuilder:rbac:groups=valhalla.ankri.io,resources=valhallas,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=pods/exec,verbs=create
// +kubebuilder:rbac:groups="",resources=pods,verbs=update;get;list;watch
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update
// +kubebuilder:rbac:groups="batch",resources=jobs,verbs=get;list;watch;create;update
// +kubebuilder:rbac:groups="apps",resources=deployments,verbs=get;list;watch;create;update
// +kubebuilder:rbac:groups="autoscaling",resources=horizontalpodautoscalers,verbs=get;list;watch;create;update
// +kubebuilder:rbac:groups="policy",resources=poddisruptionbudgets,verbs=get;list;watch;create;update
// +kubebuilder:rbac:groups="networking.k8s.io",resources=ingresses,verbs=get;list;watch;create;update
// +kubebuilder:rbac:groups="",resources=endpoints,verbs=get;watch;list
// +kubebuilder:rbac:groups=valhalla.ankri.io,resources=valhallas,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=valhalla.ankri.io,resources=valhallas/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=valhalla.ankri.io,resources=valhallas/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get;list;watch;create;update
// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=roles,verbs=get;list;watch;create;update
// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=rolebindings,verbs=get;list;watch;create;update

func (r *ValhallaReconciler) getValhallaInstance(ctx context.Context, namespacedName types.NamespacedName) (*valhallav1alpha1.Valhalla, error) {
	instance := &valhallav1alpha1.Valhalla{}
	err := r.Client.Get(ctx, namespacedName, instance)
	return instance, err
}

func (r *ValhallaReconciler) getJob(ctx context.Context, namespacedName types.NamespacedName) (*batchv1.Job, error) {
	job := &batchv1.Job{}
	err := r.Client.Get(ctx, namespacedName, job)
	return job, err
}

func (r *ValhallaReconciler) updateValhallaStatus(
	ctx context.Context,
	instance *valhallav1alpha1.Valhalla,
) {
	phaseCompleted, err := r.isPhaseComplete(ctx, instance)
	if err != nil {
		ctrl.LoggerFrom(ctx).Error(err, "Failed to fetch map builder job",
			"namespace", instance.Namespace,
			"name", instance.Name)
		return
	}

	if phaseCompleted {
		instance.Status.Phase = instance.Status.Phase.GetNextPhase()
	}

	if err = r.Status().Update(ctx, instance); err != nil {
		ctrl.LoggerFrom(ctx).Error(err, "Failed to update Custom Resource status",
			"namespace", instance.Namespace,
			"name", instance.Name)
	}
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
			return ctrl.Result{}, nil
		}

		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	if isBeingDeleted(instance) {
		return ctrl.Result{}, nil
	}

	if isPaused(instance) {
		if instance.Status.Paused {
			return ctrl.Result{}, nil
		}
		logger.Info(fmt.Sprintf("Pausing Valhalla operator on resource: %v/%v", instance.Namespace, instance.Name))
		instance.Status.Paused = true
		instance.Status.ObservedGeneration = instance.Generation
		err := r.Client.Status().Update(ctx, instance)
		return ctrl.Result{}, err
	}

	rawInstanceSpec, err := json.Marshal(instance.Spec)
	if err != nil {
		logger.Error(err, "Failed to marshal cluster spec")
	}

	logger.Info(fmt.Sprintf("Reconciling Valhalla instance - phase %d", instance.Status.Phase), "spec", string(rawInstanceSpec))

	resourceBuilder := resource.ValhallaResourceBuilder{
		Instance: instance,
		Scheme:   r.Scheme,
	}

	builders := resourceBuilder.ResourceBuilders(instance.Status.Phase)

	for _, builder := range builders {
		if builder.GetPhase() <= instance.Status.Phase {
			resource, err := builder.Build()
			if err != nil {
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
				r.updateValhallaStatus(ctx, instance)
				return ctrl.Result{}, err
			}
		}
	}

	r.updateValhallaStatus(ctx, instance)
	logger.Info("Finished reconciling")

	return ctrl.Result{}, nil
}

func (r *ValhallaReconciler) isPhaseComplete(ctx context.Context, instance *valhallav1alpha1.Valhalla) (bool, error) {
	if instance.Status.Phase == valhallav1alpha1.Empty {
		return true, nil
	}

	if instance.Status.Phase == valhallav1alpha1.MapBuilding {
		job, err := r.getJob(ctx, types.NamespacedName{
			Name:      instance.ChildResourceName("builder"),
			Namespace: instance.Namespace,
		})
		if err != nil {
			return false, err
		}

		for _, condition := range job.Status.Conditions {
			if condition.Type == "Complete" {
				return true, nil
			}
		}
	}

	return false, nil
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
		Owns(&appsv1.Deployment{}).
		Owns(&batchv1.Job{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&networkingv1.Ingress{}).
		Owns(&autoscalingv1.HorizontalPodAutoscaler{}).
		Owns(&policyv1beta1.PodDisruptionBudget{}).
		Complete(r)
}
