package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/yuankunzhang/devops-challenge/kube-bucket/pkg/apis/bucket/v1"
)

// Handler specifies methods required for a handler.
type Handler interface {
	Init() error
	ObjectCreated(obj interface{})
	ObjectDeleted(obj interface{})
	ObjectUpdated(objOld, objNew interface{})
}

// BucketHandler is the handler to deal with the Bucket resource.
type BucketHandler struct {
	s3 *s3.S3
}

// Init implements the Handler interface.
func (b *BucketHandler) Init() error {
	return nil
}

// ObjectCreated implements the Handler interface.
func (b *BucketHandler) ObjectCreated(obj interface{}) {
	bucket := obj.(*v1.Bucket)
	bucketName := bucket.Spec.BucketName

	log.Infof("checking if bucket %s exists", bucketName)
	exists, err := b.bucketExists(bucketName)
	if err != nil {
		log.Errorf("failed to check bucket %s: %v", bucketName, err)
		return
	}

	if exists {
		log.Infof("bucket %s exists, skip creating", bucketName)
		return
	}

	log.Infof("creating bucket %s", bucketName)
	err = b.createBucket(bucketName)
	if err != nil {
		log.Errorf("failed to create bucket %s: %v", bucketName, err)
		return
	}

	log.Infof("creating bucket %s completed", bucketName)
}

// ObjectDelete implements the Handler interface.
func (b *BucketHandler) ObjectDeleted(obj interface{}) {
	bucket := obj.(*v1.Bucket)
	forceDelete := bucket.Spec.ForceDelete

	if forceDelete {
		bucketName := bucket.Spec.BucketName
		log.Infof("deleting bucket %s", bucketName)
		err := b.deleteBucket(bucketName)
		if err != nil {
			log.Errorf("failed to delete bucket %s: %v", bucketName, err)
			return
		}

		log.Infof("deleting bucket %s completed", bucketName)
	}
}

// ObjectUpdated implements the Handler interface.
func (b *BucketHandler) ObjectUpdated(objOld, objNew interface{}) {
	log.Info("object updated")
	// TODO(yuankun): complete this.
}

func (b *BucketHandler) bucketExists(bucket string) (bool, error) {
	_, err := b.s3.HeadBucket(&s3.HeadBucketInput{Bucket: aws.String(bucket)})

	if err == nil {
		// Exists.
		return true, nil
	}

	if aerr, ok := err.(awserr.Error); ok {
		// TODO(yuankun): might miss some cases.
		switch aerr.Code() {
		case s3.ErrCodeNoSuchBucket:
			fallthrough
		case "NotFound":
			// Not exists.
			return false, nil
		}
	}

	// Failed to check.
	return false, err
}

func (b *BucketHandler) createBucket(bucket string) error {
	_, err := b.s3.CreateBucket(&s3.CreateBucketInput{Bucket: aws.String(bucket)})
	if err != nil {
		return err
	}

	err = b.s3.WaitUntilBucketExists(&s3.HeadBucketInput{Bucket: aws.String(bucket)})
	if err != nil {
		return err
	}

	return nil
}

func (b *BucketHandler) deleteBucket(bucket string) error {
	hasMore := true
	total := 0

	for hasMore {
		resp, err := b.s3.ListObjects(&s3.ListObjectsInput{Bucket: aws.String(bucket)})
		if err != nil {
			return err
		}

		num := len(resp.Contents)
		if num == 0 {
			break
		}

		total += num

		var items s3.Delete
		var objs = make([]*s3.ObjectIdentifier, num)

		for i, o := range resp.Contents {
			objs[i] = &s3.ObjectIdentifier{Key: aws.String(*o.Key)}
		}

		// Add list of objects to delete.
		items.SetObjects(objs)

		// Delete the items.
		_, err = b.s3.DeleteObjects(&s3.DeleteObjectsInput{Bucket: aws.String(bucket), Delete: &items})

		if err != nil {
			return err
		}

		hasMore = *resp.IsTruncated
	}

	log.Infof("%d objects deleted from bucket %s", total, bucket)

	_, err := b.s3.DeleteBucket(&s3.DeleteBucketInput{Bucket: aws.String(bucket)})

	if err != nil {
		return err
	}

	err = b.s3.WaitUntilBucketNotExists(&s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})

	if err != nil {
		return err
	}

	return nil
}
