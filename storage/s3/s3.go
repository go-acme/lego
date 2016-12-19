package s3

import (
	"bytes"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"io/ioutil"
	"strings"
)

type Storage struct {
	service *s3.S3
}

func parsePath(p string) (string, string) {
	t := strings.Split(p, "/")
	return string(t[0]), strings.Join(t[1:], "/")
}

func NewStorageProvider() (*Storage, error) {
	sess, err := session.NewSession()
	if err != nil {
		return &Storage{}, err
	}
	service := s3.New(sess)
	return &Storage{service: service}, nil
}

func (s *Storage) CheckPath(path string) error {
	// we don't create bucket so far, we might need it to be local storage compible
	bucket, path := parsePath(path)
	params := &s3.ListObjectsV2Input{
		Bucket:  aws.String(bucket),
		MaxKeys: aws.Int64(1),
	}

	_, err := s.service.ListObjectsV2(params)
	if err != nil {
		return err
	}
	return nil
}

func (s *Storage) WritePath(path string, data []byte) error {
	bucket, p := parsePath(path)
	params := &s3.PutObjectInput{
		Bucket: aws.String(bucket), // Required
		Key:    aws.String(p),      // Required
		Body:   bytes.NewReader(data),
		Metadata: map[string]*string{
			"Key": aws.String("MetadataValue"), // Required
		},
	}
	_, err := s.service.PutObject(params)

	if err != nil {
		return err
	}
	return nil

}

func (s *Storage) ReadPath(path string) ([]byte, error) {
	bucket, p := parsePath(path)
	params := &s3.GetObjectInput{
		Bucket: aws.String(bucket), // Required
		Key:    aws.String(p),      // Required
	}
	resp, err := s.service.GetObject(params)
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	return b, err
}
