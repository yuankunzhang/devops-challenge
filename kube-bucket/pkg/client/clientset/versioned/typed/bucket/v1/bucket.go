/*
Copyright The Kubernetes Authors.

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

// Code generated by client-gen. DO NOT EDIT.

package v1

import (
	v1 "github.com/yuankunzhang/devops-challenge/kube-bucket/pkg/apis/bucket/v1"
	scheme "github.com/yuankunzhang/devops-challenge/kube-bucket/pkg/client/clientset/versioned/scheme"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// BucketsGetter has a method to return a BucketInterface.
// A group's client should implement this interface.
type BucketsGetter interface {
	Buckets(namespace string) BucketInterface
}

// BucketInterface has methods to work with Bucket resources.
type BucketInterface interface {
	Create(*v1.Bucket) (*v1.Bucket, error)
	Update(*v1.Bucket) (*v1.Bucket, error)
	Delete(name string, options *meta_v1.DeleteOptions) error
	DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error
	Get(name string, options meta_v1.GetOptions) (*v1.Bucket, error)
	List(opts meta_v1.ListOptions) (*v1.BucketList, error)
	Watch(opts meta_v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.Bucket, err error)
	BucketExpansion
}

// buckets implements BucketInterface
type buckets struct {
	client rest.Interface
	ns     string
}

// newBuckets returns a Buckets
func newBuckets(c *Storagek8sV1Client, namespace string) *buckets {
	return &buckets{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the bucket, and returns the corresponding bucket object, and an error if there is any.
func (c *buckets) Get(name string, options meta_v1.GetOptions) (result *v1.Bucket, err error) {
	result = &v1.Bucket{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("buckets").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Buckets that match those selectors.
func (c *buckets) List(opts meta_v1.ListOptions) (result *v1.BucketList, err error) {
	result = &v1.BucketList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("buckets").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested buckets.
func (c *buckets) Watch(opts meta_v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("buckets").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a bucket and creates it.  Returns the server's representation of the bucket, and an error, if there is any.
func (c *buckets) Create(bucket *v1.Bucket) (result *v1.Bucket, err error) {
	result = &v1.Bucket{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("buckets").
		Body(bucket).
		Do().
		Into(result)
	return
}

// Update takes the representation of a bucket and updates it. Returns the server's representation of the bucket, and an error, if there is any.
func (c *buckets) Update(bucket *v1.Bucket) (result *v1.Bucket, err error) {
	result = &v1.Bucket{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("buckets").
		Name(bucket.Name).
		Body(bucket).
		Do().
		Into(result)
	return
}

// Delete takes name of the bucket and deletes it. Returns an error if one occurs.
func (c *buckets) Delete(name string, options *meta_v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("buckets").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *buckets) DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("buckets").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched bucket.
func (c *buckets) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.Bucket, err error) {
	result = &v1.Bucket{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("buckets").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
