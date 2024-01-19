/*
Copyright 2024.

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

package controller

import (
	"context"
	appsv1 "github.com/cloudrasayan/tugboat/api/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// BroadcastJobReconciler reconciles a BroadcastJob object
type BroadcastJobReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=apps.tugboat.cloudrasayan.com,resources=broadcastjobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.tugboat.cloudrasayan.com,resources=broadcastjobs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps.tugboat.cloudrasayan.com,resources=broadcastjobs/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=nodes,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=pods/status,verbs=get;update;patch
// move the current state of the cluster closer to the desired state.
// the BroadcastJob object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *BroadcastJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	// Fetch the BroadcastJob custom resource
	job := &appsv1.BroadcastJob{}
	err := r.Get(ctx, req.NamespacedName, job)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if job.Status.Phase == "" {
		job.Status.Phase = appsv1.PhaseRunning
	}

	// Get a list of available nodes
	nodes := &corev1.NodeList{}
	err = r.List(ctx, nodes)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Iterate through nodes and create jobs with restart limit
	for _, node := range nodes.Items {
		job := &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: req.Namespace, // Set the desired namespace here
				Name:      node.Name,
			},
			Spec: batchv1.JobSpec{
				Template:     job.Spec.Template,
				BackoffLimit: &job.Spec.RestartLimit,
			},
		}
		job.Spec.Template.Spec.RestartPolicy = "Never"
		job.Spec.Template.Spec.NodeName = node.Name
		log.Info("found node")
		// Create or update the job
		err = r.Create(ctx, job)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *BroadcastJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.BroadcastJob{}).
		Complete(r)
}
