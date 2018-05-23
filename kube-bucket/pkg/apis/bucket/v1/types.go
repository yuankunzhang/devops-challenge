package v1

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Bucket describes a Bucket resource.
type Bucket struct {
	// TypeMeta is the resource metadata, including kind and apiVersion.
	meta_v1.TypeMeta `json:",inline"`
	// ObjectMeta is the object metadata, including name, namespace, and others.
	meta_v1.ObjectMeta `json:"metadata,omitempty"`
	// Spec is the resource spec.
	Spec BucketSpec `json:"spec"`
}

// BucketSpec describes the spec for a Bucket resource.
type BucketSpec struct {
	BucketName string `json:"bucketName"`
	Region     string `json:"region"`

	// ForceDelete controls whether or not to delete the remote bucket
	// when the Bucket resource is deleted.
	ForceDelete bool `json:"forceDelete"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BucketList describes a list of Bucket resources.
type BucketList struct {
	meta_v1.TypeMeta `json:",inline"`
	meta_v1.ListMeta `json:"metadata"`
	Items            []Bucket `json:"items"`
}
