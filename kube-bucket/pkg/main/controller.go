package main

import (
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const maxRetries = 5

// Controller represents the Kubernetes controller
// dealing with the Bucket resource.
type Controller struct {
	logger       *log.Entry
	clientset    kubernetes.Interface
	queue        workqueue.RateLimitingInterface

	// informer lists and watches resource changes.
	informer     cache.SharedIndexInformer

	// handler handles resource changes.
	handler      Handler
}

// Run starts the controller.
func (c *Controller) Run(stop <-chan struct{}) {
	// Catch crash and log an error.
	defer runtime.HandleCrash()
	// Shutdown after all goroutines have finished handling
	// existing items.
	defer c.queue.ShutDown()

	c.logger.Info("Controller.Run(): initializing...")

	// Start listing and watching resource changes.
	go c.informer.Run(stop)

	// Do a cache sync.
	if !cache.WaitForCacheSync(stop, c.HasSynced) {
		runtime.HandleError(fmt.Errorf("error syncing cached"))
		return
	}

	c.logger.Info("Controller.Run(): cache sync completed")

	wait.Until(c.runWorker, time.Second, stop)
}

// HasSynced calls the informers HasSynced method.
func (c *Controller) HasSynced() bool {
	return c.informer.HasSynced()
}

// runWorker execute the loop to process new items added in the queue.
func (c *Controller) runWorker() {
	log.Info("Controller.runWorker(): starting...")

	for c.processNextItem() {
		log.Info("Controller.runWorker(): processing next item...")
	}

	log.Info("Controller.runWorker(): process completed")
}

// processNextItem retrieves each queued item and takes the necessary
// handler action based on the event type.
func (c *Controller) processNextItem() bool {
	log.Info("Controller.processNextItem(): starting...")

	// Fetch the next item from queue.
	// If a shutdown is requested, stop the process.
	key, quit := c.queue.Get()

	if quit {
		return false
	}

	defer c.queue.Done(key)

	// Format: namespace/name
	keyRaw := key.(string)

	// Get the object from the indexer using the raw string key.
	//
	// item contains the resource object. exists indicates whether
	// the resource was created (true) or deleted (false).
	item, exists, err := c.informer.GetIndexer().GetByKey(keyRaw)
	if err != nil {
		if c.queue.NumRequeues(key) < maxRetries {
			c.logger.Errorf("Controller.processNextItem(): failed to process item with key %s, retrying. error: %v", key, err)
			c.queue.AddRateLimited(key)
		} else {
			c.logger.Errorf("Controller.processNextItem(): failed to process item with key %s, no more retries. error: %v", key, err)
			c.queue.Forget(key)
			runtime.HandleError(err)
		}
	}

	if !exists {
		c.handler.ObjectDeleted(item)
		c.queue.Forget(key)
	} else {
		c.handler.ObjectCreated(item)
		c.queue.Forget(key)
	}

	// Keep the worker loop running by returning true.
	return true
}
