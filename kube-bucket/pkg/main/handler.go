package main

import (
	log "github.com/Sirupsen/logrus"
)

// Handler specifies methods required for a handler.
type Handler interface {
	Init() error
	ObjectCreated(obj interface{})
	ObjectDeleted(obj interface{})
	ObjectUpdated(objOld, objNew interface{})
}

// BucketHandler is the handler to deal with the Bucket resource.
type BucketHandler struct{}

// Init implements the Handler interface.
func (b *BucketHandler) Init() error {
	log.Info("BucketHandler.Init()")
	return nil
}

// ObjectCreated implements the Handler interface.
func (b *BucketHandler) ObjectCreated(obj interface{}) {
	log.Info("BucketHandler.ObjectCreated()")
}

// ObjectDelete implements the Handler interface.
func (b *BucketHandler) ObjectDeleted(obj interface{}) {
	log.Info("BucketHandler.ObjectDeleted()")
}

// ObjectUpdated implements the Handler interface.
func (b *BucketHandler) ObjectUpdated(objOld, objNew interface{}) {
	log.Info("BucketHandler.ObjectUpdated()")
}
