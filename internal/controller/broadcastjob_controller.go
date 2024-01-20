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
	appsv1 "github.com/cloudrasayan/tugboat/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
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
	bdjobs := &appsv1.BroadcastJob{}
	klog.Info("broadcastjob", bdjobs)
	klog.Info(req.Name, req.Namespace)
	fmt.Println("Hello World!")

	err := r.Get(ctx, req.NamespacedName, bdjobs)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// List all nodes in the cluster
	nodeList := &corev1.NodeList{}
	err = r.List(ctx, nodeList, client.MatchingLabels(bdjobs.Spec.NodeSelector))
	if err != nil {
		log.Info("Checking nodelist")
		return ctrl.Result{}, err
	}

	// List all pods owned by this BroadcastJob instance
	podList := &corev1.PodList{}
	err = r.List(ctx, podList, client.InNamespace(req.Namespace), client.MatchingLabels(bdjobs.Spec.Labels))
	if err != nil {
		log.Info("Checking podlist")
		return ctrl.Result{}, err
	}

	// Create a pod for each node that does not have one
	for _, node := range nodeList.Items {
		if !hasPodOnNode(podList, node.Name) {
			pod := newPodForCR(bdjobs, node.Name)
			err = r.Create(ctx, pod)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	// Delete any extra pods that have a different node name
	for _, pod := range podList.Items {
		if !hasNodeInList(nodeList, pod.Spec.NodeName) {
			err = r.Delete(ctx, &pod)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{}, nil
}

// newPodForCR returns a pod with the same spec as the BroadcastJob instance
func newPodForCR(cr *appsv1.BroadcastJob, nodeName string) *corev1.Pod {
	labels := cr.Spec.Labels
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", cr.Name, nodeName),
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers:         cr.Spec.Template.Spec.Containers,
			Volumes:            cr.Spec.Template.Spec.Volumes,
			ServiceAccountName: cr.Spec.Template.Spec.ServiceAccountName,
			NodeName:           nodeName,
		},
	}
}

// hasPodOnNode returns true if the pod list contains a pod that is scheduled on the given node
func hasPodOnNode(podList *corev1.PodList, nodeName string) bool {
	for _, pod := range podList.Items {
		if pod.Spec.NodeName == nodeName {
			return true
		}
	}
	return false
}

// hasNodeInList returns true if the node list contains a node with the given name
func hasNodeInList(nodeList *corev1.NodeList, nodeName string) bool {
	for _, node := range nodeList.Items {
		if node.Name == nodeName {
			return true
		}
	}
	return false
}

// SetupWithManager sets up the controller with the Manager.
func (r *BroadcastJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.BroadcastJob{}).
		Complete(r)
}
