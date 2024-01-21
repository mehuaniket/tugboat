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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CronBroadcastJobSpec defines the desired state of CronBroadcastJob
type CronBroadcastJobSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of CronBroadcastJob. Edit cronbroadcastjob_types.go to remove/update
	Schedule             string           `json:"schedule" protobuf:"bytes,1,opt,name=schedule"`
	BroadcastJobTemplate BroadcastJobSpec `json:"broadcastJobTemplate,omitempty" protobuf:"bytes,2,opt,name=broadcastJobTemplate"`
}

// CronBroadcastJobStatus defines the observed state of CronBroadcastJob
type CronBroadcastJobStatus struct {

	// A list of pointers to currently running jobs.
	// +optional
	Active []corev1.ObjectReference `json:"active,omitempty"`

	// Information when was the last time the job was successfully scheduled.
	// +optional
	LastScheduleTime *metav1.Time `json:"lastScheduleTime,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CronBroadcastJob is the Schema for the cronbroadcastjobs API
type CronBroadcastJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CronBroadcastJobSpec   `json:"spec,omitempty"`
	Status CronBroadcastJobStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CronBroadcastJobList contains a list of CronBroadcastJob
type CronBroadcastJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CronBroadcastJob `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CronBroadcastJob{}, &CronBroadcastJobList{})
}
