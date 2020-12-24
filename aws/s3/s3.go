package s3

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"golang.org/x/crypto/sha3"
)

type Manager struct {
	bucketName string
	region     string
}

func NewManager(bucket string, region string) *Manager {
	return &Manager{
		bucketName: bucket,
		region:     region,
	}
}

// Save data as a object in s3
func (m *Manager) Save(payload []byte) error {

	var bucket, key string
	var timeout time.Duration

	// generating hash and create object name
	h := sha3.Sum224(payload[:])
	objName := fmt.Sprintf("%v-%x", time.Now().Unix(), h)

	flag.StringVar(&bucket, "b", m.bucketName, "Bucket name.")
	flag.StringVar(&key, "k", objName, "Object key name.")
	flag.DurationVar(&timeout, "d", 0, "Upload timeout.")
	flag.Parse()

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(m.region)}))
	svc := s3.New(sess)

	r := bytes.NewReader(payload)

	// Uploads the object to S3. The Context will interrupt the request if the
	// timeout expires.
	_, err := svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(m.bucketName),
		Key:    aws.String(key),
		Body:   r,
	})

	return err
}

// GetKeys get all s3 bucket objects keys
func (m *Manager) GetKeys() ([]string, error) {

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(m.region)}))

	svc := s3.New(sess)

	var objects []string
	err := svc.ListObjectsPages(&s3.ListObjectsInput{
		Bucket: aws.String(m.bucketName),
	}, func(p *s3.ListObjectsOutput, lastPage bool) bool {
		for _, o := range p.Contents {
			objects = append(objects, aws.StringValue(o.Key))
		}
		return true // continue paging
	})
	if err != nil {
		panic(fmt.Sprintf("failed to list objects for bucket, %s, %v", m.bucketName, err))
		return nil, err
	}

	return objects, nil
}

// Get get a single s3 object by key
func (m *Manager) Get(key string) ([]byte, error) {

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(m.region)}))

	svc := s3.New(sess)
	obj, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(m.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(obj.Body)
	return body, nil
}

// Delete delete s3 object by key
func (m *Manager) Delete(key string) error {

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(m.region)}))

	svc := s3.New(sess)
	_, err := svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(m.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return err
	}

	return nil
}
