package main

import (
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	bucketclientset "github.com/yuankunzhang/devops-challenge/kube-bucket/pkg/client/clientset/versioned"
	bucketinformer_v1 "github.com/yuankunzhang/devops-challenge/kube-bucket/pkg/client/informers/externalversions/bucket/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func main() {
	client, bucketClient, err := getKubeClient()
	if err != nil {
		log.Fatal(err)
	}

	// Get the Bucket resource informer.
	informer := bucketinformer_v1.NewBucketInformer(
		bucketClient,
		meta_v1.NamespaceAll,
		0,
		cache.Indexers{},
	)

	// Create a new queue. When the informer gets a resource, we add an
	// identifying key to the queue so that it can be handled in the handler.
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	// Add event handlers to handle add/delete/update event.
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err != nil {
				log.Errorf("Main(): add resource error: %v", err)
			} else {
				log.Infof("Main(): add resource: %s", key)
				queue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err != nil {
				log.Errorf("Main(): delete resource error: %v", err)
			} else {
				log.Infof("Main(): delete resource: %s", key)
				queue.Add(key)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(newObj)
			if err != nil {
				log.Errorf("Main(): update resource error: %v", err)
			} else {
				log.Infof("Main(): update resource: %s", key)
				queue.Add(key)
			}
		},
	})

	awsSession, err := session.NewSession()
	if err != nil {
		log.Fatalf("Main(): failed to create aws session: %v", err)
	}

	// Create the controller
	controller := Controller{
		logger:    log.NewEntry(log.New()),
		clientset: client,
		informer:  informer,
		queue:     queue,
		handler:   &BucketHandler{s3.New(awsSession)},
	}

	stop := make(chan struct{})
	defer close(stop)

	// Run the worker loop to process items.
	go controller.Run(stop)

	term := make(chan os.Signal)
	signal.Notify(term, syscall.SIGTERM, syscall.SIGINT)

	// Waiting for SIGTERM or SIGINT.
	<-term
}

func getKubeClient() (kubernetes.Interface, bucketclientset.Interface, error) {
	config, err := getKubeConfig()
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to get kube config")
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create kube client instance")
	}

	bucketClient, err := bucketclientset.NewForConfig(config)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create bucket client")
	}

	return client, bucketClient, nil
}

func getKubeConfig() (*rest.Config, error) {
	// Try to get in-cluster config first.
	config, err := rest.InClusterConfig()
	if err == nil {
		return config, nil
	}

	// Try to get out-of-cluster config then, can be failed.
	configPath := os.Getenv("KUBECONFIG")
	if configPath == "" {
		configPath = filepath.Join(os.Getenv("HOME"), ".kube", "config")
	}
	return clientcmd.BuildConfigFromFlags("", configPath)
}
