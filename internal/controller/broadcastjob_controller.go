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
	"fmt"
	"time"

	appsv1 "github.com/cloudrasayan/tugboat/api/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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
	log.Info("Reconciling BroadcastJob", "name", req.Name, "namespace", req.Namespace)
	// Fetch the BroadcastJob custom resource
	bdjobs := &appsv1.BroadcastJob{}
	err := r.Get(ctx, req.NamespacedName, bdjobs)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Error reconciling Pods")
			// Object not found, return. Created objects are automatically garbage collected.
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	if bdjobs.Status.Phase == appsv1.PhaseCompleted || bdjobs.Status.Phase == appsv1.PhaseFailed {
		// Do nothing
		return ctrl.Result{}, nil
	}

	// List all nodes in the cluster
	nodeList := &v1.NodeList{}
	err = r.List(ctx, nodeList, client.MatchingLabels(bdjobs.Spec.NodeSelector))
	if err != nil {
		log.Error(err, "Error reconciling Nodeslist")
		return ctrl.Result{}, err
	}

	// List all pods owned by this BroadcastJob instance
	podList := &v1.PodList{}
	err = r.List(ctx, podList, client.InNamespace(req.Namespace), client.MatchingFields{jobOwnerKey: req.Name}, client.MatchingLabels(bdjobs.Spec.Labels))
	if err != nil {
		log.Error(err, "Error reconciling Pods")
		return ctrl.Result{}, err
	}

	// Count the pods that are active, succeeded, failed or desired
	activePods := int32(0)
	succeededPods := int32(0)
	failedPods := int32(0)
	desiredPods := int32(len(nodeList.Items))
	// Iterate over the pod list and update the status of each pod
	for _, pod := range podList.Items {
		// Check if the pod has completed successfully
		if pod.Status.Phase == v1.PodSucceeded {
			// Update the pod status to completed
			pod.Status.Conditions = append(pod.Status.Conditions, v1.PodCondition{
				Type:               v1.PodReady,
				Status:             v1.ConditionTrue,
				LastTransitionTime: metav1.Now(),
				Reason:             "PodCompleted",
				Message:            "Pod has completed successfully",
			})
			// Update the pod status subresource
			err = r.Status().Update(ctx, &pod)
			// Parse the cleanupAfter value as a duration
			duration, err := time.ParseDuration(bdjobs.Spec.CleanupAfter)
			if err != nil {
				klog.Warningf("Invalid cleanupAfter value for Job %s/%s: %s", bdjobs.Namespace, bdjobs.Name, bdjobs.Spec.CleanupAfter)
				continue
			}
			time.AfterFunc(duration, func() {
				err = r.Delete(ctx, &pod)
				//if err != nil {
				//	panic(err)
				//}
			})

			if err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	for _, pod := range podList.Items {
		switch pod.Status.Phase {
		case v1.PodRunning, v1.PodPending:
			activePods++
		case v1.PodSucceeded:
			succeededPods++
		case v1.PodFailed:
			failedPods++
		}
	}

	// Update the status of the BroadcastJob
	bdjobs.Status.Active = activePods
	bdjobs.Status.Succeeded = succeededPods
	bdjobs.Status.Failed = failedPods
	bdjobs.Status.Desired = desiredPods
	if activePods == 0 && desiredPods == succeededPods {
		bdjobs.Status.Phase = appsv1.PhaseCompleted
		bdjobs.Status.CompletionTime = &metav1.Time{Time: time.Now()}
	} else {
		bdjobs.Status.Phase = appsv1.PhaseRunning
	}
	err = r.Status().Update(ctx, bdjobs)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Create pods on each node
	for _, node := range nodeList.Items {
		// Check if the pod already exists on this node
		podName := fmt.Sprintf("%s-%s", bdjobs.Name, node.Name)
		pod := &v1.Pod{}
		err = r.Get(ctx, client.ObjectKey{Namespace: bdjobs.Namespace, Name: podName}, pod)
		if err == nil {
			// Pod already exists, do nothing
			continue
		}
		if !errors.IsNotFound(err) {
			// Error reading the pod, requeue the request
			return ctrl.Result{}, err
		}

		// Create a new pod on this node
		pod = &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      podName,
				Namespace: bdjobs.Namespace,
				Labels:    bdjobs.Spec.Labels,
			},
			Spec: bdjobs.Spec.Template.Spec,
		}
		// Set the BroadcastJob as the owner of the pod
		if err := controllerutil.SetControllerReference(bdjobs, pod, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}
		pod.Spec.RestartPolicy = v1.RestartPolicyOnFailure
		// Set the node selector to this node
		pod.Spec.NodeSelector = bdjobs.Spec.NodeSelector
		// Create the pod
		err = r.Create(ctx, pod)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

var (
	jobOwnerKey = ".metadata.controller"
	apiGVStr    = appsv1.GroupVersion.String()
)

// SetupWithManager sets up the controller with the Manager.
func (r *BroadcastJobReconciler) SetupWithManager(mgr ctrl.Manager) error {

	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &v1.Pod{}, jobOwnerKey, func(rawObj client.Object) []string {
		// grab the job object, extract the owner...
		pod := rawObj.(*v1.Pod)
		owner := metav1.GetControllerOf(pod)
		if owner == nil {
			return nil
		}
		// ...make sure it's a CronJob...
		if owner.APIVersion != apiGVStr || owner.Kind != "BroadcastJob" {
			return nil
		}

		// ...and if so, return it
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.BroadcastJob{}).
		Owns(&v1.Pod{}).
		Complete(r)
}
