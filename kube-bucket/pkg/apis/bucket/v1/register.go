package v1

import (
	"github.com/yuankunzhang/devops-challenge/kube-bucket/pkg/apis/bucket"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// SchemeGroupVersion is the identifier for the API which
// includes the name of the group and the API version.
var SchemeGroupVersion = schema.GroupVersion{
	Group:   bucket.GroupName,
	Version: "v1",
}

var (
	// SchemeBuilder uses custom function to add types to
	// the scheme.
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	// AddToScheme is a function that add types to the scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)

// Resource returns an instance of GroupResource.
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

// addKnownTypes adds the custom types to the API scheme by
// registering Bucket and BucketList in the scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(
		SchemeGroupVersion,
		&Bucket{},
		&BucketList{},
	)

	// Register.
	meta_v1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
