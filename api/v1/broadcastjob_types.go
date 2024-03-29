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

package v1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// BroadcastJobSpec defines the desired state of BroadcastJob
type BroadcastJobSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of BroadcastJob. Edit broadcastjob_types.go to remove/update
	Template v1.PodTemplateSpec `json:"template" protobuf:"bytes,2,opt,name=template"`
	//RestartLimit int32              `json:"restartLimit,omitempty" protobuf:"varint,2,opt,name=restartLimit"`
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	Labels       map[string]string `json:"labels,omitempty"`
	// +kubebuilder:validation:Format=duration
	CleanupAfter string `json:"cleanupafter,omitempty"`
}

// BroadcastJobStatus defines the observed state of BroadcastJob
type BroadcastJobStatus struct {
	// The latest available observations of an object's current state.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []JobCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`

	// Represents time when the job was acknowledged by the job controller.
	// It is not guaranteed to be set in happens-before order across separate operations.
	// It is represented in RFC3339 form and is in UTC.
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty" protobuf:"bytes,2,opt,name=startTime"`

	// Represents time when the job was completed. It is not guaranteed to
	// be set in happens-before order across separate operations.
	// It is represented in RFC3339 form and is in UTC.
	// +optional
	CompletionTime *metav1.Time `json:"completionTime,omitempty" protobuf:"bytes,3,opt,name=completionTime"`

	// The number of actively running pods.
	// +optional
	Active int32 `json:"active" protobuf:"varint,4,opt,name=active"`

	// The number of pods which reached phase Succeeded.
	// +optional
	Succeeded int32 `json:"succeeded" protobuf:"varint,5,opt,name=succeeded"`

	// The number of pods which reached phase Failed.
	// +optional
	Failed int32 `json:"failed" protobuf:"varint,6,opt,name=failed"`

	// The desired number of pods, this is typically equal to the number of nodes satisfied to run pods.
	// +optional
	Desired int32 `json:"desired" protobuf:"varint,7,opt,name=desired"`

	// The phase of the job.
	// +optional
	Phase BroadcastJobPhase `json:"phase" protobuf:"varint,8,opt,name=phase"`
}

// JobConditionType indicates valid conditions type of a job
type JobConditionType string

// These are valid conditions of a job.
const (
	// JobComplete means the job has completed its execution. A complete job means pods have been deployed on all
	// eligible nodes and all pods have reached succeeded or failed state. Note that the eligible nodes are defined at
	// the beginning of a reconciliation loop. If there are more nodes added within a reconciliation loop, those nodes will
	// not be considered to run pods.
	JobComplete JobConditionType = "Complete"

	// JobFailed means the job has failed its execution. A failed job means the job has either exceeded the
	// ActiveDeadlineSeconds limit, or the aggregated number of container restarts for all pods have exceeded the RestartLimit.
	JobFailed JobConditionType = "Failed"
)

// JobCondition describes current state of a job.
type JobCondition struct {
	// Type of job condition, Complete or Failed.
	Type JobConditionType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=JobConditionType"`
	// Status of the condition, one of True, False, Unknown.
	Status v1.ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status,casttype=k8s.io/api/core/v1.ConditionStatus"`
	// Last time the condition was checked.
	// +optional
	LastProbeTime metav1.Time `json:"lastProbeTime,omitempty" protobuf:"bytes,3,opt,name=lastProbeTime"`
	// Last time the condition transit from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,4,opt,name=lastTransitionTime"`
	// (brief) reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty" protobuf:"bytes,5,opt,name=reason"`
	// Human readable message indicating details about last transition.
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,6,opt,name=message"`
}

// BroadcastJobPhase indicates the phase of the job.
type BroadcastJobPhase string

const (
	// PhaseCompleted means the job is completed.
	PhaseCompleted BroadcastJobPhase = "completed"

	// PhaseRunning means the job is running.
	PhaseRunning BroadcastJobPhase = "running"

	// PhasePaused means the job is paused.
	PhasePaused BroadcastJobPhase = "paused"

	// PhaseFailed means the job is failed.
	PhaseFailed BroadcastJobPhase = "failed"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// BroadcastJob is the Schema for the broadcastjobs API
type BroadcastJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BroadcastJobSpec   `json:"spec,omitempty"`
	Status BroadcastJobStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// BroadcastJobList contains a list of BroadcastJob
type BroadcastJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BroadcastJob `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BroadcastJob{}, &BroadcastJobList{})
}
