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

// Code generated by informer-gen. DO NOT EDIT.

package v1

import (
	time "time"

	bucket_v1 "github.com/yuankunzhang/devops-challenge/kube-bucket/pkg/apis/bucket/v1"
	versioned "github.com/yuankunzhang/devops-challenge/kube-bucket/pkg/client/clientset/versioned"
	internalinterfaces "github.com/yuankunzhang/devops-challenge/kube-bucket/pkg/client/informers/externalversions/internalinterfaces"
	v1 "github.com/yuankunzhang/devops-challenge/kube-bucket/pkg/client/listers/bucket/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// BucketInformer provides access to a shared informer and lister for
// Buckets.
type BucketInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1.BucketLister
}

type bucketInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewBucketInformer constructs a new informer for Bucket type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewBucketInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredBucketInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredBucketInformer constructs a new informer for Bucket type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredBucketInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options meta_v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.Storagek8sV1().Buckets(namespace).List(options)
			},
			WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.Storagek8sV1().Buckets(namespace).Watch(options)
			},
		},
		&bucket_v1.Bucket{},
		resyncPeriod,
		indexers,
	)
}

func (f *bucketInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredBucketInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *bucketInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&bucket_v1.Bucket{}, f.defaultInformer)
}

func (f *bucketInformer) Lister() v1.BucketLister {
	return v1.NewBucketLister(f.Informer().GetIndexer())
}
