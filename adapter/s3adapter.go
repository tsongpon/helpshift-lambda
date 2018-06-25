package adapter

import (
	"bytes"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	s3Bucket = "repl-helpshift"
)

type S3Adapter struct {
	s3 *s3.S3
}

// NewS3Adapter create new NewS3Adapter by given session
func NewS3Adapter(s3 *s3.S3) *S3Adapter {
	s3Adapter := new(S3Adapter)
	s3Adapter.s3 = s3
	return s3Adapter
}

// Put put content to s3 bucket
func (adapter *S3Adapter) Put(key string, content []byte) error {
	// Config settings: this is where you choose the bucket, filename, content-type etc.
	// of the file you're uploading.
	_, err := adapter.s3.PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(s3Bucket),
		Key:                  aws.String(key),
		ACL:                  aws.String("private"),
		Body:                 bytes.NewReader(content),
		ContentLength:        aws.Int64(int64(len(content))),
		ContentType:          aws.String(http.DetectContentType(content)),
		ContentDisposition:   aws.String("attachment"),
		ServerSideEncryption: aws.String("AES256"),
	})
	return err
}
