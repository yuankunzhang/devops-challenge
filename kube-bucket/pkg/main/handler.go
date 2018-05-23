package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
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
	log.Info("BucketHandler.Init()")
	return nil
}

// ObjectCreated implements the Handler interface.
func (b *BucketHandler) ObjectCreated(obj interface{}) {
	log.Info("BucketHandler.ObjectCreated()")
	bucket := obj.(*v1.Bucket)
	bucketName := bucket.Spec.BucketName

	_, err := b.s3.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})

	if err != nil {
		log.Errorf("BucketHandler.ObjectCreated(): %v", err)
		return
	}

	err = b.s3.WaitUntilBucketExists(&s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})

	if err != nil {
		log.Errorf("BucketHandler.ObjectCreated(): %v", err)
		return
	}

	log.Infof("BucketHandler.ObjectCreated(): bucket %s created", bucketName)
}

// ObjectDelete implements the Handler interface.
func (b *BucketHandler) ObjectDeleted(obj interface{}) {
	log.Info("BucketHandler.ObjectDeleted()")
	//bucket := obj.(*v1.Bucket)
	//forceDelete := bucket.Spec.ForceDelete

	//if forceDelete {
	//bucketName := bucket.Spec.BucketName
	//b.deleteS3Bucket(bucketName)
	//}
}

// ObjectUpdated implements the Handler interface.
func (b *BucketHandler) ObjectUpdated(objOld, objNew interface{}) {
	log.Info("BucketHandler.ObjectUpdated()")
}

func (b *BucketHandler) deleteS3Bucket(bucket string) {
	hasMore := true
	total := 0

	for hasMore {
		resp, err := b.s3.ListObjects(&s3.ListObjectsInput{
			Bucket: aws.String(bucket),
		})
		if err != nil {
			log.Errorf("BucketHandler.deleteS3Bucket(): %v", err)
			return
		}

		num := len(resp.Contents)
		total += num

		var items s3.Delete
		var objs = make([]*s3.ObjectIdentifier, num)

		for i, o := range resp.Contents {
			objs[i] = &s3.ObjectIdentifier{Key: aws.String(*o.Key)}
		}

		// Add list of objects to delete.
		items.SetObjects(objs)

		// Delete the items.
		_, err = b.s3.DeleteObjects(&s3.DeleteObjectsInput{
			Bucket: aws.String(bucket),
			Delete: &items,
		})

		if err != nil {
			log.Errorf("BucketHandler.deleteS3Bucket(): %v", err)
			return
		}

		hasMore = *resp.IsTruncated
	}

	log.Infof("BucketHandler.deleteS3Bucket(): %d objects deleted from bucket %s", total, bucket)

	_, err := b.s3.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(bucket),
	})

	if err != nil {
		log.Errorf("BucketHandler.deleteS3Bucket(): %v", err)
		return
	}

	log.Infof("BucketHandler.deleteS3Bucket(): bucket %s deleted", bucket)
}
