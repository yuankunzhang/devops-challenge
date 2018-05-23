package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	bucketclientset "github.com/yuankunzhang/devops-challenge/kube-bucket/pkg/client/clientset/versioned"
	bucketinformer_v1 "github.com/yuankunzhang/devops-challenge/kube-bucket/pkg/client/informers/externalversions/bucket/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
)

const maxRetries = 5

// Event indicates the informerEvent
type Event struct {
	key    string
	action string
	oldObj interface{}
	newObj interface{}
}

// Controller represents the Kubernetes controller
// dealing with the Bucket resource.
type Controller struct {
	logger    *log.Entry
	clientset kubernetes.Interface
	queue     workqueue.RateLimitingInterface

	// informer lists and watches resource changes.
	informer cache.SharedIndexInformer

	// handler handles resource changes.
	handler Handler
}

// NewController creates a new instance of the Controller.
func NewController() (*Controller, error) {
	config, err := getKubeConfig()
	if err != nil {
		return nil, err
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	bucketClient, err := bucketclientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	// Create a new queue. When the informer gets a resource, we add an
	// identifying key to the queue so that it can be handled in the handler.
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	// Get the Bucket resource informer.
	informer := bucketinformer_v1.NewBucketInformer(
		bucketClient,
		meta_v1.NamespaceAll,
		0,
		cache.Indexers{},
	)

	var event Event
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			event.key, err = cache.MetaNamespaceKeyFunc(obj)
			event.action = "create"
			event.newObj = obj
			if err != nil {
				log.Errorf("creating resource error: %v", err)
			} else {
				log.Infof("creating resource: %s", event.key)
				queue.Add(event)
			}
		},
		DeleteFunc: func(obj interface{}) {
			event.key, err = cache.MetaNamespaceKeyFunc(obj)
			event.action = "delete"
			event.oldObj = obj
			if err != nil {
				log.Errorf("deleting resource error: %v", err)
			} else {
				log.Infof("deleting resource: %s", event.key)
				queue.Add(event)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			event.key, err = cache.MetaNamespaceKeyFunc(newObj)
			event.action = "update"
			event.oldObj = oldObj
			event.newObj = newObj
			if err != nil {
				log.Errorf("updating resource error: %v", err)
			} else {
				log.Infof("updating resource: %s", event.key)
				queue.Add(event)
			}
		},
	})

	awsSession, err := session.NewSession()
	if err != nil {
		return nil, err
	}

	// Create the controller
	controller := &Controller{
		logger:    log.NewEntry(log.New()),
		clientset: client,
		informer:  informer,
		queue:     queue,
		handler:   &BucketHandler{s3.New(awsSession)},
	}

	return controller, nil
}

// Run starts the controller.
func (c *Controller) Run(stop <-chan struct{}) {
	// Catch crash and log an error.
	defer runtime.HandleCrash()
	// Shutdown after all goroutines have finished handling
	// existing items.
	defer c.queue.ShutDown()

	// Start listing and watching resource changes.
	go c.informer.Run(stop)

	// Do a cache sync.
	if !cache.WaitForCacheSync(stop, c.HasSynced) {
		runtime.HandleError(fmt.Errorf("error syncing cached"))
		return
	}

	wait.Until(c.runWorker, time.Second, stop)
}

// HasSynced is required for the cache.Controller interface.
func (c *Controller) HasSynced() bool {
	return c.informer.HasSynced()
}

// LastSyncResourceVersion is required for the cache.Controller interface.
func (c *Controller) LastSyncResourceVersion() string {
	return c.informer.LastSyncResourceVersion()
}

// runWorker execute the loop to process new items added in the queue.
func (c *Controller) runWorker() {
	for c.processNextItem() {
	}
}

// processNextItem retrieves each queued item and takes the necessary
// handler action based on the event type.
func (c *Controller) processNextItem() bool {
	// Fetch the next item from queue.
	// If a shutdown is requested, stop the process.
	event, quit := c.queue.Get()

	if quit {
		return false
	}

	defer c.queue.Done(event)

	err := c.processItem(event.(Event))
	if err != nil {
		if c.queue.NumRequeues(event) < maxRetries {
			c.logger.Errorf("failed to process item with key %s, retrying. error: %v", event.(Event).key, err)
			c.queue.AddRateLimited(event)
		} else {
			c.logger.Errorf("failed to process item with key %s, no more retries. error: %v", event.(Event).key, err)
			c.queue.Forget(event)
			runtime.HandleError(err)
		}
	} else {
		c.queue.Forget(event)
	}

	// Keep the worker loop running by returning true.
	return true
}

// processItem processes a single item.
func (c *Controller) processItem(event Event) error {
	_, _, err := c.informer.GetIndexer().GetByKey(event.key)
	if err != nil {
		return err
	}

	switch event.action {
	case "create":
		c.handler.ObjectCreated(event.newObj)
	case "delete":
		c.handler.ObjectDeleted(event.oldObj)
	case "update":
		c.handler.ObjectUpdated(event.oldObj, event.newObj)
	}

	return nil
}

// getKubeConfig tries to get Kubernetes configuration.
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
