package main

import (
	"log"
	"time"

	"bitbucket.org/redplanethotels/helpshift-lamda/adapter"
	producer "github.com/a8m/kinesis-producer"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var initialized = false
var ginLambda *ginadapter.GinLambda

const (
	S3_REGION          = "ap-southeast-1"
	S3_BUCKET          = "repl-helpshift"
	KINESIS_TOPIC_NAME = "helpshift"
	timeLayout         = "2006-01-02T15:04:05.000"
)

// func addFileToS3(s *session.Session, content []byte) error {

// 	s3KeyName := strconv.Itoa(time.Now().Nanosecond()) + ".json"
// 	// Config settings: this is where you choose the bucket, filename, content-type etc.
// 	// of the file you're uploading.
// 	_, err := s3.New(s).PutObject(&s3.PutObjectInput{
// 		Bucket:               aws.String(S3_BUCKET),
// 		Key:                  aws.String(s3KeyName),
// 		ACL:                  aws.String("private"),
// 		Body:                 bytes.NewReader(content),
// 		ContentLength:        aws.Int64(int64(len(content))),
// 		ContentType:          aws.String(http.DetectContentType(content)),
// 		ContentDisposition:   aws.String("attachment"),
// 		ServerSideEncryption: aws.String("AES256"),
// 	})
// 	return err
// }

func sendToKinesis() {
	log := logrus.New()
	client := kinesis.New(session.New(aws.NewConfig()))
	pr := producer.New(&producer.Config{
		StreamName:   KINESIS_TOPIC_NAME,
		BacklogCount: 2000,
		Client:       client,
		Logger:       log,
	})

	pr.Start()

	err := pr.Put([]byte("foo"), "bar")
	if err != nil {
		log.WithError(err).Fatal("error producing")
	}
}

// Handler handle lamda execution
func Handler(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	if !initialized {
		s3Session, err := session.NewSession(&aws.Config{Region: aws.String(S3_REGION)})
		if err != nil {
			log.Fatal(err)
		}

		s3Instance := s3.New(s3Session)
		s3Adapter := adapter.NewS3Adapter(s3Instance)

		log := logrus.New()
		client := kinesis.New(session.New(aws.NewConfig()))
		pr := producer.New(&producer.Config{
			StreamName:    KINESIS_TOPIC_NAME,
			BacklogCount:  2000,
			Client:        client,
			Logger:        log,
			FlushInterval: 5,
		})

		pr.Start()
		defer pr.Stop()

		kinesisAdapter := adapter.NewKinesisAdapter(pr)

		// stdout and stderr are sent to AWS CloudWatch Logs
		log.Printf("Gin cold start")
		r := gin.Default()
		r.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "pong",
			})
		})

		r.POST("/hello", func(c *gin.Context) {
			body := make([]byte, c.Request.ContentLength)
			c.Request.Body.Read(body)

			key := time.Now().Format(timeLayout) + ".json"
			err = s3Adapter.Put(key, body)
			if err != nil {
				log.Fatal(err)
			}

			err := kinesisAdapter.Send(key, body)
			if err != nil {
				log.Printf("Fail send data to kinesis")
			}

			c.JSON(200, gin.H{
				"message": "ok",
			})
		})

		ginLambda = ginadapter.New(r)
		initialized = true
	}

	// If no name is provided in the HTTP request body, throw an error
	return ginLambda.Proxy(req)
}

func main() {
	lambda.Start(Handler)
}
